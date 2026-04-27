package sema

func compatibleType(a, b Type) bool {
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

func compatibleFunctionType(a, b *FunctionType) bool {
	if !compatibleType(a.Ret, b.Ret) || a.Variadic != b.Variadic || a.HasProto != b.HasProto {
		return false
	}
	if len(a.Params) != len(b.Params) {
		return false
	}
	for i := range a.Params {
		if !compatibleType(a.Params[i], b.Params[i]) {
			return false
		}
	}
	return true
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
