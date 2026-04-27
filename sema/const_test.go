package sema

import "testing"

func TestEvalIntegerConstantLiteral(t *testing.T) {
	tt := NewTypeTable()
	ev := NewEvaluator(nil)
	cv, ok := ev.EvalIntegerConstant(&IntLit{Value: 42, T: tt.Builtin(Int)})
	if !ok {
		t.Fatalf("EvalIntegerConstant on IntLit failed")
	}
	if cv.Kind != ConstInt || cv.Int != 42 {
		t.Fatalf("wrong const value: %+v", cv)
	}
}

func TestEvalArithmeticAndConditional(t *testing.T) {
	tt := NewTypeTable()
	intT := tt.Builtin(Int)
	ev := NewEvaluator(nil)
	expr := &BinOp{Op: OpAdd, L: &IntLit{Value: 3, T: intT}, R: &BinOp{Op: OpMul, L: &IntLit{Value: 4, T: intT}, R: &IntLit{Value: 2, T: intT}, T: intT}, T: intT}
	cv, ok := ev.EvalIntegerConstant(expr)
	if !ok || cv.Int != 11 {
		t.Fatalf("arithmetic eval failed: %+v ok=%v", cv, ok)
	}
	cond := &CondExpr{Cond: &IntLit{Value: 1, T: intT}, Then: &IntLit{Value: 10, T: intT}, Else: &IntLit{Value: 20, T: intT}, T: intT}
	cv, ok = ev.EvalIntegerConstant(cond)
	if !ok || cv.Int != 10 {
		t.Fatalf("cond eval failed: %+v ok=%v", cv, ok)
	}
}

func TestEvalSizeofAndRejectsNonConstant(t *testing.T) {
	tt := NewTypeTable()
	intT := tt.Builtin(Int)
	ev := NewEvaluator(nil)
	cv, ok := ev.EvalIntegerConstant(&SizeofExpr{Operand: SizeofOperand{Type: intT}, T: tt.Builtin(ULong)})
	if !ok || cv.Int != 4 {
		t.Fatalf("sizeof(int) = %+v ok=%v", cv, ok)
	}
	if _, ok := ev.EvalIntegerConstant(&VarRef{Sym: &Symbol{Name: "x"}, T: intT}); ok {
		t.Fatalf("VarRef must not be integer constant")
	}
}

func TestEvalAddressConstant(t *testing.T) {
	tt := NewTypeTable()
	intT := tt.Builtin(Int)
	sym := &Symbol{Name: "g", Storage: StorageStatic, T: tt.ArrayConstant(intT, 10)}
	addr := &UnOp{Op: UnAddr, X: &VarRef{Sym: sym, T: sym.T}, T: tt.Pointer(intT)}
	expr := &BinOp{Op: OpAdd, L: addr, R: &IntLit{Value: 3, T: intT}, T: tt.Pointer(intT)}
	cv, ok := NewEvaluator(nil).EvalConstant(expr)
	if !ok || cv.Kind != ConstAddress || cv.Addr.Sym != sym || cv.Addr.Offset != 12 {
		t.Fatalf("wrong address const: %+v ok=%v", cv, ok)
	}
}
