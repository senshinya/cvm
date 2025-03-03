package parser

import (
	"fmt"
	"shinya.click/cvm/common"
	"shinya.click/cvm/entity"
	"slices"
)

type Parser struct {
	Tokens          []entity.Token
	TokenIndex      int
	StateStack      *common.Stack[int]
	SymbolStack     *common.Stack[*entity.AstNode]
	CheckPointStack *common.Stack[*CheckPoint]
	CandidateASTs   []*entity.AstNode
}

func NewParser(tokens []entity.Token) *Parser {
	return &Parser{Tokens: tokens}
}

type CheckPoint struct {
	ChooseIndex     int
	TokenIndex      int
	StateStackSnap  []int
	SymbolStackSnap []*entity.AstNode
}

func (p *Parser) Parse() (*entity.AstNode, error) {
	p.TokenIndex = 0
	p.StateStack = common.NewStack[int]()
	p.SymbolStack = common.NewStack[*entity.AstNode]()
	p.CheckPointStack = common.NewStack[*CheckPoint]()

	p.StateStack.Push(0) // init state is always 0

	chooseOp := 0
parserIter:
	for {
		if chooseOp == -1 {
			// no checkpoint to restore
			break
		}
		// read to the end before reduce program
		if p.TokenIndex >= len(p.Tokens) {
			chooseOp = p.restore()
			continue
		}
		token := p.Tokens[p.TokenIndex]
		state, ok := p.StateStack.Peek()
		if !ok {
			chooseOp = p.restore()
			continue
		}

		ops, ok := LalrAction[state][token.Typ]
		if !ok {
			chooseOp = p.restore()
			continue
		}
		if chooseOp >= len(ops) {
			chooseOp = p.restore()
			continue
		}
		if len(ops) != 1 {
			p.addCheckPoint(chooseOp)
		}
		op := ops[chooseOp]
		if op.OperatorType == common.ACC {
			prod := Productions[op.ReduceIndex]
			var rights []*entity.AstNode
			for i := 0; i < len(prod.Right); i++ {
				p.StateStack.Pop()
				sym, ok := p.SymbolStack.Pop()
				if !ok {
					chooseOp = p.restore()
					continue parserIter
				}
				rights = append(rights, sym)
			}
			slices.Reverse(rights)
			newSym := &entity.AstNode{Typ: prod.Left}
			newSym.SetBranch(prod, rights)
			p.CandidateASTs = append(p.CandidateASTs, newSym)
			chooseOp = p.restore()
			continue
		}
		switch op.OperatorType {
		case common.SHIFT:
			p.StateStack.Push(op.StateIndex)
			p.SymbolStack.Push(&entity.AstNode{
				Typ:         token.Typ,
				Terminal:    &token,
				SourceRange: token.GetSourceRange(),
			})
			p.TokenIndex++
		case common.REDUCE:
			prod := Productions[op.ReduceIndex]
			var rights []*entity.AstNode
			for i := 0; i < len(prod.Right); i++ {
				p.StateStack.Pop()
				sym, ok := p.SymbolStack.Pop()
				if !ok {
					chooseOp = p.restore()
					continue parserIter
				}
				rights = append(rights, sym)
			}
			slices.Reverse(rights)
			newSym := &entity.AstNode{Typ: prod.Left}
			newSym.SetBranch(prod, rights)
			p.SymbolStack.Push(newSym)

			nowHeadState, ok := p.StateStack.Peek()
			if !ok {
				chooseOp = p.restore()
				continue
			}
			gotoState, ok := LalrGoto[nowHeadState][newSym.Typ]
			if !ok {
				chooseOp = p.restore()
				continue
			}
			p.StateStack.Push(gotoState)
		}
		chooseOp = 0
	}

	if len(p.CandidateASTs) == 0 {
		panic("dead end")
	}

	// merge all candidate ASTs
	res := p.CandidateASTs[0]
	for i, ast := range p.CandidateASTs {
		if i == 0 {
			continue
		}
		res.Merge(ast)
	}

	// fill parent pointer
	fillAstParent(res, nil)
	printAST(res, 0)
	return res, nil
}

func (p *Parser) addCheckPoint(chooseOp int) {
	stateAll := p.StateStack.DumpAll()
	symAll := p.SymbolStack.DumpAll()
	cp := CheckPoint{
		TokenIndex:      p.TokenIndex,
		ChooseIndex:     chooseOp,
		StateStackSnap:  common.DeepCopy(stateAll),
		SymbolStackSnap: common.DeepCopy(symAll),
	}
	p.CheckPointStack.Push(&cp)
}

func fillAstParent(node *entity.AstNode, parent *entity.AstNode) {
	node.Parent = parent
	for _, branch := range node.PossibleBranches {
		for _, child := range branch.Children {
			fillAstParent(child, node)
		}
	}
}

func (p *Parser) restore() int {
	checkPoint, ok := p.CheckPointStack.Pop()
	if !ok {
		return -1
	}
	p.TokenIndex = checkPoint.TokenIndex
	p.StateStack = common.NewStackWithElements[int](checkPoint.StateStackSnap)
	p.SymbolStack = common.NewStackWithElements[*entity.AstNode](checkPoint.SymbolStackSnap)

	return checkPoint.ChooseIndex + 1
}

func printAST(ast *entity.AstNode, level int) {
	for i := 0; i < level; i++ {
		fmt.Print("  ")
	}
	fmt.Print(ast.Typ)
	if entity.IsTerminalSymbol(string(ast.Typ)) {
		fmt.Print(" - " + ast.Terminal.Lexeme)
	}
	fmt.Println()
	if len(ast.PossibleBranches) == 1 {
		for _, child := range ast.PossibleBranches[0].Children {
			printAST(child, level+1)
		}
	} else {
		for i, branch := range ast.PossibleBranches {
			for j := 0; j < level; j++ {
				fmt.Print("  ")
			}
			fmt.Printf("branch %d\n", i)
			for _, child := range branch.Children {
				printAST(child, level+1)
			}
		}
	}
}
