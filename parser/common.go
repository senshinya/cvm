package parser

import "shinya.click/cvm/common"

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
