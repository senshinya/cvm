package sema

import "testing"

func TestIntegrationHelloWorldAndFactorial(t *testing.T) {
	hello := `int printf(const char *fmt, ...);
int main() {
	printf("hello\n");
	return 0;
}`
	r := analyzeSource(t, hello)
	if len(r.Errors) != 0 {
		t.Fatalf("hello errors: %v", r.Errors)
	}
	if len(r.Program.Funcs) != 1 || r.Program.Funcs[0].Sym.Name != "main" {
		t.Fatalf("expected main, got %+v", r.Program.Funcs)
	}
	fact := `int factorial(int n) {
	if (n <= 1) return 1;
	return n * factorial(n - 1);
}`
	r = analyzeSource(t, fact)
	if len(r.Errors) != 0 {
		t.Fatalf("factorial errors: %v", r.Errors)
	}
}
