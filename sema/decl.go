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
		return s.applyDirectDeclarator(node.Children[0], s.makeUnsizedArray(s.applyQualifierList(node.Children[2], base), node.SourceStart))
	case node.ReducedBy(parser.DirectDeclarator, 5):
		size := s.evalArraySize(node.Children[2])
		return s.applyDirectDeclarator(node.Children[0], s.makeArray(base, size, node.Children[2]))
	case node.ReducedBy(parser.DirectDeclarator, 6):
		elem := s.applyQualifierList(node.Children[2], base)
		size := s.evalArraySize(node.Children[3])
		return s.applyDirectDeclarator(node.Children[0], s.makeArray(elem, size, node.Children[3]))
	case node.ReducedBy(parser.DirectDeclarator, 7):
		return s.applyDirectDeclarator(node.Children[0], s.makeArray(base, s.evalArraySize(node.Children[3]), node.Children[3]))
	case node.ReducedBy(parser.DirectDeclarator, 8):
		return s.applyDirectDeclarator(node.Children[0], s.makeArray(base, s.evalArraySize(node.Children[4]), node.Children[4]))
	case node.ReducedBy(parser.DirectDeclarator, 9):
		return s.applyDirectDeclarator(node.Children[0], s.makeArray(base, s.evalArraySize(node.Children[4]), node.Children[4]))
	case node.ReducedBy(parser.DirectDeclarator, 10), node.ReducedBy(parser.DirectDeclarator, 11):
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

func (s *Sema) makeStarArray(elem Type, pos entity.SourcePos) Type {
	if !s.allowArrayStar {
		s.report(InvalidTypeSpec(pos, "star array size only allowed in function prototype scope"))
	}
	s.validateArrayElement(elem, pos)
	return s.Types.ArrayStar(elem)
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
		return spec.Type
	case node.ReducedBy(parser.ParameterDeclaration, 2):
		t, _ := s.applyDeclarator(node.Children[1], spec.Type)
		return s.adjustParamType(t)
	case node.ReducedBy(parser.ParameterDeclaration, 3):
		return s.adjustParamType(s.applyAbstractDeclarator(node.Children[1], spec.Type))
	}
	return ErrorTypeSingleton
}

func (s *Sema) adjustParamType(t Type) Type {
	switch x := unqual(t).(type) {
	case *ArrayType:
		return s.Types.Pointer(x.Elem)
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
	return out
}

func (s *Sema) paramDecl(node *entity.AstNode) *VarDecl {
	spec := s.parseSpec(node.Children[0])
	t := spec.Type
	name := ""
	if node.ReducedBy(parser.ParameterDeclaration, 2) {
		t, name = s.applyDeclarator(node.Children[1], spec.Type)
	} else if node.ReducedBy(parser.ParameterDeclaration, 3) {
		t = s.applyAbstractDeclarator(node.Children[1], spec.Type)
	}
	t = s.adjustParamType(t)
	sym := &Symbol{Name: name, Kind: SymParam, T: t, Storage: StorageAuto, Pos: node.SourceStart}
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
		return s.makeUnsizedArray(base, node.SourceStart)
	case node.ReducedBy(parser.DirectAbstractDeclarator, 3):
		return s.makeArray(base, s.evalArraySize(node.Children[1]), node.Children[1])
	case node.ReducedBy(parser.DirectAbstractDeclarator, 4), node.ReducedBy(parser.DirectAbstractDeclarator, 5), node.ReducedBy(parser.DirectAbstractDeclarator, 6):
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
		return s.applyDirectAbstractDeclarator(node.Children[0], s.makeUnsizedArray(base, node.SourceStart))
	case node.ReducedBy(parser.DirectAbstractDeclarator, 13):
		return s.applyDirectAbstractDeclarator(node.Children[0], s.makeArray(base, s.evalArraySize(node.Children[2]), node.Children[2]))
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

func (s *Sema) parseTypeName(node *entity.AstNode) Type {
	spec := s.parseSpec(node.Children[0])
	if node.ReducedBy(parser.TypeName, 2) {
		return s.applyAbstractDeclarator(node.Children[1], spec.Type)
	}
	return spec.Type
}

func typeNameUsesTypedef(node *entity.AstNode) bool {
	if node == nil {
		return false
	}
	if node.ReducedBy(parser.TypeSpecifier, 14) {
		return true
	}
	for _, child := range node.Children {
		if typeNameUsesTypedef(child) {
			return true
		}
	}
	return false
}

func typeNameIsBareTypedef(node *entity.AstNode) bool {
	if node == nil || !node.ReducedBy(parser.TypeName, 1) {
		return false
	}
	// 只有整个 type-name 来自 typedef 时，VM 边界才已在 typedef 声明处求值；
	// 若 abstract declarator 又写了 [m++]，那是当前 cast 新引入的边界。
	return typeNameUsesTypedef(node.Children[0])
}
