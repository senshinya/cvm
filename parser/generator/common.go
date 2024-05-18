package main

import "fmt"

type Production struct {
	Left  string
	Right []string
}

type Productions struct {
	Productions []*Production
	LeftMap     map[string][]*Production
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
	Kernel []LRItem           // kernel items
	Items  []LRItem           // all items
	Edges  map[string]*LRNode // from edges
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

var terminals = map[string]struct{}{
	"IDENTIFIER":        {},
	"STRING":            {},
	"CHARACTER":         {},
	"INTEGER_CONSTANT":  {},
	"FLOATING_CONSTANT": {},
	"AUTO":              {},
	"BREAK":             {},
	"CASE":              {},
	"CHAR":              {},
	"CONST":             {},
	"CONTINUE":          {},
	"DEFAULT":           {},
	"DO":                {},
	"DOUBLE":            {},
	"ELSE":              {},
	"ENUM":              {},
	"EXTERN":            {},
	"FLOAT":             {},
	"FOR":               {},
	"GOTO":              {},
	"IF":                {},
	"INLINE":            {},
	"INT":               {},
	"LONG":              {},
	"REGISTER":          {},
	"RESTRICT":          {},
	"RETURN":            {},
	"SHORT":             {},
	"SIGNED":            {},
	"SIZEOF":            {},
	"STATIC":            {},
	"STRUCT":            {},
	"SWITCH":            {},
	"TYPEDEF":           {},
	"UNION":             {},
	"UNSIGNED":          {},
	"VOID":              {},
	"VOLATILE":          {},
	"WHILE":             {},
	"BOOL":              {},
	"COMPLEX":           {},
	"IMAGINARY":         {},
	"LEFT_BRACKETS":     {},
	"RIGHT_BRACKETS":    {},
	"LEFT_PARENTHESES":  {},
	"RIGHT_PARENTHESES": {},
	"LEFT_BRACES":       {},
	"RIGHT_BRACES":      {},
	"PERIOD":            {},
	"ARROW":             {},
	"PLUS_PLUS":         {},
	"MINUS_MINUS":       {},
	"AND":               {},
	"ASTERISK":          {},
	"PLUS":              {},
	"MINUS":             {},
	"TILDE":             {},
	"EXCLAMATION":       {},
	"SLASH":             {},
	"PERCENT":           {},
	"LEFT_SHIFT":        {},
	"RIGHT_SHIFT":       {},
	"LESS":              {},
	"GREATER":           {},
	"LESS_EQUAL":        {},
	"GREATER_EQUAL":     {},
	"EQUAL_EQUAL":       {},
	"NOT_EQUAL":         {},
	"XOR":               {},
	"OR":                {},
	"AND_AND":           {},
	"OR_OR":             {},
	"QUESTION":          {},
	"COLON":             {},
	"SEMICOLON":         {},
	"VARIADIC":          {},
	"EQUAL":             {},
	"MULTIPLY_EQUAL":    {},
	"DIVIDE_EQUAL":      {},
	"MOD_EQUAL":         {},
	"PLUS_EQUAL":        {},
	"MINUS_EQUAL":       {},
	"LEFT_SHIFT_EQUAL":  {},
	"RIGHT_SHIFT_EQUAL": {},
	"AND_EQUAL":         {},
	"XOR_EQUAL":         {},
	"OR_EQUAL":          {},
	"COMMA":             {},
	"SHARP":             {},
	"SHARP_SHARP":       {},
}

func isTerminalSymbol(s string) bool {
	_, ok := terminals[s]
	return ok
}