package sema

import (
	"shinya.click/cvm/entity"
	"shinya.click/cvm/parser"
)

type rawInitElem struct {
	Designators []Designator
	Value       *entity.AstNode
}

type initLeaf struct {
	T           Type
	Designators []Designator
}

type initSpan struct {
	T           Type
	Designators []Designator
	Start       int
	End         int
}

func (s *Sema) typeInitializer(node *entity.AstNode, target Type) Expr {
	if expr := s.tryStringArrayInitializer(node, target); expr != nil {
		return expr
	}
	switch {
	case node.ReducedBy(parser.Initializer, 1):
		expr := s.typeExpr(node.Children[0], s.scope)
		expr = s.castFunctionDecay(s.castArrayDecay(s.castLValueToRValue(expr)))
		if isAggregateInitType(target) && isAggregateInitType(expr.GetType()) {
			return s.assignmentConversion(expr, target, node.SourceStart)
		}
		if sub := firstScalarInitializerType(target); sub != nil {
			target = sub
		}
		return s.assignmentConversion(expr, target, node.SourceStart)
	case node.ReducedBy(parser.Initializer, 2), node.ReducedBy(parser.Initializer, 3):
		return s.typeInitListForType(node.Children[1], target)
	case node.ReducedBy(parser.Initializer, 4):
		s.reportEmptyInitializerExtension(node.SourceStart)
		return &InitList{T: target, Range: node.SourceRange}
	}
	return s.errorExpr(node.SourceRange)
}

func (s *Sema) reportEmptyInitializerExtension(pos entity.SourcePos) {
	if !s.Options.GNUExtensions || s.Options.PedanticErrors {
		s.report(InvalidTypeSpec(pos, "empty initializer braces require GNU C mode"))
	}
}

func (s *Sema) tryStringArrayInitializer(node *entity.AstNode, target Type) Expr {
	at, ok := unqual(target).(*ArrayType)
	if !ok || !isStringInitializableArrayElem(unqual(at.Elem)) {
		return nil
	}
	if node.ReducedBy(parser.Initializer, 1) {
		if !hasStringLiteralToken(node.Children[0]) {
			return nil
		}
		expr := s.typeExpr(node.Children[0], s.scope)
		if _, ok := expr.(*StringLit); ok {
			return expr
		}
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
	if !hasStringLiteralToken(elem.Children[0]) {
		return nil
	}
	expr := s.typeExpr(elem.Children[0], s.scope)
	if _, ok := expr.(*StringLit); !ok {
		return nil
	}
	return expr
}

func hasStringLiteralToken(node *entity.AstNode) bool {
	if node == nil {
		return false
	}
	if node.Typ == entity.STRING {
		return true
	}
	for _, child := range node.Children {
		if hasStringLiteralToken(child) {
			return true
		}
	}
	return false
}

func isCharacterType(t Type) bool {
	bt, ok := unqual(t).(*BuiltinType)
	return ok && (bt.Kind == Char || bt.Kind == SChar || bt.Kind == UChar)
}

func isStringInitializableArrayElem(t Type) bool {
	bt, ok := unqual(t).(*BuiltinType)
	return ok && (bt.Kind == Char || bt.Kind == SChar || bt.Kind == UChar || bt.Kind == Int)
}

func firstScalarInitializerType(t Type) Type {
	switch x := unqual(t).(type) {
	case *StructType:
		if len(x.Fields) == 0 {
			return nil
		}
		return firstScalarInitializerType(x.Fields[0].T)
	case *UnionType:
		if len(x.Fields) == 0 {
			return nil
		}
		return firstScalarInitializerType(x.Fields[0].T)
	case *ArrayType:
		return firstScalarInitializerType(x.Elem)
	default:
		return t
	}
}

func (s *Sema) typeInitListForType(node *entity.AstNode, t Type) *InitList {
	il := &InitList{T: t, Range: node.SourceRange}
	at, ok := unqual(t).(*ArrayType)
	if ok && isStringInitializableArrayElem(unqual(at.Elem)) {
		if expr := s.tryStringArrayInitializerList(node); expr != nil {
			il.Elems = append(il.Elems, InitElem{Value: expr})
			return il
		}
	}
	var raw []rawInitElem
	s.collectRawInitList(node, t, &raw)
	s.typeCollectedInitList(t, raw, il)
	return il
}

func (s *Sema) completeUnsizedArrayInitializerType(t Type, init Expr) Type {
	at, ok := unqual(t).(*ArrayType)
	if !ok || at.SizeKind != ArrayUnsized {
		return t
	}
	if lit, ok := stringLiteralInitializerExpr(init); ok && isStringInitializableArrayElem(unqual(at.Elem)) {
		return s.Types.ArrayConstant(at.Elem, int64(len(lit.Value)+1))
	}
	il, ok := init.(*InitList)
	if !ok || il == nil {
		return t
	}
	if len(il.Elems) == 1 {
		if lit, ok := stringLiteralInitializerExpr(il.Elems[0].Value); ok && isStringInitializableArrayElem(unqual(at.Elem)) {
			return s.Types.ArrayConstant(at.Elem, int64(len(lit.Value)+1))
		}
	}
	elemLeaves := initLeaves(at.Elem, nil)
	leafCount := len(elemLeaves)
	if leafCount == 0 {
		leafCount = 1
	}
	cursor := 0
	maxCursor := 0
	for _, elem := range il.Elems {
		if len(elem.Designators) > 0 && elem.Designators[0].Kind == DesigArrayIndex {
			base := int(elem.Designators[0].Index) * leafCount
			relStart, relEnd := 0, leafCount
			elemType := at.Elem
			if len(elem.Designators) > 1 {
				if span, ok := designatedSpan(at.Elem, elem.Designators[1:]); ok {
					relStart, relEnd = span.Start, span.End
					elemType = span.T
				}
			}
			if !isObjectType(elemType) || initExprUsesWholeObject(elem.Value, elemType) {
				cursor = base + relEnd
			} else {
				cursor = base + relStart + 1
			}
		} else if initExprUsesWholeObject(elem.Value, at.Elem) {
			cursor += leafCount
		} else {
			cursor++
		}
		if cursor > maxCursor {
			maxCursor = cursor
		}
	}
	size := int64(0)
	if maxCursor > 0 {
		size = int64((maxCursor + leafCount - 1) / leafCount)
	}
	return s.Types.ArrayConstant(at.Elem, size)
}

func initExprUsesWholeObject(e Expr, t Type) bool {
	if _, ok := e.(*InitList); ok {
		return true
	}
	if _, ok := unqual(t).(*ArrayType); ok {
		_, ok := stringLiteralInitializerExpr(e)
		return ok
	}
	return false
}

func stringLiteralInitializerExpr(e Expr) (*StringLit, bool) {
	switch x := e.(type) {
	case *StringLit:
		return x, true
	case *ImplicitCast:
		if x.Kind == ArrayDecay || x.Kind == LValueToRValue {
			return stringLiteralInitializerExpr(x.X)
		}
	}
	return nil, false
}

func (s *Sema) collectRawInitList(node *entity.AstNode, target Type, out *[]rawInitElem) {
	switch {
	case node.ReducedBy(parser.InitializerList, 1):
		*out = append(*out, rawInitElem{Value: node.Children[0]})
	case node.ReducedBy(parser.InitializerList, 2):
		ds := s.parseDesignators(node.Children[0], target)
		*out = append(*out, rawInitElem{Designators: ds, Value: node.Children[1]})
	case node.ReducedBy(parser.InitializerList, 3):
		s.collectRawInitList(node.Children[0], target, out)
		*out = append(*out, rawInitElem{Value: node.Children[2]})
	case node.ReducedBy(parser.InitializerList, 4):
		s.collectRawInitList(node.Children[0], target, out)
		ds := s.parseDesignators(node.Children[2], target)
		*out = append(*out, rawInitElem{Designators: ds, Value: node.Children[3]})
	}
}

func (s *Sema) typeCollectedInitList(target Type, raw []rawInitElem, out *InitList) {
	leaves := initLeaves(target, nil)
	spans := directInitSpans(target, nil)
	cursor := 0
	for _, elem := range raw {
		if len(elem.Designators) > 0 {
			typed, next := s.typeDesignatedInitElem(target, elem)
			out.Elems = append(out.Elems, typed)
			cursor = next
			continue
		}
		if cursor >= len(leaves) {
			out.Elems = append(out.Elems, s.makeInitElem(nil, elem.Value, sequentialElementType(target, len(out.Elems))))
			continue
		}
		span := spanContaining(spans, cursor)
		if span != nil && cursor == span.Start && s.useWholeInitializer(elem.Value, span.T) {
			out.Elems = append(out.Elems, s.makeInitElem(span.Designators, elem.Value, span.T))
			cursor = span.End
			continue
		}
		leaf := leaves[cursor]
		out.Elems = append(out.Elems, s.makeInitElem(leaf.Designators, elem.Value, leaf.T))
		cursor++
	}
}

func (s *Sema) typeDesignatedInitElem(target Type, elem rawInitElem) (InitElem, int) {
	span, ok := designatedSpan(target, elem.Designators)
	if !ok || span.T == nil {
		return s.makeInitElem(elem.Designators, elem.Value, elementType(target, elem.Designators)), 0
	}
	if !isAggregateInitType(span.T) || s.useWholeInitializer(elem.Value, span.T) {
		return s.makeInitElem(elem.Designators, elem.Value, span.T), span.End
	}
	leaves := initLeaves(span.T, span.Designators)
	if len(leaves) == 0 {
		return s.makeInitElem(elem.Designators, elem.Value, span.T), span.End
	}
	leaf := leaves[0]
	return s.makeInitElem(leaf.Designators, elem.Value, leaf.T), span.Start + 1
}

func (s *Sema) makeInitElem(ds []Designator, value *entity.AstNode, elemType Type) InitElem {
	if at, ok := unqual(elemType).(*ArrayType); ok && at.SizeKind == ArrayUnsized && s.Options.PedanticErrors {
		s.report(InvalidTypeSpec(value.SourceStart, "cannot initialize flexible array member"))
	}
	return InitElem{Designators: ds, Value: s.typeInitializer(value, elemType)}
}

func (s *Sema) parseDesignators(node *entity.AstNode, target Type) []Designator {
	if node.ReducedBy(parser.Designation, 2) {
		s.reportObsoleteDesignator(node.SourceStart)
		out := []Designator{{Kind: DesigFieldName, Field: &Field{Name: node.Children[0].Terminal.Lexeme}}}
		s.resolveDesignators(target, out, node.SourceStart)
		return out
	}
	if node.ReducedBy(parser.Designation, 3) {
		s.reportObsoleteDesignator(node.SourceStart)
	}
	var out []Designator
	s.collectDesignatorList(node.Children[0], &out)
	s.resolveDesignators(target, out, node.SourceStart)
	return out
}

func (s *Sema) reportObsoleteDesignator(pos entity.SourcePos) {
	if !s.Options.GNUExtensions || s.Options.PedanticErrors {
		s.report(InvalidTypeSpec(pos, "obsolete GNU designated initializer requires GNU C mode"))
	}
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
		if !isInteger(expr.GetType()) {
			s.report(InvalidTypeSpec(node.SourceStart, "array designator expression must have integer type"))
			return Designator{Kind: DesigArrayIndex}
		}
		ev := NewEvaluator(s)
		cv, ok := ev.EvalC99IntegerConstantExpression(expr)
		if !ok && s.Options.GNUExtensions {
			cv, ok = ev.evalC99CastArithmeticConstant(expr, true, true, true)
		}
		if !ok {
			s.report(InvalidTypeSpec(node.SourceStart, "array designator expression must be integer constant expression"))
			return Designator{Kind: DesigArrayIndex}
		}
		if cv.Int < 0 {
			s.report(InvalidTypeSpec(node.SourceStart, "array designator index must be nonnegative"))
			return Designator{Kind: DesigArrayIndex}
		}
		return Designator{Kind: DesigArrayIndex, Index: cv.Int}
	case node.ReducedBy(parser.Designator, 2):
		return Designator{Kind: DesigFieldName, Field: &Field{Name: node.Children[1].Terminal.Lexeme}}
	case node.ReducedBy(parser.Designator, 3):
		if !s.Options.GNUExtensions && s.Options.PedanticErrors {
			s.report(InvalidTypeSpec(node.SourceStart, "range designator requires GNU C mode"))
		}
		expr := s.typeExpr(node.Children[1], s.scope)
		cv, ok := NewEvaluator(s).EvalC99IntegerConstantExpression(expr)
		if !ok {
			s.report(InvalidTypeSpec(node.SourceStart, "array designator expression must be integer constant expression"))
			return Designator{Kind: DesigArrayIndex}
		}
		if cv.Int < 0 {
			s.report(InvalidTypeSpec(node.SourceStart, "array designator index must be nonnegative"))
			return Designator{Kind: DesigArrayIndex}
		}
		return Designator{Kind: DesigArrayIndex, Index: cv.Int}
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
			if d.Field != nil && d.Field.T != nil {
				cur = d.Field.T
			}
		}
	}
	return cur
}

func initLeaves(t Type, prefix []Designator) []initLeaf {
	switch x := unqual(t).(type) {
	case *ArrayType:
		if x.SizeKind != ArrayConstantSize {
			return nil
		}
		var out []initLeaf
		for i := int64(0); i < x.Size; i++ {
			ds := appendDesignator(prefix, Designator{Kind: DesigArrayIndex, Index: i})
			out = append(out, initLeaves(x.Elem, ds)...)
		}
		return out
	case *StructType:
		var out []initLeaf
		for _, f := range x.Fields {
			if f == nil || (f.IsBitField && f.Name == "") {
				continue
			}
			ds := appendDesignator(prefix, Designator{Kind: DesigFieldName, Field: f})
			out = append(out, initLeaves(f.T, ds)...)
		}
		return out
	case *UnionType:
		for _, f := range x.Fields {
			if f == nil || (f.IsBitField && f.Name == "") {
				continue
			}
			ds := appendDesignator(prefix, Designator{Kind: DesigFieldName, Field: f})
			return initLeaves(f.T, ds)
		}
		return nil
	default:
		return []initLeaf{{T: t, Designators: copyDesignators(prefix)}}
	}
}

func isAggregateInitType(t Type) bool {
	switch unqual(t).(type) {
	case *ArrayType, *StructType, *UnionType:
		return true
	default:
		return false
	}
}

func directInitSpans(t Type, prefix []Designator) []initSpan {
	switch x := unqual(t).(type) {
	case *ArrayType:
		if x.SizeKind != ArrayConstantSize {
			return nil
		}
		out := make([]initSpan, 0, x.Size)
		cursor := 0
		for i := int64(0); i < x.Size; i++ {
			ds := appendDesignator(prefix, Designator{Kind: DesigArrayIndex, Index: i})
			n := len(initLeaves(x.Elem, ds))
			out = append(out, initSpan{T: x.Elem, Designators: ds, Start: cursor, End: cursor + n})
			cursor += n
		}
		return out
	case *StructType:
		out := make([]initSpan, 0, len(x.Fields))
		cursor := 0
		for _, f := range x.Fields {
			if f == nil || (f.IsBitField && f.Name == "") {
				continue
			}
			ds := appendDesignator(prefix, Designator{Kind: DesigFieldName, Field: f})
			n := len(initLeaves(f.T, ds))
			out = append(out, initSpan{T: f.T, Designators: ds, Start: cursor, End: cursor + n})
			cursor += n
		}
		return out
	case *UnionType:
		for _, f := range x.Fields {
			if f == nil || (f.IsBitField && f.Name == "") {
				continue
			}
			ds := appendDesignator(prefix, Designator{Kind: DesigFieldName, Field: f})
			n := len(initLeaves(f.T, ds))
			return []initSpan{{T: f.T, Designators: ds, Start: 0, End: n}}
		}
		return nil
	default:
		return []initSpan{{T: t, Designators: copyDesignators(prefix), Start: 0, End: 1}}
	}
}

func designatedSpan(t Type, ds []Designator) (initSpan, bool) {
	cur := t
	prefix := make([]Designator, 0, len(ds))
	var unionSpan *initSpan
	for _, d := range ds {
		switch d.Kind {
		case DesigArrayIndex:
			at, ok := unqual(cur).(*ArrayType)
			if !ok {
				return initSpan{}, false
			}
			prefix = appendDesignator(prefix, d)
			cur = at.Elem
		case DesigFieldName:
			if d.Field == nil || d.Field.T == nil {
				return initSpan{}, false
			}
			if _, ok := unqual(cur).(*UnionType); ok && unionSpan == nil {
				span, ok := spanForPrefix(t, prefix)
				if !ok {
					return initSpan{}, false
				}
				unionSpan = &span
			}
			prefix = appendDesignator(prefix, Designator{Kind: DesigFieldName, Field: d.Field})
			cur = d.Field.T
		default:
			return initSpan{}, false
		}
	}
	if unionSpan != nil {
		return initSpan{T: cur, Designators: copyDesignators(prefix), Start: unionSpan.Start, End: unionSpan.End}, true
	}
	span, ok := spanForPrefix(t, prefix)
	if !ok {
		return initSpan{}, false
	}
	span.T = cur
	return span, true
}

func spanForPrefix(t Type, prefix []Designator) (initSpan, bool) {
	leaves := initLeaves(t, nil)
	start := -1
	end := -1
	for i, leaf := range leaves {
		if designatorHasPrefix(leaf.Designators, prefix) {
			if start < 0 {
				start = i
			}
			end = i + 1
			continue
		}
		if start >= 0 {
			break
		}
	}
	if start < 0 {
		return initSpan{}, false
	}
	return initSpan{T: t, Designators: copyDesignators(prefix), Start: start, End: end}, true
}

func spanContaining(spans []initSpan, cursor int) *initSpan {
	for i := range spans {
		if cursor >= spans[i].Start && cursor < spans[i].End {
			return &spans[i]
		}
	}
	return nil
}

func (s *Sema) useWholeInitializer(node *entity.AstNode, t Type) bool {
	if node.ReducedBy(parser.Initializer, 2) || node.ReducedBy(parser.Initializer, 3) || node.ReducedBy(parser.Initializer, 4) {
		return true
	}
	return s.tryStringArrayInitializer(node, t) != nil
}

func appendDesignator(ds []Designator, d Designator) []Designator {
	out := copyDesignators(ds)
	out = append(out, d)
	return out
}

func copyDesignators(ds []Designator) []Designator {
	if len(ds) == 0 {
		return nil
	}
	out := make([]Designator, len(ds))
	copy(out, ds)
	return out
}

func designatorHasPrefix(ds, prefix []Designator) bool {
	if len(prefix) > len(ds) {
		return false
	}
	for i := range prefix {
		if ds[i].Kind != prefix[i].Kind {
			return false
		}
		switch prefix[i].Kind {
		case DesigArrayIndex:
			if ds[i].Index != prefix[i].Index {
				return false
			}
		case DesigFieldName:
			if ds[i].Field != prefix[i].Field {
				return false
			}
		}
	}
	return true
}

func (s *Sema) resolveDesignators(t Type, ds []Designator, pos entity.SourcePos) {
	cur := t
	for i := range ds {
		switch ds[i].Kind {
		case DesigArrayIndex:
			if at, ok := unqual(cur).(*ArrayType); ok {
				cur = at.Elem
			}
		case DesigFieldName:
			name := ds[i].Field.Name
			found := false
			if st, ok := unqual(cur).(*StructType); ok {
				for _, f := range st.Fields {
					if f.Name == name {
						ds[i].Field = f
						cur = f.T
						found = true
						break
					}
				}
			} else if ut, ok := unqual(cur).(*UnionType); ok {
				for _, f := range ut.Fields {
					if f.Name == name {
						ds[i].Field = f
						cur = f.T
						found = true
						break
					}
				}
			}
			if !found {
				s.report(InvalidTypeSpec(pos, "field designator does not match any field"))
			}
		}
	}
}
