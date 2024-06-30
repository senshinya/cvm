package glr

import (
	"fmt"
	"github.com/thoas/go-funk"
	"shinya.click/cvm/common"
	"strings"
)

type RawAstNode struct {
	Parent     *RawAstNode
	Typ        common.TokenType
	Terminal   *common.Token
	Production Production
	Children   []*RawAstNode
	common.SourceRange
}

func (n *RawAstNode) SetChildren(children []*RawAstNode) {
	n.Children = children
	for _, child := range children {
		child.Parent = n
	}
	// set the source range
	startSym := children[0]
	endSym := children[len(children)-1]
	n.SourceRange = common.SourceRange{
		SourceStart: startSym.SourceRange.SourceStart,
		SourceEnd:   endSym.SourceRange.SourceEnd,
	}
}

func (n *RawAstNode) AssertNonTerminal(nonTerminal ...common.TokenType) error {
	for _, t := range nonTerminal {
		if t == n.Typ {
			return nil
		}
	}
	return fmt.Errorf("unexpected %s", n.Typ)
}

func (n *RawAstNode) ReducedBy(nonTerminal common.TokenType, idx ...int64) bool {
	return n.Production.Left == nonTerminal && funk.ContainsInt64(idx, n.Production.Index)
}

type Production struct {
	Left  common.TokenType
	Index int64
	Right []common.TokenType
}

func (p Production) String() string {
	var rights []string
	for _, t := range p.Right {
		rights = append(rights, string(t))
	}
	return fmt.Sprintf("%s -> %s", p.Left, strings.Join(rights, " "))
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
