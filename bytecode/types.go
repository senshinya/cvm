package bytecode

import "fmt"

type ValueType int

const (
	TypeVoid ValueType = iota
	TypeBool
	TypeI8
	TypeI16
	TypeI32
	TypeI64
	TypeU8
	TypeU16
	TypeU32
	TypeU64
	TypeF32
	TypeF64
	TypeFLong
	TypePtr
	TypeObjectAddr
)

func (t ValueType) String() string {
	switch t {
	case TypeVoid:
		return "void"
	case TypeBool:
		return "bool"
	case TypeI8:
		return "i8"
	case TypeI16:
		return "i16"
	case TypeI32:
		return "i32"
	case TypeI64:
		return "i64"
	case TypeU8:
		return "u8"
	case TypeU16:
		return "u16"
	case TypeU32:
		return "u32"
	case TypeU64:
		return "u64"
	case TypeF32:
		return "f32"
	case TypeF64:
		return "f64"
	case TypeFLong:
		return "flong"
	case TypePtr:
		return "ptr"
	case TypeObjectAddr:
		return "objectaddr"
	default:
		return fmt.Sprintf("type(%d)", int(t))
	}
}

type TargetInfo struct {
	Name           string
	Endian         string
	PointerSize    int64
	PointerAlign   int64
	BoolSize       int64
	BoolAlign      int64
	BitFieldPolicy string
	LayoutVersion  string
}

const (
	CurrentModuleVersion = "1"
	DefaultExternABI     = "c"
	NoEntryGlobal        = -1
	NoFuncSig            = -1
)

func DefaultTarget() TargetInfo {
	return TargetInfo{
		Name:           "cvm-default",
		Endian:         "little",
		PointerSize:    8,
		PointerAlign:   8,
		BoolSize:       1,
		BoolAlign:      1,
		BitFieldPolicy: "cvm",
		LayoutVersion:  "1",
	}
}

func NewModule() *Module {
	return &Module{
		Version: CurrentModuleVersion,
		Entry:   &EntryPoint{Global: NoEntryGlobal},
		Target:  DefaultTarget(),
	}
}

type Module struct {
	Version   string
	Entry     *EntryPoint
	Target    TargetInfo
	Globals   []Global
	Functions []Function
	Strings   []StringConst
	Layouts   []ObjectLayout
	Sigs      []FuncSig
}

type EntryPoint struct {
	Global int
	Name   string
}

type FuncSig struct {
	ID       int
	Ret      ValueType
	Params   []ValueType
	Variadic bool
}

type GlobalKind int

const (
	GlobalVar GlobalKind = iota
	GlobalFunc
	GlobalExtern
)

type Global struct {
	ID       int
	Name     string
	Kind     GlobalKind
	Func     int
	Sig      int
	Extern   ExternRef
	Size     int64
	Align    int64
	Readonly bool
	Init     InitData
}

type ExternRef struct {
	Module string
	Name   string
	ABI    string
}

type InitData struct {
	ZeroFill    int64
	Bytes       []byte
	Relocations []Relocation
}

type RelocationKind int

const (
	RelocGlobal RelocationKind = iota
	RelocFunc
	RelocString
)

type Relocation struct {
	Offset int64
	Kind   RelocationKind
	Target int
	Addend int64
}

type StringConst struct {
	ID    int
	Value string
	Bytes []byte
}

type ObjectLayout struct {
	ID       int
	Name     string
	Size     int64
	Align    int64
	Fields   []FieldLayout
	Bit      []BitFieldLayout
	ElemSize int64
}

type FieldLayout struct {
	ID     int
	Name   string
	Offset int64
	Type   ValueType
}

type BitFieldLayout struct {
	ID           int
	Name         string
	Container    ValueType
	ByteOffset   int64
	BitOffset    int
	Width        int
	Signed       bool
	Volatile     bool
	LayoutPolicy string
}

type Function struct {
	ID             int
	GlobalID       int
	Name           string
	Sig            int
	Params         []Param
	Locals         []LocalSlot
	Objects        []LocalObject
	DynamicObjects []DynamicObject
	Labels         []Label
	Instrs         []Instr
	MaxStack       int
}

type Param struct {
	Name string
	Type ValueType
	Slot int
}

type LocalSlot struct {
	ID   int
	Name string
	Type ValueType
}

type LocalObject struct {
	ID     int
	Name   string
	Size   int64
	Align  int64
	Layout int
}

type DynamicObject struct {
	ID     int
	Name   string
	Align  int64
	Layout int
}

type Label struct {
	ID        int
	Name      string
	Stack     []ValueType
	Statement bool
}
