package parser

import (
	"fmt"
	"shinya.click/cvm/common"
	"shinya.click/cvm/entity"
)

func chopForest(forest []*entity.AstNode) ([]*entity.AstNode, error) {
	var (
		latestErr error
		result    []*entity.AstNode
	)
	for _, root := range forest {
		err := NewTimberSaw().Chop(root)
		if err != nil {
			printAST(root, 0)
			common.DebugPrintf("Chop Error: %s\n\n", err)
			latestErr = err
			continue
		}
		result = append(result, root)
	}
	return result, latestErr
}

type TimberSaw struct {
	symbolStack *symbolStack
}

type symbolStack struct {
	stack                   []*ScopeSymbols
	currentSymbolStackDepth int
}

type ScopeSymbols struct {
	TypeNames map[string]*entity.Token
	VarNames  map[string]*entity.Token
}

func (s *symbolStack) resize(length int) {
	if len(s.stack) < length {
		for i := len(s.stack); i < length; i++ {
			s.stack = append(s.stack, NewScopeSymbols())
		}
	}
	if len(s.stack) > length {
		s.stack = s.stack[:length]
	}
}

// EnterScope resize the stack to depth+1
func (s *symbolStack) EnterScope(depth int) {
	s.resize(depth + 1)
	s.currentSymbolStackDepth = depth
}

// SwitchScope only expand but not shrink the stack
func (s *symbolStack) SwitchScope(depth int) {
	if len(s.stack) < depth+1 {
		s.resize(depth + 1)
	}
	s.currentSymbolStackDepth = depth
}

func (s *symbolStack) PushVar(token *entity.Token, depth int) error {
	name := token.Lexeme
	if len(s.stack) < depth+1 {
		s.resize(depth + 1)
	}
	if previous, ok := s.stack[depth].TypeNames[name]; ok {
		return RedefinitionSymbol(token.SourceStart, previous.SourceStart, token.Lexeme)
	}
	s.stack[depth].VarNames[name] = token
	return nil
}

func (s *symbolStack) PushType(token *entity.Token, depth int) error {
	name := token.Lexeme
	if len(s.stack) < depth+1 {
		s.resize(depth + 1)
	}
	if previous, ok := s.stack[depth].VarNames[name]; ok {
		return RedefinitionSymbol(token.SourceStart, previous.SourceStart, token.Lexeme)
	}
	s.stack[depth].TypeNames[name] = token
	return nil
}

func (s *symbolStack) CheckVar(token *entity.Token, depth int) error {
	for i := depth; i >= 0; i-- {
		if previous, ok := s.stack[i].TypeNames[token.Lexeme]; ok {
			return InvalidSymbolKind(token.SourceStart, previous.SourceStart, token.Lexeme)
		}
		if _, ok := s.stack[i].VarNames[token.Lexeme]; ok {
			return nil
		}
	}
	return UndeclaredIdentifier(token.SourceStart, token.Lexeme)
}

func (s *symbolStack) CheckType(token *entity.Token, depth int) error {
	for i := depth; i >= 0; i-- {
		if previous, ok := s.stack[i].VarNames[token.Lexeme]; ok {
			return InvalidSymbolKind(token.SourceStart, previous.SourceStart, token.Lexeme)
		}
		if _, ok := s.stack[i].TypeNames[token.Lexeme]; ok {
			return nil
		}
	}
	return UndeclaredIdentifier(token.SourceStart, token.Lexeme)
}

func NewScopeSymbols() *ScopeSymbols {
	return &ScopeSymbols{
		TypeNames: map[string]*entity.Token{},
		VarNames:  map[string]*entity.Token{},
	}
}

func NewTimberSaw() *TimberSaw {
	s := &TimberSaw{
		symbolStack: &symbolStack{currentSymbolStackDepth: 0},
	}
	s.symbolStack.stack = append(s.symbolStack.stack, NewScopeSymbols())
	return s
}

func (s *TimberSaw) Chop(node *entity.AstNode) error {
	switch node.Typ {
	case FunctionDefinition:
		declarator := node.Children[1]
		for _, token := range declarator.DeclaratorID {
			if err := s.symbolStack.PushVar(token, s.symbolStack.currentSymbolStackDepth); err != nil {
				return err
			}
		}
		for i := 0; i < len(node.Children)-1; i++ {
			if err := s.Chop(node.Children[i]); err != nil {
				return err
			}
		}
		// compound_statement should be checked after function parameters parsed
		params := getParametersFromFunctionDeclarator(declarator)
		for _, param := range params {
			if err := s.symbolStack.PushVar(param, s.symbolStack.currentSymbolStackDepth+1); err != nil {
				return err
			}
		}
		if err := s.Chop(node.Children[len(node.Children)-1]); err != nil {
			return err
		}
	case Declaration:
		typedef := node.Children[0].TypeDef
		ids := node.Children[0].DeclaratorID // enum constants in declaration_specifiers
		if node.ReducedBy(Declaration, 2) {
			// declaration := declaration_specifiers init_declarator_list SEMICOLON
			ids = append(ids, node.Children[1].DeclaratorID...)
		}
		for _, id := range ids {
			if typedef {
				if err := s.symbolStack.PushType(id, s.symbolStack.currentSymbolStackDepth); err != nil {
					return err
				}
			} else {
				if err := s.symbolStack.PushVar(id, s.symbolStack.currentSymbolStackDepth); err != nil {
					return err
				}
			}
		}
		for _, child := range node.Children {
			if err := s.Chop(child); err != nil {
				return err
			}
		}
	case IterationStatement:
		if node.ReducedBy(IterationStatement, 1) || node.ReducedBy(IterationStatement, 2) {
			// while/do while
			for _, child := range node.Children {
				if err := s.Chop(child); err != nil {
					return err
				}
			}
		}
		// for
		currentSymbolStackDepth := s.symbolStack.currentSymbolStackDepth
		s.symbolStack.SwitchScope(currentSymbolStackDepth + 1)
		for i := 0; i < len(node.Children)-1; i++ {
			if err := s.Chop(node.Children[i]); err != nil {
				return err
			}
		}
		s.symbolStack.SwitchScope(currentSymbolStackDepth)
		if err := s.Chop(node.Children[len(node.Children)-1]); err != nil {
			return err
		}
	case entity.LEFT_BRACES:
		s.symbolStack.EnterScope(s.symbolStack.currentSymbolStackDepth + 1)
	case entity.RIGHT_BRACES:
		s.symbolStack.EnterScope(s.symbolStack.currentSymbolStackDepth - 1)
	case PrimaryExpression:
		if node.ReducedBy(PrimaryExpression, 1) {
			// primary_expression := IDENTIFIER
			if err := s.symbolStack.CheckVar(node.Children[0].Terminal, s.symbolStack.currentSymbolStackDepth); err != nil {
				return err
			}
		}
		for _, child := range node.Children {
			if err := s.Chop(child); err != nil {
				return err
			}
		}
	case TypedefName:
		// typedef_name := IDENTIFIER
		if err := s.symbolStack.CheckType(node.Children[0].Terminal, s.symbolStack.currentSymbolStackDepth); err != nil {
			return err
		}
	default:
		for _, child := range node.Children {
			if err := s.Chop(child); err != nil {
				return err
			}
		}
	}
	return nil
}

func getParametersFromFunctionDeclarator(declarator *entity.AstNode) []*entity.Token {
	// declarator := direct_declarator
	// declarator := pointer direct_declarator
	direct := declarator.Children[len(declarator.Children)-1]
	switch {
	case direct.ReducedBy(DirectDeclarator, 12):
		// direct_declarator := direct_declarator LEFT_PARENTHESES parameter_type_list RIGHT_PARENTHESES
		return direct.Children[2].DeclaratorID
	case direct.ReducedBy(DirectDeclarator, 13):
		// direct_declarator := direct_declarator LEFT_PARENTHESES RIGHT_PARENTHESES
		return nil
	case direct.ReducedBy(DirectDeclarator, 14):
		// direct_declarator := direct_declarator LEFT_PARENTHESES identifier_list RIGHT_PARENTHESES
		return direct.Children[2].DeclaratorID
	default:
		panic(fmt.Sprintf("getParametersFromFunctionDeclarator unexpected direct_declarator %s", direct.Production))
	}
}
