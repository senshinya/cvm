package bytecode

import "fmt"

type Opcode int

const (
	OpConst Opcode = iota
	OpAddrString
	OpAddrGlobal
	OpAddrFunc
	OpLoadConst
	OpDup
	OpPop
	OpSwap
	OpLoadLocal
	OpStoreLocal
	OpAddrLocalObject
	OpAllocDynamicObject
	OpFreeDynamicObject
	OpDynamicObjectAddr
	OpLoad
	OpStore
	OpMemCopy
	OpMemSet
	OpOffset
	OpFieldAddr
	OpBitFieldLoad
	OpBitFieldStore
	OpPtrAdd
	OpPtrDiff
	OpBinary
	OpUnary
	OpCast
	OpLabel
	OpJump
	OpJumpIfZero
	OpJumpIfNonZero
	OpSwitch
	OpReturn
	OpReturnVoid
	OpReturnObject
	OpUnreachable
	OpCall
	OpCallIndirect
	OpVaStart
	OpVaArg
	OpVaEnd
)

func (op Opcode) String() string {
	names := [...]string{
		"OpConst",
		"OpAddrString",
		"OpAddrGlobal",
		"OpAddrFunc",
		"OpLoadConst",
		"OpDup",
		"OpPop",
		"OpSwap",
		"OpLoadLocal",
		"OpStoreLocal",
		"OpAddrLocalObject",
		"OpAllocDynamicObject",
		"OpFreeDynamicObject",
		"OpDynamicObjectAddr",
		"OpLoad",
		"OpStore",
		"OpMemCopy",
		"OpMemSet",
		"OpOffset",
		"OpFieldAddr",
		"OpBitFieldLoad",
		"OpBitFieldStore",
		"OpPtrAdd",
		"OpPtrDiff",
		"OpBinary",
		"OpUnary",
		"OpCast",
		"OpLabel",
		"OpJump",
		"OpJumpIfZero",
		"OpJumpIfNonZero",
		"OpSwitch",
		"OpReturn",
		"OpReturnVoid",
		"OpReturnObject",
		"OpUnreachable",
		"OpCall",
		"OpCallIndirect",
		"OpVaStart",
		"OpVaArg",
		"OpVaEnd",
	}
	if int(op) >= 0 && int(op) < len(names) {
		return names[op]
	}
	return fmt.Sprintf("Opcode(%d)", int(op))
}

type BinaryOp int

const (
	BinAdd BinaryOp = iota
	BinSub
	BinMul
	BinDivS
	BinDivU
	BinRemS
	BinRemU
	BinAnd
	BinOr
	BinXor
	BinShl
	BinShrS
	BinShrU
	BinEq
	BinNe
	BinLtS
	BinLtU
	BinLeS
	BinLeU
	BinGtS
	BinGtU
	BinGeS
	BinGeU
	BinLtF
	BinLeF
	BinGtF
	BinGeF
)

type UnaryOp int

const (
	UnaryNeg UnaryOp = iota
)

type CastOp int

const (
	CastTrunc CastOp = iota
	CastZExt
	CastSExt
	CastFExt
	CastFTrunc
	CastIntToFloat
	CastFloatToInt
	CastPtrToInt
	CastIntToPtr
	CastBit
	CastBool
)

type SwitchCase struct {
	Value int64
	Label int
}

type Instr struct {
	Op       Opcode
	Type     ValueType
	Type2    ValueType
	Int      int64
	Float    float64
	String   string
	Slot     int
	Global   int
	Func     int
	Sig      int
	Layout   int
	Field    int
	Object   int
	Label    int
	Labels   []SwitchCase
	Align    int64
	Size     int64
	Volatile bool
	Binary   BinaryOp
	Unary    UnaryOp
	Cast     CastOp
	Argc     int
}

func Const(t ValueType, v int64) Instr       { return Instr{Op: OpConst, Type: t, Int: v} }
func I32Const(v int32) Instr                 { return Const(TypeI32, int64(v)) }
func I64Const(v int64) Instr                 { return Const(TypeI64, v) }
func U32Const(v uint32) Instr                { return Const(TypeU32, int64(v)) }
func U64Const(v uint64) Instr                { return Instr{Op: OpConst, Type: TypeU64, Int: int64(v)} }
func F32Const(v float32) Instr               { return Instr{Op: OpConst, Type: TypeF32, Float: float64(v)} }
func F64Const(v float64) Instr               { return Instr{Op: OpConst, Type: TypeF64, Float: v} }
func NullPtr() Instr                         { return Instr{Op: OpConst, Type: TypePtr, Int: 0} }
func LoadLocal(t ValueType, slot int) Instr  { return Instr{Op: OpLoadLocal, Type: t, Slot: slot} }
func StoreLocal(t ValueType, slot int) Instr { return Instr{Op: OpStoreLocal, Type: t, Slot: slot} }
func AddrLocalObject(object int) Instr {
	return Instr{Op: OpAddrLocalObject, Object: object, Type: TypeObjectAddr}
}
func AddrGlobal(global int) Instr {
	return Instr{Op: OpAddrGlobal, Global: global, Type: TypeObjectAddr}
}
func AddrString(id int) Instr { return Instr{Op: OpAddrString, Int: int64(id), Type: TypeObjectAddr} }
func Load(t ValueType, align int64, vol bool) Instr {
	return Instr{Op: OpLoad, Type: t, Align: align, Volatile: vol}
}
func Store(t ValueType, align int64, vol bool) Instr {
	return Instr{Op: OpStore, Type: t, Align: align, Volatile: vol}
}
func Binary(t ValueType, op BinaryOp) Instr { return Instr{Op: OpBinary, Type: t, Binary: op} }
func Cast(from, to ValueType, op CastOp) Instr {
	return Instr{Op: OpCast, Type: from, Type2: to, Cast: op}
}
func LabelInstr(id int) Instr { return Instr{Op: OpLabel, Label: id} }
func Jump(label int) Instr    { return Instr{Op: OpJump, Label: label} }
func JumpIfZero(t ValueType, label int) Instr {
	return Instr{Op: OpJumpIfZero, Type: t, Label: label}
}
func JumpIfNonZero(t ValueType, label int) Instr {
	return Instr{Op: OpJumpIfNonZero, Type: t, Label: label}
}
func Return(t ValueType) Instr { return Instr{Op: OpReturn, Type: t} }
func Call(global, sig, argc int) Instr {
	return Instr{Op: OpCall, Global: global, Sig: sig, Argc: argc}
}

func (i Instr) ResultType() (ValueType, bool) {
	switch i.Op {
	case OpConst, OpAddrString, OpAddrGlobal, OpAddrFunc, OpLoadConst, OpLoadLocal, OpAddrLocalObject, OpDynamicObjectAddr, OpLoad, OpBitFieldLoad:
		return i.Type, true
	case OpBinary:
		if isCompare(i.Binary) {
			return TypeBool, true
		}
		return i.Type, true
	case OpCast:
		return i.Type2, true
	}
	return TypeVoid, false
}

func isCompare(op BinaryOp) bool {
	return op >= BinEq
}

func (op BinaryOp) String() string {
	names := [...]string{"add", "sub", "mul", "divs", "divu", "rems", "remu", "and", "or", "xor", "shl", "shrs", "shru", "eq", "ne", "lts", "ltu", "les", "leu", "gts", "gtu", "ges", "geu", "ltf", "lef", "gtf", "gef"}
	if int(op) >= 0 && int(op) < len(names) {
		return names[op]
	}
	return fmt.Sprintf("binop(%d)", int(op))
}
