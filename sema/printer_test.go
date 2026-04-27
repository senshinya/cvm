package sema

import (
	"strings"
	"testing"
)

func TestPrintIR_Simple(t *testing.T) {
	prog, err := Analyze(parseCandidates(t, "int x = 42;"))
	if err != nil {
		t.Fatal(err)
	}
	out := PrintProgram(prog)
	if !strings.Contains(out, "VarDecl") || !strings.Contains(out, "x") || !strings.Contains(out, "42") {
		t.Fatalf("printer output missing expected pieces:\n%s", out)
	}
}
