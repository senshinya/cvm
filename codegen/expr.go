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
		if isFunctionDesignator(x) {
			return fg.emitFunctionAddress(x)
		}
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
			if isFunctionExpr(x.X) {
				return fg.emitFunctionAddress(x.X)
			}
			return fg.emitLValueValue(x.X, x.From)
		case sema.ArrayDecay:
			if err := fg.emitAddress(x.X); err != nil {
				return err
			}
			fg.out.Instrs = append(fg.out.Instrs, bytecode.Cast(bytecode.TypeObjectAddr, bytecode.TypePtr, bytecode.CastBit))
			return nil
		case sema.FunctionDecay:
			return fg.emitFunctionDecay(x.X)
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
		if b, ok := sema.Unqual(x.To).(*sema.BuiltinType); ok && b.Kind == sema.Void {
			fg.out.Instrs = append(fg.out.Instrs, bytecode.Instr{Op: bytecode.OpPop})
			return nil
		}
		from, err := fg.g.lowerValueType(x.X.GetType())
		if err != nil {
			return err
		}
		to, err := fg.g.lowerValueType(x.To)
		if err != nil {
			return err
		}
		kind := sema.IntegralConversion
		if to == bytecode.TypeBool {
			kind = sema.BoolConversion
		}
		fg.emitCast(from, to, kind)
	case *sema.BinOp:
		return fg.emitBinOp(x)
	case *sema.AssignExpr:
		return fg.emitAssign(x.L, x.R)
	case *sema.CompoundAssign:
		return fg.emitCompoundAssign(x)
	case *sema.CallExpr:
		return fg.emitCall(x)
	case *sema.CommaExpr:
		if err := fg.emitValue(x.L); err != nil {
			return err
		}
		if exprLeavesValue(x.L) {
			fg.out.Instrs = append(fg.out.Instrs, bytecode.Instr{Op: bytecode.OpPop})
		}
		return fg.emitValue(x.R)
	case *sema.CondExpr:
		return fg.emitCondExpr(x)
	case *sema.StmtExpr:
		return fg.emitStmtExpr(x)
	case *sema.SizeofExpr:
		return fg.emitSizeof(x)
	case *sema.UnOp:
		switch x.Op {
		case sema.UnAddr:
			if isFunctionExpr(x.X) {
				return fg.emitFunctionAddress(x.X)
			}
			if err := fg.emitAddress(x.X); err != nil {
				return err
			}
			fg.out.Instrs = append(fg.out.Instrs, bytecode.Cast(bytecode.TypeObjectAddr, bytecode.TypePtr, bytecode.CastBit))
		case sema.UnDeref:
			return fg.emitLValueValue(x, x.T)
		case sema.UnPlus:
			return fg.emitValue(x.X)
		case sema.UnMinus:
			if err := fg.emitValue(x.X); err != nil {
				return err
			}
			t, err := fg.g.lowerValueType(x.T)
			if err != nil {
				return err
			}
			fg.out.Instrs = append(fg.out.Instrs, bytecode.Instr{Op: bytecode.OpUnary, Type: t, Unary: bytecode.UnaryNeg})
		case sema.UnBitNot:
			if err := fg.emitValue(x.X); err != nil {
				return err
			}
			t, err := fg.g.lowerValueType(x.T)
			if err != nil {
				return err
			}
			fg.out.Instrs = append(fg.out.Instrs, bytecode.Const(t, -1), bytecode.Binary(t, bytecode.BinXor))
		case sema.UnLogNot:
			if err := fg.emitBoolValue(x.X); err != nil {
				return err
			}
			t, err := fg.g.lowerValueType(x.T)
			if err != nil {
				return err
			}
			fg.out.Instrs = append(fg.out.Instrs,
				bytecode.Const(bytecode.TypeBool, 0),
				bytecode.Binary(bytecode.TypeBool, bytecode.BinEq),
			)
			fg.emitCast(bytecode.TypeBool, t, sema.IntegralConversion)
		case sema.UnIncPre, sema.UnIncPost, sema.UnDecPre, sema.UnDecPost:
			return fg.emitIncDec(x)
		default:
			return &Error{Pos: e.Pos().SourceStart, Node: fmt.Sprintf("%T", e), Op: "emitValue", Reason: "unary expression lowering is not implemented for this operator"}
		}
	case *sema.MemberExpr, *sema.IndexExpr, *sema.StringLit, *sema.CompoundLit:
		return fg.emitLValueValue(e, e.GetType())
	default:
		return &Error{Pos: e.Pos().SourceStart, Node: fmt.Sprintf("%T", e), Op: "emitValue", Reason: "expression lowering is not implemented for this node"}
	}
	return nil
}

func (fg *funcGen) emitCompoundAssign(x *sema.CompoundAssign) error {
	vr, ok := x.L.(*sema.VarRef)
	if !ok {
		return fg.emitAddressableCompoundAssign(x)
	}
	st, err := fg.storageForVar(vr.Sym, vr.T)
	if err != nil {
		return err
	}
	if st.kind != storageLocalSlot {
		return fg.emitAddressableCompoundAssign(x)
	}
	if !isIntegerType(st.typ) {
		return &Error{Pos: x.Pos().SourceStart, Node: fmt.Sprintf("%T", x), Op: "emitValue", Reason: fmt.Sprintf("compound assignment lowering is not implemented for %s", st.typ)}
	}

	computeType := compoundIntegerType(st.typ)
	fg.out.Instrs = append(fg.out.Instrs, bytecode.LoadLocal(st.typ, st.slot))
	fg.emitCast(st.typ, computeType, sema.IntegralConversion)
	if err := fg.emitValue(x.R); err != nil {
		return err
	}
	rt, err := fg.g.lowerValueType(x.R.GetType())
	if err != nil {
		return err
	}
	fg.emitCast(rt, computeType, sema.IntegralConversion)
	op, err := binaryOp(x.Op, computeType)
	if err != nil {
		return &Error{Pos: x.Pos().SourceStart, Node: fmt.Sprintf("%T", x), Op: "emitValue", Reason: err.Error()}
	}
	fg.out.Instrs = append(fg.out.Instrs, bytecode.Binary(computeType, op))
	fg.emitCast(computeType, st.typ, sema.IntegralConversion)
	fg.out.Instrs = append(fg.out.Instrs, bytecode.Instr{Op: bytecode.OpDup}, bytecode.StoreLocal(st.typ, st.slot))
	return nil
}

func (fg *funcGen) emitAddressableCompoundAssign(x *sema.CompoundAssign) error {
	vt, err := fg.g.lowerValueType(x.L.GetType())
	if err != nil {
		return err
	}
	if !isIntegerType(vt) {
		return &Error{Pos: x.Pos().SourceStart, Node: fmt.Sprintf("%T", x), Op: "emitValue", Reason: fmt.Sprintf("compound assignment lowering is not implemented for %s", vt)}
	}
	computeType := compoundIntegerType(vt)
	addrSlot := fg.allocSyntheticSlot(".compound.assign.addr", bytecode.TypeObjectAddr)
	valueSlot := fg.allocSyntheticSlot(".compound.assign.value", vt)
	if err := fg.emitAddress(x.L); err != nil {
		return err
	}
	fg.out.Instrs = append(fg.out.Instrs,
		bytecode.StoreLocal(bytecode.TypeObjectAddr, addrSlot),
		bytecode.LoadLocal(bytecode.TypeObjectAddr, addrSlot),
		bytecode.Load(vt, fg.loadStoreAlign(x.L, x.L.GetType()), isVolatile(x.L.GetType())),
	)
	fg.emitCast(vt, computeType, sema.IntegralConversion)
	if err := fg.emitValue(x.R); err != nil {
		return err
	}
	rt, err := fg.g.lowerValueType(x.R.GetType())
	if err != nil {
		return err
	}
	fg.emitCast(rt, computeType, sema.IntegralConversion)
	op, err := binaryOp(x.Op, computeType)
	if err != nil {
		return &Error{Pos: x.Pos().SourceStart, Node: fmt.Sprintf("%T", x), Op: "emitValue", Reason: err.Error()}
	}
	fg.out.Instrs = append(fg.out.Instrs, bytecode.Binary(computeType, op))
	fg.emitCast(computeType, vt, sema.IntegralConversion)
	fg.out.Instrs = append(fg.out.Instrs,
		bytecode.Instr{Op: bytecode.OpDup},
		bytecode.StoreLocal(vt, valueSlot),
		bytecode.LoadLocal(bytecode.TypeObjectAddr, addrSlot),
		bytecode.Instr{Op: bytecode.OpSwap},
		bytecode.Store(vt, fg.loadStoreAlign(x.L, x.L.GetType()), isVolatile(x.L.GetType())),
		bytecode.LoadLocal(vt, valueSlot),
	)
	return nil
}

func compoundIntegerType(t bytecode.ValueType) bytecode.ValueType {
	switch t {
	case bytecode.TypeBool, bytecode.TypeI8, bytecode.TypeI16, bytecode.TypeU8, bytecode.TypeU16:
		return bytecode.TypeI32
	default:
		return t
	}
}

func (fg *funcGen) emitCondExpr(x *sema.CondExpr) error {
	var stack []bytecode.ValueType
	if exprLeavesValue(x) {
		t, err := fg.g.lowerValueType(x.T)
		if err != nil {
			return err
		}
		stack = []bytecode.ValueType{t}
	}
	elseLabel := fg.newLabel(true, nil)
	endLabel := fg.newLabel(false, stack)
	if err := fg.emitBoolValue(x.Cond); err != nil {
		return err
	}
	fg.out.Instrs = append(fg.out.Instrs, bytecode.JumpIfZero(bytecode.TypeBool, elseLabel))
	if err := fg.emitValue(x.Then); err != nil {
		return err
	}
	fg.out.Instrs = append(fg.out.Instrs, bytecode.Jump(endLabel))
	fg.mark(elseLabel)
	if err := fg.emitValue(x.Else); err != nil {
		return err
	}
	fg.mark(endLabel)
	return nil
}

func (fg *funcGen) emitStmtExpr(x *sema.StmtExpr) error {
	if x.Block == nil {
		return nil
	}
	if !exprLeavesValue(x) {
		return fg.emitStmt(x.Block)
	}
	scopeMark := len(fg.activeDynamicObjects)
	items := x.Block.Items
	last := len(items) - 1
	for i, item := range items {
		if i == last {
			if exprStmt, ok := item.(*sema.ExprStmt); ok && exprStmt.Expr != nil {
				if err := fg.emitValue(exprStmt.Expr); err != nil {
					return err
				}
				if !fg.lastInstrTerminal() {
					fg.popDynamicObjectScope(scopeMark)
				} else {
					fg.activeDynamicObjects = fg.activeDynamicObjects[:scopeMark]
				}
				return nil
			}
		}
		if err := fg.emitStmt(item); err != nil {
			return err
		}
	}
	if !fg.lastInstrTerminal() {
		fg.popDynamicObjectScope(scopeMark)
	} else {
		fg.activeDynamicObjects = fg.activeDynamicObjects[:scopeMark]
	}
	return nil
}

func (fg *funcGen) emitIncDec(x *sema.UnOp) error {
	if vr, ok := x.X.(*sema.VarRef); ok {
		st, err := fg.storageForVar(vr.Sym, vr.T)
		if err != nil {
			return err
		}
		if st.kind == storageLocalSlot {
			return fg.emitLocalSlotIncDec(x, st)
		}
	}
	return fg.emitAddressableIncDec(x)
}

func (fg *funcGen) emitLocalSlotIncDec(x *sema.UnOp, st storage) error {
	post := x.Op == sema.UnIncPost || x.Op == sema.UnDecPost
	dec := x.Op == sema.UnDecPre || x.Op == sema.UnDecPost
	fg.out.Instrs = append(fg.out.Instrs, bytecode.LoadLocal(st.typ, st.slot))
	if post {
		fg.out.Instrs = append(fg.out.Instrs, bytecode.Instr{Op: bytecode.OpDup})
	}
	if isPointerType(st.typ) {
		step := int64(1)
		if dec {
			step = -1
		}
		fg.out.Instrs = append(fg.out.Instrs, bytecode.Const(bytecode.TypeI32, step))
		if err := fg.emitPtrAddForExpr(x.X, x.X.GetType()); err != nil {
			return err
		}
	} else if isIntegerType(st.typ) {
		computeType := compoundIntegerType(st.typ)
		fg.emitCast(st.typ, computeType, sema.IntegralConversion)
		fg.out.Instrs = append(fg.out.Instrs, bytecode.Const(computeType, 1))
		op := bytecode.BinAdd
		if dec {
			op = bytecode.BinSub
		}
		fg.out.Instrs = append(fg.out.Instrs, bytecode.Binary(computeType, op))
		fg.emitCast(computeType, st.typ, sema.IntegralConversion)
	} else {
		return &Error{Pos: x.Pos().SourceStart, Node: fmt.Sprintf("%T", x), Op: "emitValue", Reason: fmt.Sprintf("++/-- lowering is not implemented for %s", st.typ)}
	}
	if !post {
		fg.out.Instrs = append(fg.out.Instrs, bytecode.Instr{Op: bytecode.OpDup})
	}
	fg.out.Instrs = append(fg.out.Instrs, bytecode.StoreLocal(st.typ, st.slot))
	return nil
}

func (fg *funcGen) emitAddressableIncDec(x *sema.UnOp) error {
	vt, err := fg.g.lowerValueType(x.X.GetType())
	if err != nil {
		return err
	}
	if !isIntegerType(vt) && !isPointerType(vt) {
		return &Error{Pos: x.Pos().SourceStart, Node: fmt.Sprintf("%T", x), Op: "emitValue", Reason: fmt.Sprintf("++/-- lowering is not implemented for %s", vt)}
	}
	addrSlot := fg.allocSyntheticSlot(".incdec.addr", bytecode.TypeObjectAddr)
	valueSlot := fg.allocSyntheticSlot(".incdec.value", vt)
	if err := fg.emitAddress(x.X); err != nil {
		return err
	}
	fg.out.Instrs = append(fg.out.Instrs, bytecode.StoreLocal(bytecode.TypeObjectAddr, addrSlot))
	fg.out.Instrs = append(fg.out.Instrs,
		bytecode.LoadLocal(bytecode.TypeObjectAddr, addrSlot),
		bytecode.Load(vt, fg.loadStoreAlign(x.X, x.X.GetType()), isVolatile(x.X.GetType())),
		bytecode.StoreLocal(vt, valueSlot),
		bytecode.LoadLocal(vt, valueSlot),
	)
	if err := fg.emitIncDecOperation(x, vt); err != nil {
		return err
	}
	if x.Op == sema.UnIncPre || x.Op == sema.UnDecPre {
		fg.out.Instrs = append(fg.out.Instrs, bytecode.Instr{Op: bytecode.OpDup}, bytecode.StoreLocal(vt, valueSlot))
	}
	fg.out.Instrs = append(fg.out.Instrs,
		bytecode.LoadLocal(bytecode.TypeObjectAddr, addrSlot),
		bytecode.Instr{Op: bytecode.OpSwap},
		bytecode.Store(vt, fg.loadStoreAlign(x.X, x.X.GetType()), isVolatile(x.X.GetType())),
		bytecode.LoadLocal(vt, valueSlot),
	)
	return nil
}

func (fg *funcGen) emitIncDecOperation(x *sema.UnOp, typ bytecode.ValueType) error {
	dec := x.Op == sema.UnDecPre || x.Op == sema.UnDecPost
	if isPointerType(typ) {
		step := int64(1)
		if dec {
			step = -1
		}
		fg.out.Instrs = append(fg.out.Instrs, bytecode.Const(bytecode.TypeI32, step))
		return fg.emitPtrAddForExpr(x.X, x.X.GetType())
	}
	if isIntegerType(typ) {
		computeType := compoundIntegerType(typ)
		fg.emitCast(typ, computeType, sema.IntegralConversion)
		fg.out.Instrs = append(fg.out.Instrs, bytecode.Const(computeType, 1))
		op := bytecode.BinAdd
		if dec {
			op = bytecode.BinSub
		}
		fg.out.Instrs = append(fg.out.Instrs, bytecode.Binary(computeType, op))
		fg.emitCast(computeType, typ, sema.IntegralConversion)
		return nil
	}
	return &Error{Pos: x.Pos().SourceStart, Node: fmt.Sprintf("%T", x), Op: "emitValue", Reason: fmt.Sprintf("++/-- lowering is not implemented for %s", typ)}
}

func (fg *funcGen) emitCall(x *sema.CallExpr) error {
	if name := tgmathPseudoCallName(x.Callee); name != "" {
		return fg.emitTgmathCall(x, name)
	}
	ft, err := functionTypeFromCallee(x.Callee.GetType())
	if err != nil {
		return &Error{Pos: x.Pos().SourceStart, Node: fmt.Sprintf("%T", x), Op: "emitValue", Reason: err.Error()}
	}
	sig, err := fg.g.lowerFuncSig(ft)
	if err != nil {
		return err
	}
	if global := directCallGlobal(x.Callee); global >= 0 {
		for _, arg := range x.Args {
			if err := fg.emitValue(arg); err != nil {
				return err
			}
		}
		fg.out.Instrs = append(fg.out.Instrs, bytecode.Call(global, sig, len(x.Args)))
		return nil
	}
	if err := fg.emitValue(x.Callee); err != nil {
		return err
	}
	for _, arg := range x.Args {
		if err := fg.emitValue(arg); err != nil {
			return err
		}
	}
	fg.out.Instrs = append(fg.out.Instrs, bytecode.Instr{Op: bytecode.OpCallIndirect, Sig: sig, Argc: len(x.Args)})
	return nil
}

func (fg *funcGen) emitTgmathCall(x *sema.CallExpr, pseudo string) error {
	ret, err := fg.g.lowerValueType(x.GetType())
	if err != nil {
		return err
	}
	params := make([]bytecode.ValueType, 0, len(x.Args))
	for _, arg := range x.Args {
		if err := fg.emitValue(arg); err != nil {
			return err
		}
		pt, err := fg.g.lowerValueType(arg.GetType())
		if err != nil {
			return err
		}
		params = append(params, pt)
	}
	global := fg.g.syntheticExtern(tgmathExternName(pseudo, x), ret, params)
	sig := fg.g.internSig(ret, params, false)
	fg.out.Instrs = append(fg.out.Instrs, bytecode.Call(global, sig, len(x.Args)))
	return nil
}

func tgmathPseudoCallName(e sema.Expr) string {
	vr := functionVarRef(e)
	if vr == nil || vr.Sym == nil {
		return ""
	}
	switch vr.Sym.Name {
	case "__cvm_tgmath_sin", "__cvm_tgmath_exp", "__cvm_tgmath_pow":
		return vr.Sym.Name
	default:
		return ""
	}
}

func tgmathExternName(pseudo string, x *sema.CallExpr) string {
	base := pseudo
	if tgmathCallIsComplex(x) {
		switch pseudo {
		case "__cvm_tgmath_sin":
			base = "__cvm_tgmath_csin"
		case "__cvm_tgmath_exp":
			base = "__cvm_tgmath_cexp"
		case "__cvm_tgmath_pow":
			base = "__cvm_tgmath_cpow"
		}
	}
	switch tgmathCallRank(x) {
	case sema.Float, sema.FloatComplex:
		return base + "f"
	case sema.LongDouble, sema.LongDoubleComplex:
		return base + "l"
	default:
		return base
	}
}

func tgmathCallIsComplex(x *sema.CallExpr) bool {
	for _, arg := range x.Args {
		if bt, ok := sema.Unqual(arg.GetType()).(*sema.BuiltinType); ok {
			switch bt.Kind {
			case sema.FloatComplex, sema.DoubleComplex, sema.LongDoubleComplex:
				return true
			}
		}
	}
	return false
}

func tgmathCallRank(x *sema.CallExpr) sema.BuiltinKind {
	rank := sema.Double
	allFloat := len(x.Args) > 0
	for _, arg := range x.Args {
		bt, ok := sema.Unqual(arg.GetType()).(*sema.BuiltinType)
		if !ok {
			allFloat = false
			continue
		}
		switch bt.Kind {
		case sema.LongDouble, sema.LongDoubleComplex:
			return bt.Kind
		case sema.Double, sema.DoubleComplex:
			rank = bt.Kind
			allFloat = false
		case sema.Float, sema.FloatComplex:
		default:
			allFloat = false
		}
	}
	if allFloat {
		if tgmathCallIsComplex(x) {
			return sema.FloatComplex
		}
		return sema.Float
	}
	return rank
}

func (fg *funcGen) emitSizeof(x *sema.SizeofExpr) error {
	outType, err := fg.g.lowerValueType(x.T)
	if err != nil {
		return err
	}
	t := x.Operand.Type
	if t == nil && x.Operand.Expr != nil {
		t = x.Operand.Expr.GetType()
	}
	if t == nil {
		return &Error{Pos: x.Pos().SourceStart, Node: fmt.Sprintf("%T", x), Op: "emitValue", Reason: "sizeof operand has no type"}
	}
	if !typeHasVariableSize(t) {
		fg.out.Instrs = append(fg.out.Instrs, bytecode.Const(outType, fg.g.sizeof(t)))
		return nil
	}
	if x.Operand.Expr != nil {
		if sym := dynamicObjectSymbol(x.Operand.Expr); sym != nil {
			if object, ok := fg.dynamicObjectMap[sym]; ok {
				fg.out.Instrs = append(fg.out.Instrs, bytecode.Instr{Op: bytecode.OpDynamicObjectAddr, Object: object, Type: bytecode.TypeObjectAddr})
				fg.out.Instrs = append(fg.out.Instrs, bytecode.Instr{Op: bytecode.OpPop})
			}
			if slot, ok := fg.dynamicSizeSlotForSymbol(sym, t); ok {
				fg.out.Instrs = append(fg.out.Instrs, bytecode.LoadLocal(bytecode.TypeI64, slot))
				fg.emitCast(bytecode.TypeI64, outType, sema.IntegralConversion)
				return nil
			}
			if slot, ok := fg.dynamicSizeSlotMap[sym]; ok && sameUnqualType(t, x.Operand.Expr.GetType()) {
				fg.out.Instrs = append(fg.out.Instrs, bytecode.LoadLocal(bytecode.TypeI64, slot))
				fg.emitCast(bytecode.TypeI64, outType, sema.IntegralConversion)
				return nil
			}
		}
		if sym := pointerSymbol(x.Operand.Expr); sym != nil {
			if slot, ok := fg.dynamicSizeSlotForPointerSymbol(sym, t); ok {
				fg.out.Instrs = append(fg.out.Instrs, bytecode.LoadLocal(bytecode.TypeI64, slot))
				fg.emitCast(bytecode.TypeI64, outType, sema.IntegralConversion)
				return nil
			}
		}
	}
	if err := fg.emitRuntimeSizeof(t); err != nil {
		return err
	}
	fg.emitCast(bytecode.TypeI64, outType, sema.IntegralConversion)
	return nil
}

func (fg *funcGen) emitRuntimeSizeof(t sema.Type) error {
	switch x := sema.Unqual(t).(type) {
	case *sema.ArrayType:
		if x.SizeKind == sema.ArrayStarSize {
			return fmt.Errorf("cannot lower runtime sizeof for star-sized array")
		}
		if x.SizeKind == sema.ArrayVLA {
			if x.SizeExpr == nil {
				return fmt.Errorf("VLA type has no bound expression")
			}
			if err := fg.emitValue(x.SizeExpr); err != nil {
				return err
			}
			from, err := fg.g.lowerValueType(x.SizeExpr.GetType())
			if err != nil {
				return err
			}
			fg.emitCast(from, bytecode.TypeI64, sema.IntegralConversion)
		} else {
			fg.out.Instrs = append(fg.out.Instrs, bytecode.I64Const(x.Size))
		}
		if err := fg.emitRuntimeSizeof(x.Elem); err != nil {
			return err
		}
		fg.out.Instrs = append(fg.out.Instrs, bytecode.Binary(bytecode.TypeI64, bytecode.BinMul))
		return nil
	default:
		fg.out.Instrs = append(fg.out.Instrs, bytecode.I64Const(fg.g.sizeof(t)))
		return nil
	}
}

func (fg *funcGen) emitFunctionAddress(e sema.Expr) error {
	vr := functionVarRef(e)
	if vr == nil || vr.Sym == nil || vr.Sym.GlobalID < 0 {
		return &Error{Pos: e.Pos().SourceStart, Node: fmt.Sprintf("%T", e), Op: "emitValue", Reason: "function address is not backed by a global symbol"}
	}
	global := vr.Sym.GlobalID
	if global >= len(fg.g.mod.Globals) {
		return &Error{Pos: e.Pos().SourceStart, Node: fmt.Sprintf("%T", e), Op: "emitValue", Reason: fmt.Sprintf("function global %d is missing", global)}
	}
	fg.out.Instrs = append(fg.out.Instrs, bytecode.AddrFunc(global))
	return nil
}

func (fg *funcGen) emitFunctionDecay(e sema.Expr) error {
	if isFunctionExpr(e) {
		return fg.emitFunctionAddress(e)
	}
	if inner := dereferencedFunctionPointer(e); inner != nil {
		return fg.emitValue(inner)
	}
	if err := fg.emitValue(e); err != nil {
		return err
	}
	return nil
}

func (fg *funcGen) emitLValueValue(e sema.Expr, t sema.Type) error {
	if isFunctionExpr(e) {
		return fg.emitFunctionAddress(e)
	}
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
	if member, ok := e.(*sema.MemberExpr); ok && member.Field != nil && member.Field.IsBitField {
		addr, err := fg.bitFieldAddress(member)
		if err != nil {
			return err
		}
		if err := addr.emit(); err != nil {
			return err
		}
		fg.out.Instrs = append(fg.out.Instrs, bytecode.Instr{Op: bytecode.OpBitFieldLoad, Type: vt, Layout: addr.layout, Field: addr.field, Volatile: addr.volatile})
		return nil
	}
	if err := fg.emitAddress(e); err != nil {
		return err
	}
	fg.out.Instrs = append(fg.out.Instrs, bytecode.Load(vt, fg.loadStoreAlign(e, t), isVolatile(t)))
	return nil
}

func (fg *funcGen) emitAssign(lhs, rhs sema.Expr) error {
	if member, ok := lhs.(*sema.MemberExpr); ok && member.Field != nil && member.Field.IsBitField {
		return fg.emitBitFieldAssign(member, rhs)
	}
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
		if isComplexType(lhs.GetType()) {
			if err := fg.emitComplexAssign(lhs, rhs); err != nil {
				return err
			}
			return fg.emitAddress(lhs)
		}
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
	fg.out.Instrs = append(fg.out.Instrs, bytecode.Store(vt, fg.loadStoreAlign(lhs, lhs.GetType()), isVolatile(lhs.GetType())))
	return nil
}

func (fg *funcGen) emitComplexAssign(lhs, rhs sema.Expr) error {
	dst := address{emit: func() error {
		return fg.emitAddress(lhs)
	}}
	return fg.emitComplexInitializer(dst, rhs, lhs.GetType())
}

func (fg *funcGen) loadStoreAlign(e sema.Expr, t sema.Type) int64 {
	if _, ok := e.(*sema.MemberExpr); ok {
		return 1
	}
	return fg.g.alignof(t)
}

func (fg *funcGen) emitBitFieldAssign(lhs *sema.MemberExpr, rhs sema.Expr) error {
	addr, err := fg.bitFieldAddress(lhs)
	if err != nil {
		return err
	}
	if err := fg.emitValue(rhs); err != nil {
		return err
	}
	fg.out.Instrs = append(fg.out.Instrs, bytecode.Instr{Op: bytecode.OpDup})
	if err := addr.emit(); err != nil {
		return err
	}
	fg.out.Instrs = append(fg.out.Instrs, bytecode.Instr{Op: bytecode.OpSwap})
	fg.out.Instrs = append(fg.out.Instrs, bytecode.Instr{Op: bytecode.OpBitFieldStore, Type: addr.valueType, Layout: addr.layout, Field: addr.field, Volatile: addr.volatile})
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
		if _, err := fg.emitPtrIndexValue(x.R); err != nil {
			return err
		}
		return fg.emitPtrAddForExpr(x.L, x.L.GetType())
	case x.Op == sema.OpAdd && isIntegerType(leftType) && isPointerType(rightType):
		if err := fg.emitValue(x.R); err != nil {
			return err
		}
		if _, err := fg.emitPtrIndexValue(x.L); err != nil {
			return err
		}
		return fg.emitPtrAddForExpr(x.R, x.R.GetType())
	case x.Op == sema.OpSub && isPointerType(leftType) && isIntegerType(rightType):
		if err := fg.emitValue(x.L); err != nil {
			return err
		}
		idxType, err := fg.emitPtrIndexValue(x.R)
		if err != nil {
			return err
		}
		fg.out.Instrs = append(fg.out.Instrs, bytecode.Instr{Op: bytecode.OpUnary, Type: idxType, Unary: bytecode.UnaryNeg})
		return fg.emitPtrAddForExpr(x.L, x.L.GetType())
	case x.Op == sema.OpSub && isPointerType(leftType) && isPointerType(rightType):
		if err := fg.emitValue(x.L); err != nil {
			return err
		}
		if err := fg.emitValue(x.R); err != nil {
			return err
		}
		if slot, ok := fg.dynamicElemSizeSlotForExpr(x.L, x.L.GetType()); ok {
			fg.out.Instrs = append(fg.out.Instrs,
				bytecode.Instr{Op: bytecode.OpPtrDiff, Size: 1},
				bytecode.LoadLocal(bytecode.TypeI64, slot),
				bytecode.Binary(bytecode.TypeI64, bytecode.BinDivS),
			)
			return nil
		}
		size := fg.g.elemSize(x.L.GetType())
		if size <= 0 {
			return fmt.Errorf("cannot lower pointer difference with zero element size for %s", x.L.GetType())
		}
		fg.out.Instrs = append(fg.out.Instrs, bytecode.Instr{Op: bytecode.OpPtrDiff, Size: size})
		return nil
	default:
		return &Error{Pos: x.Pos().SourceStart, Node: fmt.Sprintf("%T", x), Op: "emitValue", Reason: "unsupported pointer arithmetic"}
	}
}

func (fg *funcGen) emitPtrIndexValue(e sema.Expr) (bytecode.ValueType, error) {
	if err := fg.emitValue(e); err != nil {
		return bytecode.TypeVoid, err
	}
	from, err := fg.g.lowerValueType(e.GetType())
	if err != nil {
		return bytecode.TypeVoid, err
	}
	to := ptrIndexValueType(from)
	fg.emitCast(from, to, sema.IntegralConversion)
	return to, nil
}

func ptrIndexValueType(t bytecode.ValueType) bytecode.ValueType {
	switch t {
	case bytecode.TypeBool, bytecode.TypeI8, bytecode.TypeI16, bytecode.TypeU8, bytecode.TypeU16:
		return bytecode.TypeI32
	default:
		return t
	}
}

func (fg *funcGen) emitPtrAddFor(baseType sema.Type) error {
	return fg.emitPtrAddForExpr(nil, baseType)
}

func (fg *funcGen) emitPtrAddForExpr(base sema.Expr, baseType sema.Type) error {
	if slot, ok := fg.dynamicElemSizeSlotForExpr(base, baseType); ok {
		fg.out.Instrs = append(fg.out.Instrs, bytecode.LoadLocal(bytecode.TypeI64, slot), bytecode.Instr{Op: bytecode.OpPtrAddDynamic})
		return nil
	}
	size := fg.g.elemSize(baseType)
	if size <= 0 {
		return fmt.Errorf("cannot lower pointer arithmetic with zero element size for %s", baseType)
	}
	fg.out.Instrs = append(fg.out.Instrs, bytecode.Instr{Op: bytecode.OpPtrAdd, Size: size})
	return nil
}

func (fg *funcGen) dynamicElemSizeSlotForExpr(base sema.Expr, baseType sema.Type) (int, bool) {
	var elem sema.Type
	switch x := sema.Unqual(baseType).(type) {
	case *sema.PointerType:
		elem = x.Pointee
	case *sema.ArrayType:
		elem = x.Elem
	default:
		return 0, false
	}
	if !typeHasVariableSize(elem) {
		return 0, false
	}
	if sym := dynamicObjectSymbol(base); sym != nil {
		if slot, ok := fg.dynamicSizeSlotForSymbol(sym, elem); ok {
			return slot, true
		}
	}
	if sym := pointerSymbol(base); sym != nil {
		if slots := fg.dynamicPointerTypeMap[sym]; slots != nil {
			slot, ok := slots[dynamicSizeKey(elem)]
			return slot, ok
		}
	}
	return 0, false
}

func (fg *funcGen) dynamicSizeSlotForPointerSymbol(sym *sema.Symbol, t sema.Type) (int, bool) {
	if sym == nil {
		return 0, false
	}
	slots := fg.dynamicPointerTypeMap[sym]
	if slots == nil {
		return 0, false
	}
	slot, ok := slots[dynamicSizeKey(t)]
	return slot, ok
}

func isPointerArithmeticExpr(x *sema.BinOp) bool {
	if x == nil || (x.Op != sema.OpAdd && x.Op != sema.OpSub) {
		return false
	}
	_, leftPtr := sema.Unqual(x.L.GetType()).(*sema.PointerType)
	_, rightPtr := sema.Unqual(x.R.GetType()).(*sema.PointerType)
	return leftPtr || rightPtr
}

func functionTypeFromCallee(t sema.Type) (*sema.FunctionType, error) {
	if ft, ok := sema.Unqual(t).(*sema.FunctionType); ok {
		return ft, nil
	}
	if pt, ok := sema.Unqual(t).(*sema.PointerType); ok {
		if ft, ok := sema.Unqual(pt.Pointee).(*sema.FunctionType); ok {
			return ft, nil
		}
	}
	return nil, fmt.Errorf("callee type %s is not a function pointer", t)
}

func directCallGlobal(e sema.Expr) int {
	if vr := functionVarRef(e); vr != nil && vr.Sym != nil {
		return vr.Sym.GlobalID
	}
	return -1
}

func functionVarRef(e sema.Expr) *sema.VarRef {
	switch x := e.(type) {
	case *sema.VarRef:
		if isFunctionDesignator(x) {
			return x
		}
	case *sema.ImplicitCast:
		if x.Kind == sema.LValueToRValue || x.Kind == sema.FunctionDecay {
			return functionVarRef(x.X)
		}
	}
	return nil
}

func dereferencedFunctionPointer(e sema.Expr) sema.Expr {
	switch x := e.(type) {
	case *sema.ImplicitCast:
		if x.Kind == sema.LValueToRValue {
			return dereferencedFunctionPointer(x.X)
		}
	case *sema.UnOp:
		if x.Op == sema.UnDeref {
			return x.X
		}
	}
	return nil
}

func isFunctionDesignator(vr *sema.VarRef) bool {
	if vr == nil {
		return false
	}
	if vr.Sym != nil && vr.Sym.Kind == sema.SymFunc {
		return true
	}
	_, ok := sema.Unqual(vr.T).(*sema.FunctionType)
	return ok
}

func isFunctionExpr(e sema.Expr) bool {
	return functionVarRef(e) != nil
}

func dynamicObjectSymbol(e sema.Expr) *sema.Symbol {
	switch x := e.(type) {
	case *sema.VarRef:
		return x.Sym
	case *sema.ImplicitCast:
		if x.Kind == sema.LValueToRValue || x.Kind == sema.ArrayDecay {
			return dynamicObjectSymbol(x.X)
		}
	case *sema.IndexExpr:
		return dynamicObjectSymbol(x.Base)
	}
	return nil
}

func pointerSymbol(e sema.Expr) *sema.Symbol {
	switch x := e.(type) {
	case *sema.VarRef:
		return x.Sym
	case *sema.ImplicitCast:
		if x.Kind == sema.LValueToRValue || x.Kind == sema.ArrayDecay || x.Kind == sema.PointerConversion || x.Kind == sema.VoidPointerConversion {
			return pointerSymbol(x.X)
		}
	case *sema.UnOp:
		if x.Op == sema.UnDeref {
			return pointerSymbol(x.X)
		}
	case *sema.BinOp:
		if x.Op == sema.OpAdd || x.Op == sema.OpSub {
			if sym := pointerSymbol(x.L); sym != nil {
				return sym
			}
			return pointerSymbol(x.R)
		}
	}
	return nil
}

func sameUnqualType(a, b sema.Type) bool {
	return sema.Unqual(a) == sema.Unqual(b)
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
