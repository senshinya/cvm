package bytecode

import (
	"fmt"
	"strings"
)

func PrintModule(m *Module) string {
	if m == nil {
		return "<nil module>\n"
	}
	var b strings.Builder
	t := m.Target
	fmt.Fprintf(&b, "Module target=%q endian=%s ptr_size=%d ptr_align=%d bool_size=%d bool_align=%d bitfield_policy=%q layout_version=%q\n",
		t.Name, t.Endian, t.PointerSize, t.PointerAlign, t.BoolSize, t.BoolAlign, t.BitFieldPolicy, t.LayoutVersion)
	for _, g := range m.Globals {
		printGlobal(&b, m, g)
	}
	for _, s := range m.Strings {
		fmt.Fprintf(&b, "String #%d value=%q bytes=%d hex=%x\n", s.ID, s.Value, len(s.Bytes), s.Bytes)
	}
	for _, l := range m.Layouts {
		fmt.Fprintf(&b, "Layout #%d name=%q size=%d align=%d elem_size=%d\n", l.ID, l.Name, l.Size, l.Align, l.ElemSize)
		for _, f := range l.Fields {
			fmt.Fprintf(&b, "  Field #%d name=%q offset=%d type=%s\n", f.ID, f.Name, f.Offset, f.Type)
		}
		for _, bf := range l.Bit {
			fmt.Fprintf(&b, "  BitField #%d name=%q container=%s byte_offset=%d bit_offset=%d width=%d signed=%v volatile=%v policy=%q\n",
				bf.ID, bf.Name, bf.Container, bf.ByteOffset, bf.BitOffset, bf.Width, bf.Signed, bf.Volatile, bf.LayoutPolicy)
		}
	}
	for _, sig := range m.Sigs {
		fmt.Fprintf(&b, "Sig #%d ret=%s params=(", sig.ID, sig.Ret)
		writeValueTypes(&b, sig.Params)
		if sig.Variadic {
			if len(sig.Params) > 0 {
				b.WriteString(", ")
			}
			b.WriteString("...")
		}
		b.WriteString(")")
		if sig.Variadic {
			b.WriteString(" variadic")
		}
		b.WriteString("\n")
	}
	for _, f := range m.Functions {
		printFunction(&b, f)
	}
	return b.String()
}

func printGlobal(b *strings.Builder, m *Module, g Global) {
	switch g.Kind {
	case GlobalFunc:
		fmt.Fprintf(b, "Global #%d func name=%q func=%d\n", g.ID, g.Name, g.Func)
	case GlobalExtern:
		fmt.Fprintf(b, "Global #%d extern name=%q\n", g.ID, g.Name)
	default:
		fmt.Fprintf(b, "Global #%d var name=%q size=%d align=%d readonly=%v init_zero=%d init_bytes=%d init_relocs=%d\n",
			g.ID, g.Name, g.Size, g.Align, g.Readonly, g.Init.ZeroFill, len(g.Init.Bytes), len(g.Init.Relocations))
		if len(g.Init.Bytes) > 0 {
			fmt.Fprintf(b, "  InitBytes hex=%x\n", g.Init.Bytes)
		}
		for _, r := range g.Init.Relocations {
			fmt.Fprintf(b, "  reloc offset=%d kind=%s target=%s addend=%d\n", r.Offset, relocationKindName(r.Kind), relocationTarget(m, r), r.Addend)
		}
	}
}

func printFunction(b *strings.Builder, f Function) {
	fmt.Fprintf(b, "Func #%d global=%d name=%q sig=%d max_stack=%d\n", f.ID, f.GlobalID, f.Name, f.Sig, f.MaxStack)
	for _, p := range f.Params {
		fmt.Fprintf(b, "  Param slot=%d name=%q type=%s\n", p.Slot, p.Name, p.Type)
	}
	for _, l := range f.Locals {
		fmt.Fprintf(b, "  Local #%d name=%q type=%s\n", l.ID, l.Name, l.Type)
	}
	for _, o := range f.Objects {
		fmt.Fprintf(b, "  Object #%d name=%q size=%d align=%d layout=%d\n", o.ID, o.Name, o.Size, o.Align, o.Layout)
	}
	for _, o := range f.DynamicObjects {
		fmt.Fprintf(b, "  DynamicObject #%d name=%q align=%d layout=%d\n", o.ID, o.Name, o.Align, o.Layout)
	}
	for _, l := range f.Labels {
		fmt.Fprintf(b, "  Label #%d name=%q stack=(", l.ID, l.Name)
		writeValueTypes(b, l.Stack)
		fmt.Fprintf(b, ") statement=%v\n", l.Statement)
	}
	for pc, ins := range f.Instrs {
		fmt.Fprintf(b, "  %04d: %s\n", pc, FormatInstr(ins))
	}
}

func FormatInstr(i Instr) string {
	switch i.Op {
	case OpConst:
		if i.Type == TypeF32 || i.Type == TypeF64 || i.Type == TypeFLong {
			return fmt.Sprintf("%sConst %v", instrTypePrefix(i.Type), i.Float)
		}
		return fmt.Sprintf("%sConst %d", instrTypePrefix(i.Type), i.Int)
	case OpAddrString:
		return fmt.Sprintf("AddrString %d", i.Int)
	case OpAddrGlobal:
		return fmt.Sprintf("AddrGlobal %d", i.Global)
	case OpAddrFunc:
		return fmt.Sprintf("AddrFunc %d", i.Global)
	case OpLoadConst:
		return fmt.Sprintf("%sLoadConst global=%d offset=%d", instrTypePrefix(i.Type), i.Global, i.Int)
	case OpDup:
		return "Dup"
	case OpPop:
		return "Pop"
	case OpSwap:
		return "Swap"
	case OpLoadLocal:
		return fmt.Sprintf("%sLoadLocal %d", instrTypePrefix(i.Type), i.Slot)
	case OpStoreLocal:
		return fmt.Sprintf("%sStoreLocal %d", instrTypePrefix(i.Type), i.Slot)
	case OpAddrLocalObject:
		return fmt.Sprintf("AddrLocalObject %d", i.Object)
	case OpAllocDynamicObject:
		return fmt.Sprintf("AllocDynamicObject object=%d size=<stack:i64> align=%d layout=%d", i.Object, i.Align, i.Layout)
	case OpFreeDynamicObject:
		return fmt.Sprintf("FreeDynamicObject %d", i.Object)
	case OpDynamicObjectAddr:
		return fmt.Sprintf("DynamicObjectAddr %d", i.Object)
	case OpLoad:
		return fmt.Sprintf("%sLoad align=%d volatile=%v", instrTypePrefix(i.Type), i.Align, i.Volatile)
	case OpStore:
		return fmt.Sprintf("%sStore align=%d volatile=%v", instrTypePrefix(i.Type), i.Align, i.Volatile)
	case OpMemCopy:
		return fmt.Sprintf("MemCopy size=%d align=%d volatile=%v", i.Size, i.Align, i.Volatile)
	case OpMemSet:
		return fmt.Sprintf("MemSet size=%d align=%d volatile=%v", i.Size, i.Align, i.Volatile)
	case OpOffset:
		return fmt.Sprintf("%sOffset %d", instrTypePrefix(i.Type), i.Int)
	case OpFieldAddr:
		return fmt.Sprintf("FieldAddr layout=%d field=%d", i.Layout, i.Field)
	case OpBitFieldLoad:
		return fmt.Sprintf("%sBitFieldLoad layout=%d field=%d volatile=%v", instrTypePrefix(i.Type), i.Layout, i.Field, i.Volatile)
	case OpBitFieldStore:
		return fmt.Sprintf("%sBitFieldStore layout=%d field=%d volatile=%v", instrTypePrefix(i.Type), i.Layout, i.Field, i.Volatile)
	case OpPtrAdd:
		return fmt.Sprintf("PtrAdd elem_size=%d", i.Size)
	case OpPtrAddDynamic:
		return "PtrAddDynamic"
	case OpPtrDiff:
		return fmt.Sprintf("PtrDiff elem_size=%d", i.Size)
	case OpBinary:
		return fmt.Sprintf("%s%s", instrTypePrefix(i.Type), binaryName(i.Binary))
	case OpUnary:
		return fmt.Sprintf("%s%s", instrTypePrefix(i.Type), unaryName(i.Unary))
	case OpCast:
		return fmt.Sprintf("Cast %s->%s %s", i.Type, i.Type2, castName(i.Cast))
	case OpJump:
		return fmt.Sprintf("Jump L%d", i.Label)
	case OpJumpIfZero:
		return fmt.Sprintf("JumpIfZero %s L%d", i.Type, i.Label)
	case OpJumpIfNonZero:
		return fmt.Sprintf("JumpIfNonZero %s L%d", i.Type, i.Label)
	case OpLabel:
		return fmt.Sprintf("L%d:", i.Label)
	case OpSwitch:
		return fmt.Sprintf("Switch %s default=L%d cases=(%s)", i.Type, i.Label, switchCases(i.Labels))
	case OpReturn:
		return fmt.Sprintf("%sReturn", instrTypePrefix(i.Type))
	case OpReturnVoid:
		return "ReturnVoid"
	case OpReturnObject:
		return fmt.Sprintf("ReturnObject %d", i.Object)
	case OpUnreachable:
		return "Unreachable"
	case OpCall:
		return fmt.Sprintf("Call global=%d sig=%d argc=%d", i.Global, i.Sig, i.Argc)
	case OpCallIndirect:
		return fmt.Sprintf("CallIndirect sig=%d argc=%d", i.Sig, i.Argc)
	case OpVaStart:
		return fmt.Sprintf("VaStart slot=%d", i.Slot)
	case OpVaArg:
		return fmt.Sprintf("%sVaArg", instrTypePrefix(i.Type))
	case OpVaEnd:
		return fmt.Sprintf("VaEnd slot=%d", i.Slot)
	default:
		return fmt.Sprintf("InvalidOpcode(%d)", int(i.Op))
	}
}

func instrTypePrefix(t ValueType) string {
	switch t {
	case TypeBool:
		return "Bool"
	case TypeI8:
		return "I8"
	case TypeI16:
		return "I16"
	case TypeI32:
		return "I32"
	case TypeI64:
		return "I64"
	case TypeU8:
		return "U8"
	case TypeU16:
		return "U16"
	case TypeU32:
		return "U32"
	case TypeU64:
		return "U64"
	case TypeF32:
		return "F32"
	case TypeF64:
		return "F64"
	case TypeFLong:
		return "FLong"
	case TypePtr:
		return "Ptr"
	case TypeObjectAddr:
		return "ObjectAddr"
	default:
		return t.String()
	}
}

func binaryName(op BinaryOp) string {
	switch op {
	case BinAdd:
		return "Add"
	case BinSub:
		return "Sub"
	case BinMul:
		return "Mul"
	case BinDivS:
		return "DivS"
	case BinDivU:
		return "DivU"
	case BinRemS:
		return "RemS"
	case BinRemU:
		return "RemU"
	case BinAnd:
		return "And"
	case BinOr:
		return "Or"
	case BinXor:
		return "Xor"
	case BinShl:
		return "Shl"
	case BinShrS:
		return "ShrS"
	case BinShrU:
		return "ShrU"
	case BinEq:
		return "Eq"
	case BinNe:
		return "Ne"
	case BinLtS:
		return "LtS"
	case BinLtU:
		return "LtU"
	case BinLeS:
		return "LeS"
	case BinLeU:
		return "LeU"
	case BinGtS:
		return "GtS"
	case BinGtU:
		return "GtU"
	case BinGeS:
		return "GeS"
	case BinGeU:
		return "GeU"
	case BinLtF:
		return "LtF"
	case BinLeF:
		return "LeF"
	case BinGtF:
		return "GtF"
	case BinGeF:
		return "GeF"
	default:
		return fmt.Sprintf("BinaryOp(%d)", int(op))
	}
}

func unaryName(op UnaryOp) string {
	switch op {
	case UnaryNeg:
		return "Neg"
	default:
		return fmt.Sprintf("UnaryOp(%d)", int(op))
	}
}

func castName(op CastOp) string {
	switch op {
	case CastTrunc:
		return "Trunc"
	case CastZExt:
		return "ZExt"
	case CastSExt:
		return "SExt"
	case CastFExt:
		return "FExt"
	case CastFTrunc:
		return "FTrunc"
	case CastIntToFloat:
		return "IntToFloat"
	case CastFloatToInt:
		return "FloatToInt"
	case CastPtrToInt:
		return "PtrToInt"
	case CastIntToPtr:
		return "IntToPtr"
	case CastBit:
		return "Bit"
	case CastBool:
		return "Bool"
	default:
		return fmt.Sprintf("CastOp(%d)", int(op))
	}
}

func writeValueTypes(b *strings.Builder, ts []ValueType) {
	for i, t := range ts {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(t.String())
	}
}

func relocationKindName(k RelocationKind) string {
	switch k {
	case RelocGlobal:
		return "global"
	case RelocFunc:
		return "func"
	case RelocString:
		return "string"
	default:
		return fmt.Sprintf("reloc(%d)", int(k))
	}
}

func relocationTarget(m *Module, r Relocation) string {
	switch r.Kind {
	case RelocGlobal:
		if r.Target >= 0 && r.Target < len(m.Globals) {
			g := m.Globals[r.Target]
			return fmt.Sprintf("global#%d(%q)", g.ID, g.Name)
		}
		return fmt.Sprintf("global#%d(<invalid>)", r.Target)
	case RelocFunc:
		if r.Target >= 0 && r.Target < len(m.Functions) {
			f := m.Functions[r.Target]
			return fmt.Sprintf("func#%d(%q)", f.ID, f.Name)
		}
		return fmt.Sprintf("func#%d(<invalid>)", r.Target)
	case RelocString:
		if r.Target >= 0 && r.Target < len(m.Strings) {
			s := m.Strings[r.Target]
			return fmt.Sprintf("string#%d(%q)", s.ID, s.Value)
		}
		return fmt.Sprintf("string#%d(<invalid>)", r.Target)
	default:
		return fmt.Sprintf("target#%d(<unknown-kind>)", r.Target)
	}
}

func switchCases(cases []SwitchCase) string {
	var b strings.Builder
	for i, c := range cases {
		if i > 0 {
			b.WriteString(", ")
		}
		fmt.Fprintf(&b, "%d:L%d", c.Value, c.Label)
	}
	return b.String()
}
