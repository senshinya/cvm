package main

import (
	"strings"
	"testing"
)

func TestCompilerDumpIRWritesSemaProgram(t *testing.T) {
	var out strings.Builder
	c := &Compiler{DumpIR: true, Output: &out}
	if err := c.RunSource(`int main(void) { return 0; }`); err != nil {
		t.Fatalf("RunSource returned error: %v", err)
	}
	got := out.String()
	for _, want := range []string{"Program", "FuncDef name=\"main\"", "ReturnStmt"} {
		if !strings.Contains(got, want) {
			t.Fatalf("dump IR missing %q:\n%s", want, got)
		}
	}
}

func TestCompilerDumpBytecodeWritesModule(t *testing.T) {
	var out strings.Builder
	c := &Compiler{DumpBytecode: true, Output: &out}
	if err := c.RunSource(`int main(void) { return 0; }`); err != nil {
		t.Fatalf("RunSource returned error: %v", err)
	}
	got := out.String()
	for _, want := range []string{"Module target=", "Func #0", "I32Return"} {
		if !strings.Contains(got, want) {
			t.Fatalf("dump bytecode missing %q:\n%s", want, got)
		}
	}
}

func TestCompilerRejectsBothDumpModes(t *testing.T) {
	err := (&Compiler{DumpIR: true, DumpBytecode: true}).RunSource(`int main(void) { return 0; }`)
	if err == nil {
		t.Fatal("RunSource returned nil error")
	}
	if !strings.Contains(err.Error(), "mutually exclusive") {
		t.Fatalf("RunSource error = %v, want mutually exclusive", err)
	}
}
