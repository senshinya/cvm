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
