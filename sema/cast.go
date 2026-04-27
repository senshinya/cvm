package sema

import "shinya.click/cvm/entity"

func (s *Sema) castLValueToRValue(e Expr) Expr {
	if e == nil || e.GetCategory() == RValue {
		return e
	}
	t := unqual(e.GetType())
	return &ImplicitCast{From: e.GetType(), To: t, X: e, Kind: LValueToRValue, Range: e.Pos()}
}

func (s *Sema) castArrayDecay(e Expr) Expr {
	if e == nil {
		return e
	}
	at, ok := unqual(e.GetType()).(*ArrayType)
	if !ok {
		return e
	}
	return &ImplicitCast{From: e.GetType(), To: s.Types.Pointer(at.Elem), X: e, Kind: ArrayDecay, Range: e.Pos()}
}

func (s *Sema) castFunctionDecay(e Expr) Expr {
	if e == nil {
		return e
	}
	ft, ok := unqual(e.GetType()).(*FunctionType)
	if !ok {
		return e
	}
	return &ImplicitCast{From: e.GetType(), To: s.Types.Pointer(ft), X: e, Kind: FunctionDecay, Range: e.Pos()}
}

func (s *Sema) castIntegerPromotion(e Expr) Expr {
	bt, ok := unqualifiedBuiltin(e.GetType())
	if !ok {
		return e
	}
	switch bt.Kind {
	case Bool, Char, SChar, UChar, Short, UShort:
		return &ImplicitCast{From: e.GetType(), To: s.Types.Builtin(Int), X: e, Kind: IntegerPromotion, Range: e.Pos()}
	}
	return e
}

func (s *Sema) castUsualArithmetic(l, r Expr) (Expr, Expr, Type) {
	l = s.castIntegerPromotion(l)
	r = s.castIntegerPromotion(r)
	lr := arithmeticRank(l.GetType())
	rr := arithmeticRank(r.GetType())
	if lr == rr {
		return l, r, l.GetType()
	}
	if lr > rr {
		r = &ImplicitCast{From: r.GetType(), To: l.GetType(), X: r, Kind: UsualArithmetic, Range: r.Pos()}
		return l, r, l.GetType()
	}
	l = &ImplicitCast{From: l.GetType(), To: r.GetType(), X: l, Kind: UsualArithmetic, Range: l.Pos()}
	return l, r, r.GetType()
}

func (s *Sema) castBoolConversion(e Expr) Expr {
	boolT := s.Types.Builtin(Bool)
	if e.GetType() == boolT {
		return e
	}
	return &ImplicitCast{From: e.GetType(), To: boolT, X: e, Kind: BoolConversion, Range: e.Pos()}
}

func (s *Sema) castNullPointerConstant(e Expr, target Type) Expr {
	return &ImplicitCast{From: e.GetType(), To: target, X: e, Kind: NullPointerConstant, Range: e.Pos()}
}

func (s *Sema) castPointerConversion(e Expr, target Type) Expr {
	if e.GetType() == target {
		return e
	}
	return &ImplicitCast{From: e.GetType(), To: target, X: e, Kind: PointerConversion, Range: e.Pos()}
}

func (s *Sema) castVoidPointerConversion(e Expr, target Type) Expr {
	return &ImplicitCast{From: e.GetType(), To: target, X: e, Kind: VoidPointerConversion, Range: e.Pos()}
}

func (s *Sema) isNullPointerConstant(e Expr) bool {
	if isPointer(e.GetType()) {
		return false
	}
	cv, ok := NewEvaluator(s).EvalIntegerConstant(e)
	return ok && cv.Int == 0
}

func (s *Sema) assignmentConversion(e Expr, target Type, pos entity.SourcePos) Expr {
	from := e.GetType()
	if from == target || unqual(from) == unqual(target) {
		return e
	}
	if isPointer(target) && s.isNullPointerConstant(e) {
		return s.castNullPointerConstant(e, target)
	}
	if isArithmetic(from) && isArithmetic(target) {
		return s.arithmeticConversion(e, target)
	}
	if pf, ok := unqual(from).(*PointerType); ok {
		if pt, ok := unqual(target).(*PointerType); ok {
			if pointerAssignmentCompatible(pf, pt) {
				if isVoidPointer(pf) || isVoidPointer(pt) {
					return s.castVoidPointerConversion(e, target)
				}
				return s.castPointerConversion(e, target)
			}
			s.report(IncompatibleAssignment(pos, from.String(), target.String()))
			return e
		}
	}
	if bt, ok := unqual(target).(*BuiltinType); ok && bt.Kind == Bool {
		return s.castBoolConversion(e)
	}
	s.report(IncompatibleAssignment(pos, from.String(), target.String()))
	return e
}

func pointerAssignmentCompatible(from, to *PointerType) bool {
	if losesQualifier(from.Pointee, to.Pointee) {
		return false
	}
	fromFunc := isFunctionPointer(from)
	toFunc := isFunctionPointer(to)
	if fromFunc || toFunc {
		return fromFunc && toFunc && compatibleType(from.Pointee, to.Pointee)
	}
	if isVoidPointer(from) || isVoidPointer(to) {
		return true
	}
	return compatibleType(from.Pointee, to.Pointee)
}

func (s *Sema) arithmeticConversion(e Expr, target Type) Expr {
	from := e.GetType()
	if from == target {
		return e
	}
	src, srcOk := unqualifiedBuiltin(from)
	dst, dstOk := unqualifiedBuiltin(target)
	if !srcOk || !dstOk {
		return e
	}
	srcFloat := isFloating(src.Kind)
	dstFloat := isFloating(dst.Kind)
	switch {
	case srcFloat && dstFloat:
		return &ImplicitCast{From: from, To: target, X: e, Kind: FloatingConversion, Range: e.Pos()}
	case !srcFloat && !dstFloat:
		return &ImplicitCast{From: from, To: target, X: e, Kind: IntegralConversion, Range: e.Pos()}
	case !srcFloat && dstFloat:
		return &ImplicitCast{From: from, To: target, X: e, Kind: IntToFloat, Range: e.Pos()}
	case srcFloat && !dstFloat:
		return &ImplicitCast{From: from, To: target, X: e, Kind: FloatToInt, Range: e.Pos()}
	}
	return e
}

func unqual(t Type) Type {
	if q, ok := t.(*QualType); ok {
		return q.Base
	}
	return t
}

func unqualifiedBuiltin(t Type) (*BuiltinType, bool) {
	switch x := unqual(t).(type) {
	case *BuiltinType:
		return x, true
	case *EnumType:
		if x.Underlying != nil {
			return unqualifiedBuiltin(x.Underlying)
		}
	}
	return nil, false
}

func arithmeticRank(t Type) int {
	bt, ok := unqualifiedBuiltin(t)
	if !ok {
		return -1
	}
	switch bt.Kind {
	case LongDouble, LongDoubleComplex:
		return 100
	case Double, DoubleComplex:
		return 90
	case Float, FloatComplex:
		return 80
	case ULongLong:
		return 70
	case LongLong:
		return 65
	case ULong:
		return 60
	case Long:
		return 55
	case UInt:
		return 50
	case Int:
		return 45
	case UShort, Short:
		return 35
	case UChar, SChar, Char, Bool:
		return 30
	}
	return -1
}

func isFloating(k BuiltinKind) bool {
	switch k {
	case Float, Double, LongDouble, FloatComplex, DoubleComplex, LongDoubleComplex:
		return true
	}
	return false
}

func isArithmetic(t Type) bool {
	bt, ok := unqualifiedBuiltin(t)
	return ok && bt.Kind != Void
}

func isInteger(t Type) bool {
	bt, ok := unqualifiedBuiltin(t)
	return ok && !isFloating(bt.Kind) && bt.Kind != Void
}

func isPointer(t Type) bool {
	_, ok := unqual(t).(*PointerType)
	return ok
}

func isVoidPointer(p *PointerType) bool {
	bt, ok := unqual(p.Pointee).(*BuiltinType)
	return ok && bt.Kind == Void
}

func isFunctionPointer(p *PointerType) bool {
	_, ok := unqual(p.Pointee).(*FunctionType)
	return ok
}

func castAllowed(from, to Type) bool {
	if isArithmetic(from) && isArithmetic(to) {
		return true
	}
	if isPointer(from) && isPointer(to) {
		return true
	}
	if isPointer(from) && isInteger(to) {
		return true
	}
	if isInteger(from) && isPointer(to) {
		return true
	}
	bt, ok := unqual(to).(*BuiltinType)
	return ok && bt.Kind == Void
}
