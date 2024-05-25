package parser

import (
	"fmt"
	"github.com/mohae/deepcopy"
	"github.com/oleiade/lane/v2"
	"github.com/thoas/go-funk"
	"shinya.click/cvm/common"
	"shinya.click/cvm/parser/syntax"
)

type Parser struct {
	Tokens          []common.Token
	TokenIndex      int
	AST             *AstNode
	StateStack      *lane.Stack[int]
	SymbolStack     *lane.Stack[*AstNode]
	CheckPointStack *lane.Stack[*CheckPoint]
	TypeDefNames    []string
	Syntax          *syntax.Program
}

func NewParser(tokens []common.Token) *Parser {
	return &Parser{Tokens: tokens}
}

func (p *Parser) Parse() error {
	if err := p.constructAST(); err != nil {
		return err
	}

	if err := p.parseSyntax(); err != nil {
		return err
	}

	return nil
}

type CheckPoint struct {
	TokenIndex      int
	ChooseIndex     int
	StateStackSnap  *lane.Stack[int]
	SymbolStackSnap *lane.Stack[*AstNode]
}

func (p *Parser) constructAST() error {
	if p.AST != nil {
		return nil
	}

	p.TokenIndex = 0
	p.StateStack = lane.NewStack[int]()
	p.SymbolStack = lane.NewStack[*AstNode]()
	p.CheckPointStack = lane.NewStack[*CheckPoint]()

	p.StateStack.Push(0) // init state is always 0

	chooseOp := 0
	for {
		// read to the end before reduce program
		if p.TokenIndex >= len(p.Tokens) {
			chooseOp = p.restore()
			continue
		}
		token := p.Tokens[p.TokenIndex]
		state, ok := p.StateStack.Head()
		if !ok {
			chooseOp = p.restore()
			continue
		}

		ops, ok := lalrAction[state][token.Typ]
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
			newSym := &AstNode{Typ: prod.Left, ProdIndex: op.ReduceIndex}
			newSym.SetChildren(funk.Reverse(rights).([]*AstNode))
			p.SymbolStack.Push(newSym)
			nowHeadState, ok := p.StateStack.Head()
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
	p.CheckPointStack.Push(&CheckPoint{
		TokenIndex:      p.TokenIndex,
		ChooseIndex:     chooseOp,
		StateStackSnap:  deepcopy.Copy(p.StateStack).(*lane.Stack[int]),
		SymbolStackSnap: deepcopy.Copy(p.SymbolStack).(*lane.Stack[*AstNode]),
	})
}

func (p *Parser) restore() int {
	checkPoint, ok := p.CheckPointStack.Pop()
	if !ok {
		panic("total dead!")
	}

	p.TokenIndex = checkPoint.TokenIndex
	p.StateStack = checkPoint.StateStackSnap
	p.SymbolStack = checkPoint.SymbolStackSnap

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

func (p *Parser) parseSyntax() error {
	p.Syntax = &syntax.Program{}

	// flatten the outer translation unit
	units := flattenTranslationUnit(p.AST.Children[0])
	for _, unit := range units {
		switch productions[unit.ProdIndex].Right[0] {
		case function_definition:
			funcDef, err := parseFunctionDefinition(unit.Children[0])
			if err != nil {
				return err
			}
			p.Syntax.Units = append(p.Syntax.Units, funcDef)
		case declaration:
			declare, err := parseDeclaration(unit.Children[0])
			if err != nil {
				return err
			}
			printDeclaration(declare)
			p.Syntax.Units = append(p.Syntax.Units, declare)
		}
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
