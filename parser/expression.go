package parser

import (
	"shinya.click/cvm/common"
	"shinya.click/cvm/parser/syntax"
)

func ParseExpressionNodes(nodes []*AstNode) []*syntax.SingleExpression {
	var res []*syntax.SingleExpression
	for _, node := range nodes {
		res = append(res, ParseExpressionNode(node)...)
	}
	return res
}

func ParseExpressionNode(node *AstNode) []*syntax.SingleExpression {
	if node.Typ == common.IDENTIFIER {
		return []*syntax.SingleExpression{{
			ExpressionType: syntax.ExpressionTypeIdentifier,
			Terminal:       node.Terminal,
		}}
	}
	if node.Typ == common.STRING || node.Typ == common.CHARACTER ||
		node.Typ == common.INTEGER_CONSTANT || node.Typ == common.FLOATING_CONSTANT {
		return []*syntax.SingleExpression{{
			ExpressionType: syntax.ExpressionTypeConst,
			Terminal:       node.Terminal,
		}}
	}

	prod := productions[node.ProdIndex]
	if len(prod.Right) == 1 {
		return ParseExpressionNode(node.Children[0])
	}

	switch node.Typ {
	case assignment_expression:
		// assignment_expression := unary_expression assignment_operator assignment_expression
		return []*syntax.SingleExpression{{
			ExpressionType: syntax.ExpressionTypeAssignment,
			AssignmentExpressionInfo: &syntax.AssignmentExpressionInfo{
				LValue:   ParseExpressionNode(node.Children[0]),
				Operator: node.Children[1].Children[0].Typ,
				RValue:   ParseExpressionNode(node.Children[2]),
			}},
		}
	case conditional_expression:
		// conditional_expression := logical_or_expression QUESTION expression COLON conditional_expression
		return []*syntax.SingleExpression{{
			ExpressionType: syntax.ExpressionTypeCondition,
			ConditionExpressionInfo: &syntax.ConditionExpressionInfo{
				Condition:   ParseExpressionNode(node.Children[0]),
				TrueBranch:  ParseExpressionNodes(flattenExpression(node.Children[2])),
				FalseBranch: ParseExpressionNode(node.Children[4]),
			}},
		}
	case logical_or_expression, logical_and_expression, exclusive_or_expression,
		equality_expression, relational_expression:
		// logical_or_expression := logical_or_expression OR_OR logical_and_expression
		return []*syntax.SingleExpression{{
			ExpressionType: syntax.ExpressionTypeLogic,
			LogicExpressionInfo: &syntax.LogicExpressionInfo{
				Operator: node.Children[1].Typ,
				One:      ParseExpressionNode(node.Children[0]),
				Two:      ParseExpressionNode(node.Children[2]),
			}},
		}
	case inclusive_or_expression, and_expression, shift_expression:
		// inclusive_or_expression := inclusive_or_expression OR exclusive_or_expression
		return []*syntax.SingleExpression{{
			ExpressionType: syntax.ExpressionTypeBit,
			BitExpressionInfo: &syntax.BitExpressionInfo{
				Operator: node.Children[1].Typ,
				One:      ParseExpressionNode(node.Children[0]),
				Two:      ParseExpressionNode(node.Children[2]),
			}},
		}
	case additive_expression, multiplicative_expression:
		// additive_expression := additive_expression PLUS multiplicative_expression
		return []*syntax.SingleExpression{{
			ExpressionType: syntax.ExpressionTypeNumber,
			NumberExpressionInfo: &syntax.NumberExpressionInfo{
				Operator: node.Children[1].Typ,
				One:      ParseExpressionNode(node.Children[0]),
				Two:      ParseExpressionNode(node.Children[2]),
			}},
		}
	case cast_expression:
		// cast_expression := LEFT_PARENTHESES type_name RIGHT_PARENTHESES cast_expression
		return []*syntax.SingleExpression{{
			ExpressionType: syntax.ExpressionTypeCast,
			CastExpressionInfo: &syntax.CastExpressionInfo{
				Type:   ParseTypeName(node.Children[1]),
				Target: ParseExpressionNode(node.Children[3]),
			}},
		}
	case unary_expression:
		// TODO
	case postfix_expression:
		// TODO
	case primary_expression:
		// primary_expression := LEFT_PARENTHESES expression RIGHT_PARENTHESES
		return ParseExpressionNode(node.Children[1])
	default:

	}
	panic("should not happen")
}

func flattenExpression(node *AstNode) []*AstNode {
	// flatten expression := expression COMMA assignment_expression
	if len(productions[node.ProdIndex].Right) == 1 {
		return []*AstNode{node.Children[0]}
	}

	return append(flattenExpression(node.Children[0]), node.Children[2])
}
