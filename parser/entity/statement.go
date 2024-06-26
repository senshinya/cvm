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

	common.SourceRange
}

type LabeledStmtMetaInfo struct {
	Identifier *common.Token
	Body       *Statement
}

type CaseStmtMetaInfo struct {
	Condition *Expression
	Body      *Statement
}

type DefaultStmtMetaInfo struct {
	Body *Statement
}

type CompoundStmtMetaInfo struct {
	BlockItems []BlockItem
}

type ExpressionStmtMetaInfo struct {
	Expression *Expression
}

type IfStmtMetaInfo struct {
	Condition *Expression
	Body      *Statement
}

type IfElseStmtMetaInfo struct {
	Condition *Expression
	Body      *Statement
	ElseBody  *Statement
}

type SwitchStmtMetaInfo struct {
	Condition *Expression
	Body      *Statement
}

type WhileStmtMetaInfo struct {
	Condition *Expression
	Body      *Statement
}

type DoWhileStmtMetaInfo struct {
	Body      *Statement
	Condition *Expression
}

type ForStmtMetaInfo struct {
	Initialization BlockItem
	Condition      *Expression
	AfterThought   *Expression
	Body           *Statement
}

type GotoStmtMetaInfo struct {
	Identifier *common.Token
}

type ReturnStmtMetaInfo struct {
	Value *Expression
}

func (s *Statement) GetBlockItemType() common.TokenType {
	return glr.Statement
}
