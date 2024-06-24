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

type Stmt struct {
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
}

type CaseStmtMetaInfo struct {
	Condition *SingleExpression
	Body      *Stmt
}

type DefaultStmtMetaInfo struct {
	Body *Stmt
}

type CompoundStmtMetaInfo struct {
	Stmts []*Stmt
}

type ExpressionStmtMetaInfo struct {
	Expression *SingleExpression
}

type IfStmtMetaInfo struct {
	Condition *SingleExpression
	Body      *Stmt
}

type IfElseStmtMetaInfo struct {
	Condition *SingleExpression
	Body      *Stmt
	ElseBody  *Stmt
}

type SwitchStmtMetaInfo struct {
	Expression *SingleExpression
	Body       *Stmt
}

type WhileStmtMetaInfo struct {
	Condition *SingleExpression
	Body      *Stmt
}

type DoWhileStmtMetaInfo struct {
	Body      *Stmt
	Condition *SingleExpression
}

type ForStmtMetaInfo struct {
	Initialization *SingleExpression
	Condition      *SingleExpression
	AfterThought   *SingleExpression
	Body           *Stmt
}

type GotoStmtMetaInfo struct {
	Identifier string
}

type ReturnStmtMetaInfo struct {
	Value *SingleExpression
}

func (s *Stmt) GetBlockItemType() common.TokenType {
	return glr.Statement
}
