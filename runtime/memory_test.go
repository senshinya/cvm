package runtime

import (
	"bytes"
	"math"
	"strings"
	"testing"

	"shinya.click/cvm/bytecode"
)

func TestMemoryReadWriteScalar(t *testing.T) {
	mem := NewMemory(bytecode.DefaultTarget())
	addr := mustAlloc(t, mem, "global:g", 8, 4, false, blockGlobal)
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
	addr := mustAlloc(t, mem, "string:0", 4, 1, true, blockString)
	err := mem.Store(addr, bytecode.TypeI8, 1, IntValue(bytecode.TypeI8, 1))
	if err == nil || !strings.Contains(err.Error(), "readonly") {
		t.Fatalf("Store error = %v, want readonly", err)
	}
}

func TestMemoryReadCString(t *testing.T) {
	mem := NewMemory(bytecode.DefaultTarget())
	addr := mustAllocBytes(t, mem, "string:0", []byte("hello\x00extra"), true, blockString)
	got, err := mem.ReadCString(addr)
	if err != nil {
		t.Fatalf("ReadCString: %v", err)
	}
	if got != "hello" {
		t.Fatalf("ReadCString = %q, want hello", got)
	}
}

func TestMemoryReadWritePointer32BitTarget(t *testing.T) {
	target := bytecode.DefaultTarget()
	target.PointerSize = 4
	target.PointerAlign = 4
	mem := NewMemory(target)
	addr := mustAlloc(t, mem, "global:p", 4, 4, false, blockGlobal)
	if err := mem.Store(addr, bytecode.TypePtr, 4, PtrValue(0x11223344)); err != nil {
		t.Fatalf("Store: %v", err)
	}
	got, err := mem.Load(addr, bytecode.TypePtr, 4)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got.Int != 0x11223344 {
		t.Fatalf("loaded %#x, want 0x11223344", got.Int)
	}
}

func TestMemoryRejectsPointerOverflow32BitTarget(t *testing.T) {
	target := bytecode.DefaultTarget()
	target.PointerSize = 4
	target.PointerAlign = 4
	mem := NewMemory(target)
	addr := mustAlloc(t, mem, "global:p", 4, 4, false, blockGlobal)
	err := mem.Store(addr, bytecode.TypePtr, 4, PtrValue(math.MaxUint32+1))
	if err == nil || !strings.Contains(err.Error(), "exceeds 32-bit pointer") {
		t.Fatalf("Store error = %v, want pointer overflow", err)
	}
}

func TestMemoryReadWriteFloat64(t *testing.T) {
	mem := NewMemory(bytecode.DefaultTarget())
	addr := mustAlloc(t, mem, "global:f", 8, 8, false, blockGlobal)
	if err := mem.Store(addr, bytecode.TypeF64, 8, FloatValue(bytecode.TypeF64, 3.25)); err != nil {
		t.Fatalf("Store: %v", err)
	}
	got, err := mem.Load(addr, bytecode.TypeF64, 8)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got.Float != 3.25 {
		t.Fatalf("loaded %v, want 3.25", got.Float)
	}
}

func TestMemoryReadWriteFloat32(t *testing.T) {
	mem := NewMemory(bytecode.DefaultTarget())
	addr := mustAlloc(t, mem, "global:f32", 4, 4, false, blockGlobal)
	if err := mem.Store(addr, bytecode.TypeF32, 4, FloatValue(bytecode.TypeF32, 1.5)); err != nil {
		t.Fatalf("Store: %v", err)
	}
	got, err := mem.Load(addr, bytecode.TypeF32, 4)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got.Float != 1.5 {
		t.Fatalf("loaded %v, want 1.5", got.Float)
	}
}

func TestMemoryReadWriteLongDouble(t *testing.T) {
	mem := NewMemory(bytecode.DefaultTarget())
	addr := mustAlloc(t, mem, "global:flong", 16, 8, false, blockGlobal)
	if err := mem.Store(addr, bytecode.TypeFLong, 8, FloatValue(bytecode.TypeFLong, 1.5)); err != nil {
		t.Fatalf("Store: %v", err)
	}
	got, err := mem.Load(addr, bytecode.TypeFLong, 8)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got.Type != bytecode.TypeFLong || got.Float != 1.5 {
		t.Fatalf("loaded %#v, want long double 1.5", got)
	}
	if raw := mem.blocks[len(mem.blocks)-1].data[8:16]; !bytes.Equal(raw, make([]byte, 8)) {
		t.Fatalf("long double padding bytes = %v, want zero", raw)
	}
}

func TestMemoryRejectsUnknownEndian(t *testing.T) {
	target := bytecode.DefaultTarget()
	target.Endian = "middle"
	mem := NewMemory(target)
	addr := mustAlloc(t, mem, "global:g", 2, 2, false, blockGlobal)
	err := mem.Store(addr, bytecode.TypeI16, 2, IntValue(bytecode.TypeI16, 1))
	if err == nil || !strings.Contains(err.Error(), "unsupported endian") {
		t.Fatalf("Store error = %v, want unsupported endian", err)
	}
}

func TestMemoryBoundsAlignmentCopySetAndCStringErrors(t *testing.T) {
	mem := NewMemory(bytecode.DefaultTarget())
	addr := mustAlloc(t, mem, "global:g", 4, 4, false, blockGlobal)
	if _, err := mem.Load(addr+2, bytecode.TypeI32, 1); err == nil || !strings.Contains(err.Error(), "invalid memory access") {
		t.Fatalf("Load bounds error = %v, want invalid memory access", err)
	}
	if err := mem.Store(addr+1, bytecode.TypeI8, 2, IntValue(bytecode.TypeI8, 1)); err == nil || !strings.Contains(err.Error(), "unaligned") {
		t.Fatalf("Store alignment error = %v, want unaligned", err)
	}

	src := mustAlloc(t, mem, "global:src", 4, 1, false, blockGlobal)
	dst := mustAlloc(t, mem, "global:dst", 4, 1, false, blockGlobal)
	if err := mem.Set(src, 0x7f, 4); err != nil {
		t.Fatalf("Set: %v", err)
	}
	if err := mem.Copy(dst, src, 4); err != nil {
		t.Fatalf("Copy: %v", err)
	}
	got, err := mem.Load(dst+3, bytecode.TypeU8, 1)
	if err != nil {
		t.Fatalf("Load copied byte: %v", err)
	}
	if got.Int != 0x7f {
		t.Fatalf("copied byte %#x, want 0x7f", got.Int)
	}

	cstr := mustAllocBytes(t, mem, "string:unterminated", []byte("hello"), true, blockString)
	if _, err := mem.ReadCString(cstr); err == nil || !strings.Contains(err.Error(), "unterminated") {
		t.Fatalf("ReadCString error = %v, want unterminated", err)
	}
}

func TestMemoryTryAllocRejectsInvalidRanges(t *testing.T) {
	mem := NewMemory(bytecode.DefaultTarget())
	if _, err := mem.TryAlloc("bad:negative", -1, 1, false, blockGlobal); err == nil || !strings.Contains(err.Error(), "negative") {
		t.Fatalf("TryAlloc negative error = %v, want negative", err)
	}
	if len(mem.blocks) != 0 {
		t.Fatalf("TryAlloc mutated blocks after negative size")
	}

	mem.next = math.MaxUint64 - 3
	if _, err := mem.TryAlloc("bad:align", 1, 8, false, blockGlobal); err == nil || !strings.Contains(err.Error(), "overflow") {
		t.Fatalf("TryAlloc align overflow error = %v, want overflow", err)
	}

	mem.next = math.MaxUint64 - 1
	if _, err := mem.TryAlloc("bad:next", 1, 1, false, blockGlobal); err == nil || !strings.Contains(err.Error(), "next address") {
		t.Fatalf("TryAlloc next overflow error = %v, want next address overflow", err)
	}
}
