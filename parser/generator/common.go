package main

import (
	"fmt"
	mapset "github.com/deckarep/golang-set/v2"
	"shinya.click/cvm/entity"
)

type Production struct {
	Left  string
	Index int64
	Right []string
}

type Productions struct {
	Productions []*Production
	LeftMap     map[string][]*Production
}

func (p Productions) first(symbol string) []string {
	if entity.IsTerminalSymbol(symbol) {
		return []string{symbol}
	}
	result := mapset.NewSet[string]()
	defProds := p.LeftMap[symbol]
	for _, prod := range defProds {
		right := prod.Right[0]
		if right == symbol {
			continue
		}
		result.Append(p.first(right)...)
		// nothing can derive to epsilon
	}
	return result.ToSlice()
}

func NewProductions(prods []*Production) *Productions {
	leftMap := map[string][]*Production{}
	for _, prod := range prods {
		leftMap[prod.Left] = append(leftMap[prod.Left], prod)
	}
	return &Productions{Productions: prods, LeftMap: leftMap}
}

type LRItem struct {
	Production *Production
	DotIndex   int
	LookAhead  string
}

func (i LRItem) String() string {
	return fmt.Sprintf("%s := %v, %d, %s", i.Production.Left, i.Production.Right, i.DotIndex, i.LookAhead)
}

func (i LRItem) isReducible() bool {
	return i.DotIndex == len(i.Production.Right)
}

func (i LRItem) getSymbolAfterDot() string {
	return i.Production.Right[i.DotIndex]
}

type LRNode struct {
	Kernel []LRItem             // kernel items
	Items  []LRItem             // all items
	Edges  map[string][]*LRNode // from edges
}

func (n LRNode) hasSameKernel(kernel []LRItem) bool {
	nKernel := n.Kernel
	if len(nKernel) != len(kernel) {
		return false
	}
	nMap := map[LRItem]struct{}{}
	for _, item := range nKernel {
		nMap[item] = struct{}{}
	}
	for _, item := range kernel {
		if _, ok := nMap[item]; !ok {
			return false
		}
	}
	return true
}

type LRDFA struct {
	StartState *LRNode
	EndStates  []*LRNode
	AllNodes   []*LRNode
}

type StateSpecificLRItem struct {
	Node *LRNode
	Item LRItem
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
