package sema

import "testing"

func unwrapCasts(e Expr) Expr {
	for {
		ic, ok := e.(*ImplicitCast)
		if !ok {
			return e
		}
		e = ic.X
	}
}

func TestTypeExprBinaryUnaryCallMemberIndex(t *testing.T) {
	r := analyzeSource(t, `struct S { int x; } s; int a[5]; int f(int);
void g() {
	int v = s.x + a[2] + f(3);
	int *p = &v;
	int y = *p;
}`)
	if len(r.Errors) != 0 {
		t.Fatalf("unexpected errors: %v", r.Errors)
	}
	def := r.Program.Funcs[0]
	v := def.Locals[0]
	if _, ok := unwrapCasts(v.Init).(*BinOp); !ok {
		t.Fatalf("expected binop init, got %T", v.Init)
	}
	p := def.Locals[1]
	if _, ok := unwrapCasts(p.Init).(*UnOp); !ok {
		t.Fatalf("expected address init, got %T", p.Init)
	}
	y := def.Locals[2]
	if _, ok := unwrapCasts(y.Init).(*UnOp); !ok {
		t.Fatalf("expected deref init, got %T", y.Init)
	}
}

func TestTypeExprAssignmentConditionalAndCast(t *testing.T) {
	r := analyzeSource(t, "int x; void g() { int y = (x = 5); int z = x ? (int)3.5 : 2; }")
	if len(r.Errors) != 0 {
		t.Fatalf("unexpected errors: %v", r.Errors)
	}
	def := r.Program.Funcs[0]
	if _, ok := unwrapCasts(def.Locals[0].Init).(*AssignExpr); !ok {
		t.Fatalf("expected assignment expression")
	}
	if _, ok := unwrapCasts(def.Locals[1].Init).(*CondExpr); !ok {
		t.Fatalf("expected conditional expression")
	}
}
