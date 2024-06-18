package entity

import "shinya.click/cvm/common"

type AstNode struct {
	Parent     *AstNode
	Typ        common.TokenType
	Terminal   *common.Token
	Production Production
	Children   []*AstNode
}

func (n *AstNode) SetChildren(children []*AstNode) {
	n.Children = children
	for _, child := range children {
		child.Parent = n
	}
}

func (n *AstNode) AssertNonTerminal(nonTerminal common.TokenType) bool {
	return n.Production.Left == nonTerminal
}

func (n *AstNode) ReducedBy(nonTerminal common.TokenType, idx int64) bool {
	return n.AssertNonTerminal(nonTerminal) && n.Production.Index == idx
}

type Production struct {
	Left  common.TokenType
	Index int64
	Right []common.TokenType
}

type OperatorType uint

const (
	SHIFT  OperatorType = 1
	REDUCE OperatorType = 2
	ACC    OperatorType = 3
)

type DFAOperator struct {
	OperatorType OperatorType
	StateIndex   int
	ReduceIndex  int
}
