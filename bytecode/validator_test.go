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

func TestValidateModuleRejectsReturnTypeMismatchWithSignature(t *testing.T) {
	mod := minimalModule()
	mod.Functions[0].Instrs = []Instr{
		I32Const(0),
		Return(TypeI32),
	}
	mod.Sigs[0].Ret = TypeI64

	if err := ValidateModule(mod); err == nil {
		t.Fatal("ValidateModule accepted a return instruction that mismatches the function signature")
	}
}

func TestValidateModuleRejectsReturnVoidForNonVoidSignature(t *testing.T) {
	mod := minimalModule()
	mod.Functions[0].Instrs = []Instr{Instr{Op: OpReturnVoid}}

	if err := ValidateModule(mod); err == nil {
		t.Fatal("ValidateModule accepted ReturnVoid for a non-void function signature")
	}
}

func TestValidateModuleRejectsReturnValueForVoidSignature(t *testing.T) {
	mod := minimalModule()
	mod.Sigs[0].Ret = TypeVoid
	mod.Functions[0].Instrs = []Instr{
		I32Const(0),
		Return(TypeI32),
	}

	if err := ValidateModule(mod); err == nil {
		t.Fatal("ValidateModule accepted Return(value) for a void function signature")
	}
}

func TestValidateModuleRejectsFunctionIDNotMatchingIndex(t *testing.T) {
	mod := minimalModule()
	mod.Functions = append(mod.Functions, Function{
		ID:       0,
		GlobalID: 1,
		Name:     "other",
		Sig:      0,
		Instrs: []Instr{
			I32Const(0),
			Return(TypeI32),
		},
	})
	mod.Globals = append(mod.Globals, Global{ID: 1, Name: "other", Kind: GlobalFunc, Func: 1})

	if err := ValidateModule(mod); err == nil {
		t.Fatal("ValidateModule accepted a function id that does not match its slice index")
	}
}

func TestValidateModuleRejectsBadFunctionGlobalBackLink(t *testing.T) {
	t.Run("global is not function", func(t *testing.T) {
		mod := minimalModule()
		mod.Globals[0].Kind = GlobalVar

		if err := ValidateModule(mod); err == nil {
			t.Fatal("ValidateModule accepted a function whose global is not GlobalFunc")
		}
	})

	t.Run("global points elsewhere", func(t *testing.T) {
		mod := minimalModule()
		mod.Functions = append(mod.Functions, Function{
			ID:       1,
			GlobalID: 0,
			Name:     "other",
			Sig:      0,
			Instrs: []Instr{
				I32Const(0),
				Return(TypeI32),
			},
		})
		mod.Globals[0].Func = 1

		if err := ValidateModule(mod); err == nil {
			t.Fatal("ValidateModule accepted a function whose global does not point back to it")
		}
	})
}

func TestValidateModuleRejectsPopUnderflow(t *testing.T) {
	mod := minimalModule()
	mod.Functions[0].Instrs = []Instr{{Op: OpPop}}

	if err := ValidateModule(mod); err == nil {
		t.Fatal("ValidateModule accepted pop with an empty stack")
	}
}

func TestValidateModuleRejectsSwapUnderflow(t *testing.T) {
	mod := minimalModule()
	mod.Functions[0].Instrs = []Instr{
		I32Const(0),
		{Op: OpSwap},
	}

	if err := ValidateModule(mod); err == nil {
		t.Fatal("ValidateModule accepted swap with fewer than two stack values")
	}
}

func TestValidateModuleAcceptsPopAndSwapWithValidStack(t *testing.T) {
	mod := minimalModule()
	mod.Functions[0].Instrs = []Instr{
		I32Const(1),
		I64Const(2),
		{Op: OpSwap},
		{Op: OpPop},
		Return(TypeI64),
	}
	mod.Sigs[0].Ret = TypeI64

	if err := ValidateModule(mod); err != nil {
		t.Fatalf("ValidateModule rejected valid pop/swap stack use: %v", err)
	}
}

func minimalModule() *Module {
	return &Module{
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
}
