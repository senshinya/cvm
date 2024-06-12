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
	MetaTypeUserDefined
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
	StructMetaInfo   *StructUnionMetaInfo
	UnionMetaInfo    *StructUnionMetaInfo
	FunctionMetaInfo *FunctionMetaInfo
	ArrayMetaInfo    *ArrayMetaInfo
	UserDefinedType  *Type
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
	Identifier string

	Incomplete    bool
	FieldMetaInfo []*FieldMetaInfo
}

type FieldMetaInfo struct {
	Type       Type
	Identifier *string
	BitWidth   *SingleExpression
}

type ArrayMetaInfo struct {
	InnerType *Type

	// should only be in array declare inside function prototype
	Static bool
	TypeQualifiers
	Asterisk bool

	Incomplete bool
	Size       *SingleExpression
}

type FunctionParameter struct {
	Specifiers Specifiers
	Type       Type
	Identifier *string
}

type FunctionMetaInfo struct {
	ReturnType     *Type
	Parameters     []*FunctionParameter
	Variadic       bool
	IdentifierList []string
}
