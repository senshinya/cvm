package sema

type ConstKind int

const (
	ConstInt ConstKind = iota
	ConstUint
	ConstFloat
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
	sema *Sema
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
		v, ok := evalBinOpInt(x.Op, l.Int, r.Int)
		if !ok {
			return ConstValue{}, false
		}
		return ConstValue{Kind: ConstInt, Int: v, Uint: uint64(v), T: x.T}, true
	case *UnOp:
		v, ok := e.EvalIntegerConstant(x.X)
		if !ok {
			return ConstValue{}, false
		}
		switch x.Op {
		case UnPlus:
			return v, true
		case UnMinus:
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
		return e.EvalIntegerConstant(x.X)
	case *ExplicitCast:
		if v, ok := e.EvalIntegerConstant(x.X); ok {
			return ConstValue{Kind: ConstInt, Int: v.Int, Uint: uint64(v.Int), T: x.To}, true
		}
		if f, ok := e.EvalArithmetic(x.X); ok && f.Kind == ConstFloat {
			return ConstValue{Kind: ConstInt, Int: int64(f.Float), Uint: uint64(int64(f.Float)), T: x.To}, true
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
		r, rok := e.evalC99IntegerConstantExpression(x.R)
		if !lok || !rok {
			return ConstValue{}, false
		}
		if x.Op == OpShl && l.Int < 0 {
			return ConstValue{}, false
		}
		v, ok := evalBinOpInt(x.Op, l.Int, r.Int)
		if !ok {
			return ConstValue{}, false
		}
		return ConstValue{Kind: ConstInt, Int: v, Uint: uint64(v), T: x.T}, true
	case *UnOp:
		v, ok := e.evalC99IntegerConstantExpression(x.X)
		if !ok {
			return ConstValue{}, false
		}
		switch x.Op {
		case UnPlus:
			return ConstValue{Kind: ConstInt, Int: v.Int, Uint: uint64(v.Int), T: x.T}, true
		case UnMinus:
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
			return e.evalC99IntegerConstantExpression(x.Then)
		}
		return e.evalC99IntegerConstantExpression(x.Else)
	case *ImplicitCast:
		cv, ok := e.evalC99IntegerConstantExpression(x.X)
		if !ok || !isInteger(x.To) {
			return ConstValue{}, false
		}
		return ConstValue{Kind: ConstInt, Int: cv.Int, Uint: uint64(cv.Int), T: x.To}, true
	case *ExplicitCast:
		if !isInteger(x.To) {
			return ConstValue{}, false
		}
		if cv, ok := e.evalC99IntegerConstantExpression(x.X); ok {
			return ConstValue{Kind: ConstInt, Int: cv.Int, Uint: uint64(cv.Int), T: x.To}, true
		}
		if cv, ok := e.evalC99CastArithmeticConstant(x.X); ok {
			if cv.Kind == ConstFloat {
				v := int64(cv.Float)
				return ConstValue{Kind: ConstInt, Int: v, Uint: uint64(v), T: x.To}, true
			}
			return ConstValue{Kind: ConstInt, Int: cv.Int, Uint: uint64(cv.Int), T: x.To}, true
		}
	}
	return ConstValue{}, false
}

// 整数 cast 的操作数可以是算术常量表达式，例如 (int)(double)0.0 或
// (int)-1.0；这里只展开 cast 和一元 +/-，避免把含运行时语义的表达式误判为 ICE。
func (e *Evaluator) evalC99CastArithmeticConstant(expr Expr) (ConstValue, bool) {
	switch x := expr.(type) {
	case *FloatLit:
		return ConstValue{Kind: ConstFloat, Float: x.Value, T: x.T}, true
	case *ExplicitCast:
		if !isArithmetic(x.To) {
			return ConstValue{}, false
		}
		cv, ok := e.evalC99CastArithmeticConstant(x.X)
		if !ok {
			return ConstValue{}, false
		}
		return castC99ArithmeticConstant(cv, x.To)
	case *ImplicitCast:
		if !isArithmetic(x.To) {
			return ConstValue{}, false
		}
		cv, ok := e.evalC99CastArithmeticConstant(x.X)
		if !ok {
			return ConstValue{}, false
		}
		return castC99ArithmeticConstant(cv, x.To)
	case *UnOp:
		cv, ok := e.evalC99CastArithmeticConstant(x.X)
		if !ok {
			return ConstValue{}, false
		}
		switch x.Op {
		case UnPlus:
			return cv, true
		case UnMinus:
			if cv.Kind == ConstFloat {
				return ConstValue{Kind: ConstFloat, Float: -cv.Float, T: x.T}, true
			}
			return ConstValue{Kind: ConstInt, Int: -cv.Int, Uint: uint64(-cv.Int), T: x.T}, true
		}
	default:
		return e.evalC99IntegerConstantExpression(expr)
	}
	return ConstValue{}, false
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

func (e *Evaluator) EvalConstant(expr Expr) (ConstValue, bool) {
	if cv, ok := e.EvalIntegerConstant(expr); ok {
		return cv, true
	}
	switch x := expr.(type) {
	case *FloatLit:
		return ConstValue{Kind: ConstFloat, Float: x.Value, T: x.T}, true
	case *StringLit:
		return ConstValue{Kind: ConstString, T: x.T}, true
	case *UnOp:
		if x.Op != UnAddr {
			return ConstValue{}, false
		}
		if vr, ok := x.X.(*VarRef); ok && vr.Sym.Storage != StorageAuto && vr.Sym.Storage != StorageRegister {
			return ConstValue{Kind: ConstAddress, Addr: ConstValueAddr{Sym: vr.Sym}, T: x.T}, true
		}
		if _, ok := x.X.(*CompoundLit); ok {
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
		scale := sizeofType(ptr.Pointee)
		switch {
		case l.Kind == ConstAddress && r.Kind == ConstInt:
			return ConstValue{Kind: ConstAddress, Addr: ConstValueAddr{Sym: l.Addr.Sym, Offset: l.Addr.Offset + signedOp(x.Op)*r.Int*scale}, T: x.T}, true
		case x.Op == OpAdd && l.Kind == ConstInt && r.Kind == ConstAddress:
			return ConstValue{Kind: ConstAddress, Addr: ConstValueAddr{Sym: r.Addr.Sym, Offset: r.Addr.Offset + l.Int*scale}, T: x.T}, true
		}
	case *ImplicitCast:
		if x.Kind == ArrayDecay {
			inner := x.X
			if ic, ok := inner.(*ImplicitCast); ok && ic.Kind == LValueToRValue {
				inner = ic.X
			}
			if _, ok := inner.(*CompoundLit); ok {
				return ConstValue{Kind: ConstAddress, T: x.To}, true
			}
		}
		return e.EvalConstant(x.X)
	case *ExplicitCast:
		return e.EvalConstant(x.X)
	case *AddrConst:
		return ConstValue{Kind: ConstAddress, Addr: ConstValueAddr{Sym: x.Sym, Offset: x.Offset}, T: x.T}, true
	}
	return ConstValue{}, false
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
		return typeHasVariableSize(x.Pointee)
	case *QualType:
		return typeHasVariableSize(x.Base)
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
