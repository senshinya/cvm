package codegen

import (
	"shinya.click/cvm/bytecode"
	"shinya.click/cvm/sema"
)

type storageKind int

const (
	storageLocalSlot storageKind = iota
	storageAddress
)

type storage struct {
	kind storageKind
	slot int
	typ  bytecode.ValueType
	sym  *sema.Symbol
}

func (fg *funcGen) storageForVar(sym *sema.Symbol, t sema.Type) (storage, error) {
	vt, err := fg.g.lowerValueType(t)
	if err != nil {
		return storage{}, err
	}
	if sym != nil && sym.SlotID >= 0 && isSlotType(vt) {
		return storage{kind: storageLocalSlot, slot: sym.SlotID, typ: vt, sym: sym}, nil
	}
	return storage{kind: storageAddress, typ: vt, sym: sym}, nil
}

func isSlotType(t bytecode.ValueType) bool {
	switch t {
	case bytecode.TypeBool,
		bytecode.TypeI8, bytecode.TypeI16, bytecode.TypeI32, bytecode.TypeI64,
		bytecode.TypeU8, bytecode.TypeU16, bytecode.TypeU32, bytecode.TypeU64,
		bytecode.TypeF32, bytecode.TypeF64, bytecode.TypeFLong,
		bytecode.TypePtr:
		return true
	default:
		return false
	}
}
