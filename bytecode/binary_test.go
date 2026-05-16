package bytecode

import (
	"bytes"
	"reflect"
	"strings"
	"testing"
)

func TestBinaryModuleRoundTrip(t *testing.T) {
	mod := binaryFixtureModule()
	var buf bytes.Buffer
	if err := EncodeModule(&buf, mod); err != nil {
		t.Fatalf("EncodeModule: %v", err)
	}
	if bytes.Contains(buf.Bytes(), []byte("Module version=")) {
		t.Fatalf("binary encoding contains printer text:\n%x", buf.Bytes())
	}
	got, err := DecodeModule(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("DecodeModule: %v", err)
	}
	if !reflect.DeepEqual(got, mod) {
		t.Fatalf("round-trip mismatch\nwant:\n%s\ngot:\n%s", PrintModule(mod), PrintModule(got))
	}
	if err := ValidateModule(got); err != nil {
		t.Fatalf("decoded module failed validation: %v", err)
	}
}

func TestDecodeModuleRejectsCorruptPayload(t *testing.T) {
	mod := binaryFixtureModule()
	var buf bytes.Buffer
	if err := EncodeModule(&buf, mod); err != nil {
		t.Fatalf("EncodeModule: %v", err)
	}
	data := append([]byte(nil), buf.Bytes()...)
	data[len(data)-1] ^= 0xff
	_, err := DecodeModule(bytes.NewReader(data))
	if err == nil {
		t.Fatal("DecodeModule accepted corrupt payload")
	}
	if !strings.Contains(err.Error(), "checksum") {
		t.Fatalf("DecodeModule error = %v, want checksum failure", err)
	}
}

func TestDecodeModuleRejectsTrailingFileBytes(t *testing.T) {
	mod := binaryFixtureModule()
	var buf bytes.Buffer
	if err := EncodeModule(&buf, mod); err != nil {
		t.Fatalf("EncodeModule: %v", err)
	}
	buf.Write([]byte{0xde, 0xad})
	_, err := DecodeModule(bytes.NewReader(buf.Bytes()))
	if err == nil {
		t.Fatal("DecodeModule accepted trailing file bytes")
	}
	if !strings.Contains(err.Error(), "trailing") {
		t.Fatalf("DecodeModule error = %v, want trailing data failure", err)
	}
}

func binaryFixtureModule() *Module {
	return &Module{
		Version: CurrentModuleVersion,
		Entry:   &EntryPoint{Global: 1, Name: "main"},
		Target:  DefaultTarget(),
		Globals: []Global{
			{
				ID:    0,
				Name:  "g",
				Kind:  GlobalVar,
				Func:  NoEntryGlobal,
				Sig:   NoFuncSig,
				Size:  8,
				Align: 8,
				Init: InitData{
					Bytes:       []byte{1, 2, 3, 4, 5, 6, 7, 8},
					Relocations: []Relocation{{Offset: 0, Kind: RelocString, Target: 0, Addend: 2}},
				},
			},
			{ID: 1, Name: "main", Kind: GlobalFunc, Func: 0, Sig: 0},
			{ID: 2, Name: "puts", Kind: GlobalExtern, Func: -1, Sig: 1, Extern: ExternRef{Name: "puts", ABI: DefaultExternABI}},
		},
		Strings: []StringConst{{ID: 0, Value: "ok", Bytes: []byte{'o', 'k', 0}}},
		Layouts: []ObjectLayout{{
			ID:     0,
			Name:   "int",
			Size:   4,
			Align:  4,
			Fields: []FieldLayout{{ID: 0, Name: "value", Offset: 0, Type: TypeI32}},
			Bit:    []BitFieldLayout{{ID: 0, Name: "flag", Container: TypeU32, ByteOffset: 0, BitOffset: 1, Width: 3, Signed: true, LayoutPolicy: "cvm"}},
		}},
		Sigs: []FuncSig{
			{ID: 0, Ret: TypeI32},
			{ID: 1, Ret: TypeI32, Params: []ValueType{TypePtr}},
		},
		Functions: []Function{{
			ID:       0,
			GlobalID: 1,
			Name:     "main",
			Sig:      0,
			Params:   []Param{{Name: "argc", Type: TypeI32, Slot: 0}},
			Locals:   []LocalSlot{{ID: 1, Name: "tmp", Type: TypeI32}},
			Objects:  []LocalObject{{ID: 0, Name: "obj", Size: 4, Align: 4, Layout: 0}},
			Labels:   []Label{{ID: 0, Name: "done", Stack: []ValueType{TypeI32}, Statement: true}},
			MaxStack: 2,
			Instrs: []Instr{
				AddrString(0),
				Cast(TypeObjectAddr, TypePtr, CastBit),
				Call(2, 1, 1),
				{Op: OpPop},
				I32Const(0),
				Return(TypeI32),
			},
		}},
	}
}
