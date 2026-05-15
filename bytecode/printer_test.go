package bytecode

import (
	"strings"
	"testing"
)

func TestPrintModuleStableMinimalFunction(t *testing.T) {
	mod := &Module{
		Target:  DefaultTarget(),
		Globals: []Global{{ID: 0, Name: "main", Kind: GlobalFunc, Func: 0}},
		Sigs:    []FuncSig{{ID: 0, Ret: TypeI32}},
		Functions: []Function{{
			ID:       0,
			GlobalID: 0,
			Name:     "main",
			Sig:      0,
			Instrs: []Instr{
				I32Const(0),
				Return(TypeI32),
			},
		}},
	}
	got := PrintModule(mod)
	for _, want := range []string{
		"Module target=\"cvm-default\"",
		"Global #0 func name=\"main\" func=0",
		"Sig #0 ret=i32 params=()",
		"Func #0 global=0 name=\"main\" sig=0",
		"0000: I32Const 0",
		"0001: I32Return",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("PrintModule missing %q:\n%s", want, got)
		}
	}
}
