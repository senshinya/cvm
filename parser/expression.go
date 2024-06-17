package parser

import (
	"shinya.click/cvm/common"
	"shinya.click/cvm/parser/entity"
	"shinya.click/cvm/parser/glr"
)

func ParseExpressionNode(node *entity.AstNode) *entity.SingleExpression {
	return SimplifyExpression(parseExpressionNodeInner(node))
}

func parseExpressionNodeInner(node *entity.AstNode) *entity.SingleExpression {
	if node.Typ == common.IDENTIFIER {
		return &entity.SingleExpression{
			ExpressionType: entity.ExpressionTypeIdentifier,
			Terminal:       node.Terminal,
		}
	}
	if node.Typ == common.STRING || node.Typ == common.CHARACTER ||
		node.Typ == common.INTEGER_CONSTANT || node.Typ == common.FLOATING_CONSTANT {
		return &entity.SingleExpression{
			ExpressionType: entity.ExpressionTypeConst,
			Terminal:       node.Terminal,
		}
	}

	prod := glr.Productions[node.ProdIndex]
	if len(prod.Right) == 1 {
		return parseExpressionNodeInner(node.Children[0])
	}

	switch node.Typ {
	case glr.AssignmentExpression:
		// assignment_expression := unary_expression assignment_operator assignment_expression
		return &entity.SingleExpression{
			ExpressionType: entity.ExpressionTypeAssignment,
			AssignmentExpressionInfo: &entity.AssignmentExpressionInfo{
				LValue:   parseExpressionNodeInner(node.Children[0]),
				Operator: node.Children[1].Children[0].Typ,
				RValue:   parseExpressionNodeInner(node.Children[2]),
			},
		}
	case glr.ConditionalExpression:
		// conditional_expression := logical_or_expression QUESTION expression COLON conditional_expression
		return &entity.SingleExpression{
			ExpressionType: entity.ExpressionTypeCondition,
			ConditionExpressionInfo: &entity.ConditionExpressionInfo{
				Condition:   parseExpressionNodeInner(node.Children[0]),
				TrueBranch:  parseExpressionNodeInner(node.Children[2]),
				FalseBranch: parseExpressionNodeInner(node.Children[4]),
			},
		}
	case glr.LogicalOrExpression, glr.LogicalAndExpression,
		glr.EqualityExpression, glr.RelationalExpression:
		// logical_or_expression := logical_or_expression OR_OR logical_and_expression
		return &entity.SingleExpression{
			ExpressionType: entity.ExpressionTypeLogic,
			LogicExpressionInfo: &entity.LogicExpressionInfo{
				Operator: node.Children[1].Typ,
				One:      parseExpressionNodeInner(node.Children[0]),
				Two:      parseExpressionNodeInner(node.Children[2]),
			},
		}
	case glr.InclusiveOrExpression, glr.AndExpression,
		glr.ShiftExpression, glr.ExclusiveOrExpression:
		// inclusive_or_expression := inclusive_or_expression OR exclusive_or_expression
		return &entity.SingleExpression{
			ExpressionType: entity.ExpressionTypeBit,
			BitExpressionInfo: &entity.BitExpressionInfo{
				Operator: node.Children[1].Typ,
				One:      parseExpressionNodeInner(node.Children[0]),
				Two:      parseExpressionNodeInner(node.Children[2]),
			},
		}
	case glr.AdditiveExpression, glr.MultiplicativeExpression:
		// additive_expression := additive_expression PLUS multiplicative_expression
		return &entity.SingleExpression{
			ExpressionType: entity.ExpressionTypeNumber,
			NumberExpressionInfo: &entity.NumberExpressionInfo{
				Operator: node.Children[1].Typ,
				One:      parseExpressionNodeInner(node.Children[0]),
				Two:      parseExpressionNodeInner(node.Children[2]),
			},
		}
	case glr.CastExpression:
		// cast_expression := LEFT_PARENTHESES type_name RIGHT_PARENTHESES cast_expression
		return &entity.SingleExpression{
			ExpressionType: entity.ExpressionTypeCast,
			CastExpressionInfo: &entity.CastExpressionInfo{
				Type:   ParseTypeName(node.Children[1]),
				Source: parseExpressionNodeInner(node.Children[3]),
			},
		}
	case glr.UnaryExpression:
		return parseUnary(node)
	case glr.PostfixExpression:
		return parsePostfix(node)
	case glr.PrimaryExpression:
		// primary_expression := LEFT_PARENTHESES expression RIGHT_PARENTHESES
		return parseExpressionNodeInner(node.Children[1])
	case glr.Expression:
		// expression := expression COMMA assignment_expression
		var exps []*entity.SingleExpression
		for _, n := range flattenExpression(node) {
			exps = append(exps, parseExpressionNodeInner(n))
		}
		if len(exps) == 1 {
			return exps[0]
		}
		return &entity.SingleExpression{
			ExpressionType: entity.ExpressionTypeExpressions,
			Expressions:    exps,
		}
	}
	panic("should not happen")
}

func parsePostfix(node *entity.AstNode) *entity.SingleExpression {
	prod := glr.Productions[node.ProdIndex]
	if len(prod.Right) == 2 {
		// postfix_expression := postfix_expression PLUS_PLUS
		return &entity.SingleExpression{
			ExpressionType: entity.ExpressionTypePostfix,
			PostfixExpressionInfo: &entity.PostfixExpressionInfo{
				Operator: node.Children[1].Typ,
				Target:   parseExpressionNodeInner(node.Children[0]),
			},
		}
	}
	if prod.Right[1] == common.LEFT_BRACKETS {
		// postfix_expression := postfix_expression LEFT_BRACKETS expression RIGHT_BRACKETS
		return &entity.SingleExpression{
			ExpressionType: entity.ExpressionTypeArrayAccess,
			ArrayAccessExpressionInfo: &entity.ArrayAccessExpressionInfo{
				Array:  parseExpressionNodeInner(node.Children[0]),
				Target: parseExpressionNodeInner(node.Children[2]),
			},
		}
	}
	if prod.Right[1] == common.LEFT_PARENTHESES {
		// TODO function call
		return &entity.SingleExpression{
			ExpressionType: entity.ExpressionTypeFunctionCall,
			FunctionCallExpressionInfo: &entity.FunctionCallExpressionInfo{
				Function: parseExpressionNodeInner(node.Children[0]),
			},
		}
	}
	if prod.Right[1] == common.PERIOD ||
		prod.Right[1] == common.ARROW {
		return &entity.SingleExpression{
			ExpressionType: entity.ExpressionTypeStructAccess,
			StructAccessExpressionInfo: &entity.StructAccessExpressionInfo{
				Struct: parseExpressionNodeInner(node.Children[0]),
				Field:  node.Children[1].Terminal.Lexeme,
				Access: prod.Right[1],
			},
		}
	}
	// postfix_expression := LEFT_PARENTHESES type_name RIGHT_PARENTHESES LEFT_BRACES initializer_list RIGHT_BRACES
	// TODO initializer_list
	return &entity.SingleExpression{
		ExpressionType: entity.ExpressionTypeConstruct,
		ConstructExpressionInfo: &entity.ConstructExpressionInfo{
			Type: ParseTypeName(node.Children[1]),
		},
	}
}

func parseUnary(node *entity.AstNode) *entity.SingleExpression {
	prod := glr.Productions[node.ProdIndex]
	if len(prod.Right) == 4 {
		// unary_expression := SIZEOF LEFT_PARENTHESES type_name RIGHT_PARENTHESES
		return &entity.SingleExpression{
			ExpressionType: entity.ExpressionTypeSizeOf,
			SizeOfExpressionInfo: &entity.SizeOfExpressionInfo{
				Type: ParseTypeName(node.Children[2]),
			},
		}
	}
	if prod.Right[0] == glr.UnaryOperator {
		// unary_expression := unary_operator cast_expression
		return &entity.SingleExpression{
			ExpressionType: entity.ExpressionTypeUnary,
			UnaryExpressionInfo: &entity.UnaryExpressionInfo{
				Operator: node.Children[0].Children[0].Typ,
				Target:   parseExpressionNodeInner(node.Children[1]),
			},
		}
	}
	return &entity.SingleExpression{
		ExpressionType: entity.ExpressionTypeUnary,
		UnaryExpressionInfo: &entity.UnaryExpressionInfo{
			Operator: node.Children[0].Typ,
			Target:   parseExpressionNodeInner(node.Children[1]),
		},
	}
}

func flattenExpression(node *entity.AstNode) []*entity.AstNode {
	// flatten expression := expression COMMA assignment_expression
	if len(glr.Productions[node.ProdIndex].Right) == 1 {
		return []*entity.AstNode{node.Children[0]}
	}

	return append(flattenExpression(node.Children[0]), node.Children[2])
}
