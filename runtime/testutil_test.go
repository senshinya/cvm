package runtime

import (
	"math"
	"testing"

	"shinya.click/cvm/bytecode"
)

func testMainModule(instrs ...bytecode.Instr) *bytecode.Module {
	mod := bytecode.NewModule()
	mod.Sigs = []bytecode.FuncSig{{ID: 0, Ret: bytecode.TypeI32}}
	mod.Globals = []bytecode.Global{{ID: 0, Name: "main", Kind: bytecode.GlobalFunc, Func: 0, Sig: 0}}
	mod.Functions = []bytecode.Function{{
		ID: 0, GlobalID: 0, Name: "main", Sig: 0,
		Instrs:   instrs,
		MaxStack: 8,
	}}
	mod.Entry = &bytecode.EntryPoint{Global: 0, Name: "main"}
	return mod
}

func mustAlloc(t testing.TB, mem *Memory, name string, size, align int64, readonly bool, kind blockKind) uint64 {
	t.Helper()
	addr, err := mem.TryAlloc(name, size, align, readonly, kind)
	if err != nil {
		t.Fatalf("TryAlloc(%q): %v", name, err)
	}
	return addr
}

func mustAllocBytes(t testing.TB, mem *Memory, name string, data []byte, readonly bool, kind blockKind) uint64 {
	t.Helper()
	addr := mustAlloc(t, mem, name, int64(len(data)), 1, readonly, kind)
	copy(mem.blocks[len(mem.blocks)-1].data, data)
	return addr
}

func mustAllocVaList(t testing.TB, mem *Memory, name string, args ...Value) uint64 {
	t.Helper()
	const (
		testVaListTagI32 = 1
		testVaListTagU32 = 2
		testVaListTagI64 = 3
		testVaListTagU64 = 4
		testVaListTagPtr = 5
		testVaListTagF64 = 6
	)
	addr := mustAlloc(t, mem, name, int64(8+len(args)*16), 8, false, blockLocal)
	if err := mem.Store(addr, bytecode.TypeU64, 8, UIntValue(bytecode.TypeU64, uint64(len(args)))); err != nil {
		t.Fatalf("store va_list count: %v", err)
	}
	for i, arg := range args {
		tag := uint64(0)
		payload := arg.Int
		switch arg.Type {
		case bytecode.TypeI32:
			tag = testVaListTagI32
		case bytecode.TypeU32:
			tag = testVaListTagU32
		case bytecode.TypeI64:
			tag = testVaListTagI64
		case bytecode.TypeU64:
			tag = testVaListTagU64
		case bytecode.TypePtr, bytecode.TypeObjectAddr:
			tag = testVaListTagPtr
		case bytecode.TypeF64:
			tag = testVaListTagF64
			payload = math.Float64bits(arg.Float)
		default:
			t.Fatalf("unsupported va_list test argument type %s", arg.Type)
		}
		entry := addr + uint64(8+i*16)
		if err := mem.Store(entry, bytecode.TypeU64, 8, UIntValue(bytecode.TypeU64, tag)); err != nil {
			t.Fatalf("store va_list tag %d: %v", i, err)
		}
		if err := mem.Store(entry+8, bytecode.TypeU64, 8, UIntValue(bytecode.TypeU64, payload)); err != nil {
			t.Fatalf("store va_list payload %d: %v", i, err)
		}
	}
	return addr
}
