package parser

import (
	"shinya.click/cvm/common"
	"shinya.click/cvm/parser/syntax"
)

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
	// short circuit
	exp.LogicExpressionInfo.One = simplifyConstExpression(exp.LogicExpressionInfo.One)
	oneConst := exp.LogicExpressionInfo.One.ExpressionType == syntax.ExpressionTypeConst
	if oneConst {
		if exp.LogicExpressionInfo.Operator == common.OR_OR &&
			cConstToBool(exp.LogicExpressionInfo.One.Terminal.Lexeme) {
			return constructTrueExpression()
		}
		if exp.LogicExpressionInfo.Operator == common.AND_AND &&
			!cConstToBool(exp.LogicExpressionInfo.One.Terminal.Lexeme) {
			return constructFalseExpression()
		}
	}

	exp.LogicExpressionInfo.Two = simplifyConstExpression(exp.LogicExpressionInfo.Two)
	twoConst := exp.LogicExpressionInfo.Two.ExpressionType == syntax.ExpressionTypeConst

	if !oneConst || !twoConst {
		return exp
	}

	oneLexeme, twoLexeme := exp.LogicExpressionInfo.One.Terminal.Lexeme, exp.LogicExpressionInfo.One.Terminal.Lexeme

	switch exp.LogicExpressionInfo.Operator {
	case common.OR_OR:
		return goBoolToExpression(cConstToBool(oneLexeme) || cConstToBool(twoLexeme))
	case common.AND_AND:
		return goBoolToExpression(cConstToBool(oneLexeme) && cConstToBool(twoLexeme))
	case common.EQUAL_EQUAL, common.NOT_EQUAL, common.LESS,
		common.GREATER, common.LESS_EQUAL, common.GREATER_EQUAL:
		return CalTwoConstOperate(oneLexeme, twoLexeme, exp.LogicExpressionInfo.Operator)
	}
	panic(nil)
}

func goBoolToExpression(b bool) *syntax.SingleExpression {
	if b {
		return constructTrueExpression()
	}
	return constructFalseExpression()
}

func constructTrueExpression() *syntax.SingleExpression {
	return &syntax.SingleExpression{
		ExpressionType: syntax.ExpressionTypeConst,
		Terminal: &common.Token{
			Typ:     common.INTEGER_CONSTANT,
			Literal: 1,
		},
	}
}

func constructFalseExpression() *syntax.SingleExpression {
	return &syntax.SingleExpression{
		ExpressionType: syntax.ExpressionTypeConst,
		Terminal: &common.Token{
			Typ:     common.INTEGER_CONSTANT,
			Literal: 0,
		},
	}
}

func simplifyBitExpression(exp *syntax.SingleExpression) *syntax.SingleExpression {
	return nil
}

func CalTwoConstOperate(one, two any, op common.TokenType) *syntax.SingleExpression {
	_, oneString := one.(string)
	_, oneFloat64 := one.(float64)
	_, oneFloat32 := one.(float32)
	_, oneUint64 := one.(uint64)
	_, oneInt64 := one.(int64)
	_, oneUInt32 := one.(uint32)
	_, oneInt32 := one.(int32)
	_, oneByte := one.(byte)

	_, twoString := two.(string)
	_, twoFloat64 := two.(float64)
	_, twoFloat32 := two.(float32)
	_, twoUint64 := two.(uint64)
	_, twoInt64 := two.(int64)
	_, twoUInt32 := two.(uint32)
	_, twoInt32 := two.(int32)
	_, twoByte := two.(byte)

	// TODO

	return nil
}
