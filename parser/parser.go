package parser

import (
	"fmt"
	"github.com/oleiade/lane/v2"
	"github.com/thoas/go-funk"
	"shinya.click/cvm/common"
)

type Parser struct {
	Tokens       []common.Token
	AST          *AstNode
	TypeDefNames []string
}

func NewParser(tokens []common.Token) *Parser {
	return &Parser{Tokens: tokens}
}

func (p *Parser) ConstructAST() error {
	if p.AST != nil {
		return nil
	}

	stateStack := lane.NewStack[int]()
	symbolStack := lane.NewStack[*AstNode]()

	stateStack.Push(0) // init state is always 0

	tokenIndex := 0
	for {
		// read to the end before reduce program
		if tokenIndex >= len(p.Tokens) {
			panic("read to the end!")
		}
		token := p.Tokens[tokenIndex]
		state, ok := stateStack.Head()
		if !ok {
			panic("stateStack is empty")
		}

		op, ok := lalrAction[state][token.Typ]
		if !ok {
			panic("Unexpected token!" + token.Typ)
		}
		if op.OperatorType == ACC {
			break
		}
		switch op.OperatorType {
		case SHIFT:
			stateStack.Push(op.StateIndex)
			symbolStack.Push(&AstNode{
				Typ:      token.Typ,
				Terminal: &token,
			})
			tokenIndex++
		case REDUCE:
			prod := productions[op.ReduceIndex]
			var rights []*AstNode
			for i := 0; i < len(prod.Right); i++ {
				stateStack.Pop()
				sym, ok := symbolStack.Pop()
				if !ok {
					panic("symbolStack is empty")
				}
				rights = append(rights, sym)
			}
			newSym := &AstNode{Typ: prod.Left}
			newSym.SetChildren(funk.Reverse(rights).([]*AstNode))
			symbolStack.Push(newSym)
			nowHeadState, ok := stateStack.Head()
			if !ok {
				panic("stateStack is empty")
			}
			gotoState, ok := lalrGoto[nowHeadState][newSym.Typ]
			if !ok {
				panic("gotoState is empty")
			}
			stateStack.Push(gotoState)
		}
	}

	res, ok := symbolStack.Pop()
	if !ok {
		panic("symbolStack is empty")
	}
	p.AST = res
	printAST(p.AST, 0)

	return nil
}

func printAST(ast *AstNode, level int) {
	for i := 0; i < level; i++ {
		fmt.Print("  ")
	}
	fmt.Print(ast.Typ)
	if common.IsTerminalSymbol(string(ast.Typ)) {
		fmt.Print(" - " + ast.Terminal.Lexeme)
	}
	fmt.Println()
	for _, child := range ast.Children {
		printAST(child, level+1)
	}
}
