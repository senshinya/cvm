package sema

import (
	"shinya.click/cvm/common"
	"shinya.click/cvm/entity"
	"shinya.click/cvm/parser"
)

type Sema struct {
	Types   *TypeTable
	SymTab  *SymbolTable
	Options SemaOptions
	scope   *Scope
	errors  []*common.CvmError

	pendingFuncs   []*pendingFunc
	funcCtx        *funcCtx
	allowArrayStar bool
}

type SemaOptions struct {
	PedanticErrors                  bool
	GNUExtensions                   bool
	Permissive                      bool
	WErrorDeclarationAfterStatement bool
}

type pendingFunc struct {
	def         *FuncDef
	bodyAst     *entity.AstNode
	parentScope *Scope
}

type SemaResult struct {
	Program *Program
	Errors  []*common.CvmError
	Source  *entity.AstNode
}

func NewSema() *Sema {
	return NewSemaWithOptions(SemaOptions{})
}

func NewSemaWithOptions(opts SemaOptions) *Sema {
	s := &Sema{Types: NewTypeTable(), SymTab: NewSymbolTable(), Options: opts}
	s.scope = s.SymTab.File
	return s
}

func (s *Sema) analyzeOne(root *entity.AstNode) *SemaResult {
	prog := &Program{Types: s.Types, SymTab: s.SymTab}
	s.walkTranslationUnit(root, prog)
	for i := 0; i < len(s.pendingFuncs); i++ {
		s.walkFunctionBody(s.pendingFuncs[i], prog)
	}
	s.foldStaticInitializers(prog)
	s.markStaticFunctionUsesInGlobals(prog)
	s.validateStaticFunctionDefinitions(prog)
	FinalizeProgramLayout(prog)
	return &SemaResult{Program: prog, Errors: s.errors, Source: root}
}

func (s *Sema) report(err *common.CvmError) {
	if err != nil {
		s.errors = append(s.errors, err)
	}
}

func (s *Sema) walkTranslationUnit(node *entity.AstNode, prog *Program) {
	if node == nil {
		return
	}
	if node.Typ == parser.Program {
		s.walkTranslationUnit(node.Children[0], prog)
		return
	}
	switch {
	case node.ReducedBy(parser.TranslationUnit, 1):
		s.walkExternalDeclaration(node.Children[0], prog)
	case node.ReducedBy(parser.TranslationUnit, 2):
		s.walkTranslationUnit(node.Children[0], prog)
		s.walkExternalDeclaration(node.Children[1], prog)
	}
}

func (s *Sema) walkExternalDeclaration(node *entity.AstNode, prog *Program) {
	switch {
	case node.ReducedBy(parser.ExternalDeclaration, 1):
		s.walkFunctionDefinition(node.Children[0], prog)
	case node.ReducedBy(parser.ExternalDeclaration, 2):
		s.walkDeclaration(node.Children[0], prog)
	case node.ReducedBy(parser.ExternalDeclaration, 3):
		return
	}
}

func (s *Sema) walkFunctionDefinition(node *entity.AstNode, prog *Program) {
	spec := s.parseSpec(node.Children[0])
	s.validateDeclaratorArrayQualifiers(node.Children[1], false)
	t, name := s.applyDeclarator(node.Children[1], spec.Type)
	s.validateRestrictType(t, node.Children[1].SourceStart)
	s.validateOldStyleImplicitIntParams(node)
	ft, ok := t.(*FunctionType)
	if !ok {
		s.report(InvalidTypeSpec(node.SourceStart, "function definition declarator did not yield a function type"))
		return
	}
	oldStyleParamTypes := s.oldStyleDefinitionParamTypes(node, functionDeclaratorIdentifierList(node.Children[1]))
	s.validateInlineSpecifier(spec, t, name, node.Children[1].SourceStart, false)
	sym := s.scope.LookupCurrent(name, NSOrdinary)
	if sym == nil {
		sym = &Symbol{Name: name, Kind: SymFunc, T: ft, Storage: spec.Storage, Linkage: LinkageExternal, Pos: node.Children[1].SourceStart}
		s.scope.Insert(name, sym)
	} else if sym.Kind != SymFunc {
		s.report(RedefinitionSymbol(node.Children[1].SourceStart, sym.Pos, name))
		return
	} else if prev := firstFunctionDefinition(sym); prev != nil {
		s.report(RedefinitionSymbol(node.Children[1].SourceStart, prev.Pos().SourceStart, name))
		return
	} else {
		if prev, ok := unqual(sym.T).(*FunctionType); ok {
			if !sameQualifiers(prev.Ret, ft.Ret) {
				s.report(RedefinitionSymbol(node.Children[1].SourceStart, sym.Pos, name))
				return
			}
		}
		s.validateOldStyleDefinitionPrototype(node.SourceStart, oldStyleParamTypes, sym)
	}
	def := &FuncDef{
		Sym:                sym,
		T:                  ft,
		OldStyleParamTypes: oldStyleParamTypes,
		IsInlineDefinition: spec.IsInline && spec.Storage != StorageExtern && spec.Storage != StorageStatic,
		Range:              node.SourceRange,
		Labels:             map[string]*LabeledStmt{},
	}
	def.Params = s.collectParamDecls(node.Children[1], ft)
	sym.Decl = def
	sym.Defs = append(sym.Defs, def)
	prog.Funcs = append(prog.Funcs, def)
	s.pendingFuncs = append(s.pendingFuncs, &pendingFunc{def: def, bodyAst: node.Children[len(node.Children)-1], parentScope: s.scope})
}

func firstFunctionDefinition(sym *Symbol) *FuncDef {
	if sym == nil {
		return nil
	}
	for _, d := range sym.Defs {
		if def, ok := d.(*FuncDef); ok {
			return def
		}
	}
	return nil
}

func (s *Sema) walkDeclaration(node *entity.AstNode, prog *Program) {
	if node.ReducedBy(parser.Declaration, 3) {
		s.typeStaticAssert(node.Children[0])
		return
	}
	invalidEmptyTagRedecl := s.qualifiedEmptyTagRedeclaration(node.Children[0])
	spec := s.parseSpec(node.Children[0])
	if s.Options.PedanticErrors && hasEnumReferenceSpecifier(node.Children[0]) {
		s.report(InvalidTypeSpec(node.SourceStart, "ISO C forbids forward references to enum types"))
	}
	if node.ReducedBy(parser.Declaration, 1) {
		s.validateRestrictType(spec.Type, node.SourceStart)
		if invalidEmptyTagRedecl {
			s.report(InvalidTypeSpec(node.SourceStart, "empty declaration with type qualifier or storage class does not redeclare tag"))
		}
		if isTagType(spec.Type) {
			prog.Globals = append(prog.Globals, &TagDecl{T: spec.Type, Range: node.SourceRange})
		}
		return
	}
	s.walkInitDeclaratorList(node.Children[1], spec, prog, node.SourceRange)
}

func (s *Sema) typeStaticAssert(node *entity.AstNode) {
	expr := s.typeExpr(node.Children[2], s.scope)
	cv, ok := NewEvaluator(s).EvalIntegerConstant(expr)
	if !ok {
		s.report(InvalidTypeSpec(node.SourceStart, "static assertion expression is not an integer constant"))
		return
	}
	if cv.Int == 0 {
		s.report(InvalidTypeSpec(node.SourceStart, "static assertion failed"))
		return
	}
	if s.Options.PedanticErrors {
		s.report(InvalidTypeSpec(node.SourceStart, "ISO C99 does not support '_Static_assert'"))
	}
}

func isTagType(t Type) bool {
	switch unqual(t).(type) {
	case *StructType, *UnionType, *EnumType:
		return true
	}
	return false
}

func hasEnumReferenceSpecifier(node *entity.AstNode) bool {
	if node == nil {
		return false
	}
	if node.ReducedBy(parser.TypeSpecifier, 13) {
		return node.Children[0].ReducedBy(parser.EnumSpecifier, 5)
	}
	for _, child := range node.Children {
		if hasEnumReferenceSpecifier(child) {
			return true
		}
	}
	return false
}

func (s *Sema) qualifiedEmptyTagRedeclaration(specs *entity.AstNode) bool {
	if !declarationSpecHasQualifierOrStorage(specs) {
		return false
	}
	name, ok := tagReferenceWithoutDefinition(specs)
	if !ok {
		return false
	}
	return s.scope.LookupTag(name) != nil
}

func declarationSpecHasQualifierOrStorage(node *entity.AstNode) bool {
	if node == nil {
		return false
	}
	switch node.Typ {
	case parser.TypeQualifier, parser.StorageClassSpecifier:
		return true
	}
	for _, child := range node.Children {
		if declarationSpecHasQualifierOrStorage(child) {
			return true
		}
	}
	return false
}

func tagReferenceWithoutDefinition(node *entity.AstNode) (string, bool) {
	if node == nil {
		return "", false
	}
	if node.ReducedBy(parser.TypeSpecifier, 12) {
		return structUnionReferenceWithoutDefinition(node.Children[0])
	}
	if node.ReducedBy(parser.TypeSpecifier, 13) {
		enum := node.Children[0]
		if enum.ReducedBy(parser.EnumSpecifier, 5) {
			return enum.Children[1].Terminal.Lexeme, true
		}
		return "", false
	}
	for _, child := range node.Children {
		if name, ok := tagReferenceWithoutDefinition(child); ok {
			return name, true
		}
	}
	return "", false
}

func structUnionReferenceWithoutDefinition(node *entity.AstNode) (string, bool) {
	if node.ReducedBy(parser.StructOrUnionSpecifier, 3) {
		return node.Children[1].Terminal.Lexeme, true
	}
	return "", false
}

func (s *Sema) walkInitDeclaratorList(node *entity.AstNode, spec SpecResult, prog *Program, srcRange entity.SourceRange) {
	switch {
	case node.ReducedBy(parser.InitDeclaratorList, 1):
		s.walkInitDeclarator(node.Children[0], spec, prog, srcRange)
	case node.ReducedBy(parser.InitDeclaratorList, 2):
		s.walkInitDeclaratorList(node.Children[0], spec, prog, srcRange)
		s.walkInitDeclarator(node.Children[2], spec, prog, srcRange)
	}
}

func (s *Sema) walkInitDeclarator(node *entity.AstNode, spec SpecResult, prog *Program, srcRange entity.SourceRange) {
	s.validateDeclaratorArrayQualifiers(node.Children[0], false)
	t, name := s.applyDeclarator(node.Children[0], spec.Type)
	pos := node.Children[0].SourceStart
	if name == "" {
		s.report(InvalidTypeSpec(pos, "missing declarator name"))
		return
	}
	s.validateInlineSpecifier(spec, t, name, pos, false)
	s.validateRestrictType(t, pos)
	if spec.IsTypedef {
		if typeHasDisallowedFileScopeVMType(t) {
			s.report(InvalidTypeSpec(pos, "array size must be integer constant expression"))
		}
		markTypedefVMBounds(t)
		sym := &Symbol{Name: name, Kind: SymTypedef, T: t, Storage: StorageTypedef, Pos: pos}
		td := &TypedefDecl{Sym: sym, T: t, Range: srcRange}
		sym.Decl = td
		if err := s.scope.InsertChecked(name, sym); err != nil {
			s.report(err.(*common.CvmError))
			return
		}
		prog.Globals = append(prog.Globals, td)
		return
	}
	if ft, ok := unqual(t).(*FunctionType); ok {
		if node.ReducedBy(parser.InitDeclarator, 2) {
			s.report(InvalidTypeSpec(pos, "function declarator cannot have initializer"))
			return
		}
		s.declareFunction(name, ft, spec.Storage, pos, srcRange, prog)
		return
	}
	storage := spec.Storage
	if typeHasDisallowedFileScopeVMType(t) {
		s.report(InvalidTypeSpec(pos, "array size must be integer constant expression"))
	}
	sym := &Symbol{Name: name, Kind: SymVar, T: t, Storage: storage, Pos: pos}
	if s.scope.Kind == ScopeFile {
		sym.Linkage = LinkageExternal
		if storage == StorageStatic {
			sym.Linkage = LinkageInternal
		}
	}
	vd := &VarDecl{Sym: sym, T: t, Storage: storage, Range: srcRange}
	sym.Decl = vd
	if err := s.scope.InsertChecked(name, sym); err != nil {
		s.report(err.(*common.CvmError))
		return
	}
	if node.ReducedBy(parser.InitDeclarator, 2) {
		vd.Init = s.typeInitializer(node.Children[2], t)
		vd.T = s.completeUnsizedArrayInitializerType(vd.T, vd.Init)
		sym.T = vd.T
		if il, ok := vd.Init.(*InitList); ok {
			il.T = vd.T
		}
	}
	prog.Globals = append(prog.Globals, vd)
}

func (s *Sema) declareFunction(name string, ft *FunctionType, storage StorageClass, pos entity.SourcePos, srcRange entity.SourceRange, prog *Program) {
	s.validateFunctionVMReturn(ft, pos)
	sym := s.scope.LookupCurrent(name, NSOrdinary)
	if sym == nil {
		sym = &Symbol{Name: name, Kind: SymFunc, T: ft, Storage: storage, Pos: pos, Linkage: LinkageExternal}
		s.scope.Insert(name, sym)
	} else if sym.Kind != SymFunc {
		s.report(RedefinitionSymbol(pos, sym.Pos, name))
		return
	} else if !s.mergeFunctionDeclaration(sym, ft, pos) {
		return
	}
	fd := &FuncDecl{Sym: sym, T: ft, Storage: storage, Range: srcRange}
	sym.Decl = fd
	sym.Defs = append(sym.Defs, fd)
	prog.Globals = append(prog.Globals, fd)
}

func (s *Sema) validateFunctionVMReturn(ft *FunctionType, pos entity.SourcePos) {
	if ft != nil && typeHasVariablyModifiedType(ft.Ret) {
		s.report(InvalidTypeSpec(pos, "function return type cannot be variably modified"))
	}
}

func (s *Sema) mergeFunctionDeclaration(sym *Symbol, ft *FunctionType, pos entity.SourcePos) bool {
	prev, ok := unqual(sym.T).(*FunctionType)
	if ok && !compatibleFunctionType(prev, ft) {
		if s.Options.GNUExtensions && compatibleFunctionTypeWithTransparentUnion(prev, ft) {
			sym.T = ft
			return true
		}
		s.report(RedefinitionSymbol(pos, sym.Pos, sym.Name))
		return false
	}
	if ok && !sameQualifiers(prev.Ret, ft.Ret) {
		s.report(RedefinitionSymbol(pos, sym.Pos, sym.Name))
		return false
	}
	if ft.HasProto {
		for _, d := range sym.Defs {
			def, ok := d.(*FuncDef)
			if !ok || len(def.OldStyleParamTypes) == 0 {
				continue
			}
			if !compatiblePrototypeWithOldStyleParams(ft, def.OldStyleParamTypes) {
				s.report(InvalidTypeSpec(pos, "old-style definition parameter type does not match prototype"))
				return false
			}
		}
		// 后续带 prototype 的声明会细化早先的旧式声明；反过来不能丢掉
		// 已知 prototype，因为旧式函数定义需要用它做兼容性校验。
		if !ok || !prev.HasProto {
			sym.T = ft
		}
	}
	return true
}

func (s *Sema) walkFunctionBody(pf *pendingFunc, prog *Program) {
	parent := pf.parentScope
	if parent == nil {
		parent = s.SymTab.File
	}
	bodyScope := NewScope(ScopeFunc, parent)
	for _, p := range pf.def.Params {
		if p.Sym != nil && p.Sym.Name != "" {
			_ = bodyScope.InsertChecked(p.Sym.Name, p.Sym)
		}
	}
	// C99 要求每个函数体内隐式声明 __func__，类型等价于 static const char[N]。
	funcType := s.Types.ArrayConstant(s.Types.Qualified(s.Types.Builtin(Char), true, false, false), int64(len(pf.def.Sym.Name)+1))
	funcSym := &Symbol{
		Name:    "__func__",
		Kind:    SymVar,
		T:       funcType,
		Storage: StorageStatic,
		Linkage: LinkageNone,
		Pos:     pf.def.Range.SourceStart,
	}
	funcDecl := &VarDecl{
		Sym:     funcSym,
		T:       funcType,
		Init:    &StringLit{Value: pf.def.Sym.Name, T: funcType, Range: pf.def.Range},
		Storage: StorageStatic,
		Range:   pf.def.Range,
	}
	funcSym.Decl = funcDecl
	_ = bodyScope.InsertChecked("__func__", funcSym)
	prev := s.scope
	s.scope = bodyScope
	defer func() { s.scope = prev }()
	ctx := &funcCtx{def: pf.def, prog: prog}
	prevCtx := s.funcCtx
	s.funcCtx = ctx
	defer func() { s.funcCtx = prevCtx }()
	body, _ := s.typeStmt(pf.bodyAst, bodyScope, ctx).(*Block)
	if body == nil {
		body = &Block{Range: pf.bodyAst.SourceRange, Scope: bodyScope}
	}
	if funcSym.Used {
		pf.def.Locals = append([]*VarDecl{funcDecl}, pf.def.Locals...)
	}
	pf.def.Body = body
	pf.def.Labels = map[string]*LabeledStmt{}
	collectLabels(body, pf.def.Labels)
	resolveGotos(ctx.pendingGotos, pf.def.Labels, ctx.vmScopes, s)
	s.validateInlineDefinitionBody(pf.def)
	s.markStaticFunctionUsesInStmt(body)
	_ = prog
}

func (s *Sema) validateInlineDefinitionBody(def *FuncDef) {
	if def == nil || !def.IsInlineDefinition || !s.Options.PedanticErrors {
		return
	}
	s.validateInlineDefinitionStmt(def.Body)
}

func (s *Sema) validateInlineDefinitionStmt(stmt Stmt) {
	switch x := stmt.(type) {
	case nil:
		return
	case *Block:
		for _, it := range x.Items {
			s.validateInlineDefinitionStmt(it)
		}
	case *IfStmt:
		s.validateInlineDefinitionExpr(x.Cond)
		s.validateInlineDefinitionStmt(x.Then)
		s.validateInlineDefinitionStmt(x.Else)
	case *WhileStmt:
		s.validateInlineDefinitionExpr(x.Cond)
		s.validateInlineDefinitionStmt(x.Body)
	case *ForStmt:
		s.validateInlineDefinitionStmt(x.Init)
		s.validateInlineDefinitionExpr(x.Cond)
		s.validateInlineDefinitionExpr(x.Post)
		s.validateInlineDefinitionStmt(x.Body)
	case *SwitchStmt:
		s.validateInlineDefinitionExpr(x.Cond)
		s.validateInlineDefinitionStmt(x.Body)
	case *CaseStmt:
		s.validateInlineDefinitionStmt(x.Body)
	case *DefaultStmt:
		s.validateInlineDefinitionStmt(x.Body)
	case *LabeledStmt:
		s.validateInlineDefinitionStmt(x.Body)
	case *ExprStmt:
		s.validateInlineDefinitionExpr(x.Expr)
	case *ReturnStmt:
		s.validateInlineDefinitionExpr(x.Value)
	case *DeclStmt:
		for _, d := range x.Decls {
			s.validateInlineDefinitionDecl(d)
		}
	}
}

func (s *Sema) validateInlineDefinitionDecl(decl Decl) {
	vd, ok := decl.(*VarDecl)
	if !ok {
		return
	}
	if vd.Storage == StorageStatic && typeIsModifiableObject(vd.T) {
		s.report(InvalidTypeSpec(vd.Range.SourceStart, "modifiable static object declared in inline definition"))
	}
	s.validateInlineDefinitionExpr(vd.Init)
}

func (s *Sema) validateInlineDefinitionExpr(expr Expr) {
	switch x := expr.(type) {
	case nil:
		return
	case *VarRef:
		if x.Sym != nil && x.Sym.Linkage == LinkageInternal {
			s.report(InvalidTypeSpec(x.Range.SourceStart, "internal linkage identifier used in inline definition"))
		}
	case *BinOp:
		s.validateInlineDefinitionExpr(x.L)
		s.validateInlineDefinitionExpr(x.R)
	case *UnOp:
		s.validateInlineDefinitionExpr(x.X)
	case *AssignExpr:
		s.validateInlineDefinitionExpr(x.L)
		s.validateInlineDefinitionExpr(x.R)
	case *CompoundAssign:
		s.validateInlineDefinitionExpr(x.L)
		s.validateInlineDefinitionExpr(x.R)
	case *CallExpr:
		s.validateInlineDefinitionExpr(x.Callee)
		for _, arg := range x.Args {
			s.validateInlineDefinitionExpr(arg)
		}
	case *MemberExpr:
		s.validateInlineDefinitionExpr(x.Base)
	case *IndexExpr:
		s.validateInlineDefinitionExpr(x.Base)
		s.validateInlineDefinitionExpr(x.Index)
	case *CondExpr:
		s.validateInlineDefinitionExpr(x.Cond)
		s.validateInlineDefinitionExpr(x.Then)
		s.validateInlineDefinitionExpr(x.Else)
	case *SizeofExpr:
		if x.Operand.Expr != nil && typeHasVariableSize(x.Operand.Expr.GetType()) {
			s.validateInlineDefinitionExpr(x.Operand.Expr)
		}
	case *CommaExpr:
		s.validateInlineDefinitionExpr(x.L)
		s.validateInlineDefinitionExpr(x.R)
	case *CompoundLit:
		s.validateInlineDefinitionInitList(x.Init)
	case *InitList:
		s.validateInlineDefinitionInitList(x)
	case *ImplicitCast:
		s.validateInlineDefinitionExpr(x.X)
	case *ExplicitCast:
		s.validateInlineDefinitionExpr(x.X)
	}
}

func (s *Sema) validateInlineDefinitionInitList(il *InitList) {
	if il == nil {
		return
	}
	for _, elem := range il.Elems {
		s.validateInlineDefinitionExpr(elem.Value)
	}
}

func typeIsModifiableObject(t Type) bool {
	if q, ok := t.(*QualType); ok && q.Const {
		return false
	}
	if at, ok := unqual(t).(*ArrayType); ok {
		return typeIsModifiableObject(at.Elem)
	}
	return true
}

func (s *Sema) validateOldStyleImplicitIntParams(node *entity.AstNode) {
	names := functionDeclaratorIdentifierList(node.Children[1])
	if len(names) == 0 {
		return
	}
	declared := map[string]bool{}
	if node.ReducedBy(parser.FunctionDefinition, 2) {
		collectOldStyleParameterDeclNames(node.Children[2], declared)
	}
	for _, name := range names {
		if !declared[name] {
			s.report(InvalidTypeSpec(node.SourceStart, "old-style function parameter defaults to int"))
		}
	}
}

func (s *Sema) validateOldStyleDefinitionPrototype(pos entity.SourcePos, paramTypes []Type, sym *Symbol) {
	if len(paramTypes) == 0 || sym == nil {
		return
	}
	prev, ok := unqual(sym.T).(*FunctionType)
	if !ok || !prev.HasProto {
		return
	}
	if !compatiblePrototypeWithOldStyleParams(prev, paramTypes) {
		s.report(InvalidTypeSpec(pos, "old-style definition parameter type does not match prototype"))
	}
}

func compatiblePrototypeWithOldStyleParams(proto *FunctionType, params []Type) bool {
	if proto == nil || !proto.HasProto || proto.Variadic || len(params) != len(proto.Params) {
		return false
	}
	for i := range params {
		if !compatibleTypeIgnoringTopLevelQualifiers(defaultPromotedType(params[i]), proto.Params[i]) {
			return false
		}
	}
	return true
}

func (s *Sema) oldStyleDefinitionParamTypes(node *entity.AstNode, names []string) []Type {
	declared := map[string]Type{}
	if node.ReducedBy(parser.FunctionDefinition, 2) {
		s.collectOldStyleParameterTypes(node.Children[2], declared)
	}
	params := make([]Type, 0, len(names))
	for _, name := range names {
		if t := declared[name]; t != nil {
			params = append(params, s.adjustParamType(t))
			continue
		}
		params = append(params, s.Types.Builtin(Int))
	}
	return params
}

func (s *Sema) collectOldStyleParameterTypes(node *entity.AstNode, out map[string]Type) {
	if node == nil {
		return
	}
	switch {
	case node.ReducedBy(parser.DeclarationList, 1):
		s.collectOldStyleDeclarationTypes(node.Children[0], out)
	case node.ReducedBy(parser.DeclarationList, 2):
		s.collectOldStyleParameterTypes(node.Children[0], out)
		s.collectOldStyleDeclarationTypes(node.Children[1], out)
	}
}

func (s *Sema) collectOldStyleDeclarationTypes(node *entity.AstNode, out map[string]Type) {
	if node == nil || !node.ReducedBy(parser.Declaration, 2) {
		return
	}
	spec := s.parseSpec(node.Children[0])
	s.collectOldStyleInitDeclaratorTypes(node.Children[1], spec.Type, out)
}

func (s *Sema) collectOldStyleInitDeclaratorTypes(node *entity.AstNode, base Type, out map[string]Type) {
	switch {
	case node.ReducedBy(parser.InitDeclaratorList, 1):
		t, name := s.applyDeclarator(node.Children[0].Children[0], base)
		if name != "" {
			out[name] = t
		}
	case node.ReducedBy(parser.InitDeclaratorList, 2):
		s.collectOldStyleInitDeclaratorTypes(node.Children[0], base, out)
		t, name := s.applyDeclarator(node.Children[2].Children[0], base)
		if name != "" {
			out[name] = t
		}
	}
}

func functionDeclaratorIdentifierList(declarator *entity.AstNode) []string {
	direct := declarator.Children[0]
	if declarator.ReducedBy(parser.Declarator, 2) {
		direct = declarator.Children[1]
	}
	for {
		switch {
		case direct.ReducedBy(parser.DirectDeclarator, 14):
			return collectIdentifierListNames(direct.Children[2])
		case direct.ReducedBy(parser.DirectDeclarator, 1), direct.ReducedBy(parser.DirectDeclarator, 12), direct.ReducedBy(parser.DirectDeclarator, 13):
			return nil
		default:
			if len(direct.Children) == 0 {
				return nil
			}
			direct = direct.Children[0]
		}
	}
}

func collectIdentifierListNames(node *entity.AstNode) []string {
	switch {
	case node.ReducedBy(parser.IdentifierList, 1):
		return []string{node.Children[0].Terminal.Lexeme}
	case node.ReducedBy(parser.IdentifierList, 2):
		out := collectIdentifierListNames(node.Children[0])
		return append(out, node.Children[2].Terminal.Lexeme)
	}
	return nil
}

func collectOldStyleParameterDeclNames(node *entity.AstNode, out map[string]bool) {
	if node == nil {
		return
	}
	if node.ReducedBy(parser.DeclarationList, 1) {
		collectDeclarationTopLevelDeclarators(node.Children[0], out)
		return
	}
	if node.ReducedBy(parser.DeclarationList, 2) {
		collectOldStyleParameterDeclNames(node.Children[0], out)
		collectDeclarationTopLevelDeclarators(node.Children[1], out)
	}
}

func collectDeclarationTopLevelDeclarators(node *entity.AstNode, out map[string]bool) {
	if node == nil || !node.ReducedBy(parser.Declaration, 2) {
		return
	}
	// K&R 参数声明只声明 identifier-list 里的顶层声明符；嵌套函数原型参数名
	// 不能反过来满足外层旧式参数的声明要求。
	collectInitDeclaratorListTopLevelNames(node.Children[1], out)
}

func collectInitDeclaratorListTopLevelNames(node *entity.AstNode, out map[string]bool) {
	switch {
	case node.ReducedBy(parser.InitDeclaratorList, 1):
		if name := topLevelDeclaratorName(node.Children[0].Children[0]); name != "" {
			out[name] = true
		}
	case node.ReducedBy(parser.InitDeclaratorList, 2):
		collectInitDeclaratorListTopLevelNames(node.Children[0], out)
		if name := topLevelDeclaratorName(node.Children[2].Children[0]); name != "" {
			out[name] = true
		}
	}
}

func topLevelDeclaratorName(node *entity.AstNode) string {
	if node == nil {
		return ""
	}
	switch {
	case node.ReducedBy(parser.Declarator, 1):
		return topLevelDirectDeclaratorName(node.Children[0])
	case node.ReducedBy(parser.Declarator, 2):
		return topLevelDirectDeclaratorName(node.Children[1])
	}
	return ""
}

func topLevelDirectDeclaratorName(node *entity.AstNode) string {
	if node == nil {
		return ""
	}
	switch {
	case node.ReducedBy(parser.DirectDeclarator, 1):
		return node.Children[0].Terminal.Lexeme
	case node.ReducedBy(parser.DirectDeclarator, 2):
		return topLevelDeclaratorName(node.Children[1])
	default:
		if len(node.Children) == 0 {
			return ""
		}
		return topLevelDirectDeclaratorName(node.Children[0])
	}
}

func (s *Sema) validateStaticFunctionDefinitions(prog *Program) {
	for _, d := range prog.Globals {
		fd, ok := d.(*FuncDecl)
		if !ok || fd.Storage != StorageStatic || fd.Sym == nil || !fd.Sym.Used {
			continue
		}
		hasDefinition := false
		for _, def := range fd.Sym.Defs {
			if _, ok := def.(*FuncDef); ok {
				hasDefinition = true
				break
			}
		}
		if !hasDefinition {
			s.report(InvalidTypeSpec(fd.Range.SourceStart, "static function used but never defined"))
		}
	}
}

func (s *Sema) markStaticFunctionUsesInGlobals(prog *Program) {
	for _, d := range prog.Globals {
		s.markStaticFunctionUsesInDecl(d)
	}
}

func (s *Sema) markStaticFunctionUsesInStmt(stmt Stmt) {
	switch x := stmt.(type) {
	case *Block:
		for _, it := range x.Items {
			s.markStaticFunctionUsesInStmt(it)
		}
	case *IfStmt:
		s.markStaticFunctionUsesInExpr(x.Cond)
		s.markStaticFunctionUsesInStmt(x.Then)
		s.markStaticFunctionUsesInStmt(x.Else)
	case *WhileStmt:
		s.markStaticFunctionUsesInExpr(x.Cond)
		s.markStaticFunctionUsesInStmt(x.Body)
	case *ForStmt:
		s.markStaticFunctionUsesInStmt(x.Init)
		s.markStaticFunctionUsesInExpr(x.Cond)
		s.markStaticFunctionUsesInExpr(x.Post)
		s.markStaticFunctionUsesInStmt(x.Body)
	case *SwitchStmt:
		s.markStaticFunctionUsesInExpr(x.Cond)
		s.markStaticFunctionUsesInStmt(x.Body)
	case *CaseStmt:
		s.markStaticFunctionUsesInStmt(x.Body)
	case *DefaultStmt:
		s.markStaticFunctionUsesInStmt(x.Body)
	case *LabeledStmt:
		s.markStaticFunctionUsesInStmt(x.Body)
	case *ExprStmt:
		s.markStaticFunctionUsesInExpr(x.Expr)
	case *ReturnStmt:
		s.markStaticFunctionUsesInExpr(x.Value)
	case *DeclStmt:
		for _, d := range x.Decls {
			s.markStaticFunctionUsesInDecl(d)
		}
	}
}

func (s *Sema) markStaticFunctionUsesInDecl(decl Decl) {
	switch x := decl.(type) {
	case *VarDecl:
		s.markStaticFunctionUsesInType(x.T)
		s.markStaticFunctionUsesInExpr(x.Init)
	case *TypedefDecl:
		s.markStaticFunctionUsesInType(x.T)
	}
}

func (s *Sema) markStaticFunctionUsesInExpr(expr Expr) {
	switch x := expr.(type) {
	case nil:
		return
	case *VarRef:
		if x.Sym != nil && x.Sym.Kind == SymFunc {
			x.Sym.Used = true
		}
	case *BinOp:
		s.markStaticFunctionUsesInExpr(x.L)
		s.markStaticFunctionUsesInExpr(x.R)
	case *UnOp:
		s.markStaticFunctionUsesInExpr(x.X)
	case *AssignExpr:
		s.markStaticFunctionUsesInExpr(x.L)
		s.markStaticFunctionUsesInExpr(x.R)
	case *CompoundAssign:
		s.markStaticFunctionUsesInExpr(x.L)
		s.markStaticFunctionUsesInExpr(x.R)
	case *CallExpr:
		s.markStaticFunctionUsesInExpr(x.Callee)
		for _, arg := range x.Args {
			s.markStaticFunctionUsesInExpr(arg)
		}
	case *MemberExpr:
		s.markStaticFunctionUsesInExpr(x.Base)
	case *IndexExpr:
		s.markStaticFunctionUsesInExpr(x.Base)
		s.markStaticFunctionUsesInExpr(x.Index)
	case *CondExpr:
		s.markStaticFunctionUsesInExpr(x.Cond)
		s.markStaticFunctionUsesInExpr(x.Then)
		s.markStaticFunctionUsesInExpr(x.Else)
	case *SizeofExpr:
		// sizeof 只有在操作数是实际求值的 VLA 时才算使用其中的静态函数。
		if x.Operand.Type != nil {
			s.markStaticFunctionUsesInVariablySizedType(x.Operand.Type)
		}
		if x.Operand.Expr != nil && typeHasVariableSize(x.Operand.Expr.GetType()) {
			s.markStaticFunctionUsesInExpr(x.Operand.Expr)
		}
	case *CommaExpr:
		s.markStaticFunctionUsesInExpr(x.L)
		s.markStaticFunctionUsesInExpr(x.R)
	case *CompoundLit:
		s.markStaticFunctionUsesInType(x.T)
		s.markStaticFunctionUsesInInitList(x.Init)
	case *InitList:
		s.markStaticFunctionUsesInInitList(x)
	case *ImplicitCast:
		s.markStaticFunctionUsesInExpr(x.X)
	case *ExplicitCast:
		s.markStaticFunctionUsesInType(x.To)
		s.markStaticFunctionUsesInExpr(x.X)
	case *AddrConst:
		if x.Sym != nil && x.Sym.Kind == SymFunc {
			x.Sym.Used = true
		}
	}
}

func (s *Sema) markStaticFunctionUsesInInitList(il *InitList) {
	if il == nil {
		return
	}
	for _, elem := range il.Elems {
		s.markStaticFunctionUsesInExpr(elem.Value)
	}
}

func (s *Sema) markStaticFunctionUsesInType(t Type) {
	s.markStaticFunctionUsesInTypeSeen(t, map[Type]bool{}, true)
}

func (s *Sema) markStaticFunctionUsesInVariablySizedType(t Type) {
	if !typeHasVariableSize(t) {
		return
	}
	s.markStaticFunctionUsesInTypeSeen(t, map[Type]bool{}, false)
}

func (s *Sema) markStaticFunctionUsesInTypeSeen(t Type, seen map[Type]bool, throughPointer bool) {
	t = unqual(t)
	if t == nil || seen[t] {
		return
	}
	seen[t] = true
	switch x := t.(type) {
	case *ArrayType:
		if x.SizeKind == ArrayVLA || x.SizeKind == ArrayStarSize {
			if x.SizeExpr != nil {
				s.markStaticFunctionUsesInExpr(x.SizeExpr)
			}
		}
		s.markStaticFunctionUsesInTypeSeen(x.Elem, seen, throughPointer)
	case *PointerType:
		if throughPointer {
			s.markStaticFunctionUsesInTypeSeen(x.Pointee, seen, throughPointer)
		}
	case *FunctionType:
		s.markStaticFunctionUsesInTypeSeen(x.Ret, seen, throughPointer)
		for _, p := range x.Params {
			s.markStaticFunctionUsesInTypeSeen(p, seen, throughPointer)
		}
	case *StructType:
		for _, f := range x.Fields {
			s.markStaticFunctionUsesInTypeSeen(f.T, seen, throughPointer)
		}
	case *UnionType:
		for _, f := range x.Fields {
			s.markStaticFunctionUsesInTypeSeen(f.T, seen, throughPointer)
		}
	}
}

func (s *Sema) foldStaticInitializers(prog *Program) {
	ev := NewEvaluator(s)
	for _, d := range prog.Globals {
		vd, ok := d.(*VarDecl)
		if !ok {
			continue
		}
		ev.allowCompoundLiteralAddress = true
		s.foldStaticVarInitializer(ev, vd)
	}
	// 块作用域 static 也具有静态存储期，初始化式必须走同一套常量表达式约束。
	for _, fn := range prog.Funcs {
		for _, vd := range fn.Locals {
			if vd.Storage == StorageStatic {
				ev.allowCompoundLiteralAddress = false
				s.foldStaticVarInitializer(ev, vd)
			}
		}
	}
}

func (s *Sema) foldStaticVarInitializer(ev *Evaluator, vd *VarDecl) {
	if vd.Init == nil {
		return
	}
	folded, ok := s.foldStaticInitializerExpr(ev, vd.Init)
	if !ok {
		return
	}
	vd.Init = folded
}

func (s *Sema) foldStaticInitializerExpr(ev *Evaluator, expr Expr) (Expr, bool) {
	if il, ok := expr.(*InitList); ok {
		ok := true
		// braced initializer 只是语法分组；静态存储期的每个叶子表达式仍须满足常量表达式约束。
		for i := range il.Elems {
			folded, elemOK := s.foldStaticInitializerExpr(ev, il.Elems[i].Value)
			if !elemOK {
				ok = false
				continue
			}
			il.Elems[i].Value = folded
		}
		return il, ok
	}
	cv, ok := ev.EvalConstant(expr)
	if !ok {
		s.report(InvalidTypeSpec(expr.Pos().SourceStart, "static initializer must be constant"))
		return expr, false
	}
	if folded := constToExpr(cv, s); folded != nil {
		return folded, true
	}
	return expr, true
}

func constToExpr(cv ConstValue, s *Sema) Expr {
	switch cv.Kind {
	case ConstInt, ConstUint:
		return &IntLit{Value: cv.Int, T: cv.T}
	case ConstFloat:
		return &FloatLit{Value: cv.Float, T: cv.T}
	case ConstAddress:
		if cv.Addr.Sym == nil {
			return nil
		}
		return &AddrConst{Sym: cv.Addr.Sym, Offset: cv.Addr.Offset, T: cv.T}
	}
	return nil
}
