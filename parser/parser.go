package parser

import (
	"errors"
	"fmt"
	"github.com/hyphennn/glambda/gslice"
	"shinya.click/cvm/common"
	"shinya.click/cvm/entity"
	"slices"
)

type Parser struct {
	Tokens             []entity.Token
	TokenIndex         int
	StateStack         *common.Stack[int]
	SymbolStack        *common.Stack[*entity.AstNode]
	TypeDefSymbolTable *SymbolTable
	CheckPointStack    *common.Stack[*CheckPoint]
	CandidateASTs      []*entity.AstNode
}

type CheckPoint struct {
	ChooseIndex        int
	TokenIndex         int
	StateStackSnap     []int
	SymbolStackSnap    []*entity.AstNode
	TypeDefSymbolTable *SymbolTable
}

type SymbolTable struct {
	Stack [][]string
}

func (st *SymbolTable) Push(sym ...string) {
	st.Stack[len(st.Stack)-1] = append(st.Stack[len(st.Stack)-1], sym...)
}

func (st *SymbolTable) Contain(sym string) bool {
	for i := len(st.Stack) - 1; i >= 0; i-- {
		if gslice.Contains(st.Stack[i], sym) {
			return true
		}
	}
	return false
}

func (st *SymbolTable) EnterScope() {
	st.Stack = append(st.Stack, []string{})
}

func (st *SymbolTable) LeaveScope() {
	st.Stack = st.Stack[:len(st.Stack)-1]
}

func NewParser(tokens []entity.Token) *Parser {
	return &Parser{Tokens: tokens}
}

func (p *Parser) Parse() (*entity.AstNode, error) {
	p.TokenIndex = 0
	p.StateStack = common.NewStack[int]()
	p.StateStack.Push(0) // init state is always 0
	p.SymbolStack = common.NewStack[*entity.AstNode]()
	p.TypeDefSymbolTable = &SymbolTable{
		Stack: [][]string{{}},
	}

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
			newSym.SetBranch(prod, rights)
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
			newSym.SetBranch(prod, rights)
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
		TokenIndex:         p.TokenIndex,
		ChooseIndex:        chooseOp,
		StateStackSnap:     common.DeepCopy(stateAll),
		SymbolStackSnap:    common.DeepCopy(symAll),
		TypeDefSymbolTable: common.DeepCopy(p.TypeDefSymbolTable),
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
	p.TypeDefSymbolTable = checkPoint.TypeDefSymbolTable

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

func (p *Parser) operatePostProcess(node *entity.AstNode) error {
	switch node.Typ {
	case DirectDeclarator:
		// direct_declarator := IDENTIFIER
		branch := node.GetProductionBranch(DirectDeclarator, 1)
		if branch == nil {
			return nil
		}
		node.DeclaratorID = append(node.DeclaratorID, branch.Children[0].Terminal.Lexeme)
	case StorageClassSpecifier:
		// storage_class_specifier := TYPEDEF
		branch := node.GetProductionBranch(StorageClassSpecifier, 1)
		if branch == nil {
			return nil
		}
		node.TypeDef = true
	case Declaration:
		// declaration := declaration_specifiers init_declarator_list SEMICOLON
		branch := node.GetProductionBranch(Declaration, 2)
		if branch != nil && branch.Children[0].TypeDef {
			// typedef declaration
			p.TypeDefSymbolTable.Push(branch.Children[1].DeclaratorID...)
		}
		// clear label
		node.TypeDef = false
		node.DeclaratorID = nil
	case entity.LEFT_BRACES:
		p.TypeDefSymbolTable.EnterScope()
	case entity.RIGHT_BRACES:
		p.TypeDefSymbolTable.LeaveScope()
	case TypedefName:
		// typedef_name := IDENTIFIER
		branch := node.GetProductionBranch(TypedefName, 1)
		if branch == nil {
			return errors.New("typedef_name := IDENTIFIER branch not found")
		}
		if !p.TypeDefSymbolTable.Contain(branch.Children[0].Terminal.Lexeme) {
			return fmt.Errorf("type name %s not found", branch.Children[0].Terminal.Lexeme)
		}
	}
	return nil
}
