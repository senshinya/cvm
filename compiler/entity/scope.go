package entity

import "shinya.click/cvm/common"

type ScopeType int8

const (
	ScopeTypeFile ScopeType = iota + 1
	ScopeTypeFunction
	ScopeTypeBlock
	ScopeTypeFunctionPrototype
)

type Scope struct {
	Type ScopeType

	ParentScope *Scope

	SymbolTable

	Children common.Stack[*Scope]
}

type SymbolTable struct {
	Objects map[string]*Object
	Tags    map[string]*Tag
}

func newSymbolTable() SymbolTable {
	return SymbolTable{Objects: map[string]*Object{}, Tags: map[string]*Tag{}}
}

func (s *Scope) LookForObject(symbol string) *Object {
	current := s
	for current != nil {
		if _, ok := current.Objects[symbol]; ok {
			return current.Objects[symbol]
		}
		current = current.ParentScope
	}
	return nil
}

func (s *Scope) LookForTag(symbol string) *Tag {
	current := s
	for current != nil {
		if _, ok := current.Tags[symbol]; ok {
			return current.Tags[symbol]
		}
		current = current.ParentScope
	}
	return nil
}

func (s *Scope) AddChildScope(ChildScope *Scope) {
	ChildScope.ParentScope = s
	s.Children.Push(ChildScope)
}

func (s *Scope) GetType() ScopeType {
	return s.Type
}

func NewFileScope() *Scope {
	return &Scope{Type: ScopeTypeFile, SymbolTable: newSymbolTable()}
}

func NewFunctionScope() *Scope {
	return &Scope{Type: ScopeTypeFunction, SymbolTable: newSymbolTable()}
}

func NewBlockScope() *Scope {
	return &Scope{Type: ScopeTypeBlock, SymbolTable: newSymbolTable()}
}

func NewFunctionPrototypeScope() *Scope {
	return &Scope{Type: ScopeTypeFunctionPrototype, SymbolTable: newSymbolTable()}
}
