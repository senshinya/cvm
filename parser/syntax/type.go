package syntax

type TypeDeclaration struct {
	TypeQualifiers TypeQualifiers

	CoreType Type
}

type TypeQualifiers struct {
	Const    bool
	Restrict bool
	Volatile bool
}

type MetaType uint8

const (
	MetaTypeVoid MetaType = iota
	MetaTypeNumber
	MetaTypePointer
	MetaTypeStruct
	MetaTypeUnion
	MetaTypeFunction
)

type Type struct {
	TypeQualifiers TypeQualifiers

	MetaType MetaType

	PointerInnerType *Type
	StructMetaInfo   *StructMetaInfo
	UnionMetaInfo    *UnionMetaInfo
	FunctionMetaInfo *FunctionMetaInfo
}

type StructMetaInfo struct {
}

type UnionMetaInfo struct {
}

type FunctionMetaInfo struct {
	ReturnType Type
}
