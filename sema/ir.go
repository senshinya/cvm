package sema

import "shinya.click/cvm/entity"

type Node interface {
	Pos() entity.SourceRange
}

type Stmt interface {
	Node
	isStmt()
}

type Expr interface {
	Node
	isExpr()
	GetType() Type
	GetCategory() ValueCategory
}

type Decl interface {
	Node
	isDecl()
}

type ValueCategory int

const (
	RValue ValueCategory = iota
	LValue
)

type Program struct {
	Globals []Decl
	Funcs   []*FuncDef
	Types   *TypeTable
	SymTab  *SymbolTable
}

type Block struct {
	Items []Stmt
	Scope *Scope
	Range entity.SourceRange
}

func (b *Block) Pos() entity.SourceRange { return b.Range }
func (*Block) isStmt()                   {}

type IfStmt struct {
	Cond  Expr
	Then  Stmt
	Else  Stmt
	Range entity.SourceRange
}

func (s *IfStmt) Pos() entity.SourceRange { return s.Range }
func (*IfStmt) isStmt()                   {}

type WhileStmt struct {
	Cond    Expr
	Body    Stmt
	DoWhile bool
	Range   entity.SourceRange
}

func (s *WhileStmt) Pos() entity.SourceRange { return s.Range }
func (*WhileStmt) isStmt()                   {}

type ForStmt struct {
	Init  Stmt
	Cond  Expr
	Post  Expr
	Body  Stmt
	Scope *Scope
	Range entity.SourceRange
}

func (s *ForStmt) Pos() entity.SourceRange { return s.Range }
func (*ForStmt) isStmt()                   {}

type SwitchStmt struct {
	Cond    Expr
	Body    Stmt
	Cases   []*CaseStmt
	Default *DefaultStmt
	Range   entity.SourceRange
}

func (s *SwitchStmt) Pos() entity.SourceRange { return s.Range }
func (*SwitchStmt) isStmt()                   {}

type CaseStmt struct {
	Value int64
	Body  Stmt
	Range entity.SourceRange
}

func (s *CaseStmt) Pos() entity.SourceRange { return s.Range }
func (*CaseStmt) isStmt()                   {}

type DefaultStmt struct {
	Body  Stmt
	Range entity.SourceRange
}

func (s *DefaultStmt) Pos() entity.SourceRange { return s.Range }
func (*DefaultStmt) isStmt()                   {}

type ReturnStmt struct {
	Value Expr
	Range entity.SourceRange
}

func (s *ReturnStmt) Pos() entity.SourceRange { return s.Range }
func (*ReturnStmt) isStmt()                   {}

type BreakStmt struct{ Range entity.SourceRange }

func (s *BreakStmt) Pos() entity.SourceRange { return s.Range }
func (*BreakStmt) isStmt()                   {}

type ContinueStmt struct{ Range entity.SourceRange }

func (s *ContinueStmt) Pos() entity.SourceRange { return s.Range }
func (*ContinueStmt) isStmt()                   {}

type GotoStmt struct {
	Target *LabeledStmt
	Name   string
	Range  entity.SourceRange
}

func (s *GotoStmt) Pos() entity.SourceRange { return s.Range }
func (*GotoStmt) isStmt()                   {}

type LabeledStmt struct {
	Name  string
	Body  Stmt
	Range entity.SourceRange
}

func (s *LabeledStmt) Pos() entity.SourceRange { return s.Range }
func (*LabeledStmt) isStmt()                   {}

type ExprStmt struct {
	Expr  Expr
	Range entity.SourceRange
}

func (s *ExprStmt) Pos() entity.SourceRange { return s.Range }
func (*ExprStmt) isStmt()                   {}

type DeclStmt struct {
	Decls []Decl
	Range entity.SourceRange
}

func (s *DeclStmt) Pos() entity.SourceRange { return s.Range }
func (*DeclStmt) isStmt()                   {}

type EmptyStmt struct{ Range entity.SourceRange }

func (s *EmptyStmt) Pos() entity.SourceRange { return s.Range }
func (*EmptyStmt) isStmt()                   {}

type IntLit struct {
	Value int64
	T     Type
	Range entity.SourceRange
}

func (l *IntLit) Pos() entity.SourceRange  { return l.Range }
func (*IntLit) isExpr()                    {}
func (l *IntLit) GetType() Type            { return l.T }
func (*IntLit) GetCategory() ValueCategory { return RValue }

type FloatLit struct {
	Value float64
	T     Type
	Range entity.SourceRange
}

func (l *FloatLit) Pos() entity.SourceRange  { return l.Range }
func (*FloatLit) isExpr()                    {}
func (l *FloatLit) GetType() Type            { return l.T }
func (*FloatLit) GetCategory() ValueCategory { return RValue }

type CharLit struct {
	Value int32
	T     Type
	Range entity.SourceRange
}

func (l *CharLit) Pos() entity.SourceRange  { return l.Range }
func (*CharLit) isExpr()                    {}
func (l *CharLit) GetType() Type            { return l.T }
func (*CharLit) GetCategory() ValueCategory { return RValue }

type StringLit struct {
	Value string
	T     Type
	Range entity.SourceRange
}

func (l *StringLit) Pos() entity.SourceRange  { return l.Range }
func (*StringLit) isExpr()                    {}
func (l *StringLit) GetType() Type            { return l.T }
func (*StringLit) GetCategory() ValueCategory { return LValue }

type VarRef struct {
	Sym   *Symbol
	T     Type
	Range entity.SourceRange
}

func (e *VarRef) Pos() entity.SourceRange  { return e.Range }
func (*VarRef) isExpr()                    {}
func (e *VarRef) GetType() Type            { return e.T }
func (*VarRef) GetCategory() ValueCategory { return LValue }

type EnumRef struct {
	Enumerator *Enumerator
	T          Type
	Range      entity.SourceRange
}

func (e *EnumRef) Pos() entity.SourceRange  { return e.Range }
func (*EnumRef) isExpr()                    {}
func (e *EnumRef) GetType() Type            { return e.T }
func (*EnumRef) GetCategory() ValueCategory { return RValue }

type BinOp struct {
	Op    BinaryOp
	L, R  Expr
	T     Type
	Range entity.SourceRange
}

func (e *BinOp) Pos() entity.SourceRange  { return e.Range }
func (*BinOp) isExpr()                    {}
func (e *BinOp) GetType() Type            { return e.T }
func (*BinOp) GetCategory() ValueCategory { return RValue }

type UnOp struct {
	Op       UnaryOp
	X        Expr
	T        Type
	Category ValueCategory
	Range    entity.SourceRange
}

func (e *UnOp) Pos() entity.SourceRange    { return e.Range }
func (*UnOp) isExpr()                      {}
func (e *UnOp) GetType() Type              { return e.T }
func (e *UnOp) GetCategory() ValueCategory { return e.Category }

type AssignExpr struct {
	L, R  Expr
	T     Type
	Range entity.SourceRange
}

func (e *AssignExpr) Pos() entity.SourceRange  { return e.Range }
func (*AssignExpr) isExpr()                    {}
func (e *AssignExpr) GetType() Type            { return e.T }
func (*AssignExpr) GetCategory() ValueCategory { return RValue }

type CompoundAssign struct {
	Op    BinaryOp
	L, R  Expr
	T     Type
	Range entity.SourceRange
}

func (e *CompoundAssign) Pos() entity.SourceRange  { return e.Range }
func (*CompoundAssign) isExpr()                    {}
func (e *CompoundAssign) GetType() Type            { return e.T }
func (*CompoundAssign) GetCategory() ValueCategory { return RValue }

type CallExpr struct {
	Callee Expr
	Args   []Expr
	T      Type
	Range  entity.SourceRange
}

func (e *CallExpr) Pos() entity.SourceRange  { return e.Range }
func (*CallExpr) isExpr()                    {}
func (e *CallExpr) GetType() Type            { return e.T }
func (*CallExpr) GetCategory() ValueCategory { return RValue }

type MemberExpr struct {
	Base     Expr
	Field    *Field
	Arrow    bool
	T        Type
	Category ValueCategory
	Range    entity.SourceRange
}

func (e *MemberExpr) Pos() entity.SourceRange    { return e.Range }
func (*MemberExpr) isExpr()                      {}
func (e *MemberExpr) GetType() Type              { return e.T }
func (e *MemberExpr) GetCategory() ValueCategory { return e.Category }

type IndexExpr struct {
	Base, Index Expr
	T           Type
	Range       entity.SourceRange
}

func (e *IndexExpr) Pos() entity.SourceRange  { return e.Range }
func (*IndexExpr) isExpr()                    {}
func (e *IndexExpr) GetType() Type            { return e.T }
func (*IndexExpr) GetCategory() ValueCategory { return LValue }

type CondExpr struct {
	Cond, Then, Else Expr
	T                Type
	Range            entity.SourceRange
}

func (e *CondExpr) Pos() entity.SourceRange  { return e.Range }
func (*CondExpr) isExpr()                    {}
func (e *CondExpr) GetType() Type            { return e.T }
func (*CondExpr) GetCategory() ValueCategory { return RValue }

type SizeofOperand struct {
	Expr Expr
	Type Type
}

type SizeofExpr struct {
	Operand SizeofOperand
	T       Type
	Range   entity.SourceRange
}

func (e *SizeofExpr) Pos() entity.SourceRange  { return e.Range }
func (*SizeofExpr) isExpr()                    {}
func (e *SizeofExpr) GetType() Type            { return e.T }
func (*SizeofExpr) GetCategory() ValueCategory { return RValue }

type CommaExpr struct {
	L, R  Expr
	T     Type
	Range entity.SourceRange
}

func (e *CommaExpr) Pos() entity.SourceRange { return e.Range }
func (*CommaExpr) isExpr()                   {}
func (e *CommaExpr) GetType() Type           { return e.T }
func (e *CommaExpr) GetCategory() ValueCategory {
	return e.R.GetCategory()
}

type CompoundLit struct {
	T     Type
	Init  *InitList
	Range entity.SourceRange
}

func (e *CompoundLit) Pos() entity.SourceRange  { return e.Range }
func (*CompoundLit) isExpr()                    {}
func (e *CompoundLit) GetType() Type            { return e.T }
func (*CompoundLit) GetCategory() ValueCategory { return LValue }

type InitList struct {
	Elems []InitElem
	T     Type
	Range entity.SourceRange
}

func (e *InitList) Pos() entity.SourceRange  { return e.Range }
func (*InitList) isExpr()                    {}
func (e *InitList) GetType() Type            { return e.T }
func (*InitList) GetCategory() ValueCategory { return RValue }

type InitElem struct {
	Designators []Designator
	Value       Expr
}

type Designator struct {
	Kind  DesignatorKind
	Index int64
	Field *Field
}

type ImplicitCast struct {
	From, To Type
	X        Expr
	Kind     CastKind
	Range    entity.SourceRange
}

func (e *ImplicitCast) Pos() entity.SourceRange  { return e.Range }
func (*ImplicitCast) isExpr()                    {}
func (e *ImplicitCast) GetType() Type            { return e.To }
func (*ImplicitCast) GetCategory() ValueCategory { return RValue }

type ExplicitCast struct {
	To              Type
	X               Expr
	TypeNameTypedef bool
	Range           entity.SourceRange
}

func (e *ExplicitCast) Pos() entity.SourceRange  { return e.Range }
func (*ExplicitCast) isExpr()                    {}
func (e *ExplicitCast) GetType() Type            { return e.To }
func (*ExplicitCast) GetCategory() ValueCategory { return RValue }

type AddrConst struct {
	Sym    *Symbol
	Offset int64
	T      Type
	Range  entity.SourceRange
}

func (e *AddrConst) Pos() entity.SourceRange  { return e.Range }
func (*AddrConst) isExpr()                    {}
func (e *AddrConst) GetType() Type            { return e.T }
func (*AddrConst) GetCategory() ValueCategory { return RValue }

type VarDecl struct {
	Sym     *Symbol
	T       Type
	Init    Expr
	Storage StorageClass
	IsParam bool
	Range   entity.SourceRange
}

func (d *VarDecl) Pos() entity.SourceRange { return d.Range }
func (*VarDecl) isDecl()                   {}

type FuncDecl struct {
	Sym     *Symbol
	T       *FunctionType
	Storage StorageClass
	Range   entity.SourceRange
}

func (d *FuncDecl) Pos() entity.SourceRange { return d.Range }
func (*FuncDecl) isDecl()                   {}

type FuncDef struct {
	Sym    *Symbol
	T      *FunctionType
	Params []*VarDecl
	Body   *Block
	Locals []*VarDecl
	Labels map[string]*LabeledStmt
	Range  entity.SourceRange
}

func (d *FuncDef) Pos() entity.SourceRange { return d.Range }
func (*FuncDef) isDecl()                   {}

type TypedefDecl struct {
	Sym   *Symbol
	T     Type
	Range entity.SourceRange
}

func (d *TypedefDecl) Pos() entity.SourceRange { return d.Range }
func (*TypedefDecl) isDecl()                   {}

type TagDecl struct {
	Tag   *TagID
	T     Type
	Range entity.SourceRange
}

func (d *TagDecl) Pos() entity.SourceRange { return d.Range }
func (*TagDecl) isDecl()                   {}
