package entity

import "shinya.click/cvm/common"

type InitializerType int64

const (
	InitializerTypeExpression    InitializerType = 1
	InitializerTypeStructOrArray InitializerType = 2
)

type Initializer struct {
	Type InitializerType

	Expression      *Expression
	InitializerList []*InitializerItem
	common.SourceRange
}

type InitializerItem struct {
	Designators []*Designator
	Initializer *Initializer
	common.SourceRange
}

type Designator struct {
	Expression *Expression
	Identifier *string
	common.SourceRange
}
