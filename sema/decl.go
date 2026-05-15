package sema

import (
	"shinya.click/cvm/entity"
	"shinya.click/cvm/parser"
)

type arraySizeInfo struct {
	Expr       Expr
	Value      int64
	Constant   bool
	KnownValue bool
}

func (s *Sema) applyDeclarator(node *entity.AstNode, base Type) (Type, string) {
	switch {
	case node.ReducedBy(parser.Declarator, 1):
		return s.applyDirectDeclarator(node.Children[0], base)
	case node.ReducedBy(parser.Declarator, 2):
		base = s.applyPointer(node.Children[0], base)
		return s.applyDirectDeclarator(node.Children[1], base)
	}
	return base, ""
}

func (s *Sema) applyPointer(node *entity.AstNode, base Type) Type {
	switch {
	case node.ReducedBy(parser.Pointer, 1):
		return s.Types.Pointer(base)
	case node.ReducedBy(parser.Pointer, 2):
		return s.applyQualifierList(node.Children[1], s.Types.Pointer(base))
	case node.ReducedBy(parser.Pointer, 3):
		return s.applyPointer(node.Children[1], s.Types.Pointer(base))
	case node.ReducedBy(parser.Pointer, 4):
		return s.applyPointer(node.Children[2], s.applyQualifierList(node.Children[1], s.Types.Pointer(base)))
	}
	return base
}

func (s *Sema) applyQualifierList(node *entity.AstNode, base Type) Type {
	var c, v, r bool
	s.collectTypeQualifierList(node, &c, &v, &r)
	if c || v || r {
		return s.Types.Qualified(base, c, v, r)
	}
	return base
}

func (s *Sema) collectTypeQualifierList(node *entity.AstNode, c, v, r *bool) {
	switch {
	case node.ReducedBy(parser.TypeQualifierList, 1):
		s.handleQualifier(node.Children[0], c, v, r)
	case node.ReducedBy(parser.TypeQualifierList, 2):
		s.collectTypeQualifierList(node.Children[0], c, v, r)
		s.handleQualifier(node.Children[1], c, v, r)
	}
}

func (s *Sema) applyDirectDeclarator(node *entity.AstNode, base Type) (Type, string) {
	switch {
	case node.ReducedBy(parser.DirectDeclarator, 1):
		return base, node.Children[0].Terminal.Lexeme
	case node.ReducedBy(parser.DirectDeclarator, 2):
		return s.applyDeclarator(node.Children[1], base)
	case node.ReducedBy(parser.DirectDeclarator, 3):
		return s.applyDirectDeclarator(node.Children[0], s.makeUnsizedArray(base, node.SourceStart))
	case node.ReducedBy(parser.DirectDeclarator, 4):
		return s.applyDirectDeclarator(node.Children[0], s.makeUnsizedArrayWithParamQualifiers(base, node.SourceStart, s.arrayParamQualifiers(node.Children[2])))
	case node.ReducedBy(parser.DirectDeclarator, 5):
		size := s.evalArraySize(node.Children[2])
		return s.applyDirectDeclarator(node.Children[0], s.makeArray(base, size, node.Children[2]))
	case node.ReducedBy(parser.DirectDeclarator, 6):
		size := s.evalArraySize(node.Children[3])
		return s.applyDirectDeclarator(node.Children[0], s.makeArrayWithParamQualifiers(base, size, node.Children[3], s.arrayParamQualifiers(node.Children[2])))
	case node.ReducedBy(parser.DirectDeclarator, 7):
		return s.applyDirectDeclarator(node.Children[0], s.makeArray(base, s.evalArraySize(node.Children[3]), node.Children[3]))
	case node.ReducedBy(parser.DirectDeclarator, 8):
		return s.applyDirectDeclarator(node.Children[0], s.makeArrayWithParamQualifiers(base, s.evalArraySize(node.Children[4]), node.Children[4], s.arrayParamQualifiers(node.Children[3])))
	case node.ReducedBy(parser.DirectDeclarator, 9):
		return s.applyDirectDeclarator(node.Children[0], s.makeArrayWithParamQualifiers(base, s.evalArraySize(node.Children[4]), node.Children[4], s.arrayParamQualifiers(node.Children[2])))
	case node.ReducedBy(parser.DirectDeclarator, 10), node.ReducedBy(parser.DirectDeclarator, 11):
		if node.ReducedBy(parser.DirectDeclarator, 11) {
			return s.applyDirectDeclarator(node.Children[0], s.makeStarArrayWithParamQualifiers(base, node.SourceStart, s.arrayParamQualifiers(node.Children[2])))
		}
		return s.applyDirectDeclarator(node.Children[0], s.makeStarArray(base, node.SourceStart))
	case node.ReducedBy(parser.DirectDeclarator, 12):
		return s.applyDirectDeclarator(node.Children[0], s.buildFunctionType(node.Children[2], base))
	case node.ReducedBy(parser.DirectDeclarator, 13):
		s.validateFunctionReturn(base, node.SourceStart)
		return s.applyDirectDeclarator(node.Children[0], s.Types.Function(base, nil, false, false))
	case node.ReducedBy(parser.DirectDeclarator, 14):
		s.validateFunctionReturn(base, node.SourceStart)
		return s.applyDirectDeclarator(node.Children[0], s.Types.Function(base, nil, false, false))
	}
	return base, ""
}

type arrayDeclaratorSyntax struct {
	Pos       entity.SourcePos
	Decorated bool
	Blocked   bool
}

func (s *Sema) validateDeclaratorArrayQualifiers(node *entity.AstNode, parameter bool) {
	s.validateArrayDeclaratorSyntax(declaratorArraySyntax(node), parameter)
}

func (s *Sema) validateAbstractDeclaratorArrayQualifiers(node *entity.AstNode, parameter bool) {
	s.validateArrayDeclaratorSyntax(abstractDeclaratorArraySyntax(node), parameter)
}

func (s *Sema) validateArrayDeclaratorSyntax(arrays []arrayDeclaratorSyntax, parameter bool) {
	for i, arr := range arrays {
		if !arr.Decorated {
			continue
		}
		switch {
		case !parameter:
			s.report(InvalidTypeSpec(arr.Pos, "array static or type qualifiers only allowed in function parameter"))
		case i != 0 || arr.Blocked:
			// C99 只允许参数声明最外层数组派生携带 static 或限定符；
			// 指向数组的指针、以及多维数组的内层派生都不能使用这些标记。
			s.report(InvalidTypeSpec(arr.Pos, "array static or type qualifiers only allowed in outermost parameter array"))
		}
	}
}

func declaratorArraySyntax(node *entity.AstNode) []arrayDeclaratorSyntax {
	if node == nil {
		return nil
	}
	switch {
	case node.ReducedBy(parser.Declarator, 1):
		return directDeclaratorArraySyntax(node.Children[0])
	case node.ReducedBy(parser.Declarator, 2):
		return directDeclaratorArraySyntax(node.Children[1])
	}
	return nil
}

func directDeclaratorArraySyntax(node *entity.AstNode) []arrayDeclaratorSyntax {
	if node == nil {
		return nil
	}
	switch {
	case node.ReducedBy(parser.DirectDeclarator, 2):
		return declaratorArraySyntax(node.Children[1])
	case node.ReducedBy(parser.DirectDeclarator, 3), node.ReducedBy(parser.DirectDeclarator, 4),
		node.ReducedBy(parser.DirectDeclarator, 5), node.ReducedBy(parser.DirectDeclarator, 6),
		node.ReducedBy(parser.DirectDeclarator, 7), node.ReducedBy(parser.DirectDeclarator, 8),
		node.ReducedBy(parser.DirectDeclarator, 9), node.ReducedBy(parser.DirectDeclarator, 10),
		node.ReducedBy(parser.DirectDeclarator, 11):
		out := directDeclaratorArraySyntax(node.Children[0])
		out = append(out, arrayDeclaratorSyntax{
			Pos:       node.SourceStart,
			Decorated: directDeclaratorHasArrayQualifierOrStatic(node),
			Blocked:   directDeclaratorHasParenthesizedPointerBase(node.Children[0]),
		})
		return out
	case node.ReducedBy(parser.DirectDeclarator, 12), node.ReducedBy(parser.DirectDeclarator, 13),
		node.ReducedBy(parser.DirectDeclarator, 14):
		return directDeclaratorArraySyntax(node.Children[0])
	}
	return nil
}

func directDeclaratorHasArrayQualifierOrStatic(node *entity.AstNode) bool {
	return node.ReducedBy(parser.DirectDeclarator, 4) ||
		node.ReducedBy(parser.DirectDeclarator, 6) ||
		node.ReducedBy(parser.DirectDeclarator, 7) ||
		node.ReducedBy(parser.DirectDeclarator, 8) ||
		node.ReducedBy(parser.DirectDeclarator, 9) ||
		node.ReducedBy(parser.DirectDeclarator, 11)
}

func directDeclaratorHasParenthesizedPointerBase(node *entity.AstNode) bool {
	if node == nil {
		return false
	}
	if node.ReducedBy(parser.DirectDeclarator, 2) {
		return declaratorHasPointer(node.Children[1])
	}
	if len(node.Children) > 0 {
		return directDeclaratorHasParenthesizedPointerBase(node.Children[0])
	}
	return false
}

func declaratorHasPointer(node *entity.AstNode) bool {
	if node == nil {
		return false
	}
	switch {
	case node.ReducedBy(parser.Declarator, 2):
		return true
	case node.ReducedBy(parser.Declarator, 1):
		return directDeclaratorHasParenthesizedPointerBase(node.Children[0])
	}
	return false
}

type typeQualifiers struct {
	Const, Volatile, Restrict bool
}

func (s *Sema) arrayParamQualifiers(node *entity.AstNode) typeQualifiers {
	var q typeQualifiers
	s.collectTypeQualifierList(node, &q.Const, &q.Volatile, &q.Restrict)
	return q
}

func (s *Sema) evalArraySize(node *entity.AstNode) arraySizeInfo {
	if node == nil {
		return arraySizeInfo{}
	}
	expr := s.typeExpr(node, s.scope)
	ev := NewEvaluator(s)
	if cv, ok := ev.EvalC99IntegerConstantExpression(expr); ok {
		return arraySizeInfo{Expr: expr, Value: cv.Int, Constant: true, KnownValue: true}
	}
	if cv, ok := ev.EvalC99ArraySizeConstantExpression(expr); ok {
		return arraySizeInfo{Expr: expr, Value: cv.Int, KnownValue: true}
	}
	return arraySizeInfo{Expr: expr}
}

func (s *Sema) makeArray(elem Type, size arraySizeInfo, sizeNode *entity.AstNode) Type {
	pos := entity.SourcePos{}
	if sizeNode != nil {
		pos = sizeNode.SourceStart
	}
	s.validateArrayElement(elem, pos)
	if !size.Constant {
		// C99 仍要求非 ICE 的常量数组大小为正，但这类大小不能把数组固定化。
		if size.KnownValue && size.Value <= 0 {
			s.report(InvalidTypeSpec(pos, "array size must be positive"))
		}
		return s.Types.ArrayVLA(elem, size.Expr)
	}
	if size.Value <= 0 {
		s.report(InvalidTypeSpec(pos, "array size must be positive"))
	}
	return s.Types.ArrayConstant(elem, size.Value)
}

func (s *Sema) makeUnsizedArray(elem Type, pos entity.SourcePos) Type {
	s.validateArrayElement(elem, pos)
	return s.Types.ArrayUnsized(elem)
}

func (s *Sema) makeUnsizedArrayWithParamQualifiers(elem Type, pos entity.SourcePos, q typeQualifiers) Type {
	return s.withArrayParamQualifiers(s.makeUnsizedArray(elem, pos), q)
}

func (s *Sema) makeStarArray(elem Type, pos entity.SourcePos) Type {
	if !s.allowArrayStar {
		s.report(InvalidTypeSpec(pos, "star array size only allowed in function prototype scope"))
	}
	s.validateArrayElement(elem, pos)
	return s.Types.ArrayStar(elem)
}

func (s *Sema) makeStarArrayWithParamQualifiers(elem Type, pos entity.SourcePos, q typeQualifiers) Type {
	return s.withArrayParamQualifiers(s.makeStarArray(elem, pos), q)
}

func (s *Sema) makeArrayWithParamQualifiers(elem Type, size arraySizeInfo, sizeNode *entity.AstNode, q typeQualifiers) Type {
	return s.withArrayParamQualifiers(s.makeArray(elem, size, sizeNode), q)
}

func (s *Sema) withArrayParamQualifiers(t Type, q typeQualifiers) Type {
	if !q.Const && !q.Volatile && !q.Restrict {
		return t
	}
	at, ok := t.(*ArrayType)
	if !ok {
		return t
	}
	// C99 中参数数组 declarator 方括号内的限定符限定调整后的指针，
	// 不能先落到元素类型上，否则 int a[restrict] 会被误判为 restrict int[]。
	cp := *at
	cp.ParamConst = q.Const
	cp.ParamVolatile = q.Volatile
	cp.ParamRestrict = q.Restrict
	return &cp
}

func (s *Sema) validateArrayElement(elem Type, pos entity.SourcePos) {
	if !isObjectType(elem) && !(s.allowArrayStar && isPrototypeArrayObjectType(elem)) {
		s.report(InvalidTypeSpec(pos, "array element type must be complete object type"))
		return
	}
	if typeContainsFlexibleArrayMember(elem) {
		s.report(InvalidTypeSpec(pos, "array element type contains flexible array member"))
	}
}

func (s *Sema) buildFunctionType(paramList *entity.AstNode, ret Type) *FunctionType {
	s.validateFunctionReturn(ret, paramList.SourceStart)
	prevScope := s.scope
	prevAllowArrayStar := s.allowArrayStar
	s.scope = NewScope(ScopeFuncProto, prevScope)
	s.allowArrayStar = true
	defer func() {
		s.scope = prevScope
		s.allowArrayStar = prevAllowArrayStar
	}()
	var params []Type
	var variadic bool
	switch {
	case paramList.ReducedBy(parser.ParameterTypeList, 1):
		params = s.collectParameterList(paramList.Children[0])
	case paramList.ReducedBy(parser.ParameterTypeList, 2):
		params = s.collectParameterList(paramList.Children[0])
		variadic = true
	}
	if len(params) == 1 {
		if bt, ok := unqual(params[0]).(*BuiltinType); ok && bt.Kind == Void {
			params = nil
		}
	}
	return s.Types.Function(ret, params, variadic, true)
}

func (s *Sema) validateFunctionReturn(ret Type, pos entity.SourcePos) {
	switch unqual(ret).(type) {
	case *ArrayType, *FunctionType:
		s.report(InvalidTypeSpec(pos, "function cannot return array or function type"))
	}
}

func (s *Sema) validateRestrictType(t Type, pos entity.SourcePos) {
	s.validateRestrictTypeSeen(t, pos, map[Type]bool{})
}

func (s *Sema) validateRestrictTypeSeen(t Type, pos entity.SourcePos, seen map[Type]bool) {
	if t == nil || seen[t] {
		return
	}
	seen[t] = true
	switch x := t.(type) {
	case *QualType:
		if x.Restrict && !restrictQualifierTargetOK(x.Base) {
			s.report(InvalidTypeSpec(pos, "restrict qualifier requires pointer to object type"))
		}
		s.validateRestrictTypeSeen(x.Base, pos, seen)
	case *PointerType:
		s.validateRestrictTypeSeen(x.Pointee, pos, seen)
	case *ArrayType:
		s.validateRestrictTypeSeen(x.Elem, pos, seen)
	case *FunctionType:
		s.validateRestrictTypeSeen(x.Ret, pos, seen)
		for _, p := range x.Params {
			s.validateRestrictTypeSeen(p, pos, seen)
		}
	case *StructType:
		for _, f := range x.Fields {
			s.validateRestrictTypeSeen(f.T, pos, seen)
		}
	case *UnionType:
		for _, f := range x.Fields {
			s.validateRestrictTypeSeen(f.T, pos, seen)
		}
	}
}

func restrictQualifierTargetOK(t Type) bool {
	switch x := unqual(t).(type) {
	case *PointerType:
		return isRestrictPointerTarget(x.Pointee)
	case *ArrayType:
		return restrictQualifierTargetOK(x.Elem)
	}
	return false
}

func (s *Sema) collectParameterList(node *entity.AstNode) []Type {
	var out []Type
	switch {
	case node.ReducedBy(parser.ParameterList, 1):
		out = append(out, s.parameterDeclarationType(node.Children[0]))
	case node.ReducedBy(parser.ParameterList, 2):
		out = append(out, s.collectParameterList(node.Children[0])...)
		out = append(out, s.parameterDeclarationType(node.Children[2]))
	}
	return out
}

func (s *Sema) parameterDeclarationType(node *entity.AstNode) Type {
	spec := s.parseSpec(node.Children[0])
	prevAllowArrayStar := s.allowArrayStar
	s.allowArrayStar = true
	defer func() { s.allowArrayStar = prevAllowArrayStar }()
	switch {
	case node.ReducedBy(parser.ParameterDeclaration, 1):
		s.validateRestrictType(spec.Type, node.SourceStart)
		return spec.Type
	case node.ReducedBy(parser.ParameterDeclaration, 2):
		s.validateDeclaratorArrayQualifiers(node.Children[1], true)
		t, name := s.applyDeclarator(node.Children[1], spec.Type)
		s.validateRestrictType(t, node.SourceStart)
		t = s.adjustParamType(t)
		s.insertPrototypeParameter(name, t, node.SourceStart)
		return t
	case node.ReducedBy(parser.ParameterDeclaration, 3):
		s.validateAbstractDeclaratorArrayQualifiers(node.Children[1], true)
		t := s.applyAbstractDeclarator(node.Children[1], spec.Type)
		s.validateRestrictType(t, node.SourceStart)
		return s.adjustParamType(t)
	}
	return ErrorTypeSingleton
}

func (s *Sema) insertPrototypeParameter(name string, t Type, pos entity.SourcePos) {
	if name == "" || s.scope == nil || s.scope.Kind != ScopeFuncProto {
		return
	}
	s.scope.Insert(name, &Symbol{Name: name, Kind: SymParam, T: t, Storage: StorageAuto, Pos: pos})
}

func (s *Sema) adjustParamType(t Type) Type {
	switch x := unqual(t).(type) {
	case *ArrayType:
		ptr := Type(s.Types.Pointer(x.Elem))
		if x.ParamConst || x.ParamVolatile || x.ParamRestrict {
			ptr = s.Types.Qualified(ptr, x.ParamConst, x.ParamVolatile, x.ParamRestrict)
		}
		return ptr
	case *FunctionType:
		return s.Types.Pointer(x)
	}
	return t
}

func (s *Sema) collectParamDecls(declarator *entity.AstNode, ft *FunctionType) []*VarDecl {
	direct := declarator.Children[0]
	if declarator.ReducedBy(parser.Declarator, 2) {
		direct = declarator.Children[1]
	}
	for {
		switch {
		case direct.ReducedBy(parser.DirectDeclarator, 12):
			return s.paramDeclsFromList(direct.Children[2])
		case direct.ReducedBy(parser.DirectDeclarator, 13), direct.ReducedBy(parser.DirectDeclarator, 14), direct.ReducedBy(parser.DirectDeclarator, 1):
			return nil
		default:
			direct = direct.Children[0]
		}
		_ = ft
	}
}

func (s *Sema) paramDeclsFromList(node *entity.AstNode) []*VarDecl {
	prevScope := s.scope
	prevAllowArrayStar := s.allowArrayStar
	s.scope = NewScope(ScopeFuncProto, prevScope)
	s.allowArrayStar = true
	defer func() {
		s.scope = prevScope
		s.allowArrayStar = prevAllowArrayStar
	}()
	var out []*VarDecl
	var walk func(*entity.AstNode)
	walk = func(p *entity.AstNode) {
		switch {
		case p.ReducedBy(parser.ParameterList, 1):
			out = append(out, s.paramDecl(p.Children[0]))
		case p.ReducedBy(parser.ParameterList, 2):
			walk(p.Children[0])
			out = append(out, s.paramDecl(p.Children[2]))
		}
	}
	switch {
	case node.ReducedBy(parser.ParameterTypeList, 1), node.ReducedBy(parser.ParameterTypeList, 2):
		walk(node.Children[0])
	}
	if len(out) == 1 && out[0] != nil && out[0].Sym != nil && out[0].Sym.Name == "" && isPlainVoidType(out[0].T) {
		return nil
	}
	return out
}

func (s *Sema) paramDecl(node *entity.AstNode) *VarDecl {
	spec := s.parseSpec(node.Children[0])
	t := spec.Type
	name := ""
	if node.ReducedBy(parser.ParameterDeclaration, 2) {
		s.validateDeclaratorArrayQualifiers(node.Children[1], true)
		t, name = s.applyDeclarator(node.Children[1], spec.Type)
	} else if node.ReducedBy(parser.ParameterDeclaration, 3) {
		s.validateAbstractDeclaratorArrayQualifiers(node.Children[1], true)
		t = s.applyAbstractDeclarator(node.Children[1], spec.Type)
	}
	s.validateInlineSpecifier(spec, t, name, node.SourceStart, true)
	s.validateRestrictType(t, node.SourceStart)
	t = s.adjustParamType(t)
	sym := &Symbol{Name: name, Kind: SymParam, T: t, Storage: StorageAuto, Pos: node.SourceStart}
	s.insertPrototypeParameter(name, t, node.SourceStart)
	return &VarDecl{Sym: sym, T: t, Storage: StorageAuto, IsParam: true, Range: node.SourceRange}
}

func (s *Sema) applyAbstractDeclarator(node *entity.AstNode, base Type) Type {
	switch {
	case node.ReducedBy(parser.AbstractDeclarator, 1):
		return s.applyPointer(node.Children[0], base)
	case node.ReducedBy(parser.AbstractDeclarator, 2):
		return s.applyDirectAbstractDeclarator(node.Children[0], base)
	case node.ReducedBy(parser.AbstractDeclarator, 3):
		return s.applyDirectAbstractDeclarator(node.Children[1], s.applyPointer(node.Children[0], base))
	}
	return base
}

func (s *Sema) applyDirectAbstractDeclarator(node *entity.AstNode, base Type) Type {
	switch {
	case node.ReducedBy(parser.DirectAbstractDeclarator, 1), node.ReducedBy(parser.DirectAbstractDeclarator, 2):
		if node.ReducedBy(parser.DirectAbstractDeclarator, 2) {
			return s.makeUnsizedArrayWithParamQualifiers(base, node.SourceStart, s.arrayParamQualifiers(node.Children[1]))
		}
		return s.makeUnsizedArray(base, node.SourceStart)
	case node.ReducedBy(parser.DirectAbstractDeclarator, 3):
		return s.makeArray(base, s.evalArraySize(node.Children[1]), node.Children[1])
	case node.ReducedBy(parser.DirectAbstractDeclarator, 4), node.ReducedBy(parser.DirectAbstractDeclarator, 5), node.ReducedBy(parser.DirectAbstractDeclarator, 6):
		if node.ReducedBy(parser.DirectAbstractDeclarator, 5) {
			return s.makeArrayWithParamQualifiers(base, s.evalArraySize(node.Children[len(node.Children)-2]), node.Children[len(node.Children)-2], s.arrayParamQualifiers(node.Children[2]))
		}
		if node.ReducedBy(parser.DirectAbstractDeclarator, 6) {
			return s.makeArrayWithParamQualifiers(base, s.evalArraySize(node.Children[len(node.Children)-2]), node.Children[len(node.Children)-2], s.arrayParamQualifiers(node.Children[1]))
		}
		return s.makeArray(base, s.evalArraySize(node.Children[len(node.Children)-2]), node.Children[len(node.Children)-2])
	case node.ReducedBy(parser.DirectAbstractDeclarator, 7):
		return s.makeStarArray(base, node.SourceStart)
	case node.ReducedBy(parser.DirectAbstractDeclarator, 8):
		s.validateFunctionReturn(base, node.SourceStart)
		return s.Types.Function(base, nil, false, false)
	case node.ReducedBy(parser.DirectAbstractDeclarator, 9):
		return s.buildFunctionType(node.Children[1], base)
	case node.ReducedBy(parser.DirectAbstractDeclarator, 10):
		return s.applyAbstractDeclarator(node.Children[1], base)
	case node.ReducedBy(parser.DirectAbstractDeclarator, 11), node.ReducedBy(parser.DirectAbstractDeclarator, 12):
		if node.ReducedBy(parser.DirectAbstractDeclarator, 12) {
			return s.applyDirectAbstractDeclarator(node.Children[0], s.makeUnsizedArrayWithParamQualifiers(base, node.SourceStart, s.arrayParamQualifiers(node.Children[2])))
		}
		return s.applyDirectAbstractDeclarator(node.Children[0], s.makeUnsizedArray(base, node.SourceStart))
	case node.ReducedBy(parser.DirectAbstractDeclarator, 13):
		return s.applyDirectAbstractDeclarator(node.Children[0], s.makeArray(base, s.evalArraySize(node.Children[2]), node.Children[2]))
	case node.ReducedBy(parser.DirectAbstractDeclarator, 14):
		return s.applyDirectAbstractDeclarator(node.Children[0], s.makeArrayWithParamQualifiers(base, s.evalArraySize(node.Children[3]), node.Children[3], s.arrayParamQualifiers(node.Children[2])))
	case node.ReducedBy(parser.DirectAbstractDeclarator, 16):
		return s.applyDirectAbstractDeclarator(node.Children[0], s.makeArrayWithParamQualifiers(base, s.evalArraySize(node.Children[4]), node.Children[4], s.arrayParamQualifiers(node.Children[3])))
	case node.ReducedBy(parser.DirectAbstractDeclarator, 17):
		return s.applyDirectAbstractDeclarator(node.Children[0], s.makeArrayWithParamQualifiers(base, s.evalArraySize(node.Children[4]), node.Children[4], s.arrayParamQualifiers(node.Children[2])))
	case node.ReducedBy(parser.DirectAbstractDeclarator, 18):
		return s.applyDirectAbstractDeclarator(node.Children[0], s.makeStarArray(base, node.SourceStart))
	case node.ReducedBy(parser.DirectAbstractDeclarator, 19):
		s.validateFunctionReturn(base, node.SourceStart)
		return s.applyDirectAbstractDeclarator(node.Children[0], s.Types.Function(base, nil, false, false))
	case node.ReducedBy(parser.DirectAbstractDeclarator, 20):
		return s.applyDirectAbstractDeclarator(node.Children[0], s.buildFunctionType(node.Children[2], base))
	}
	return base
}

func abstractDeclaratorArraySyntax(node *entity.AstNode) []arrayDeclaratorSyntax {
	if node == nil {
		return nil
	}
	switch {
	case node.ReducedBy(parser.AbstractDeclarator, 2):
		return directAbstractDeclaratorArraySyntax(node.Children[0])
	case node.ReducedBy(parser.AbstractDeclarator, 3):
		return directAbstractDeclaratorArraySyntax(node.Children[1])
	}
	return nil
}

func directAbstractDeclaratorArraySyntax(node *entity.AstNode) []arrayDeclaratorSyntax {
	if node == nil {
		return nil
	}
	switch {
	case node.ReducedBy(parser.DirectAbstractDeclarator, 1), node.ReducedBy(parser.DirectAbstractDeclarator, 2),
		node.ReducedBy(parser.DirectAbstractDeclarator, 3), node.ReducedBy(parser.DirectAbstractDeclarator, 4),
		node.ReducedBy(parser.DirectAbstractDeclarator, 5), node.ReducedBy(parser.DirectAbstractDeclarator, 6),
		node.ReducedBy(parser.DirectAbstractDeclarator, 7):
		return []arrayDeclaratorSyntax{{
			Pos:       node.SourceStart,
			Decorated: directAbstractDeclaratorHasArrayQualifierOrStatic(node),
		}}
	case node.ReducedBy(parser.DirectAbstractDeclarator, 10):
		return abstractDeclaratorArraySyntax(node.Children[1])
	case node.ReducedBy(parser.DirectAbstractDeclarator, 11), node.ReducedBy(parser.DirectAbstractDeclarator, 12),
		node.ReducedBy(parser.DirectAbstractDeclarator, 13), node.ReducedBy(parser.DirectAbstractDeclarator, 14),
		node.ReducedBy(parser.DirectAbstractDeclarator, 15), node.ReducedBy(parser.DirectAbstractDeclarator, 16),
		node.ReducedBy(parser.DirectAbstractDeclarator, 17), node.ReducedBy(parser.DirectAbstractDeclarator, 18):
		out := directAbstractDeclaratorArraySyntax(node.Children[0])
		out = append(out, arrayDeclaratorSyntax{
			Pos:       node.SourceStart,
			Decorated: directAbstractDeclaratorHasArrayQualifierOrStatic(node),
			Blocked:   directAbstractDeclaratorHasParenthesizedPointerBase(node.Children[0]),
		})
		return out
	case node.ReducedBy(parser.DirectAbstractDeclarator, 19), node.ReducedBy(parser.DirectAbstractDeclarator, 20):
		return directAbstractDeclaratorArraySyntax(node.Children[0])
	}
	return nil
}

func directAbstractDeclaratorHasArrayQualifierOrStatic(node *entity.AstNode) bool {
	return node.ReducedBy(parser.DirectAbstractDeclarator, 2) ||
		node.ReducedBy(parser.DirectAbstractDeclarator, 4) ||
		node.ReducedBy(parser.DirectAbstractDeclarator, 5) ||
		node.ReducedBy(parser.DirectAbstractDeclarator, 6) ||
		node.ReducedBy(parser.DirectAbstractDeclarator, 12) ||
		node.ReducedBy(parser.DirectAbstractDeclarator, 14) ||
		node.ReducedBy(parser.DirectAbstractDeclarator, 15) ||
		node.ReducedBy(parser.DirectAbstractDeclarator, 16) ||
		node.ReducedBy(parser.DirectAbstractDeclarator, 17)
}

func directAbstractDeclaratorHasParenthesizedPointerBase(node *entity.AstNode) bool {
	if node == nil {
		return false
	}
	if node.ReducedBy(parser.DirectAbstractDeclarator, 10) {
		return abstractDeclaratorHasPointer(node.Children[1])
	}
	if len(node.Children) > 0 {
		return directAbstractDeclaratorHasParenthesizedPointerBase(node.Children[0])
	}
	return false
}

func abstractDeclaratorHasPointer(node *entity.AstNode) bool {
	if node == nil {
		return false
	}
	switch {
	case node.ReducedBy(parser.AbstractDeclarator, 1), node.ReducedBy(parser.AbstractDeclarator, 3):
		return true
	case node.ReducedBy(parser.AbstractDeclarator, 2):
		return directAbstractDeclaratorHasParenthesizedPointerBase(node.Children[0])
	}
	return false
}

func (s *Sema) parseTypeName(node *entity.AstNode) Type {
	spec := s.parseSpec(node.Children[0])
	if node.ReducedBy(parser.TypeName, 2) {
		s.validateAbstractDeclaratorArrayQualifiers(node.Children[1], false)
		t := s.applyAbstractDeclarator(node.Children[1], spec.Type)
		s.validateRestrictType(t, node.SourceStart)
		return t
	}
	s.validateRestrictType(spec.Type, node.SourceStart)
	return spec.Type
}
