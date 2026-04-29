package sema

type BinaryOp int

const (
	OpAdd BinaryOp = iota
	OpSub
	OpMul
	OpDiv
	OpMod
	OpAnd
	OpOr
	OpXor
	OpShl
	OpShr
	OpLAnd
	OpLOr
	OpEq
	OpNe
	OpLt
	OpLe
	OpGt
	OpGe
)

type UnaryOp int

const (
	UnPlus UnaryOp = iota
	UnMinus
	UnBitNot
	UnLogNot
	UnAddr
	UnDeref
	UnIncPre
	UnIncPost
	UnDecPre
	UnDecPost
)

type CastKind int

const (
	LValueToRValue CastKind = iota
	ArrayDecay
	FunctionDecay
	IntegerPromotion
	UsualArithmetic
	IntegralConversion
	FloatingConversion
	IntToFloat
	FloatToInt
	IntToPointer
	PointerToInt
	PointerConversion
	VoidPointerConversion
	NullPointerConstant
	BoolConversion
	DefaultArgPromotion
	CastUnion
)

type DesignatorKind int

const (
	DesigArrayIndex DesignatorKind = iota
	DesigFieldName
)
