package parser

import (
	"fmt"
	"github.com/mohae/deepcopy"
	"github.com/thoas/go-funk"
	"shinya.click/cvm/common"
	"shinya.click/cvm/parser/syntax"
)

type Parser struct {
	Tokens           []common.Token
	TokenIndex       int
	AST              *AstNode
	StateStack       *Stack[int]
	SymbolStack      *Stack[*AstNode]
	CheckPointStack  *Stack[*CheckPoint]
	TranslationUnits []syntax.TranslationUnit
}

func NewParser(tokens []common.Token) *Parser {
	return &Parser{Tokens: tokens}
}

type CheckPoint struct {
	TokenIndex       int
	ChooseIndex      int
	StateStackSnap   []int
	SymbolStackSnap  []*AstNode
	TranslationUnits []syntax.TranslationUnit
}

func (p *Parser) Parse() error {
	if p.AST != nil {
		return nil
	}

	p.TokenIndex = 0
	p.StateStack = NewStack[int]()
	p.SymbolStack = NewStack[*AstNode]()
	p.CheckPointStack = NewStack[*CheckPoint]()
	p.TranslationUnits = []syntax.TranslationUnit{}

	p.StateStack.Push(0) // init state is always 0

	chooseOp := 0
	for {
		// read to the end before reduce program
		if p.TokenIndex >= len(p.Tokens) {
			fmt.Println("alo1")
			chooseOp = p.restore()
			continue
		}
		token := p.Tokens[p.TokenIndex]
		state, ok := p.StateStack.Peek()
		if !ok {
			fmt.Println("alo2")
			chooseOp = p.restore()
			continue
		}

		ops, ok := lalrAction[state][token.Typ]
		if !ok {
			fmt.Println("alo3")
			chooseOp = p.restore()
			continue
		}
		if chooseOp >= len(ops) {
			fmt.Println("alo4")
			chooseOp = p.restore()
			continue
		}
		if len(ops) != 1 {
			p.addCheckPoint(chooseOp)
		}
		op := ops[chooseOp]
		if op.OperatorType == ACC {
			prod := productions[op.ReduceIndex]
			var rights []*AstNode
			for i := 0; i < len(prod.Right); i++ {
				p.StateStack.Pop()
				sym, ok := p.SymbolStack.Pop()
				if !ok {
					chooseOp = p.restore()
					continue
				}
				rights = append(rights, sym)
			}
			newSym := &AstNode{Typ: prod.Left, ProdIndex: op.ReduceIndex}
			newSym.SetChildren(funk.Reverse(rights).([]*AstNode))
			p.SymbolStack.Push(newSym)
			break
		}
		switch op.OperatorType {
		case SHIFT:
			p.StateStack.Push(op.StateIndex)
			p.SymbolStack.Push(&AstNode{
				Typ:      token.Typ,
				Terminal: &token,
			})
			p.TokenIndex++
		case REDUCE:
			prod := productions[op.ReduceIndex]
			var rights []*AstNode
			for i := 0; i < len(prod.Right); i++ {
				p.StateStack.Pop()
				sym, ok := p.SymbolStack.Pop()
				if !ok {
					chooseOp = p.restore()
					continue
				}
				rights = append(rights, sym)
			}
			fmt.Printf("Reduced by %+v\n", prod)
			newSym := &AstNode{Typ: prod.Left, ProdIndex: op.ReduceIndex}
			newSym.SetChildren(funk.Reverse(rights).([]*AstNode))
			p.SymbolStack.Push(newSym)

			if prod.Left == translation_unit {
				err := p.parseTranslationUnit(newSym)
				if err != nil {
					fmt.Println(err)
					chooseOp = p.restore()
					continue
				}
			}

			nowHeadState, ok := p.StateStack.Peek()
			if !ok {
				chooseOp = p.restore()
				continue
			}
			gotoState, ok := lalrGoto[nowHeadState][newSym.Typ]
			if !ok {
				chooseOp = p.restore()
				continue
			}
			p.StateStack.Push(gotoState)
		}
		chooseOp = 0
	}

	res, ok := p.SymbolStack.Pop()
	if !ok {
		panic("symbolStack is empty")
	}
	p.AST = res
	// printAST(p.AST, 0)

	return nil
}

func (p *Parser) addCheckPoint(chooseOp int) {
	stateAll := p.StateStack.DumpAll()
	symAll := p.SymbolStack.DumpAll()
	cp := CheckPoint{
		TokenIndex:       p.TokenIndex,
		ChooseIndex:      chooseOp,
		StateStackSnap:   deepcopy.Copy(stateAll).([]int),
		SymbolStackSnap:  deepCopyAstNodeSlice(symAll),
		TranslationUnits: deepcopy.Copy(p.TranslationUnits).([]syntax.TranslationUnit),
	}
	p.CheckPointStack.Push(&cp)
	fmt.Printf("addCheckPoint %+v\n", cp)
}

func deepCopyAstNodeSlice(origins []*AstNode) []*AstNode {
	// due to a node contains parent and children, it cannot be deepcopy.Copy(), or stack overflow
	var res []*AstNode
	for _, origin := range origins {
		res = append(res, copyAstNode(origin))
	}
	for _, node := range res {
		fillAstParent(node, nil)
	}
	return res
}

func copyAstNode(origin *AstNode) *AstNode {
	root := &AstNode{
		Typ:       origin.Typ,
		Terminal:  origin.Terminal,
		ProdIndex: origin.ProdIndex,
	}
	for _, child := range origin.Children {
		root.Children = append(root.Children, copyAstNode(child))
	}
	return root
}

func fillAstParent(node *AstNode, parent *AstNode) {
	node.Parent = parent
	for _, child := range node.Children {
		fillAstParent(child, node)
	}
}

func (p *Parser) restore() int {
	checkPoint, ok := p.CheckPointStack.Pop()
	if !ok {
		panic("total dead!")
	}
	fmt.Printf("restore to %+v\n", *checkPoint)
	p.TokenIndex = checkPoint.TokenIndex
	p.StateStack = NewStackWithElements[int](checkPoint.StateStackSnap)
	p.SymbolStack = NewStackWithElements[*AstNode](checkPoint.SymbolStackSnap)
	p.TranslationUnits = checkPoint.TranslationUnits

	return checkPoint.ChooseIndex + 1
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

func (p *Parser) parseTranslationUnit(unit *AstNode) error {
	switch productions[unit.ProdIndex].Right[0] {
	case function_definition:
		funcDef, err := parseFunctionDefinition(unit.Children[0])
		if err != nil {
			return err
		}
		p.TranslationUnits = append(p.TranslationUnits, funcDef)
	case declaration:
		declare, err := parseDeclaration(unit.Children[0])
		if err != nil {
			return err
		}
		printDeclaration(declare)
		p.TranslationUnits = append(p.TranslationUnits, declare)
	}
	return nil
}

func printDeclaration(unit syntax.TranslationUnit) {
	declares := unit.(*syntax.Declaration)
	for _, declare := range declares.Declarators {
		fmt.Printf("declare %s as ", declare.Identifier)
		typ := &declare.Type
		for {
			if typ.TypeQualifiers.Const {
				fmt.Print("const ")
			}
			if typ.TypeQualifiers.Volatile {
				fmt.Print("volatile ")
			}
			if typ.TypeQualifiers.Restrict {
				fmt.Print("restrict ")
			}
			switch typ.MetaType {
			case syntax.MetaTypeVoid:
				print("void")
				goto out
			case syntax.MetaTypeNumber:
				numMeta := typ.NumberMetaInfo
				if numMeta.Signed {
					print("signed ")
				}
				if numMeta.Unsigned {
					print("unsigned ")
				}
				switch numMeta.BaseNumType {
				case syntax.BaseNumTypeChar:
					print("char")
				case syntax.BaseNumTypeShort:
					print("short")
				case syntax.BaseNumTypeInt:
					print("int")
				case syntax.BaseNumTypeLong:
					print("long")
				case syntax.BaseNumTypeFloat:
					print("float")
				case syntax.BaseNumTypeDouble:
					print("double")
				case syntax.BaseNumTypeBool:
					print("bool")
				case syntax.BaseNumTypeLongLong:
					print("long long")
				case syntax.BaseNumTypeLongDouble:
					print("long double")
				}
				goto out
			case syntax.MetaTypeEnum:
			case syntax.MetaTypePointer:
				print("pointer to ")
				typ = typ.PointerInnerType
			case syntax.MetaTypeStruct:
			case syntax.MetaTypeUnion:
			case syntax.MetaTypeFunction:
				print("function returning ")
				typ = typ.FunctionMetaInfo.ReturnType
			case syntax.MetaTypeArray:
				print("array of ")
				typ = typ.ArrayMetaInfo.InnerType
			case syntax.MetaTypeUserDefined:
				// TODO user defined
				print("user defined")
				goto out
			}
		}
	out:
		fmt.Println()
	}
}

func flattenTranslationUnit(ast *AstNode) []*AstNode {
	// ast must be program_units node
	if ast.Typ != program_units {
		return nil
	}

	// if reduced by program_units := translation_unit
	if len(productions[ast.ProdIndex].Right) == 1 {
		return []*AstNode{ast.Children[0]}
	}

	// reduced by program_units := program_units translation_unit
	return append(flattenTranslationUnit(ast.Children[0]), ast.Children[1])
}
