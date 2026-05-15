package bytecode

import "testing"

func TestValidateModuleAcceptsMinimalReturningFunction(t *testing.T) {
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
	if err := ValidateModule(mod); err != nil {
		t.Fatalf("ValidateModule rejected minimal module: %v", err)
	}
}

func TestValidateModuleRejectsBadGlobalFunctionReference(t *testing.T) {
	mod := &Module{
		Target:  DefaultTarget(),
		Globals: []Global{{ID: 0, Name: "main", Kind: GlobalFunc, Func: 99}},
		Sigs:    []FuncSig{{ID: 0, Ret: TypeI32}},
	}
	if err := ValidateModule(mod); err == nil {
		t.Fatal("ValidateModule accepted a function global with an invalid function index")
	}
}
