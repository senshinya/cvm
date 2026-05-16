package runtime

import (
	"bytes"
	"context"
	"errors"
	"math"
	"strings"
	"testing"

	"shinya.click/cvm/bytecode"
)

func runModule(t *testing.T, mod *bytecode.Module) (ExitStatus, error) {
	return runModuleWithOptions(t, context.Background(), mod, RunOptions{})
}

func runModuleWithOptions(t *testing.T, ctx context.Context, mod *bytecode.Module, opts RunOptions) (ExitStatus, error) {
	t.Helper()
	return runModuleWithExterns(t, ctx, mod, opts, DefaultExternRegistry(nil, nil))
}

func runModuleWithExterns(t *testing.T, ctx context.Context, mod *bytecode.Module, opts RunOptions, externs *ExternRegistry) (ExitStatus, error) {
	t.Helper()
	var buf bytes.Buffer
	if err := bytecode.EncodeModule(&buf, mod); err != nil {
		t.Fatalf("EncodeModule: %v", err)
	}
	p, err := Load(bytes.NewReader(buf.Bytes()), LoadOptions{Externs: externs})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	return Run(ctx, p, opts)
}

func runProgram(t *testing.T, ctx context.Context, mod *bytecode.Module, opts RunOptions) (ExitStatus, error) {
	t.Helper()
	return Run(ctx, &Program{module: mod, memory: NewMemory(mod.Target), entryFunc: 0}, opts)
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

func TestRunDirectCall(t *testing.T) {
	mod := testMainModule(
		bytecode.I32Const(20),
		bytecode.I32Const(22),
		bytecode.Call(1, 1, 2),
		bytecode.Return(bytecode.TypeI32),
	)
	mod.Sigs = append(mod.Sigs, bytecode.FuncSig{ID: 1, Ret: bytecode.TypeI32, Params: []bytecode.ValueType{bytecode.TypeI32, bytecode.TypeI32}})
	mod.Globals = append(mod.Globals, bytecode.Global{ID: 1, Name: "add", Kind: bytecode.GlobalFunc, Func: 1, Sig: 1})
	mod.Functions = append(mod.Functions, bytecode.Function{
		ID:       1,
		GlobalID: 1,
		Name:     "add",
		Sig:      1,
		Params: []bytecode.Param{
			{Name: "a", Type: bytecode.TypeI32, Slot: 0},
			{Name: "b", Type: bytecode.TypeI32, Slot: 1},
		},
		Instrs: []bytecode.Instr{
			bytecode.LoadLocal(bytecode.TypeI32, 0),
			bytecode.LoadLocal(bytecode.TypeI32, 1),
			bytecode.Binary(bytecode.TypeI32, bytecode.BinAdd),
			bytecode.Return(bytecode.TypeI32),
		},
		MaxStack: 2,
	})

	st, err := runModule(t, mod)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if st.Code != 42 {
		t.Fatalf("exit code = %d, want 42", st.Code)
	}
}

func TestRunVariadicDirectCallConsumesExtraArgs(t *testing.T) {
	mod := testMainModule(
		bytecode.I32Const(7),
		bytecode.I64Const(9),
		bytecode.Call(1, 1, 2),
		bytecode.Return(bytecode.TypeI32),
	)
	mod.Sigs = append(mod.Sigs, bytecode.FuncSig{ID: 1, Ret: bytecode.TypeI32, Params: []bytecode.ValueType{bytecode.TypeI32}, Variadic: true})
	mod.Globals = append(mod.Globals, bytecode.Global{ID: 1, Name: "first", Kind: bytecode.GlobalFunc, Func: 1, Sig: 1})
	mod.Functions = append(mod.Functions, bytecode.Function{
		ID:       1,
		GlobalID: 1,
		Name:     "first",
		Sig:      1,
		Params: []bytecode.Param{
			{Name: "a", Type: bytecode.TypeI32, Slot: 0},
		},
		Instrs: []bytecode.Instr{
			bytecode.LoadLocal(bytecode.TypeI32, 0),
			bytecode.Return(bytecode.TypeI32),
		},
		MaxStack: 1,
	})

	st, err := runModule(t, mod)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if st.Code != 7 {
		t.Fatalf("exit code = %d, want 7", st.Code)
	}
}

func TestRunIndirectCall(t *testing.T) {
	mod := testMainModule(
		bytecode.AddrFunc(1),
		bytecode.I32Const(40),
		bytecode.I32Const(2),
		bytecode.Instr{Op: bytecode.OpCallIndirect, Sig: 1, Argc: 2},
		bytecode.Return(bytecode.TypeI32),
	)
	mod.Sigs = append(mod.Sigs, bytecode.FuncSig{ID: 1, Ret: bytecode.TypeI32, Params: []bytecode.ValueType{bytecode.TypeI32, bytecode.TypeI32}})
	mod.Globals = append(mod.Globals, bytecode.Global{ID: 1, Name: "add", Kind: bytecode.GlobalFunc, Func: 1, Sig: 1})
	mod.Functions = append(mod.Functions, bytecode.Function{
		ID:       1,
		GlobalID: 1,
		Name:     "add",
		Sig:      1,
		Params: []bytecode.Param{
			{Name: "a", Type: bytecode.TypeI32, Slot: 0},
			{Name: "b", Type: bytecode.TypeI32, Slot: 1},
		},
		Instrs: []bytecode.Instr{
			bytecode.LoadLocal(bytecode.TypeI32, 0),
			bytecode.LoadLocal(bytecode.TypeI32, 1),
			bytecode.Binary(bytecode.TypeI32, bytecode.BinAdd),
			bytecode.Return(bytecode.TypeI32),
		},
		MaxStack: 2,
	})

	st, err := runModule(t, mod)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if st.Code != 42 {
		t.Fatalf("exit code = %d, want 42", st.Code)
	}
}

func TestRunExitExtern(t *testing.T) {
	mod := testMainModule(
		bytecode.I32Const(17),
		bytecode.Call(1, 1, 1),
		bytecode.I32Const(0),
		bytecode.Return(bytecode.TypeI32),
	)
	mod.Sigs = append(mod.Sigs, bytecode.FuncSig{ID: 1, Ret: bytecode.TypeVoid, Params: []bytecode.ValueType{bytecode.TypeI32}})
	mod.Globals = append(mod.Globals, bytecode.Global{
		ID:     1,
		Name:   "exit",
		Kind:   bytecode.GlobalExtern,
		Sig:    1,
		Extern: bytecode.ExternRef{Name: "exit", ABI: bytecode.DefaultExternABI},
	})

	st, err := runModule(t, mod)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if st.Code != 17 {
		t.Fatalf("exit code = %d, want 17", st.Code)
	}
}

func TestRunPutsExtern(t *testing.T) {
	var stdout bytes.Buffer
	mod := testMainModule(
		bytecode.AddrString(0),
		bytecode.Call(1, 1, 1),
		bytecode.Return(bytecode.TypeI32),
	)
	mod.Sigs = append(mod.Sigs, bytecode.FuncSig{ID: 1, Ret: bytecode.TypeI32, Params: []bytecode.ValueType{bytecode.TypeObjectAddr}})
	mod.Globals = append(mod.Globals, bytecode.Global{
		ID:     1,
		Name:   "puts",
		Kind:   bytecode.GlobalExtern,
		Sig:    1,
		Extern: bytecode.ExternRef{Name: "puts", ABI: bytecode.DefaultExternABI},
	})
	mod.Strings = []bytecode.StringConst{{ID: 0, Value: "hi", Bytes: []byte("hi\x00")}}

	st, err := runModuleWithExterns(t, context.Background(), mod, RunOptions{}, DefaultExternRegistry(&stdout, nil))
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if st.Code != 3 {
		t.Fatalf("exit code = %d, want 3", st.Code)
	}
	if got := stdout.String(); got != "hi\n" {
		t.Fatalf("stdout = %q, want %q", got, "hi\n")
	}
}

func TestRunVariadicExternPreservesArgumentOrder(t *testing.T) {
	reg := NewExternRegistry(nil, nil)
	reg.Register("probe", func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 3 {
			return Value{}, nil, errors.New("probe expects 3 arguments")
		}
		if args[0].Type != bytecode.TypeI32 || args[0].Int != 7 {
			return Value{}, nil, errors.New("arg0 mismatch")
		}
		if args[1].Type != bytecode.TypeI64 || args[1].Int != 9 {
			return Value{}, nil, errors.New("arg1 mismatch")
		}
		if args[2].Type != bytecode.TypeObjectAddr {
			return Value{}, nil, errors.New("arg2 mismatch")
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, errors.New("missing extern context")
		}
		return IntValue(bytecode.TypeI32, 23), nil, nil
	})

	mod := testMainModule(
		bytecode.I32Const(7),
		bytecode.I64Const(9),
		bytecode.AddrString(0),
		bytecode.Call(1, 1, 3),
		bytecode.Return(bytecode.TypeI32),
	)
	mod.Sigs = append(mod.Sigs, bytecode.FuncSig{ID: 1, Ret: bytecode.TypeI32, Params: []bytecode.ValueType{bytecode.TypeI32}, Variadic: true})
	mod.Globals = append(mod.Globals, bytecode.Global{
		ID:     1,
		Name:   "probe",
		Kind:   bytecode.GlobalExtern,
		Sig:    1,
		Extern: bytecode.ExternRef{Name: "probe", ABI: bytecode.DefaultExternABI},
	})
	mod.Strings = []bytecode.StringConst{{ID: 0, Value: "x", Bytes: []byte("x\x00")}}

	st, err := runModuleWithExterns(t, context.Background(), mod, RunOptions{}, reg)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if st.Code != 23 {
		t.Fatalf("exit code = %d, want 23", st.Code)
	}
}

func TestRunInvalidIndirectCallTargetTraps(t *testing.T) {
	mod := testMainModule(
		bytecode.Instr{Op: bytecode.OpConst, Type: bytecode.TypePtr, Int: 0x1234},
		bytecode.Instr{Op: bytecode.OpCallIndirect, Sig: 1, Argc: 0},
		bytecode.Return(bytecode.TypeI32),
	)
	mod.Sigs = append(mod.Sigs, bytecode.FuncSig{ID: 1, Ret: bytecode.TypeI32})

	_, err := runProgram(t, context.Background(), mod, RunOptions{})
	if err == nil || !strings.Contains(err.Error(), "invalid indirect call target") {
		t.Fatalf("Run error = %v, want invalid indirect call target", err)
	}
}

func TestRunCallRejectsBadArgc(t *testing.T) {
	mod := testMainModule(
		bytecode.I32Const(1),
		bytecode.I32Const(2),
		bytecode.Call(1, 1, 2),
		bytecode.Return(bytecode.TypeI32),
	)
	mod.Sigs = append(mod.Sigs, bytecode.FuncSig{ID: 1, Ret: bytecode.TypeI32, Params: []bytecode.ValueType{bytecode.TypeI32}})
	mod.Globals = append(mod.Globals, bytecode.Global{
		ID:     1,
		Name:   "id",
		Kind:   bytecode.GlobalExtern,
		Sig:    1,
		Extern: bytecode.ExternRef{Name: "id", ABI: bytecode.DefaultExternABI},
	})

	_, err := Run(context.Background(), &Program{
		module: mod,
		memory: NewMemory(mod.Target),
		externs: map[int]ExternFunc{1: func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
			return args[0], nil, nil
		}},
		entryFunc: 0,
	}, RunOptions{})
	if err == nil || !strings.Contains(err.Error(), "argc") {
		t.Fatalf("Run error = %v, want argc trap", err)
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

func TestRunLocalObjectAddress(t *testing.T) {
	mod := testMainModule(
		bytecode.AddrLocalObject(0),
		bytecode.I32Const(44),
		bytecode.Store(bytecode.TypeI32, 4, false),
		bytecode.AddrLocalObject(0),
		bytecode.Load(bytecode.TypeI32, 4, false),
		bytecode.Return(bytecode.TypeI32),
	)
	mod.Layouts = []bytecode.ObjectLayout{{ID: 0, Name: "local", Size: 4, Align: 4}}
	mod.Functions[0].Objects = []bytecode.LocalObject{{ID: 0, Name: "obj", Size: 4, Align: 4, Layout: 0}}

	st, err := runModule(t, mod)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if st.Code != 44 {
		t.Fatalf("exit code = %d, want 44", st.Code)
	}
}

func TestRunLocalObjectBadIDTraps(t *testing.T) {
	_, err := runProgram(t, context.Background(), testMainModule(bytecode.AddrLocalObject(99)), RunOptions{})
	if err == nil || !strings.Contains(err.Error(), "invalid local object 99") {
		t.Fatalf("Run error = %v, want invalid local object trap", err)
	}
}

func TestRunFieldAddress(t *testing.T) {
	mod := testMainModule(
		bytecode.AddrLocalObject(0),
		bytecode.Instr{Op: bytecode.OpFieldAddr, Layout: 0, Field: 0},
		bytecode.I32Const(44),
		bytecode.Store(bytecode.TypeI32, 4, false),
		bytecode.AddrLocalObject(0),
		bytecode.Instr{Op: bytecode.OpFieldAddr, Layout: 0, Field: 0},
		bytecode.Load(bytecode.TypeI32, 4, false),
		bytecode.Return(bytecode.TypeI32),
	)
	mod.Layouts = []bytecode.ObjectLayout{{
		ID:     0,
		Name:   "record",
		Size:   8,
		Align:  4,
		Fields: []bytecode.FieldLayout{{ID: 0, Name: "field", Offset: 4, Type: bytecode.TypeI32}},
	}}
	mod.Functions[0].Objects = []bytecode.LocalObject{{ID: 0, Name: "obj", Size: 8, Align: 4, Layout: 0}}

	st, err := runModule(t, mod)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if st.Code != 44 {
		t.Fatalf("exit code = %d, want 44", st.Code)
	}
}

func TestRunFieldAddressBadFieldTraps(t *testing.T) {
	mod := testMainModule(
		bytecode.Instr{Op: bytecode.OpConst, Type: bytecode.TypeObjectAddr, Int: 0x1000},
		bytecode.Instr{Op: bytecode.OpFieldAddr, Layout: 0, Field: 99},
	)
	mod.Layouts = []bytecode.ObjectLayout{{ID: 0, Name: "record"}}

	_, err := runProgram(t, context.Background(), mod, RunOptions{})
	if err == nil || !strings.Contains(err.Error(), "invalid field 99 in layout 0") {
		t.Fatalf("Run error = %v, want invalid field trap", err)
	}
}

func TestRunFieldAddressOffsetOverflowTraps(t *testing.T) {
	mod := testMainModule(
		bytecode.Instr{Op: bytecode.OpConst, Type: bytecode.TypeObjectAddr, Int: -1},
		bytecode.Instr{Op: bytecode.OpFieldAddr, Layout: 0, Field: 0},
	)
	mod.Layouts = []bytecode.ObjectLayout{{
		ID:     0,
		Name:   "record",
		Fields: []bytecode.FieldLayout{{ID: 0, Name: "field", Offset: 1, Type: bytecode.TypeI32}},
	}}

	_, err := runProgram(t, context.Background(), mod, RunOptions{})
	if err == nil || !strings.Contains(err.Error(), "field offset overflow") {
		t.Fatalf("Run error = %v, want field offset overflow trap", err)
	}
}

func TestRunDynamicObjectAddress(t *testing.T) {
	mod := testMainModule(
		bytecode.I64Const(4),
		bytecode.Instr{Op: bytecode.OpAllocDynamicObject, Object: 0, Type: bytecode.TypeI64, Align: 4, Layout: 0},
		bytecode.Instr{Op: bytecode.OpDynamicObjectAddr, Object: 0, Type: bytecode.TypeObjectAddr},
		bytecode.I32Const(44),
		bytecode.Store(bytecode.TypeI32, 4, false),
		bytecode.Instr{Op: bytecode.OpDynamicObjectAddr, Object: 0, Type: bytecode.TypeObjectAddr},
		bytecode.Load(bytecode.TypeI32, 4, false),
		bytecode.Instr{Op: bytecode.OpFreeDynamicObject, Object: 0},
		bytecode.Return(bytecode.TypeI32),
	)
	mod.Layouts = []bytecode.ObjectLayout{{ID: 0, Name: "vla", Align: 4}}
	mod.Functions[0].DynamicObjects = []bytecode.DynamicObject{{ID: 0, Name: "dyn", Align: 4, Layout: 0}}

	st, err := runModule(t, mod)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if st.Code != 44 {
		t.Fatalf("exit code = %d, want 44", st.Code)
	}
}

func TestRunDynamicObjectBadIDTraps(t *testing.T) {
	mod := testMainModule(
		bytecode.I64Const(4),
		bytecode.Instr{Op: bytecode.OpAllocDynamicObject, Object: 99, Type: bytecode.TypeI64, Align: 4, Layout: 0},
	)

	_, err := runProgram(t, context.Background(), mod, RunOptions{})
	if err == nil || !strings.Contains(err.Error(), "invalid dynamic object 99") {
		t.Fatalf("Run error = %v, want invalid dynamic object trap", err)
	}
}

func TestRunDynamicObjectAddressBeforeAllocTraps(t *testing.T) {
	mod := testMainModule(
		bytecode.Instr{Op: bytecode.OpDynamicObjectAddr, Object: 0, Type: bytecode.TypeObjectAddr},
		bytecode.Return(bytecode.TypeI32),
	)
	mod.Layouts = []bytecode.ObjectLayout{{ID: 0, Name: "vla", Align: 4}}
	mod.Functions[0].DynamicObjects = []bytecode.DynamicObject{{ID: 0, Name: "dyn", Align: 4, Layout: 0}}

	_, err := runProgram(t, context.Background(), mod, RunOptions{})
	if err == nil || !strings.Contains(err.Error(), "dynamic object 0 is not allocated") {
		t.Fatalf("Run error = %v, want dynamic object before alloc trap", err)
	}
}

func TestRunDynamicObjectAddressAfterFreeTraps(t *testing.T) {
	mod := testMainModule(
		bytecode.I64Const(4),
		bytecode.Instr{Op: bytecode.OpAllocDynamicObject, Object: 0, Type: bytecode.TypeI64, Align: 4, Layout: 0},
		bytecode.Instr{Op: bytecode.OpFreeDynamicObject, Object: 0},
		bytecode.Instr{Op: bytecode.OpDynamicObjectAddr, Object: 0, Type: bytecode.TypeObjectAddr},
		bytecode.Return(bytecode.TypeI32),
	)
	mod.Layouts = []bytecode.ObjectLayout{{ID: 0, Name: "vla", Align: 4}}
	mod.Functions[0].DynamicObjects = []bytecode.DynamicObject{{ID: 0, Name: "dyn", Align: 4, Layout: 0}}

	_, err := runProgram(t, context.Background(), mod, RunOptions{})
	if err == nil || !strings.Contains(err.Error(), "dynamic object 0 is not allocated") {
		t.Fatalf("Run error = %v, want dynamic object after free trap", err)
	}
}

func TestRunDynamicObjectStaleAddressAfterFreeTraps(t *testing.T) {
	mod := testMainModule(
		bytecode.I64Const(4),
		bytecode.Instr{Op: bytecode.OpAllocDynamicObject, Object: 0, Type: bytecode.TypeI64, Align: 4, Layout: 0},
		bytecode.Instr{Op: bytecode.OpDynamicObjectAddr, Object: 0, Type: bytecode.TypeObjectAddr},
		bytecode.Instr{Op: bytecode.OpDup},
		bytecode.Instr{Op: bytecode.OpFreeDynamicObject, Object: 0},
		bytecode.I32Const(1),
		bytecode.Store(bytecode.TypeI32, 4, false),
		bytecode.Return(bytecode.TypeI32),
	)
	mod.Layouts = []bytecode.ObjectLayout{{ID: 0, Name: "vla", Align: 4}}
	mod.Functions[0].DynamicObjects = []bytecode.DynamicObject{{ID: 0, Name: "dyn", Align: 4, Layout: 0}}

	_, err := runProgram(t, context.Background(), mod, RunOptions{})
	if err == nil || !strings.Contains(err.Error(), "use after free") {
		t.Fatalf("Run error = %v, want use after free trap", err)
	}
}

func TestRunDynamicObjectRejectsDuplicateAlloc(t *testing.T) {
	mod := testMainModule(
		bytecode.I64Const(4),
		bytecode.Instr{Op: bytecode.OpAllocDynamicObject, Object: 0, Type: bytecode.TypeI64, Align: 4, Layout: 0},
		bytecode.I64Const(4),
		bytecode.Instr{Op: bytecode.OpAllocDynamicObject, Object: 0, Type: bytecode.TypeI64, Align: 4, Layout: 0},
		bytecode.Return(bytecode.TypeI32),
	)
	mod.Layouts = []bytecode.ObjectLayout{{ID: 0, Name: "vla", Align: 4}}
	mod.Functions[0].DynamicObjects = []bytecode.DynamicObject{{ID: 0, Name: "dyn", Align: 4, Layout: 0}}

	_, err := runProgram(t, context.Background(), mod, RunOptions{})
	if err == nil || !strings.Contains(err.Error(), "dynamic object 0 is already allocated") {
		t.Fatalf("Run error = %v, want duplicate allocation trap", err)
	}
}

func TestRunDynamicObjectRejectsLayoutMismatch(t *testing.T) {
	mod := testMainModule(
		bytecode.I64Const(4),
		bytecode.Instr{Op: bytecode.OpAllocDynamicObject, Object: 0, Type: bytecode.TypeI64, Align: 4, Layout: 1},
		bytecode.Return(bytecode.TypeI32),
	)
	mod.Layouts = []bytecode.ObjectLayout{
		{ID: 0, Name: "vla", Align: 4},
		{ID: 1, Name: "other", Align: 4},
	}
	mod.Functions[0].DynamicObjects = []bytecode.DynamicObject{{ID: 0, Name: "dyn", Align: 4, Layout: 0}}

	_, err := runProgram(t, context.Background(), mod, RunOptions{})
	if err == nil || !strings.Contains(err.Error(), "does not match instruction layout") {
		t.Fatalf("Run error = %v, want layout mismatch trap", err)
	}
}

func TestRunDynamicObjectRejectsNonObjectAddrType(t *testing.T) {
	mod := testMainModule(
		bytecode.I64Const(4),
		bytecode.Instr{Op: bytecode.OpAllocDynamicObject, Object: 0, Type: bytecode.TypeI64, Align: 4, Layout: 0},
		bytecode.Instr{Op: bytecode.OpDynamicObjectAddr, Object: 0, Type: bytecode.TypePtr},
		bytecode.Return(bytecode.TypeI32),
	)
	mod.Layouts = []bytecode.ObjectLayout{{ID: 0, Name: "vla", Align: 4}}
	mod.Functions[0].DynamicObjects = []bytecode.DynamicObject{{ID: 0, Name: "dyn", Align: 4, Layout: 0}}

	_, err := runProgram(t, context.Background(), mod, RunOptions{})
	if err == nil || !strings.Contains(err.Error(), "dynamic object address has type ptr") {
		t.Fatalf("Run error = %v, want dynamic object address type trap", err)
	}
}

func TestRunReturnedLocalObjectAddressTrapsAfterFrameExit(t *testing.T) {
	mod := testMainModule(
		bytecode.Call(1, 1, 0),
		bytecode.Load(bytecode.TypeI32, 4, false),
		bytecode.Return(bytecode.TypeI32),
	)
	mod.Sigs = append(mod.Sigs, bytecode.FuncSig{ID: 1, Ret: bytecode.TypeObjectAddr})
	mod.Globals = append(mod.Globals, bytecode.Global{ID: 1, Name: "leak", Kind: bytecode.GlobalFunc, Func: 1, Sig: 1})
	mod.Functions = append(mod.Functions, bytecode.Function{
		ID:       1,
		GlobalID: 1,
		Name:     "leak",
		Sig:      1,
		Objects:  []bytecode.LocalObject{{ID: 0, Name: "obj", Size: 4, Align: 4, Layout: 0}},
		Instrs: []bytecode.Instr{
			bytecode.AddrLocalObject(0),
			bytecode.I32Const(55),
			bytecode.Store(bytecode.TypeI32, 4, false),
			bytecode.AddrLocalObject(0),
			bytecode.Return(bytecode.TypeObjectAddr),
		},
		MaxStack: 2,
	})
	mod.Layouts = []bytecode.ObjectLayout{{ID: 0, Name: "local", Size: 4, Align: 4}}

	_, err := runProgram(t, context.Background(), mod, RunOptions{})
	if err == nil || !strings.Contains(err.Error(), "use after free") {
		t.Fatalf("Run error = %v, want use after free trap", err)
	}
}

func TestRunGlobalLoadStore(t *testing.T) {
	mod := testMainModule(
		bytecode.AddrGlobal(1),
		bytecode.I32Const(17),
		bytecode.Store(bytecode.TypeI32, 4, false),
		bytecode.AddrGlobal(1),
		bytecode.Load(bytecode.TypeI32, 4, false),
		bytecode.Return(bytecode.TypeI32),
	)
	mod.Globals = append(mod.Globals, bytecode.Global{ID: 1, Name: "g", Kind: bytecode.GlobalVar, Size: 4, Align: 4})
	st, err := runModule(t, mod)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if st.Code != 17 {
		t.Fatalf("exit code = %d, want 17", st.Code)
	}
}

func TestRunMemSetAndMemCopy(t *testing.T) {
	mod := testMainModule(
		bytecode.AddrGlobal(1),
		bytecode.I32Const(65),
		bytecode.Instr{Op: bytecode.OpMemSet, Size: 4, Align: 1},
		bytecode.AddrGlobal(2),
		bytecode.AddrGlobal(1),
		bytecode.Instr{Op: bytecode.OpMemCopy, Size: 4, Align: 1},
		bytecode.AddrGlobal(2),
		bytecode.Load(bytecode.TypeI32, 4, false),
		bytecode.Return(bytecode.TypeI32),
	)
	mod.Globals = append(mod.Globals,
		bytecode.Global{ID: 1, Name: "a", Kind: bytecode.GlobalVar, Size: 4, Align: 4},
		bytecode.Global{ID: 2, Name: "b", Kind: bytecode.GlobalVar, Size: 4, Align: 4},
	)
	st, err := runModule(t, mod)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if st.Code == 0 {
		t.Fatalf("exit code = 0, want copied nonzero bytes")
	}
}

func TestRunOffsetFollowsObjectAddrStackContract(t *testing.T) {
	mod := testMainModule(
		bytecode.AddrGlobal(1),
		bytecode.Instr{Op: bytecode.OpOffset, Type: bytecode.TypePtr, Int: 4},
		bytecode.I32Const(23),
		bytecode.Store(bytecode.TypeI32, 4, false),
		bytecode.AddrGlobal(1),
		bytecode.Instr{Op: bytecode.OpOffset, Type: bytecode.TypePtr, Int: 4},
		bytecode.Load(bytecode.TypeI32, 4, false),
		bytecode.Return(bytecode.TypeI32),
	)
	mod.Globals = append(mod.Globals, bytecode.Global{ID: 1, Name: "g", Kind: bytecode.GlobalVar, Size: 8, Align: 4})
	if err := bytecode.ValidateModule(mod); err != nil {
		t.Fatalf("ValidateModule: %v", err)
	}
	st, err := runModule(t, mod)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if st.Code != 23 {
		t.Fatalf("exit code = %d, want 23", st.Code)
	}
}

func TestRunPtrAddDynamic(t *testing.T) {
	mod := testMainModule(
		bytecode.AddrGlobal(1),
		bytecode.I32Const(2),
		bytecode.I64Const(4),
		bytecode.Instr{Op: bytecode.OpPtrAddDynamic},
		bytecode.I32Const(31),
		bytecode.Store(bytecode.TypeI32, 4, false),
		bytecode.AddrGlobal(1),
		bytecode.I32Const(2),
		bytecode.I64Const(4),
		bytecode.Instr{Op: bytecode.OpPtrAddDynamic},
		bytecode.Load(bytecode.TypeI32, 4, false),
		bytecode.Return(bytecode.TypeI32),
	)
	mod.Globals = append(mod.Globals, bytecode.Global{ID: 1, Name: "g", Kind: bytecode.GlobalVar, Size: 12, Align: 4})
	st, err := runModule(t, mod)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if st.Code != 31 {
		t.Fatalf("exit code = %d, want 31", st.Code)
	}
}

func TestRunPtrAddDynamicRejectsZeroStride(t *testing.T) {
	mod := testMainModule(
		bytecode.AddrGlobal(1),
		bytecode.I32Const(1),
		bytecode.I64Const(0),
		bytecode.Instr{Op: bytecode.OpPtrAddDynamic},
		bytecode.Instr{Op: bytecode.OpPop},
		bytecode.I32Const(0),
		bytecode.Return(bytecode.TypeI32),
	)
	mod.Globals = append(mod.Globals, bytecode.Global{ID: 1, Name: "g", Kind: bytecode.GlobalVar, Size: 4, Align: 4})
	_, err := runModule(t, mod)
	if err == nil || !strings.Contains(err.Error(), "invalid element size") {
		t.Fatalf("Run error = %v, want invalid element size", err)
	}
}

func TestRunLoadConst(t *testing.T) {
	mod := testMainModule(
		bytecode.Instr{Op: bytecode.OpLoadConst, Type: bytecode.TypeI32, Global: 1, Int: 4},
		bytecode.Return(bytecode.TypeI32),
	)
	mod.Globals = append(mod.Globals, bytecode.Global{
		ID:    1,
		Name:  "c",
		Kind:  bytecode.GlobalVar,
		Size:  8,
		Align: 4,
		Init:  bytecode.InitData{Bytes: []byte{0, 0, 0, 0, 41, 0, 0, 0}},
	})
	st, err := runModule(t, mod)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if st.Code != 41 {
		t.Fatalf("exit code = %d, want 41", st.Code)
	}
}

func TestRunAddressAccessorsTrapOnInvalidIDsWithoutPanic(t *testing.T) {
	tests := []struct {
		name string
		ins  bytecode.Instr
		want string
	}{
		{"global", bytecode.AddrGlobal(9), "invalid global id"},
		{"string", bytecode.AddrString(9), "invalid string id"},
		{"func", bytecode.AddrFunc(9), "invalid function address id"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("Run panicked: %v", r)
				}
			}()
			_, err := runProgram(t, context.Background(), testMainModule(tt.ins), RunOptions{})
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("Run error = %v, want %q", err, tt.want)
			}
		})
	}
}

func TestRunReadonlyGlobalWriteTraps(t *testing.T) {
	mod := testMainModule(
		bytecode.AddrGlobal(1),
		bytecode.I32Const(99),
		bytecode.Store(bytecode.TypeI32, 4, false),
		bytecode.I32Const(0),
		bytecode.Return(bytecode.TypeI32),
	)
	mod.Globals = append(mod.Globals, bytecode.Global{ID: 1, Name: "ro", Kind: bytecode.GlobalVar, Size: 4, Align: 4, Readonly: true})
	_, err := runModule(t, mod)
	if err == nil || !strings.Contains(err.Error(), "readonly") {
		t.Fatalf("Run error = %v, want readonly", err)
	}
}

func TestRunStringWriteTraps(t *testing.T) {
	mod := testMainModule(
		bytecode.AddrString(0),
		bytecode.I32Const(99),
		bytecode.Store(bytecode.TypeI32, 4, false),
		bytecode.I32Const(0),
		bytecode.Return(bytecode.TypeI32),
	)
	mod.Strings = []bytecode.StringConst{{ID: 0, Value: "test", Bytes: []byte("test\x00")}}
	_, err := runModule(t, mod)
	if err == nil || !strings.Contains(err.Error(), "readonly") {
		t.Fatalf("Run error = %v, want readonly", err)
	}
}

func TestRunPtrAddNegativeIndex(t *testing.T) {
	mod := testMainModule(
		bytecode.AddrGlobal(1),
		bytecode.Instr{Op: bytecode.OpOffset, Type: bytecode.TypeObjectAddr, Int: 8},
		bytecode.I32Const(-1),
		bytecode.Instr{Op: bytecode.OpPtrAdd, Size: 4},
		bytecode.I32Const(37),
		bytecode.Store(bytecode.TypeI32, 4, false),
		bytecode.AddrGlobal(1),
		bytecode.Instr{Op: bytecode.OpOffset, Type: bytecode.TypeObjectAddr, Int: 4},
		bytecode.Load(bytecode.TypeI32, 4, false),
		bytecode.Return(bytecode.TypeI32),
	)
	mod.Globals = append(mod.Globals, bytecode.Global{ID: 1, Name: "g", Kind: bytecode.GlobalVar, Size: 12, Align: 4})
	st, err := runModule(t, mod)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if st.Code != 37 {
		t.Fatalf("exit code = %d, want 37", st.Code)
	}
}

func TestRunPtrAddRejectsInvalidElementSize(t *testing.T) {
	mod := testMainModule(
		bytecode.AddrGlobal(1),
		bytecode.I32Const(1),
		bytecode.Instr{Op: bytecode.OpPtrAdd, Size: 0},
		bytecode.Instr{Op: bytecode.OpPop},
		bytecode.I32Const(0),
		bytecode.Return(bytecode.TypeI32),
	)
	mod.Globals = append(mod.Globals, bytecode.Global{ID: 1, Name: "g", Kind: bytecode.GlobalVar, Size: 4, Align: 4})
	_, err := runModule(t, mod)
	if err == nil || !strings.Contains(err.Error(), "invalid element size") {
		t.Fatalf("Run error = %v, want invalid element size", err)
	}
}

func TestVMPointerDiffChecksScaledQuotient(t *testing.T) {
	got, err := pointerDiff(uint64(math.MaxInt64)*2, 0, 2)
	if err != nil {
		t.Fatalf("pointerDiff: %v", err)
	}
	if got != math.MaxInt64 {
		t.Fatalf("pointerDiff = %d, want %d", got, int64(math.MaxInt64))
	}
}

func TestVMPointerDiffAllowsMinInt64(t *testing.T) {
	got, err := pointerDiff(0, uint64(1)<<63, 1)
	if err != nil {
		t.Fatalf("pointerDiff: %v", err)
	}
	if got != math.MinInt64 {
		t.Fatalf("pointerDiff = %d, want %d", got, int64(math.MinInt64))
	}
}

func TestVMPointerDiffRejectsNonDivisibleDistance(t *testing.T) {
	tests := []struct {
		name       string
		left       uint64
		right      uint64
		wantSubstr string
	}{
		{name: "positive", left: 9, right: 0, wantSubstr: "pointer difference 9 is not divisible"},
		{name: "negative", left: 0, right: 9, wantSubstr: "pointer difference -9 is not divisible"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := pointerDiff(tt.left, tt.right, 4)
			if err == nil || !strings.Contains(err.Error(), tt.wantSubstr) {
				t.Fatalf("pointerDiff error = %v, want %q", err, tt.wantSubstr)
			}
		})
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

func TestRunUnreachableTrap(t *testing.T) {
	_, err := runProgram(t, context.Background(), testMainModule(bytecode.Instr{Op: bytecode.OpUnreachable}), RunOptions{})
	if err == nil || !strings.Contains(err.Error(), "unreachable") || strings.Contains(err.Error(), "unsupported opcode") {
		t.Fatalf("Run error = %v, want unreachable trap", err)
	}
}

func TestRunUnsupportedOpcodeTrap(t *testing.T) {
	tests := []struct {
		name string
		ins  bytecode.Instr
		want string
	}{
		{name: "return object", ins: bytecode.Instr{Op: bytecode.OpReturnObject}, want: "unsupported opcode OpReturnObject"},
		{name: "bitfield load", ins: bytecode.Instr{Op: bytecode.OpBitFieldLoad, Type: bytecode.TypeI32}, want: "unsupported opcode OpBitFieldLoad"},
		{name: "bitfield store", ins: bytecode.Instr{Op: bytecode.OpBitFieldStore, Type: bytecode.TypeI32}, want: "unsupported opcode OpBitFieldStore"},
		{name: "va start", ins: bytecode.Instr{Op: bytecode.OpVaStart, Slot: 0}, want: "unsupported opcode OpVaStart"},
		{name: "va arg", ins: bytecode.Instr{Op: bytecode.OpVaArg, Type: bytecode.TypeI32}, want: "unsupported opcode OpVaArg"},
		{name: "va end", ins: bytecode.Instr{Op: bytecode.OpVaEnd, Slot: 0}, want: "unsupported opcode OpVaEnd"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := runProgram(t, context.Background(), testMainModule(tt.ins), RunOptions{})
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("Run error = %v, want %q", err, tt.want)
			}
		})
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
