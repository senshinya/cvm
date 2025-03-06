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

type CheckPoint struct {
	ChooseIndex     int
	TokenIndex      int
	StateStackSnap  []int
	SymbolStackSnap []*entity.AstNode
}

func NewParser(tokens []entity.Token) *Parser {
	return &Parser{Tokens: tokens}
}

func (p *Parser) Parse() (*entity.AstNode, error) {
	p.TokenIndex = 0
	p.StateStack = common.NewStack[int]()
	p.StateStack.Push(0) // init state is always 0
	p.SymbolStack = common.NewStack[*entity.AstNode]()

	p.CheckPointStack = common.NewStack[*CheckPoint]()

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
			newSym.SetChildren(prod, rights)
			p.CandidateASTs = append(p.CandidateASTs, newSym)
			chooseOp = p.restore()
			continue
		}
		switch op.OperatorType {
		case common.SHIFT:
			p.StateStack.Push(op.StateIndex)
			node := &entity.AstNode{
				Typ:         token.Typ,
				Terminal:    &token,
				SourceRange: token.GetSourceRange(),
			}
			p.SymbolStack.Push(node)
			if err := p.operatePostProcess(node); err != nil {
				chooseOp = p.restore()
				continue
			}
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
			newSym.SetChildren(prod, rights)
			p.SymbolStack.Push(newSym)
			if err := p.operatePostProcess(newSym); err != nil {
				chooseOp = p.restore()
				continue
			}

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

	// eliminate the wrong tree
	candidates, latestErr := chopForest(p.CandidateASTs)
	if len(candidates) == 0 {
		return nil, latestErr
	}
	for _, tree := range candidates {
		fillAstParent(tree, nil)
	}
	fmt.Printf("Parse Result: %d candidates\n", len(candidates))
	for i, candidate := range candidates {
		fmt.Printf("Tree %d\n", i)
		printAST(candidate, 0)
		fmt.Println()
		fmt.Println()
	}
	return candidates[0], nil
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

func fillAstParent(node *entity.AstNode, parent *entity.AstNode) {
	node.Parent = parent
	for _, child := range node.Children {
		fillAstParent(child, node)
	}
}

func printAST(ast *entity.AstNode, level int) {
	for i := 0; i < level; i++ {
		fmt.Print("  ")
	}
	fmt.Print(ast.Typ)
	if entity.IsTerminalSymbol(string(ast.Typ)) {
		fmt.Print(" - " + ast.Terminal.Lexeme)
	}
	fmt.Printf(" %v %v", ast.TypeDef, ast.DeclaratorID)
	fmt.Println()
	for _, child := range ast.Children {
		printAST(child, level+1)
	}
}

func (p *Parser) operatePostProcess(node *entity.AstNode) error {
	switch node.Typ {
	case DirectDeclarator:
		switch {
		case node.ReducedBy(DirectDeclarator, 1):
			// direct_declarator := IDENTIFIER
			node.DeclaratorID = []*entity.Token{node.Children[0].Terminal}
		case node.ReducedBy(DirectDeclarator, 2):
			// direct_declarator := LEFT_PARENTHESES declarator RIGHT_PARENTHESES
			node.DeclaratorID = nil
		default:
			// direct_declarator := direct_declarator ...
			node.DeclaratorID = node.Children[0].DeclaratorID
		}
	case IdentifierList:
		if node.ReducedBy(IdentifierList, 1) {
			// identifier_list := IDENTIFIER
			node.DeclaratorID = []*entity.Token{node.Children[0].Terminal}
		}
		if node.ReducedBy(IdentifierList, 2) {
			// identifier_list := identifier_list COMMA IDENTIFIER
			node.DeclaratorID = append(node.DeclaratorID, node.Children[2].Terminal)
		}
	case StructOrUnionSpecifier:
		node.DeclaratorID = nil
	case EnumerationConstant:
		// enumeration_constant := IDENTIFIER
		node.DeclaratorID = []*entity.Token{node.Children[0].Terminal}
	case AbstractDeclarator:
		// prevent ids in abstract declarator from passing to the parent node
		node.DeclaratorID = nil
	case StorageClassSpecifier:
		// storage_class_specifier := TYPEDEF
		if node.ReducedBy(StorageClassSpecifier, 1) {
			node.TypeDef = true
		}
	case Declaration, FunctionDefinition:
		// clear label
		node.TypeDef = false
		node.DeclaratorID = nil
	}
	return nil
}
