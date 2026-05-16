package sema

func FinalizeProgramLayout(prog *Program) {
	if prog == nil {
		return
	}
	clearProgramLayout(prog)
	nextGlobal := 0
	assignedGlobals := map[*Symbol]bool{}
	assignGlobal := func(sym *Symbol) {
		if sym == nil || assignedGlobals[sym] {
			return
		}
		sym.GlobalID = nextGlobal
		nextGlobal++
		assignedGlobals[sym] = true
	}
	for _, d := range prog.Globals {
		switch x := d.(type) {
		case *VarDecl:
			assignGlobal(x.Sym)
		case *FuncDecl:
			assignGlobal(x.Sym)
		}
	}
	for _, fn := range prog.Funcs {
		assignGlobal(fn.Sym)
		nextSlot := 0
		for _, p := range fn.Params {
			if p == nil || p.Sym == nil {
				continue
			}
			p.Sym.SlotID = nextSlot
			nextSlot++
		}
		for _, local := range fn.Locals {
			if local == nil || local.Sym == nil {
				continue
			}
			if local.Storage == StorageStatic {
				assignGlobal(local.Sym)
				continue
			}
			if local.Storage == StorageExtern {
				continue
			}
			local.Sym.SlotID = nextSlot
			nextSlot++
		}
	}
}

func clearProgramLayout(prog *Program) {
	seen := map[*Symbol]bool{}
	clear := func(sym *Symbol) {
		if sym == nil || seen[sym] {
			return
		}
		sym.SlotID = -1
		sym.GlobalID = -1
		seen[sym] = true
	}
	for _, d := range prog.Globals {
		switch x := d.(type) {
		case *VarDecl:
			clear(x.Sym)
		case *FuncDecl:
			clear(x.Sym)
		}
	}
	for _, fn := range prog.Funcs {
		clear(fn.Sym)
		for _, p := range fn.Params {
			if p != nil {
				clear(p.Sym)
			}
		}
		for _, local := range fn.Locals {
			if local != nil {
				clear(local.Sym)
			}
		}
	}
}
