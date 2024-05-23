package syntax

type MetaType uint8

const (
	MetaTypeUnknown MetaType = iota
	MetaTypeVoid
	MetaTypeNumber
	MetaTypePointer
	MetaTypeStruct
	MetaTypeUnion
	MetaTypeFunction
)

type Type struct {
	TypeQualifiers TypeQualifiers

	MetaType MetaType

	NumberMetaInfo   *NumberMetaInfo
	PointerInnerType *Type
	StructMetaInfo   *StructMetaInfo
	UnionMetaInfo    *UnionMetaInfo
	FunctionMetaInfo *FunctionMetaInfo
}

type BaseNumType uint8

const (
	BaseNumTypeUnknown BaseNumType = iota
	BaseNumTypeChar
	BaseNumTypeShort
	BaseNumTypeInt
	BaseNumTypeLong
	BaseNumTypeFloat
	BaseNumTypeDouble
	BaseNumTypeBool
	BaseNumTypeLongLong
	BaseNumTypeLongDouble
)

type NumberMetaInfo struct {
	Unsigned    bool
	Signed      bool
	BaseNumType BaseNumType
}

type StructMetaInfo struct {
}

type UnionMetaInfo struct {
}

type FunctionMetaInfo struct {
	ReturnType *Type
}
