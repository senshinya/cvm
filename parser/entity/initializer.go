package entity

type InitializerType int64

const (
	InitializerTypeExpression    InitializerType = 1
	InitializerTypeStructOrArray InitializerType = 2
)

type Initializer struct {
	Type InitializerType

	Expression      *SingleExpression
	InitializerList []*InitializerItem
}

type InitializerItem struct {
	Designators []*Designator
	Initializer *Initializer
}

type Designator struct {
	Expression *SingleExpression
	Identifier *string
}
