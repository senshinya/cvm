package parser

import "shinya.click/cvm/common"

type AstNode struct {
	Parent    *AstNode
	Typ       common.TokenType
	Terminal  *common.Token
	ProdIndex int // reduce by which production
	Children  []*AstNode
}

func (n *AstNode) SetChildren(children []*AstNode) {
	n.Children = children
	for _, child := range children {
		child.Parent = n
	}
}

type Production struct {
	Left  common.TokenType
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
