package codegen

import (
	"fmt"

	"shinya.click/cvm/bytecode"
	"shinya.click/cvm/sema"
)

type storageKind int

const (
	storageLocalSlot storageKind = iota
	storageAddress
)

type storage struct {
	kind   storageKind
	slot   int
	object int
	global int
	typ    bytecode.ValueType
	sym    *sema.Symbol
}

func (fg *funcGen) storageForVar(sym *sema.Symbol, t sema.Type) (storage, error) {
	vt, err := fg.g.lowerValueType(t)
	if err != nil {
		return storage{}, err
	}
	if sym != nil && fg.addressTaken[sym] {
		if id, ok := fg.objectMap[sym]; ok {
			return storage{kind: storageAddress, object: id, global: -1, typ: vt, sym: sym}, nil
		}
	}
	if sym != nil && sym.SlotID >= 0 && isSlotType(vt) {
		return storage{kind: storageLocalSlot, slot: sym.SlotID, object: -1, global: -1, typ: vt, sym: sym}, nil
	}
	if sym != nil && sym.Storage != sema.StorageAuto && sym.Storage != sema.StorageRegister && sym.GlobalID >= 0 {
		return storage{kind: storageAddress, global: sym.GlobalID, typ: vt, sym: sym}, nil
	}
	if sym != nil {
		if id, ok := fg.objectMap[sym]; ok {
			return storage{kind: storageAddress, object: id, global: -1, typ: vt, sym: sym}, nil
		}
	}
	return storage{kind: storageAddress, object: -1, global: -1, typ: vt, sym: sym}, nil
}

func (fg *funcGen) emitAddress(e sema.Expr) error {
	switch x := e.(type) {
	case *sema.VarRef:
		if object, ok := fg.dynamicObjectMap[x.Sym]; ok {
			fg.out.Instrs = append(fg.out.Instrs, bytecode.Instr{Op: bytecode.OpDynamicObjectAddr, Object: object, Type: bytecode.TypeObjectAddr})
			return nil
		}
		st, err := fg.storageForVar(x.Sym, x.T)
		if err != nil {
			return err
		}
		if st.global >= 0 {
			fg.out.Instrs = append(fg.out.Instrs, bytecode.AddrGlobal(st.global))
			return nil
		}
		if st.object >= 0 {
			fg.out.Instrs = append(fg.out.Instrs, bytecode.AddrLocalObject(st.object))
			return nil
		}
	case *sema.UnOp:
		switch x.Op {
		case sema.UnAddr:
			return fg.emitAddress(x.X)
		case sema.UnDeref:
			if err := fg.emitValue(x.X); err != nil {
				return err
			}
			return fg.ensureObjectAddr(x.X.GetType())
		}
	case *sema.MemberExpr:
		layoutType := x.Base.GetType()
		if x.Arrow {
			if err := fg.emitValue(x.Base); err != nil {
				return err
			}
			if err := fg.ensureObjectAddr(x.Base.GetType()); err != nil {
				return err
			}
			if pt, ok := sema.Unqual(x.Base.GetType()).(*sema.PointerType); ok {
				layoutType = pt.Pointee
			}
		} else if err := fg.emitAddress(x.Base); err != nil {
			return err
		}
		layout, err := fg.g.lowerLayout(layoutType)
		if err != nil {
			return err
		}
		field, err := fg.g.fieldID(layout.ID, x.Field)
		if err != nil {
			return err
		}
		fg.out.Instrs = append(fg.out.Instrs, bytecode.Instr{Op: bytecode.OpFieldAddr, Layout: layout.ID, Field: field})
		return nil
	case *sema.IndexExpr:
		if err := fg.emitValue(x.Base); err != nil {
			return err
		}
		if err := fg.ensureObjectAddr(x.Base.GetType()); err != nil {
			return err
		}
		if err := fg.emitValue(x.Index); err != nil {
			return err
		}
		return fg.emitPtrAddForExpr(x.Base, x.Base.GetType())
	case *sema.StringLit:
		fg.out.Instrs = append(fg.out.Instrs, bytecode.AddrString(fg.g.internString(x.Value)))
		return nil
	case *sema.ImplicitCast:
		if x.Kind == sema.ArrayDecay || x.Kind == sema.LValueToRValue {
			return fg.emitAddress(x.X)
		}
	}
	return &Error{Pos: e.Pos().SourceStart, Node: fmt.Sprintf("%T", e), Op: "emitAddress", Reason: "expression is not addressable"}
}

func (fg *funcGen) ensureObjectAddr(t sema.Type) error {
	vt, err := fg.g.lowerValueType(t)
	if err != nil {
		return err
	}
	if vt == bytecode.TypePtr {
		fg.out.Instrs = append(fg.out.Instrs, bytecode.Cast(bytecode.TypePtr, bytecode.TypeObjectAddr, bytecode.CastBit))
	}
	return nil
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
