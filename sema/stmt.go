package sema

import (
	"shinya.click/cvm/common"
	"shinya.click/cvm/entity"
	"shinya.click/cvm/parser"
)

type funcCtx struct {
	def          *FuncDef
	loopDepth    int
	switchStack  []*SwitchStmt
	pendingGotos []*GotoStmt
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
	prev := s.scope
	s.scope = scope
	defer func() { s.scope = prev }()
	block := &Block{Range: node.SourceRange, Scope: scope}
	if node.ReducedBy(parser.CompoundStatement, 2) {
		s.collectBlockItems(node.Children[1], scope, ctx, &block.Items)
	}
	return block
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
		var decls []Decl
		s.walkBlockDecl(node.Children[0], scope, ctx, &decls)
		if len(decls) > 0 {
			*out = append(*out, &DeclStmt{Decls: decls, Range: node.SourceRange})
		}
	case node.ReducedBy(parser.BlockItem, 2):
		*out = append(*out, s.typeStmt(node.Children[0], scope, ctx))
	}
}

func (s *Sema) walkBlockDecl(node *entity.AstNode, scope *Scope, ctx *funcCtx, out *[]Decl) {
	if node.ReducedBy(parser.Declaration, 3) {
		s.typeStaticAssert(node.Children[0])
		return
	}
	spec := s.parseSpec(node.Children[0])
	if s.Options.PedanticErrors && hasEnumReferenceSpecifier(node.Children[0]) {
		s.report(InvalidTypeSpec(node.SourceStart, "ISO C forbids forward references to enum types"))
	}
	if node.ReducedBy(parser.Declaration, 1) {
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

func (s *Sema) walkBlockInitDecl(node *entity.AstNode, spec SpecResult, scope *Scope, ctx *funcCtx, srcRange entity.SourceRange) Decl {
	t, name := s.applyDeclarator(node.Children[0], spec.Type)
	pos := node.Children[0].SourceStart
	if spec.IsTypedef {
		markTypedefVMBounds(t)
		sym := &Symbol{Name: name, Kind: SymTypedef, T: t, Storage: StorageTypedef, Pos: pos}
		td := &TypedefDecl{Sym: sym, T: t, Range: srcRange}
		sym.Decl = td
		if err := scope.InsertChecked(name, sym); err != nil {
			s.report(err.(*common.CvmError))
		}
		return td
	}
	storage := spec.Storage
	if storage == StorageNone {
		storage = StorageAuto
	}
	if storage == StorageStatic && typeHasDisallowedStaticArrayBound(t) {
		s.report(InvalidTypeSpec(pos, "array size must be integer constant expression"))
	}
	sym := &Symbol{Name: name, Kind: SymVar, T: t, Storage: storage, Pos: pos}
	vd := &VarDecl{Sym: sym, T: t, Storage: storage, Range: srcRange}
	sym.Decl = vd
	if node.ReducedBy(parser.InitDeclarator, 2) {
		vd.Init = s.typeInitializer(node.Children[2], t)
	}
	if err := scope.InsertChecked(name, sym); err != nil {
		s.report(err.(*common.CvmError))
	}
	if ctx != nil && ctx.def != nil {
		ctx.def.Locals = append(ctx.def.Locals, vd)
	}
	return vd
}

func (s *Sema) typeSelection(node *entity.AstNode, scope *Scope, ctx *funcCtx) Stmt {
	switch {
	case node.ReducedBy(parser.SelectionStatement, 1):
		cond := s.castBoolConversion(s.castLValueToRValue(s.typeExpr(node.Children[2], scope)))
		return &IfStmt{Cond: cond, Then: s.typeStmt(node.Children[4], scope, ctx), Range: node.SourceRange}
	case node.ReducedBy(parser.SelectionStatement, 2):
		cond := s.castBoolConversion(s.castLValueToRValue(s.typeExpr(node.Children[2], scope)))
		return &IfStmt{Cond: cond, Then: s.typeStmt(node.Children[4], scope, ctx), Else: s.typeStmt(node.Children[6], scope, ctx), Range: node.SourceRange}
	case node.ReducedBy(parser.SelectionStatement, 3):
		cond := s.castIntegerPromotion(s.castLValueToRValue(s.typeExpr(node.Children[2], scope)))
		sw := &SwitchStmt{Cond: cond, Range: node.SourceRange}
		ctx.switchStack = append(ctx.switchStack, sw)
		sw.Body = s.typeStmt(node.Children[4], scope, ctx)
		ctx.switchStack = ctx.switchStack[:len(ctx.switchStack)-1]
		collectCasesAndDefault(sw.Body, sw, s)
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
	}
}

func (s *Sema) typeLabeled(node *entity.AstNode, scope *Scope, ctx *funcCtx) Stmt {
	switch {
	case node.ReducedBy(parser.LabeledStatement, 1):
		return &LabeledStmt{Name: node.Children[0].Terminal.Lexeme, Body: s.typeStmt(node.Children[2], scope, ctx), Range: node.SourceRange}
	case node.ReducedBy(parser.LabeledStatement, 2):
		expr := s.typeExpr(node.Children[1], scope)
		cv, ok := NewEvaluator(s).EvalC99IntegerConstantExpression(expr)
		if !ok {
			s.report(InvalidTypeSpec(node.SourceStart, "case value must be integer constant expression"))
		}
		return &CaseStmt{Value: cv.Int, Body: s.typeStmt(node.Children[3], scope, ctx), Range: node.SourceRange}
	case node.ReducedBy(parser.LabeledStatement, 3):
		return &DefaultStmt{Body: s.typeStmt(node.Children[2], scope, ctx), Range: node.SourceRange}
	}
	return &EmptyStmt{Range: node.SourceRange}
}

func (s *Sema) typeIteration(node *entity.AstNode, scope *Scope, ctx *funcCtx) Stmt {
	switch {
	case node.ReducedBy(parser.IterationStatement, 1):
		cond := s.castBoolConversion(s.castLValueToRValue(s.typeExpr(node.Children[2], scope)))
		ctx.loopDepth++
		body := s.typeStmt(node.Children[4], scope, ctx)
		ctx.loopDepth--
		return &WhileStmt{Cond: cond, Body: body, Range: node.SourceRange}
	case node.ReducedBy(parser.IterationStatement, 2):
		ctx.loopDepth++
		body := s.typeStmt(node.Children[1], scope, ctx)
		ctx.loopDepth--
		cond := s.castBoolConversion(s.castLValueToRValue(s.typeExpr(node.Children[4], scope)))
		return &WhileStmt{Cond: cond, Body: body, DoWhile: true, Range: node.SourceRange}
	}
	forScope := NewScope(ScopeBlock, scope)
	prev := s.scope
	s.scope = forScope
	defer func() { s.scope = prev }()
	fp := s.collectForParts(node, forScope, ctx)
	fs := &ForStmt{Init: fp.init, Cond: fp.cond, Post: fp.post, Scope: forScope, Range: node.SourceRange}
	ctx.loopDepth++
	fs.Body = s.typeStmt(fp.body, forScope, ctx)
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
		s.walkBlockDecl(node.Children[2], scope, ctx, &decls)
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

func (s *Sema) typeJump(node *entity.AstNode, scope *Scope, ctx *funcCtx) Stmt {
	switch {
	case node.ReducedBy(parser.JumpStatement, 1):
		g := &GotoStmt{Name: node.Children[1].Terminal.Lexeme, Range: node.SourceRange}
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
	}
	return &EmptyStmt{Range: node.SourceRange}
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

func resolveGotos(pending []*GotoStmt, labels map[string]*LabeledStmt, s *Sema) {
	for _, g := range pending {
		target := labels[g.Name]
		if target == nil {
			s.report(UndeclaredIdentifier(g.Range.SourceStart, g.Name))
			continue
		}
		g.Target = target
		g.Name = ""
	}
}
