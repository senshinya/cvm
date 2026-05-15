package main

import (
	"strings"
	"testing"
)

func TestCompilerDumpIRWritesSemaProgram(t *testing.T) {
	var out strings.Builder
	c := &Compiler{DumpIR: true, Output: &out}
	c.RunSource(`int main(void) { return 0; }`)
	got := out.String()
	for _, want := range []string{"Program", "FuncDef name=\"main\"", "ReturnStmt"} {
		if !strings.Contains(got, want) {
			t.Fatalf("dump IR missing %q:\n%s", want, got)
		}
	}
}
