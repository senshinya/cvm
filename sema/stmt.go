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
	vmScopes     []vmScopeBarrier
}

type vmScopeBarrier struct {
	decl  entity.SourceRange
	scope entity.SourceRange
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

func (s *Sema) walkBlockInitDecl(node *entity.AstNode, spec SpecResult, scope *Scope, ctx *funcCtx, srcRange entity.SourceRange) Decl {
	s.validateDeclaratorArrayQualifiers(node.Children[0], false)
	t, name := s.applyDeclarator(node.Children[0], spec.Type)
	pos := node.Children[0].SourceStart
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
	recordVMScopeBarrier(ctx, scope, srcRange, t)
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
	forScope.Range = node.SourceRange
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
		decl:  declRange,
		scope: scope.Range,
	})
}

func validateGotoVMJump(g *GotoStmt, target *LabeledStmt, barriers []vmScopeBarrier, s *Sema) {
	if g == nil || target == nil {
		return
	}
	for _, barrier := range barriers {
		if jumpEntersVMBarrier(g.Range.SourceStart, target.Range.SourceStart, barrier) {
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
		if switchEntersVMBarrier(sw.Range.SourceStart, c.Range.SourceStart, barriers) {
			s.report(InvalidTypeSpec(c.Range.SourceStart, "switch jumps into scope of identifier with variably modified type"))
		}
	}
	if sw.Default != nil && switchEntersVMBarrier(sw.Range.SourceStart, sw.Default.Range.SourceStart, barriers) {
		s.report(InvalidTypeSpec(sw.Default.Range.SourceStart, "switch jumps into scope of identifier with variably modified type"))
	}
}

func switchEntersVMBarrier(from, target entity.SourcePos, barriers []vmScopeBarrier) bool {
	for _, barrier := range barriers {
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

func compareSourcePos(a, b entity.SourcePos) int {
	if a.Line != b.Line {
		return a.Line - b.Line
	}
	return a.Column - b.Column
}

func posInRange(pos entity.SourcePos, r entity.SourceRange) bool {
	return compareSourcePos(pos, r.SourceStart) >= 0 && compareSourcePos(pos, r.SourceEnd) <= 0
}
