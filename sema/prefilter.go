package sema

import (
	"fmt"

	"shinya.click/cvm/common"
	"shinya.click/cvm/entity"
	"shinya.click/cvm/parser"
)

// PreFilter 遍历每棵候选语法树，丢弃 typedef 使用方式明显不一致的候选。
// 被剔除候选产生的错误会一并返回，调用方可把它们纳入最佳错误池。
func PreFilter(forest []*entity.AstNode) (survivors []*entity.AstNode, errs []*common.CvmError) {
	for _, root := range forest {
		err := newTimberSaw().chop(root)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		survivors = append(survivors, root)
	}
	return survivors, errs
}

type timberSaw struct {
	symbolStack *symbolStack
}

type symbolStack struct {
	stack []*scopeSymbols
	depth int
}

type scopeSymbols struct {
	TypeNames map[string]*entity.Token
	VarNames  map[string]*entity.Token
}

func newScopeSymbols() *scopeSymbols {
	return &scopeSymbols{
		TypeNames: map[string]*entity.Token{},
		VarNames:  map[string]*entity.Token{},
	}
}

func newTimberSaw() *timberSaw {
	s := &timberSaw{symbolStack: &symbolStack{}}
	s.symbolStack.stack = append(s.symbolStack.stack, newScopeSymbols())
	return s
}

func (s *symbolStack) resize(length int) {
	for len(s.stack) < length {
		s.stack = append(s.stack, newScopeSymbols())
	}
	if len(s.stack) > length {
		s.stack = s.stack[:length]
	}
}

func (s *symbolStack) enterScope(depth int) {
	s.resize(depth + 1)
	s.depth = depth
}

func (s *symbolStack) switchScope(depth int) {
	if len(s.stack) < depth+1 {
		s.resize(depth + 1)
	}
	s.depth = depth
}

func (s *symbolStack) pushVar(token *entity.Token) *common.CvmError {
	name := token.Lexeme
	if previous, ok := s.stack[s.depth].TypeNames[name]; ok {
		return invalidSymbolKind(token.SourceStart, previous.SourceStart, name)
	}
	s.stack[s.depth].VarNames[name] = token
	return nil
}

func (s *symbolStack) pushType(token *entity.Token) *common.CvmError {
	name := token.Lexeme
	if previous, ok := s.stack[s.depth].VarNames[name]; ok {
		return invalidSymbolKind(token.SourceStart, previous.SourceStart, name)
	}
	s.stack[s.depth].TypeNames[name] = token
	return nil
}

func (s *symbolStack) isTypeName(name string) (bool, *entity.Token) {
	for i := s.depth; i >= 0; i-- {
		if t, ok := s.stack[i].TypeNames[name]; ok {
			return true, t
		}
		if _, ok := s.stack[i].VarNames[name]; ok {
			return false, nil
		}
	}
	return false, nil
}

func (s *timberSaw) chop(node *entity.AstNode) *common.CvmError {
	switch node.Typ {
	case parser.FunctionDefinition:
		declarator := node.Children[1]
		for _, token := range declarator.DeclaratorID {
			if err := s.symbolStack.pushVar(token); err != nil {
				return err
			}
		}
		for i := 0; i < 2; i++ {
			if err := s.chop(node.Children[i]); err != nil {
				return err
			}
		}
		for _, param := range getParametersFromFunctionDeclarator(declarator) {
			depth := s.symbolStack.depth
			s.symbolStack.enterScope(depth + 1)
			if err := s.symbolStack.pushVar(param); err != nil {
				return err
			}
			s.symbolStack.enterScope(depth)
		}
		current := s.symbolStack.depth
		if err := s.chop(node.Children[len(node.Children)-1]); err != nil {
			return err
		}
		s.symbolStack.enterScope(current)

	case parser.Declaration:
		typedef := node.Children[0].TypeDef
		ids := node.Children[0].DeclaratorID
		if node.ReducedBy(parser.Declaration, 2) {
			ids = append(ids, node.Children[1].DeclaratorID...)
		}
		for _, id := range ids {
			if typedef {
				if err := s.symbolStack.pushType(id); err != nil {
					return err
				}
			} else {
				if err := s.symbolStack.pushVar(id); err != nil {
					return err
				}
			}
		}
		for _, child := range node.Children {
			if err := s.chop(child); err != nil {
				return err
			}
		}

	case parser.IterationStatement:
		if node.ReducedBy(parser.IterationStatement, 1) || node.ReducedBy(parser.IterationStatement, 2) {
			for _, child := range node.Children {
				if err := s.chop(child); err != nil {
					return err
				}
			}
			break
		}
		current := s.symbolStack.depth
		s.symbolStack.switchScope(current + 1)
		for i := 0; i < len(node.Children)-1; i++ {
			if err := s.chop(node.Children[i]); err != nil {
				return err
			}
		}
		s.symbolStack.switchScope(current)
		if err := s.chop(node.Children[len(node.Children)-1]); err != nil {
			return err
		}
		s.symbolStack.enterScope(current)

	case entity.LEFT_BRACES:
		s.symbolStack.enterScope(s.symbolStack.depth + 1)
	case entity.RIGHT_BRACES:
		s.symbolStack.enterScope(s.symbolStack.depth - 1)

	case parser.PrimaryExpression:
		if node.ReducedBy(parser.PrimaryExpression, 1) {
			name := node.Children[0].Terminal.Lexeme
			isType, prev := s.symbolStack.isTypeName(name)
			if isType {
				return invalidSymbolKind(node.Children[0].Terminal.SourceStart, prev.SourceStart, name)
			}
		}
		for _, child := range node.Children {
			if err := s.chop(child); err != nil {
				return err
			}
		}

	case parser.TypedefName:
		name := node.Children[0].Terminal.Lexeme
		isType, _ := s.symbolStack.isTypeName(name)
		if !isType {
			return UndeclaredIdentifier(node.Children[0].Terminal.SourceStart, name)
		}

	default:
		for _, child := range node.Children {
			if err := s.chop(child); err != nil {
				return err
			}
		}
	}
	return nil
}

func getParametersFromFunctionDeclarator(declarator *entity.AstNode) []*entity.Token {
	direct := declarator.Children[len(declarator.Children)-1]
	switch {
	case direct.ReducedBy(parser.DirectDeclarator, 12):
		return direct.Children[2].DeclaratorID
	case direct.ReducedBy(parser.DirectDeclarator, 13):
		return nil
	case direct.ReducedBy(parser.DirectDeclarator, 14):
		return direct.Children[2].DeclaratorID
	}
	panic(fmt.Sprintf("unexpected direct_declarator: %s", direct.Production))
}

func invalidSymbolKind(pos, previous entity.SourcePos, name string) *common.CvmError {
	return InvalidSymbolKind(pos, previous, name)
}
