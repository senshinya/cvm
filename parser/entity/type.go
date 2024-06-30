package entity

import "shinya.click/cvm/common"

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
	MetaTypeUserDefined
)

type TypeQualifiers struct {
	Const    bool
	Restrict bool
	Volatile bool

	common.SourceRange
}

type Type struct {
	TypeQualifiers TypeQualifiers

	MetaType MetaType

	NumberMetaInfo      *NumberMetaInfo
	PointerInnerType    *Type
	StructMetaInfo      *StructUnionMetaInfo
	UnionMetaInfo       *StructUnionMetaInfo
	EnumMetaInfo        *EnumMetaInfo
	FunctionMetaInfo    *FunctionMetaInfo
	ArrayMetaInfo       *ArrayMetaInfo
	UserDefinedTypeName *string

	common.SourceRange
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

type StructUnionMetaInfo struct {
	Identifier *common.Token

	Incomplete    bool
	FieldMetaInfo []*FieldMetaInfo
}

type FieldMetaInfo struct {
	Type       Type
	Identifier *common.Token
	BitWidth   *Expression
	common.SourceRange
}

type EnumMetaInfo struct {
	Identifier *common.Token

	Incomplete bool
	EnumFields []*EnumFieldMetaInfo
}

type EnumFieldMetaInfo struct {
	Identifier *common.Token

	Value *Expression
	common.SourceRange
}

type ArrayMetaInfo struct {
	InnerType *Type

	// should only be in array declare inside function prototype
	Static bool
	TypeQualifiers
	Asterisk bool

	Incomplete bool
	Size       *Expression
}

type FunctionParameter struct {
	Specifiers Specifiers
	Type       Type
	Identifier *common.Token

	common.SourceRange
}

type FunctionMetaInfo struct {
	ReturnType     *Type
	Parameters     []*FunctionParameter
	Variadic       bool
	IdentifierList []*common.Token
}
