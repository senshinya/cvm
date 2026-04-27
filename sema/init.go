package sema

import (
	"shinya.click/cvm/entity"
	"shinya.click/cvm/parser"
)

func (s *Sema) typeInitializer(node *entity.AstNode, target Type) Expr {
	if expr := s.tryStringArrayInitializer(node, target); expr != nil {
		return expr
	}
	switch {
	case node.ReducedBy(parser.Initializer, 1):
		expr := s.typeExpr(node.Children[0], s.scope)
		expr = s.castFunctionDecay(s.castArrayDecay(s.castLValueToRValue(expr)))
		return s.assignmentConversion(expr, target, node.SourceStart)
	case node.ReducedBy(parser.Initializer, 2), node.ReducedBy(parser.Initializer, 3):
		return s.typeInitListForType(node.Children[1], target)
	}
	return s.errorExpr(node.SourceRange)
}

func (s *Sema) tryStringArrayInitializer(node *entity.AstNode, target Type) Expr {
	at, ok := unqual(target).(*ArrayType)
	if !ok || !isCharacterType(unqual(at.Elem)) {
		return nil
	}
	if !node.ReducedBy(parser.Initializer, 2) && !node.ReducedBy(parser.Initializer, 3) {
		return nil
	}
	return s.tryStringArrayInitializerList(node.Children[1])
}

func (s *Sema) tryStringArrayInitializerList(list *entity.AstNode) Expr {
	if !list.ReducedBy(parser.InitializerList, 1) {
		return nil
	}
	elem := list.Children[0]
	if !elem.ReducedBy(parser.Initializer, 1) {
		return nil
	}
	expr := s.typeExpr(elem.Children[0], s.scope)
	if _, ok := expr.(*StringLit); !ok {
		return nil
	}
	return expr
}

func isCharacterType(t Type) bool {
	bt, ok := unqual(t).(*BuiltinType)
	return ok && (bt.Kind == Char || bt.Kind == SChar || bt.Kind == UChar)
}

func (s *Sema) typeInitListForType(node *entity.AstNode, t Type) *InitList {
	il := &InitList{T: t, Range: node.SourceRange}
	at, ok := unqual(t).(*ArrayType)
	if ok && isCharacterType(unqual(at.Elem)) {
		if expr := s.tryStringArrayInitializerList(node); expr != nil {
			il.Elems = append(il.Elems, InitElem{Value: expr})
			return il
		}
	}
	s.collectInitList(node, t, il)
	return il
}

func (s *Sema) collectInitList(node *entity.AstNode, target Type, out *InitList) {
	switch {
	case node.ReducedBy(parser.InitializerList, 1):
		elemT := sequentialElementType(target, len(out.Elems))
		out.Elems = append(out.Elems, s.makeInitElem(nil, node.Children[0], elemT))
	case node.ReducedBy(parser.InitializerList, 2):
		ds := s.parseDesignators(node.Children[0], target)
		out.Elems = append(out.Elems, s.makeInitElem(ds, node.Children[1], elementType(target, ds)))
	case node.ReducedBy(parser.InitializerList, 3):
		s.collectInitList(node.Children[0], target, out)
		elemT := sequentialElementType(target, len(out.Elems))
		out.Elems = append(out.Elems, s.makeInitElem(nil, node.Children[2], elemT))
	case node.ReducedBy(parser.InitializerList, 4):
		s.collectInitList(node.Children[0], target, out)
		ds := s.parseDesignators(node.Children[2], target)
		out.Elems = append(out.Elems, s.makeInitElem(ds, node.Children[3], elementType(target, ds)))
	}
}

func (s *Sema) makeInitElem(ds []Designator, value *entity.AstNode, elemType Type) InitElem {
	return InitElem{Designators: ds, Value: s.typeInitializer(value, elemType)}
}

func (s *Sema) parseDesignators(node *entity.AstNode, target Type) []Designator {
	var out []Designator
	s.collectDesignatorList(node.Children[0], &out)
	resolveDesignators(target, out)
	return out
}

func (s *Sema) collectDesignatorList(node *entity.AstNode, out *[]Designator) {
	switch {
	case node.ReducedBy(parser.DesignatorList, 1):
		*out = append(*out, s.parseDesignator(node.Children[0]))
	case node.ReducedBy(parser.DesignatorList, 2):
		s.collectDesignatorList(node.Children[0], out)
		*out = append(*out, s.parseDesignator(node.Children[1]))
	}
}

func (s *Sema) parseDesignator(node *entity.AstNode) Designator {
	switch {
	case node.ReducedBy(parser.Designator, 1):
		expr := s.typeExpr(node.Children[1], s.scope)
		cv, _ := NewEvaluator(s).EvalIntegerConstant(expr)
		return Designator{Kind: DesigArrayIndex, Index: cv.Int}
	case node.ReducedBy(parser.Designator, 2):
		return Designator{Kind: DesigFieldName, Field: &Field{Name: node.Children[1].Terminal.Lexeme}}
	}
	return Designator{}
}

func sequentialElementType(t Type, index int) Type {
	switch x := unqual(t).(type) {
	case *ArrayType:
		return x.Elem
	case *StructType:
		if index >= 0 && index < len(x.Fields) {
			return x.Fields[index].T
		}
	case *UnionType:
		if len(x.Fields) > 0 {
			return x.Fields[0].T
		}
	}
	return t
}

func elementType(t Type, ds []Designator) Type {
	cur := t
	for _, d := range ds {
		switch d.Kind {
		case DesigArrayIndex:
			if at, ok := unqual(cur).(*ArrayType); ok {
				cur = at.Elem
			}
		case DesigFieldName:
			if d.Field != nil {
				cur = d.Field.T
			}
		}
	}
	return cur
}

func resolveDesignators(t Type, ds []Designator) {
	cur := t
	for i := range ds {
		switch ds[i].Kind {
		case DesigArrayIndex:
			if at, ok := unqual(cur).(*ArrayType); ok {
				cur = at.Elem
			}
		case DesigFieldName:
			name := ds[i].Field.Name
			if st, ok := unqual(cur).(*StructType); ok {
				for _, f := range st.Fields {
					if f.Name == name {
						ds[i].Field = f
						cur = f.T
						break
					}
				}
			} else if ut, ok := unqual(cur).(*UnionType); ok {
				for _, f := range ut.Fields {
					if f.Name == name {
						ds[i].Field = f
						cur = f.T
						break
					}
				}
			}
		}
	}
}
