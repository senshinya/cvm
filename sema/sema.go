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

	pendingFuncs []*pendingFunc
}

type SemaOptions struct {
	PedanticErrors bool
}

type pendingFunc struct {
	def     *FuncDef
	bodyAst *entity.AstNode
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
	for _, pf := range s.pendingFuncs {
		s.walkFunctionBody(pf, prog)
	}
	s.foldStaticInitializers(prog)
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
	}
}

func (s *Sema) walkFunctionDefinition(node *entity.AstNode, prog *Program) {
	spec := s.parseSpec(node.Children[0])
	t, name := s.applyDeclarator(node.Children[1], spec.Type)
	ft, ok := t.(*FunctionType)
	if !ok {
		s.report(InvalidTypeSpec(node.SourceStart, "function definition declarator did not yield a function type"))
		return
	}
	sym := s.scope.LookupCurrent(name, NSOrdinary)
	if sym == nil {
		sym = &Symbol{Name: name, Kind: SymFunc, T: ft, Storage: spec.Storage, Linkage: LinkageExternal, Pos: node.Children[1].SourceStart}
		s.scope.Insert(name, sym)
	} else if sym.Kind != SymFunc {
		s.report(RedefinitionSymbol(node.Children[1].SourceStart, sym.Pos, name))
		return
	}
	def := &FuncDef{Sym: sym, T: ft, Range: node.SourceRange, Labels: map[string]*LabeledStmt{}}
	def.Params = s.collectParamDecls(node.Children[1], ft)
	sym.Decl = def
	sym.Defs = append(sym.Defs, def)
	prog.Funcs = append(prog.Funcs, def)
	s.pendingFuncs = append(s.pendingFuncs, &pendingFunc{def: def, bodyAst: node.Children[len(node.Children)-1]})
}

func (s *Sema) walkDeclaration(node *entity.AstNode, prog *Program) {
	if node.ReducedBy(parser.Declaration, 3) {
		s.typeStaticAssert(node.Children[0])
		return
	}
	spec := s.parseSpec(node.Children[0])
	if node.ReducedBy(parser.Declaration, 1) {
		if s.Options.PedanticErrors && hasEnumReferenceSpecifier(node.Children[0]) {
			s.report(InvalidTypeSpec(node.SourceStart, "ISO C forbids forward references to enum types"))
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
	t, name := s.applyDeclarator(node.Children[0], spec.Type)
	pos := node.Children[0].SourceStart
	if name == "" {
		s.report(InvalidTypeSpec(pos, "missing declarator name"))
		return
	}
	if spec.IsTypedef {
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
	if ft, ok := t.(*FunctionType); ok {
		if node.ReducedBy(parser.InitDeclarator, 2) {
			s.report(InvalidTypeSpec(pos, "function declarator cannot have initializer"))
			return
		}
		s.declareFunction(name, ft, spec.Storage, pos, srcRange, prog)
		return
	}
	storage := spec.Storage
	sym := &Symbol{Name: name, Kind: SymVar, T: t, Storage: storage, Pos: pos}
	if s.scope.Kind == ScopeFile {
		sym.Linkage = LinkageExternal
		if storage == StorageStatic {
			sym.Linkage = LinkageInternal
		}
	}
	vd := &VarDecl{Sym: sym, T: t, Storage: storage, Range: srcRange}
	sym.Decl = vd
	if node.ReducedBy(parser.InitDeclarator, 2) {
		vd.Init = s.typeInitializer(node.Children[2], t)
	}
	if err := s.scope.InsertChecked(name, sym); err != nil {
		s.report(err.(*common.CvmError))
		return
	}
	prog.Globals = append(prog.Globals, vd)
}

func (s *Sema) declareFunction(name string, ft *FunctionType, storage StorageClass, pos entity.SourcePos, srcRange entity.SourceRange, prog *Program) {
	sym := s.scope.LookupCurrent(name, NSOrdinary)
	if sym == nil {
		sym = &Symbol{Name: name, Kind: SymFunc, T: ft, Storage: storage, Pos: pos, Linkage: LinkageExternal}
		s.scope.Insert(name, sym)
	} else if sym.Kind != SymFunc {
		s.report(RedefinitionSymbol(pos, sym.Pos, name))
		return
	}
	fd := &FuncDecl{Sym: sym, T: ft, Storage: storage, Range: srcRange}
	sym.Decl = fd
	sym.Defs = append(sym.Defs, fd)
	prog.Globals = append(prog.Globals, fd)
}

func (s *Sema) walkFunctionBody(pf *pendingFunc, prog *Program) {
	bodyScope := NewScope(ScopeFunc, s.SymTab.File)
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
	_ = bodyScope.InsertChecked("__func__", funcSym)
	prev := s.scope
	s.scope = bodyScope
	defer func() { s.scope = prev }()
	ctx := &funcCtx{def: pf.def}
	body, _ := s.typeStmt(pf.bodyAst, bodyScope, ctx).(*Block)
	if body == nil {
		body = &Block{Range: pf.bodyAst.SourceRange, Scope: bodyScope}
	}
	pf.def.Body = body
	pf.def.Labels = map[string]*LabeledStmt{}
	collectLabels(body, pf.def.Labels)
	resolveGotos(ctx.pendingGotos, pf.def.Labels, s)
	_ = prog
}

func (s *Sema) foldStaticInitializers(prog *Program) {
	ev := NewEvaluator(s)
	for _, d := range prog.Globals {
		vd, ok := d.(*VarDecl)
		if !ok || vd.Init == nil {
			continue
		}
		if _, ok := vd.Init.(*InitList); ok {
			continue
		}
		cv, ok := ev.EvalConstant(vd.Init)
		if !ok {
			s.report(InvalidTypeSpec(vd.Range.SourceStart, "static initializer must be constant"))
			continue
		}
		if folded := constToExpr(cv, s); folded != nil {
			vd.Init = folded
		}
	}
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
