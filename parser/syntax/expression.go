package syntax

import "shinya.click/cvm/common"

type ExpressionType uint

const (
	ExpressionTypeConst ExpressionType = iota
	ExpressionTypeIdentifier
	ExpressionTypeAssignment
	ExpressionTypeCondition
	ExpressionTypeLogic
	ExpressionTypeBit
	ExpressionTypeNumber
	ExpressionTypeCast
	ExpressionTypeUnary
)

type SingleExpression struct {
	ExpressionType ExpressionType

	Terminal *common.Token

	AssignmentExpressionInfo *AssignmentExpressionInfo
	ConditionExpressionInfo  *ConditionExpressionInfo
	LogicExpressionInfo      *LogicExpressionInfo
	BitExpressionInfo        *BitExpressionInfo
	NumberExpressionInfo     *NumberExpressionInfo
	CastExpressionInfo       *CastExpressionInfo
}

type AssignmentExpressionInfo struct {
	LValue   []*SingleExpression
	Operator common.TokenType
	RValue   []*SingleExpression
}

type ConditionExpressionInfo struct {
	Condition   []*SingleExpression
	TrueBranch  []*SingleExpression
	FalseBranch []*SingleExpression
}

type LogicExpressionInfo struct {
	Operator common.TokenType
	One      []*SingleExpression
	Two      []*SingleExpression
}

type BitExpressionInfo struct {
	Operator common.TokenType
	One      []*SingleExpression
	Two      []*SingleExpression
}

type NumberExpressionInfo struct {
	Operator common.TokenType
	One      []*SingleExpression
	Two      []*SingleExpression
}

type CastExpressionInfo struct {
	Type   Type
	Target []*SingleExpression
}
