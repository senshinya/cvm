package runtime

import (
	"fmt"
	"math"

	"shinya.click/cvm/bytecode"
)

type Value struct {
	Type  bytecode.ValueType
	Int   uint64
	Float float64
}

func IntValue(t bytecode.ValueType, v int64) Value {
	return Value{Type: t, Int: uint64(v)}
}

func UIntValue(t bytecode.ValueType, v uint64) Value {
	return Value{Type: t, Int: v}
}

func PtrValue(addr uint64) Value {
	return Value{Type: bytecode.TypePtr, Int: addr}
}

func ObjectAddrValue(addr uint64) Value {
	return Value{Type: bytecode.TypeObjectAddr, Int: addr}
}

func FloatValue(t bytecode.ValueType, v float64) Value {
	return Value{Type: t, Float: v}
}

func (v Value) ExitCode() (int, error) {
	switch v.Type {
	case bytecode.TypeI8, bytecode.TypeI16, bytecode.TypeI32, bytecode.TypeI64:
		signed := int64(v.Int)
		if int64(int(signed)) != signed {
			return 0, fmt.Errorf("exit code %d exceeds int range", signed)
		}
		return int(signed), nil
	case bytecode.TypeBool, bytecode.TypeU8, bytecode.TypeU16, bytecode.TypeU32, bytecode.TypeU64:
		if v.Int > uint64(math.MaxInt) {
			return 0, fmt.Errorf("exit code %d exceeds int range", v.Int)
		}
		return int(v.Int), nil
	default:
		return 0, fmt.Errorf("cannot use %s as exit code", v.Type)
	}
}

func (v Value) IsZero() bool {
	switch v.Type {
	case bytecode.TypeF32, bytecode.TypeF64, bytecode.TypeFLong:
		return v.Float == 0
	default:
		return v.Int == 0
	}
}
