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
	_, err := runProgram(t, context.Background(), testMainModule(bytecode.Jump(0)), RunOptions{})
	if err == nil || !strings.Contains(err.Error(), "unsupported opcode OpJump") {
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
