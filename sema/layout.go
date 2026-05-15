package sema

import "fmt"

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

func ValidateProgramInvariants(prog *Program) error {
	if prog == nil {
		return fmt.Errorf("sema program is nil")
	}
	for i, d := range prog.Globals {
		switch x := d.(type) {
		case *VarDecl:
			if x.Sym == nil {
				return fmt.Errorf("global declaration %d has nil symbol", i)
			}
			if x.Sym.GlobalID < 0 {
				return fmt.Errorf("global %q has no global id", x.Sym.Name)
			}
		case *FuncDecl:
			if x.Sym == nil {
				return fmt.Errorf("function declaration %d has nil symbol", i)
			}
			if x.Sym.GlobalID < 0 {
				return fmt.Errorf("function declaration %q has no global id", x.Sym.Name)
			}
		}
	}
	for i, fn := range prog.Funcs {
		if fn == nil {
			return fmt.Errorf("function definition %d is nil", i)
		}
		if fn.Sym == nil {
			return fmt.Errorf("function definition %d has nil symbol", i)
		}
		if fn.T == nil {
			return fmt.Errorf("function definition %q has nil type", fn.Sym.Name)
		}
		if fn.Body == nil {
			return fmt.Errorf("function definition %q has nil body", fn.Sym.Name)
		}
		if fn.Sym.GlobalID < 0 {
			return fmt.Errorf("function definition %q has no global id", fn.Sym.Name)
		}
		for j, p := range fn.Params {
			if p == nil || p.Sym == nil {
				return fmt.Errorf("function %q parameter %d has nil symbol", fn.Sym.Name, j)
			}
			if p.Sym.SlotID < 0 {
				return fmt.Errorf("function %q parameter %q has no slot id", fn.Sym.Name, p.Sym.Name)
			}
		}
		for _, local := range fn.Locals {
			if local == nil || local.Sym == nil {
				continue
			}
			switch local.Storage {
			case StorageStatic:
				if local.Sym.GlobalID < 0 {
					return fmt.Errorf("function %q static local %q has no global id", fn.Sym.Name, local.Sym.Name)
				}
			case StorageExtern:
			default:
				if local.Sym.SlotID < 0 {
					return fmt.Errorf("function %q local %q has no slot id", fn.Sym.Name, local.Sym.Name)
				}
			}
		}
	}
	return nil
}
