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
	common := usualArithmeticType(s, l.GetType(), r.GetType())
	if common == nil || l.GetType() == common && r.GetType() == common {
		return l, r, l.GetType()
	}
	if l.GetType() != common {
		l = &ImplicitCast{From: l.GetType(), To: common, X: l, Kind: UsualArithmetic, Range: l.Pos()}
	}
	if r.GetType() != common {
		r = &ImplicitCast{From: r.GetType(), To: common, X: r, Kind: UsualArithmetic, Range: r.Pos()}
	}
	return l, r, common
}

func usualArithmeticType(s *Sema, a, b Type) Type {
	ab, aok := unqualifiedBuiltin(a)
	bb, bok := unqualifiedBuiltin(b)
	if !aok || !bok {
		return a
	}
	for _, k := range []BuiltinKind{LongDoubleComplex, DoubleComplex, FloatComplex, LongDouble, Double, Float} {
		if ab.Kind == k || bb.Kind == k {
			return s.Types.Builtin(k)
		}
	}
	if ab.Kind == bb.Kind {
		return s.Types.Builtin(ab.Kind)
	}
	as, bs := isSignedIntegerKind(ab.Kind), isSignedIntegerKind(bb.Kind)
	ar, br := arithmeticRankBuiltin(ab), arithmeticRankBuiltin(bb)
	if as == bs {
		if ar >= br {
			return s.Types.Builtin(ab.Kind)
		}
		return s.Types.Builtin(bb.Kind)
	}
	signed, unsigned := ab.Kind, bb.Kind
	signedRank, unsignedRank := ar, br
	if !as {
		signed, unsigned = bb.Kind, ab.Kind
		signedRank, unsignedRank = br, ar
	}
	if unsignedRank >= signedRank {
		return s.Types.Builtin(unsigned)
	}
	if signedCanRepresentUnsigned(signed, unsigned) {
		return s.Types.Builtin(signed)
	}
	return s.Types.Builtin(unsignedVersion(signed))
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
	if pt, ok := unqual(e.GetType()).(*PointerType); ok {
		if isVoidPointer(pt) {
			return s.isVoidPointerZero(e)
		}
		return false
	}
	cv, ok := s.evalNullPointerIntegerZero(e)
	return ok && cv.Int == 0
}

func (s *Sema) isVoidPointerZero(e Expr) bool {
	pt, ok := unqual(e.GetType()).(*PointerType)
	if !ok {
		return false
	}
	if _, qualified := pt.Pointee.(*QualType); qualified {
		return false
	}
	if !isVoidPointer(pt) {
		return false
	}
	cast, ok := e.(*ExplicitCast)
	if !ok {
		if implicit, implicitOK := e.(*ImplicitCast); implicitOK {
			cast, ok = implicit.X.(*ExplicitCast)
		}
	}
	if !ok {
		return false
	}
	if exprContainsFloatLiteral(cast.X) {
		return false
	}
	cv, ok := s.evalNullPointerIntegerZero(cast.X)
	return ok && cv.Int == 0
}

func (s *Sema) evalNullPointerIntegerZero(e Expr) (ConstValue, bool) {
	switch x := e.(type) {
	case *ExplicitCast:
		if !isInteger(x.To) {
			return ConstValue{}, false
		}
		if f, ok := x.X.(*FloatLit); ok {
			return ConstValue{Kind: ConstInt, Int: int64(f.Value), T: x.To}, true
		}
		cv, ok := s.evalNullPointerIntegerZero(x.X)
		if !ok {
			return ConstValue{}, false
		}
		return ConstValue{Kind: ConstInt, Int: cv.Int, T: x.To}, true
	case *ImplicitCast:
		if !isInteger(x.To) {
			return ConstValue{}, false
		}
		return s.evalNullPointerIntegerZero(x.X)
	case *FloatLit:
		return ConstValue{}, false
	}
	return NewEvaluator(s).EvalC99IntegerConstantExpression(e)
}

func exprContainsFloatLiteral(e Expr) bool {
	switch x := e.(type) {
	case *FloatLit:
		return true
	case *ExplicitCast:
		return exprContainsFloatLiteral(x.X)
	case *ImplicitCast:
		return exprContainsFloatLiteral(x.X)
	case *UnOp:
		return exprContainsFloatLiteral(x.X)
	case *BinOp:
		return exprContainsFloatLiteral(x.L) || exprContainsFloatLiteral(x.R)
	case *CondExpr:
		return exprContainsFloatLiteral(x.Cond) || exprContainsFloatLiteral(x.Then) || exprContainsFloatLiteral(x.Else)
	case *CommaExpr:
		return exprContainsFloatLiteral(x.L) || exprContainsFloatLiteral(x.R)
	}
	return false
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
	if compatibleVMPointerPointee(from.Pointee, to.Pointee) {
		return true
	}
	if correspondingSignedIntegerTypes(from.Pointee, to.Pointee) {
		return true
	}
	return compatibleTypeIgnoringTopLevelQualifiers(from.Pointee, to.Pointee)
}

func correspondingSignedIntegerTypes(a, b Type) bool {
	ak, aok := unqualifiedBuiltin(a)
	bk, bok := unqualifiedBuiltin(b)
	if !aok || !bok || integerValueBits(ak.Kind) == 0 || integerValueBits(bk.Kind) == 0 {
		return false
	}
	return integerValueBits(ak.Kind) == integerValueBits(bk.Kind) && isSignedIntegerKind(ak.Kind) != isSignedIntegerKind(bk.Kind)
}

func compatibleVMPointerPointee(from, to Type) bool {
	fa, fok := unqual(from).(*ArrayType)
	ta, tok := unqual(to).(*ArrayType)
	if !fok || !tok {
		return false
	}
	// 指向 VLA 的指针赋值按运行期边界兼容处理；元素类型仍必须兼容。
	if fa.SizeKind != ArrayVLA && ta.SizeKind != ArrayVLA {
		return false
	}
	return compatibleTypeIgnoringTopLevelQualifiers(fa.Elem, ta.Elem)
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
	return arithmeticRankBuiltin(bt)
}

func arithmeticRankBuiltin(bt *BuiltinType) int {
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

func isSignedIntegerKind(k BuiltinKind) bool {
	switch k {
	case Bool, Char, SChar, Short, Int, Long, LongLong:
		return true
	}
	return false
}

func unsignedVersion(k BuiltinKind) BuiltinKind {
	switch k {
	case Char, SChar, UChar:
		return UChar
	case Short, UShort:
		return UShort
	case Int, UInt:
		return UInt
	case Long, ULong:
		return ULong
	case LongLong, ULongLong:
		return ULongLong
	}
	return k
}

func signedCanRepresentUnsigned(signed, unsigned BuiltinKind) bool {
	return integerValueBits(signed) > integerValueBits(unsigned)
}

func integerValueBits(k BuiltinKind) int {
	switch k {
	case Bool:
		return 1
	case Char, SChar, UChar:
		return 8
	case Short, UShort:
		return 16
	case Int, UInt:
		return 32
	case Long, ULong, LongLong, ULongLong:
		return 64
	}
	return 0
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

func (s *Sema) castAllowed(from, to Type) bool {
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
	if s.Options.GNUExtensions && unionCastAllowed(from, to) {
		return true
	}
	bt, ok := unqual(to).(*BuiltinType)
	return ok && bt.Kind == Void
}

func unionCastAllowed(from, to Type) bool {
	u, ok := unqual(to).(*UnionType)
	if !ok || !u.Complete {
		return false
	}
	for _, field := range u.Fields {
		if compatibleTypeIgnoringTopLevelQualifiers(from, field.T) {
			return true
		}
	}
	return false
}
