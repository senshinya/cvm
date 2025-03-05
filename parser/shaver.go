package parser

import (
	"shinya.click/cvm/common"
	"shinya.click/cvm/entity"
)

func shaveForest(forest []*entity.AstNode) ([]*entity.AstNode, error) {
	var (
		latestErr error
		result    []*entity.AstNode
	)
	for _, root := range forest {
		err := NewShaver(root).Shave()
		if err != nil {
			latestErr = err
			continue
		}
		result = append(result, root)
	}
	return result, latestErr
}

type Shaver struct {
	NodeStack   *common.Stack[*entity.AstNode]
	SymbolStack *common.Stack[*ScopeSymbols]
}

type ScopeSymbols struct {
	TypeDefNames       []string
	StructOrUnionNames []string
	EnumNames          []string
	VarNames           []string
}

func NewShaver(root *entity.AstNode) *Shaver {
	s := &Shaver{
		NodeStack:   common.NewStack[*entity.AstNode](),
		SymbolStack: common.NewStack[*ScopeSymbols](),
	}
	s.NodeStack.Push(root)
	s.SymbolStack.Push(&ScopeSymbols{})
	return s
}

func (s *Shaver) Shave() error {
	for !s.NodeStack.IsEmpty() {
		// pre-order
		node, _ := s.NodeStack.Pop()
		switch {
		// TODO
		default:
			for _, child := range node.Children {
				s.NodeStack.Push(child)
			}
		}
	}
	return nil
}
