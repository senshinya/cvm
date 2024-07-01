package entity

import "shinya.click/cvm/parser/entity"

type ObjectType int8

const (
	ObjectTypeFunction ObjectType = iota + 1
	ObjectTypeStructMember
	ObjectTypeUnionMember
	ObjectTypeEnumConstant
	ObjectTypeTypedefName
	ObjectTypeLabel
	ObjectTypeMacroName      // no macro yet
	ObjectTypeMacroParameter // no macro yet
)

type Object struct {
	Type ObjectType

	FunctionDefinition *entity.FunctionDefinition
	StructMemberType   *entity.FieldMetaInfo
	UnionMemberType    *entity.FieldMetaInfo
	EnumConstantValue  *int64
	TypedefType        *entity.Type
	LabeledStatement   *entity.Statement
}

type TagType int8

const (
	EnumTag TagType = iota + 1
	StructTag
)

type Tag struct {
	Type TagType
}
