package entity

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
	ExpressionTypeSizeOf
	ExpressionTypeArrayAccess
	ExpressionTypeFunctionCall
	ExpressionTypeStructAccess
	ExpressionTypePostfix
	ExpressionTypeConstruct
	ExpressionTypeExpressions
)

type Expression struct {
	ExpressionType ExpressionType

	Terminal *common.Token

	AssignmentExpressionInfo   *AssignmentExpressionInfo
	ConditionExpressionInfo    *ConditionExpressionInfo
	LogicExpressionInfo        *LogicExpressionInfo
	BitExpressionInfo          *BitExpressionInfo
	NumberExpressionInfo       *NumberExpressionInfo
	CastExpressionInfo         *CastExpressionInfo
	UnaryExpressionInfo        *UnaryExpressionInfo
	SizeOfExpressionInfo       *SizeOfExpressionInfo
	ArrayAccessExpressionInfo  *ArrayAccessExpressionInfo
	PostfixExpressionInfo      *PostfixExpressionInfo
	FunctionCallExpressionInfo *FunctionCallExpressionInfo
	StructAccessExpressionInfo *StructAccessExpressionInfo
	ConstructExpressionInfo    *ConstructExpressionInfo
	Expressions                []*Expression

	common.SourceRange
}

type AssignmentExpressionInfo struct {
	LValue   *Expression
	Operator common.TokenType
	RValue   *Expression
}

type ConditionExpressionInfo struct {
	Condition   *Expression
	TrueBranch  *Expression
	FalseBranch *Expression
}

type LogicExpressionInfo struct {
	Operator common.TokenType
	One      *Expression
	Two      *Expression
}

type BitExpressionInfo struct {
	Operator common.TokenType
	One      *Expression
	Two      *Expression
}

type NumberExpressionInfo struct {
	Operator common.TokenType
	One      *Expression
	Two      *Expression
}

type CastExpressionInfo struct {
	Type   Type
	Source *Expression
}

type UnaryExpressionInfo struct {
	Operator common.TokenType
	Target   *Expression
}

type SizeOfExpressionInfo struct {
	Type   Type
	Target *Expression
}

type ArrayAccessExpressionInfo struct {
	Array  *Expression
	Target *Expression
}

type PostfixExpressionInfo struct {
	Operator common.TokenType
	Target   *Expression
}

type FunctionCallExpressionInfo struct {
	Function  *Expression
	Arguments []*Expression
}

type StructAccessExpressionInfo struct {
	Struct *Expression
	Field  string
	Access common.TokenType
}

type ConstructExpressionInfo struct {
	Type         Type
	Initializers []*InitializerItem
}
