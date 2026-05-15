package bytecode

import (
	"strings"
	"testing"
)

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

func TestValidateModuleRejectsReturnTypeMismatch(t *testing.T) {
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
				F64Const(1),
				Return(TypeI32),
			},
		}},
	}
	if err := ValidateModule(mod); err == nil {
		t.Fatal("ValidateModule accepted return type mismatch")
	}
}

func TestValidateModuleRejectsMissingJumpLabel(t *testing.T) {
	mod := &Module{
		Target:  DefaultTarget(),
		Globals: []Global{{ID: 0, Name: "main", Kind: GlobalFunc, Func: 0}},
		Sigs:    []FuncSig{{ID: 0, Ret: TypeVoid}},
		Functions: []Function{{
			ID:       0,
			GlobalID: 0,
			Name:     "main",
			Sig:      0,
			Instrs: []Instr{
				Jump(7),
				{Op: OpReturnVoid},
			},
		}},
	}
	if err := ValidateModule(mod); err == nil {
		t.Fatal("ValidateModule accepted jump to missing label")
	}
}

func TestValidateModuleRejectsBadCallSignature(t *testing.T) {
	mod := &Module{
		Target:  DefaultTarget(),
		Globals: []Global{{ID: 0, Name: "main", Kind: GlobalFunc, Func: 0}},
		Sigs:    []FuncSig{{ID: 0, Ret: TypeI32, Params: []ValueType{TypeI32}}},
		Functions: []Function{{
			ID:       0,
			GlobalID: 0,
			Name:     "main",
			Sig:      0,
			Instrs: []Instr{
				Call(0, 9, 1),
				Return(TypeI32),
			},
		}},
	}
	if err := ValidateModule(mod); err == nil {
		t.Fatal("ValidateModule accepted call with invalid signature")
	}
}

func TestValidateModuleAcceptsValidReferencesAndStackEffects(t *testing.T) {
	mod := &Module{
		Target: DefaultTarget(),
		Globals: []Global{
			{ID: 0, Name: "main", Kind: GlobalFunc, Func: 0},
			{ID: 1, Name: "callee", Kind: GlobalFunc, Func: 1},
			{ID: 2, Name: "storage", Kind: GlobalVar, Size: 8, Align: 8},
		},
		Strings: []StringConst{{ID: 0, Value: "ok", Bytes: []byte("ok\x00")}},
		Layouts: []ObjectLayout{{
			ID:    0,
			Name:  "pair",
			Size:  8,
			Align: 4,
			Fields: []FieldLayout{{
				ID:     0,
				Name:   "first",
				Offset: 0,
				Type:   TypeI32,
			}},
			Bit: []BitFieldLayout{{
				ID:         0,
				Name:       "flag",
				Container:  TypeU32,
				ByteOffset: 0,
				BitOffset:  0,
				Width:      1,
			}},
		}},
		Sigs: []FuncSig{
			{ID: 0, Ret: TypeI32},
			{ID: 1, Ret: TypeI32, Params: []ValueType{TypeI32}},
		},
		Functions: []Function{
			{
				ID:       0,
				GlobalID: 0,
				Name:     "main",
				Sig:      0,
				Locals:   []LocalSlot{{ID: 0, Name: "tmp", Type: TypeI32}},
				Objects:  []LocalObject{{ID: 0, Name: "obj", Size: 8, Align: 4, Layout: 0}},
				Labels:   []Label{{ID: 0, Name: "done", Statement: true}},
				Instrs: []Instr{
					AddrLocalObject(0),
					Load(TypeI32, 4, false),
					StoreLocal(TypeI32, 0),
					AddrGlobal(2),
					AddrString(0),
					{Op: OpMemCopy, Size: 4, Align: 4},
					AddrLocalObject(0),
					{Op: OpOffset, Type: TypeObjectAddr, Int: 4},
					{Op: OpPop},
					AddrLocalObject(0),
					I32Const(1),
					{Op: OpPtrAdd, Size: 4},
					{Op: OpPop},
					AddrLocalObject(0),
					{Op: OpFieldAddr, Layout: 0, Field: 0},
					{Op: OpPop},
					AddrLocalObject(0),
					{Op: OpBitFieldLoad, Type: TypeI32, Layout: 0, Field: 0},
					{Op: OpPop},
					AddrLocalObject(0),
					I32Const(1),
					{Op: OpBitFieldStore, Type: TypeI32, Layout: 0, Field: 0},
					I32Const(0),
					JumpIfZero(TypeI32, 0),
					I32Const(1),
					Call(1, 1, 1),
					Return(TypeI32),
					LabelInstr(0),
					I32Const(0),
					Return(TypeI32),
				},
			},
			{
				ID:       1,
				GlobalID: 1,
				Name:     "callee",
				Sig:      1,
				Params:   []Param{{Name: "v", Type: TypeI32, Slot: 0}},
				Instrs: []Instr{
					LoadLocal(TypeI32, 0),
					Return(TypeI32),
				},
			},
		},
	}

	if err := ValidateModule(mod); err != nil {
		t.Fatalf("ValidateModule rejected valid references and stack effects: %v", err)
	}
}

func TestValidateModuleAcceptsAddrFuncUsingGlobalField(t *testing.T) {
	mod := moduleWithCallee()
	mod.Functions[0].Instrs = []Instr{
		AddrFunc(1),
		{Op: OpPop},
		I32Const(0),
		Return(TypeI32),
	}

	if err := ValidateModule(mod); err != nil {
		t.Fatalf("ValidateModule rejected AddrFunc with valid global field: %v", err)
	}
}

func TestValidateModuleRejectsInvalidAddrFuncGlobal(t *testing.T) {
	mod := moduleWithCallee()
	mod.Functions[0].Instrs = []Instr{
		AddrFunc(99),
		{Op: OpPop},
		I32Const(0),
		Return(TypeI32),
	}

	if err := ValidateModule(mod); err == nil {
		t.Fatal("ValidateModule accepted AddrFunc with invalid global field")
	}
}

func TestValidateModuleRejectsAddrFuncExternVariable(t *testing.T) {
	mod := minimalModule()
	mod.Globals = append(mod.Globals, Global{ID: 1, Name: "extern_var", Kind: GlobalExtern, Size: 4, Align: 4})
	mod.Functions[0].Instrs = []Instr{
		AddrFunc(1),
		{Op: OpPop},
		I32Const(0),
		Return(TypeI32),
	}

	if err := ValidateModule(mod); err == nil {
		t.Fatal("ValidateModule accepted AddrFunc for extern variable")
	}
}

func TestValidateModuleAcceptsDynamicPtrAdd(t *testing.T) {
	mod := minimalModule()
	mod.Functions[0].Instrs = []Instr{
		{Op: OpConst, Type: TypePtr, Int: 0},
		I32Const(2),
		I64Const(16),
		{Op: OpPtrAddDynamic},
		{Op: OpPop},
		I32Const(0),
		Return(TypeI32),
	}

	if err := ValidateModule(mod); err != nil {
		t.Fatalf("ValidateModule rejected dynamic ptr add: %v", err)
	}
}

func TestValidateModuleAcceptsVariadicCallWithExtraArgs(t *testing.T) {
	mod := moduleWithCallee()
	mod.Sigs[1] = FuncSig{ID: 1, Ret: TypeI32, Params: []ValueType{TypeI32}, Variadic: true}
	mod.Functions[0].Instrs = []Instr{
		I32Const(1),
		I64Const(2),
		Call(1, 1, 2),
		Return(TypeI32),
	}

	if err := ValidateModule(mod); err != nil {
		t.Fatalf("ValidateModule rejected valid variadic call with extra args: %v", err)
	}
}

func TestValidateModuleRejectsNonVariadicCallWithWrongArgc(t *testing.T) {
	mod := moduleWithCallee()
	mod.Sigs[1] = FuncSig{ID: 1, Ret: TypeI32, Params: []ValueType{TypeI32}}
	mod.Functions[0].Instrs = []Instr{
		I32Const(1),
		I64Const(2),
		Call(1, 1, 2),
		Return(TypeI32),
	}

	if err := ValidateModule(mod); err == nil {
		t.Fatal("ValidateModule accepted non-variadic call with wrong argc")
	}
}

func TestValidateModuleRejectsScalarReturnObject(t *testing.T) {
	mod := minimalModule()
	mod.Layouts = []ObjectLayout{{ID: 0, Name: "object", Size: 4, Align: 4}}
	mod.Functions[0].Objects = []LocalObject{{ID: 0, Name: "obj", Size: 4, Align: 4, Layout: 0}}
	mod.Functions[0].Instrs = []Instr{
		AddrLocalObject(0),
		{Op: OpReturnObject},
	}

	if err := ValidateModule(mod); err == nil {
		t.Fatal("ValidateModule accepted ReturnObject in scalar-returning function")
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

func TestValidateModuleRejectsStrayGlobalFuncBackLink(t *testing.T) {
	mod := minimalModule()
	mod.Globals = append(mod.Globals, Global{ID: 1, Name: "main_alias", Kind: GlobalFunc, Func: 0})

	if err := ValidateModule(mod); err == nil {
		t.Fatal("ValidateModule accepted a stray function global whose function points elsewhere")
	}
}

func TestValidateModuleRejectsReturnWithLeftoverStack(t *testing.T) {
	mod := minimalModule()
	mod.Functions[0].Instrs = []Instr{
		I64Const(1),
		I32Const(0),
		Return(TypeI32),
	}

	if err := ValidateModule(mod); err == nil {
		t.Fatal("ValidateModule accepted a non-void return with leftover stack values")
	}
}

func TestValidateModuleRejectsUnhandledOpcode(t *testing.T) {
	t.Run("known unsupported opcode", func(t *testing.T) {
		mod := minimalModule()
		mod.Functions[0].Instrs = []Instr{
			Call(0, 0, 0),
		}

		if err := ValidateModule(mod); err == nil {
			t.Fatal("ValidateModule accepted an unhandled call opcode")
		}
	})

	t.Run("invalid opcode", func(t *testing.T) {
		mod := minimalModule()
		mod.Functions[0].Instrs = []Instr{
			{Op: Opcode(999)},
		}

		if err := ValidateModule(mod); err == nil {
			t.Fatal("ValidateModule accepted an invalid opcode")
		}
	})
}

func TestValidateModuleStackErrorIncludesReadableOpcodeName(t *testing.T) {
	mod := minimalModule()
	mod.Functions[0].Instrs = []Instr{{Op: OpPop}}

	err := ValidateModule(mod)
	if err == nil {
		t.Fatal("ValidateModule accepted pop with an empty stack")
	}
	if !strings.Contains(err.Error(), "OpPop") {
		t.Fatalf("ValidateModule error %q does not include readable opcode name", err)
	}
}

func TestValidateModuleRejectsNonVoidFunctionWithNoInstructions(t *testing.T) {
	mod := minimalModule()
	mod.Functions[0].Instrs = nil

	if err := ValidateModule(mod); err == nil {
		t.Fatal("ValidateModule accepted a non-void function with no instructions")
	}
}

func TestValidateModuleRejectsNonVoidFunctionEndingWithoutReturn(t *testing.T) {
	mod := minimalModule()
	mod.Functions[0].Instrs = []Instr{I32Const(1)}

	if err := ValidateModule(mod); err == nil {
		t.Fatal("ValidateModule accepted a non-void function ending without a return")
	}
}

func TestValidateModuleRejectsVoidFunctionEndingWithLeftoverStack(t *testing.T) {
	mod := minimalModule()
	mod.Sigs[0].Ret = TypeVoid
	mod.Functions[0].Instrs = []Instr{I32Const(1)}

	if err := ValidateModule(mod); err == nil {
		t.Fatal("ValidateModule accepted a void function ending with leftover stack values")
	}
}

func TestValidateModuleDupSwapErrorsIncludeReadableOpcodeName(t *testing.T) {
	t.Run("dup", func(t *testing.T) {
		mod := minimalModule()
		mod.Functions[0].Instrs = []Instr{{Op: OpDup}}

		err := ValidateModule(mod)
		if err == nil {
			t.Fatal("ValidateModule accepted dup with an empty stack")
		}
		if !strings.Contains(err.Error(), "OpDup") {
			t.Fatalf("ValidateModule error %q does not include readable opcode name", err)
		}
	})

	t.Run("swap", func(t *testing.T) {
		mod := minimalModule()
		mod.Functions[0].Instrs = []Instr{I32Const(1), {Op: OpSwap}}

		err := ValidateModule(mod)
		if err == nil {
			t.Fatal("ValidateModule accepted swap with fewer than two stack values")
		}
		if !strings.Contains(err.Error(), "OpSwap") {
			t.Fatalf("ValidateModule error %q does not include readable opcode name", err)
		}
	})
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

func moduleWithCallee() *Module {
	return &Module{
		Target: DefaultTarget(),
		Globals: []Global{
			{ID: 0, Name: "main", Kind: GlobalFunc, Func: 0},
			{ID: 1, Name: "callee", Kind: GlobalFunc, Func: 1},
		},
		Sigs: []FuncSig{
			{ID: 0, Ret: TypeI32},
			{ID: 1, Ret: TypeI32},
		},
		Functions: []Function{
			{
				ID:       0,
				GlobalID: 0,
				Name:     "main",
				Sig:      0,
				Instrs: []Instr{
					I32Const(0),
					Return(TypeI32),
				},
			},
			{
				ID:       1,
				GlobalID: 1,
				Name:     "callee",
				Sig:      1,
				Instrs: []Instr{
					I32Const(0),
					Return(TypeI32),
				},
			},
		},
	}
}
