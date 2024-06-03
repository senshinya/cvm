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
	ExpressionTypeTypeSize
	ExpressionTypeArrayAccess
	ExpressionTypeFunctionCall
	ExpressionTypeStructAccess
	ExpressionTypePostfix
	ExpressionTypeConstruct
	ExpressionTypeExpressions
)

type SingleExpression struct {
	ExpressionType ExpressionType

	Terminal *common.Token

	AssignmentExpressionInfo   *AssignmentExpressionInfo
	ConditionExpressionInfo    *ConditionExpressionInfo
	LogicExpressionInfo        *LogicExpressionInfo
	BitExpressionInfo          *BitExpressionInfo
	NumberExpressionInfo       *NumberExpressionInfo
	CastExpressionInfo         *CastExpressionInfo
	UnaryExpressionInfo        *UnaryExpressionInfo
	TypeSizeExpressionInfo     *TypeSizeExpressionInfo
	ArrayAccessExpressionInfo  *ArrayAccessExpressionInfo
	PostfixExpressionInfo      *PostfixExpressionInfo
	FunctionCallExpressionInfo *FunctionCallExpressionInfo
	StructAccessExpressionInfo *StructAccessExpressionInfo
	ConstructExpressionInfo    *ConstructExpressionInfo
	Expressions                []*SingleExpression
}

type AssignmentExpressionInfo struct {
	LValue   *SingleExpression
	Operator common.TokenType
	RValue   *SingleExpression
}

type ConditionExpressionInfo struct {
	Condition   *SingleExpression
	TrueBranch  *SingleExpression
	FalseBranch *SingleExpression
}

type LogicExpressionInfo struct {
	Operator common.TokenType
	One      *SingleExpression
	Two      *SingleExpression
}

type BitExpressionInfo struct {
	Operator common.TokenType
	One      *SingleExpression
	Two      *SingleExpression
}

type NumberExpressionInfo struct {
	Operator common.TokenType
	One      *SingleExpression
	Two      *SingleExpression
}

type CastExpressionInfo struct {
	Type   Type
	Target *SingleExpression
}

type UnaryExpressionInfo struct {
	Operator common.TokenType
	Target   *SingleExpression
}

type TypeSizeExpressionInfo struct {
	Type Type
}

type ArrayAccessExpressionInfo struct {
	Array  *SingleExpression
	Target *SingleExpression
}

type PostfixExpressionInfo struct {
	Operator common.TokenType
	Target   *SingleExpression
}

type FunctionCallExpressionInfo struct {
	Function *SingleExpression
}

type StructAccessExpressionInfo struct {
	Struct *SingleExpression
	Field  string
}

type ConstructExpressionInfo struct {
	Type Type
}
