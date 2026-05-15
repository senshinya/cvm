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
		fg.emitCast(from, to, x.Kind)
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
		fg.emitCast(from, to, sema.IntegralConversion)
	case *sema.BinOp:
		return fg.emitBinOp(x)
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

func (fg *funcGen) emitBinOp(x *sema.BinOp) error {
	resultType, err := fg.g.lowerValueType(x.T)
	if err != nil {
		return err
	}
	if x.Op == sema.OpLAnd || x.Op == sema.OpLOr {
		if err := fg.emitBoolValue(x.L); err != nil {
			return err
		}
		if err := fg.emitBoolValue(x.R); err != nil {
			return err
		}
		op := bytecode.BinAnd
		if x.Op == sema.OpLOr {
			op = bytecode.BinOr
		}
		fg.out.Instrs = append(fg.out.Instrs, bytecode.Binary(bytecode.TypeBool, op))
		fg.emitCast(bytecode.TypeBool, resultType, sema.IntegralConversion)
		return nil
	}

	leftType, err := fg.g.lowerValueType(x.L.GetType())
	if err != nil {
		return err
	}
	if err := fg.emitValue(x.L); err != nil {
		return err
	}
	if err := fg.emitValue(x.R); err != nil {
		return err
	}
	rightType, err := fg.g.lowerValueType(x.R.GetType())
	if err != nil {
		return err
	}
	if x.Op == sema.OpShl || x.Op == sema.OpShr {
		fg.emitCast(rightType, leftType, sema.IntegralConversion)
	}
	op, err := binaryOp(x.Op, leftType)
	if err != nil {
		return &Error{Pos: x.Pos().SourceStart, Node: fmt.Sprintf("%T", x), Op: "emitValue", Reason: err.Error()}
	}
	fg.out.Instrs = append(fg.out.Instrs, bytecode.Binary(leftType, op))
	if isCompareOp(x.Op) {
		fg.emitCast(bytecode.TypeBool, resultType, sema.IntegralConversion)
	}
	return nil
}

func (fg *funcGen) emitBoolValue(e sema.Expr) error {
	if err := fg.emitValue(e); err != nil {
		return err
	}
	from, err := fg.g.lowerValueType(e.GetType())
	if err != nil {
		return err
	}
	fg.emitCast(from, bytecode.TypeBool, sema.BoolConversion)
	return nil
}

func (fg *funcGen) emitCast(from, to bytecode.ValueType, kind sema.CastKind) {
	if from == to {
		return
	}
	fg.out.Instrs = append(fg.out.Instrs, bytecode.Cast(from, to, castOpFor(kind, from, to)))
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

func binaryOp(op sema.BinaryOp, t bytecode.ValueType) (bytecode.BinaryOp, error) {
	switch op {
	case sema.OpAdd:
		return bytecode.BinAdd, nil
	case sema.OpSub:
		return bytecode.BinSub, nil
	case sema.OpMul:
		return bytecode.BinMul, nil
	case sema.OpDiv:
		if isUnsignedType(t) {
			return bytecode.BinDivU, nil
		}
		return bytecode.BinDivS, nil
	case sema.OpMod:
		if isUnsignedType(t) {
			return bytecode.BinRemU, nil
		}
		return bytecode.BinRemS, nil
	case sema.OpAnd:
		return bytecode.BinAnd, nil
	case sema.OpOr:
		return bytecode.BinOr, nil
	case sema.OpXor:
		return bytecode.BinXor, nil
	case sema.OpShl:
		return bytecode.BinShl, nil
	case sema.OpShr:
		if isUnsignedType(t) {
			return bytecode.BinShrU, nil
		}
		return bytecode.BinShrS, nil
	case sema.OpEq:
		return bytecode.BinEq, nil
	case sema.OpNe:
		return bytecode.BinNe, nil
	case sema.OpLt:
		if isFloatType(t) {
			return bytecode.BinLtF, nil
		}
		if isUnsignedType(t) {
			return bytecode.BinLtU, nil
		}
		return bytecode.BinLtS, nil
	case sema.OpLe:
		if isFloatType(t) {
			return bytecode.BinLeF, nil
		}
		if isUnsignedType(t) {
			return bytecode.BinLeU, nil
		}
		return bytecode.BinLeS, nil
	case sema.OpGt:
		if isFloatType(t) {
			return bytecode.BinGtF, nil
		}
		if isUnsignedType(t) {
			return bytecode.BinGtU, nil
		}
		return bytecode.BinGtS, nil
	case sema.OpGe:
		if isFloatType(t) {
			return bytecode.BinGeF, nil
		}
		if isUnsignedType(t) {
			return bytecode.BinGeU, nil
		}
		return bytecode.BinGeS, nil
	}
	return 0, fmt.Errorf("unsupported binary operator %d", op)
}

func isCompareOp(op sema.BinaryOp) bool {
	switch op {
	case sema.OpEq, sema.OpNe, sema.OpLt, sema.OpLe, sema.OpGt, sema.OpGe:
		return true
	default:
		return false
	}
}
