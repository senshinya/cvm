package codegen

import (
	"fmt"

	"shinya.click/cvm/bytecode"
	"shinya.click/cvm/sema"
)

func (fg *funcGen) emitValue(e sema.Expr) error {
	switch x := e.(type) {
	case *sema.IntLit:
		t, err := fg.g.lowerValueType(x.T)
		if err != nil {
			return err
		}
		fg.out.Instrs = append(fg.out.Instrs, bytecode.Const(t, x.Value))
	case *sema.FloatLit:
		t, err := fg.g.lowerValueType(x.T)
		if err != nil {
			return err
		}
		fg.out.Instrs = append(fg.out.Instrs, bytecode.Instr{Op: bytecode.OpConst, Type: t, Float: x.Value})
	case *sema.CharLit:
		t, err := fg.g.lowerValueType(x.T)
		if err != nil {
			return err
		}
		fg.out.Instrs = append(fg.out.Instrs, bytecode.Const(t, int64(x.Value)))
	case *sema.EnumRef:
		t, err := fg.g.lowerValueType(x.T)
		if err != nil {
			return err
		}
		var value int64
		if x.Enumerator != nil {
			value = x.Enumerator.Value
		}
		fg.out.Instrs = append(fg.out.Instrs, bytecode.Const(t, value))
	case *sema.VarRef:
		st, err := fg.storageForVar(x.Sym, x.T)
		if err != nil {
			return err
		}
		if st.kind != storageLocalSlot {
			return &Error{Pos: e.Pos().SourceStart, Node: fmt.Sprintf("%T", e), Op: "emitValue", Reason: "address storage expressions are not lowered in this task"}
		}
		fg.out.Instrs = append(fg.out.Instrs, bytecode.LoadLocal(st.typ, st.slot))
	case *sema.ImplicitCast:
		if err := fg.emitValue(x.X); err != nil {
			return err
		}
		from, err := fg.g.lowerValueType(x.From)
		if err != nil {
			return err
		}
		to, err := fg.g.lowerValueType(x.To)
		if err != nil {
			return err
		}
		if from != to {
			fg.out.Instrs = append(fg.out.Instrs, bytecode.Cast(from, to, castOpFor(x.Kind, from, to)))
		}
	case *sema.ExplicitCast:
		if err := fg.emitValue(x.X); err != nil {
			return err
		}
		from, err := fg.g.lowerValueType(x.X.GetType())
		if err != nil {
			return err
		}
		to, err := fg.g.lowerValueType(x.To)
		if err != nil {
			return err
		}
		if from != to {
			fg.out.Instrs = append(fg.out.Instrs, bytecode.Cast(from, to, castOpFor(sema.IntegralConversion, from, to)))
		}
	case *sema.BinOp:
		if err := fg.emitValue(x.L); err != nil {
			return err
		}
		if err := fg.emitValue(x.R); err != nil {
			return err
		}
		t, err := fg.g.lowerValueType(x.L.GetType())
		if err != nil {
			return err
		}
		fg.out.Instrs = append(fg.out.Instrs, bytecode.Binary(t, binaryOp(x.Op, t)))
	case *sema.AssignExpr:
		lhs, ok := x.L.(*sema.VarRef)
		if !ok {
			return &Error{Pos: e.Pos().SourceStart, Node: fmt.Sprintf("%T", e), Op: "emitValue", Reason: "only local variable assignment is lowered in this task"}
		}
		st, err := fg.storageForVar(lhs.Sym, lhs.T)
		if err != nil {
			return err
		}
		if st.kind != storageLocalSlot {
			return &Error{Pos: e.Pos().SourceStart, Node: fmt.Sprintf("%T", e), Op: "emitValue", Reason: "only local slot assignment is lowered in this task"}
		}
		if err := fg.emitValue(x.R); err != nil {
			return err
		}
		fg.out.Instrs = append(fg.out.Instrs, bytecode.Instr{Op: bytecode.OpDup}, bytecode.StoreLocal(st.typ, st.slot))
	default:
		return &Error{Pos: e.Pos().SourceStart, Node: fmt.Sprintf("%T", e), Op: "emitValue", Reason: "expression lowering is not implemented for this node"}
	}
	return nil
}

func exprLeavesValue(e sema.Expr) bool {
	if e == nil {
		return false
	}
	if b, ok := sema.Unqual(e.GetType()).(*sema.BuiltinType); ok && b.Kind == sema.Void {
		return false
	}
	return true
}

func binaryOp(op sema.BinaryOp, t bytecode.ValueType) bytecode.BinaryOp {
	switch op {
	case sema.OpAdd:
		return bytecode.BinAdd
	case sema.OpSub:
		return bytecode.BinSub
	case sema.OpMul:
		return bytecode.BinMul
	case sema.OpDiv:
		if isUnsignedType(t) {
			return bytecode.BinDivU
		}
		return bytecode.BinDivS
	case sema.OpMod:
		if isUnsignedType(t) {
			return bytecode.BinRemU
		}
		return bytecode.BinRemS
	case sema.OpAnd:
		return bytecode.BinAnd
	case sema.OpOr:
		return bytecode.BinOr
	case sema.OpXor:
		return bytecode.BinXor
	case sema.OpShl:
		return bytecode.BinShl
	case sema.OpShr:
		if isUnsignedType(t) {
			return bytecode.BinShrU
		}
		return bytecode.BinShrS
	case sema.OpEq:
		return bytecode.BinEq
	case sema.OpNe:
		return bytecode.BinNe
	case sema.OpLt:
		if isFloatType(t) {
			return bytecode.BinLtF
		}
		if isUnsignedType(t) {
			return bytecode.BinLtU
		}
		return bytecode.BinLtS
	case sema.OpLe:
		if isFloatType(t) {
			return bytecode.BinLeF
		}
		if isUnsignedType(t) {
			return bytecode.BinLeU
		}
		return bytecode.BinLeS
	case sema.OpGt:
		if isFloatType(t) {
			return bytecode.BinGtF
		}
		if isUnsignedType(t) {
			return bytecode.BinGtU
		}
		return bytecode.BinGtS
	case sema.OpGe:
		if isFloatType(t) {
			return bytecode.BinGeF
		}
		if isUnsignedType(t) {
			return bytecode.BinGeU
		}
		return bytecode.BinGeS
	}
	return bytecode.BinAdd
}
