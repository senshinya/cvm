package runtime

import (
	"fmt"
	"io"
	"math"

	"shinya.click/cvm/bytecode"
)

const funcAddrBase = uint64(0x80000000)

type LoadOptions struct {
	Externs *ExternRegistry
}

type Program struct {
	module     *bytecode.Module
	memory     *Memory
	globalAddr []uint64
	funcAddr   []uint64
	stringAddr []uint64
	externs    map[int]ExternFunc
	entryFunc  int
}

func Load(r io.Reader, opts LoadOptions) (*Program, error) {
	mod, err := bytecode.DecodeModule(r)
	if err != nil {
		return nil, &LoadError{Reason: "decode module", Cause: err}
	}
	if mod.Entry == nil || mod.Entry.Global == bytecode.NoEntryGlobal {
		return nil, &LoadError{Reason: "module has no runnable entry"}
	}

	reg := opts.Externs
	if reg == nil {
		reg = DefaultExternRegistry(nil, nil)
	}

	p := &Program{
		module:     mod,
		memory:     NewMemory(mod.Target),
		globalAddr: make([]uint64, len(mod.Globals)),
		funcAddr:   make([]uint64, len(mod.Globals)),
		stringAddr: make([]uint64, len(mod.Strings)),
		externs:    make(map[int]ExternFunc),
	}
	if err := p.allocateGlobals(reg); err != nil {
		return nil, err
	}
	if err := p.applyRelocations(); err != nil {
		return nil, err
	}
	p.protectReadonlyGlobals()

	entryGlobal := mod.Globals[mod.Entry.Global]
	p.entryFunc = entryGlobal.Func
	if err := p.checkEntrySignature(entryGlobal.Sig); err != nil {
		return nil, err
	}
	return p, nil
}

func (p *Program) Module() *bytecode.Module { return p.module }

func (p *Program) Memory() *Memory { return p.memory }

func (p *Program) GlobalAddr(id int) uint64 { return p.globalAddr[id] }

func (p *Program) allocateGlobals(reg *ExternRegistry) error {
	for i, g := range p.module.Globals {
		switch g.Kind {
		case bytecode.GlobalVar:
			addr, err := p.memory.TryAlloc("global:"+g.Name, g.Size, g.Align, false, blockGlobal)
			if err != nil {
				return &LoadError{Reason: "allocate global " + g.Name, Cause: err}
			}
			p.globalAddr[i] = addr
			if err := p.initializeGlobal(g, addr); err != nil {
				return err
			}
		case bytecode.GlobalFunc:
			p.funcAddr[i] = funcAddrBase + uint64(i)
		case bytecode.GlobalExtern:
			if isExternFunction(g) {
				fn, ok := reg.Lookup(g.Extern.Name)
				if !ok {
					return &LoadError{Reason: "unresolved extern " + g.Extern.Name}
				}
				p.funcAddr[i] = funcAddrBase + uint64(i)
				p.externs[i] = fn
				continue
			}
			addr, ok := reg.LookupVariable(g.Extern.Name, p.memory)
			if !ok {
				return &LoadError{Reason: "unsupported extern variable " + g.Extern.Name}
			}
			p.globalAddr[i] = addr
		default:
			return &LoadError{Reason: fmt.Sprintf("invalid global kind %d", g.Kind)}
		}
	}

	for _, s := range p.module.Strings {
		addr, err := p.memory.TryAlloc(fmt.Sprintf("string:%d", s.ID), int64(len(s.Bytes)), 1, false, blockString)
		if err != nil {
			return &LoadError{Reason: fmt.Sprintf("allocate string %d", s.ID), Cause: err}
		}
		if len(s.Bytes) != 0 {
			b, off, err := p.memory.rangeAccess(addr, int64(len(s.Bytes)), true)
			if err != nil {
				return &LoadError{Reason: fmt.Sprintf("initialize string %d", s.ID), Cause: err}
			}
			copy(b.data[off:], s.Bytes)
		}
		if b, _, err := p.memory.rangeAccess(addr, int64(len(s.Bytes)), false); err == nil {
			b.readonly = true
		}
		p.stringAddr[s.ID] = addr
	}

	return nil
}

func (p *Program) protectReadonlyGlobals() {
	for _, g := range p.module.Globals {
		if g.Kind != bytecode.GlobalVar || !g.Readonly {
			continue
		}
		if b, _, err := p.memory.rangeAccess(p.globalAddr[g.ID], g.Size, false); err == nil {
			b.readonly = true
		}
	}
}

func (p *Program) initializeGlobal(g bytecode.Global, addr uint64) error {
	if len(g.Init.Bytes) == 0 {
		return nil
	}
	b, off, err := p.memory.rangeAccess(addr, int64(len(g.Init.Bytes)), true)
	if err != nil {
		return &LoadError{Reason: "initialize global " + g.Name, Cause: err}
	}
	copy(b.data[off:], g.Init.Bytes)
	return nil
}

func (p *Program) applyRelocations() error {
	for _, g := range p.module.Globals {
		if g.Kind != bytecode.GlobalVar {
			continue
		}
		base := p.globalAddr[g.ID]
		for _, r := range g.Init.Relocations {
			target, err := p.relocationTarget(r)
			if err != nil {
				return &LoadError{Reason: "relocation for global " + g.Name, Cause: err}
			}
			value, err := addRelocationAddend(target, r.Addend)
			if err != nil {
				return &LoadError{Reason: "relocation for global " + g.Name, Cause: err}
			}
			if err := p.memory.WritePointer(base+uint64(r.Offset), value); err != nil {
				return &LoadError{Reason: "write relocation for global " + g.Name, Cause: err}
			}
		}
	}
	return nil
}

func (p *Program) relocationTarget(r bytecode.Relocation) (uint64, error) {
	switch r.Kind {
	case bytecode.RelocGlobal:
		return p.globalAddr[r.Target], nil
	case bytecode.RelocFunc:
		return p.funcAddr[r.Target], nil
	case bytecode.RelocString:
		return p.stringAddr[r.Target], nil
	default:
		return 0, fmt.Errorf("unsupported relocation kind %d", r.Kind)
	}
}

func (p *Program) checkEntrySignature(sigID int) error {
	if sigID < 0 || sigID >= len(p.module.Sigs) {
		return &LoadError{Reason: "entry has invalid signature"}
	}
	sig := p.module.Sigs[sigID]
	if len(sig.Params) != 0 || !isIntegerReturn(sig.Ret) {
		return &LoadError{Reason: "phase 1 supports no-argument integer-returning main"}
	}
	return nil
}

func isExternFunction(g bytecode.Global) bool {
	return g.Kind == bytecode.GlobalExtern && g.Size == 0 && g.Align == 0
}

func isIntegerReturn(t bytecode.ValueType) bool {
	switch t {
	case bytecode.TypeBool,
		bytecode.TypeI8, bytecode.TypeI16, bytecode.TypeI32, bytecode.TypeI64,
		bytecode.TypeU8, bytecode.TypeU16, bytecode.TypeU32, bytecode.TypeU64:
		return true
	default:
		return false
	}
}

func addRelocationAddend(base uint64, addend int64) (uint64, error) {
	if addend >= 0 {
		if uint64(addend) > math.MaxUint64-base {
			return 0, fmt.Errorf("relocation value overflows: %#x + %d", base, addend)
		}
		return base + uint64(addend), nil
	}
	if addend == math.MinInt64 {
		return 0, fmt.Errorf("relocation addend %d is unsupported", addend)
	}
	delta := uint64(-addend)
	if delta > base {
		return 0, fmt.Errorf("relocation value underflows: %#x + %d", base, addend)
	}
	return base - delta, nil
}
