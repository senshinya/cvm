package runtime

import (
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
