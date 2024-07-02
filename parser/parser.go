package parser

import (
	"fmt"
	"github.com/mohae/deepcopy"
	"github.com/thoas/go-funk"
	"shinya.click/cvm/common"
	"shinya.click/cvm/parser/entity"
	"shinya.click/cvm/parser/glr"
	"strings"
)

type Parser struct {
	Tokens               []common.Token
	TokenIndex           int
	AST                  *glr.RawAstNode
	StateStack           *common.Stack[int]
	SymbolStack          *common.Stack[*glr.RawAstNode]
	CheckPointStack      *common.Stack[*CheckPoint]
	ExternalDeclarations []entity.ExternalDeclaration
}

func NewParser(tokens []common.Token) *Parser {
	return &Parser{Tokens: tokens}
}

type CheckPoint struct {
	TokenIndex           int
	ChooseIndex          int
	StateStackSnap       []int
	SymbolStackSnap      []*glr.RawAstNode
	ExternalDeclarations []entity.ExternalDeclaration
}

func (p *Parser) Parse() error {
	if p.AST != nil {
		return nil
	}

	p.TokenIndex = 0
	p.StateStack = common.NewStack[int]()
	p.SymbolStack = common.NewStack[*glr.RawAstNode]()
	p.CheckPointStack = common.NewStack[*CheckPoint]()
	p.ExternalDeclarations = []entity.ExternalDeclaration{}

	p.StateStack.Push(0) // init state is always 0

	chooseOp := 0
	for {
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

		ops, ok := glr.LalrAction[state][token.Typ]
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
		if op.OperatorType == glr.ACC {
			prod := glr.Productions[op.ReduceIndex]
			var rights []*glr.RawAstNode
			for i := 0; i < len(prod.Right); i++ {
				p.StateStack.Pop()
				sym, ok := p.SymbolStack.Pop()
				if !ok {
					chooseOp = p.restore()
					continue
				}
				rights = append(rights, sym)
			}
			newSym := &glr.RawAstNode{Typ: prod.Left, Production: prod}
			newSym.SetChildren(funk.Reverse(rights).([]*glr.RawAstNode))
			p.SymbolStack.Push(newSym)
			break
		}
		switch op.OperatorType {
		case glr.SHIFT:
			p.StateStack.Push(op.StateIndex)
			p.SymbolStack.Push(&glr.RawAstNode{
				Typ:         token.Typ,
				Terminal:    &token,
				SourceRange: token.GetSourceRange(),
			})
			p.TokenIndex++
		case glr.REDUCE:
			prod := glr.Productions[op.ReduceIndex]
			var rights []*glr.RawAstNode
			for i := 0; i < len(prod.Right); i++ {
				p.StateStack.Pop()
				sym, ok := p.SymbolStack.Pop()
				if !ok {
					chooseOp = p.restore()
					continue
				}
				rights = append(rights, sym)
			}
			newSym := &glr.RawAstNode{Typ: prod.Left, Production: prod}
			newSym.SetChildren(funk.Reverse(rights).([]*glr.RawAstNode))
			p.SymbolStack.Push(newSym)

			if prod.Left == glr.ExternalDeclaration {
				//printAST(newSym, 0)
				err := p.parseExternalDeclaration(newSym)
				if err != nil {
					chooseOp = p.restore()
					continue
				}
			}

			nowHeadState, ok := p.StateStack.Peek()
			if !ok {
				chooseOp = p.restore()
				continue
			}
			gotoState, ok := glr.LalrGoto[nowHeadState][newSym.Typ]
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
	for _, declare := range p.ExternalDeclarations {
		if declare.GetExternalDeclarationType() == entity.ExternalDeclarationTypeDeclaration {
			printDeclaration(declare)
		}
	}

	return nil
}

func (p *Parser) addCheckPoint(chooseOp int) {
	stateAll := p.StateStack.DumpAll()
	symAll := p.SymbolStack.DumpAll()
	cp := CheckPoint{
		TokenIndex:           p.TokenIndex,
		ChooseIndex:          chooseOp,
		StateStackSnap:       deepcopy.Copy(stateAll).([]int),
		SymbolStackSnap:      deepCopyAstNodeSlice(symAll),
		ExternalDeclarations: deepcopy.Copy(p.ExternalDeclarations).([]entity.ExternalDeclaration),
	}
	p.CheckPointStack.Push(&cp)
}

func deepCopyAstNodeSlice(origins []*glr.RawAstNode) []*glr.RawAstNode {
	// due to a node contains parent and children, it cannot be deepcopy.Copy(), or stack overflow
	var res []*glr.RawAstNode
	for _, origin := range origins {
		res = append(res, copyAstNode(origin))
	}
	for _, node := range res {
		fillAstParent(node, nil)
	}
	return res
}

func copyAstNode(origin *glr.RawAstNode) *glr.RawAstNode {
	root := &glr.RawAstNode{
		Typ:        origin.Typ,
		Terminal:   origin.Terminal,
		Production: origin.Production,
	}
	for _, child := range origin.Children {
		root.Children = append(root.Children, copyAstNode(child))
	}
	return root
}

func fillAstParent(node *glr.RawAstNode, parent *glr.RawAstNode) {
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
	p.TokenIndex = checkPoint.TokenIndex
	p.StateStack = common.NewStackWithElements[int](checkPoint.StateStackSnap)
	p.SymbolStack = common.NewStackWithElements[*glr.RawAstNode](checkPoint.SymbolStackSnap)
	p.ExternalDeclarations = checkPoint.ExternalDeclarations

	return checkPoint.ChooseIndex + 1
}

func printAST(ast *glr.RawAstNode, level int) {
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

func (p *Parser) parseExternalDeclaration(declare *glr.RawAstNode) error {
	switch {
	case declare.ReducedBy(glr.ExternalDeclaration, 1):
		// external_declaration := function_definition
		funcDef, err := parseFunctionDefinition(declare.Children[0])
		if err != nil {
			return err
		}
		p.ExternalDeclarations = append(p.ExternalDeclarations, funcDef)
	case declare.ReducedBy(glr.ExternalDeclaration, 2):
		// external_declaration := declaration
		res, err := parseDeclaration(declare.Children[0])
		if err != nil {
			return err
		}
		p.ExternalDeclarations = append(p.ExternalDeclarations, res)
	default:
		panic("unreachable")
	}
	return nil
}

func printDeclaration(unit entity.ExternalDeclaration) {
	declares := unit.(*entity.Declaration)
	for _, initDeclare := range declares.InitDeclarators {
		fmt.Printf("declare %s as ", initDeclare.Declarator.Identifier.Lexeme)
		typ := &initDeclare.Declarator.Type
		printType(typ)
		fmt.Println()
	}
}

func printType(typ *entity.Type) {
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
	case entity.MetaTypeVoid:
		print("void")
		return
	case entity.MetaTypeNumber:
		numMeta := typ.NumberMetaInfo
		if numMeta.Signed {
			print("signed ")
		}
		if numMeta.Unsigned {
			print("unsigned ")
		}
		switch numMeta.BaseNumType {
		case entity.BaseNumTypeChar:
			print("char")
		case entity.BaseNumTypeShort:
			print("short")
		case entity.BaseNumTypeInt:
			print("int")
		case entity.BaseNumTypeLong:
			print("long")
		case entity.BaseNumTypeFloat:
			print("float")
		case entity.BaseNumTypeDouble:
			print("double")
		case entity.BaseNumTypeBool:
			print("bool")
		case entity.BaseNumTypeLongLong:
			print("long long")
		case entity.BaseNumTypeLongDouble:
			print("long double")
		}
		return
	case entity.MetaTypeEnum:
	case entity.MetaTypePointer:
		print("pointer to ( ")
		printType(typ.PointerInnerType)
		print(" ) ")
	case entity.MetaTypeStruct:
	case entity.MetaTypeUnion:
	case entity.MetaTypeFunction:
		print("function with parameter ( ")
		for i, param := range typ.FunctionMetaInfo.Parameters {
			if i != 0 {
				print(" , ")
			}
			printType(&param.Type)
		}
		if len(typ.FunctionMetaInfo.IdentifierList) != 0 {
			var identifiers []string
			for _, identifier := range typ.FunctionMetaInfo.IdentifierList {
				identifiers = append(identifiers, identifier.Lexeme)
			}
			print(strings.Join(identifiers, ", "))
		}
		print(" ) ")
		if typ.FunctionMetaInfo.Variadic {
			print("... ")
		}
		print("and returning ( ")
		printType(typ.FunctionMetaInfo.ReturnType)
		print(" ) ")
	case entity.MetaTypeArray:
		print("array ")
		sizeExp := typ.ArrayMetaInfo.Size
		if sizeExp != nil {
			if sizeExp.ExpressionType == entity.ExpressionTypeConst {
				fmt.Printf("%+v ", sizeExp.Terminal.Literal)
			}
			if sizeExp.ExpressionType == entity.ExpressionTypeSizeOf {
				fmt.Printf("( with the size of type (")
				printType(&sizeExp.SizeOfExpressionInfo.Type)
				fmt.Printf(") ) ")
			}
		}
		print("of ( ")
		printType(typ.ArrayMetaInfo.InnerType)
		print(" ) ")
	case entity.MetaTypeUserDefined:
		fmt.Printf("user defined %s ", *typ.UserDefinedTypeName)
		return
	}

}
