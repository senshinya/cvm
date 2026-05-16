package runtime

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"shinya.click/cvm/bytecode"
)

func runModule(t *testing.T, mod *bytecode.Module) (ExitStatus, error) {
	return runModuleWithOptions(t, context.Background(), mod, RunOptions{})
}

func runModuleWithOptions(t *testing.T, ctx context.Context, mod *bytecode.Module, opts RunOptions) (ExitStatus, error) {
	t.Helper()
	var buf bytes.Buffer
	if err := bytecode.EncodeModule(&buf, mod); err != nil {
		t.Fatalf("EncodeModule: %v", err)
	}
	p, err := Load(bytes.NewReader(buf.Bytes()), LoadOptions{Externs: DefaultExternRegistry(nil, nil)})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	return Run(ctx, p, opts)
}

func runProgram(t *testing.T, ctx context.Context, mod *bytecode.Module, opts RunOptions) (ExitStatus, error) {
	t.Helper()
	return Run(ctx, &Program{module: mod, entryFunc: 0}, opts)
}

func TestRunReturnsMainConstant(t *testing.T) {
	st, err := runModule(t, testMainModule(bytecode.I32Const(5), bytecode.Return(bytecode.TypeI32)))
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if st.Code != 5 {
		t.Fatalf("exit code = %d, want 5", st.Code)
	}
}

func TestRunLocalStoreLoad(t *testing.T) {
	mod := testMainModule(
		bytecode.I32Const(9),
		bytecode.StoreLocal(bytecode.TypeI32, 0),
		bytecode.LoadLocal(bytecode.TypeI32, 0),
		bytecode.Return(bytecode.TypeI32),
	)
	mod.Functions[0].Locals = []bytecode.LocalSlot{{ID: 0, Name: "x", Type: bytecode.TypeI32}}
	st, err := runModule(t, mod)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if st.Code != 9 {
		t.Fatalf("exit code = %d, want 9", st.Code)
	}
}

func TestRunIntegerArithmeticAndBranch(t *testing.T) {
	mod := testMainModule(
		bytecode.I32Const(3),
		bytecode.I32Const(4),
		bytecode.Binary(bytecode.TypeI32, bytecode.BinAdd),
		bytecode.I32Const(7),
		bytecode.Binary(bytecode.TypeI32, bytecode.BinEq),
		bytecode.JumpIfZero(bytecode.TypeBool, 1),
		bytecode.I32Const(11),
		bytecode.Return(bytecode.TypeI32),
		bytecode.LabelInstr(1),
		bytecode.I32Const(12),
		bytecode.Return(bytecode.TypeI32),
	)
	mod.Functions[0].Labels = []bytecode.Label{{ID: 1, Name: "else"}}
	st, err := runModule(t, mod)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if st.Code != 11 {
		t.Fatalf("exit code = %d, want 11", st.Code)
	}
}

func TestRunSwitch(t *testing.T) {
	mod := testMainModule(
		bytecode.I32Const(2),
		bytecode.Instr{Op: bytecode.OpSwitch, Type: bytecode.TypeI32, Label: 9, Labels: []bytecode.SwitchCase{{Value: 2, Label: 2}}},
		bytecode.LabelInstr(9),
		bytecode.I32Const(1),
		bytecode.Return(bytecode.TypeI32),
		bytecode.LabelInstr(2),
		bytecode.I32Const(22),
		bytecode.Return(bytecode.TypeI32),
	)
	mod.Functions[0].Labels = []bytecode.Label{{ID: 9, Name: "default"}, {ID: 2, Name: "two"}}
	st, err := runModule(t, mod)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if st.Code != 22 {
		t.Fatalf("exit code = %d, want 22", st.Code)
	}
}

func TestRunUnsignedSwitchUsesUnsignedCaseValues(t *testing.T) {
	mod := testMainModule(
		bytecode.U32Const(0x80000000),
		bytecode.Instr{Op: bytecode.OpSwitch, Type: bytecode.TypeU32, Label: 9, Labels: []bytecode.SwitchCase{{Value: 2147483648, Label: 2}}},
		bytecode.LabelInstr(9),
		bytecode.I32Const(1),
		bytecode.Return(bytecode.TypeI32),
		bytecode.LabelInstr(2),
		bytecode.I32Const(22),
		bytecode.Return(bytecode.TypeI32),
	)
	mod.Functions[0].Labels = []bytecode.Label{{ID: 9, Name: "default"}, {ID: 2, Name: "high"}}
	st, err := runModule(t, mod)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if st.Code != 22 {
		t.Fatalf("exit code = %d, want 22", st.Code)
	}
}

func TestRunUnsignedSwitchDoesNotMatchNegativeCase(t *testing.T) {
	mod := testMainModule(
		bytecode.U32Const(0xffffffff),
		bytecode.Instr{Op: bytecode.OpSwitch, Type: bytecode.TypeU32, Label: 9, Labels: []bytecode.SwitchCase{{Value: -1, Label: 2}}},
		bytecode.LabelInstr(9),
		bytecode.I32Const(11),
		bytecode.Return(bytecode.TypeI32),
		bytecode.LabelInstr(2),
		bytecode.I32Const(22),
		bytecode.Return(bytecode.TypeI32),
	)
	mod.Functions[0].Labels = []bytecode.Label{{ID: 9, Name: "default"}, {ID: 2, Name: "negative"}}
	st, err := runModule(t, mod)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if st.Code != 11 {
		t.Fatalf("exit code = %d, want 11", st.Code)
	}
}

func TestRunU64SwitchMaxValueDoesNotMatchNegativeCase(t *testing.T) {
	mod := testMainModule(
		bytecode.U64Const(^uint64(0)),
		bytecode.Instr{Op: bytecode.OpSwitch, Type: bytecode.TypeU64, Label: 9, Labels: []bytecode.SwitchCase{{Value: -1, Label: 2}}},
		bytecode.LabelInstr(9),
		bytecode.I32Const(11),
		bytecode.Return(bytecode.TypeI32),
		bytecode.LabelInstr(2),
		bytecode.I32Const(22),
		bytecode.Return(bytecode.TypeI32),
	)
	mod.Functions[0].Labels = []bytecode.Label{{ID: 9, Name: "default"}, {ID: 2, Name: "negative"}}
	st, err := runModule(t, mod)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if st.Code != 11 {
		t.Fatalf("exit code = %d, want 11", st.Code)
	}
}

func TestRunShiftTrapsOnNegativeCount(t *testing.T) {
	_, err := runProgram(t, context.Background(), testMainModule(
		bytecode.I32Const(1),
		bytecode.I32Const(-1),
		bytecode.Binary(bytecode.TypeI32, bytecode.BinShl),
		bytecode.Return(bytecode.TypeI32),
	), RunOptions{})
	if err == nil || !strings.Contains(err.Error(), "invalid shift count") {
		t.Fatalf("Run error = %v, want invalid shift count", err)
	}
}

func TestRunShiftTrapsOnOversizedCount(t *testing.T) {
	_, err := runProgram(t, context.Background(), testMainModule(
		bytecode.U32Const(1),
		bytecode.U32Const(32),
		bytecode.Binary(bytecode.TypeU32, bytecode.BinShl),
		bytecode.Return(bytecode.TypeU32),
	), RunOptions{})
	if err == nil || !strings.Contains(err.Error(), "invalid shift count") {
		t.Fatalf("Run error = %v, want invalid shift count", err)
	}
}

func TestRunUnsignedComparisonUsesUnsignedValues(t *testing.T) {
	mod := testMainModule(
		bytecode.U32Const(0xffffffff),
		bytecode.U32Const(1),
		bytecode.Binary(bytecode.TypeU32, bytecode.BinGtU),
		bytecode.Cast(bytecode.TypeBool, bytecode.TypeI32, bytecode.CastZExt),
		bytecode.Return(bytecode.TypeI32),
	)
	st, err := runModule(t, mod)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if st.Code != 1 {
		t.Fatalf("exit code = %d, want 1", st.Code)
	}
}

func TestRunIntegerCasts(t *testing.T) {
	t.Run("trunc zext", func(t *testing.T) {
		mod := testMainModule(
			bytecode.U32Const(0x1ff),
			bytecode.Cast(bytecode.TypeU32, bytecode.TypeU8, bytecode.CastTrunc),
			bytecode.Cast(bytecode.TypeU8, bytecode.TypeI32, bytecode.CastZExt),
			bytecode.Return(bytecode.TypeI32),
		)
		st, err := runModule(t, mod)
		if err != nil {
			t.Fatalf("Run: %v", err)
		}
		if st.Code != 255 {
			t.Fatalf("exit code = %d, want 255", st.Code)
		}
	})

	t.Run("sext", func(t *testing.T) {
		mod := testMainModule(
			bytecode.Const(bytecode.TypeI8, -1),
			bytecode.Cast(bytecode.TypeI8, bytecode.TypeI32, bytecode.CastSExt),
			bytecode.Return(bytecode.TypeI32),
		)
		st, err := runModule(t, mod)
		if err != nil {
			t.Fatalf("Run: %v", err)
		}
		if st.Code != -1 {
			t.Fatalf("exit code = %d, want -1", st.Code)
		}
	})
}

func TestRunBoolCastRejectsNonBoolResultType(t *testing.T) {
	_, err := runModule(t, testMainModule(
		bytecode.I32Const(7),
		bytecode.Cast(bytecode.TypeI32, bytecode.TypeI32, bytecode.CastBool),
		bytecode.Return(bytecode.TypeI32),
	))
	if err == nil || !strings.Contains(err.Error(), "bool cast result type") {
		t.Fatalf("Run error = %v, want bool cast result type", err)
	}
}

func TestRunBoolCastFromIntAndPointer(t *testing.T) {
	t.Run("int", func(t *testing.T) {
		mod := testMainModule(
			bytecode.I32Const(7),
			bytecode.Cast(bytecode.TypeI32, bytecode.TypeBool, bytecode.CastBool),
			bytecode.Cast(bytecode.TypeBool, bytecode.TypeI32, bytecode.CastZExt),
			bytecode.Return(bytecode.TypeI32),
		)
		st, err := runModule(t, mod)
		if err != nil {
			t.Fatalf("Run: %v", err)
		}
		if st.Code != 1 {
			t.Fatalf("exit code = %d, want 1", st.Code)
		}
	})

	t.Run("pointer", func(t *testing.T) {
		st, err := runProgram(t, context.Background(), testMainModule(
			bytecode.Instr{Op: bytecode.OpConst, Type: bytecode.TypePtr, Int: 5},
			bytecode.Cast(bytecode.TypePtr, bytecode.TypeBool, bytecode.CastBool),
			bytecode.Return(bytecode.TypeBool),
		), RunOptions{})
		if err != nil {
			t.Fatalf("Run: %v", err)
		}
		if st.Code != 1 {
			t.Fatalf("exit code = %d, want 1", st.Code)
		}
	})
}

func TestRunMissingLabelTrap(t *testing.T) {
	_, err := runProgram(t, context.Background(), testMainModule(bytecode.Jump(99)), RunOptions{})
	if err == nil || !strings.Contains(err.Error(), "missing label L99") {
		t.Fatalf("Run error = %v, want missing label", err)
	}
}

func TestRunRejectsNegativeLocalSlotWithoutPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Run panicked: %v", r)
		}
	}()
	mod := testMainModule(bytecode.I32Const(0), bytecode.Return(bytecode.TypeI32))
	mod.Functions[0].Locals = []bytecode.LocalSlot{{ID: -1, Name: "bad", Type: bytecode.TypeI32}}
	_, err := runProgram(t, context.Background(), mod, RunOptions{})
	if err == nil || !strings.Contains(err.Error(), "negative local slot") {
		t.Fatalf("Run error = %v, want negative local slot", err)
	}
}

func TestRunRejectsNegativeParamSlotWithoutPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Run panicked: %v", r)
		}
	}()
	mod := testMainModule(bytecode.I32Const(0), bytecode.Return(bytecode.TypeI32))
	mod.Functions[0].Params = []bytecode.Param{{Name: "bad", Type: bytecode.TypeI32, Slot: -1}}
	_, err := runProgram(t, context.Background(), mod, RunOptions{})
	if err == nil || !strings.Contains(err.Error(), "negative param slot") {
		t.Fatalf("Run error = %v, want negative param slot", err)
	}
}

func TestRunStepLimitReportsNextPC(t *testing.T) {
	_, err := runModuleWithOptions(t, context.Background(),
		testMainModule(bytecode.I32Const(5), bytecode.Return(bytecode.TypeI32)),
		RunOptions{StepLimit: 1},
	)
	if err == nil {
		t.Fatal("Run error = nil, want step limit")
	}
	msg := err.Error()
	if !strings.Contains(msg, "step limit") {
		t.Fatalf("Run error = %v, want step limit", err)
	}
	if !strings.Contains(msg, "pc=1") || strings.Contains(msg, "pc=0 opcode=OpConst") {
		t.Fatalf("Run error = %v, want next pc context", err)
	}
}

func TestRunUnsupportedOpcodeTrap(t *testing.T) {
	_, err := runProgram(t, context.Background(), testMainModule(bytecode.Instr{Op: bytecode.OpUnreachable}), RunOptions{})
	if err == nil || !strings.Contains(err.Error(), "unsupported opcode OpUnreachable") {
		t.Fatalf("Run error = %v, want unsupported opcode trap", err)
	}
}

func TestRunStackUnderflowTrap(t *testing.T) {
	_, err := runProgram(t, context.Background(), testMainModule(bytecode.Return(bytecode.TypeI32)), RunOptions{})
	if err == nil || !strings.Contains(err.Error(), "stack underflow") {
		t.Fatalf("Run error = %v, want stack underflow", err)
	}
}

func TestRunTypeMismatchTrap(t *testing.T) {
	_, err := runProgram(t, context.Background(), testMainModule(bytecode.I32Const(1), bytecode.Return(bytecode.TypeI64)), RunOptions{})
	if err == nil || !strings.Contains(err.Error(), "type i32, want i64") {
		t.Fatalf("Run error = %v, want type mismatch", err)
	}
}

func TestRunInvalidLocalSlotTrap(t *testing.T) {
	_, err := runProgram(t, context.Background(), testMainModule(bytecode.LoadLocal(bytecode.TypeI32, 0)), RunOptions{})
	if err == nil || !strings.Contains(err.Error(), "local slot out of range") {
		t.Fatalf("Run error = %v, want invalid local slot", err)
	}
}

func TestRunReturnVoidFromMainTrap(t *testing.T) {
	_, err := runProgram(t, context.Background(), testMainModule(bytecode.Instr{Op: bytecode.OpReturnVoid}), RunOptions{})
	if err == nil || !strings.Contains(err.Error(), "void return") {
		t.Fatalf("Run error = %v, want void return trap", err)
	}
}

func TestRunCanceledContextTrap(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := runProgram(t, ctx, testMainModule(bytecode.I32Const(0), bytecode.Return(bytecode.TypeI32)), RunOptions{})
	if err == nil {
		t.Fatal("Run error = nil, want canceled context")
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Run error = %v, want context.Canceled", err)
	}
}
