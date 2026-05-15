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
		if st.kind == storageLocalSlot {
			fg.out.Instrs = append(fg.out.Instrs, bytecode.LoadLocal(st.typ, st.slot))
			return nil
		}
		return fg.emitLValueValue(x, x.T)
	case *sema.ImplicitCast:
		switch x.Kind {
		case sema.LValueToRValue:
			return fg.emitLValueValue(x.X, x.From)
		case sema.ArrayDecay:
			if err := fg.emitAddress(x.X); err != nil {
				return err
			}
			fg.out.Instrs = append(fg.out.Instrs, bytecode.Cast(bytecode.TypeObjectAddr, bytecode.TypePtr, bytecode.CastBit))
			return nil
		}
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
		return fg.emitAssign(x.L, x.R)
	case *sema.UnOp:
		switch x.Op {
		case sema.UnAddr:
			if err := fg.emitAddress(x.X); err != nil {
				return err
			}
			fg.out.Instrs = append(fg.out.Instrs, bytecode.Cast(bytecode.TypeObjectAddr, bytecode.TypePtr, bytecode.CastBit))
		case sema.UnDeref:
			return fg.emitLValueValue(x, x.T)
		default:
			return &Error{Pos: e.Pos().SourceStart, Node: fmt.Sprintf("%T", e), Op: "emitValue", Reason: "unary expression lowering is not implemented for this operator"}
		}
	case *sema.MemberExpr, *sema.IndexExpr, *sema.StringLit:
		return fg.emitLValueValue(e, e.GetType())
	default:
		return &Error{Pos: e.Pos().SourceStart, Node: fmt.Sprintf("%T", e), Op: "emitValue", Reason: "expression lowering is not implemented for this node"}
	}
	return nil
}

func (fg *funcGen) emitLValueValue(e sema.Expr, t sema.Type) error {
	vt, err := fg.g.lowerValueType(t)
	if err != nil {
		return err
	}
	if vr, ok := e.(*sema.VarRef); ok {
		st, err := fg.storageForVar(vr.Sym, vr.T)
		if err != nil {
			return err
		}
		if st.kind == storageLocalSlot {
			fg.out.Instrs = append(fg.out.Instrs, bytecode.LoadLocal(st.typ, st.slot))
			return nil
		}
	}
	if vt == bytecode.TypeObjectAddr {
		return fg.emitAddress(e)
	}
	if err := fg.emitAddress(e); err != nil {
		return err
	}
	fg.out.Instrs = append(fg.out.Instrs, bytecode.Load(vt, fg.g.alignof(t), isVolatile(t)))
	return nil
}

func (fg *funcGen) emitAssign(lhs, rhs sema.Expr) error {
	if vr, ok := lhs.(*sema.VarRef); ok {
		st, err := fg.storageForVar(vr.Sym, vr.T)
		if err != nil {
			return err
		}
		if st.kind == storageLocalSlot {
			if err := fg.emitValue(rhs); err != nil {
				return err
			}
			fg.out.Instrs = append(fg.out.Instrs, bytecode.Instr{Op: bytecode.OpDup}, bytecode.StoreLocal(st.typ, st.slot))
			return nil
		}
	}
	vt, err := fg.g.lowerValueType(lhs.GetType())
	if err != nil {
		return err
	}
	if vt == bytecode.TypeObjectAddr {
		if err := fg.emitAddress(lhs); err != nil {
			return err
		}
		if err := fg.emitAddress(rhs); err != nil {
			return err
		}
		fg.out.Instrs = append(fg.out.Instrs, bytecode.Instr{Op: bytecode.OpMemCopy, Size: fg.g.sizeof(lhs.GetType()), Align: fg.g.alignof(lhs.GetType()), Volatile: isVolatile(lhs.GetType())})
		return fg.emitAddress(lhs)
	}
	if err := fg.emitValue(rhs); err != nil {
		return err
	}
	fg.out.Instrs = append(fg.out.Instrs, bytecode.Instr{Op: bytecode.OpDup})
	if err := fg.emitAddress(lhs); err != nil {
		return err
	}
	fg.out.Instrs = append(fg.out.Instrs, bytecode.Instr{Op: bytecode.OpSwap})
	fg.out.Instrs = append(fg.out.Instrs, bytecode.Store(vt, fg.g.alignof(lhs.GetType()), isVolatile(lhs.GetType())))
	return nil
}

func (fg *funcGen) emitBinOp(x *sema.BinOp) error {
	resultType, err := fg.g.lowerValueType(x.T)
	if err != nil {
		return err
	}
	if isPointerArithmeticExpr(x) {
		return fg.emitPointerArithmetic(x)
	}
	if x.Op == sema.OpLAnd || x.Op == sema.OpLOr {
		return fg.emitLogical(x, resultType)
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

func (fg *funcGen) emitLogical(x *sema.BinOp, resultType bytecode.ValueType) error {
	shortLabel := fg.newLabel(true, nil)
	endLabel := fg.newLabel(false, []bytecode.ValueType{bytecode.TypeBool})
	if err := fg.emitBoolValue(x.L); err != nil {
		return err
	}
	if x.Op == sema.OpLAnd {
		fg.out.Instrs = append(fg.out.Instrs, bytecode.JumpIfZero(bytecode.TypeBool, shortLabel))
	} else {
		fg.out.Instrs = append(fg.out.Instrs, bytecode.JumpIfNonZero(bytecode.TypeBool, shortLabel))
	}
	if err := fg.emitBoolValue(x.R); err != nil {
		return err
	}
	fg.out.Instrs = append(fg.out.Instrs, bytecode.Jump(endLabel))
	fg.mark(shortLabel)
	if x.Op == sema.OpLAnd {
		fg.out.Instrs = append(fg.out.Instrs, bytecode.Const(bytecode.TypeBool, 0))
	} else {
		fg.out.Instrs = append(fg.out.Instrs, bytecode.Const(bytecode.TypeBool, 1))
	}
	fg.mark(endLabel)
	fg.emitCast(bytecode.TypeBool, resultType, sema.IntegralConversion)
	return nil
}

func (fg *funcGen) emitPointerArithmetic(x *sema.BinOp) error {
	leftType, err := fg.g.lowerValueType(x.L.GetType())
	if err != nil {
		return err
	}
	rightType, err := fg.g.lowerValueType(x.R.GetType())
	if err != nil {
		return err
	}
	switch {
	case x.Op == sema.OpAdd && isPointerType(leftType) && isIntegerType(rightType):
		if err := fg.emitValue(x.L); err != nil {
			return err
		}
		if err := fg.emitValue(x.R); err != nil {
			return err
		}
		fg.out.Instrs = append(fg.out.Instrs, bytecode.Instr{Op: bytecode.OpPtrAdd, Size: fg.g.elemSize(x.L.GetType())})
		return nil
	case x.Op == sema.OpAdd && isIntegerType(leftType) && isPointerType(rightType):
		if err := fg.emitValue(x.R); err != nil {
			return err
		}
		if err := fg.emitValue(x.L); err != nil {
			return err
		}
		fg.out.Instrs = append(fg.out.Instrs, bytecode.Instr{Op: bytecode.OpPtrAdd, Size: fg.g.elemSize(x.R.GetType())})
		return nil
	case x.Op == sema.OpSub && isPointerType(leftType) && isIntegerType(rightType):
		if err := fg.emitValue(x.L); err != nil {
			return err
		}
		if err := fg.emitValue(x.R); err != nil {
			return err
		}
		fg.out.Instrs = append(fg.out.Instrs, bytecode.Instr{Op: bytecode.OpUnary, Type: rightType, Unary: bytecode.UnaryNeg})
		fg.out.Instrs = append(fg.out.Instrs, bytecode.Instr{Op: bytecode.OpPtrAdd, Size: fg.g.elemSize(x.L.GetType())})
		return nil
	case x.Op == sema.OpSub && isPointerType(leftType) && isPointerType(rightType):
		return &Error{Pos: x.Pos().SourceStart, Node: fmt.Sprintf("%T", x), Op: "emitValue", Reason: "pointer difference lowering is not implemented"}
	default:
		return &Error{Pos: x.Pos().SourceStart, Node: fmt.Sprintf("%T", x), Op: "emitValue", Reason: "unsupported pointer arithmetic"}
	}
}

func isPointerArithmeticExpr(x *sema.BinOp) bool {
	if x == nil || (x.Op != sema.OpAdd && x.Op != sema.OpSub) {
		return false
	}
	_, leftPtr := sema.Unqual(x.L.GetType()).(*sema.PointerType)
	_, rightPtr := sema.Unqual(x.R.GetType()).(*sema.PointerType)
	return leftPtr || rightPtr
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
