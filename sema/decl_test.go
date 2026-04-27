package sema

import "testing"

func TestApplyDeclaratorShapesViaAnalyze(t *testing.T) {
	r := analyzeSource(t, "int *p; int a[5]; int f(int x, double y);")
	if len(r.Errors) != 0 {
		t.Fatalf("unexpected errors: %v", r.Errors)
	}
	if _, ok := r.Program.Globals[0].(*VarDecl).T.(*PointerType); !ok {
		t.Fatalf("p should be pointer, got %T", r.Program.Globals[0].(*VarDecl).T)
	}
	arr, ok := r.Program.Globals[1].(*VarDecl).T.(*ArrayType)
	if !ok || arr.SizeKind != ArrayConstantSize || arr.Size != 5 {
		t.Fatalf("a should be int[5], got %v", r.Program.Globals[1].(*VarDecl).T)
	}
	fn := r.Program.Globals[2].(*FuncDecl).T
	if len(fn.Params) != 2 || !fn.HasProto {
		t.Fatalf("f should have 2-parameter prototype, got %+v", r.Program.Globals[2])
	}
}
