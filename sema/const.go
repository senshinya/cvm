package sema

import "math"

type ConstKind int

const (
	ConstInt ConstKind = iota
	ConstUint
	ConstFloat
	ConstComplex
	ConstAddress
	ConstString
)

type ConstValueAddr struct {
	Sym    *Symbol
	Offset int64
}

type ConstValue struct {
	Kind  ConstKind
	Int   int64
	Uint  uint64
	Float float64
	Addr  ConstValueAddr
	T     Type
}

type Evaluator struct {
	sema                        *Sema
	allowCompoundLiteralAddress bool
}

func NewEvaluator(s *Sema) *Evaluator {
	return &Evaluator{sema: s}
}

func (e *Evaluator) EvalIntegerConstant(expr Expr) (ConstValue, bool) {
	switch x := expr.(type) {
	case *IntLit:
		return ConstValue{Kind: ConstInt, Int: x.Value, Uint: uint64(x.Value), T: x.T}, true
	case *CharLit:
		return ConstValue{Kind: ConstInt, Int: int64(x.Value), Uint: uint64(x.Value), T: x.T}, true
	case *EnumRef:
		return ConstValue{Kind: ConstInt, Int: x.Enumerator.Value, Uint: uint64(x.Enumerator.Value), T: x.T}, true
	case *SizeofExpr:
		if x.Operand.Type != nil {
			return ConstValue{Kind: ConstInt, Int: sizeofType(x.Operand.Type), T: x.T}, true
		}
		if x.Operand.Expr != nil {
			return ConstValue{Kind: ConstInt, Int: sizeofType(x.Operand.Expr.GetType()), T: x.T}, true
		}
	case *BinOp:
		l, lok := e.EvalIntegerConstant(x.L)
		r, rok := e.EvalIntegerConstant(x.R)
		if !lok || !rok {
			return ConstValue{}, false
		}
		cv, ok := evalC99BinOpConst(x.Op, l, r, x.L.GetType(), x.T)
		if !ok {
			return ConstValue{}, false
		}
		return cv, true
	case *UnOp:
		v, ok := e.EvalIntegerConstant(x.X)
		if !ok {
			return ConstValue{}, false
		}
		switch x.Op {
		case UnPlus:
			return v, true
		case UnMinus:
			if !validC99UnaryMinus(x.T, v.Int) {
				return ConstValue{}, false
			}
			return ConstValue{Kind: ConstInt, Int: -v.Int, Uint: uint64(-v.Int), T: x.T}, true
		case UnBitNot:
			return ConstValue{Kind: ConstInt, Int: ^v.Int, Uint: uint64(^v.Int), T: x.T}, true
		case UnLogNot:
			return ConstValue{Kind: ConstInt, Int: boolToInt(v.Int == 0), T: x.T}, true
		}
	case *CondExpr:
		c, ok := e.EvalIntegerConstant(x.Cond)
		if !ok {
			return ConstValue{}, false
		}
		if c.Int != 0 {
			return e.EvalIntegerConstant(x.Then)
		}
		return e.EvalIntegerConstant(x.Else)
	case *ImplicitCast:
		cv, ok := e.EvalIntegerConstant(x.X)
		if !ok || !isInteger(x.To) {
			return ConstValue{}, false
		}
		return castConstInteger(cv, x.To), true
	case *ExplicitCast:
		if v, ok := e.EvalIntegerConstant(x.X); ok {
			return castConstInteger(v, x.To), true
		}
		if f, ok := e.EvalArithmetic(x.X); ok && f.Kind == ConstFloat {
			return castConstInteger(ConstValue{Kind: ConstInt, Int: int64(f.Float), Uint: uint64(int64(f.Float)), T: x.To}, x.To), true
		}
	}
	return ConstValue{}, false
}

func (e *Evaluator) EvalC99IntegerConstantExpression(expr Expr) (ConstValue, bool) {
	cv, ok := e.evalC99IntegerConstantExpression(expr)
	if !ok || !isInteger(cv.T) {
		return ConstValue{}, false
	}
	return cv, true
}

// C99 的 ICE 允许不求值分支里的副作用表达式，但实际选中的分支和固定数组大小
// 必须能折成整数；这里故意独立于更宽松的 EvalIntegerConstant。
func (e *Evaluator) evalC99IntegerConstantExpression(expr Expr) (ConstValue, bool) {
	switch x := expr.(type) {
	case *IntLit:
		return ConstValue{Kind: ConstInt, Int: x.Value, Uint: uint64(x.Value), T: x.T}, true
	case *CharLit:
		return ConstValue{Kind: ConstInt, Int: int64(x.Value), Uint: uint64(x.Value), T: x.T}, true
	case *EnumRef:
		return ConstValue{Kind: ConstInt, Int: x.Enumerator.Value, Uint: uint64(x.Enumerator.Value), T: x.T}, true
	case *SizeofExpr:
		if x.Operand.Type != nil {
			if typeHasVariableSize(x.Operand.Type) {
				return ConstValue{}, false
			}
			size := sizeofType(x.Operand.Type)
			if size <= 0 {
				return ConstValue{}, false
			}
			return ConstValue{Kind: ConstInt, Int: size, Uint: uint64(size), T: x.T}, true
		}
		if x.Operand.Expr != nil {
			if typeHasVariableSize(x.Operand.Expr.GetType()) {
				return ConstValue{}, false
			}
			size := sizeofType(x.Operand.Expr.GetType())
			if size <= 0 {
				return ConstValue{}, false
			}
			return ConstValue{Kind: ConstInt, Int: size, Uint: uint64(size), T: x.T}, true
		}
	case *BinOp:
		l, lok := e.evalC99IntegerConstantExpression(x.L)
		if !lok {
			return ConstValue{}, false
		}
		// C99 短路分支不求值，但仍必须只由常量表达式允许的操作数组成。
		switch x.Op {
		case OpLAnd:
			if l.Int == 0 {
				if !e.isC99UnevaluatedIntegerConstantOperand(x.R) {
					return ConstValue{}, false
				}
				return ConstValue{Kind: ConstInt, Int: 0, Uint: 0, T: x.T}, true
			}
			r, ok := e.evalC99IntegerConstantExpression(x.R)
			if !ok {
				return ConstValue{}, false
			}
			v := boolToInt(r.Int != 0)
			return ConstValue{Kind: ConstInt, Int: v, Uint: uint64(v), T: x.T}, true
		case OpLOr:
			if l.Int != 0 {
				if !e.isC99UnevaluatedIntegerConstantOperand(x.R) {
					return ConstValue{}, false
				}
				return ConstValue{Kind: ConstInt, Int: 1, Uint: 1, T: x.T}, true
			}
			r, ok := e.evalC99IntegerConstantExpression(x.R)
			if !ok {
				return ConstValue{}, false
			}
			v := boolToInt(r.Int != 0)
			return ConstValue{Kind: ConstInt, Int: v, Uint: uint64(v), T: x.T}, true
		}
		r, rok := e.evalC99IntegerConstantExpression(x.R)
		if !rok {
			return ConstValue{}, false
		}
		if x.Op == OpShl && !validC99LeftShift(x.L.GetType(), l.Int, r.Int) {
			return ConstValue{}, false
		}
		cv, ok := evalC99BinOpConst(x.Op, l, r, x.L.GetType(), x.T)
		if !ok {
			return ConstValue{}, false
		}
		return cv, true
	case *UnOp:
		v, ok := e.evalC99IntegerConstantExpression(x.X)
		if !ok {
			return ConstValue{}, false
		}
		switch x.Op {
		case UnPlus:
			return ConstValue{Kind: ConstInt, Int: v.Int, Uint: uint64(v.Int), T: x.T}, true
		case UnMinus:
			if !validC99UnaryMinus(x.T, v.Int) {
				return ConstValue{}, false
			}
			return ConstValue{Kind: ConstInt, Int: -v.Int, Uint: uint64(-v.Int), T: x.T}, true
		case UnBitNot:
			return ConstValue{Kind: ConstInt, Int: ^v.Int, Uint: uint64(^v.Int), T: x.T}, true
		case UnLogNot:
			return ConstValue{Kind: ConstInt, Int: boolToInt(v.Int == 0), Uint: uint64(boolToInt(v.Int == 0)), T: x.T}, true
		}
	case *CondExpr:
		c, ok := e.evalC99IntegerConstantExpression(x.Cond)
		if !ok {
			return ConstValue{}, false
		}
		if c.Int != 0 {
			if !e.isC99UnevaluatedIntegerConstantOperand(x.Else) {
				return ConstValue{}, false
			}
			return e.evalC99IntegerConstantExpression(x.Then)
		}
		if !e.isC99UnevaluatedIntegerConstantOperand(x.Then) {
			return ConstValue{}, false
		}
		return e.evalC99IntegerConstantExpression(x.Else)
	case *ImplicitCast:
		cv, ok := e.evalC99IntegerConstantExpression(x.X)
		if !ok || !isInteger(x.To) {
			return ConstValue{}, false
		}
		return castConstInteger(cv, x.To), true
	case *ExplicitCast:
		if !isInteger(x.To) {
			return ConstValue{}, false
		}
		if cv, ok := e.evalC99IntegerConstantExpression(x.X); ok {
			return castConstInteger(cv, x.To), true
		}
		if f, ok := x.X.(*FloatLit); ok {
			v := int64(f.Value)
			return castConstInteger(ConstValue{Kind: ConstInt, Int: v, Uint: uint64(v), T: x.To}, x.To), true
		}
	}
	return ConstValue{}, false
}

func (e *Evaluator) isC99UnevaluatedIntegerConstantOperand(expr Expr) bool {
	switch x := expr.(type) {
	case *IntLit, *CharLit, *EnumRef:
		return true
	case *SizeofExpr:
		return true
	case *BinOp:
		return isInteger(x.T) &&
			e.isC99UnevaluatedIntegerConstantOperand(x.L) &&
			e.isC99UnevaluatedIntegerConstantOperand(x.R)
	case *UnOp:
		switch x.Op {
		case UnPlus, UnMinus, UnBitNot, UnLogNot:
			return isInteger(x.T) && e.isC99UnevaluatedIntegerConstantOperand(x.X)
		}
	case *CondExpr:
		return isInteger(x.T) &&
			e.isC99UnevaluatedIntegerConstantOperand(x.Cond) &&
			e.isC99UnevaluatedIntegerConstantOperand(x.Then) &&
			e.isC99UnevaluatedIntegerConstantOperand(x.Else)
	case *ImplicitCast:
		return isInteger(x.To) && e.isC99UnevaluatedIntegerConstantOperand(x.X)
	case *ExplicitCast:
		if !isInteger(x.To) {
			return false
		}
		if _, ok := x.X.(*FloatLit); ok {
			return true
		}
		return e.isC99UnevaluatedIntegerConstantOperand(x.X)
	case *CommaExpr:
		// 未求值子表达式里的逗号可以出现；它的操作数仍不能引用运行期值。
		return isInteger(x.T) &&
			e.isC99UnevaluatedIntegerConstantOperand(x.L) &&
			e.isC99UnevaluatedIntegerConstantOperand(x.R)
	}
	return false
}

func (e *Evaluator) EvalC99ArraySizeConstantExpression(expr Expr) (ConstValue, bool) {
	if cv, ok := e.EvalC99IntegerConstantExpression(expr); ok {
		return cv, true
	}
	if x, ok := expr.(*ExplicitCast); ok && isInteger(x.To) {
		if cv, ok := e.evalC99CastArithmeticConstant(x.X, true, false, false); ok {
			if cv.Kind == ConstFloat {
				v := int64(cv.Float)
				return ConstValue{Kind: ConstInt, Int: v, Uint: uint64(v), T: x.To}, true
			}
			return ConstValue{Kind: ConstInt, Int: cv.Int, Uint: uint64(cv.Int), T: x.To}, true
		}
	}
	return ConstValue{}, false
}

func (e *Evaluator) EvalC99ArithmeticConstantExpression(expr Expr) (ConstValue, bool) {
	cv, ok := e.evalC99CastArithmeticConstant(expr, true, true, true)
	if !ok || !isArithmetic(cv.T) {
		return ConstValue{}, false
	}
	return cv, true
}

// GCC/C99 把直接的 (int)1.0 当作 ICE；一元 +/- 或 (double) 这类中间算术 cast
// 只能作为普通算术常量表达式。数组大小正负检查还要避免把浮点二元表达式当作已知值。
func (e *Evaluator) evalC99CastArithmeticConstant(expr Expr, allowUnaryFloat, allowFloatBinOp, allowUnevaluatedNonConstant bool) (ConstValue, bool) {
	switch x := expr.(type) {
	case *FloatLit:
		return ConstValue{Kind: ConstFloat, Float: x.Value, T: x.T}, true
	case *ExplicitCast:
		if !isArithmetic(x.To) {
			return ConstValue{}, false
		}
		cv, ok := e.evalC99CastArithmeticConstant(x.X, allowUnaryFloat, allowFloatBinOp, allowUnevaluatedNonConstant)
		if !ok {
			return ConstValue{}, false
		}
		return castC99ArithmeticConstant(cv, x.To)
	case *ImplicitCast:
		if !isArithmetic(x.To) {
			return ConstValue{}, false
		}
		cv, ok := e.evalC99CastArithmeticConstant(x.X, allowUnaryFloat, allowFloatBinOp, allowUnevaluatedNonConstant)
		if !ok {
			return ConstValue{}, false
		}
		return castC99ArithmeticConstant(cv, x.To)
	case *UnOp:
		cv, ok := e.evalC99CastArithmeticConstant(x.X, allowUnaryFloat, allowFloatBinOp, allowUnevaluatedNonConstant)
		if !ok {
			return ConstValue{}, false
		}
		switch x.Op {
		case UnPlus:
			if cv.Kind == ConstFloat && !allowUnaryFloat {
				return ConstValue{}, false
			}
			return cv, true
		case UnMinus:
			if cv.Kind == ConstFloat {
				if !allowUnaryFloat {
					return ConstValue{}, false
				}
				return ConstValue{Kind: ConstFloat, Float: -cv.Float, T: x.T}, true
			}
			if !validC99UnaryMinus(x.T, cv.Int) {
				return ConstValue{}, false
			}
			return ConstValue{Kind: ConstInt, Int: -cv.Int, Uint: uint64(-cv.Int), T: x.T}, true
		case UnLogNot:
			return ConstValue{Kind: ConstInt, Int: boolToInt(!constNonZero(cv)), Uint: uint64(boolToInt(!constNonZero(cv))), T: x.T}, true
		case UnBitNot:
			if cv.Kind != ConstInt {
				return ConstValue{}, false
			}
			return ConstValue{Kind: ConstInt, Int: ^cv.Int, Uint: uint64(^cv.Int), T: x.T}, true
		}
	case *BinOp:
		return e.evalC99ArithmeticBinOp(x, allowUnaryFloat, allowFloatBinOp, allowUnevaluatedNonConstant)
	case *CondExpr:
		c, ok := e.evalC99CastArithmeticConstant(x.Cond, allowUnaryFloat, allowFloatBinOp, allowUnevaluatedNonConstant)
		if !ok {
			return ConstValue{}, false
		}
		if constNonZero(c) {
			if !allowUnevaluatedNonConstant && !e.isC99UnevaluatedArithmeticConstantOperand(x.Else) {
				return ConstValue{}, false
			}
			return e.evalC99CastArithmeticConstant(x.Then, allowUnaryFloat, allowFloatBinOp, allowUnevaluatedNonConstant)
		}
		if !allowUnevaluatedNonConstant && !e.isC99UnevaluatedArithmeticConstantOperand(x.Then) {
			return ConstValue{}, false
		}
		return e.evalC99CastArithmeticConstant(x.Else, allowUnaryFloat, allowFloatBinOp, allowUnevaluatedNonConstant)
	default:
		return e.evalC99IntegerConstantExpression(expr)
	}
	return ConstValue{}, false
}

func (e *Evaluator) evalC99ArithmeticBinOp(x *BinOp, allowUnaryFloat, allowFloatBinOp, allowUnevaluatedNonConstant bool) (ConstValue, bool) {
	l, lok := e.evalC99CastArithmeticConstant(x.L, allowUnaryFloat, allowFloatBinOp, allowUnevaluatedNonConstant)
	if !lok {
		return ConstValue{}, false
	}
	// 算术常量表达式同样遵循短路求值；未求值右侧不应触发约束错误。
	switch x.Op {
	case OpLAnd:
		if !constNonZero(l) {
			if !allowUnevaluatedNonConstant && !e.isC99UnevaluatedArithmeticConstantOperand(x.R) {
				return ConstValue{}, false
			}
			return ConstValue{Kind: ConstInt, Int: 0, Uint: 0, T: x.T}, true
		}
		r, ok := e.evalC99CastArithmeticConstant(x.R, allowUnaryFloat, allowFloatBinOp, allowUnevaluatedNonConstant)
		if !ok {
			return ConstValue{}, false
		}
		v := boolToInt(constNonZero(r))
		return ConstValue{Kind: ConstInt, Int: v, Uint: uint64(v), T: x.T}, true
	case OpLOr:
		if constNonZero(l) {
			if !allowUnevaluatedNonConstant && !e.isC99UnevaluatedArithmeticConstantOperand(x.R) {
				return ConstValue{}, false
			}
			return ConstValue{Kind: ConstInt, Int: 1, Uint: 1, T: x.T}, true
		}
		r, ok := e.evalC99CastArithmeticConstant(x.R, allowUnaryFloat, allowFloatBinOp, allowUnevaluatedNonConstant)
		if !ok {
			return ConstValue{}, false
		}
		v := boolToInt(constNonZero(r))
		return ConstValue{Kind: ConstInt, Int: v, Uint: uint64(v), T: x.T}, true
	}
	r, rok := e.evalC99CastArithmeticConstant(x.R, allowUnaryFloat, allowFloatBinOp, allowUnevaluatedNonConstant)
	if !rok {
		return ConstValue{}, false
	}
	if l.Kind == ConstFloat || r.Kind == ConstFloat {
		if !allowFloatBinOp {
			return ConstValue{}, false
		}
		return evalC99FloatArithmeticBinOp(x.Op, constToFloat(l), constToFloat(r), x.T)
	}
	if x.Op == OpShl && !validC99LeftShift(x.L.GetType(), l.Int, r.Int) {
		return ConstValue{}, false
	}
	v, ok := evalC99BinOpInt(x.Op, l.Int, r.Int, x.T)
	if !ok {
		return ConstValue{}, false
	}
	return ConstValue{Kind: ConstInt, Int: v, Uint: uint64(v), T: x.T}, true
}

func (e *Evaluator) isC99UnevaluatedArithmeticConstantOperand(expr Expr) bool {
	switch x := expr.(type) {
	case *FloatLit:
		return true
	case *IntLit, *CharLit, *EnumRef:
		return true
	case *SizeofExpr:
		return true
	case *BinOp:
		return isArithmetic(x.T) &&
			e.isC99UnevaluatedArithmeticConstantOperand(x.L) &&
			e.isC99UnevaluatedArithmeticConstantOperand(x.R)
	case *UnOp:
		switch x.Op {
		case UnPlus, UnMinus, UnBitNot, UnLogNot:
			return isArithmetic(x.T) && e.isC99UnevaluatedArithmeticConstantOperand(x.X)
		}
	case *CondExpr:
		return isArithmetic(x.T) &&
			e.isC99UnevaluatedArithmeticConstantOperand(x.Cond) &&
			e.isC99UnevaluatedArithmeticConstantOperand(x.Then) &&
			e.isC99UnevaluatedArithmeticConstantOperand(x.Else)
	case *ImplicitCast:
		return isArithmetic(x.To) && e.isC99UnevaluatedArithmeticConstantOperand(x.X)
	case *ExplicitCast:
		return isArithmetic(x.To) && e.isC99UnevaluatedArithmeticConstantOperand(x.X)
	case *CommaExpr:
		// 未求值算术分支同样只做形状检查，避免把逗号里的除零或负移位当作已求值错误。
		return isArithmetic(x.T) &&
			e.isC99UnevaluatedArithmeticConstantOperand(x.L) &&
			e.isC99UnevaluatedArithmeticConstantOperand(x.R)
	}
	return false
}

func castC99ArithmeticConstant(cv ConstValue, to Type) (ConstValue, bool) {
	if !isArithmetic(to) {
		return ConstValue{}, false
	}
	if isInteger(to) {
		if cv.Kind == ConstFloat {
			v := int64(cv.Float)
			return ConstValue{Kind: ConstInt, Int: v, Uint: uint64(v), T: to}, true
		}
		return ConstValue{Kind: ConstInt, Int: cv.Int, Uint: uint64(cv.Int), T: to}, true
	}
	if cv.Kind == ConstFloat {
		return ConstValue{Kind: ConstFloat, Float: cv.Float, T: to}, true
	}
	return ConstValue{Kind: ConstFloat, Float: float64(cv.Int), T: to}, true
}

func evalC99FloatArithmeticBinOp(op BinaryOp, l, r float64, t Type) (ConstValue, bool) {
	switch op {
	case OpAdd:
		return ConstValue{Kind: ConstFloat, Float: l + r, T: t}, true
	case OpSub:
		return ConstValue{Kind: ConstFloat, Float: l - r, T: t}, true
	case OpMul:
		return ConstValue{Kind: ConstFloat, Float: l * r, T: t}, true
	case OpDiv:
		if r == 0 {
			return ConstValue{}, false
		}
		return ConstValue{Kind: ConstFloat, Float: l / r, T: t}, true
	case OpEq:
		v := boolToInt(l == r)
		return ConstValue{Kind: ConstInt, Int: v, Uint: uint64(v), T: t}, true
	case OpNe:
		v := boolToInt(l != r)
		return ConstValue{Kind: ConstInt, Int: v, Uint: uint64(v), T: t}, true
	case OpLt:
		v := boolToInt(l < r)
		return ConstValue{Kind: ConstInt, Int: v, Uint: uint64(v), T: t}, true
	case OpLe:
		v := boolToInt(l <= r)
		return ConstValue{Kind: ConstInt, Int: v, Uint: uint64(v), T: t}, true
	case OpGt:
		v := boolToInt(l > r)
		return ConstValue{Kind: ConstInt, Int: v, Uint: uint64(v), T: t}, true
	case OpGe:
		v := boolToInt(l >= r)
		return ConstValue{Kind: ConstInt, Int: v, Uint: uint64(v), T: t}, true
	}
	return ConstValue{}, false
}

func constToFloat(cv ConstValue) float64 {
	if cv.Kind == ConstFloat {
		return cv.Float
	}
	return float64(cv.Int)
}

func constNonZero(cv ConstValue) bool {
	if cv.Kind == ConstFloat {
		return cv.Float != 0
	}
	return cv.Int != 0
}

func isSignedIntegerType(t Type) bool {
	bt, ok := unqualifiedBuiltin(t)
	if !ok || !isInteger(t) {
		return false
	}
	switch bt.Kind {
	case Bool, UChar, UShort, UInt, ULong, ULongLong:
		return false
	}
	return true
}

func validC99LeftShift(t Type, left, right int64) bool {
	if right < 0 {
		return false
	}
	if !isSignedIntegerType(t) {
		return true
	}
	if left < 0 {
		return false
	}
	bits := sizeofType(t) * 8
	if bits <= 0 || right >= bits {
		return false
	}
	if bits >= 63 {
		return right < 63 && left <= (int64Max>>uint(right))
	}
	max := (int64(1) << uint(bits-1)) - 1
	return left <= (max >> uint(right))
}

const int64Max = int64(^uint64(0) >> 1)
const int64Min = -int64Max - 1

func evalC99BinOpInt(op BinaryOp, l, r int64, t Type) (int64, bool) {
	if !validC99SignedBinOp(op, l, r, t) {
		return 0, false
	}
	v, ok := evalBinOpInt(op, l, r)
	if !ok || !valueFitsSignedIntegerType(t, v) {
		return 0, false
	}
	return v, true
}

func validC99UnaryMinus(t Type, v int64) bool {
	min, _, ok := signedIntegerRange(t)
	return !ok || v != min
}

func validC99SignedBinOp(op BinaryOp, l, r int64, t Type) bool {
	min, max, ok := signedIntegerRange(t)
	if !ok {
		return true
	}
	switch op {
	case OpAdd:
		return (r <= 0 || l <= max-r) && (r >= 0 || l >= min-r)
	case OpSub:
		return (r >= 0 || l <= max+r) && (r <= 0 || l >= min+r)
	case OpMul:
		if l == 0 || r == 0 {
			return true
		}
		v := l * r
		return v/r == l && v >= min && v <= max
	case OpDiv, OpMod:
		return !(l == min && r == -1)
	}
	return true
}

func valueFitsSignedIntegerType(t Type, v int64) bool {
	min, max, ok := signedIntegerRange(t)
	return !ok || (v >= min && v <= max)
}

func signedIntegerRange(t Type) (int64, int64, bool) {
	if !isSignedIntegerType(t) {
		return 0, 0, false
	}
	bits := sizeofType(t) * 8
	if bits >= 64 {
		return int64Min, int64Max, true
	}
	if bits <= 0 {
		return 0, 0, false
	}
	max := (int64(1) << uint(bits-1)) - 1
	return -max - 1, max, true
}

func (e *Evaluator) EvalConstant(expr Expr) (ConstValue, bool) {
	if cv, ok := e.EvalC99IntegerConstantExpression(expr); ok {
		return cv, true
	}
	if isArithmetic(expr.GetType()) {
		if cv, ok := e.EvalC99ArithmeticConstantExpression(expr); ok {
			return cv, true
		}
	}
	switch x := expr.(type) {
	case *FloatLit:
		return ConstValue{Kind: ConstFloat, Float: x.Value, T: x.T}, true
	case *ImagLit:
		return ConstValue{Kind: ConstComplex, T: x.T}, true
	case *StringLit:
		return ConstValue{Kind: ConstString, T: x.T}, true
	case *UnOp:
		if x.Op != UnAddr {
			return ConstValue{}, false
		}
		if vr, ok := x.X.(*VarRef); ok && vr.Sym.Storage != StorageAuto && vr.Sym.Storage != StorageRegister {
			return ConstValue{Kind: ConstAddress, Addr: ConstValueAddr{Sym: vr.Sym}, T: x.T}, true
		}
		if cl, ok := x.X.(*CompoundLit); ok && e.allowCompoundLiteralAddress && e.initListIsStaticConstant(cl.Init) {
			return ConstValue{Kind: ConstAddress, T: x.T}, true
		}
	case *BinOp:
		if x.Op != OpAdd && x.Op != OpSub {
			return ConstValue{}, false
		}
		l, lok := e.EvalConstant(x.L)
		r, rok := e.EvalConstant(x.R)
		if !lok || !rok {
			return ConstValue{}, false
		}
		ptr, ok := x.T.(*PointerType)
		if !ok {
			return ConstValue{}, false
		}
		// VLA 步长依赖运行期 sizeof，不能把 VM 指针算术折成静态地址偏移。
		if typeHasVariableSize(ptr.Pointee) {
			return ConstValue{}, false
		}
		scale := sizeofType(ptr.Pointee)
		switch {
		case l.Kind == ConstAddress && r.Kind == ConstInt:
			return ConstValue{Kind: ConstAddress, Addr: ConstValueAddr{Sym: l.Addr.Sym, Offset: l.Addr.Offset + signedOp(x.Op)*r.Int*scale}, T: x.T}, true
		case x.Op == OpAdd && l.Kind == ConstInt && r.Kind == ConstAddress:
			return ConstValue{Kind: ConstAddress, Addr: ConstValueAddr{Sym: r.Addr.Sym, Offset: r.Addr.Offset + l.Int*scale}, T: x.T}, true
		}
	case *ImplicitCast:
		if x.Kind == ArrayDecay {
			if cv, ok := e.staticAddressConstant(x.X, x.To); ok {
				return cv, true
			}
		}
		if x.Kind == FunctionDecay {
			if cv, ok := e.staticAddressConstant(x.X, x.To); ok {
				return cv, true
			}
		}
		return e.EvalConstant(x.X)
	case *ExplicitCast:
		if isArithmetic(x.To) {
			if cv, ok := e.evalC99CastArithmeticConstant(x.X, true, true, true); ok {
				return castC99ArithmeticConstant(cv, x.To)
			}
		}
		if isPointer(x.To) {
			if typeHasForbiddenAddressConstantVMSize(x.To) {
				return ConstValue{}, false
			}
			if cv, ok := e.EvalConstant(x.X); ok && cv.Kind == ConstAddress {
				cv.T = x.To
				return cv, true
			}
			if cv, ok := e.staticAddressConstant(x.X, x.To); ok {
				return cv, true
			}
		}
		return e.EvalConstant(x.X)
	case *CallExpr:
		if cv, ok := e.evalBuiltinConstantCall(x); ok {
			return cv, true
		}
	case *AddrConst:
		return ConstValue{Kind: ConstAddress, Addr: ConstValueAddr{Sym: x.Sym, Offset: x.Offset}, T: x.T}, true
	}
	return ConstValue{}, false
}

func (e *Evaluator) evalBuiltinConstantCall(call *CallExpr) (ConstValue, bool) {
	callee := call.Callee
	for {
		ic, ok := callee.(*ImplicitCast)
		if !ok || (ic.Kind != FunctionDecay && ic.Kind != LValueToRValue) {
			break
		}
		callee = ic.X
	}
	ref, ok := callee.(*VarRef)
	if !ok || ref.Sym == nil {
		return ConstValue{}, false
	}
	switch ref.Sym.Name {
	case "__builtin_nan", "nan":
		return ConstValue{Kind: ConstFloat, Float: math.NaN(), T: call.T}, true
	case "__builtin_huge_val", "__builtin_huge_valf", "__builtin_huge_vall":
		return ConstValue{Kind: ConstFloat, Float: math.Inf(1), T: call.T}, true
	case "__builtin_complex":
		if len(call.Args) != 2 {
			return ConstValue{}, false
		}
		for _, arg := range call.Args {
			if _, ok := e.EvalConstant(arg); !ok {
				return ConstValue{}, false
			}
		}
		return ConstValue{Kind: ConstComplex, T: call.T}, true
	}
	return ConstValue{}, false
}

func (e *Evaluator) staticAddressConstant(expr Expr, t Type) (ConstValue, bool) {
	if ic, ok := expr.(*ImplicitCast); ok && ic.Kind == LValueToRValue {
		expr = ic.X
	}
	switch x := expr.(type) {
	case *CompoundLit:
		if e.allowCompoundLiteralAddress && e.initListIsStaticConstant(x.Init) {
			return ConstValue{Kind: ConstAddress, T: t}, true
		}
	case *VarRef:
		if x.Sym.Kind == SymFunc || (x.Sym.Storage != StorageAuto && x.Sym.Storage != StorageRegister) {
			switch unqual(x.GetType()).(type) {
			case *ArrayType, *FunctionType:
				return ConstValue{Kind: ConstAddress, Addr: ConstValueAddr{Sym: x.Sym}, T: t}, true
			}
		}
	}
	return ConstValue{}, false
}

func (e *Evaluator) initListIsStaticConstant(il *InitList) bool {
	if il == nil {
		return true
	}
	// 文件作用域 compound literal 可提供静态地址，但它自己的初始化器仍需满足
	// 静态初始化约束；否则 &(int){g} 会绕过叶子表达式检查。
	for _, elem := range il.Elems {
		if nested, ok := elem.Value.(*InitList); ok {
			if !e.initListIsStaticConstant(nested) {
				return false
			}
			continue
		}
		if _, ok := e.EvalConstant(elem.Value); !ok {
			return false
		}
	}
	return true
}

func (e *Evaluator) EvalArithmetic(expr Expr) (ConstValue, bool) {
	return e.EvalConstant(expr)
}

func evalBinOpInt(op BinaryOp, l, r int64) (int64, bool) {
	switch op {
	case OpAdd:
		return l + r, true
	case OpSub:
		return l - r, true
	case OpMul:
		return l * r, true
	case OpDiv:
		if r == 0 {
			return 0, false
		}
		return l / r, true
	case OpMod:
		if r == 0 {
			return 0, false
		}
		return l % r, true
	case OpAnd:
		return l & r, true
	case OpOr:
		return l | r, true
	case OpXor:
		return l ^ r, true
	case OpShl:
		if r < 0 || r >= 64 {
			return 0, false
		}
		return l << uint(r), true
	case OpShr:
		if r < 0 || r >= 64 {
			return 0, false
		}
		return l >> uint(r), true
	case OpEq:
		return boolToInt(l == r), true
	case OpNe:
		return boolToInt(l != r), true
	case OpLt:
		return boolToInt(l < r), true
	case OpLe:
		return boolToInt(l <= r), true
	case OpGt:
		return boolToInt(l > r), true
	case OpGe:
		return boolToInt(l >= r), true
	case OpLAnd:
		return boolToInt(l != 0 && r != 0), true
	case OpLOr:
		return boolToInt(l != 0 || r != 0), true
	}
	return 0, false
}

func evalC99BinOpConst(op BinaryOp, l, r ConstValue, operandType, resultType Type) (ConstValue, bool) {
	if isUnsignedIntegerType(operandType) {
		v, ok := evalBinOpUint(op, l.Uint, r.Uint)
		if !ok {
			return ConstValue{}, false
		}
		if isRelationalConstOp(op) || op == OpLAnd || op == OpLOr {
			return ConstValue{Kind: ConstInt, Int: int64(v), Uint: v, T: resultType}, true
		}
		return castConstInteger(ConstValue{Kind: ConstInt, Int: int64(v), Uint: v, T: resultType}, resultType), true
	}
	v, ok := evalC99BinOpInt(op, l.Int, r.Int, resultType)
	if !ok {
		return ConstValue{}, false
	}
	return ConstValue{Kind: ConstInt, Int: v, Uint: uint64(v), T: resultType}, true
}

func evalBinOpUint(op BinaryOp, l, r uint64) (uint64, bool) {
	switch op {
	case OpAdd:
		return l + r, true
	case OpSub:
		return l - r, true
	case OpMul:
		return l * r, true
	case OpDiv:
		if r == 0 {
			return 0, false
		}
		return l / r, true
	case OpMod:
		if r == 0 {
			return 0, false
		}
		return l % r, true
	case OpAnd:
		return l & r, true
	case OpOr:
		return l | r, true
	case OpXor:
		return l ^ r, true
	case OpShl:
		if r >= 64 {
			return 0, false
		}
		return l << uint(r), true
	case OpShr:
		if r >= 64 {
			return 0, false
		}
		return l >> uint(r), true
	case OpEq:
		return uint64(boolToInt(l == r)), true
	case OpNe:
		return uint64(boolToInt(l != r)), true
	case OpLt:
		return uint64(boolToInt(l < r)), true
	case OpLe:
		return uint64(boolToInt(l <= r)), true
	case OpGt:
		return uint64(boolToInt(l > r)), true
	case OpGe:
		return uint64(boolToInt(l >= r)), true
	case OpLAnd:
		return uint64(boolToInt(l != 0 && r != 0)), true
	case OpLOr:
		return uint64(boolToInt(l != 0 || r != 0)), true
	}
	return 0, false
}

func isRelationalConstOp(op BinaryOp) bool {
	switch op {
	case OpEq, OpNe, OpLt, OpLe, OpGt, OpGe:
		return true
	}
	return false
}

func castConstInteger(cv ConstValue, to Type) ConstValue {
	u := cv.Uint
	if isSignedIntegerType(cv.T) {
		u = uint64(cv.Int)
	}
	if bits := integerValueBitsOfType(to); bits > 0 && bits < 64 {
		u &= (uint64(1) << uint(bits)) - 1
	}
	if isSignedIntegerType(to) {
		bits := integerValueBitsOfType(to)
		if bits > 0 && bits < 64 {
			sign := uint64(1) << uint(bits-1)
			if u&sign != 0 {
				return ConstValue{Kind: ConstInt, Int: int64(u | ^((uint64(1) << uint(bits)) - 1)), Uint: u, T: to}
			}
		}
		return ConstValue{Kind: ConstInt, Int: int64(u), Uint: u, T: to}
	}
	return ConstValue{Kind: ConstInt, Int: int64(u), Uint: u, T: to}
}

func isUnsignedIntegerType(t Type) bool {
	bt, ok := unqualifiedBuiltin(t)
	if !ok {
		return false
	}
	switch bt.Kind {
	case Bool, UChar, UShort, UInt, ULong, ULongLong:
		return true
	}
	return false
}

func integerValueBitsOfType(t Type) int {
	bt, ok := unqualifiedBuiltin(t)
	if !ok {
		return 0
	}
	return integerValueBits(bt.Kind)
}

func boolToInt(b bool) int64 {
	if b {
		return 1
	}
	return 0
}

func signedOp(op BinaryOp) int64 {
	if op == OpSub {
		return -1
	}
	return 1
}

// sizeof(VLA) 会求值长度表达式，不能把这种结果当作 C99 ICE；固定大小数组
// 仍然通过 sizeofType 折叠。
func typeHasVariableSize(t Type) bool {
	switch x := unqual(t).(type) {
	case *ArrayType:
		return x.SizeKind == ArrayVLA || x.SizeKind == ArrayStarSize || typeHasVariableSize(x.Elem)
	case *PointerType:
		// 指向变长类型的指针本身仍是固定大小；sizeof(int (*)[n]) 不会求值 n。
		return false
	case *QualType:
		return typeHasVariableSize(x.Base)
	}
	return false
}

func typeHasVariablyModifiedType(t Type) bool {
	switch x := unqual(t).(type) {
	case *ArrayType:
		return x.SizeKind == ArrayVLA || x.SizeKind == ArrayStarSize || typeHasVariablyModifiedType(x.Elem)
	case *PointerType:
		// 指向 VLA 的指针自身大小固定，但类型仍是 variably modified type。
		return typeHasVariablyModifiedType(x.Pointee)
	case *FunctionType:
		if typeHasVariablyModifiedType(x.Ret) {
			return true
		}
		for _, p := range x.Params {
			if typeHasVariablyModifiedType(p) {
				return true
			}
		}
	case *QualType:
		return typeHasVariablyModifiedType(x.Base)
	}
	return false
}

// 静态/文件作用域声明只拒绝对象自身的数组边界；指向 VM 类型的指针对象大小固定，
// 不能因为 pointee 变长就当作静态 VLA 数组。
func typeHasDisallowedStaticArrayBound(t Type) bool {
	switch x := unqual(t).(type) {
	case *ArrayType:
		if x.SizeKind == ArrayStarSize {
			return true
		}
		if x.SizeKind == ArrayVLA && !isNonRuntimeSizeofBound(x.SizeExpr) {
			return true
		}
		return typeHasDisallowedStaticArrayBound(x.Elem)
	case *PointerType:
		return false
	case *FunctionType:
		// 函数原型参数里的 [*]/VLA 不改变函数类型对象自身的大小。
		return typeHasDisallowedStaticArrayBound(x.Ret)
	case *QualType:
		return typeHasDisallowedStaticArrayBound(x.Base)
	}
	return false
}

// 文件作用域不能声明 variably modified type；这里要穿过指针，
// 与块作用域 static 指针对象的固定大小规则区分开。
func typeHasDisallowedFileScopeVMType(t Type) bool {
	switch x := unqual(t).(type) {
	case *ArrayType:
		if x.SizeKind == ArrayStarSize {
			return true
		}
		if x.SizeKind == ArrayVLA && !isNonRuntimeSizeofBound(x.SizeExpr) {
			return true
		}
		return typeHasDisallowedFileScopeVMType(x.Elem)
	case *PointerType:
		return typeHasDisallowedFileScopeVMType(x.Pointee)
	case *FunctionType:
		// 文件作用域的函数原型允许参数含 VM 形式；只有返回类型会让函数类型本身非法。
		return typeHasDisallowedFileScopeVMType(x.Ret)
	case *QualType:
		return typeHasDisallowedFileScopeVMType(x.Base)
	}
	return false
}

func markTypedefVMBounds(t Type) {
	switch x := unqual(t).(type) {
	case *ArrayType:
		if x.SizeKind == ArrayVLA {
			// typedef 声明处已经求值过 VM 边界；后续 cast 使用该类型时不要把
			// 保存下来的 m++ 等表达式当作初始化器中新出现的副作用。
			x.VMFromTypedef = true
		}
		markTypedefVMBounds(x.Elem)
	case *PointerType:
		markTypedefVMBounds(x.Pointee)
	case *FunctionType:
		markTypedefVMBounds(x.Ret)
		for _, p := range x.Params {
			markTypedefVMBounds(p)
		}
	case *QualType:
		markTypedefVMBounds(x.Base)
	}
}

func isNonRuntimeSizeofBound(expr Expr) bool {
	if expr == nil {
		return false
	}
	x, ok := expr.(*SizeofExpr)
	if !ok {
		return false
	}
	if x.Operand.Type != nil {
		return !typeHasVariableSize(x.Operand.Type)
	}
	if x.Operand.Expr != nil {
		return !typeHasVariableSize(x.Operand.Expr.GetType())
	}
	return false
}

// VM 指针 cast 可作为地址常量，但其中会被求值的 VLA 长度不能含有赋值、调用、
// 逗号或自增自减这类 C99 常量表达式禁止求值的操作；普通标识符和 compound literal
// 只表示运行期边界，不会破坏地址常量本身。
func typeHasForbiddenAddressConstantVMSize(t Type) bool {
	switch x := unqual(t).(type) {
	case *ArrayType:
		if x.SizeKind == ArrayVLA && !x.VMFromTypedef {
			if x.SizeExpr != nil && exprHasForbiddenAddressConstantVMSize(x.SizeExpr) {
				return true
			}
		}
		return typeHasForbiddenAddressConstantVMSize(x.Elem)
	case *PointerType:
		return typeHasForbiddenAddressConstantVMSize(x.Pointee)
	case *FunctionType:
		if typeHasForbiddenAddressConstantVMSize(x.Ret) {
			return true
		}
		for _, p := range x.Params {
			if typeHasForbiddenAddressConstantVMSize(p) {
				return true
			}
		}
	case *QualType:
		return typeHasForbiddenAddressConstantVMSize(x.Base)
	}
	return false
}

func exprHasForbiddenAddressConstantVMSize(expr Expr) bool {
	switch x := expr.(type) {
	case *AssignExpr, *CompoundAssign, *CallExpr, *CommaExpr:
		return true
	case *CompoundLit:
		return initListHasForbiddenAddressConstantVMSize(x.Init)
	case *UnOp:
		switch x.Op {
		case UnIncPre, UnIncPost, UnDecPre, UnDecPost:
			return true
		}
		return exprHasForbiddenAddressConstantVMSize(x.X)
	case *BinOp:
		return exprHasForbiddenAddressConstantVMSize(x.L) || exprHasForbiddenAddressConstantVMSize(x.R)
	case *CondExpr:
		return exprHasForbiddenAddressConstantVMSize(x.Cond) ||
			exprHasForbiddenAddressConstantVMSize(x.Then) ||
			exprHasForbiddenAddressConstantVMSize(x.Else)
	case *ImplicitCast:
		return exprHasForbiddenAddressConstantVMSize(x.X) || typeHasForbiddenAddressConstantVMSize(x.To)
	case *ExplicitCast:
		return exprHasForbiddenAddressConstantVMSize(x.X) || typeHasForbiddenAddressConstantVMSize(x.To)
	case *SizeofExpr:
		if x.Operand.Type != nil && typeHasVariableSize(x.Operand.Type) {
			return typeHasForbiddenAddressConstantVMSize(x.Operand.Type)
		}
		if x.Operand.Expr != nil && typeHasVariableSize(x.Operand.Expr.GetType()) {
			return typeHasForbiddenAddressConstantVMSize(x.Operand.Expr.GetType())
		}
		return false
	case *MemberExpr:
		return exprHasForbiddenAddressConstantVMSize(x.Base)
	case *IndexExpr:
		return exprHasForbiddenAddressConstantVMSize(x.Base) || exprHasForbiddenAddressConstantVMSize(x.Index)
	}
	return false
}

func initListHasForbiddenAddressConstantVMSize(il *InitList) bool {
	if il == nil {
		return false
	}
	for _, elem := range il.Elems {
		if exprHasForbiddenAddressConstantVMSize(elem.Value) {
			return true
		}
	}
	return false
}

// sizeofType 使用 cvm 当前 64 位目标模型。这里保持简单布局，后续后端布局阶段可以替换为
// 带对齐的 layout 计算。
func sizeofType(t Type) int64 {
	switch x := t.(type) {
	case *BuiltinType:
		switch x.Kind {
		case Void:
			return 1
		case Bool, Char, SChar, UChar:
			return 1
		case Short, UShort:
			return 2
		case Int, UInt, Float:
			return 4
		case Long, ULong, LongLong, ULongLong, Double:
			return 8
		case LongDouble:
			return 16
		case FloatComplex:
			return 8
		case DoubleComplex:
			return 16
		case LongDoubleComplex:
			return 32
		}
	case *PointerType, *FunctionType:
		return 8
	case *ArrayType:
		if x.SizeKind == ArrayConstantSize {
			return x.Size * sizeofType(x.Elem)
		}
	case *StructType:
		if !x.Complete {
			return 0
		}
		var end int64
		for _, f := range x.Fields {
			if n := f.Offset + sizeofType(f.T); n > end {
				end = n
			}
		}
		return end
	case *UnionType:
		var max int64
		for _, f := range x.Fields {
			if n := sizeofType(f.T); n > max {
				max = n
			}
		}
		return max
	case *EnumType:
		return sizeofType(x.Underlying)
	case *QualType:
		return sizeofType(x.Base)
	}
	return 0
}
