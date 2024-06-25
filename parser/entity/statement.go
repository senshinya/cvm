package entity

import (
	"shinya.click/cvm/common"
	"shinya.click/cvm/parser/glr"
)

type StmtType int8

const (
	StmtTypeLabeled StmtType = iota + 1
	StmtTypeCase
	StmtTypeDefault
	StmtTypeCompound
	StmtTypeExpression
	StmtTypeIf
	StmtTypeIfElse
	StmtTypeSwitch
	StmtTypeWhile
	StmtTypeDoWhile
	StmtTypeFor
	StmtTypeGoto
	StmtTypeContinue
	StmtTypeBreak
	StmtTypeReturn
)

type Statement struct {
	StmtType StmtType

	LabeledStmtMetaInfo    *LabeledStmtMetaInfo
	CaseStmtMetaInfo       *CaseStmtMetaInfo
	DefaultStmtMetaInfo    *DefaultStmtMetaInfo
	CompoundStmtMetaInfo   *CompoundStmtMetaInfo
	ExpressionStmtMetaInfo *ExpressionStmtMetaInfo
	IfStmtMetaInfo         *IfStmtMetaInfo
	IfElseStmtMetaInfo     *IfElseStmtMetaInfo
	SwitchStmtMetaInfo     *SwitchStmtMetaInfo
	WhileStmtMetaInfo      *WhileStmtMetaInfo
	DoWhileStmtMetaInfo    *DoWhileStmtMetaInfo
	ForStmtMetaInfo        *ForStmtMetaInfo
	GotoStmtMetaInfo       *GotoStmtMetaInfo
	ReturnStmtMetaInfo     *ReturnStmtMetaInfo
}

type LabeledStmtMetaInfo struct {
	Identifier string
	Body       *Statement
}

type CaseStmtMetaInfo struct {
	Condition *SingleExpression
	Body      *Statement
}

type DefaultStmtMetaInfo struct {
	Body *Statement
}

type CompoundStmtMetaInfo struct {
	BlockItems []BlockItem
}

type ExpressionStmtMetaInfo struct {
	Expression *SingleExpression
}

type IfStmtMetaInfo struct {
	Condition *SingleExpression
	Body      *Statement
}

type IfElseStmtMetaInfo struct {
	Condition *SingleExpression
	Body      *Statement
	ElseBody  *Statement
}

type SwitchStmtMetaInfo struct {
	Expression *SingleExpression
	Body       *Statement
}

type WhileStmtMetaInfo struct {
	Condition *SingleExpression
	Body      *Statement
}

type DoWhileStmtMetaInfo struct {
	Body      *Statement
	Condition *SingleExpression
}

type ForStmtMetaInfo struct {
	Initialization *SingleExpression
	Condition      *SingleExpression
	AfterThought   *SingleExpression
	Body           *Statement
}

type GotoStmtMetaInfo struct {
	Identifier string
}

type ReturnStmtMetaInfo struct {
	Value *SingleExpression
}

func (s *Statement) GetBlockItemType() common.TokenType {
	return glr.Statement
}
