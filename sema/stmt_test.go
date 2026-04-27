package sema

import "testing"

func TestStmtWalkerControlFlow(t *testing.T) {
	src := `int x; int f(int n) {
	if (n <= 1) return 1;
	for (int i = 0; i < n; i++) { x = x + i; }
	while (x) break;
	goto L;
L:
	return n * f(n - 1);
}`
	r := analyzeSource(t, src)
	if len(r.Errors) != 0 {
		t.Fatalf("unexpected errors: %v", r.Errors)
	}
	def := r.Program.Funcs[0]
	if def.Body == nil || len(def.Body.Items) == 0 {
		t.Fatalf("body missing: %+v", def.Body)
	}
	if def.Labels["L"] == nil {
		t.Fatalf("label L missing: %+v", def.Labels)
	}
}

func TestStmtWalkerSwitch(t *testing.T) {
	src := `int x; void f() {
	switch (x) {
	case 1: x = 10; break;
	case 2: x = 20; break;
	default: x = 0;
	}
}`
	r := analyzeSource(t, src)
	if len(r.Errors) != 0 {
		t.Fatalf("unexpected errors: %v", r.Errors)
	}
	sw, ok := r.Program.Funcs[0].Body.Items[0].(*SwitchStmt)
	if !ok || len(sw.Cases) != 2 || sw.Default == nil {
		t.Fatalf("switch wrong: %T %+v", r.Program.Funcs[0].Body.Items[0], r.Program.Funcs[0].Body.Items[0])
	}
}
