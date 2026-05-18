package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"shinya.click/cvm/bytecode"
)

func TestError(t *testing.T) {
	if err := (&Compiler{}).RunSource(`typedef int a;
int main() {
	int a;
	int b;
	a*b;
}`); err != nil {
		t.Fatalf("RunSource returned error: %v", err)
	}
}

func TestCompilerRunsPreprocessor(t *testing.T) {
	if err := (&Compiler{}).RunSource(`#define T int
T main(void) {
	return 0;
}`); err != nil {
		t.Fatalf("RunSource returned error: %v", err)
	}
}

func TestAbstractDeclaration(t *testing.T) {

	if err := (&Compiler{}).RunSource(`void func2(int (*)[10]);

int main()
{
  func2();
  return 0;
}`); err == nil {
		t.Fatal("RunSource returned nil error")
	}
}

func TestEnumDeclaration(t *testing.T) {
	if err := (&Compiler{}).RunSource(`enum a { b, c };
int main() {
	enum a d = b;
}`); err != nil {
		t.Fatalf("RunSource returned error: %v", err)
	}
}

func TestStructDeclaration(t *testing.T) {
	if err := (&Compiler{}).RunSource(`struct a { int b; };
int main() {
	int ccc = b;
	struct a c;
	c.b = 1;
	return 0;
}`); err == nil {
		t.Fatal("RunSource returned nil error")
	}
}

func TestFunctionDeclaration(t *testing.T) {
	if err := (&Compiler{}).RunSource(`int (*f)();
int main() {
	f();
}`); err != nil {
		t.Fatalf("RunSource returned error: %v", err)
	}
}

func TestKAndRFunctionDeclaration(t *testing.T) {
	if err := (&Compiler{}).RunSource(`void example(a, b, c)
    int a;
    char b;
    float c;
{
	int ccc = a;
}`); err == nil {
		t.Fatal("RunSource returned nil error")
	}
}

func TestFunctionParameterShadow(t *testing.T) {
	if err := (&Compiler{}).RunSource(`typedef int a;
void example(int a) {
	a*b;
}`); err == nil {
		t.Fatal("RunSource returned nil error")
	}
}

func TestCompilerRejectsBothDumpModesBeforeCompilation(t *testing.T) {
	err := (&Compiler{DumpIR: true, DumpBytecode: true}).RunSource(`not valid c`)
	if err == nil {
		t.Fatal("RunSource returned nil error")
	}
	if !strings.Contains(err.Error(), "mutually exclusive") {
		t.Fatalf("RunSource error = %v, want mutually exclusive", err)
	}
}

func TestCompilerEmitBytecodeWritesLoadableBinaryModule(t *testing.T) {
	path := filepath.Join(t.TempDir(), "main.cvmbc")
	c := &Compiler{EmitBytecode: path}
	if err := c.RunSource(`int main(void) { return 0; }`); err != nil {
		t.Fatalf("RunSource returned error: %v", err)
	}
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open emitted bytecode: %v", err)
	}
	defer f.Close()
	mod, err := bytecode.DecodeModule(f)
	if err != nil {
		t.Fatalf("DecodeModule: %v", err)
	}
	if mod.Entry == nil || mod.Entry.Global != 0 || mod.Entry.Name != "main" {
		t.Fatalf("entry = %#v, want main at global 0", mod.Entry)
	}
	if len(mod.Functions) != 1 || mod.Functions[0].Name != "main" {
		t.Fatalf("functions = %#v, want one main function", mod.Functions)
	}
}

func TestMainEmitBytecodeFlagWritesLoadableBinaryModule(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "main.c")
	out := filepath.Join(dir, "main.cvmbc")
	if err := os.WriteFile(src, []byte(`int main(void) { return 0; }`), 0644); err != nil {
		t.Fatalf("write source: %v", err)
	}
	oldArgs := os.Args
	os.Args = []string{"cvm", "--emit-bytecode", out, src}
	defer func() { os.Args = oldArgs }()

	if code := runMain(os.Args[1:]); code != 0 {
		t.Fatalf("runMain exit code = %d, want 0", code)
	}

	f, err := os.Open(out)
	if err != nil {
		t.Fatalf("open emitted bytecode: %v", err)
	}
	defer f.Close()
	mod, err := bytecode.DecodeModule(f)
	if err != nil {
		t.Fatalf("DecodeModule: %v", err)
	}
	if mod.Entry == nil || mod.Entry.Name != "main" {
		t.Fatalf("entry = %#v, want main", mod.Entry)
	}
}

func TestMainRunBytecodeUsesExitCode(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "main.c")
	out := filepath.Join(dir, "main.cvmbc")
	if err := os.WriteFile(src, []byte(`int main(void) { return 7; }`), 0644); err != nil {
		t.Fatalf("write source: %v", err)
	}
	if err := (&Compiler{EmitBytecode: out}).RunFile(src); err != nil {
		t.Fatalf("emit bytecode: %v", err)
	}
	if code := runMain([]string{"run", out}); code != 7 {
		t.Fatalf("runMain exit code = %d, want 7", code)
	}
}

func TestRunBytecodeForwardsArgs(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "main.c")
	out := filepath.Join(dir, "main.cvmbc")
	source := `int main(int argc, char **argv) {
	if (argc != 3) return 1;
	if (argv[1][0] != 'x') return 2;
	if (argv[2][0] != 'y') return 3;
	return 9;
}`
	if err := os.WriteFile(src, []byte(source), 0644); err != nil {
		t.Fatalf("write source: %v", err)
	}
	if err := (&Compiler{EmitBytecode: out}).RunFile(src); err != nil {
		t.Fatalf("emit bytecode: %v", err)
	}
	if code := runMain([]string{"run", out, "xray", "yak"}); code != 9 {
		t.Fatalf("runMain exit code = %d, want 9", code)
	}
}
