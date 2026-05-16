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

func TestMemoryReadWritePointer32BitTarget(t *testing.T) {
	target := bytecode.DefaultTarget()
	target.PointerSize = 4
	target.PointerAlign = 4
	mem := NewMemory(target)
	addr := mem.Alloc("global:p", 4, 4, false, blockGlobal)
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
