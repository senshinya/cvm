package sema

import (
	"shinya.click/cvm/common"
	"shinya.click/cvm/entity"
	"shinya.click/cvm/parser"
)

type funcCtx struct {
	def           *FuncDef
	prog          *Program
	loopDepth     int
	switchStack   []*SwitchStmt
	namedBreak    []string
	namedContinue []string
	pendingGotos  []*GotoStmt
	vmScopes      []vmScopeBarrier
	order         int
	seenStatement bool
}

type vmScopeBarrier struct {
	decl      entity.SourceRange
	scope     entity.SourceRange
	declOrder int
}

func (ctx *funcCtx) nextOrder() int {
	if ctx == nil {
		return 0
	}
	ctx.order++
	return ctx.order
}

func (s *Sema) typeStmt(node *entity.AstNode, scope *Scope, ctx *funcCtx) Stmt {
	if node == nil {
		return &EmptyStmt{}
	}
	for len(node.Children) == 1 && node.Terminal == nil && !isMeaningfulSingleChildStmt(node.Typ) {
		node = node.Children[0]
	}
	switch node.Typ {
	case parser.Statement:
		return s.typeStmt(node.Children[0], scope, ctx)
	case parser.CompoundStatement:
		return s.typeBlock(node, scope, ctx)
	case parser.ExpressionStatement:
		return s.typeExprStmt(node, scope)
	case parser.SelectionStatement:
		return s.typeSelection(node, scope, ctx)
	case parser.IterationStatement:
		return s.typeIteration(node, scope, ctx)
	case parser.JumpStatement:
		return s.typeJump(node, scope, ctx)
	case parser.LabeledStatement:
		return s.typeLabeled(node, scope, ctx)
	}
	return &EmptyStmt{Range: node.SourceRange}
}

func isMeaningfulSingleChildStmt(t entity.TokenType) bool {
	switch t {
	case parser.Statement, parser.CompoundStatement, parser.ExpressionStatement, parser.SelectionStatement,
		parser.IterationStatement, parser.JumpStatement, parser.LabeledStatement:
		return true
	}
	return false
}

func (s *Sema) typeExprStmt(node *entity.AstNode, scope *Scope) Stmt {
	switch {
	case node.ReducedBy(parser.ExpressionStatement, 1):
		return &EmptyStmt{Range: node.SourceRange}
	case node.ReducedBy(parser.ExpressionStatement, 2):
		return &ExprStmt{Expr: s.typeExpr(node.Children[0], scope), Range: node.SourceRange}
	}
	return &EmptyStmt{Range: node.SourceRange}
}

func (s *Sema) typeBlock(node *entity.AstNode, parent *Scope, ctx *funcCtx) *Block {
	scope := NewScope(ScopeBlock, parent)
	scope.Range = node.SourceRange
	prev := s.scope
	s.scope = scope
	defer func() { s.scope = prev }()
	block := &Block{Range: node.SourceRange, Scope: scope}
	if node.ReducedBy(parser.CompoundStatement, 2) {
		s.collectBlockItems(node.Children[1], scope, ctx, &block.Items)
	}
	return block
}

func (s *Sema) typeScopedStmt(node *entity.AstNode, parent *Scope, ctx *funcCtx) Stmt {
	scope := NewScope(ScopeBlock, parent)
	scope.Range = node.SourceRange
	prev := s.scope
	s.scope = scope
	defer func() { s.scope = prev }()
	return s.typeStmt(node, scope, ctx)
}

func (s *Sema) collectBlockItems(node *entity.AstNode, scope *Scope, ctx *funcCtx, out *[]Stmt) {
	switch {
	case node.ReducedBy(parser.BlockItemList, 1):
		s.appendBlockItem(node.Children[0], scope, ctx, out)
	case node.ReducedBy(parser.BlockItemList, 2):
		s.collectBlockItems(node.Children[0], scope, ctx, out)
		s.appendBlockItem(node.Children[1], scope, ctx, out)
	}
}

func (s *Sema) appendBlockItem(node *entity.AstNode, scope *Scope, ctx *funcCtx, out *[]Stmt) {
	switch {
	case node.ReducedBy(parser.BlockItem, 1):
		if ctx != nil && ctx.seenStatement && s.Options.WErrorDeclarationAfterStatement {
			s.report(InvalidTypeSpec(node.SourceStart, "declaration after statement"))
		}
		var decls []Decl
		s.walkBlockDecl(node.Children[0], scope, ctx, &decls)
		if len(decls) > 0 {
			*out = append(*out, &DeclStmt{Decls: decls, Range: node.SourceRange})
		}
	case node.ReducedBy(parser.BlockItem, 2):
		if ctx != nil {
			ctx.seenStatement = true
		}
		*out = append(*out, s.typeStmt(node.Children[0], scope, ctx))
	case node.ReducedBy(parser.BlockItem, 3):
		if ctx != nil && ctx.prog != nil {
			s.walkFunctionDefinition(node.Children[0], ctx.prog)
		}
	}
}

func (s *Sema) walkBlockDecl(node *entity.AstNode, scope *Scope, ctx *funcCtx, out *[]Decl) {
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
			*out = append(*out, &TagDecl{T: spec.Type, Range: node.SourceRange})
		}
		return
	}
	s.walkBlockInitDeclList(node.Children[1], spec, scope, ctx, out, node.SourceRange)
}

func (s *Sema) walkBlockInitDeclList(node *entity.AstNode, spec SpecResult, scope *Scope, ctx *funcCtx, out *[]Decl, srcRange entity.SourceRange) {
	switch {
	case node.ReducedBy(parser.InitDeclaratorList, 1):
		if d := s.walkBlockInitDecl(node.Children[0], spec, scope, ctx, srcRange); d != nil {
			*out = append(*out, d)
		}
	case node.ReducedBy(parser.InitDeclaratorList, 2):
		s.walkBlockInitDeclList(node.Children[0], spec, scope, ctx, out, srcRange)
		if d := s.walkBlockInitDecl(node.Children[2], spec, scope, ctx, srcRange); d != nil {
			*out = append(*out, d)
		}
	}
}

func (s *Sema) walkConditionDecl(node *entity.AstNode, scope *Scope, ctx *funcCtx, out *[]Decl) {
	if !node.ReducedBy(parser.ConditionDeclaration, 1) {
		return
	}
	spec := s.parseSpec(node.Children[0])
	s.walkBlockInitDeclList(node.Children[1], spec, scope, ctx, out, node.SourceRange)
}

func (s *Sema) walkBlockInitDecl(node *entity.AstNode, spec SpecResult, scope *Scope, ctx *funcCtx, srcRange entity.SourceRange) Decl {
	s.validateDeclaratorArrayQualifiers(node.Children[0], false)
	t, name := s.applyDeclarator(node.Children[0], spec.Type)
	pos := node.Children[0].SourceStart
	s.validateInlineSpecifier(spec, t, name, pos, false)
	s.validateRestrictType(t, pos)
	if spec.IsTypedef {
		markTypedefVMBounds(t)
		sym := &Symbol{Name: name, Kind: SymTypedef, T: t, Storage: StorageTypedef, Pos: pos}
		td := &TypedefDecl{Sym: sym, T: t, Range: srcRange}
		sym.Decl = td
		if err := scope.InsertChecked(name, sym); err != nil {
			s.report(err.(*common.CvmError))
		}
		recordVMScopeBarrier(ctx, scope, srcRange, t)
		return td
	}
	storage := spec.Storage
	if ft, ok := unqual(t).(*FunctionType); ok {
		if node.ReducedBy(parser.InitDeclarator, 2) {
			s.report(InvalidTypeSpec(pos, "function declarator cannot have initializer"))
			return nil
		}
		return s.declareBlockFunction(name, ft, storage, pos, srcRange, scope, ctx)
	}
	if storage == StorageNone {
		storage = StorageAuto
	}
	if storage == StorageStatic && typeHasDisallowedStaticArrayBound(t) {
		s.report(InvalidTypeSpec(pos, "array size must be integer constant expression"))
	}
	sym := &Symbol{Name: name, Kind: SymVar, T: t, Storage: storage, Pos: pos}
	vd := &VarDecl{Sym: sym, T: t, Storage: storage, Range: srcRange}
	sym.Decl = vd
	if err := scope.InsertChecked(name, sym); err != nil {
		s.report(err.(*common.CvmError))
	}
	if node.ReducedBy(parser.InitDeclarator, 2) {
		vd.Init = s.typeInitializer(node.Children[2], t)
	}
	if ctx != nil && ctx.def != nil {
		ctx.def.Locals = append(ctx.def.Locals, vd)
	}
	recordVMScopeBarrier(ctx, scope, srcRange, t)
	return vd
}

func (s *Sema) declareBlockFunction(name string, ft *FunctionType, storage StorageClass, pos entity.SourcePos, srcRange entity.SourceRange, scope *Scope, ctx *funcCtx) Decl {
	s.validateFunctionVMReturn(ft, pos)
	if storage != StorageNone && storage != StorageExtern {
		s.report(InvalidTypeSpec(pos, "block-scope function declaration must be extern"))
		return nil
	}
	fileSym := s.SymTab.File.LookupCurrent(name, NSOrdinary)
	if fileSym == nil {
		fileSym = &Symbol{Name: name, Kind: SymFunc, T: ft, Storage: StorageExtern, Linkage: LinkageExternal, Pos: pos}
		s.SymTab.File.Insert(name, fileSym)
	} else if fileSym.Kind != SymFunc {
		s.report(RedefinitionSymbol(pos, fileSym.Pos, name))
		return nil
	} else if !s.mergeFunctionDeclaration(fileSym, ft, pos) {
		return nil
	}
	if cur := scope.LookupCurrent(name, NSOrdinary); cur != nil && cur.Kind != SymFunc {
		s.report(RedefinitionSymbol(pos, cur.Pos, name))
		return nil
	}
	scope.Insert(name, fileSym)
	fd := &FuncDecl{Sym: fileSym, T: ft, Storage: StorageExtern, Range: srcRange}
	if fileSym.Decl == nil {
		fileSym.Decl = fd
	}
	fileSym.Defs = append(fileSym.Defs, fd)
	if ctx != nil && ctx.prog != nil {
		ctx.prog.Globals = append(ctx.prog.Globals, fd)
	}
	return fd
}

func (s *Sema) typeSelection(node *entity.AstNode, scope *Scope, ctx *funcCtx) Stmt {
	switch {
	case node.ReducedBy(parser.SelectionStatement, 1):
		stmtScope := NewScope(ScopeBlock, scope)
		stmtScope.Range = node.SourceRange
		prev := s.scope
		s.scope = stmtScope
		defer func() { s.scope = prev }()
		cond := s.castBoolConversion(s.castLValueToRValue(s.typeExpr(node.Children[2], stmtScope)))
		return &IfStmt{Cond: cond, Then: s.typeScopedStmt(node.Children[4], stmtScope, ctx), Range: node.SourceRange}
	case node.ReducedBy(parser.SelectionStatement, 2):
		stmtScope := NewScope(ScopeBlock, scope)
		stmtScope.Range = node.SourceRange
		prev := s.scope
		s.scope = stmtScope
		defer func() { s.scope = prev }()
		cond := s.castBoolConversion(s.castLValueToRValue(s.typeExpr(node.Children[2], stmtScope)))
		return &IfStmt{Cond: cond, Then: s.typeScopedStmt(node.Children[4], stmtScope, ctx), Else: s.typeScopedStmt(node.Children[6], stmtScope, ctx), Range: node.SourceRange}
	case node.ReducedBy(parser.SelectionStatement, 3):
		stmtScope := NewScope(ScopeBlock, scope)
		stmtScope.Range = node.SourceRange
		prev := s.scope
		s.scope = stmtScope
		defer func() { s.scope = prev }()
		cond := s.castIntegerPromotion(s.castLValueToRValue(s.typeExpr(node.Children[2], stmtScope)))
		sw := &SwitchStmt{Cond: cond, Order: ctx.nextOrder(), Range: node.SourceRange}
		ctx.switchStack = append(ctx.switchStack, sw)
		sw.Body = s.typeScopedStmt(node.Children[4], stmtScope, ctx)
		ctx.switchStack = ctx.switchStack[:len(ctx.switchStack)-1]
		collectCasesAndDefault(sw.Body, sw, s)
		validateSwitchVMJumps(sw, ctx.vmScopes, s)
		return sw
	case node.ReducedBy(parser.SelectionStatement, 4),
		node.ReducedBy(parser.SelectionStatement, 5),
		node.ReducedBy(parser.SelectionStatement, 6),
		node.ReducedBy(parser.SelectionStatement, 7):
		declScope := NewScope(ScopeBlock, scope)
		declScope.Range = node.SourceRange
		prev := s.scope
		s.scope = declScope
		defer func() { s.scope = prev }()
		var decls []Decl
		s.walkConditionDecl(node.Children[2], declScope, ctx, &decls)
		stmtIdx := 4
		cond := Expr(&IntLit{Value: 1, T: s.Types.Builtin(Int), Range: node.Children[2].SourceRange})
		if node.ReducedBy(parser.SelectionStatement, 5) || node.ReducedBy(parser.SelectionStatement, 7) {
			cond = s.castBoolConversion(s.castLValueToRValue(s.typeExpr(node.Children[4], declScope)))
			stmtIdx = 6
		}
		stmt := &IfStmt{Cond: cond, Then: s.typeScopedStmt(node.Children[stmtIdx], declScope, ctx), Range: node.SourceRange}
		if node.ReducedBy(parser.SelectionStatement, 6) || node.ReducedBy(parser.SelectionStatement, 7) {
			stmt.Else = s.typeScopedStmt(node.Children[len(node.Children)-1], declScope, ctx)
		}
		return stmt
	case node.ReducedBy(parser.SelectionStatement, 8),
		node.ReducedBy(parser.SelectionStatement, 9):
		declScope := NewScope(ScopeBlock, scope)
		declScope.Range = node.SourceRange
		prev := s.scope
		s.scope = declScope
		defer func() { s.scope = prev }()
		var decls []Decl
		s.walkConditionDecl(node.Children[2], declScope, ctx, &decls)
		stmtIdx := 4
		cond := Expr(&IntLit{Value: 0, T: s.Types.Builtin(Int), Range: node.Children[2].SourceRange})
		if node.ReducedBy(parser.SelectionStatement, 9) {
			cond = s.castIntegerPromotion(s.castLValueToRValue(s.typeExpr(node.Children[4], declScope)))
			stmtIdx = 6
		}
		sw := &SwitchStmt{Cond: cond, Order: ctx.nextOrder(), Range: node.SourceRange}
		ctx.switchStack = append(ctx.switchStack, sw)
		sw.Body = s.typeScopedStmt(node.Children[stmtIdx], declScope, ctx)
		ctx.switchStack = ctx.switchStack[:len(ctx.switchStack)-1]
		collectCasesAndDefault(sw.Body, sw, s)
		validateSwitchVMJumps(sw, ctx.vmScopes, s)
		return sw
	}
	return &EmptyStmt{Range: node.SourceRange}
}

func collectCasesAndDefault(stmt Stmt, sw *SwitchStmt, s *Sema) {
	switch x := stmt.(type) {
	case *Block:
		for _, it := range x.Items {
			collectCasesAndDefault(it, sw, s)
		}
	case *CaseStmt:
		for _, prev := range sw.Cases {
			if prev.Value == x.Value {
				s.report(InvalidTypeSpec(x.Range.SourceStart, "duplicate case value"))
			}
		}
		sw.Cases = append(sw.Cases, x)
		collectCasesAndDefault(x.Body, sw, s)
	case *DefaultStmt:
		if sw.Default != nil {
			s.report(InvalidTypeSpec(x.Range.SourceStart, "multiple default labels"))
		}
		sw.Default = x
		collectCasesAndDefault(x.Body, sw, s)
	case *LabeledStmt:
		collectCasesAndDefault(x.Body, sw, s)
	case *IfStmt:
		collectCasesAndDefault(x.Then, sw, s)
		collectCasesAndDefault(x.Else, sw, s)
	case *WhileStmt:
		collectCasesAndDefault(x.Body, sw, s)
	case *ForStmt:
		collectCasesAndDefault(x.Body, sw, s)
	case *SwitchStmt:
		// 嵌套 switch 的 case/default 归内层 switch 管理；外层只需要看到
		// 自己语句树里的标签，即使它们藏在循环或 if 里面。
		return
	}
}

func (s *Sema) typeLabeled(node *entity.AstNode, scope *Scope, ctx *funcCtx) Stmt {
	switch {
	case node.ReducedBy(parser.LabeledStatement, 1):
		order := ctx.nextOrder()
		name := node.Children[0].Terminal.Lexeme
		target := namedLabelTargetKind(node.Children[2])
		if target == namedLabelLoop || target == namedLabelSwitch {
			ctx.namedBreak = append(ctx.namedBreak, name)
			defer func() { ctx.namedBreak = ctx.namedBreak[:len(ctx.namedBreak)-1] }()
		}
		if target == namedLabelLoop {
			ctx.namedContinue = append(ctx.namedContinue, name)
			defer func() { ctx.namedContinue = ctx.namedContinue[:len(ctx.namedContinue)-1] }()
		}
		return &LabeledStmt{Name: name, Body: s.typeStmt(node.Children[2], scope, ctx), Order: order, Range: node.SourceRange}
	case node.ReducedBy(parser.LabeledStatement, 2):
		expr := s.typeExpr(node.Children[1], scope)
		cv, ok := NewEvaluator(s).EvalC99IntegerConstantExpression(expr)
		if !ok {
			s.report(InvalidTypeSpec(node.SourceStart, "case value must be integer constant expression"))
		}
		return &CaseStmt{Value: cv.Int, Body: s.typeStmt(node.Children[3], scope, ctx), Order: ctx.nextOrder(), Range: node.SourceRange}
	case node.ReducedBy(parser.LabeledStatement, 3):
		return &DefaultStmt{Body: s.typeStmt(node.Children[2], scope, ctx), Order: ctx.nextOrder(), Range: node.SourceRange}
	}
	return &EmptyStmt{Range: node.SourceRange}
}

type namedLabelTarget int

const (
	namedLabelNone namedLabelTarget = iota
	namedLabelLoop
	namedLabelSwitch
)

func namedLabelTargetKind(node *entity.AstNode) namedLabelTarget {
	for node != nil && node.Typ == parser.Statement {
		node = node.Children[0]
	}
	if node == nil {
		return namedLabelNone
	}
	if node.ReducedBy(parser.LabeledStatement, 1) {
		return namedLabelTargetKind(node.Children[2])
	}
	if node.ReducedBy(parser.LabeledStatement, 2) || node.ReducedBy(parser.LabeledStatement, 3) {
		return namedLabelSwitch
	}
	if node.Typ == parser.IterationStatement {
		return namedLabelLoop
	}
	if node.ReducedBy(parser.SelectionStatement, 3) ||
		node.ReducedBy(parser.SelectionStatement, 8) ||
		node.ReducedBy(parser.SelectionStatement, 9) {
		return namedLabelSwitch
	}
	return namedLabelNone
}

func (s *Sema) typeIteration(node *entity.AstNode, scope *Scope, ctx *funcCtx) Stmt {
	switch {
	case node.ReducedBy(parser.IterationStatement, 1):
		stmtScope := NewScope(ScopeBlock, scope)
		stmtScope.Range = node.SourceRange
		prev := s.scope
		s.scope = stmtScope
		defer func() { s.scope = prev }()
		cond := s.castBoolConversion(s.castLValueToRValue(s.typeExpr(node.Children[2], stmtScope)))
		ctx.loopDepth++
		body := s.typeScopedStmt(node.Children[4], stmtScope, ctx)
		ctx.loopDepth--
		return &WhileStmt{Cond: cond, Body: body, Range: node.SourceRange}
	case node.ReducedBy(parser.IterationStatement, 2):
		stmtScope := NewScope(ScopeBlock, scope)
		stmtScope.Range = node.SourceRange
		prev := s.scope
		s.scope = stmtScope
		defer func() { s.scope = prev }()
		ctx.loopDepth++
		body := s.typeScopedStmt(node.Children[1], stmtScope, ctx)
		ctx.loopDepth--
		cond := s.castBoolConversion(s.castLValueToRValue(s.typeExpr(node.Children[4], stmtScope)))
		return &WhileStmt{Cond: cond, Body: body, DoWhile: true, Range: node.SourceRange}
	}
	forScope := NewScope(ScopeBlock, scope)
	forScope.Range = node.SourceRange
	prev := s.scope
	s.scope = forScope
	defer func() { s.scope = prev }()
	fp := s.collectForParts(node, forScope, ctx)
	fs := &ForStmt{Init: fp.init, Cond: fp.cond, Post: fp.post, Scope: forScope, Range: node.SourceRange}
	ctx.loopDepth++
	fs.Body = s.typeScopedStmt(fp.body, forScope, ctx)
	ctx.loopDepth--
	return fs
}

type forParts struct {
	init Stmt
	cond Expr
	post Expr
	body *entity.AstNode
}

func (s *Sema) collectForParts(node *entity.AstNode, scope *Scope, ctx *funcCtx) forParts {
	fp := forParts{body: node.Children[len(node.Children)-1]}
	if node.ReducedBy(parser.IterationStatement, 11) || node.ReducedBy(parser.IterationStatement, 12) ||
		node.ReducedBy(parser.IterationStatement, 13) || node.ReducedBy(parser.IterationStatement, 14) {
		var decls []Decl
		declNode := node.Children[2]
		s.walkBlockDecl(declNode, scope, ctx, &decls)
		s.validateForInitDeclaration(declNode, decls)
		fp.init = &DeclStmt{Decls: decls, Range: node.Children[2].SourceRange}
		slot := 0
		for i := 3; i < len(node.Children)-1; i++ {
			ch := node.Children[i]
			if ch.Typ == entity.SEMICOLON {
				slot++
				continue
			}
			if ch.Typ == parser.Expression {
				if slot == 0 {
					fp.cond = s.castBoolConversion(s.castLValueToRValue(s.typeExpr(ch, scope)))
				} else {
					fp.post = s.typeExpr(ch, scope)
				}
			}
		}
		return fp
	}
	slot := 0
	for i := 2; i < len(node.Children)-1; i++ {
		ch := node.Children[i]
		if ch.Typ == entity.SEMICOLON {
			slot++
			continue
		}
		if ch.Typ != parser.Expression {
			continue
		}
		expr := s.typeExpr(ch, scope)
		switch slot {
		case 0:
			fp.init = &ExprStmt{Expr: expr, Range: ch.SourceRange}
		case 1:
			fp.cond = s.castBoolConversion(s.castLValueToRValue(expr))
		case 2:
			fp.post = expr
		}
	}
	return fp
}

func (s *Sema) validateForInitDeclaration(node *entity.AstNode, decls []Decl) {
	if forDeclarationDefinesTagOrEnum(node) {
		s.report(InvalidTypeSpec(node.SourceStart, "tag or enumerator declaration is not allowed in for init declaration"))
	}
	for _, d := range decls {
		switch x := d.(type) {
		case *FuncDecl:
			s.report(InvalidTypeSpec(x.Range.SourceStart, "function declaration is not allowed in for init declaration"))
		case *VarDecl:
			if _, ok := unqual(x.T).(*FunctionType); ok {
				s.report(InvalidTypeSpec(x.Range.SourceStart, "function declaration is not allowed in for init declaration"))
			}
			if x.Storage == StorageStatic || x.Storage == StorageExtern {
				s.report(InvalidTypeSpec(x.Range.SourceStart, "static or extern declaration is not allowed in for init declaration"))
			}
		case *TypedefDecl:
			s.report(InvalidTypeSpec(x.Range.SourceStart, "non-variable declaration is not allowed in for init declaration"))
		case *TagDecl:
			s.report(InvalidTypeSpec(x.Range.SourceStart, "tag declaration is not allowed in for init declaration"))
		}
	}
}

func forDeclarationDefinesTagOrEnum(node *entity.AstNode) bool {
	if node == nil {
		return false
	}
	switch {
	case node.ReducedBy(parser.StructOrUnionSpecifier, 1),
		node.ReducedBy(parser.StructOrUnionSpecifier, 2),
		node.ReducedBy(parser.EnumSpecifier, 1),
		node.ReducedBy(parser.EnumSpecifier, 2),
		node.ReducedBy(parser.EnumSpecifier, 3),
		node.ReducedBy(parser.EnumSpecifier, 4):
		return true
	}
	for _, child := range node.Children {
		if forDeclarationDefinesTagOrEnum(child) {
			return true
		}
	}
	return false
}

func (s *Sema) typeJump(node *entity.AstNode, scope *Scope, ctx *funcCtx) Stmt {
	switch {
	case node.ReducedBy(parser.JumpStatement, 1):
		g := &GotoStmt{Name: node.Children[1].Terminal.Lexeme, Order: ctx.nextOrder(), Range: node.SourceRange}
		ctx.pendingGotos = append(ctx.pendingGotos, g)
		return g
	case node.ReducedBy(parser.JumpStatement, 2):
		if ctx.loopDepth == 0 {
			s.report(InvalidTypeSpec(node.SourceStart, "continue outside loop"))
		}
		return &ContinueStmt{Range: node.SourceRange}
	case node.ReducedBy(parser.JumpStatement, 3):
		if ctx.loopDepth == 0 && len(ctx.switchStack) == 0 {
			s.report(InvalidTypeSpec(node.SourceStart, "break outside loop or switch"))
		}
		return &BreakStmt{Range: node.SourceRange}
	case node.ReducedBy(parser.JumpStatement, 4):
		return &ReturnStmt{Range: node.SourceRange}
	case node.ReducedBy(parser.JumpStatement, 5):
		expr := s.castArrayDecay(s.castLValueToRValue(s.typeExpr(node.Children[1], scope)))
		if ctx != nil && ctx.def != nil && ctx.def.T != nil {
			expr = s.assignmentConversion(expr, ctx.def.T.Ret, node.SourceStart)
		}
		return &ReturnStmt{Value: expr, Range: node.SourceRange}
	case node.ReducedBy(parser.JumpStatement, 6):
		name := node.Children[1].Terminal.Lexeme
		if ctx.loopDepth == 0 {
			s.report(InvalidTypeSpec(node.SourceStart, "continue outside loop"))
		}
		if !namedTargetActive(ctx.namedContinue, name) {
			s.report(InvalidTypeSpec(node.SourceStart, "continue target is not an enclosing named loop"))
		}
		return &ContinueStmt{Name: name, Range: node.SourceRange}
	case node.ReducedBy(parser.JumpStatement, 7):
		name := node.Children[1].Terminal.Lexeme
		if ctx.loopDepth == 0 && len(ctx.switchStack) == 0 {
			s.report(InvalidTypeSpec(node.SourceStart, "break outside loop or switch"))
		}
		if !namedTargetActive(ctx.namedBreak, name) {
			s.report(InvalidTypeSpec(node.SourceStart, "break target is not an enclosing named loop or switch"))
		}
		return &BreakStmt{Name: name, Range: node.SourceRange}
	}
	return &EmptyStmt{Range: node.SourceRange}
}

func namedTargetActive(stack []string, name string) bool {
	for i := len(stack) - 1; i >= 0; i-- {
		if stack[i] == name {
			return true
		}
	}
	return false
}

func collectLabels(stmt Stmt, out map[string]*LabeledStmt) {
	switch x := stmt.(type) {
	case *LabeledStmt:
		out[x.Name] = x
		collectLabels(x.Body, out)
	case *Block:
		for _, it := range x.Items {
			collectLabels(it, out)
		}
	case *IfStmt:
		collectLabels(x.Then, out)
		collectLabels(x.Else, out)
	case *WhileStmt:
		collectLabels(x.Body, out)
	case *ForStmt:
		collectLabels(x.Body, out)
	case *SwitchStmt:
		collectLabels(x.Body, out)
	case *CaseStmt:
		collectLabels(x.Body, out)
	case *DefaultStmt:
		collectLabels(x.Body, out)
	}
}

func resolveGotos(pending []*GotoStmt, labels map[string]*LabeledStmt, barriers []vmScopeBarrier, s *Sema) {
	for _, g := range pending {
		target := labels[g.Name]
		if target == nil {
			s.report(UndeclaredIdentifier(g.Range.SourceStart, g.Name))
			continue
		}
		validateGotoVMJump(g, target, barriers, s)
		g.Target = target
		g.Name = ""
	}
}

func recordVMScopeBarrier(ctx *funcCtx, scope *Scope, declRange entity.SourceRange, t Type) {
	if ctx == nil || scope == nil || !typeHasVariablyModifiedType(t) {
		return
	}
	ctx.vmScopes = append(ctx.vmScopes, vmScopeBarrier{
		decl:      declRange,
		scope:     scope.Range,
		declOrder: ctx.nextOrder(),
	})
}

func validateGotoVMJump(g *GotoStmt, target *LabeledStmt, barriers []vmScopeBarrier, s *Sema) {
	if g == nil || target == nil {
		return
	}
	for _, barrier := range barriers {
		if orderJumpsIntoVMBarrier(g.Order, target.Order, barrier.declOrder) ||
			jumpEntersVMBarrier(g.Range.SourceStart, target.Range.SourceStart, barrier) {
			s.report(InvalidTypeSpec(g.Range.SourceStart, "jump into scope of identifier with variably modified type"))
			return
		}
	}
}

func validateSwitchVMJumps(sw *SwitchStmt, barriers []vmScopeBarrier, s *Sema) {
	if sw == nil {
		return
	}
	for _, c := range sw.Cases {
		if switchEntersVMBarrier(sw.Range.SourceStart, c.Range.SourceStart, sw.Order, c.Order, barriers) {
			s.report(InvalidTypeSpec(c.Range.SourceStart, "switch jumps into scope of identifier with variably modified type"))
		}
	}
	if sw.Default != nil && switchEntersVMBarrier(sw.Range.SourceStart, sw.Default.Range.SourceStart, sw.Order, sw.Default.Order, barriers) {
		s.report(InvalidTypeSpec(sw.Default.Range.SourceStart, "switch jumps into scope of identifier with variably modified type"))
	}
}

func switchEntersVMBarrier(from, target entity.SourcePos, fromOrder, targetOrder int, barriers []vmScopeBarrier) bool {
	for _, barrier := range barriers {
		if orderJumpsIntoVMBarrier(fromOrder, targetOrder, barrier.declOrder) {
			return true
		}
		if jumpEntersVMBarrier(from, target, barrier) {
			return true
		}
	}
	return false
}

func jumpEntersVMBarrier(from, target entity.SourcePos, barrier vmScopeBarrier) bool {
	// VM 标识符的作用域从声明之后开始，到声明所在作用域结束；
	// 跳转源点若还不在这个区间内，落到区间内部就是 C99 禁止的进入。
	if !posInRange(target, barrier.scope) || compareSourcePos(target, barrier.decl.SourceEnd) <= 0 {
		return false
	}
	return !posInRange(from, barrier.scope) || compareSourcePos(from, barrier.decl.SourceEnd) <= 0
}

func orderJumpsIntoVMBarrier(from, target, decl int) bool {
	return decl > 0 && from > 0 && target > 0 && from < decl && target > decl
}

func compareSourcePos(a, b entity.SourcePos) int {
	if a.Line != b.Line {
		return a.Line - b.Line
	}
	return a.Column - b.Column
}

func posInRange(pos entity.SourcePos, r entity.SourceRange) bool {
	return compareSourcePos(pos, r.SourceStart) >= 0 && compareSourcePos(pos, r.SourceEnd) <= 0
}
