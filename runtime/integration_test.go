package runtime

import (
	"bytes"
	"context"
	"testing"

	"shinya.click/cvm/bytecode"
	"shinya.click/cvm/codegen"
	"shinya.click/cvm/parser"
	"shinya.click/cvm/preprocessor"
	"shinya.click/cvm/sema"
)

func compileAndRun(t *testing.T, src string, stdout *bytes.Buffer) (ExitStatus, error) {
	t.Helper()
	return compileAndRunWithOptions(t, src, stdout, sema.SemaOptions{})
}

func compileAndRunWithOptions(t *testing.T, src string, stdout *bytes.Buffer, opts sema.SemaOptions) (ExitStatus, error) {
	t.Helper()

	pp, err := preprocessor.PreprocessSource("main.c", src, preprocessor.Options{})
	if err != nil {
		t.Fatalf("preprocess: %v", err)
	}
	candidates, err := parser.NewParser(pp.Tokens).Parse()
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	prog, err := sema.AnalyzeWithOptions(candidates, opts)
	if err != nil {
		t.Fatalf("sema: %v", err)
	}
	mod, err := codegen.Generate(prog)
	if err != nil {
		t.Fatalf("codegen: %v", err)
	}

	var encoded bytes.Buffer
	if err := bytecode.EncodeModule(&encoded, mod); err != nil {
		t.Fatalf("EncodeModule: %v", err)
	}

	p, err := Load(bytes.NewReader(encoded.Bytes()), LoadOptions{
		Externs: DefaultExternRegistry(stdout, nil),
	})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	return Run(context.Background(), p, RunOptions{})
}

func TestCompileAndRunReturnArithmetic(t *testing.T) {
	st, err := compileAndRun(t, `int main(void) { int x = 3; int y = 4; return x * y + 2; }`, nil)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if st.Code != 14 {
		t.Fatalf("exit code = %d, want 14", st.Code)
	}
}

func TestCompileAndRunGlobalAndLoop(t *testing.T) {
	st, err := compileAndRun(t, `int g = 2; int main(void) { int i = 0; while (i < 3) { g = g + 1; i = i + 1; } return g; }`, nil)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if st.Code != 5 {
		t.Fatalf("exit code = %d, want 5", st.Code)
	}
}

func TestCompileAndRunPuts(t *testing.T) {
	var out bytes.Buffer
	st, err := compileAndRun(t, `int puts(const char *); int main(void) { puts("hi"); return 0; }`, &out)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
	if out.String() != "hi\n" {
		t.Fatalf("stdout = %q, want %q", out.String(), "hi\n")
	}
}

func TestCompileAndRunLocalArrayAddressing(t *testing.T) {
	st, err := compileAndRun(t, `int main(void) { int a[2]; a[0] = 4; a[1] = 7; return a[0] + a[1]; }`, nil)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if st.Code != 11 {
		t.Fatalf("exit code = %d, want 11", st.Code)
	}
}

func TestCompileAndRunCommaExpressionSequencesAndReturnsRightValue(t *testing.T) {
	st, err := compileAndRun(t, `int main(void) { int x = 1; return (x = x + 2, x * 3); }`, nil)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if st.Code != 9 {
		t.Fatalf("exit code = %d, want 9", st.Code)
	}
}

func TestCompileAndRunPostDecrementReturnsOldValueAndUpdatesObject(t *testing.T) {
	st, err := compileAndRun(t, `int main(void) { int x = 2; int old = x--; return old * 10 + x; }`, nil)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if st.Code != 21 {
		t.Fatalf("exit code = %d, want 21", st.Code)
	}
}

func TestCompileAndRunBitwiseNotUsesPromotedIntegerWidth(t *testing.T) {
	st, err := compileAndRun(t, `int main(void) { _Bool u = 0; if (~u != -1) return 1; u = 1; if (~u != -2) return 2; return 7; }`, nil)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if st.Code != 7 {
		t.Fatalf("exit code = %d, want 7", st.Code)
	}
}

func TestCompileAndRunConditionalExpressionSelectsOnlyChosenBranch(t *testing.T) {
	st, err := compileAndRun(t, `int main(void) { int x = 0; int y = 1 ? (x = 4) : (x = 9); return y * 10 + x; }`, nil)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if st.Code != 44 {
		t.Fatalf("exit code = %d, want 44", st.Code)
	}
}

func TestCompileAndRunCompoundAssignUsesArithmeticThenAssignmentConversion(t *testing.T) {
	st, err := compileAndRun(t, `int main(void) { _Bool u = 1; if ((u /= 2) != 0) return 1; int j = 1; j += 4; return j; }`, nil)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if st.Code != 5 {
		t.Fatalf("exit code = %d, want 5", st.Code)
	}
}

func TestCompileAndRunBoolArraySubscriptPromotesIndexForPointerAdd(t *testing.T) {
	st, err := compileAndRun(t, `int main(void) { const char *t = "_B"; _Bool u = 1; return t[u] == 'B' && u[t] == 'B' ? 7 : 1; }`, nil)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if st.Code != 7 {
		t.Fatalf("exit code = %d, want 7", st.Code)
	}
}

func TestCompileAndRunBoolDecrementStoresBoolConvertedValue(t *testing.T) {
	st, err := compileAndRun(t, `int main(void) { _Bool u = 0; if (u-- != 0) return 1; if (u != 1) return 2; if (--u != 0) return 3; return 7; }`, nil)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if st.Code != 7 {
		t.Fatalf("exit code = %d, want 7", st.Code)
	}
}

func TestCompileAndRunBoolBitFieldStoresAndLoadsConvertedValue(t *testing.T) {
	st, err := compileAndRun(t, `struct S { _Bool b : 1; }; int main(void) { struct S s; s.b = 2; if (s.b != 1) return 1; s.b = 0; if (s.b != 0) return 2; s.b = 0.2; if (s.b != 1) return 3; s.b = &s; if (s.b != 1) return 4; return 7; }`, nil)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if st.Code != 7 {
		t.Fatalf("exit code = %d, want 7", st.Code)
	}
}

func TestCompileAndRunStaticCompoundLiteralPointerInitializers(t *testing.T) {
	st, err := compileAndRun(t, `int *p = &(int){3}; int *q = (int[]){4, 5}; int main(void) { if (*p != 3 || q[1] != 5) return 1; *p = 7; q[0] = 8; return *p + q[0]; }`, nil)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if st.Code != 15 {
		t.Fatalf("exit code = %d, want 15", st.Code)
	}
}

func TestCompileAndRunIncDecAddressableMemberAndIndex(t *testing.T) {
	st, err := compileAndRun(t, `struct S { int a; int b; }; int main(void) { struct S s = {1, 4}; int a[2] = {3, 9}; if (s.a++ != 1) return 1; if (--s.b != 3) return 2; if (a[0]++ != 3) return 3; return s.a + s.b + a[0]; }`, nil)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if st.Code != 9 {
		t.Fatalf("exit code = %d, want 9", st.Code)
	}
}

func TestCompileAndRunCompoundAssignAddressableIndex(t *testing.T) {
	st, err := compileAndRun(t, `int main(void) { int a[4] = {0, 2, 5, 1}; if ((a[1] *= a[2]) != 10) return 1; if ((a[2] -= a[3]) != 4) return 2; if ((a[3] += 7) != 8) return 3; return a[1] + a[2] + a[3]; }`, nil)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if st.Code != 22 {
		t.Fatalf("exit code = %d, want 22", st.Code)
	}
}

func TestCompileAndRunVoidCastDiscardsValue(t *testing.T) {
	st, err := compileAndRun(t, `int main(void) { if (1) { (void)0; } return 7; }`, nil)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if st.Code != 7 {
		t.Fatalf("exit code = %d, want 7", st.Code)
	}
}

func TestCompileAndRunFuncIdentifierCString(t *testing.T) {
	st, err := compileAndRun(t, `
extern int strcmp(const char *, const char *);
int main(void) {
	return strcmp(__func__, "main") || sizeof(__func__) != 5;
}`, nil)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestCompileAndRunFuncIdentifierHasDistinctAddress(t *testing.T) {
	st, err := compileAndRun(t, `int main(void) { return "main" == __func__; }`, nil)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestCompileAndRunCapturingNestedFunctionDirectCall(t *testing.T) {
	st, err := compileAndRunWithOptions(t, `
int main(void) {
	int x = 4;
	int inner(void) { x += 2; x++; return x; }
	return inner();
}`, nil, sema.SemaOptions{GNUExtensions: true})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if st.Code != 7 {
		t.Fatalf("exit code = %d, want 7", st.Code)
	}
}

func TestCompileAndRunCapturingNestedFunctionVLA(t *testing.T) {
	st, err := compileAndRunWithOptions(t, `
int main(void) {
	int n = 3;
	int a[n];
	a[0] = 5;
	int inner(void) { a[0] += 2; return a[0]; }
	return inner();
}`, nil, sema.SemaOptions{GNUExtensions: true})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if st.Code != 7 {
		t.Fatalf("exit code = %d, want 7", st.Code)
	}
}

func TestCompileAndRunNestedFunctionForwardsOuterCapture(t *testing.T) {
	st, err := compileAndRunWithOptions(t, `
int main(void) {
	int x = 2;
	int middle(void) {
		int inner(void) { return x + 5; }
		return inner();
	}
	return middle();
}`, nil, sema.SemaOptions{GNUExtensions: true})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if st.Code != 7 {
		t.Fatalf("exit code = %d, want 7", st.Code)
	}
}
