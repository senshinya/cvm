package parser

import (
	"errors"
	"github.com/hyphennn/glambda/gslice"
	"shinya.click/cvm/common"
	"shinya.click/cvm/entity"
	"slices"
)

type Parser struct {
	Tokens          []entity.Token
	TokenIndex      int
	StateStack      *common.Stack[int]
	SymbolStack     *common.Stack[*entity.AstNode]
	TypeDefSymbols  [][]string
	CheckPointStack *common.Stack[*CheckPoint]
	CandidateASTs   []*entity.AstNode
	bestError       error
	bestErrorPos    entity.SourcePos
	ForkCount       int
}

type CheckPoint struct {
	ChooseIndex        int
	TokenIndex         int
	StateStackSnap     []int
	SymbolStackSnap    []*entity.AstNode
	TypeDefSymbolsSnap [][]string
}

func NewParser(tokens []entity.Token) *Parser {
	return &Parser{Tokens: tokens}
}

func (p *Parser) Parse() ([]*entity.AstNode, error) {
	p.TokenIndex = 0
	p.StateStack = common.NewStack[int]()
	p.StateStack.Push(0) // init state is always 0
	p.SymbolStack = common.NewStack[*entity.AstNode]()
	p.TypeDefSymbols = [][]string{{}}

	p.CheckPointStack = common.NewStack[*CheckPoint]()
	p.bestError = nil
	p.bestErrorPos = entity.SourcePos{}
	p.ForkCount = 0

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
			p.recordError(UnexpectedToken(token.SourceStart, token.Lexeme))
			chooseOp = p.restore()
			continue
		}
		if len(ops) > 1 {
			ops = p.pruneTypedefFork(token, ops)
		}
		if chooseOp >= len(ops) {
			p.recordError(UnexpectedToken(token.SourceStart, token.Lexeme))
			chooseOp = p.restore()
			continue
		}
		if len(ops) != 1 {
			p.addCheckPoint(chooseOp)
		}
		op := ops[chooseOp]
		if op.OperatorType == ACC {
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
		case SHIFT:
			p.StateStack.Push(op.StateIndex)
			node := &entity.AstNode{
				Typ:         token.Typ,
				Terminal:    &token,
				SourceRange: token.GetSourceRange(),
			}
			p.SymbolStack.Push(node)
			if err := p.operatePostProcess(node); err != nil {
				p.recordError(err)
				chooseOp = p.restore()
				continue
			}
			p.TokenIndex++
		case REDUCE:
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
				p.recordError(err)
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
		return nil, p.bestError
	}
	for _, tree := range p.CandidateASTs {
		fillAstParent(tree, nil)
	}
	return p.CandidateASTs, nil
}

func (p *Parser) addCheckPoint(chooseOp int) {
	stateAll := p.StateStack.DumpAll()
	symAll := p.SymbolStack.DumpAll()
	cp := CheckPoint{
		TokenIndex:         p.TokenIndex,
		ChooseIndex:        chooseOp,
		StateStackSnap:     slices.Clone(stateAll),
		SymbolStackSnap:    slices.Clone(symAll),
		TypeDefSymbolsSnap: cloneTypeDefSymbols(p.TypeDefSymbols),
	}
	p.CheckPointStack.Push(&cp)
	p.ForkCount++
}

func cloneTypeDefSymbols(symbols [][]string) [][]string {
	cloned := make([][]string, len(symbols))
	for i := range symbols {
		cloned[i] = slices.Clone(symbols[i])
	}
	return cloned
}

func (p *Parser) restore() int {
	checkPoint, ok := p.CheckPointStack.Pop()
	if !ok {
		return -1
	}
	p.TokenIndex = checkPoint.TokenIndex
	p.StateStack = common.NewStackWithElements[int](checkPoint.StateStackSnap)
	p.SymbolStack = common.NewStackWithElements[*entity.AstNode](checkPoint.SymbolStackSnap)
	p.TypeDefSymbols = checkPoint.TypeDefSymbolsSnap

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
		common.DebugPrint("  ")
	}
	common.DebugPrint(ast.Typ)
	if entity.IsTerminalSymbol(string(ast.Typ)) {
		common.DebugPrint(" - " + ast.Terminal.Lexeme)
	}
	common.DebugPrintf(" %v %v", ast.TypeDef, ast.DeclaratorID)
	common.DebugPrintln()
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
			node.DeclaratorID = node.Children[1].DeclaratorID
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
	case EnumerationConstant:
		// enumeration_constant := IDENTIFIER
		node.DeclaratorID = []*entity.Token{node.Children[0].Terminal}
	case StructOrUnionSpecifier:
		// clear DeclaratorID
		node.DeclaratorID = nil
	case ParameterDeclaration:
		if err := checkDeclarationSpecifiers(node.Children[0]); err != nil {
			return err
		}
	case StorageClassSpecifier:
		// storage_class_specifier := TYPEDEF
		if node.ReducedBy(StorageClassSpecifier, 1) {
			node.TypeDef = true
		}
	case Declaration, FunctionDefinition:
		// add typedef name to the typedef stack
		if node.TypeDef {
			p.TypeDefSymbols[len(p.TypeDefSymbols)-1] = append(p.TypeDefSymbols[len(p.TypeDefSymbols)-1],
				gslice.Map(node.DeclaratorID, func(token *entity.Token) string {
					return token.Lexeme
				})...)
		}
		// when Declaration specifier contains typedef name, it should be the only type specifier
		if err := checkDeclarationSpecifiers(node.Children[0]); err != nil {
			return err
		}
		// clear label
		node.TypeDef = false
		node.DeclaratorID = nil
	case entity.LEFT_BRACES:
		p.TypeDefSymbols = append(p.TypeDefSymbols, []string{})
	case entity.RIGHT_BRACES:
		p.TypeDefSymbols = p.TypeDefSymbols[:len(p.TypeDefSymbols)-1]
	case TypedefName:
		// typedef_name := IDENTIFIER
		// check if the typedef name is in the typedef stack
		id := node.Children[0].Terminal.Lexeme
		for _, symbols := range p.TypeDefSymbols {
			if slices.Contains(symbols, id) {
				return nil
			}
		}
		return UndeclaredIdentifier(node.SourceStart, id)
	}
	return nil
}

func checkDeclarationSpecifiers(node *entity.AstNode) error {
	if node.Typ != DeclarationSpecifiers {
		return nil
	}
	typeSpecifiers := getAllTypeSpecifiers(node)
	for _, typeSpecifier := range typeSpecifiers {
		if typeSpecifier.ReducedBy(TypeSpecifier, 14) && len(typeSpecifiers) > 1 {
			// type_specifier := typedef_name
			return InvalidTypeSpecifier(node.SourceStart)
		}
	}
	return nil
}

// recordError keeps the error from the deepest source position seen so far.
// Branches that fail very early produce noisy, misleading errors; the error
// from the branch that consumed the most input is usually the closest to the
// real syntax problem.
func (p *Parser) recordError(err error) {
	if err == nil {
		return
	}
	pos := errorPos(err)
	if p.bestError == nil || comparePos(pos, p.bestErrorPos) > 0 {
		p.bestError = err
		p.bestErrorPos = pos
	}
}

func errorPos(err error) entity.SourcePos {
	var cvmErr *common.CvmError
	if errors.As(err, &cvmErr) && len(cvmErr.Messages) > 0 {
		return cvmErr.Messages[0].SourcePos
	}
	return entity.SourcePos{}
}

func comparePos(a, b entity.SourcePos) int {
	if a.Line != b.Line {
		return a.Line - b.Line
	}
	return a.Column - b.Column
}

// pruneTypedefFork resolves the two known typedef-name conflicts at fork time
// using p.TypeDefSymbols.
//
// Pattern A: SHIFT vs REDUCE on lookahead IDENTIFIER. SHIFT consumes IDENTIFIER
// as a typedef_name continuation of declaration_specifiers; REDUCE finalizes
// the existing declaration_specifiers and treats IDENTIFIER as the declarator.
//
// Pattern B: REDUCE[typedef_name -> IDENTIFIER] vs REDUCE[primary_expression
// -> IDENTIFIER] on lookahead LEFT_PARENTHESES. The IDENTIFIER is already on
// the symbol stack.
//
// Pruning is asymmetric on purpose. TypeDefSymbols only ever appends typedef
// declarations and never tracks variable declarations that shadow them, so:
//   - Identifier NOT in stack -> definitely not a typedef -> drop typedef branch.
//   - Identifier IN stack -> might be a typedef OR a shadowed name -> keep both
//     branches but reorder so the typedef branch runs first.
//
// The shadowing case must remain forkable for the future semantic layer to pick
// the right tree.
func (p *Parser) pruneTypedefFork(token entity.Token, ops []DFAOperator) []DFAOperator {
	if len(ops) != 2 {
		return ops
	}

	// Pattern B: both REDUCE, exactly one is the typedef_name reduction.
	if ops[0].OperatorType == REDUCE && ops[1].OperatorType == REDUCE {
		typedefIdx, otherIdx := -1, -1
		for i, op := range ops {
			if Productions[op.ReduceIndex].Left == TypedefName {
				typedefIdx = i
			} else {
				otherIdx = i
			}
		}
		if typedefIdx == -1 || otherIdx == -1 {
			return ops
		}
		top, ok := p.SymbolStack.Peek()
		if !ok || top.Terminal == nil {
			return ops
		}
		if p.isTypedefName(top.Terminal.Lexeme) {
			return []DFAOperator{ops[typedefIdx], ops[otherIdx]}
		}
		return []DFAOperator{ops[otherIdx]}
	}

	// Pattern A: one SHIFT, one REDUCE, lookahead IDENTIFIER.
	if token.Typ != entity.IDENTIFIER {
		return ops
	}
	shiftIdx, reduceIdx := -1, -1
	for i, op := range ops {
		switch op.OperatorType {
		case SHIFT:
			shiftIdx = i
		case REDUCE:
			reduceIdx = i
		}
	}
	if shiftIdx == -1 || reduceIdx == -1 {
		return ops
	}
	if p.isTypedefName(token.Lexeme) {
		return []DFAOperator{ops[shiftIdx], ops[reduceIdx]}
	}
	return []DFAOperator{ops[reduceIdx]}
}

func (p *Parser) isTypedefName(name string) bool {
	for _, symbols := range p.TypeDefSymbols {
		if slices.Contains(symbols, name) {
			return true
		}
	}
	return false
}

func getAllTypeSpecifiers(node *entity.AstNode) []*entity.AstNode {
	switch {
	case node.ReducedBy(DeclarationSpecifiers, 2):
		// declaration_specifiers := type_specifier
		return []*entity.AstNode{node.Children[0]}
	case node.ReducedBy(DeclarationSpecifiers, 6):
		// declaration_specifiers := type_specifier declaration_specifiers
		return append(getAllTypeSpecifiers(node.Children[1]), node.Children[0])
	case node.ReducedBy(DeclarationSpecifiers, 5),
		node.ReducedBy(DeclarationSpecifiers, 7),
		node.ReducedBy(DeclarationSpecifiers, 8):
		// declaration_specifiers := ... declaration_specifiers
		return getAllTypeSpecifiers(node.Children[1])
	}
	return nil
}
