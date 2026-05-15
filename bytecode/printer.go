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
	fmt.Fprintf(&b, "Module target=%q endian=%s ptr=%d\n", m.Target.Name, m.Target.Endian, m.Target.PointerSize)
	for _, g := range m.Globals {
		printGlobal(&b, g)
	}
	for _, s := range m.Strings {
		fmt.Fprintf(&b, "String #%d value=%q size=%d\n", s.ID, s.Value, len(s.Bytes))
	}
	for _, l := range m.Layouts {
		fmt.Fprintf(&b, "Layout #%d name=%q size=%d align=%d\n", l.ID, l.Name, l.Size, l.Align)
	}
	for _, sig := range m.Sigs {
		fmt.Fprintf(&b, "Sig #%d ret=%s params=(", sig.ID, sig.Ret)
		for i, p := range sig.Params {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString(p.String())
		}
		if sig.Variadic {
			if len(sig.Params) > 0 {
				b.WriteString(", ")
			}
			b.WriteString("...")
		}
		b.WriteString(")\n")
	}
	for _, f := range m.Functions {
		printFunction(&b, f)
	}
	return b.String()
}

func printGlobal(b *strings.Builder, g Global) {
	switch g.Kind {
	case GlobalFunc:
		fmt.Fprintf(b, "Global #%d func name=%q func=%d\n", g.ID, g.Name, g.Func)
	case GlobalExtern:
		fmt.Fprintf(b, "Global #%d extern name=%q\n", g.ID, g.Name)
	default:
		fmt.Fprintf(b, "Global #%d var name=%q size=%d align=%d readonly=%v\n", g.ID, g.Name, g.Size, g.Align, g.Readonly)
	}
}

func printFunction(b *strings.Builder, f Function) {
	fmt.Fprintf(b, "Func #%d global=%d name=%q sig=%d\n", f.ID, f.GlobalID, f.Name, f.Sig)
	for _, l := range f.Locals {
		fmt.Fprintf(b, "  Local #%d name=%q type=%s\n", l.ID, l.Name, l.Type)
	}
	for _, o := range f.Objects {
		fmt.Fprintf(b, "  Object #%d name=%q size=%d align=%d layout=%d\n", o.ID, o.Name, o.Size, o.Align, o.Layout)
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
	case OpLoadLocal:
		return fmt.Sprintf("%sLoadLocal %d", instrTypePrefix(i.Type), i.Slot)
	case OpStoreLocal:
		return fmt.Sprintf("%sStoreLocal %d", instrTypePrefix(i.Type), i.Slot)
	case OpAddrLocalObject:
		return fmt.Sprintf("AddrLocalObject %d", i.Object)
	case OpAddrGlobal:
		return fmt.Sprintf("AddrGlobal %d", i.Global)
	case OpLoad:
		return fmt.Sprintf("%sLoad align=%d volatile=%v", instrTypePrefix(i.Type), i.Align, i.Volatile)
	case OpStore:
		return fmt.Sprintf("%sStore align=%d volatile=%v", instrTypePrefix(i.Type), i.Align, i.Volatile)
	case OpBinary:
		return fmt.Sprintf("%s%s", instrTypePrefix(i.Type), binaryName(i.Binary))
	case OpJump:
		return fmt.Sprintf("Jump L%d", i.Label)
	case OpJumpIfZero:
		return fmt.Sprintf("JumpIfZero %s L%d", i.Type, i.Label)
	case OpJumpIfNonZero:
		return fmt.Sprintf("JumpIfNonZero %s L%d", i.Type, i.Label)
	case OpLabel:
		return fmt.Sprintf("L%d:", i.Label)
	case OpReturn:
		return fmt.Sprintf("%sReturn", instrTypePrefix(i.Type))
	case OpReturnVoid:
		return "ReturnVoid"
	case OpCall:
		return fmt.Sprintf("Call global=%d sig=%d argc=%d", i.Global, i.Sig, i.Argc)
	default:
		return fmt.Sprintf("Opcode(%d)", int(i.Op))
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
	case BinEq:
		return "Eq"
	case BinNe:
		return "Ne"
	case BinLtS:
		return "LtS"
	case BinLtU:
		return "LtU"
	default:
		return titleASCII(op.String())
	}
}

func titleASCII(s string) string {
	if s == "" {
		return ""
	}
	if s[0] >= 'a' && s[0] <= 'z' {
		return string(s[0]-'a'+'A') + s[1:]
	}
	return s
}
