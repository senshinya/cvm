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

func TestLoadAlignsStringConstantsForWideAccess(t *testing.T) {
	mod := testMainModule(bytecode.I32Const(0), bytecode.Return(bytecode.TypeI32))
	mod.Strings = []bytecode.StringConst{
		{ID: 0, Value: "x", Bytes: []byte{'x', 0}},
		{ID: 1, Value: "%d", Bytes: []byte{'%', 0, 0, 0, 'd', 0, 0, 0, 0, 0, 0, 0}},
	}
	var buf bytes.Buffer
	if err := bytecode.EncodeModule(&buf, mod); err != nil {
		t.Fatalf("EncodeModule: %v", err)
	}
	p, err := Load(bytes.NewReader(buf.Bytes()), LoadOptions{})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	for _, s := range mod.Strings {
		addr := p.stringAddr[s.ID]
		if addr%4 != 0 {
			t.Fatalf("string %d address %#x is not 4-byte aligned", s.ID, addr)
		}
	}
	if got, err := p.Memory().Load(p.stringAddr[1], bytecode.TypeI32, 4); err != nil || got.Int != '%' {
		t.Fatalf("wide string first element = %#v err=%v, want %%", got, err)
	}
}

func TestLoadReturnsErrorForStdoutExternVariableWithUnsupportedPointerSize(t *testing.T) {
	mod := testMainModule(bytecode.I32Const(0), bytecode.Return(bytecode.TypeI32))
	mod.Target.PointerSize = 3
	mod.Globals = append(mod.Globals,
		bytecode.Global{
			ID:     1,
			Name:   "stdout",
			Kind:   bytecode.GlobalExtern,
			Size:   mod.Target.PointerSize,
			Align:  mod.Target.PointerAlign,
			Extern: bytecode.ExternRef{Name: "stdout", ABI: bytecode.DefaultExternABI},
		},
	)

	var buf bytes.Buffer
	if err := bytecode.EncodeModule(&buf, mod); err != nil {
		t.Fatalf("EncodeModule: %v", err)
	}
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Load panicked: %v", r)
		}
	}()
	_, err := Load(bytes.NewReader(buf.Bytes()), LoadOptions{})
	if err == nil {
		t.Fatal("Load error = nil, want unsupported stdout extern variable failure")
	}
	if msg := err.Error(); !strings.Contains(msg, "stdout") && !strings.Contains(msg, "pointer size") {
		t.Fatalf("Load error = %v, want stdout or pointer size", err)
	}
}

func TestLoadRejectsNilExternFunction(t *testing.T) {
	mod := testMainModule(bytecode.I32Const(0), bytecode.Return(bytecode.TypeI32))
	mod.Sigs = append(mod.Sigs, bytecode.FuncSig{ID: 1, Ret: bytecode.TypeVoid})
	mod.Globals = append(mod.Globals, bytecode.Global{
		ID:     1,
		Name:   "nil_ext",
		Kind:   bytecode.GlobalExtern,
		Sig:    1,
		Extern: bytecode.ExternRef{Name: "nil_ext", ABI: bytecode.DefaultExternABI},
	})
	var buf bytes.Buffer
	if err := bytecode.EncodeModule(&buf, mod); err != nil {
		t.Fatalf("EncodeModule: %v", err)
	}
	reg := NewExternRegistry(nil, nil)
	reg.Register("nil_ext", nil)
	_, err := Load(bytes.NewReader(buf.Bytes()), LoadOptions{Externs: reg})
	if err == nil || !strings.Contains(err.Error(), "unresolved extern nil_ext") {
		t.Fatalf("Load error = %v, want unresolved nil extern", err)
	}
}
