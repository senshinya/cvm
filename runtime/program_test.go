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
