package codegen

import (
	"fmt"

	"shinya.click/cvm/bytecode"
	"shinya.click/cvm/sema"
)

func (g *generator) lowerValueType(t sema.Type) (bytecode.ValueType, error) {
	switch x := sema.Unqual(t).(type) {
	case *sema.BuiltinType:
		switch x.Kind {
		case sema.Void:
			return bytecode.TypeVoid, nil
		case sema.Bool:
			return bytecode.TypeBool, nil
		case sema.Char, sema.SChar:
			return bytecode.TypeI8, nil
		case sema.UChar:
			return bytecode.TypeU8, nil
		case sema.Short:
			return bytecode.TypeI16, nil
		case sema.UShort:
			return bytecode.TypeU16, nil
		case sema.Int:
			return bytecode.TypeI32, nil
		case sema.UInt:
			return bytecode.TypeU32, nil
		case sema.Long, sema.LongLong:
			return bytecode.TypeI64, nil
		case sema.ULong, sema.ULongLong:
			return bytecode.TypeU64, nil
		case sema.Float:
			return bytecode.TypeF32, nil
		case sema.Double:
			return bytecode.TypeF64, nil
		case sema.LongDouble:
			return bytecode.TypeFLong, nil
		case sema.FloatComplex, sema.DoubleComplex, sema.LongDoubleComplex:
			return bytecode.TypeObjectAddr, nil
		}
	case *sema.PointerType, *sema.FunctionType:
		return bytecode.TypePtr, nil
	case *sema.ArrayType, *sema.StructType, *sema.UnionType:
		return bytecode.TypeObjectAddr, nil
	case *sema.EnumType:
		return bytecode.TypeI32, nil
	}
	return bytecode.TypeVoid, fmt.Errorf("cannot lower sema type %T (%s)", t, t)
}

func (g *generator) lowerFuncSig(t *sema.FunctionType) (int, error) {
	if t == nil {
		return 0, fmt.Errorf("cannot lower nil function type")
	}
	ret, err := g.lowerValueType(t.Ret)
	if err != nil {
		return 0, err
	}
	params := make([]bytecode.ValueType, 0, len(t.Params))
	for _, p := range t.Params {
		pt, err := g.lowerValueType(p)
		if err != nil {
			return 0, err
		}
		params = append(params, pt)
	}
	return g.internSig(ret, params, t.Variadic), nil
}

func (g *generator) lowerLayout(t sema.Type) (bytecode.ObjectLayout, error) {
	key := sema.Unqual(t)
	if id, ok := g.layoutMap[key]; ok {
		return g.mod.Layouts[id], nil
	}
	id := len(g.mod.Layouts)
	layout := bytecode.ObjectLayout{
		ID:    id,
		Name:  key.String(),
		Size:  g.sizeof(key),
		Align: g.alignof(key),
	}
	g.layoutMap[key] = id
	g.mod.Layouts = append(g.mod.Layouts, layout)

	switch x := key.(type) {
	case *sema.ArrayType:
		layout.ElemSize = g.sizeof(x.Elem)
		if isObjectType(x.Elem) {
			if _, err := g.lowerLayout(x.Elem); err != nil {
				return bytecode.ObjectLayout{}, err
			}
		}
	case *sema.StructType:
		layout.Fields, layout.Bit = g.lowerFields(x.Fields)
	case *sema.UnionType:
		layout.Fields, layout.Bit = g.lowerFields(x.Fields)
	}
	g.mod.Layouts[id] = layout
	return layout, nil
}

func (g *generator) lowerFields(fields []*sema.Field) ([]bytecode.FieldLayout, []bytecode.BitFieldLayout) {
	out := make([]bytecode.FieldLayout, 0, len(fields))
	bits := []bytecode.BitFieldLayout{}
	for _, f := range fields {
		if f == nil {
			continue
		}
		vt, err := g.lowerValueType(f.T)
		if err != nil {
			vt = bytecode.TypeVoid
		}
		if f.IsBitField {
			bits = append(bits, bytecode.BitFieldLayout{
				ID:           len(bits),
				Name:         f.Name,
				Container:    vt,
				ByteOffset:   f.Offset,
				Width:        f.BitWidth,
				Signed:       g.isSigned(f.T),
				Volatile:     isVolatile(f.T),
				LayoutPolicy: g.mod.Target.BitFieldPolicy,
			})
			continue
		}
		out = append(out, bytecode.FieldLayout{
			ID:     len(out),
			Name:   f.Name,
			Offset: f.Offset,
			Type:   vt,
		})
	}
	return out, bits
}

func (g *generator) fieldID(layoutID int, field *sema.Field) (int, error) {
	if field == nil {
		return 0, fmt.Errorf("nil field")
	}
	if layoutID < 0 || layoutID >= len(g.mod.Layouts) {
		return 0, fmt.Errorf("invalid layout %d", layoutID)
	}
	for _, f := range g.mod.Layouts[layoutID].Fields {
		if f.Name == field.Name && f.Offset == field.Offset {
			return f.ID, nil
		}
	}
	return 0, fmt.Errorf("field %q not found in layout %d", field.Name, layoutID)
}

func (g *generator) bitFieldID(layoutID int, field *sema.Field) (int, error) {
	if field == nil {
		return 0, fmt.Errorf("nil bit-field")
	}
	if layoutID < 0 || layoutID >= len(g.mod.Layouts) {
		return 0, fmt.Errorf("invalid layout %d", layoutID)
	}
	for _, f := range g.mod.Layouts[layoutID].Bit {
		if f.Name == field.Name && f.ByteOffset == field.Offset {
			return f.ID, nil
		}
	}
	return 0, fmt.Errorf("bit-field %q not found in layout %d", field.Name, layoutID)
}

func (g *generator) elemSize(t sema.Type) int64 {
	if pt, ok := sema.Unqual(t).(*sema.PointerType); ok {
		return g.sizeof(pt.Pointee)
	}
	if at, ok := sema.Unqual(t).(*sema.ArrayType); ok {
		return g.sizeof(at.Elem)
	}
	return 1
}

func (g *generator) sizeof(t sema.Type) int64 {
	switch x := sema.Unqual(t).(type) {
	case *sema.BuiltinType:
		switch x.Kind {
		case sema.Void:
			return 1
		case sema.Bool, sema.Char, sema.SChar, sema.UChar:
			return 1
		case sema.Short, sema.UShort:
			return 2
		case sema.Int, sema.UInt, sema.Float:
			return 4
		case sema.Long, sema.ULong, sema.LongLong, sema.ULongLong, sema.Double:
			return 8
		case sema.LongDouble:
			return 16
		case sema.FloatComplex:
			return 8
		case sema.DoubleComplex:
			return 16
		case sema.LongDoubleComplex:
			return 32
		}
	case *sema.PointerType, *sema.FunctionType:
		return g.mod.Target.PointerSize
	case *sema.ArrayType:
		if x.SizeKind == sema.ArrayConstantSize {
			return x.Size * g.sizeof(x.Elem)
		}
	case *sema.StructType:
		var end int64
		for _, f := range x.Fields {
			if f == nil {
				continue
			}
			if n := f.Offset + g.sizeof(f.T); n > end {
				end = n
			}
		}
		return end
	case *sema.UnionType:
		var max int64
		for _, f := range x.Fields {
			if f == nil {
				continue
			}
			if n := g.sizeof(f.T); n > max {
				max = n
			}
		}
		return max
	case *sema.EnumType:
		return g.sizeof(x.Underlying)
	}
	return 0
}

func isObjectType(t sema.Type) bool {
	switch x := sema.Unqual(t).(type) {
	case *sema.ArrayType, *sema.StructType, *sema.UnionType:
		return true
	case *sema.BuiltinType:
		switch x.Kind {
		case sema.FloatComplex, sema.DoubleComplex, sema.LongDoubleComplex:
			return true
		}
	default:
		return false
	}
	return false
}

func isVLAType(t sema.Type) bool {
	switch x := sema.Unqual(t).(type) {
	case *sema.ArrayType:
		return x.SizeKind == sema.ArrayVLA || x.SizeKind == sema.ArrayStarSize || isVLAType(x.Elem)
	default:
		return false
	}
}

func typeHasVariableSize(t sema.Type) bool {
	switch x := sema.Unqual(t).(type) {
	case *sema.ArrayType:
		return x.SizeKind == sema.ArrayVLA || x.SizeKind == sema.ArrayStarSize || typeHasVariableSize(x.Elem)
	case *sema.StructType:
		for _, f := range x.Fields {
			if f != nil && typeHasVariableSize(f.T) {
				return true
			}
		}
	case *sema.UnionType:
		for _, f := range x.Fields {
			if f != nil && typeHasVariableSize(f.T) {
				return true
			}
		}
	case *sema.PointerType:
		return false
	default:
		return false
	}
	return false
}

func isComplexType(t sema.Type) bool {
	bt, ok := sema.Unqual(t).(*sema.BuiltinType)
	if !ok {
		return false
	}
	switch bt.Kind {
	case sema.FloatComplex, sema.DoubleComplex, sema.LongDoubleComplex:
		return true
	default:
		return false
	}
}

func typeHasVariablyModifiedType(t sema.Type) bool {
	switch x := sema.Unqual(t).(type) {
	case *sema.ArrayType:
		return x.SizeKind == sema.ArrayVLA || x.SizeKind == sema.ArrayStarSize || typeHasVariablyModifiedType(x.Elem)
	case *sema.PointerType:
		return typeHasVariablyModifiedType(x.Pointee)
	default:
		return false
	}
}

func (g *generator) alignof(t sema.Type) int64 {
	switch x := sema.Unqual(t).(type) {
	case *sema.BuiltinType:
		switch x.Kind {
		case sema.Void, sema.Bool, sema.Char, sema.SChar, sema.UChar:
			return 1
		case sema.Short, sema.UShort:
			return 2
		case sema.Int, sema.UInt, sema.Float:
			return 4
		case sema.Long, sema.ULong, sema.LongLong, sema.ULongLong, sema.Double:
			return 8
		case sema.LongDouble:
			return 16
		case sema.FloatComplex:
			return 4
		case sema.DoubleComplex:
			return 8
		case sema.LongDoubleComplex:
			return 16
		}
	case *sema.PointerType, *sema.FunctionType:
		return g.mod.Target.PointerAlign
	case *sema.ArrayType:
		return g.alignof(x.Elem)
	case *sema.StructType, *sema.UnionType:
		return 1
	case *sema.EnumType:
		return g.alignof(x.Underlying)
	}
	return 1
}

func (g *generator) isSigned(t sema.Type) bool {
	bt, ok := sema.Unqual(t).(*sema.BuiltinType)
	if !ok {
		return false
	}
	switch bt.Kind {
	case sema.Bool, sema.UChar, sema.UShort, sema.UInt, sema.ULong, sema.ULongLong:
		return false
	default:
		return true
	}
}

func isVolatile(t sema.Type) bool {
	if q, ok := t.(*sema.QualType); ok {
		return q.Volatile
	}
	return false
}
