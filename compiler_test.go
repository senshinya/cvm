package main

import (
	"strings"
	"testing"
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
