package sema

func compatibleType(a, b Type) bool {
	if !sameQualifiers(a, b) {
		return false
	}
	a = unqual(a)
	b = unqual(b)
	if a == b {
		return true
	}
	switch ax := a.(type) {
	case *BuiltinType:
		bx, ok := b.(*BuiltinType)
		return ok && ax.Kind == bx.Kind
	case *PointerType:
		bx, ok := b.(*PointerType)
		return ok && compatibleType(ax.Pointee, bx.Pointee)
	case *ArrayType:
		bx, ok := b.(*ArrayType)
		return ok && compatibleType(ax.Elem, bx.Elem) && compatibleArraySize(ax, bx)
	case *FunctionType:
		bx, ok := b.(*FunctionType)
		return ok && compatibleFunctionType(ax, bx)
	case *StructType:
		bx, ok := b.(*StructType)
		return ok && ax.Tag != nil && ax.Tag == bx.Tag
	case *UnionType:
		bx, ok := b.(*UnionType)
		return ok && ax.Tag != nil && ax.Tag == bx.Tag
	case *EnumType:
		bx, ok := b.(*EnumType)
		return ok && ax.Tag != nil && ax.Tag == bx.Tag
	}
	return false
}

func compatibleTypeIgnoringTopLevelQualifiers(a, b Type) bool {
	return compatibleType(unqual(a), unqual(b))
}

func compatibleArraySize(a, b *ArrayType) bool {
	if a.SizeKind == ArrayUnsized || b.SizeKind == ArrayUnsized {
		return true
	}
	if a.SizeKind != b.SizeKind {
		return false
	}
	if a.SizeKind == ArrayConstantSize {
		return a.Size == b.Size
	}
	return true
}

func isObjectType(t Type) bool {
	switch x := unqual(t).(type) {
	case *ErrorType:
		return true
	case *BuiltinType:
		return x.Kind != Void
	case *FunctionType:
		return false
	case *ArrayType:
		return x.SizeKind != ArrayUnsized && x.SizeKind != ArrayStarSize && isObjectType(x.Elem)
	case *StructType:
		return x.Complete
	case *UnionType:
		return x.Complete
	case *EnumType:
		return x.Complete
	default:
		return true
	}
}

func isRestrictPointerTarget(t Type) bool {
	switch unqual(t).(type) {
	case *ErrorType:
		return true
	case *FunctionType:
		return false
	default:
		return true
	}
}

func isPrototypeArrayObjectType(t Type) bool {
	switch x := unqual(t).(type) {
	case *ArrayType:
		return x.SizeKind != ArrayUnsized && isPrototypeArrayObjectType(x.Elem)
	default:
		return isObjectType(t)
	}
}

func isFlexibleArrayMember(t Type) bool {
	at, ok := unqual(t).(*ArrayType)
	return ok && at.SizeKind == ArrayUnsized
}

func typeContainsFlexibleArrayMember(t Type) bool {
	return typeContainsFlexibleArrayMemberSeen(t, map[Type]bool{})
}

func typeContainsFlexibleArrayMemberSeen(t Type, seen map[Type]bool) bool {
	t = unqual(t)
	if seen[t] {
		return false
	}
	seen[t] = true
	switch x := t.(type) {
	case *ArrayType:
		return x.SizeKind == ArrayUnsized || typeContainsFlexibleArrayMemberSeen(x.Elem, seen)
	case *StructType:
		for _, f := range x.Fields {
			if isFlexibleArrayMember(f.T) || typeContainsFlexibleArrayMemberSeen(f.T, seen) {
				return true
			}
		}
	case *UnionType:
		for _, f := range x.Fields {
			if isFlexibleArrayMember(f.T) || typeContainsFlexibleArrayMemberSeen(f.T, seen) {
				return true
			}
		}
	}
	return false
}

func compatibleFunctionType(a, b *FunctionType) bool {
	if !compatibleTypeIgnoringTopLevelQualifiers(a.Ret, b.Ret) {
		return false
	}
	if a.HasProto && b.HasProto {
		return compatiblePrototypeParams(a, b)
	}
	if !a.HasProto && !b.HasProto {
		return true
	}
	if a.HasProto {
		return compatiblePrototypeWithNoProto(a, b)
	}
	return compatiblePrototypeWithNoProto(b, a)
}

func compatiblePrototypeParams(a, b *FunctionType) bool {
	if a.Variadic != b.Variadic || len(a.Params) != len(b.Params) {
		return false
	}
	for i := range a.Params {
		if !compatibleTypeIgnoringTopLevelQualifiers(a.Params[i], b.Params[i]) {
			return false
		}
	}
	return true
}

func compatiblePrototypeWithNoProto(proto, noProto *FunctionType) bool {
	if proto.Variadic || len(noProto.Params) != 0 {
		return false
	}
	for _, param := range proto.Params {
		if !compatibleTypeIgnoringTopLevelQualifiers(param, defaultPromotedType(param)) {
			return false
		}
	}
	return true
}

func defaultPromotedType(t Type) Type {
	bt, ok := unqualifiedBuiltin(t)
	if !ok {
		return t
	}
	switch bt.Kind {
	case Bool, Char, SChar, UChar, Short, UShort:
		return &BuiltinType{Kind: Int}
	case Float:
		return &BuiltinType{Kind: Double}
	}
	return t
}

func losesQualifier(from, to Type) bool {
	fq, fok := from.(*QualType)
	tq, tok := to.(*QualType)
	if fok {
		if !tok {
			return fq.Const || fq.Volatile || fq.Restrict
		}
		if fq.Const && !tq.Const {
			return true
		}
		if fq.Volatile && !tq.Volatile {
			return true
		}
		if fq.Restrict && !tq.Restrict {
			return true
		}
	}
	from = unqual(from)
	to = unqual(to)
	switch fx := from.(type) {
	case *PointerType:
		tx, ok := to.(*PointerType)
		return ok && losesQualifier(fx.Pointee, tx.Pointee)
	case *ArrayType:
		tx, ok := to.(*ArrayType)
		return ok && losesQualifier(fx.Elem, tx.Elem)
	}
	return false
}

func completeIncompleteArrayMismatch(a, b Type) bool {
	a = unqual(a)
	b = unqual(b)
	ax, aok := a.(*ArrayType)
	bx, bok := b.(*ArrayType)
	if !aok || !bok {
		return false
	}
	if (ax.SizeKind == ArrayUnsized) != (bx.SizeKind == ArrayUnsized) {
		return true
	}
	return completeIncompleteArrayMismatch(ax.Elem, bx.Elem)
}

func sameQualifiers(a, b Type) bool {
	aq, aok := a.(*QualType)
	bq, bok := b.(*QualType)
	if !aok && !bok {
		return true
	}
	if aok != bok {
		return false
	}
	return aq.Const == bq.Const && aq.Volatile == bq.Volatile && aq.Restrict == bq.Restrict
}

func isComplexType(t Type) bool {
	bt, ok := unqualifiedBuiltin(t)
	if !ok {
		return false
	}
	switch bt.Kind {
	case FloatComplex, DoubleComplex, LongDoubleComplex:
		return true
	}
	return false
}
