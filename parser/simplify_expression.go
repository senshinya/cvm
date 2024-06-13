package parser

import (
	"golang.org/x/exp/constraints"
	"shinya.click/cvm/common"
	"shinya.click/cvm/parser/syntax"
)

func simplifyConstExpression(exp *syntax.SingleExpression) *syntax.SingleExpression {
	// TODO Simplify Inner Expression
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
		return simplifyNumberCalExpression(exp)
	case syntax.ExpressionTypeCast:
		return simplifyCastExpression(exp)
	case syntax.ExpressionTypeUnary:
		return exp
	case syntax.ExpressionTypeTypeSize:
		return exp
	case syntax.ExpressionTypeArrayAccess:
		return exp
	case syntax.ExpressionTypeFunctionCall:
		return exp
	case syntax.ExpressionTypeStructAccess:
		return exp
	case syntax.ExpressionTypePostfix:
		return exp
	case syntax.ExpressionTypeConstruct:
		return exp
	default:
		return exp
	}
}

func simplifyCastExpression(exp *syntax.SingleExpression) *syntax.SingleExpression {
	exp.CastExpressionInfo.Target = simplifyConstExpression(exp.CastExpressionInfo.Target)
	if exp.CastExpressionInfo.Target.ExpressionType != syntax.ExpressionTypeConst {
		return exp
	}

	info := exp.CastExpressionInfo
	if info.Type.MetaType != syntax.MetaTypeNumber {
		return exp
	}

	literal := exp.CastExpressionInfo.Target.Terminal.Literal
	if _, ok := literal.(string); ok {
		return exp
	}

	return castNumberConst(exp.CastExpressionInfo.Target.Terminal.Literal, info.Type.NumberMetaInfo)
}

func castNumberConst(origin any, numberType *syntax.NumberMetaInfo) *syntax.SingleExpression {
	// TODO
	return nil
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
			cConstToBool(exp.LogicExpressionInfo.One.Terminal.Literal) {
			return constructTrueExpression()
		}
		if exp.LogicExpressionInfo.Operator == common.AND_AND &&
			!cConstToBool(exp.LogicExpressionInfo.One.Terminal.Literal) {
			return constructFalseExpression()
		}
	}

	exp.LogicExpressionInfo.Two = simplifyConstExpression(exp.LogicExpressionInfo.Two)
	twoConst := exp.LogicExpressionInfo.Two.ExpressionType == syntax.ExpressionTypeConst

	if !oneConst || !twoConst {
		return exp
	}

	oneLiteral, twoLiteral := exp.LogicExpressionInfo.One.Terminal.Literal, exp.LogicExpressionInfo.One.Terminal.Literal

	switch exp.LogicExpressionInfo.Operator {
	case common.OR_OR:
		return goBoolToExpression(cConstToBool(oneLiteral) || cConstToBool(twoLiteral))
	case common.AND_AND:
		return goBoolToExpression(cConstToBool(oneLiteral) && cConstToBool(twoLiteral))
	case common.EQUAL_EQUAL, common.NOT_EQUAL, common.LESS,
		common.GREATER, common.LESS_EQUAL, common.GREATER_EQUAL:
		return calTwoConstOperate(oneLiteral, twoLiteral, exp.LogicExpressionInfo.Operator, exp)
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

func constructConstExpression(literal any) *syntax.SingleExpression {
	return &syntax.SingleExpression{
		ExpressionType: syntax.ExpressionTypeConst,
		Terminal: &common.Token{
			Typ:     common.INTEGER_CONSTANT,
			Literal: literal,
		},
	}
}

func simplifyBitExpression(exp *syntax.SingleExpression) *syntax.SingleExpression {
	exp.BitExpressionInfo.One = simplifyConstExpression(exp.BitExpressionInfo.One)
	oneConst := exp.BitExpressionInfo.One.ExpressionType == syntax.ExpressionTypeConst

	if !oneConst {
		return exp
	}

	exp.BitExpressionInfo.Two = simplifyConstExpression(exp.BitExpressionInfo.Two)
	twoConst := exp.BitExpressionInfo.Two.ExpressionType == syntax.ExpressionTypeConst

	if !twoConst {
		return exp
	}

	oneLiteral, twoLiteral := exp.BitExpressionInfo.One.Terminal.Literal, exp.BitExpressionInfo.One.Terminal.Literal
	return calTwoConstOperate(oneLiteral, twoLiteral, exp.BitExpressionInfo.Operator, exp)
}

func simplifyNumberCalExpression(exp *syntax.SingleExpression) *syntax.SingleExpression {
	exp.NumberExpressionInfo.One = simplifyConstExpression(exp.NumberExpressionInfo.One)
	oneConst := exp.NumberExpressionInfo.One.ExpressionType == syntax.ExpressionTypeConst

	exp.NumberExpressionInfo.Two = simplifyConstExpression(exp.NumberExpressionInfo.Two)
	twoConst := exp.NumberExpressionInfo.Two.ExpressionType == syntax.ExpressionTypeConst

	if !oneConst || !twoConst {
		return exp
	}

	oneLiteral, twoLiteral := exp.NumberExpressionInfo.One.Terminal.Literal, exp.NumberExpressionInfo.Two.Terminal.Literal
	return calTwoConstOperate(oneLiteral, twoLiteral, exp.NumberExpressionInfo.Operator, exp)
}

func calTwoConstOperate(one, two any, op common.TokenType, origin *syntax.SingleExpression) *syntax.SingleExpression {
	_, oneString := one.(string)
	_, twoString := two.(string)

	if oneString {
		return calStringRelatedConstOperate(one.(string), two, op, origin)
	}
	if twoString {
		return calStringRelatedConstOperate(two.(string), one, op, origin)
	}

	return calTwoNumberConstOperate(one, two, op, origin)
}

func calStringRelatedConstOperate(str string, two any, op common.TokenType, origin *syntax.SingleExpression) *syntax.SingleExpression {
	switch two.(type) {
	case string, float32, float64:
		panic("invalid operands to binary expression")
	}
	switch op {
	case common.OR, common.AND, common.XOR,
		common.LEFT_SHIFT, common.RIGHT_SHIFT:
		// bit-wise op to const str is not supported
		panic("invalid operands to binary expression")
	case common.PLUS, common.MINUS:
		// cannot handle right now
		return origin
	case common.ASTERISK, common.SLASH, common.PERCENT:
		panic("invalid operands to binary expression")
	}
	return nil
}

func calTwoNumberConstOperate(one, two any, op common.TokenType, origin *syntax.SingleExpression) *syntax.SingleExpression {
	_, oneFloat64 := one.(float64)
	_, oneFloat32 := one.(float32)
	_, oneUint64 := one.(uint64)
	_, oneInt64 := one.(int64)
	_, oneUInt32 := one.(uint32)
	//_, oneInt32 := one.(int32)
	//_, oneByte := one.(byte)

	_, twoFloat64 := two.(float64)
	_, twoFloat32 := two.(float32)
	_, twoUint64 := two.(uint64)
	_, twoInt64 := two.(int64)
	_, twoUInt32 := two.(uint32)
	//_, twoInt32 := two.(int32)
	//_, twoByte := two.(byte)

	switch {
	case oneFloat64 || twoFloat64:
		return calFloatConstOperate(convNumConstToFloat64(one), convNumConstToFloat64(two), op, origin)
	case oneFloat32 || twoFloat32:
		return calFloatConstOperate(convNumConstToFloat32(one), convNumConstToFloat32(two), op, origin)
	case oneUint64 || twoUint64:
		return calIntegerConstOperate(convNumConstToUint64(one), convNumConstToUint64(two), op, origin)
	case oneInt64 || twoInt64:
		return calIntegerConstOperate(convNumConstToInt64(one), convNumConstToInt64(two), op, origin)
	case oneUInt32 || twoUInt32:
		return calIntegerConstOperate(convNumConstToUint32(one), convNumConstToUint32(two), op, origin)
	default:
		return calIntegerConstOperate(convNumConstToInt32(one), convNumConstToInt32(two), op, origin)
	}
}

func calIntegerConstOperate[T constraints.Integer](one, two T, op common.TokenType, origin *syntax.SingleExpression) *syntax.SingleExpression {
	switch op {
	case common.EQUAL_EQUAL, common.NOT_EQUAL, common.LESS,
		common.GREATER, common.LESS_EQUAL, common.GREATER_EQUAL:
		return calNumConstLogicOperate(one, two, op)
	case common.OR, common.AND, common.XOR,
		common.LEFT_SHIFT, common.RIGHT_SHIFT:
		calNumConstBitOperate(one, two, op)
	case common.PLUS:
		return constructConstExpression(one + two)
	case common.MINUS:
		return constructConstExpression(one - two)
	case common.ASTERISK:
		return constructConstExpression(one * two)
	case common.SLASH:
		if two == 0 {
			// ub
			return origin
		}
		return constructConstExpression(one / two)
	case common.PERCENT:
		return constructConstExpression(one % two)
	}
	panic("invalid operands to binary expression")
}

func calFloatConstOperate[T constraints.Float](one, two T, op common.TokenType, origin *syntax.SingleExpression) *syntax.SingleExpression {
	switch op {
	case common.EQUAL_EQUAL, common.NOT_EQUAL, common.LESS,
		common.GREATER, common.LESS_EQUAL, common.GREATER_EQUAL:
		return calNumConstLogicOperate(one, two, op)
	case common.OR, common.AND, common.XOR,
		common.LEFT_SHIFT, common.RIGHT_SHIFT:
		panic("invalid operand to binary expression")
	case common.PLUS:
		return constructConstExpression(one + two)
	case common.MINUS:
		return constructConstExpression(one - two)
	case common.ASTERISK:
		return constructConstExpression(one * two)
	case common.SLASH:
		if two == 0.0 {
			// ub
			return origin
		}
		return constructConstExpression(one / two)
	case common.PERCENT:
		panic("invalid operands to binary expression")
	}
	panic("invalid operands to binary expression")
}

func calNumConstBitOperate[T constraints.Integer](one T, two T, op common.TokenType) *syntax.SingleExpression {
	switch op {
	case common.OR:
		return constructConstExpression(one | two)
	case common.AND:
		return constructConstExpression(one & two)
	case common.XOR:
		return constructConstExpression(one ^ two)
	case common.LEFT_SHIFT:
		return constructConstExpression(one << two)
	case common.RIGHT_SHIFT:
		return constructConstExpression(one >> two)
	}
	panic("invalid operands to binary expression")
}

func calNumConstLogicOperate[T constraints.Integer | constraints.Float](one, two T, op common.TokenType) *syntax.SingleExpression {
	switch op {
	case common.EQUAL_EQUAL:
		if one == two {
			return constructTrueExpression()
		}
		return constructFalseExpression()
	case common.NOT_EQUAL:
		if one != two {
			return constructTrueExpression()
		}
		return constructFalseExpression()
	case common.LESS:
		if one < two {
			return constructTrueExpression()
		}
		return constructFalseExpression()
	case common.GREATER:
		if one > two {
			return constructTrueExpression()
		}
		return constructFalseExpression()
	case common.LESS_EQUAL:
		if one <= two {
			return constructTrueExpression()
		}
		return constructFalseExpression()
	case common.GREATER_EQUAL:
		if one >= two {
			return constructTrueExpression()
		}
		return constructFalseExpression()
	}
	panic("invalid logic operate")
}

func convNumConstToFloat64(one any) float64 {
	switch one.(type) {
	case float64:
		return one.(float64)
	case float32:
		return float64(one.(float32))
	case uint64:
		return float64(one.(uint64))
	case int64:
		return float64(one.(int64))
	case uint32:
		return float64(one.(uint32))
	case int32:
		return float64(one.(int32))
	default:
		return float64(one.(byte)) // last is byte
	}
}

func convNumConstToFloat32(one any) float32 {
	switch one.(type) {
	case float32:
		return one.(float32)
	case uint64:
		return float32(one.(uint64))
	case int64:
		return float32(one.(int64))
	case uint32:
		return float32(one.(uint32))
	case int32:
		return float32(one.(int32))
	default:
		return float32(one.(byte)) // last is byte
	}
}

func convNumConstToUint64(one any) uint64 {
	switch one.(type) {
	case uint64:
		return one.(uint64)
	case int64:
		return uint64(one.(int64))
	case uint32:
		return uint64(one.(uint32))
	case int32:
		return uint64(one.(int32))
	default:
		return uint64(one.(byte)) // last is byte
	}
}

func convNumConstToInt64(one any) int64 {
	switch one.(type) {
	case int64:
		return one.(int64)
	case uint32:
		return int64(one.(uint32))
	case int32:
		return int64(one.(int32))
	default:
		return int64(one.(byte)) // last is byte
	}
}

func convNumConstToUint32(one any) uint32 {
	switch one.(type) {
	case uint32:
		return one.(uint32)
	case int32:
		return uint32(one.(int32))
	default:
		return uint32(one.(byte)) // last is byte
	}
}

func convNumConstToInt32(one any) int32 {
	switch one.(type) {
	case int32:
		return one.(int32)
	default:
		return int32(one.(byte)) // last is byte
	}
}
