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

	pp, err := preprocessor.PreprocessSource("main.c", src, preprocessor.Options{})
	if err != nil {
		t.Fatalf("preprocess: %v", err)
	}
	candidates, err := parser.NewParser(pp.Tokens).Parse()
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	prog, err := sema.Analyze(candidates)
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
