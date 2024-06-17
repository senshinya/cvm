package parser

import (
	"golang.org/x/exp/constraints"
	"shinya.click/cvm/common"
	"shinya.click/cvm/parser/entity"
)

func SimplifyExpression(exp *entity.SingleExpression) *entity.SingleExpression {
	switch exp.ExpressionType {
	case entity.ExpressionTypeConst:
		return exp
	case entity.ExpressionTypeIdentifier:
		return exp
	case entity.ExpressionTypeAssignment:
		return simplifyAssignmentExpression(exp)
	case entity.ExpressionTypeCondition:
		return simplifyConditionExpression(exp)
	case entity.ExpressionTypeLogic:
		return simplifyLogicExpression(exp)
	case entity.ExpressionTypeBit:
		return simplifyBitExpression(exp)
	case entity.ExpressionTypeNumber:
		return simplifyNumberCalExpression(exp)
	case entity.ExpressionTypeCast:
		return simplifyCastExpression(exp)
	case entity.ExpressionTypeUnary:
		return simplifyUnaryExpression(exp)
	case entity.ExpressionTypeSizeOf:
		return simplifySizeOfExpression(exp)
	case entity.ExpressionTypeArrayAccess:
		return simplifyArrayAccessExpression(exp)
	case entity.ExpressionTypeFunctionCall:
		return simplifyFunctionCallExpression(exp)
	case entity.ExpressionTypeStructAccess:
		return simplifyStructAccessExpression(exp)
	case entity.ExpressionTypePostfix:
		return simplifyPostfixExpression(exp)
	case entity.ExpressionTypeConstruct:
		return simplifyConstructExpression(exp)
	case entity.ExpressionTypeExpressions:
		return simplifyExpressionsExpression(exp)
	}
	return exp
}

func simplifyAssignmentExpression(exp *entity.SingleExpression) *entity.SingleExpression {
	exp.AssignmentExpressionInfo.RValue = SimplifyExpression(exp.AssignmentExpressionInfo.RValue)
	exp.AssignmentExpressionInfo.LValue = SimplifyExpression(exp.AssignmentExpressionInfo.LValue)
	return exp
}

func simplifyConditionExpression(exp *entity.SingleExpression) *entity.SingleExpression {
	exp.ConditionExpressionInfo.Condition = SimplifyExpression(exp.ConditionExpressionInfo.Condition)
	if exp.ConditionExpressionInfo.Condition.ExpressionType != entity.ExpressionTypeConst {
		exp.ConditionExpressionInfo.TrueBranch = SimplifyExpression(exp.ConditionExpressionInfo.TrueBranch)
		exp.ConditionExpressionInfo.FalseBranch = SimplifyExpression(exp.ConditionExpressionInfo.FalseBranch)
		return exp
	}
	cond := exp.ConditionExpressionInfo.Condition.Terminal
	if cConstToBool(cond.Literal) {
		return SimplifyExpression(exp.ConditionExpressionInfo.TrueBranch)
	}
	return SimplifyExpression(exp.ConditionExpressionInfo.FalseBranch)
}

func simplifyLogicExpression(exp *entity.SingleExpression) *entity.SingleExpression {
	// short circuit
	exp.LogicExpressionInfo.One = SimplifyExpression(exp.LogicExpressionInfo.One)
	oneConst := exp.LogicExpressionInfo.One.ExpressionType == entity.ExpressionTypeConst
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

	exp.LogicExpressionInfo.Two = SimplifyExpression(exp.LogicExpressionInfo.Two)
	twoConst := exp.LogicExpressionInfo.Two.ExpressionType == entity.ExpressionTypeConst
	if twoConst {
		if exp.LogicExpressionInfo.Operator == common.OR_OR &&
			cConstToBool(exp.LogicExpressionInfo.Two.Terminal.Literal) {
			return constructTrueExpression()
		}
		if exp.LogicExpressionInfo.Operator == common.AND_AND &&
			!cConstToBool(exp.LogicExpressionInfo.Two.Terminal.Literal) {
			return constructFalseExpression()
		}
	}

	if !oneConst || !twoConst {
		return exp
	}

	oneLiteral, twoLiteral := exp.LogicExpressionInfo.One.Terminal.Literal, exp.LogicExpressionInfo.Two.Terminal.Literal

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

func simplifyBitExpression(exp *entity.SingleExpression) *entity.SingleExpression {
	exp.BitExpressionInfo.One = SimplifyExpression(exp.BitExpressionInfo.One)
	oneConst := exp.BitExpressionInfo.One.ExpressionType == entity.ExpressionTypeConst

	exp.BitExpressionInfo.Two = SimplifyExpression(exp.BitExpressionInfo.Two)
	twoConst := exp.BitExpressionInfo.Two.ExpressionType == entity.ExpressionTypeConst

	if !oneConst || !twoConst {
		return exp
	}

	oneLiteral, twoLiteral := exp.BitExpressionInfo.One.Terminal.Literal, exp.BitExpressionInfo.Two.Terminal.Literal
	return calTwoConstOperate(oneLiteral, twoLiteral, exp.BitExpressionInfo.Operator, exp)
}

func simplifyNumberCalExpression(exp *entity.SingleExpression) *entity.SingleExpression {
	exp.NumberExpressionInfo.One = SimplifyExpression(exp.NumberExpressionInfo.One)
	oneConst := exp.NumberExpressionInfo.One.ExpressionType == entity.ExpressionTypeConst

	exp.NumberExpressionInfo.Two = SimplifyExpression(exp.NumberExpressionInfo.Two)
	twoConst := exp.NumberExpressionInfo.Two.ExpressionType == entity.ExpressionTypeConst

	if !oneConst || !twoConst {
		return exp
	}

	oneLiteral, twoLiteral := exp.NumberExpressionInfo.One.Terminal.Literal, exp.NumberExpressionInfo.Two.Terminal.Literal
	return calTwoConstOperate(oneLiteral, twoLiteral, exp.NumberExpressionInfo.Operator, exp)
}

func simplifyCastExpression(exp *entity.SingleExpression) *entity.SingleExpression {
	exp.CastExpressionInfo.Source = SimplifyExpression(exp.CastExpressionInfo.Source)
	if exp.CastExpressionInfo.Source.ExpressionType != entity.ExpressionTypeConst {
		return exp
	}

	info := exp.CastExpressionInfo
	if info.Type.MetaType != entity.MetaTypeNumber {
		return exp
	}

	literal := exp.CastExpressionInfo.Source.Terminal.Literal
	if _, ok := literal.(string); ok {
		return exp
	}

	return castNumberConst(exp.CastExpressionInfo.Source.Terminal.Literal, info.Type.NumberMetaInfo)
}

func simplifyUnaryExpression(exp *entity.SingleExpression) *entity.SingleExpression {
	exp.UnaryExpressionInfo.Target = SimplifyExpression(exp.UnaryExpressionInfo.Target)
	return exp
}

func simplifySizeOfExpression(exp *entity.SingleExpression) *entity.SingleExpression {
	return exp
}

func simplifyArrayAccessExpression(exp *entity.SingleExpression) *entity.SingleExpression {
	exp.ArrayAccessExpressionInfo.Array = SimplifyExpression(exp.ArrayAccessExpressionInfo.Array)
	exp.ArrayAccessExpressionInfo.Target = SimplifyExpression(exp.ArrayAccessExpressionInfo.Target)
	return exp
}

func simplifyFunctionCallExpression(exp *entity.SingleExpression) *entity.SingleExpression {
	exp.FunctionCallExpressionInfo.Function = SimplifyExpression(exp.FunctionCallExpressionInfo.Function)
	return exp
}

func simplifyStructAccessExpression(exp *entity.SingleExpression) *entity.SingleExpression {
	exp.StructAccessExpressionInfo.Struct = SimplifyExpression(exp.StructAccessExpressionInfo.Struct)
	return exp
}

func simplifyPostfixExpression(exp *entity.SingleExpression) *entity.SingleExpression {
	exp.PostfixExpressionInfo.Target = SimplifyExpression(exp.PostfixExpressionInfo.Target)
	return exp
}

func simplifyConstructExpression(exp *entity.SingleExpression) *entity.SingleExpression {
	return exp
}

func simplifyExpressionsExpression(exp *entity.SingleExpression) *entity.SingleExpression {
	var res []*entity.SingleExpression
	for _, singleExpression := range exp.Expressions {
		res = append(res, SimplifyExpression(singleExpression))
	}
	exp.Expressions = res
	return exp
}

func castNumberConst(origin any, numberType *entity.NumberMetaInfo) *entity.SingleExpression {
	switch numberType.BaseNumType {
	case entity.BaseNumTypeChar:
		if numberType.Signed {
			return &entity.SingleExpression{
				ExpressionType: entity.ExpressionTypeConst,
				Terminal: &common.Token{
					Typ:     common.INTEGER_CONSTANT,
					Literal: convNumConstToInt8(origin),
				},
			}
		}
		return &entity.SingleExpression{
			ExpressionType: entity.ExpressionTypeConst,
			Terminal: &common.Token{
				Typ:     common.INTEGER_CONSTANT,
				Literal: convNumConstToUint8(origin),
			},
		}
	case entity.BaseNumTypeShort:
		if numberType.Signed {
			return &entity.SingleExpression{
				ExpressionType: entity.ExpressionTypeConst,
				Terminal: &common.Token{
					Typ:     common.INTEGER_CONSTANT,
					Literal: convNumConstToInt16(origin),
				},
			}
		}
		return &entity.SingleExpression{
			ExpressionType: entity.ExpressionTypeConst,
			Terminal: &common.Token{
				Typ:     common.INTEGER_CONSTANT,
				Literal: convNumConstToUint16(origin),
			},
		}
	case entity.BaseNumTypeInt:
		if numberType.Signed {
			return &entity.SingleExpression{
				ExpressionType: entity.ExpressionTypeConst,
				Terminal: &common.Token{
					Typ:     common.INTEGER_CONSTANT,
					Literal: convNumConstToInt32(origin),
				},
			}
		}
		return &entity.SingleExpression{
			ExpressionType: entity.ExpressionTypeConst,
			Terminal: &common.Token{
				Typ:     common.INTEGER_CONSTANT,
				Literal: convNumConstToUint32(origin),
			},
		}
	case entity.BaseNumTypeLong:
		if numberType.Signed {
			return &entity.SingleExpression{
				ExpressionType: entity.ExpressionTypeConst,
				Terminal: &common.Token{
					Typ:     common.INTEGER_CONSTANT,
					Literal: convNumConstToInt64(origin),
				},
			}
		}
		return &entity.SingleExpression{
			ExpressionType: entity.ExpressionTypeConst,
			Terminal: &common.Token{
				Typ:     common.INTEGER_CONSTANT,
				Literal: convNumConstToUint64(origin),
			},
		}
	case entity.BaseNumTypeFloat:
		return &entity.SingleExpression{
			ExpressionType: entity.ExpressionTypeConst,
			Terminal: &common.Token{
				Typ:     common.FLOATING_CONSTANT,
				Literal: convNumConstToFloat32(origin),
			},
		}
	case entity.BaseNumTypeDouble:
		return &entity.SingleExpression{
			ExpressionType: entity.ExpressionTypeConst,
			Terminal: &common.Token{
				Typ:     common.FLOATING_CONSTANT,
				Literal: convNumConstToFloat64(origin),
			},
		}
	case entity.BaseNumTypeBool:
		if cConstToBool(origin) {
			return constructTrueExpression()
		}
		return constructFalseExpression()
	case entity.BaseNumTypeLongLong:
		if numberType.Signed {
			return &entity.SingleExpression{
				ExpressionType: entity.ExpressionTypeConst,
				Terminal: &common.Token{
					Typ:     common.INTEGER_CONSTANT,
					Literal: convNumConstToInt64(origin),
				},
			}
		}
		return &entity.SingleExpression{
			ExpressionType: entity.ExpressionTypeConst,
			Terminal: &common.Token{
				Typ:     common.INTEGER_CONSTANT,
				Literal: convNumConstToUint64(origin),
			},
		}
	case entity.BaseNumTypeLongDouble:
		return &entity.SingleExpression{
			ExpressionType: entity.ExpressionTypeConst,
			Terminal: &common.Token{
				Typ:     common.FLOATING_CONSTANT,
				Literal: convNumConstToFloat64(origin),
			},
		}
	default:
		panic("Impossible")
	}
}

func cConstToBool(num any) bool {
	switch num.(type) {
	case string:
		return true
	case int8:
		return num.(int8) != 0
	case int16:
		return num.(int16) != 0
	case int32:
		return num.(int32) != 0
	case int64:
		return num.(int64) != 0
	case uint8:
		return num.(uint8) != 0
	case uint16:
		return num.(uint16) != 0
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

func goBoolToExpression(b bool) *entity.SingleExpression {
	if b {
		return constructTrueExpression()
	}
	return constructFalseExpression()
}

func constructTrueExpression() *entity.SingleExpression {
	return &entity.SingleExpression{
		ExpressionType: entity.ExpressionTypeConst,
		Terminal: &common.Token{
			Typ:     common.INTEGER_CONSTANT,
			Literal: 1,
		},
	}
}

func constructFalseExpression() *entity.SingleExpression {
	return &entity.SingleExpression{
		ExpressionType: entity.ExpressionTypeConst,
		Terminal: &common.Token{
			Typ:     common.INTEGER_CONSTANT,
			Literal: 0,
		},
	}
}

func constructConstExpression(literal any) *entity.SingleExpression {
	return &entity.SingleExpression{
		ExpressionType: entity.ExpressionTypeConst,
		Terminal: &common.Token{
			Typ:     common.INTEGER_CONSTANT,
			Literal: literal,
		},
	}
}

func calTwoConstOperate(one, two any, op common.TokenType, origin *entity.SingleExpression) *entity.SingleExpression {
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

func calStringRelatedConstOperate(str string, two any, op common.TokenType, origin *entity.SingleExpression) *entity.SingleExpression {
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

func calTwoNumberConstOperate(one, two any, op common.TokenType, origin *entity.SingleExpression) *entity.SingleExpression {
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

func calIntegerConstOperate[T constraints.Integer](one, two T, op common.TokenType, origin *entity.SingleExpression) *entity.SingleExpression {
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

func calFloatConstOperate[T constraints.Float](one, two T, op common.TokenType, origin *entity.SingleExpression) *entity.SingleExpression {
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

func calNumConstBitOperate[T constraints.Integer](one T, two T, op common.TokenType) *entity.SingleExpression {
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

func calNumConstLogicOperate[T constraints.Integer | constraints.Float](one, two T, op common.TokenType) *entity.SingleExpression {
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
	case uint16:
		return float64(one.(uint16))
	case int16:
		return float64(one.(int16))
	case uint8:
		return float64(one.(uint8))
	default:
		return float64(one.(int8))
	}
}

func convNumConstToFloat32(one any) float32 {
	switch one.(type) {
	case float64:
		return float32(one.(float64))
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
	case uint16:
		return float32(one.(uint16))
	case int16:
		return float32(one.(int16))
	case uint8:
		return float32(one.(uint8))
	default:
		return float32(one.(int8))
	}
}

func convNumConstToUint64(one any) uint64 {
	switch one.(type) {
	case float64:
		return uint64(one.(float64))
	case float32:
		return uint64(one.(float32))
	case uint64:
		return one.(uint64)
	case int64:
		return uint64(one.(int64))
	case uint32:
		return uint64(one.(uint32))
	case int32:
		return uint64(one.(int32))
	case uint16:
		return uint64(one.(uint16))
	case int16:
		return uint64(one.(int16))
	case uint8:
		return uint64(one.(uint8))
	default:
		return uint64(one.(int8))
	}
}

func convNumConstToInt64(one any) int64 {
	switch one.(type) {
	case float64:
		return int64(one.(float64))
	case float32:
		return int64(one.(float32))
	case uint64:
		return int64(one.(uint64))
	case int64:
		return one.(int64)
	case uint32:
		return int64(one.(uint32))
	case int32:
		return int64(one.(int32))
	case uint16:
		return int64(one.(uint16))
	case int16:
		return int64(one.(int16))
	case uint8:
		return int64(one.(uint8))
	default:
		return int64(one.(int8))
	}
}

func convNumConstToUint32(one any) uint32 {
	switch one.(type) {
	case float64:
		return uint32(one.(float64))
	case float32:
		return uint32(one.(float32))
	case uint64:
		return uint32(one.(uint64))
	case int64:
		return uint32(one.(int64))
	case uint32:
		return one.(uint32)
	case int32:
		return uint32(one.(int32))
	case uint16:
		return uint32(one.(uint16))
	case int16:
		return uint32(one.(int16))
	case uint8:
		return uint32(one.(uint8))
	default:
		return uint32(one.(int8))
	}
}

func convNumConstToInt32(one any) int32 {
	switch one.(type) {
	case float64:
		return int32(one.(float64))
	case float32:
		return int32(one.(float32))
	case uint64:
		return int32(one.(uint64))
	case int64:
		return int32(one.(int64))
	case uint32:
		return int32(one.(uint32))
	case int32:
		return one.(int32)
	case uint16:
		return int32(one.(uint16))
	case int16:
		return int32(one.(int16))
	case uint8:
		return int32(one.(uint8))
	default:
		return int32(one.(int8))
	}
}

func convNumConstToUint16(one any) uint16 {
	switch one.(type) {
	case float64:
		return uint16(one.(float64))
	case float32:
		return uint16(one.(float32))
	case uint64:
		return uint16(one.(uint64))
	case int64:
		return uint16(one.(int64))
	case uint32:
		return uint16(one.(uint32))
	case int32:
		return uint16(one.(int32))
	case uint16:
		return one.(uint16)
	case int16:
		return uint16(one.(int16))
	case uint8:
		return uint16(one.(uint8))
	default:
		return uint16(one.(int8))
	}
}

func convNumConstToInt16(one any) int16 {
	switch one.(type) {
	case float64:
		return int16(one.(float64))
	case float32:
		return int16(one.(float32))
	case uint64:
		return int16(one.(uint64))
	case int64:
		return int16(one.(int64))
	case uint32:
		return int16(one.(uint32))
	case int32:
		return int16(one.(int32))
	case uint16:
		return int16(one.(uint16))
	case int16:
		return one.(int16)
	case uint8:
		return int16(one.(uint8))
	default:
		return int16(one.(int8))
	}
}

func convNumConstToUint8(one any) uint8 {
	switch one.(type) {
	case float64:
		return uint8(one.(float64))
	case float32:
		return uint8(one.(float32))
	case uint64:
		return uint8(one.(uint64))
	case int64:
		return uint8(one.(int64))
	case uint32:
		return uint8(one.(uint32))
	case int32:
		return uint8(one.(int32))
	case uint16:
		return uint8(one.(uint16))
	case int16:
		return uint8(one.(int16))
	case uint8:
		return one.(uint8)
	default:
		return uint8(one.(int8))
	}
}

func convNumConstToInt8(one any) int8 {
	switch one.(type) {
	case float64:
		return int8(one.(float64))
	case float32:
		return int8(one.(float32))
	case uint64:
		return int8(one.(uint64))
	case int64:
		return int8(one.(int64))
	case uint32:
		return int8(one.(uint32))
	case int32:
		return int8(one.(int32))
	case uint16:
		return int8(one.(uint16))
	case int16:
		return int8(one.(int16))
	case uint8:
		return int8(one.(uint8))
	default:
		return one.(int8)
	}
}
