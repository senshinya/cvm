package parser

import (
	"github.com/thoas/go-funk"
	"shinya.click/cvm/parser/entity"
	"shinya.click/cvm/parser/glr"
)

func ParseInitializer(root *entity.RawAstNode) (*entity.Initializer, error) {
	if err := root.AssertNonTerminal(glr.Initializer); err != nil {
		panic(err)
	}

	res := &entity.Initializer{}
	var err error
	switch {
	case root.ReducedBy(glr.Initializer, 1):
		// initializer := assignment_expression
		res.Type = entity.InitializerTypeExpression
		res.Expression, err = ParseExpressionNode(root.Children[0])
		if err != nil {
			return nil, err
		}
	case root.ReducedBy(glr.Initializer, 2), root.ReducedBy(glr.Initializer, 3):
		// initializer := LEFT_BRACES initializer_list RIGHT_BRACES
		// initializer := LEFT_BRACES initializer_list COMMA RIGHT_BRACES
		res.Type = entity.InitializerTypeStructOrArray
		res.InitializerList, err = ParseInitializerList(root.Children[1])
		if err != nil {
			return nil, err
		}
	default:
		panic("unreachable")
	}

	return res, nil
}

func ParseInitializerList(listNode *entity.RawAstNode) ([]*entity.InitializerItem, error) {
	if err := listNode.AssertNonTerminal(glr.InitializerList); err != nil {
		panic(err)
	}

	var (
		res []*entity.InitializerItem
		err error
	)
	for {
		endLoop := false
		switch {
		case listNode.ReducedBy(glr.InitializerList, 1):
			// initializer_list := initializer
			item := &entity.InitializerItem{}
			item.Initializer, err = ParseInitializer(listNode.Children[0])
			if err != nil {
				return nil, err
			}
			res = append(res, item)
			endLoop = true
		case listNode.ReducedBy(glr.InitializerList, 2):
			// initializer_list := designation initializer
			item := &entity.InitializerItem{}
			item.Designators, err = ParseDesignation(listNode.Children[0])
			if err != nil {
				return nil, err
			}
			item.Initializer, err = ParseInitializer(listNode.Children[1])
			if err != nil {
				return nil, err
			}
			res = append(res, item)
			endLoop = true
		case listNode.ReducedBy(glr.InitializerList, 3):
			// initializer_list := initializer_list COMMA initializer
			item := &entity.InitializerItem{}
			item.Initializer, err = ParseInitializer(listNode.Children[2])
			if err != nil {
				return nil, err
			}
			res = append(res, item)
			listNode = listNode.Children[0]
		case listNode.ReducedBy(glr.InitializerList, 4):
			// initializer_list := initializer_list COMMA designation initializer
			item := &entity.InitializerItem{}
			item.Designators, err = ParseDesignation(listNode.Children[2])
			if err != nil {
				return nil, err
			}
			item.Initializer, err = ParseInitializer(listNode.Children[3])
			if err != nil {
				return nil, err
			}
			res = append(res, item)
			listNode = listNode.Children[0]
		default:
			panic("unreachable")
		}
		if endLoop {
			break
		}
	}
	return funk.Reverse(res).([]*entity.InitializerItem), nil
}

func ParseDesignation(root *entity.RawAstNode) ([]*entity.Designator, error) {
	if err := root.AssertNonTerminal(glr.Designation); err != nil {
		panic(err)
	}

	// designation := designator_list EQUAL
	listNode := root.Children[0]
	var res []*entity.Designator
	designatorNodes := flattenDesignatorList(listNode)
	for _, designatorNode := range designatorNodes {
		switch {
		case designatorNode.ReducedBy(glr.Designator, 1):
			// designator := LEFT_BRACKETS constant_expression RIGHT_BRACKETS
			exp, err := ParseExpressionNode(designatorNode.Children[1])
			if err != nil {
				return nil, err
			}
			res = append(res, &entity.Designator{
				Expression: exp,
			})
		case designatorNode.ReducedBy(glr.Designator, 2):
			// designator := PERIOD IDENTIFIER
			res = append(res, &entity.Designator{
				Identifier: &designatorNode.Children[1].Terminal.Lexeme,
			})
		default:
			panic("unreachable")
		}
	}
	return res, nil
}
