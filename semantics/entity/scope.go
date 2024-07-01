package entity

type ScopeType int8

const (
	FunctionScope ScopeType = iota + 1
	FileScope
	BlockScope
	FunctionPrototypeScope
)

type Scope struct {
	Type ScopeType

	Identifiers map[string]Object
	Tags        map[string]Tag
}
