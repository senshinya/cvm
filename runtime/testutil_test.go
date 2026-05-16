package runtime

import "shinya.click/cvm/bytecode"

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
