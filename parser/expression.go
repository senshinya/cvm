package parser

import (
	"shinya.click/cvm/common"
	"shinya.click/cvm/parser/syntax"
)

func ParseExpressionNode(node *AstNode) *syntax.SingleExpression {
	return simplifyConstExpression(parseExpressionNodeInner(node))
}

func parseExpressionNodeInner(node *AstNode) *syntax.SingleExpression {
	if node.Typ == common.IDENTIFIER {
		return &syntax.SingleExpression{
			ExpressionType: syntax.ExpressionTypeIdentifier,
			Terminal:       node.Terminal,
		}
	}
	if node.Typ == common.STRING || node.Typ == common.CHARACTER ||
		node.Typ == common.INTEGER_CONSTANT || node.Typ == common.FLOATING_CONSTANT {
		return &syntax.SingleExpression{
			ExpressionType: syntax.ExpressionTypeConst,
			Terminal:       node.Terminal,
		}
	}

	prod := productions[node.ProdIndex]
	if len(prod.Right) == 1 {
		return parseExpressionNodeInner(node.Children[0])
	}

	switch node.Typ {
	case assignment_expression:
		// assignment_expression := unary_expression assignment_operator assignment_expression
		return &syntax.SingleExpression{
			ExpressionType: syntax.ExpressionTypeAssignment,
			AssignmentExpressionInfo: &syntax.AssignmentExpressionInfo{
				LValue:   parseExpressionNodeInner(node.Children[0]),
				Operator: node.Children[1].Children[0].Typ,
				RValue:   parseExpressionNodeInner(node.Children[2]),
			},
		}
	case conditional_expression:
		// conditional_expression := logical_or_expression QUESTION expression COLON conditional_expression
		return &syntax.SingleExpression{
			ExpressionType: syntax.ExpressionTypeCondition,
			ConditionExpressionInfo: &syntax.ConditionExpressionInfo{
				Condition:   parseExpressionNodeInner(node.Children[0]),
				TrueBranch:  parseExpressionNodeInner(node.Children[2]),
				FalseBranch: parseExpressionNodeInner(node.Children[4]),
			},
		}
	case logical_or_expression, logical_and_expression, exclusive_or_expression,
		equality_expression, relational_expression:
		// logical_or_expression := logical_or_expression OR_OR logical_and_expression
		return &syntax.SingleExpression{
			ExpressionType: syntax.ExpressionTypeLogic,
			LogicExpressionInfo: &syntax.LogicExpressionInfo{
				Operator: node.Children[1].Typ,
				One:      parseExpressionNodeInner(node.Children[0]),
				Two:      parseExpressionNodeInner(node.Children[2]),
			},
		}
	case inclusive_or_expression, and_expression, shift_expression:
		// inclusive_or_expression := inclusive_or_expression OR exclusive_or_expression
		return &syntax.SingleExpression{
			ExpressionType: syntax.ExpressionTypeBit,
			BitExpressionInfo: &syntax.BitExpressionInfo{
				Operator: node.Children[1].Typ,
				One:      parseExpressionNodeInner(node.Children[0]),
				Two:      parseExpressionNodeInner(node.Children[2]),
			},
		}
	case additive_expression, multiplicative_expression:
		// additive_expression := additive_expression PLUS multiplicative_expression
		return &syntax.SingleExpression{
			ExpressionType: syntax.ExpressionTypeNumber,
			NumberExpressionInfo: &syntax.NumberExpressionInfo{
				Operator: node.Children[1].Typ,
				One:      parseExpressionNodeInner(node.Children[0]),
				Two:      parseExpressionNodeInner(node.Children[2]),
			},
		}
	case cast_expression:
		// cast_expression := LEFT_PARENTHESES type_name RIGHT_PARENTHESES cast_expression
		return &syntax.SingleExpression{
			ExpressionType: syntax.ExpressionTypeCast,
			CastExpressionInfo: &syntax.CastExpressionInfo{
				Type:   ParseTypeName(node.Children[1]),
				Target: parseExpressionNodeInner(node.Children[3]),
			},
		}
	case unary_expression:
		return parseUnary(node)
	case postfix_expression:
		return parsePostfix(node)
	case primary_expression:
		// primary_expression := LEFT_PARENTHESES expression RIGHT_PARENTHESES
		return parseExpressionNodeInner(node.Children[1])
	case expression:
		// expression := expression COMMA assignment_expression
		var exps []*syntax.SingleExpression
		for _, n := range flattenExpression(node) {
			exps = append(exps, parseExpressionNodeInner(n))
		}
		if len(exps) == 1 {
			return exps[0]
		}
		return &syntax.SingleExpression{
			ExpressionType: syntax.ExpressionTypeExpressions,
			Expressions:    exps,
		}
	}
	panic("should not happen")
}

func parsePostfix(node *AstNode) *syntax.SingleExpression {
	prod := productions[node.ProdIndex]
	if len(prod.Right) == 2 {
		// postfix_expression := postfix_expression PLUS_PLUS
		return &syntax.SingleExpression{
			ExpressionType: syntax.ExpressionTypePostfix,
			PostfixExpressionInfo: &syntax.PostfixExpressionInfo{
				Operator: node.Children[1].Typ,
				Target:   parseExpressionNodeInner(node.Children[0]),
			},
		}
	}
	if prod.Right[1] == common.LEFT_BRACKETS {
		// postfix_expression := postfix_expression LEFT_BRACKETS expression RIGHT_BRACKETS
		return &syntax.SingleExpression{
			ExpressionType: syntax.ExpressionTypeArrayAccess,
			ArrayAccessExpressionInfo: &syntax.ArrayAccessExpressionInfo{
				Array:  parseExpressionNodeInner(node.Children[0]),
				Target: parseExpressionNodeInner(node.Children[2]),
			},
		}
	}
	if prod.Right[1] == common.LEFT_PARENTHESES {
		// TODO function call
		return &syntax.SingleExpression{
			ExpressionType: syntax.ExpressionTypeFunctionCall,
			FunctionCallExpressionInfo: &syntax.FunctionCallExpressionInfo{
				Function: parseExpressionNodeInner(node.Children[0]),
			},
		}
	}
	if prod.Right[1] == common.PERIOD ||
		prod.Right[1] == common.ARROW {
		return &syntax.SingleExpression{
			ExpressionType: syntax.ExpressionTypeStructAccess,
			StructAccessExpressionInfo: &syntax.StructAccessExpressionInfo{
				Struct: parseExpressionNodeInner(node.Children[0]),
				Field:  node.Children[1].Terminal.Lexeme,
			},
		}
	}
	// postfix_expression := LEFT_PARENTHESES type_name RIGHT_PARENTHESES LEFT_BRACES initializer_list RIGHT_BRACES
	return &syntax.SingleExpression{
		ExpressionType: syntax.ExpressionTypeConstruct,
		ConstructExpressionInfo: &syntax.ConstructExpressionInfo{
			Type: ParseTypeName(node.Children[1]),
		},
	}
}

func parseUnary(node *AstNode) *syntax.SingleExpression {
	prod := productions[node.ProdIndex]
	if len(prod.Right) == 4 {
		// unary_expression := SIZEOF LEFT_PARENTHESES type_name RIGHT_PARENTHESES
		return &syntax.SingleExpression{
			ExpressionType: syntax.ExpressionTypeUnary,
			UnaryExpressionInfo: &syntax.UnaryExpressionInfo{
				Operator: common.SIZEOF,
				Target:   parseExpressionNodeInner(node.Children[2]),
			},
		}
	}
	if node.Children[0].Typ == unary_expression {
		// unary_expression := unary_operator cast_expression
		return &syntax.SingleExpression{
			ExpressionType: syntax.ExpressionTypeUnary,
			UnaryExpressionInfo: &syntax.UnaryExpressionInfo{
				Operator: node.Children[0].Children[0].Typ,
				Target:   parseExpressionNodeInner(node.Children[1]),
			},
		}
	}
	return &syntax.SingleExpression{
		ExpressionType: syntax.ExpressionTypeUnary,
		UnaryExpressionInfo: &syntax.UnaryExpressionInfo{
			Operator: node.Children[0].Typ,
			Target:   parseExpressionNodeInner(node.Children[1]),
		},
	}
}

func flattenExpression(node *AstNode) []*AstNode {
	// flatten expression := expression COMMA assignment_expression
	if len(productions[node.ProdIndex].Right) == 1 {
		return []*AstNode{node.Children[0]}
	}

	return append(flattenExpression(node.Children[0]), node.Children[2])
}

func simplifyConstExpression(exp *syntax.SingleExpression) *syntax.SingleExpression {
	switch exp.ExpressionType {
	case syntax.ExpressionTypeConst, syntax.ExpressionTypeIdentifier:
		return exp
	case syntax.ExpressionTypeAssignment, syntax.ExpressionTypeExpressions:
		// may contains side effect
		return exp
	case syntax.ExpressionTypeCondition:
		return simplifyConditionExpression(exp)
	case syntax.ExpressionTypeLogic:
		return simplifyLogicExpression(exp)
	case syntax.ExpressionTypeBit:
		return simplifyBitExpression(exp)
	case syntax.ExpressionTypeNumber:
	case syntax.ExpressionTypeCast:
	case syntax.ExpressionTypeUnary:
	case syntax.ExpressionTypeArrayAccess:
	case syntax.ExpressionTypeFunctionCall:
	case syntax.ExpressionTypeStructAccess:
	case syntax.ExpressionTypePostfix:
	case syntax.ExpressionTypeConstruct:
	default:
		panic(nil)
	}
	panic(nil)
}

func simplifyConditionExpression(exp *syntax.SingleExpression) *syntax.SingleExpression {
	exp.ConditionExpressionInfo.Condition = simplifyConstExpression(exp.ConditionExpressionInfo.Condition)
	exp.ConditionExpressionInfo.TrueBranch = simplifyConstExpression(exp.ConditionExpressionInfo.TrueBranch)
	exp.ConditionExpressionInfo.FalseBranch = simplifyConstExpression(exp.ConditionExpressionInfo.FalseBranch)
	if exp.ConditionExpressionInfo.Condition.ExpressionType != syntax.ExpressionTypeConst {
		return exp
	}
	ter := exp.ConditionExpressionInfo.Condition.Terminal
	if cConstToBool(ter.Literal) {
		return exp.ConditionExpressionInfo.TrueBranch
	}
	return exp.ConditionExpressionInfo.FalseBranch
}

func cConstToBool(num any) bool {
	switch num.(type) {
	case string:
		return true
	case byte:
		return num.(byte) != 0
	case int32:
		return num.(int32) != 0
	case int64:
		return num.(int64) != 0
	case uint32:
		return num.(uint32) != 0
	case uint64:
		return num.(uint64) != 0
	case float32:
		return num.(float32) != 0
	case float64:
		return num.(float64) != 0
	}
	panic(nil)
}

func simplifyLogicExpression(exp *syntax.SingleExpression) *syntax.SingleExpression {
	exp.LogicExpressionInfo.One = simplifyConditionExpression(exp.LogicExpressionInfo.One)
	if exp.LogicExpressionInfo.Operator == common.OR_OR &&
		cConstToBool(exp.LogicExpressionInfo.One.Terminal.Lexeme) {
		return exp.LogicExpressionInfo.One
	}
	if exp.LogicExpressionInfo.Operator == common.AND_AND &&
		!cConstToBool(exp.LogicExpressionInfo.One.Terminal.Lexeme) {
		return exp.LogicExpressionInfo.One
	}

	exp.LogicExpressionInfo.Two = simplifyLogicExpression(exp.LogicExpressionInfo.Two)
	if exp.LogicExpressionInfo.Operator == common.OR_OR &&
		cConstToBool(exp.LogicExpressionInfo.Two.Terminal.Lexeme) {
		return exp.LogicExpressionInfo.Two
	}
	if exp.LogicExpressionInfo.Operator == common.AND_AND &&
		!cConstToBool(exp.LogicExpressionInfo.Two.Terminal.Lexeme) {
		return exp.LogicExpressionInfo.Two
	}
	switch exp.LogicExpressionInfo.Operator {
	case common.OR_OR:

	case common.AND_AND:

	case common.XOR:

	case common.EQUAL_EQUAL:
	case common.NOT_EQUAL:
	case common.LESS:
	case common.GREATER:
	case common.LESS_EQUAL:
	case common.GREATER_EQUAL:
	}
	panic(nil)
}

func simplifyBitExpression(exp *syntax.SingleExpression) *syntax.SingleExpression {
	return nil
}
