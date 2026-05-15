package bytecode

import (
	"fmt"
	"strings"
	"testing"
)

func TestPrintModuleStableMinimalFunction(t *testing.T) {
	mod := &Module{
		Target:  DefaultTarget(),
		Globals: []Global{{ID: 0, Name: "main", Kind: GlobalFunc, Func: 0}},
		Sigs:    []FuncSig{{ID: 0, Ret: TypeI32}},
		Functions: []Function{{
			ID:       0,
			GlobalID: 0,
			Name:     "main",
			Sig:      0,
			Instrs: []Instr{
				I32Const(0),
				Return(TypeI32),
			},
		}},
	}
	got := PrintModule(mod)
	for _, want := range []string{
		"Module target=\"cvm-default\"",
		"Global #0 func name=\"main\" func=0",
		"Sig #0 ret=i32 params=()",
		"Func #0 global=0 name=\"main\" sig=0",
		"0000: I32Const 0",
		"0001: I32Return",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("PrintModule missing %q:\n%s", want, got)
		}
	}
}

func TestPrintModuleFullStateExactOutput(t *testing.T) {
	mod := &Module{
		Target: TargetInfo{
			Name:           "test-target",
			Endian:         "big",
			PointerSize:    4,
			PointerAlign:   4,
			BoolSize:       1,
			BoolAlign:      1,
			BitFieldPolicy: "test-policy",
			LayoutVersion:  "7",
		},
		Globals: []Global{
			{
				ID:       0,
				Name:     "data",
				Kind:     GlobalVar,
				Size:     16,
				Align:    8,
				Readonly: true,
				Init: InitData{
					ZeroFill: 4,
					Bytes:    []byte{0x01, 0x02, 0xff},
					Relocations: []Relocation{
						{Offset: 0, Kind: RelocGlobal, Target: 0, Addend: 1},
						{Offset: 1, Kind: RelocFunc, Target: 0, Addend: 2},
						{Offset: 2, Kind: RelocString, Target: 0, Addend: 9},
					},
				},
			},
			{ID: 1, Name: "main", Kind: GlobalFunc, Func: 0},
			{ID: 2, Name: "puts", Kind: GlobalExtern},
		},
		Strings: []StringConst{{ID: 0, Value: "hi", Bytes: []byte{'h', 'i', 0}}},
		Layouts: []ObjectLayout{{
			ID:       0,
			Name:     "Pair",
			Size:     12,
			Align:    4,
			ElemSize: 6,
			Fields: []FieldLayout{{
				ID:     0,
				Name:   "x",
				Offset: 0,
				Type:   TypeI32,
			}},
			Bit: []BitFieldLayout{{
				ID:           1,
				Name:         "flag",
				Container:    TypeU32,
				ByteOffset:   4,
				BitOffset:    3,
				Width:        5,
				Signed:       true,
				Volatile:     true,
				LayoutPolicy: "test-policy",
			}},
		}},
		Sigs: []FuncSig{{ID: 0, Ret: TypeI32, Params: []ValueType{TypeI32, TypePtr}, Variadic: true}},
		Functions: []Function{{
			ID:       0,
			GlobalID: 1,
			Name:     "main",
			Sig:      0,
			Params: []Param{{
				Name: "argc",
				Type: TypeI32,
				Slot: 0,
			}},
			Locals: []LocalSlot{{
				ID:   1,
				Name: "tmp",
				Type: TypeI64,
			}},
			Objects: []LocalObject{{
				ID:     0,
				Name:   "obj",
				Size:   12,
				Align:  4,
				Layout: 0,
			}},
			DynamicObjects: []DynamicObject{{
				ID:     1,
				Name:   "dyn",
				Align:  8,
				Layout: 0,
			}},
			Labels: []Label{{
				ID:        3,
				Name:      "done",
				Stack:     []ValueType{TypeI32, TypePtr},
				Statement: true,
			}},
			MaxStack: 5,
			Instrs: []Instr{
				AddrString(0),
				AddrGlobal(0),
				{Op: OpLoad, Type: TypeI32, Align: 4, Volatile: true},
				{Op: OpBinary, Type: TypeI32, Binary: BinRemU},
				{Op: OpCast, Type: TypeI32, Type2: TypeI64, Cast: CastSExt},
				{Op: OpSwitch, Type: TypeI32, Label: 3, Labels: []SwitchCase{{Value: 7, Label: 4}, {Value: 8, Label: 5}}},
				Call(2, 0, 1),
				Return(TypeI32),
			},
		}},
	}

	const want = "Module target=\"test-target\" endian=big ptr_size=4 ptr_align=4 bool_size=1 bool_align=1 bitfield_policy=\"test-policy\" layout_version=\"7\"\n" +
		"Global #0 var name=\"data\" size=16 align=8 readonly=true init_zero=4 init_bytes=3 init_relocs=3\n" +
		"  InitBytes hex=0102ff\n" +
		"  reloc offset=0 kind=global target=global#0(\"data\") addend=1\n" +
		"  reloc offset=1 kind=func target=func#0(\"main\") addend=2\n" +
		"  reloc offset=2 kind=string target=string#0(\"hi\") addend=9\n" +
		"Global #1 func name=\"main\" func=0\n" +
		"Global #2 extern name=\"puts\"\n" +
		"String #0 value=\"hi\" bytes=3 hex=686900\n" +
		"Layout #0 name=\"Pair\" size=12 align=4 elem_size=6\n" +
		"  Field #0 name=\"x\" offset=0 type=i32\n" +
		"  BitField #1 name=\"flag\" container=u32 byte_offset=4 bit_offset=3 width=5 signed=true volatile=true policy=\"test-policy\"\n" +
		"Sig #0 ret=i32 params=(i32, ptr, ...) variadic\n" +
		"Func #0 global=1 name=\"main\" sig=0 max_stack=5\n" +
		"  Param slot=0 name=\"argc\" type=i32\n" +
		"  Local #1 name=\"tmp\" type=i64\n" +
		"  Object #0 name=\"obj\" size=12 align=4 layout=0\n" +
		"  DynamicObject #1 name=\"dyn\" align=8 layout=0\n" +
		"  Label #3 name=\"done\" stack=(i32, ptr) statement=true\n" +
		"  0000: AddrString 0\n" +
		"  0001: AddrGlobal 0\n" +
		"  0002: I32Load align=4 volatile=true\n" +
		"  0003: I32RemU\n" +
		"  0004: Cast i32->i64 SExt\n" +
		"  0005: Switch i32 default=L3 cases=(7:L4, 8:L5)\n" +
		"  0006: Call global=2 sig=0 argc=1\n" +
		"  0007: I32Return\n"
	if got := PrintModule(mod); got != want {
		t.Fatalf("PrintModule mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestFormatInstrCoversCurrentOpcodeEnum(t *testing.T) {
	tests := []struct {
		name string
		in   Instr
		want string
	}{
		{"const", I64Const(7), "I64Const 7"},
		{"addr-string", AddrString(1), "AddrString 1"},
		{"addr-global", AddrGlobal(2), "AddrGlobal 2"},
		{"addr-func", AddrFunc(3), "AddrFunc 3"},
		{"load-const", Instr{Op: OpLoadConst, Type: TypeI32, Global: 4, Int: 8}, "I32LoadConst global=4 offset=8"},
		{"dup", Instr{Op: OpDup}, "Dup"},
		{"pop", Instr{Op: OpPop}, "Pop"},
		{"swap", Instr{Op: OpSwap}, "Swap"},
		{"load-local", LoadLocal(TypeI16, 5), "I16LoadLocal 5"},
		{"store-local", StoreLocal(TypeI16, 5), "I16StoreLocal 5"},
		{"addr-local-object", AddrLocalObject(6), "AddrLocalObject 6"},
		{"alloc-dynamic-object", Instr{Op: OpAllocDynamicObject, Object: 7, Size: 32, Align: 8, Layout: 9}, "AllocDynamicObject object=7 size=<stack:i64> align=8 layout=9"},
		{"free-dynamic-object", Instr{Op: OpFreeDynamicObject, Object: 7}, "FreeDynamicObject 7"},
		{"dynamic-object-addr", Instr{Op: OpDynamicObjectAddr, Object: 7, Type: TypeObjectAddr}, "DynamicObjectAddr 7"},
		{"load", Load(TypeU32, 4, true), "U32Load align=4 volatile=true"},
		{"store", Store(TypeU32, 4, true), "U32Store align=4 volatile=true"},
		{"mem-copy", Instr{Op: OpMemCopy, Size: 12, Align: 4, Volatile: true}, "MemCopy size=12 align=4 volatile=true"},
		{"mem-set", Instr{Op: OpMemSet, Size: 12, Align: 4, Volatile: true}, "MemSet size=12 align=4 volatile=true"},
		{"offset", Instr{Op: OpOffset, Type: TypePtr, Int: 16}, "PtrOffset 16"},
		{"field-addr", Instr{Op: OpFieldAddr, Layout: 1, Field: 2}, "FieldAddr layout=1 field=2"},
		{"bitfield-load", Instr{Op: OpBitFieldLoad, Type: TypeI32, Layout: 1, Field: 2, Volatile: true}, "I32BitFieldLoad layout=1 field=2 volatile=true"},
		{"bitfield-store", Instr{Op: OpBitFieldStore, Type: TypeI32, Layout: 1, Field: 2, Volatile: true}, "I32BitFieldStore layout=1 field=2 volatile=true"},
		{"ptr-add", Instr{Op: OpPtrAdd, Size: 4}, "PtrAdd elem_size=4"},
		{"ptr-add-dynamic", Instr{Op: OpPtrAddDynamic}, "PtrAddDynamic"},
		{"ptr-diff", Instr{Op: OpPtrDiff, Size: 4}, "PtrDiff elem_size=4"},
		{"binary", Binary(TypeI32, BinShrS), "I32ShrS"},
		{"unary", Instr{Op: OpUnary, Type: TypeI32, Unary: UnaryNeg}, "I32Neg"},
		{"cast", Cast(TypeI32, TypeI64, CastSExt), "Cast i32->i64 SExt"},
		{"label", LabelInstr(8), "L8:"},
		{"jump", Jump(8), "Jump L8"},
		{"jump-if-zero", JumpIfZero(TypeI32, 8), "JumpIfZero i32 L8"},
		{"jump-if-non-zero", JumpIfNonZero(TypeI32, 8), "JumpIfNonZero i32 L8"},
		{"switch", Instr{Op: OpSwitch, Type: TypeI32, Label: 9, Labels: []SwitchCase{{Value: 1, Label: 2}}}, "Switch i32 default=L9 cases=(1:L2)"},
		{"return", Return(TypeI32), "I32Return"},
		{"return-void", Instr{Op: OpReturnVoid}, "ReturnVoid"},
		{"return-object", Instr{Op: OpReturnObject, Object: 3}, "ReturnObject 3"},
		{"unreachable", Instr{Op: OpUnreachable}, "Unreachable"},
		{"call", Call(1, 2, 3), "Call global=1 sig=2 argc=3"},
		{"call-indirect", Instr{Op: OpCallIndirect, Sig: 2, Argc: 3}, "CallIndirect sig=2 argc=3"},
		{"va-start", Instr{Op: OpVaStart, Slot: 1}, "VaStart slot=1"},
		{"va-arg", Instr{Op: OpVaArg, Type: TypeI64}, "I64VaArg"},
		{"va-end", Instr{Op: OpVaEnd, Slot: 1}, "VaEnd slot=1"},
		{"invalid", Instr{Op: Opcode(999), Type: TypeI32, Int: 7}, "InvalidOpcode(999)"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FormatInstr(tt.in); got != tt.want {
				t.Fatalf("FormatInstr(%v) = %q, want %q", tt.in.Op, got, tt.want)
			}
		})
	}
	for op := OpConst; op <= OpVaEnd; op++ {
		t.Run(fmt.Sprintf("valid-op-%d", op), func(t *testing.T) {
			got := FormatInstr(Instr{Op: op})
			if strings.HasPrefix(got, "InvalidOpcode(") || strings.HasPrefix(got, "Opcode(") {
				t.Fatalf("FormatInstr(%v) did not cover valid opcode: %q", op, got)
			}
		})
	}
}
