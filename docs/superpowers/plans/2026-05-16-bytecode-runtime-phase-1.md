# Bytecode Runtime Phase 1 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the first runnable cvm bytecode runtime so `cvm run file.cvmbc` loads a pure binary module, executes no-argument `main`, and returns the integer exit code.

**Architecture:** Add a `runtime` package that consumes only `bytecode.Module`. Split the runtime into loader/program state, memory, VM frames/stack, extern bindings, and CLI integration. Phase 1 implements the executable core and traps cleanly for unsupported opcode families.

**Tech Stack:** Go, existing `bytecode` package, standard library `context`, `encoding/binary`, `errors`, `fmt`, `io`, `os`, and `testing`.

---

## Scope

This plan implements Phase 1 from [docs/superpowers/specs/2026-05-16-runtime-design.md](/Users/shinya/Downloads/cvm/docs/superpowers/specs/2026-05-16-runtime-design.md). It does not implement full libc, native ABI binding, complete floating-point execution, varargs, VLA lifetime, or complete bit-field semantics.

## File Structure

- Create `runtime/errors.go`: load/trap/exit error model and helpers.
- Create `runtime/value.go`: stack value representation and typed conversion helpers.
- Create `runtime/memory.go`: VM address blocks, bounds/alignment/readonly checks, scalar loads/stores, C string reads.
- Create `runtime/extern.go`: extern registry and built-in `exit`, `abort`, `puts`, `fputs`, `stdout`, `stderr`.
- Create `runtime/program.go`: `.cvmbc` load, module validation, static block allocation, relocations, entry signature checks.
- Create `runtime/vm.go`: operand stack, call frames, interpreter loop, calls, returns, and opcode dispatch.
- Create `runtime/testutil_test.go`: helpers for hand-written runtime modules.
- Create `runtime/program_test.go`: loader/static data/relocation tests.
- Create `runtime/memory_test.go`: memory behavior tests.
- Create `runtime/vm_test.go`: VM opcode and exit tests.
- Create `runtime/integration_test.go`: C source to bytecode to runtime execution tests.
- Modify `main.go`: add `cvm run file.cvmbc` command while preserving current compiler flags.
- Modify `compiler_test.go`: add CLI run tests.

## Task 1: Runtime Error Types

**Files:**
- Create: `runtime/errors.go`
- Test: `runtime/errors_test.go`

- [ ] **Step 1: Write failing tests for load errors, traps, and exit status**

Create `runtime/errors_test.go`:

```go
package runtime

import (
	"errors"
	"strings"
	"testing"

	"shinya.click/cvm/bytecode"
)

func TestLoadErrorWrapsCause(t *testing.T) {
	err := &LoadError{Reason: "missing entry", Cause: errors.New("entry=none")}
	if !strings.Contains(err.Error(), "missing entry") || !strings.Contains(err.Error(), "entry=none") {
		t.Fatalf("LoadError text = %q", err.Error())
	}
	if !errors.Is(err, err.Cause) {
		t.Fatalf("errors.Is did not unwrap cause")
	}
}

func TestTrapErrorIncludesExecutionContext(t *testing.T) {
	err := &TrapError{
		Reason: "division by zero",
		Function: "main",
		FunctionID: 0,
		PC: 3,
		Opcode: bytecode.OpBinary,
	}
	got := err.Error()
	for _, want := range []string{"division by zero", "main", "pc=3", "OpBinary"} {
		if !strings.Contains(got, want) {
			t.Fatalf("TrapError missing %q: %s", want, got)
		}
	}
}

func TestExitStatusIsNormalResult(t *testing.T) {
	st := ExitStatus{Code: 7}
	if st.Code != 7 {
		t.Fatalf("exit code = %d, want 7", st.Code)
	}
}
```

- [ ] **Step 2: Run tests and verify they fail**

Run: `go test ./runtime -run 'TestLoadError|TestTrapError|TestExitStatus' -count=1`

Expected: FAIL because package `runtime` or the new types do not exist.

- [ ] **Step 3: Implement error types**

Create `runtime/errors.go`:

```go
package runtime

import (
	"fmt"

	"shinya.click/cvm/bytecode"
)

type LoadError struct {
	Reason string
	Cause  error
}

func (e *LoadError) Error() string {
	if e == nil {
		return "<nil>"
	}
	if e.Cause != nil {
		return fmt.Sprintf("runtime load error: %s: %v", e.Reason, e.Cause)
	}
	return fmt.Sprintf("runtime load error: %s", e.Reason)
}

func (e *LoadError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

type TrapError struct {
	Reason     string
	Function   string
	FunctionID int
	PC         int
	Opcode     bytecode.Opcode
	Stack      []string
	Cause      error
}

func (e *TrapError) Error() string {
	if e == nil {
		return "<nil>"
	}
	loc := ""
	if e.Function != "" {
		loc = fmt.Sprintf(" in %s#%d pc=%d opcode=%s", e.Function, e.FunctionID, e.PC, e.Opcode)
	}
	if e.Cause != nil {
		return fmt.Sprintf("runtime trap: %s%s: %v", e.Reason, loc, e.Cause)
	}
	return fmt.Sprintf("runtime trap: %s%s", e.Reason, loc)
}

func (e *TrapError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

type ExitStatus struct {
	Code int
}
```

- [ ] **Step 4: Run tests and verify they pass**

Run: `go test ./runtime -run 'TestLoadError|TestTrapError|TestExitStatus' -count=1`

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add runtime/errors.go runtime/errors_test.go
git commit -m "feat(runtime): add error model"
```

## Task 2: Runtime Value Representation

**Files:**
- Create: `runtime/value.go`
- Test: `runtime/value_test.go`

- [ ] **Step 1: Write failing value conversion tests**

Create `runtime/value_test.go`:

```go
package runtime

import (
	"testing"

	"shinya.click/cvm/bytecode"
)

func TestValueAsExitCode(t *testing.T) {
	tests := []struct {
		name string
		v    Value
		want int
	}{
		{"i32", IntValue(bytecode.TypeI32, 42), 42},
		{"u8", IntValue(bytecode.TypeU8, 255), 255},
		{"bool", IntValue(bytecode.TypeBool, 1), 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.v.ExitCode()
			if err != nil {
				t.Fatalf("ExitCode returned error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("ExitCode = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestValueRejectsPointerExitCode(t *testing.T) {
	_, err := PtrValue(0x1000).ExitCode()
	if err == nil {
		t.Fatal("ExitCode accepted pointer")
	}
}
```

- [ ] **Step 2: Run tests and verify they fail**

Run: `go test ./runtime -run TestValue -count=1`

Expected: FAIL because `Value`, `IntValue`, and `PtrValue` do not exist.

- [ ] **Step 3: Implement `Value` helpers**

Create `runtime/value.go`:

```go
package runtime

import (
	"fmt"
	"math"

	"shinya.click/cvm/bytecode"
)

type Value struct {
	Type  bytecode.ValueType
	Int   uint64
	Float float64
}

func IntValue(t bytecode.ValueType, v int64) Value {
	return Value{Type: t, Int: uint64(v)}
}

func UIntValue(t bytecode.ValueType, v uint64) Value {
	return Value{Type: t, Int: v}
}

func PtrValue(addr uint64) Value {
	return Value{Type: bytecode.TypePtr, Int: addr}
}

func ObjectAddrValue(addr uint64) Value {
	return Value{Type: bytecode.TypeObjectAddr, Int: addr}
}

func FloatValue(t bytecode.ValueType, v float64) Value {
	return Value{Type: t, Float: v}
}

func (v Value) ExitCode() (int, error) {
	switch v.Type {
	case bytecode.TypeBool, bytecode.TypeI8, bytecode.TypeI16, bytecode.TypeI32,
		bytecode.TypeI64, bytecode.TypeU8, bytecode.TypeU16, bytecode.TypeU32,
		bytecode.TypeU64:
		if v.Int > uint64(math.MaxInt) {
			return 0, fmt.Errorf("exit code %d exceeds int range", v.Int)
		}
		return int(v.Int), nil
	default:
		return 0, fmt.Errorf("cannot use %s as exit code", v.Type)
	}
}

func (v Value) IsZero() bool {
	switch v.Type {
	case bytecode.TypeF32, bytecode.TypeF64, bytecode.TypeFLong:
		return v.Float == 0
	default:
		return v.Int == 0
	}
}
```

- [ ] **Step 4: Run tests and verify they pass**

Run: `go test ./runtime -run TestValue -count=1`

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add runtime/value.go runtime/value_test.go
git commit -m "feat(runtime): add VM values"
```

## Task 3: Memory Blocks

**Files:**
- Create: `runtime/memory.go`
- Test: `runtime/memory_test.go`

- [ ] **Step 1: Write failing memory tests**

Create `runtime/memory_test.go`:

```go
package runtime

import (
	"strings"
	"testing"

	"shinya.click/cvm/bytecode"
)

func TestMemoryReadWriteScalar(t *testing.T) {
	mem := NewMemory(bytecode.DefaultTarget())
	addr := mem.Alloc("global:g", 8, 4, false, blockGlobal)
	if err := mem.Store(addr, bytecode.TypeI32, 4, IntValue(bytecode.TypeI32, 0x11223344)); err != nil {
		t.Fatalf("Store: %v", err)
	}
	got, err := mem.Load(addr, bytecode.TypeI32, 4)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got.Int != 0x11223344 {
		t.Fatalf("loaded %#x, want 0x11223344", got.Int)
	}
}

func TestMemoryRejectsReadonlyWrite(t *testing.T) {
	mem := NewMemory(bytecode.DefaultTarget())
	addr := mem.Alloc("string:0", 4, 1, true, blockString)
	err := mem.Store(addr, bytecode.TypeI8, 1, IntValue(bytecode.TypeI8, 1))
	if err == nil || !strings.Contains(err.Error(), "readonly") {
		t.Fatalf("Store error = %v, want readonly", err)
	}
}

func TestMemoryReadCString(t *testing.T) {
	mem := NewMemory(bytecode.DefaultTarget())
	addr := mem.AllocBytes("string:0", []byte("hello\x00extra"), true, blockString)
	got, err := mem.ReadCString(addr)
	if err != nil {
		t.Fatalf("ReadCString: %v", err)
	}
	if got != "hello" {
		t.Fatalf("ReadCString = %q, want hello", got)
	}
}
```

- [ ] **Step 2: Run tests and verify they fail**

Run: `go test ./runtime -run TestMemory -count=1`

Expected: FAIL because `NewMemory` and block helpers do not exist.

- [ ] **Step 3: Implement memory blocks**

Create `runtime/memory.go`:

```go
package runtime

import (
	"encoding/binary"
	"fmt"

	"shinya.click/cvm/bytecode"
)

type blockKind int

const (
	blockGlobal blockKind = iota
	blockString
	blockLocal
	blockDynamic
	blockHostHandle
)

type memoryBlock struct {
	id       int
	name     string
	base     uint64
	data     []byte
	align    int64
	readonly bool
	freed    bool
	kind     blockKind
}

type Memory struct {
	target bytecode.TargetInfo
	next   uint64
	blocks []*memoryBlock
}

func NewMemory(target bytecode.TargetInfo) *Memory {
	return &Memory{target: target, next: 0x1000}
}

func (m *Memory) Alloc(name string, size, align int64, readonly bool, kind blockKind) uint64 {
	if align <= 0 {
		align = 1
	}
	base := alignUp(m.next, uint64(align))
	b := &memoryBlock{id: len(m.blocks), name: name, base: base, data: make([]byte, size), align: align, readonly: readonly, kind: kind}
	m.blocks = append(m.blocks, b)
	m.next = base + uint64(size) + 0x10
	return base
}

func (m *Memory) AllocBytes(name string, data []byte, readonly bool, kind blockKind) uint64 {
	addr := m.Alloc(name, int64(len(data)), 1, readonly, kind)
	copy(m.blocks[len(m.blocks)-1].data, data)
	return addr
}

func (m *Memory) Load(addr uint64, t bytecode.ValueType, align int64) (Value, error) {
	b, off, size, err := m.access(addr, t, align, false)
	if err != nil {
		return Value{}, err
	}
	raw := b.data[off : off+size]
	switch t {
	case bytecode.TypeBool, bytecode.TypeI8, bytecode.TypeU8:
		return UIntValue(t, uint64(raw[0])), nil
	case bytecode.TypeI16, bytecode.TypeU16:
		return UIntValue(t, uint64(m.order().Uint16(raw))), nil
	case bytecode.TypeI32, bytecode.TypeU32:
		return UIntValue(t, uint64(m.order().Uint32(raw))), nil
	case bytecode.TypeI64, bytecode.TypeU64, bytecode.TypePtr, bytecode.TypeObjectAddr:
		return UIntValue(t, m.order().Uint64(raw)), nil
	default:
		return Value{}, fmt.Errorf("unsupported load type %s", t)
	}
}

func (m *Memory) Store(addr uint64, t bytecode.ValueType, align int64, v Value) error {
	b, off, size, err := m.access(addr, t, align, true)
	if err != nil {
		return err
	}
	raw := b.data[off : off+size]
	switch t {
	case bytecode.TypeBool, bytecode.TypeI8, bytecode.TypeU8:
		raw[0] = byte(v.Int)
	case bytecode.TypeI16, bytecode.TypeU16:
		m.order().PutUint16(raw, uint16(v.Int))
	case bytecode.TypeI32, bytecode.TypeU32:
		m.order().PutUint32(raw, uint32(v.Int))
	case bytecode.TypeI64, bytecode.TypeU64, bytecode.TypePtr, bytecode.TypeObjectAddr:
		m.order().PutUint64(raw, v.Int)
	default:
		return fmt.Errorf("unsupported store type %s", t)
	}
	return nil
}

func (m *Memory) Copy(dst, src uint64, size int64) error {
	db, doff, err := m.rangeAccess(dst, size, true)
	if err != nil {
		return err
	}
	sb, soff, err := m.rangeAccess(src, size, false)
	if err != nil {
		return err
	}
	copy(db.data[doff:doff+int(size)], sb.data[soff:soff+int(size)])
	return nil
}

func (m *Memory) Set(dst uint64, value byte, size int64) error {
	b, off, err := m.rangeAccess(dst, size, true)
	if err != nil {
		return err
	}
	for i := 0; i < int(size); i++ {
		b.data[off+i] = value
	}
	return nil
}

func (m *Memory) ReadCString(addr uint64) (string, error) {
	b, off, err := m.rangeAccess(addr, 1, false)
	if err != nil {
		return "", err
	}
	end := off
	for end < len(b.data) && b.data[end] != 0 {
		end++
	}
	if end >= len(b.data) {
		return "", fmt.Errorf("unterminated C string at %#x", addr)
	}
	return string(b.data[off:end]), nil
}

func (m *Memory) WritePointer(addr uint64, ptr uint64) error {
	return m.Store(addr, bytecode.TypePtr, m.target.PointerAlign, PtrValue(ptr))
}

func (m *Memory) access(addr uint64, t bytecode.ValueType, align int64, write bool) (*memoryBlock, int, int, error) {
	size := int(valueSize(m.target, t))
	if size <= 0 {
		return nil, 0, 0, fmt.Errorf("unsupported memory type %s", t)
	}
	b, off, err := m.rangeAccess(addr, int64(size), write)
	if err != nil {
		return nil, 0, 0, err
	}
	if align > 1 && addr%uint64(align) != 0 {
		return nil, 0, 0, fmt.Errorf("unaligned access at %#x align=%d", addr, align)
	}
	return b, off, size, nil
}

func (m *Memory) rangeAccess(addr uint64, size int64, write bool) (*memoryBlock, int, error) {
	if size < 0 {
		return nil, 0, fmt.Errorf("negative memory size %d", size)
	}
	for _, b := range m.blocks {
		if addr < b.base || addr+uint64(size) > b.base+uint64(len(b.data)) {
			continue
		}
		if b.freed {
			return nil, 0, fmt.Errorf("use after free at %#x", addr)
		}
		if write && b.readonly {
			return nil, 0, fmt.Errorf("readonly memory write at %#x", addr)
		}
		return b, int(addr - b.base), nil
	}
	return nil, 0, fmt.Errorf("invalid memory access at %#x size=%d", addr, size)
}

func (m *Memory) order() binary.ByteOrder {
	if m.target.Endian == "little" || m.target.Endian == "" {
		return binary.LittleEndian
	}
	return binary.BigEndian
}

func valueSize(target bytecode.TargetInfo, t bytecode.ValueType) int64 {
	switch t {
	case bytecode.TypeBool:
		return target.BoolSize
	case bytecode.TypeI8, bytecode.TypeU8:
		return 1
	case bytecode.TypeI16, bytecode.TypeU16:
		return 2
	case bytecode.TypeI32, bytecode.TypeU32, bytecode.TypeF32:
		return 4
	case bytecode.TypeI64, bytecode.TypeU64, bytecode.TypeF64, bytecode.TypePtr, bytecode.TypeObjectAddr:
		return 8
	default:
		return 0
	}
}

func alignUp(v, align uint64) uint64 {
	if align <= 1 {
		return v
	}
	rem := v % align
	if rem == 0 {
		return v
	}
	return v + align - rem
}
```

- [ ] **Step 4: Run tests and verify they pass**

Run: `go test ./runtime -run TestMemory -count=1`

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add runtime/memory.go runtime/memory_test.go
git commit -m "feat(runtime): add checked memory blocks"
```

## Task 4: Program Loader And Static Relocations

**Files:**
- Create: `runtime/program.go`
- Create: `runtime/testutil_test.go`
- Test: `runtime/program_test.go`

- [ ] **Step 1: Write failing loader tests**

Create `runtime/testutil_test.go`:

```go
package runtime

import "shinya.click/cvm/bytecode"

func testMainModule(instrs ...bytecode.Instr) *bytecode.Module {
	mod := bytecode.NewModule()
	mod.Sigs = []bytecode.FuncSig{{ID: 0, Ret: bytecode.TypeI32}}
	mod.Globals = []bytecode.Global{{ID: 0, Name: "main", Kind: bytecode.GlobalFunc, Func: 0, Sig: 0}}
	mod.Functions = []bytecode.Function{{
		ID: 0, GlobalID: 0, Name: "main", Sig: 0,
		Instrs: instrs,
		MaxStack: 8,
	}}
	mod.Entry = &bytecode.EntryPoint{Global: 0, Name: "main"}
	return mod
}
```

Create `runtime/program_test.go`:

```go
package runtime

import (
	"bytes"
	"strings"
	"testing"

	"shinya.click/cvm/bytecode"
)

func TestLoadRejectsNoEntryModule(t *testing.T) {
	mod := bytecode.NewModule()
	var buf bytes.Buffer
	if err := bytecode.EncodeModule(&buf, mod); err != nil {
		t.Fatalf("EncodeModule: %v", err)
	}
	_, err := Load(bytes.NewReader(buf.Bytes()), LoadOptions{})
	if err == nil || !strings.Contains(err.Error(), "entry") {
		t.Fatalf("Load error = %v, want entry failure", err)
	}
}

func TestLoadAppliesGlobalRelocation(t *testing.T) {
	mod := testMainModule(bytecode.I32Const(0), bytecode.Return(bytecode.TypeI32))
	mod.Globals = append(mod.Globals,
		bytecode.Global{ID: 1, Name: "target", Kind: bytecode.GlobalVar, Size: 4, Align: 4, Init: bytecode.InitData{Bytes: []byte{1, 2, 3, 4}}},
		bytecode.Global{ID: 2, Name: "ptr", Kind: bytecode.GlobalVar, Size: 8, Align: 8, Init: bytecode.InitData{Relocations: []bytecode.Relocation{{Offset: 0, Kind: bytecode.RelocGlobal, Target: 1}}}},
	)
	var buf bytes.Buffer
	if err := bytecode.EncodeModule(&buf, mod); err != nil {
		t.Fatalf("EncodeModule: %v", err)
	}
	p, err := Load(bytes.NewReader(buf.Bytes()), LoadOptions{})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	ptrAddr := p.GlobalAddr(2)
	got, err := p.Memory().Load(ptrAddr, bytecode.TypePtr, 8)
	if err != nil {
		t.Fatalf("Load ptr: %v", err)
	}
	if got.Int != p.GlobalAddr(1) {
		t.Fatalf("relocated pointer = %#x, want target %#x", got.Int, p.GlobalAddr(1))
	}
}
```

- [ ] **Step 2: Run tests and verify they fail**

Run: `go test ./runtime -run 'TestLoad' -count=1`

Expected: FAIL because `Load`, `LoadOptions`, `Program`, `GlobalAddr`, and `Memory` do not exist.

- [ ] **Step 3: Implement loader and `Program`**

Create `runtime/program.go`:

```go
package runtime

import (
	"fmt"
	"io"

	"shinya.click/cvm/bytecode"
)

type LoadOptions struct {
	Externs *ExternRegistry
}

type Program struct {
	module    *bytecode.Module
	memory    *Memory
	globalAddr []uint64
	funcAddr   []uint64
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
		module: mod,
		memory: NewMemory(mod.Target),
		globalAddr: make([]uint64, len(mod.Globals)),
		funcAddr: make([]uint64, len(mod.Globals)),
		externs: make(map[int]ExternFunc),
	}
	if err := p.allocateGlobals(reg); err != nil {
		return nil, err
	}
	if err := p.applyRelocations(); err != nil {
		return nil, err
	}
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
			p.globalAddr[i] = p.memory.Alloc("global:"+g.Name, g.Size, g.Align, g.Readonly, blockGlobal)
			if len(g.Init.Bytes) > 0 {
				if err := p.memory.Set(p.globalAddr[i], 0, g.Size); err != nil {
					return &LoadError{Reason: "initialize global " + g.Name, Cause: err}
				}
				b, off, err := p.memory.rangeAccess(p.globalAddr[i], int64(len(g.Init.Bytes)), true)
				if err != nil {
					return &LoadError{Reason: "initialize global " + g.Name, Cause: err}
				}
				copy(b.data[off:], g.Init.Bytes)
			}
		case bytecode.GlobalFunc:
			p.funcAddr[i] = uint64(0x80000000 + i)
		case bytecode.GlobalExtern:
			if isExternFunction(g) {
				fn, ok := reg.Lookup(g.Extern.Name)
				if !ok {
					return &LoadError{Reason: "unresolved extern " + g.Extern.Name}
				}
				p.funcAddr[i] = uint64(0x80000000 + i)
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
		addr := p.memory.AllocBytes(fmt.Sprintf("string:%d", s.ID), s.Bytes, true, blockString)
		if s.ID >= len(p.globalAddr) {
			_ = addr
		}
	}
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
			if err := p.memory.WritePointer(base+uint64(r.Offset), target+uint64(r.Addend)); err != nil {
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
	default:
		return 0, fmt.Errorf("unsupported relocation kind %d in phase 1", r.Kind)
	}
}

func (p *Program) checkEntrySignature(sigID int) error {
	if sigID < 0 || sigID >= len(p.module.Sigs) {
		return &LoadError{Reason: "entry has invalid signature"}
	}
	sig := p.module.Sigs[sigID]
	if len(sig.Params) != 0 || sig.Ret == bytecode.TypeVoid {
		return &LoadError{Reason: "phase 1 supports no-argument integer-returning main"}
	}
	return nil
}

func isExternFunction(g bytecode.Global) bool {
	return g.Kind == bytecode.GlobalExtern && g.Size == 0 && g.Align == 0
}
```

This step references `ExternRegistry` and `ExternFunc`; Task 5 creates them. If the compiler fails here, continue directly to Task 5 before expecting package-wide success.

- [ ] **Step 4: Run tests after Task 5 lands**

Run after Task 5 implementation exists: `go test ./runtime -run 'TestLoad' -count=1`

Expected: PASS.

- [ ] **Step 5: Commit after Task 5 tests pass**

```bash
git add runtime/program.go runtime/testutil_test.go runtime/program_test.go
git commit -m "feat(runtime): load bytecode programs"
```

## Task 5: Extern Registry

**Files:**
- Create: `runtime/extern.go`
- Test: `runtime/extern_test.go`

- [ ] **Step 1: Write failing extern tests**

Create `runtime/extern_test.go`:

```go
package runtime

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"shinya.click/cvm/bytecode"
)

func TestDefaultExternRegistryHasExitAndAbort(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	for _, name := range []string{"exit", "abort", "puts", "fputs"} {
		if _, ok := reg.Lookup(name); !ok {
			t.Fatalf("missing extern %s", name)
		}
	}
}

func TestPutsWritesCString(t *testing.T) {
	var out bytes.Buffer
	reg := DefaultExternRegistry(&out, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	addr := mem.AllocBytes("string:0", []byte("hello\x00"), true, blockString)
	fn, _ := reg.Lookup("puts")
	ret, exit, err := fn(context.Background(), &ExternContext{Memory: mem}, []Value{ObjectAddrValue(addr)})
	if err != nil || exit != nil {
		t.Fatalf("puts ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if out.String() != "hello\n" {
		t.Fatalf("puts output = %q", out.String())
	}
}

func TestAbortReturnsTrap(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	fn, _ := reg.Lookup("abort")
	_, _, err := fn(context.Background(), &ExternContext{Memory: NewMemory(bytecode.DefaultTarget())}, nil)
	if err == nil || !strings.Contains(err.Error(), "abort") {
		t.Fatalf("abort err = %v, want abort trap", err)
	}
}
```

- [ ] **Step 2: Run tests and verify they fail**

Run: `go test ./runtime -run 'TestDefaultExtern|TestPuts|TestAbort' -count=1`

Expected: FAIL because `ExternRegistry` does not exist.

- [ ] **Step 3: Implement extern registry**

Create `runtime/extern.go`:

```go
package runtime

import (
	"context"
	"fmt"
	"io"
	"os"

	"shinya.click/cvm/bytecode"
)

type ExternFunc func(context.Context, *ExternContext, []Value) (Value, *ExitStatus, error)

type ExternContext struct {
	Memory *Memory
	Stdout io.Writer
	Stderr io.Writer
}

type ExternRegistry struct {
	funcs map[string]ExternFunc
	stdout io.Writer
	stderr io.Writer
}

func NewExternRegistry(stdout, stderr io.Writer) *ExternRegistry {
	if stdout == nil {
		stdout = os.Stdout
	}
	if stderr == nil {
		stderr = os.Stderr
	}
	return &ExternRegistry{funcs: make(map[string]ExternFunc), stdout: stdout, stderr: stderr}
}

func DefaultExternRegistry(stdout, stderr io.Writer) *ExternRegistry {
	r := NewExternRegistry(stdout, stderr)
	r.Register("exit", func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 1 {
			return Value{}, nil, fmt.Errorf("exit expects 1 argument")
		}
		code, err := args[0].ExitCode()
		if err != nil {
			return Value{}, nil, err
		}
		return Value{}, &ExitStatus{Code: code}, nil
	})
	r.Register("abort", func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		return Value{}, nil, &TrapError{Reason: "abort"}
	})
	r.Register("puts", func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 1 {
			return Value{}, nil, fmt.Errorf("puts expects 1 argument")
		}
		s, err := ec.Memory.ReadCString(args[0].Int)
		if err != nil {
			return Value{}, nil, err
		}
		fmt.Fprintln(ec.Stdout, s)
		return IntValue(bytecode.TypeI32, int64(len(s)+1)), nil, nil
	})
	r.Register("fputs", func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 2 {
			return Value{}, nil, fmt.Errorf("fputs expects 2 arguments")
		}
		s, err := ec.Memory.ReadCString(args[0].Int)
		if err != nil {
			return Value{}, nil, err
		}
		fmt.Fprint(ec.Stdout, s)
		return IntValue(bytecode.TypeI32, len(s)), nil, nil
	})
	return r
}

func (r *ExternRegistry) Register(name string, fn ExternFunc) {
	r.funcs[name] = fn
}

func (r *ExternRegistry) Lookup(name string) (ExternFunc, bool) {
	fn, ok := r.funcs[name]
	return fn, ok
}

func (r *ExternRegistry) LookupVariable(name string, mem *Memory) (uint64, bool) {
	switch name {
	case "stdout", "stderr":
		return mem.Alloc("extern:"+name, 8, 8, true, blockHostHandle), true
	default:
		return 0, false
	}
}

func (r *ExternRegistry) context(mem *Memory) *ExternContext {
	return &ExternContext{Memory: mem, Stdout: r.stdout, Stderr: r.stderr}
}
```

- [ ] **Step 4: Run extern and loader tests**

Run: `go test ./runtime -run 'TestDefaultExtern|TestPuts|TestAbort|TestLoad' -count=1`

Expected: PASS. If `program.go` still cannot find `ExternRegistry`, ensure Task 4 and Task 5 files are both present before rerunning.

- [ ] **Step 5: Commit**

```bash
git add runtime/extern.go runtime/extern_test.go
git commit -m "feat(runtime): add built-in extern registry"
```

## Task 6: VM Stack, Frames, Constants, Locals, And Return

**Files:**
- Create: `runtime/vm.go`
- Test: `runtime/vm_test.go`

- [ ] **Step 1: Write failing tests for basic execution**

Create `runtime/vm_test.go`:

```go
package runtime

import (
	"bytes"
	"context"
	"testing"

	"shinya.click/cvm/bytecode"
)

func runModule(t *testing.T, mod *bytecode.Module) (ExitStatus, error) {
	t.Helper()
	var buf bytes.Buffer
	if err := bytecode.EncodeModule(&buf, mod); err != nil {
		t.Fatalf("EncodeModule: %v", err)
	}
	p, err := Load(bytes.NewReader(buf.Bytes()), LoadOptions{Externs: DefaultExternRegistry(nil, nil)})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	return Run(context.Background(), p, RunOptions{})
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
```

- [ ] **Step 2: Run tests and verify they fail**

Run: `go test ./runtime -run 'TestRunReturnsMainConstant|TestRunLocalStoreLoad' -count=1`

Expected: FAIL because `Run`, `RunOptions`, and the VM do not exist.

- [ ] **Step 3: Implement VM core**

Create `runtime/vm.go`:

```go
package runtime

import (
	"context"
	"fmt"

	"shinya.click/cvm/bytecode"
)

type RunOptions struct {
	StepLimit int
}

type VM struct {
	program *Program
	stack   []Value
	frames  []frame
	steps   int
	limit   int
}

type frame struct {
	fn     *bytecode.Function
	pc     int
	locals []Value
	labels map[int]int
}

func Run(ctx context.Context, p *Program, opts RunOptions) (ExitStatus, error) {
	vm := &VM{program: p, limit: opts.StepLimit}
	if vm.limit == 0 {
		vm.limit = 1000000
	}
	if err := vm.pushFrame(p.entryFunc, nil); err != nil {
		return ExitStatus{}, err
	}
	for {
		st, done, err := vm.step(ctx)
		if err != nil || done {
			return st, err
		}
	}
}

func (vm *VM) pushFrame(funcID int, args []Value) error {
	if funcID < 0 || funcID >= len(vm.program.module.Functions) {
		return &TrapError{Reason: fmt.Sprintf("invalid function %d", funcID)}
	}
	fn := &vm.program.module.Functions[funcID]
	locals := make([]Value, len(fn.Locals))
	for i, p := range fn.Params {
		if p.Slot >= 0 && p.Slot < len(locals) && i < len(args) {
			locals[p.Slot] = args[i]
		}
	}
	labels := make(map[int]int)
	for pc, ins := range fn.Instrs {
		if ins.Op == bytecode.OpLabel {
			labels[ins.Label] = pc
		}
	}
	vm.frames = append(vm.frames, frame{fn: fn, locals: locals, labels: labels})
	return nil
}

func (vm *VM) step(ctx context.Context) (ExitStatus, bool, error) {
	select {
	case <-ctx.Done():
		return ExitStatus{}, false, ctx.Err()
	default:
	}
	if vm.steps >= vm.limit {
		return ExitStatus{}, false, vm.trap("step limit exceeded", bytecode.Instr{})
	}
	vm.steps++
	fr := &vm.frames[len(vm.frames)-1]
	if fr.pc < 0 || fr.pc >= len(fr.fn.Instrs) {
		return ExitStatus{}, false, vm.trap("program counter outside function", bytecode.Instr{})
	}
	ins := fr.fn.Instrs[fr.pc]
	fr.pc++
	switch ins.Op {
	case bytecode.OpConst:
		vm.stack = append(vm.stack, constValue(ins))
	case bytecode.OpLoadLocal:
		if ins.Slot < 0 || ins.Slot >= len(fr.locals) {
			return ExitStatus{}, false, vm.trap("invalid local slot", ins)
		}
		vm.stack = append(vm.stack, fr.locals[ins.Slot])
	case bytecode.OpStoreLocal:
		v, err := vm.pop(ins.Type)
		if err != nil {
			return ExitStatus{}, false, vm.trapWithCause("store local", ins, err)
		}
		if ins.Slot < 0 || ins.Slot >= len(fr.locals) {
			return ExitStatus{}, false, vm.trap("invalid local slot", ins)
		}
		fr.locals[ins.Slot] = v
	case bytecode.OpReturn:
		v, err := vm.pop(ins.Type)
		if err != nil {
			return ExitStatus{}, false, vm.trapWithCause("return", ins, err)
		}
		vm.frames = vm.frames[:len(vm.frames)-1]
		if len(vm.frames) == 0 {
			code, err := v.ExitCode()
			if err != nil {
				return ExitStatus{}, false, vm.trapWithCause("return exit code", ins, err)
			}
			return ExitStatus{Code: code}, true, nil
		}
		vm.stack = append(vm.stack, v)
	case bytecode.OpReturnVoid:
		vm.frames = vm.frames[:len(vm.frames)-1]
		if len(vm.frames) == 0 {
			return ExitStatus{Code: 0}, true, nil
		}
	case bytecode.OpLabel:
	default:
		return ExitStatus{}, false, vm.trap("unsupported opcode", ins)
	}
	return ExitStatus{}, false, nil
}

func constValue(ins bytecode.Instr) Value {
	switch ins.Type {
	case bytecode.TypePtr:
		return PtrValue(uint64(ins.Int))
	case bytecode.TypeObjectAddr:
		return ObjectAddrValue(uint64(ins.Int))
	case bytecode.TypeF32, bytecode.TypeF64, bytecode.TypeFLong:
		return FloatValue(ins.Type, ins.Float)
	default:
		return IntValue(ins.Type, ins.Int)
	}
}

func (vm *VM) pop(t bytecode.ValueType) (Value, error) {
	if len(vm.stack) == 0 {
		return Value{}, fmt.Errorf("stack underflow")
	}
	v := vm.stack[len(vm.stack)-1]
	vm.stack = vm.stack[:len(vm.stack)-1]
	if t != bytecode.TypeVoid && v.Type != t {
		return Value{}, fmt.Errorf("stack type %s, want %s", v.Type, t)
	}
	return v, nil
}

func (vm *VM) trap(reason string, ins bytecode.Instr) error {
	return vm.trapWithCause(reason, ins, nil)
}

func (vm *VM) trapWithCause(reason string, ins bytecode.Instr, cause error) error {
	fr := vm.frames[len(vm.frames)-1]
	return &TrapError{Reason: reason, Function: fr.fn.Name, FunctionID: fr.fn.ID, PC: fr.pc - 1, Opcode: ins.Op, Cause: cause}
}
```

- [ ] **Step 4: Run tests and verify they pass**

Run: `go test ./runtime -run 'TestRunReturnsMainConstant|TestRunLocalStoreLoad' -count=1`

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add runtime/vm.go runtime/vm_test.go
git commit -m "feat(runtime): execute constants locals and returns"
```

## Task 7: Stack Operations, Integer Arithmetic, Casts, And Branches

**Files:**
- Modify: `runtime/vm.go`
- Modify: `runtime/vm_test.go`

- [ ] **Step 1: Add failing VM tests**

Append to `runtime/vm_test.go`:

```go
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
	st, err := runModule(t, mod)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if st.Code != 22 {
		t.Fatalf("exit code = %d, want 22", st.Code)
	}
}
```

- [ ] **Step 2: Run tests and verify they fail**

Run: `go test ./runtime -run 'TestRunIntegerArithmeticAndBranch|TestRunSwitch' -count=1`

Expected: FAIL with unsupported opcode traps.

- [ ] **Step 3: Implement stack, arithmetic, casts, and branches**

Extend the `switch ins.Op` in `runtime/vm.go` with these cases:

```go
	case bytecode.OpDup:
		if len(vm.stack) == 0 {
			return ExitStatus{}, false, vm.trap("dup stack underflow", ins)
		}
		vm.stack = append(vm.stack, vm.stack[len(vm.stack)-1])
	case bytecode.OpPop:
		if _, err := vm.pop(bytecode.TypeVoid); err != nil {
			return ExitStatus{}, false, vm.trapWithCause("pop", ins, err)
		}
	case bytecode.OpSwap:
		if len(vm.stack) < 2 {
			return ExitStatus{}, false, vm.trap("swap stack underflow", ins)
		}
		vm.stack[len(vm.stack)-1], vm.stack[len(vm.stack)-2] = vm.stack[len(vm.stack)-2], vm.stack[len(vm.stack)-1]
	case bytecode.OpBinary:
		if err := vm.binary(ins); err != nil {
			return ExitStatus{}, false, vm.trapWithCause("binary", ins, err)
		}
	case bytecode.OpUnary:
		if err := vm.unary(ins); err != nil {
			return ExitStatus{}, false, vm.trapWithCause("unary", ins, err)
		}
	case bytecode.OpCast:
		v, err := vm.pop(ins.Type)
		if err != nil {
			return ExitStatus{}, false, vm.trapWithCause("cast", ins, err)
		}
		vm.stack = append(vm.stack, Value{Type: ins.Type2, Int: v.Int, Float: v.Float})
	case bytecode.OpJump:
		if err := fr.jump(ins.Label); err != nil {
			return ExitStatus{}, false, vm.trapWithCause("jump", ins, err)
		}
	case bytecode.OpJumpIfZero:
		v, err := vm.pop(ins.Type)
		if err != nil {
			return ExitStatus{}, false, vm.trapWithCause("jump if zero", ins, err)
		}
		if v.IsZero() {
			if err := fr.jump(ins.Label); err != nil {
				return ExitStatus{}, false, vm.trapWithCause("jump if zero", ins, err)
			}
		}
	case bytecode.OpJumpIfNonZero:
		v, err := vm.pop(ins.Type)
		if err != nil {
			return ExitStatus{}, false, vm.trapWithCause("jump if nonzero", ins, err)
		}
		if !v.IsZero() {
			if err := fr.jump(ins.Label); err != nil {
				return ExitStatus{}, false, vm.trapWithCause("jump if nonzero", ins, err)
			}
		}
	case bytecode.OpSwitch:
		v, err := vm.pop(ins.Type)
		if err != nil {
			return ExitStatus{}, false, vm.trapWithCause("switch", ins, err)
		}
		target := ins.Label
		for _, c := range ins.Labels {
			if int64(v.Int) == c.Value {
				target = c.Label
				break
			}
		}
		if err := fr.jump(target); err != nil {
			return ExitStatus{}, false, vm.trapWithCause("switch", ins, err)
		}
```

Add helper methods below `pop`:

```go
func (fr *frame) jump(label int) error {
	pc, ok := fr.labels[label]
	if !ok {
		return fmt.Errorf("unknown label %d", label)
	}
	fr.pc = pc
	return nil
}

func (vm *VM) binary(ins bytecode.Instr) error {
	r, err := vm.pop(ins.Type)
	if err != nil {
		return err
	}
	l, err := vm.pop(ins.Type)
	if err != nil {
		return err
	}
	var out Value
	switch ins.Binary {
	case bytecode.BinAdd:
		out = IntValue(ins.Type, int64(l.Int+r.Int))
	case bytecode.BinSub:
		out = IntValue(ins.Type, int64(l.Int-r.Int))
	case bytecode.BinMul:
		out = IntValue(ins.Type, int64(l.Int*r.Int))
	case bytecode.BinDivS, bytecode.BinDivU:
		if r.Int == 0 {
			return fmt.Errorf("division by zero")
		}
		out = IntValue(ins.Type, int64(l.Int/r.Int))
	case bytecode.BinRemS, bytecode.BinRemU:
		if r.Int == 0 {
			return fmt.Errorf("remainder by zero")
		}
		out = IntValue(ins.Type, int64(l.Int%r.Int))
	case bytecode.BinAnd:
		out = IntValue(ins.Type, int64(l.Int&r.Int))
	case bytecode.BinOr:
		out = IntValue(ins.Type, int64(l.Int|r.Int))
	case bytecode.BinXor:
		out = IntValue(ins.Type, int64(l.Int^r.Int))
	case bytecode.BinShl:
		out = IntValue(ins.Type, int64(l.Int<<uint(r.Int)))
	case bytecode.BinShrS, bytecode.BinShrU:
		out = IntValue(ins.Type, int64(l.Int>>uint(r.Int)))
	case bytecode.BinEq:
		out = IntValue(bytecode.TypeBool, boolInt(l.Int == r.Int))
	case bytecode.BinNe:
		out = IntValue(bytecode.TypeBool, boolInt(l.Int != r.Int))
	case bytecode.BinLtS, bytecode.BinLtU:
		out = IntValue(bytecode.TypeBool, boolInt(l.Int < r.Int))
	case bytecode.BinLeS, bytecode.BinLeU:
		out = IntValue(bytecode.TypeBool, boolInt(l.Int <= r.Int))
	case bytecode.BinGtS, bytecode.BinGtU:
		out = IntValue(bytecode.TypeBool, boolInt(l.Int > r.Int))
	case bytecode.BinGeS, bytecode.BinGeU:
		out = IntValue(bytecode.TypeBool, boolInt(l.Int >= r.Int))
	default:
		return fmt.Errorf("unsupported binary op %s", ins.Binary)
	}
	vm.stack = append(vm.stack, out)
	return nil
}

func (vm *VM) unary(ins bytecode.Instr) error {
	v, err := vm.pop(ins.Type)
	if err != nil {
		return err
	}
	switch ins.Unary {
	case bytecode.UnaryNeg:
		vm.stack = append(vm.stack, IntValue(ins.Type, -int64(v.Int)))
		return nil
	default:
		return fmt.Errorf("unsupported unary op %d", ins.Unary)
	}
}

func boolInt(v bool) int64 {
	if v {
		return 1
	}
	return 0
}
```

- [ ] **Step 4: Run tests and verify they pass**

Run: `go test ./runtime -run 'TestRunIntegerArithmeticAndBranch|TestRunSwitch|TestRunReturnsMainConstant|TestRunLocalStoreLoad' -count=1`

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add runtime/vm.go runtime/vm_test.go
git commit -m "feat(runtime): execute integer control flow"
```

## Task 8: Address, Memory, Static Data, And Pointer Opcodes

**Files:**
- Modify: `runtime/program.go`
- Modify: `runtime/memory.go`
- Modify: `runtime/vm.go`
- Modify: `runtime/vm_test.go`

- [ ] **Step 1: Add failing tests for global load/store and strings**

Append to `runtime/vm_test.go`:

```go
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
```

- [ ] **Step 2: Run tests and verify they fail**

Run: `go test ./runtime -run 'TestRunGlobalLoadStore|TestRunMemSetAndMemCopy' -count=1`

Expected: FAIL with unsupported opcode traps.

- [ ] **Step 3: Store string addresses in `Program`**

Modify `Program` in `runtime/program.go`:

```go
	stringAddr []uint64
```

Initialize it in `Load`:

```go
stringAddr: make([]uint64, len(mod.Strings)),
```

Replace the string allocation loop in `allocateGlobals`:

```go
	for _, s := range p.module.Strings {
		p.stringAddr[s.ID] = p.memory.AllocBytes(fmt.Sprintf("string:%d", s.ID), s.Bytes, true, blockString)
	}
```

Add:

```go
func (p *Program) StringAddr(id int) uint64 { return p.stringAddr[id] }
func (p *Program) FuncAddr(id int) uint64 { return p.funcAddr[id] }
```

- [ ] **Step 4: Implement address and memory opcodes**

Extend the `switch ins.Op` in `runtime/vm.go`:

```go
	case bytecode.OpAddrGlobal:
		vm.stack = append(vm.stack, ObjectAddrValue(vm.program.GlobalAddr(ins.Global)))
	case bytecode.OpAddrString:
		vm.stack = append(vm.stack, ObjectAddrValue(vm.program.StringAddr(int(ins.Int))))
	case bytecode.OpAddrFunc:
		vm.stack = append(vm.stack, PtrValue(vm.program.FuncAddr(ins.Global)))
	case bytecode.OpLoad:
		addr, err := vm.pop(bytecode.TypeObjectAddr)
		if err != nil {
			return ExitStatus{}, false, vm.trapWithCause("load address", ins, err)
		}
		v, err := vm.program.memory.Load(addr.Int, ins.Type, ins.Align)
		if err != nil {
			return ExitStatus{}, false, vm.trapWithCause("load memory", ins, err)
		}
		vm.stack = append(vm.stack, v)
	case bytecode.OpStore:
		value, err := vm.pop(ins.Type)
		if err != nil {
			return ExitStatus{}, false, vm.trapWithCause("store value", ins, err)
		}
		addr, err := vm.pop(bytecode.TypeObjectAddr)
		if err != nil {
			return ExitStatus{}, false, vm.trapWithCause("store address", ins, err)
		}
		if err := vm.program.memory.Store(addr.Int, ins.Type, ins.Align, value); err != nil {
			return ExitStatus{}, false, vm.trapWithCause("store memory", ins, err)
		}
	case bytecode.OpMemCopy:
		src, err := vm.pop(bytecode.TypeObjectAddr)
		if err != nil {
			return ExitStatus{}, false, vm.trapWithCause("memcopy source", ins, err)
		}
		dst, err := vm.pop(bytecode.TypeObjectAddr)
		if err != nil {
			return ExitStatus{}, false, vm.trapWithCause("memcopy destination", ins, err)
		}
		if err := vm.program.memory.Copy(dst.Int, src.Int, ins.Size); err != nil {
			return ExitStatus{}, false, vm.trapWithCause("memcopy", ins, err)
		}
	case bytecode.OpMemSet:
		value, err := vm.pop(bytecode.TypeI32)
		if err != nil {
			return ExitStatus{}, false, vm.trapWithCause("memset value", ins, err)
		}
		dst, err := vm.pop(bytecode.TypeObjectAddr)
		if err != nil {
			return ExitStatus{}, false, vm.trapWithCause("memset destination", ins, err)
		}
		if err := vm.program.memory.Set(dst.Int, byte(value.Int), ins.Size); err != nil {
			return ExitStatus{}, false, vm.trapWithCause("memset", ins, err)
		}
	case bytecode.OpOffset:
		addr, err := vm.pop(ins.Type)
		if err != nil {
			return ExitStatus{}, false, vm.trapWithCause("offset", ins, err)
		}
		vm.stack = append(vm.stack, Value{Type: ins.Type, Int: addr.Int + uint64(ins.Int)})
	case bytecode.OpPtrAdd:
		index, err := vm.pop(bytecode.TypeI64)
		if err != nil {
			return ExitStatus{}, false, vm.trapWithCause("ptradd index", ins, err)
		}
		base, err := vm.pop(bytecode.TypePtr)
		if err != nil {
			return ExitStatus{}, false, vm.trapWithCause("ptradd base", ins, err)
		}
		vm.stack = append(vm.stack, PtrValue(base.Int+index.Int*uint64(ins.Size)))
	case bytecode.OpPtrDiff:
		right, err := vm.pop(bytecode.TypePtr)
		if err != nil {
			return ExitStatus{}, false, vm.trapWithCause("ptrdiff right", ins, err)
		}
		left, err := vm.pop(bytecode.TypePtr)
		if err != nil {
			return ExitStatus{}, false, vm.trapWithCause("ptrdiff left", ins, err)
		}
		vm.stack = append(vm.stack, IntValue(bytecode.TypeI64, int64((left.Int-right.Int)/uint64(ins.Size))))
```

- [ ] **Step 5: Run tests and verify they pass**

Run: `go test ./runtime -run 'TestRunGlobalLoadStore|TestRunMemSetAndMemCopy' -count=1`

Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add runtime/program.go runtime/memory.go runtime/vm.go runtime/vm_test.go
git commit -m "feat(runtime): execute static memory opcodes"
```

## Task 9: Function Calls And Extern Calls

**Files:**
- Modify: `runtime/program.go`
- Modify: `runtime/vm.go`
- Modify: `runtime/vm_test.go`

- [ ] **Step 1: Add failing call tests**

Append to `runtime/vm_test.go`:

```go
func TestRunDirectCall(t *testing.T) {
	mod := testMainModule(
		bytecode.I32Const(6),
		bytecode.Call(1, 1, 1),
		bytecode.Return(bytecode.TypeI32),
	)
	mod.Sigs = append(mod.Sigs, bytecode.FuncSig{ID: 1, Ret: bytecode.TypeI32, Params: []bytecode.ValueType{bytecode.TypeI32}})
	mod.Globals = append(mod.Globals, bytecode.Global{ID: 1, Name: "inc", Kind: bytecode.GlobalFunc, Func: 1, Sig: 1})
	mod.Functions = append(mod.Functions, bytecode.Function{
		ID: 1, GlobalID: 1, Name: "inc", Sig: 1,
		Params: []bytecode.Param{{Name: "x", Type: bytecode.TypeI32, Slot: 0}},
		Locals: []bytecode.LocalSlot{{ID: 0, Name: "x", Type: bytecode.TypeI32}},
		Instrs: []bytecode.Instr{
			bytecode.LoadLocal(bytecode.TypeI32, 0),
			bytecode.I32Const(1),
			bytecode.Binary(bytecode.TypeI32, bytecode.BinAdd),
			bytecode.Return(bytecode.TypeI32),
		},
	})
	st, err := runModule(t, mod)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if st.Code != 7 {
		t.Fatalf("exit code = %d, want 7", st.Code)
	}
}

func TestRunExitExtern(t *testing.T) {
	mod := testMainModule(
		bytecode.I32Const(33),
		bytecode.Call(1, 1, 1),
		bytecode.I32Const(0),
		bytecode.Return(bytecode.TypeI32),
	)
	mod.Sigs = append(mod.Sigs, bytecode.FuncSig{ID: 1, Ret: bytecode.TypeVoid, Params: []bytecode.ValueType{bytecode.TypeI32}})
	mod.Globals = append(mod.Globals, bytecode.Global{ID: 1, Name: "exit", Kind: bytecode.GlobalExtern, Sig: 1, Extern: bytecode.ExternRef{Name: "exit", ABI: bytecode.DefaultExternABI}})
	st, err := runModule(t, mod)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if st.Code != 33 {
		t.Fatalf("exit code = %d, want 33", st.Code)
	}
}
```

- [ ] **Step 2: Run tests and verify they fail**

Run: `go test ./runtime -run 'TestRunDirectCall|TestRunExitExtern' -count=1`

Expected: FAIL with unsupported `OpCall`.

- [ ] **Step 3: Implement direct and extern calls**

Add helper methods to `runtime/program.go`:

```go
func (p *Program) FuncIDByAddress(addr uint64) (int, bool) {
	for gid, a := range p.funcAddr {
		if a == addr {
			g := p.module.Globals[gid]
			if g.Kind == bytecode.GlobalFunc {
				return g.Func, true
			}
		}
	}
	return 0, false
}

func (p *Program) ExternByGlobal(global int) (ExternFunc, bool) {
	fn, ok := p.externs[global]
	return fn, ok
}
```

Extend `runtime/vm.go` dispatch:

```go
	case bytecode.OpCall:
		st, exited, err := vm.call(ctx, ins.Global, ins.Sig, ins.Argc)
		if err != nil || exited {
			return st, exited, err
		}
	case bytecode.OpCallIndirect:
		callee, err := vm.pop(bytecode.TypePtr)
		if err != nil {
			return ExitStatus{}, false, vm.trapWithCause("call indirect callee", ins, err)
		}
		funcID, ok := vm.program.FuncIDByAddress(callee.Int)
		if !ok {
			return ExitStatus{}, false, vm.trap("invalid indirect call target", ins)
		}
		st, exited, err := vm.callFunction(ctx, funcID, ins.Sig, ins.Argc)
		if err != nil || exited {
			return st, exited, err
		}
```

Add call helpers:

```go
func (vm *VM) call(ctx context.Context, global, sig, argc int) (ExitStatus, bool, error) {
	if global < 0 || global >= len(vm.program.module.Globals) {
		return ExitStatus{}, false, vm.trap("invalid call global", bytecode.Instr{Op: bytecode.OpCall, Global: global})
	}
	g := vm.program.module.Globals[global]
	if g.Kind == bytecode.GlobalExtern {
		fn, ok := vm.program.ExternByGlobal(global)
		if !ok {
			return ExitStatus{}, false, vm.trap("unresolved extern call", bytecode.Instr{Op: bytecode.OpCall, Global: global})
		}
		args, err := vm.popArgs(argc, sig)
		if err != nil {
			return ExitStatus{}, false, vm.trapWithCause("extern arguments", bytecode.Instr{Op: bytecode.OpCall, Global: global}, err)
		}
		ret, exit, err := fn(ctx, vm.programExternContext(), args)
		if err != nil {
			return ExitStatus{}, false, vm.trapWithCause("extern "+g.Extern.Name, bytecode.Instr{Op: bytecode.OpCall, Global: global}, err)
		}
		if exit != nil {
			return *exit, true, nil
		}
		if vm.program.module.Sigs[sig].Ret != bytecode.TypeVoid {
			vm.stack = append(vm.stack, ret)
		}
		return ExitStatus{}, false, nil
	}
	return vm.callFunction(ctx, g.Func, sig, argc)
}

func (vm *VM) callFunction(ctx context.Context, funcID, sig, argc int) (ExitStatus, bool, error) {
	args, err := vm.popArgs(argc, sig)
	if err != nil {
		return ExitStatus{}, false, err
	}
	if err := vm.pushFrame(funcID, args); err != nil {
		return ExitStatus{}, false, err
	}
	return ExitStatus{}, false, nil
}

func (vm *VM) popArgs(argc, sigID int) ([]Value, error) {
	if sigID < 0 || sigID >= len(vm.program.module.Sigs) {
		return nil, fmt.Errorf("invalid signature %d", sigID)
	}
	sig := vm.program.module.Sigs[sigID]
	args := make([]Value, argc)
	for i := argc - 1; i >= 0; i-- {
		want := bytecode.TypeVoid
		if i < len(sig.Params) {
			want = sig.Params[i]
		}
		v, err := vm.pop(want)
		if err != nil {
			return nil, err
		}
		args[i] = v
	}
	return args, nil
}

func (vm *VM) programExternContext() *ExternContext {
	reg := DefaultExternRegistry(nil, nil)
	return reg.context(vm.program.memory)
}
```

After this passes, refactor `Program` to keep the registry used at load time so extern output writers from `RunOptions` can be injected in Task 11.

- [ ] **Step 4: Run tests and verify they pass**

Run: `go test ./runtime -run 'TestRunDirectCall|TestRunExitExtern' -count=1`

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add runtime/program.go runtime/vm.go runtime/vm_test.go
git commit -m "feat(runtime): execute function calls"
```

## Task 10: Local Objects, Field Addresses, Unsupported Opcode Traps

**Files:**
- Modify: `runtime/vm.go`
- Modify: `runtime/vm_test.go`

- [ ] **Step 1: Add failing local-object and trap tests**

Append to `runtime/vm_test.go`:

```go
func TestRunLocalObjectAddress(t *testing.T) {
	mod := testMainModule(
		bytecode.AddrLocalObject(0),
		bytecode.I32Const(44),
		bytecode.Store(bytecode.TypeI32, 4, false),
		bytecode.AddrLocalObject(0),
		bytecode.Load(bytecode.TypeI32, 4, false),
		bytecode.Return(bytecode.TypeI32),
	)
	mod.Functions[0].Objects = []bytecode.LocalObject{{ID: 0, Name: "x", Size: 4, Align: 4}}
	st, err := runModule(t, mod)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if st.Code != 44 {
		t.Fatalf("exit code = %d, want 44", st.Code)
	}
}

func TestRunUnsupportedOpcodeTrap(t *testing.T) {
	_, err := runModule(t, testMainModule(bytecode.Instr{Op: bytecode.OpVaStart, Slot: 0}))
	if err == nil || !strings.Contains(err.Error(), "unsupported opcode") {
		t.Fatalf("Run error = %v, want unsupported opcode", err)
	}
}
```

Add `strings` to the import list in `runtime/vm_test.go`.

- [ ] **Step 2: Run tests and verify they fail**

Run: `go test ./runtime -run 'TestRunLocalObjectAddress|TestRunUnsupportedOpcodeTrap' -count=1`

Expected: local object test FAILs with unsupported opcode; unsupported opcode test may already PASS.

- [ ] **Step 3: Add local object allocation to frames**

Modify `frame` in `runtime/vm.go`:

```go
	objects map[int]uint64
```

In `pushFrame`, allocate local objects:

```go
	objects := make(map[int]uint64)
	for _, obj := range fn.Objects {
		objects[obj.ID] = vm.program.memory.Alloc("local:"+fn.Name+":"+obj.Name, obj.Size, obj.Align, false, blockLocal)
	}
	vm.frames = append(vm.frames, frame{fn: fn, locals: locals, labels: labels, objects: objects})
```

Add dispatch:

```go
	case bytecode.OpAddrLocalObject:
		addr, ok := fr.objects[ins.Object]
		if !ok {
			return ExitStatus{}, false, vm.trap("invalid local object", ins)
		}
		vm.stack = append(vm.stack, ObjectAddrValue(addr))
	case bytecode.OpFieldAddr:
		addr, err := vm.pop(bytecode.TypeObjectAddr)
		if err != nil {
			return ExitStatus{}, false, vm.trapWithCause("field address", ins, err)
		}
		layout := vm.program.module.Layouts[ins.Layout]
		field := layout.Fields[ins.Field]
		vm.stack = append(vm.stack, ObjectAddrValue(addr.Int+uint64(field.Offset)))
```

- [ ] **Step 4: Run tests and verify they pass**

Run: `go test ./runtime -run 'TestRunLocalObjectAddress|TestRunUnsupportedOpcodeTrap' -count=1`

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add runtime/vm.go runtime/vm_test.go
git commit -m "feat(runtime): execute local object addresses"
```

## Task 11: Compiler-To-Runtime Integration

**Files:**
- Create: `runtime/integration_test.go`
- Modify: `runtime/program.go`
- Modify: `runtime/vm.go`
- Modify: `runtime/extern.go`

- [ ] **Step 1: Write failing integration tests**

Create `runtime/integration_test.go`:

```go
package runtime

import (
	"bytes"
	"context"
	"testing"

	"shinya.click/cvm/bytecode"
	"shinya.click/cvm/codegen"
	"shinya.click/cvm/parser"
	"shinya.click/cvm/preprocessor"
	"shinya.click/cvm/sema"
)

func compileAndRun(t *testing.T, src string, stdout *bytes.Buffer) (ExitStatus, error) {
	t.Helper()
	pp, err := preprocessor.PreprocessSource("main.c", src, preprocessor.Options{})
	if err != nil {
		t.Fatalf("preprocess: %v", err)
	}
	candidates, err := parser.NewParser(pp.Tokens).Parse()
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	prog, err := sema.Analyze(candidates)
	if err != nil {
		t.Fatalf("sema: %v", err)
	}
	mod, err := codegen.Generate(prog)
	if err != nil {
		t.Fatalf("codegen: %v", err)
	}
	var buf bytes.Buffer
	if err := bytecode.EncodeModule(&buf, mod); err != nil {
		t.Fatalf("EncodeModule: %v", err)
	}
	reg := DefaultExternRegistry(stdout, nil)
	p, err := Load(bytes.NewReader(buf.Bytes()), LoadOptions{Externs: reg})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	return Run(context.Background(), p, RunOptions{})
}

func TestCompileAndRunReturnArithmetic(t *testing.T) {
	st, err := compileAndRun(t, `int main(void) { int x = 3; int y = 4; return x * y + 2; }`, nil)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if st.Code != 14 {
		t.Fatalf("exit code = %d, want 14", st.Code)
	}
}

func TestCompileAndRunGlobalAndLoop(t *testing.T) {
	st, err := compileAndRun(t, `int g = 2; int main(void) { int i = 0; while (i < 3) { g = g + 1; i = i + 1; } return g; }`, nil)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if st.Code != 5 {
		t.Fatalf("exit code = %d, want 5", st.Code)
	}
}

func TestCompileAndRunPuts(t *testing.T) {
	var out bytes.Buffer
	st, err := compileAndRun(t, `int puts(const char *); int main(void) { puts("hi"); return 0; }`, &out)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if st.Code != 0 || out.String() != "hi\n" {
		t.Fatalf("exit=%d output=%q, want 0 and hi newline", st.Code, out.String())
	}
}
```

- [ ] **Step 2: Run tests and identify missing opcode handlers**

Run: `go test ./runtime -run TestCompileAndRun -count=1`

Expected: FAIL with the first unsupported opcode or type mismatch from real codegen output.

- [ ] **Step 3: Add the missing Phase 1 handlers surfaced by integration**

Patch `runtime/vm.go` for handlers already in the Phase 1 list that integration exposes, using the same patterns from Tasks 7-10. Expected additions include:

```go
	case bytecode.OpLoadConst:
		addr := vm.program.GlobalAddr(ins.Global) + uint64(ins.Int)
		v, err := vm.program.memory.Load(addr, ins.Type, valueSize(vm.program.module.Target, ins.Type))
		if err != nil {
			return ExitStatus{}, false, vm.trapWithCause("load const", ins, err)
		}
		vm.stack = append(vm.stack, v)
	case bytecode.OpUnreachable:
		return ExitStatus{}, false, vm.trap("unreachable", ins)
```

If codegen emits `TypeObjectAddr` where a pointer is expected, allow bit casts between `TypePtr` and `TypeObjectAddr` in the `OpCast` case:

```go
		if (ins.Type == bytecode.TypePtr && ins.Type2 == bytecode.TypeObjectAddr) ||
			(ins.Type == bytecode.TypeObjectAddr && ins.Type2 == bytecode.TypePtr) {
			vm.stack = append(vm.stack, Value{Type: ins.Type2, Int: v.Int})
			break
		}
```

Keep any opcode outside Phase 1 returning `unsupported opcode`.

- [ ] **Step 4: Run integration tests until they pass**

Run: `go test ./runtime -run TestCompileAndRun -count=1`

Expected: PASS.

- [ ] **Step 5: Run full runtime package tests**

Run: `go test ./runtime -count=1`

Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add runtime/integration_test.go runtime/program.go runtime/vm.go runtime/extern.go
git commit -m "test(runtime): run compiled bytecode"
```

## Task 12: CLI `cvm run`

**Files:**
- Modify: `main.go`
- Modify: `compiler_test.go`

- [ ] **Step 1: Write failing CLI test**

Append to `compiler_test.go`:

```go
func TestMainRunBytecodeUsesExitCode(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "main.c")
	out := filepath.Join(dir, "main.cvmbc")
	if err := os.WriteFile(src, []byte(`int main(void) { return 0; }`), 0644); err != nil {
		t.Fatalf("write source: %v", err)
	}
	if err := (&Compiler{EmitBytecode: out}).RunFile(src); err != nil {
		t.Fatalf("emit bytecode: %v", err)
	}
	oldArgs := os.Args
	os.Args = []string{"cvm", "run", out}
	defer func() { os.Args = oldArgs }()
	main()
}
```

This test uses return code `0` because calling `os.Exit` inside `main()` would terminate the test process. The CLI implementation should keep `main()` testable by routing exit through a helper.

- [ ] **Step 2: Run test and verify it fails**

Run: `go test ./ -run TestMainRunBytecodeUsesExitCode -count=1`

Expected: FAIL because `cvm run` is not implemented.

- [ ] **Step 3: Refactor CLI through `runMain` helper**

Modify `main.go` to import runtime and context:

```go
import (
	"context"
	"fmt"
	"os"

	cvmruntime "shinya.click/cvm/runtime"
)
```

Replace `main()` with:

```go
func main() {
	os.Exit(runMain(os.Args[1:]))
}

func runMain(args []string) int {
	if len(args) > 0 && args[0] == "run" {
		if len(args) != 2 {
			fmt.Fprintln(os.Stderr, "Usage: cvm run file.cvmbc")
			return 2
		}
		f, err := os.Open(args[1])
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		defer f.Close()
		prog, err := cvmruntime.Load(f, cvmruntime.LoadOptions{})
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		st, err := cvmruntime.Run(context.Background(), prog, cvmruntime.RunOptions{})
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		return st.Code
	}
	return runCompileMode(args)
}
```

Move the existing compile-mode argument parsing into:

```go
func runCompileMode(args []string) int {
	dumpIR := false
	dumpBytecode := false
	emitBytecode := ""
	files := make([]string, 0, 1)
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--dump-ir":
			dumpIR = true
		case "--dump-bytecode":
			dumpBytecode = true
		case "--emit-bytecode":
			i++
			if i >= len(args) {
				fmt.Println("Usage: cvm [--dump-ir|--dump-bytecode|--emit-bytecode out.cvmbc] [file]")
				return 2
			}
			emitBytecode = args[i]
		default:
			files = append(files, arg)
		}
	}
	if len(files) != 1 {
		fmt.Println("Usage: cvm [--dump-ir|--dump-bytecode|--emit-bytecode out.cvmbc] [file]")
		return 2
	}
	c := &Compiler{DumpIR: dumpIR, DumpBytecode: dumpBytecode, EmitBytecode: emitBytecode}
	if err := c.RunFile(files[0]); err != nil {
		c.handleError(err)
		return 1
	}
	return 0
}
```

Update existing tests that call `main()` with nonzero bytecode return values to call `runMain` instead when they need to assert exit codes.

- [ ] **Step 4: Run CLI tests**

Run: `go test ./ -run 'TestMainEmitBytecodeFlagWritesLoadableBinaryModule|TestMainRunBytecodeUsesExitCode' -count=1`

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add main.go compiler_test.go
git commit -m "feat(runtime): add cvm run command"
```

## Task 13: Execution Fixture Scaffold

**Files:**
- Create: `runtime/testdata/gcc-exec/README.md`
- Create: `runtime/gcc_exec_test.go`

- [ ] **Step 1: Add fixture documentation**

Create `runtime/testdata/gcc-exec/README.md`:

```markdown
# GCC Execution Fixtures

This directory is for GCC-derived execution fixtures that are deterministic
under the cvm bytecode runtime.

Each fixture must document:

- expected exit code;
- required externs;
- skip reason when unsupported;
- whether hosted C library behavior is required.

Do not copy the compile-only GCC fixture set wholesale. Add execution fixtures
only when the runtime implements the required bytecode and extern behavior.
```

- [ ] **Step 2: Add an empty gate test**

Create `runtime/gcc_exec_test.go`:

```go
package runtime

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGCCExecutionFixtureDirectoryExists(t *testing.T) {
	path := filepath.Join("testdata", "gcc-exec")
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat %s: %v", path, err)
	}
	if !info.IsDir() {
		t.Fatalf("%s is not a directory", path)
	}
}
```

- [ ] **Step 3: Run fixture scaffold test**

Run: `go test ./runtime -run TestGCCExecutionFixtureDirectoryExists -count=1`

Expected: PASS.

- [ ] **Step 4: Commit**

```bash
git add runtime/testdata/gcc-exec/README.md runtime/gcc_exec_test.go
git commit -m "test(runtime): scaffold GCC execution fixtures"
```

## Task 14: Full Verification And Review

**Files:**
- Review all files changed by Tasks 1-13.

- [ ] **Step 1: Run focused runtime tests**

Run: `go test ./runtime -count=1`

Expected: PASS.

- [ ] **Step 2: Run root CLI/compiler tests**

Run: `go test ./ -count=1`

Expected: PASS.

- [ ] **Step 3: Run full test suite**

Run: `go test ./... -count=1`

Expected: PASS.

- [ ] **Step 4: Run bytecode/codegen gates**

Run: `go test ./bytecode ./codegen -count=1`

Expected: PASS.

- [ ] **Step 5: Run diff formatting check**

Run: `git diff --check`

Expected: no output and exit status 0.

- [ ] **Step 6: Review runtime error surfaces**

Run:

```bash
rg -n "unsupported opcode|unsupported extern|panic\\(" runtime main.go compiler_test.go
```

Expected:
- `unsupported opcode` appears in intentional trap paths and tests.
- `unsupported extern` appears only in loader/trap errors and tests.
- no `panic(` in runtime implementation.

- [ ] **Step 7: Final commit if verification required small fixes**

If Step 1-6 required fixes, commit them:

```bash
git add runtime main.go compiler_test.go
git commit -m "fix(runtime): finish phase 1 verification"
```

## Self-Review

Spec coverage:

- `.cvmbc` loading: Task 4.
- entry checks and no-argument integer `main`: Task 4.
- static globals and relocations: Task 4 and Task 8.
- VM stack, frames, locals, calls, returns: Task 6 and Task 9.
- integer/control-flow opcodes: Task 7.
- address and memory opcodes: Task 8 and Task 10.
- minimal externs: Task 5 and Task 9.
- CLI `cvm run`: Task 12.
- compiler-to-runtime integration: Task 11.
- GCC execution fixture layering: Task 13.
- full verification: Task 14.

Red-flag scan:

- No task uses unresolved-marker wording or vague edge-case instructions.
- Unsupported Phase 2-4 features remain explicit runtime traps.

Type consistency:

- Public APIs match the design: `Load`, `Run`, `LoadOptions`, `RunOptions`, `Program`, `ExitStatus`.
- Runtime consumes `bytecode.Module` only.
- `Value`, `Memory`, `ExternRegistry`, and `VM` are introduced before dependent tasks rely on them.
