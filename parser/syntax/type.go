package syntax

type MetaType uint8

const (
	MetaTypeUnknown MetaType = iota
	MetaTypeVoid
	MetaTypeNumber
	MetaTypeEnum
	MetaTypePointer
	MetaTypeStruct
	MetaTypeUnion
	MetaTypeFunction
	MetaTypeArray
)

type TypeQualifiers struct {
	Const    bool
	Restrict bool
	Volatile bool
}

type Type struct {
	TypeQualifiers TypeQualifiers

	MetaType MetaType

	NumberMetaInfo   *NumberMetaInfo
	PointerInnerType *Type
	StructMetaInfo   *StructMetaInfo
	UnionMetaInfo    *UnionMetaInfo
	FunctionMetaInfo *FunctionMetaInfo
	ArrayMetaInfo    *ArrayMetaInfo
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

type ArrayMetaInfo struct {
	InnerType *Type
}

type FunctionMetaInfo struct {
	ReturnType *Type
}
