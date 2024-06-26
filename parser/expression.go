package parser

import (
	"shinya.click/cvm/common"
	"shinya.click/cvm/parser/entity"
	"shinya.click/cvm/parser/glr"
)

func ParseExpression(node *glr.RawAstNode) (*entity.Expression, error) {
	if node.Typ == common.IDENTIFIER {
		return &entity.Expression{
			ExpressionType: entity.ExpressionTypeIdentifier,
			Terminal:       node.Terminal,
			SourceRange:    node.GetSourceRange(),
		}, nil
	}
	if node.Typ == common.STRING || node.Typ == common.CHARACTER ||
		node.Typ == common.INTEGER_CONSTANT || node.Typ == common.FLOATING_CONSTANT {
		return &entity.Expression{
			ExpressionType: entity.ExpressionTypeConst,
			Terminal:       node.Terminal,
			SourceRange:    node.GetSourceRange(),
		}, nil
	}

	// production like assignment_expression := conditional_expression
	if len(node.Children) == 1 {
		return ParseExpression(node.Children[0])
	}

	switch {
	case node.ReducedBy(glr.AssignmentExpression, 2):
		// assignment_expression := unary_expression assignment_operator assignment_expression
		lv, err := ParseExpression(node.Children[0])
		if err != nil {
			return nil, err
		}
		rv, err := ParseExpression(node.Children[2])
		if err != nil {
			return nil, err
		}
		return &entity.Expression{
			ExpressionType: entity.ExpressionTypeAssignment,
			AssignmentExpressionInfo: &entity.AssignmentExpressionInfo{
				LValue:   lv,
				Operator: node.Children[1].Children[0].Typ,
				RValue:   rv,
			},
			SourceRange: node.GetSourceRange(),
		}, nil
	case node.ReducedBy(glr.ConditionalExpression, 2):
		// conditional_expression := logical_or_expression QUESTION expression COLON conditional_expression
		cond, err := ParseExpression(node.Children[0])
		if err != nil {
			return nil, err
		}
		tr, err := ParseExpression(node.Children[2])
		if err != nil {
			return nil, err
		}
		fa, err := ParseExpression(node.Children[4])
		if err != nil {
			return nil, err
		}
		return &entity.Expression{
			ExpressionType: entity.ExpressionTypeCondition,
			ConditionExpressionInfo: &entity.ConditionExpressionInfo{
				Condition:   cond,
				TrueBranch:  tr,
				FalseBranch: fa,
			},
			SourceRange: node.GetSourceRange(),
		}, nil
	case node.ReducedBy(glr.LogicalOrExpression, 2),
		node.ReducedBy(glr.LogicalAndExpression, 2),
		node.ReducedBy(glr.EqualityExpression, 2, 3),
		node.ReducedBy(glr.RelationalExpression, 2, 3, 4, 5):
		// logical_or_expression := logical_or_expression OR_OR logical_and_expression
		one, err := ParseExpression(node.Children[0])
		if err != nil {
			return nil, err
		}
		two, err := ParseExpression(node.Children[2])
		if err != nil {
			return nil, err
		}
		return &entity.Expression{
			ExpressionType: entity.ExpressionTypeLogic,
			LogicExpressionInfo: &entity.LogicExpressionInfo{
				Operator: node.Children[1].Typ,
				One:      one,
				Two:      two,
			},
			SourceRange: node.GetSourceRange(),
		}, nil
	case node.ReducedBy(glr.InclusiveOrExpression, 2),
		node.ReducedBy(glr.AndExpression, 2),
		node.ReducedBy(glr.ExclusiveOrExpression, 2),
		node.ReducedBy(glr.ShiftExpression, 2, 3):
		// inclusive_or_expression := inclusive_or_expression OR exclusive_or_expression
		one, err := ParseExpression(node.Children[0])
		if err != nil {
			return nil, err
		}
		two, err := ParseExpression(node.Children[2])
		if err != nil {
			return nil, err
		}
		return &entity.Expression{
			ExpressionType: entity.ExpressionTypeBit,
			BitExpressionInfo: &entity.BitExpressionInfo{
				Operator: node.Children[1].Typ,
				One:      one,
				Two:      two,
			},
			SourceRange: node.GetSourceRange(),
		}, nil
	case node.ReducedBy(glr.AdditiveExpression, 2, 3),
		node.ReducedBy(glr.MultiplicativeExpression, 2, 3, 4):
		// additive_expression := additive_expression PLUS multiplicative_expression
		one, err := ParseExpression(node.Children[0])
		if err != nil {
			return nil, err
		}
		two, err := ParseExpression(node.Children[2])
		if err != nil {
			return nil, err
		}
		return &entity.Expression{
			ExpressionType: entity.ExpressionTypeNumber,
			NumberExpressionInfo: &entity.NumberExpressionInfo{
				Operator: node.Children[1].Typ,
				One:      one,
				Two:      two,
			},
			SourceRange: node.GetSourceRange(),
		}, nil
	case node.ReducedBy(glr.CastExpression, 2):
		// cast_expression := LEFT_PARENTHESES type_name RIGHT_PARENTHESES cast_expression
		typ, err := ParseTypeName(node.Children[1])
		if err != nil {
			return nil, err
		}
		source, err := ParseExpression(node.Children[3])
		if err != nil {
			return nil, err
		}
		return &entity.Expression{
			ExpressionType: entity.ExpressionTypeCast,
			CastExpressionInfo: &entity.CastExpressionInfo{
				Type:   typ,
				Source: source,
			},
			SourceRange: node.GetSourceRange(),
		}, nil
	case node.ReducedBy(glr.PrimaryExpression, 6):
		// primary_expression := LEFT_PARENTHESES expression RIGHT_PARENTHESES
		return ParseExpression(node.Children[1])
	case node.ReducedBy(glr.Expression, 2):
		// expression := expression COMMA assignment_expression
		var exps []*entity.Expression
		for _, n := range flattenExpression(node) {
			exp, err := ParseExpression(n)
			if err != nil {
				return nil, err
			}
			exps = append(exps, exp)
		}
		if len(exps) == 1 {
			return exps[0], nil
		}
		return &entity.Expression{
			ExpressionType: entity.ExpressionTypeExpressions,
			Expressions:    exps,
			SourceRange:    node.GetSourceRange(),
		}, nil
	case node.Typ == glr.UnaryExpression:
		return parseUnary(node)
	case node.Typ == glr.PostfixExpression:
		return parsePostfix(node)
	default:
		panic("unreachable")
	}
}

func parsePostfix(node *glr.RawAstNode) (*entity.Expression, error) {
	if err := node.AssertNonTerminal(glr.PostfixExpression); err != nil {
		panic(err)
	}

	switch {
	case node.ReducedBy(glr.PostfixExpression, 7, 8):
		// postfix_expression := postfix_expression PLUS_PLUS
		// postfix_expression := postfix_expression MINUS_MINUS
		target, err := ParseExpression(node.Children[0])
		if err != nil {
			return nil, err
		}
		return &entity.Expression{
			ExpressionType: entity.ExpressionTypePostfix,
			PostfixExpressionInfo: &entity.PostfixExpressionInfo{
				Operator: node.Children[1].Typ,
				Target:   target,
			},
			SourceRange: node.GetSourceRange(),
		}, nil
	case node.ReducedBy(glr.PostfixExpression, 2):
		// postfix_expression := postfix_expression LEFT_BRACKETS expression RIGHT_BRACKETS
		arr, err := ParseExpression(node.Children[0])
		if err != nil {
			return nil, err
		}
		target, err := ParseExpression(node.Children[2])
		if err != nil {
			return nil, err
		}
		return &entity.Expression{
			ExpressionType: entity.ExpressionTypeArrayAccess,
			ArrayAccessExpressionInfo: &entity.ArrayAccessExpressionInfo{
				Array:  arr,
				Target: target,
			},
			SourceRange: node.GetSourceRange(),
		}, nil
	case node.ReducedBy(glr.PostfixExpression, 3):
		// postfix_expression := postfix_expression LEFT_PARENTHESES RIGHT_PARENTHESES
		fun, err := ParseExpression(node.Children[0])
		if err != nil {
			return nil, err
		}
		return &entity.Expression{
			ExpressionType: entity.ExpressionTypeFunctionCall,
			FunctionCallExpressionInfo: &entity.FunctionCallExpressionInfo{
				Function: fun,
			},
			SourceRange: node.GetSourceRange(),
		}, nil
	case node.ReducedBy(glr.PostfixExpression, 4):
		// postfix_expression := postfix_expression LEFT_PARENTHESES argument_expression_list RIGHT_PARENTHESES
		fun, err := ParseExpression(node.Children[0])
		if err != nil {
			return nil, err
		}
		arguments, err := parseArgumentExpressionList(node.Children[2])
		if err != nil {
			return nil, err
		}
		return &entity.Expression{
			ExpressionType: entity.ExpressionTypeFunctionCall,
			FunctionCallExpressionInfo: &entity.FunctionCallExpressionInfo{
				Function:  fun,
				Arguments: arguments,
			},
			SourceRange: node.GetSourceRange(),
		}, nil
	case node.ReducedBy(glr.PostfixExpression, 5, 6):
		// postfix_expression := postfix_expression PERIOD IDENTIFIER
		// postfix_expression := postfix_expression ARROW IDENTIFIER
		str, err := ParseExpression(node.Children[0])
		if err != nil {
			return nil, err
		}
		return &entity.Expression{
			ExpressionType: entity.ExpressionTypeStructAccess,
			StructAccessExpressionInfo: &entity.StructAccessExpressionInfo{
				Struct: str,
				Field:  node.Children[2].Terminal.Lexeme,
				Access: node.Children[1].Typ,
			},
			SourceRange: node.GetSourceRange(),
		}, nil
	case node.ReducedBy(glr.PostfixExpression, 9, 10):
		// postfix_expression := LEFT_PARENTHESES type_name RIGHT_PARENTHESES LEFT_BRACES initializer_list RIGHT_BRACES
		// postfix_expression := LEFT_PARENTHESES type_name RIGHT_PARENTHESES LEFT_BRACES initializer_list COMMA RIGHT_BRACES
		typ, err := ParseTypeName(node.Children[1])
		if err != nil {
			return nil, err
		}
		initList, err := ParseInitializerList(node.Children[4])
		if err != nil {
			return nil, err
		}
		return &entity.Expression{
			ExpressionType: entity.ExpressionTypeConstruct,
			ConstructExpressionInfo: &entity.ConstructExpressionInfo{
				Type:         typ,
				Initializers: initList,
			},
			SourceRange: node.GetSourceRange(),
		}, nil
	default:
		panic("unreachable")
	}
}

func parseUnary(node *glr.RawAstNode) (*entity.Expression, error) {
	if err := node.AssertNonTerminal(glr.UnaryExpression); err != nil {
		panic(err)
	}

	switch {
	case node.ReducedBy(glr.UnaryExpression, 6):
		// unary_expression := SIZEOF LEFT_PARENTHESES type_name RIGHT_PARENTHESES
		typ, err := ParseTypeName(node.Children[2])
		if err != nil {
			return nil, err
		}
		return &entity.Expression{
			ExpressionType: entity.ExpressionTypeSizeOf,
			SizeOfExpressionInfo: &entity.SizeOfExpressionInfo{
				Type: typ,
			},
			SourceRange: node.GetSourceRange(),
		}, nil
	case node.ReducedBy(glr.UnaryExpression, 5):
		// unary_expression := SIZEOF unary_expression
		target, err := ParseExpression(node.Children[1])
		if err != nil {
			return nil, err
		}
		return &entity.Expression{
			ExpressionType: entity.ExpressionTypeSizeOf,
			SizeOfExpressionInfo: &entity.SizeOfExpressionInfo{
				Target: target,
			},
			SourceRange: node.GetSourceRange(),
		}, nil
	case node.ReducedBy(glr.UnaryExpression, 4):
		// unary_expression := unary_operator cast_expression
		target, err := ParseExpression(node.Children[1])
		if err != nil {
			return nil, err
		}
		return &entity.Expression{
			ExpressionType: entity.ExpressionTypeUnary,
			UnaryExpressionInfo: &entity.UnaryExpressionInfo{
				Operator: node.Children[0].Children[0].Typ,
				Target:   target,
			},
			SourceRange: node.GetSourceRange(),
		}, nil
	case node.ReducedBy(glr.UnaryExpression, 2, 3):
		// unary_expression := PLUS_PLUS unary_expression
		// unary_expression := MINUS_MINUS unary_expression
		target, err := ParseExpression(node.Children[1])
		if err != nil {
			return nil, err
		}
		return &entity.Expression{
			ExpressionType: entity.ExpressionTypeUnary,
			UnaryExpressionInfo: &entity.UnaryExpressionInfo{
				Operator: node.Children[0].Typ,
				Target:   target,
			},
			SourceRange: node.GetSourceRange(),
		}, nil
	default:
		panic("unreachable")
	}
}

func parseArgumentExpressionList(root *glr.RawAstNode) ([]*entity.Expression, error) {
	if err := root.AssertNonTerminal(glr.ArgumentExpressionList); err != nil {
		panic(err)
	}

	var res []*entity.Expression
	expNodes := flattenArgumentExpressions(root)
	for _, expNode := range expNodes {
		exp, err := ParseExpression(expNode)
		if err != nil {
			return nil, err
		}
		res = append(res, exp)
	}

	return res, nil
}
