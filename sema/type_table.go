package sema

type TypeTable struct {
	builtins       [len(builtinNames)]*BuiltinType
	pointers       map[pointerKey]*PointerType
	arraysConstant map[arrayConstantKey]*ArrayType
	arraysUnsized  map[Type]*ArrayType
	arraysStar     map[Type]*ArrayType
}

func NewTypeTable() *TypeTable {
	tt := &TypeTable{
		pointers:       map[pointerKey]*PointerType{},
		arraysConstant: map[arrayConstantKey]*ArrayType{},
		arraysUnsized:  map[Type]*ArrayType{},
		arraysStar:     map[Type]*ArrayType{},
	}
	for k := Void; int(k) < len(builtinNames); k++ {
		tt.builtins[k] = &BuiltinType{Kind: k}
	}
	return tt
}

func (tt *TypeTable) Builtin(k BuiltinKind) *BuiltinType {
	return tt.builtins[k]
}

type pointerKey struct{ pointee Type }

func (tt *TypeTable) Pointer(pointee Type) *PointerType {
	key := pointerKey{pointee}
	if p, ok := tt.pointers[key]; ok {
		return p
	}
	p := &PointerType{Pointee: pointee}
	tt.pointers[key] = p
	return p
}

type arrayConstantKey struct {
	elem Type
	size int64
}

func (tt *TypeTable) ArrayConstant(elem Type, size int64) *ArrayType {
	key := arrayConstantKey{elem, size}
	if a, ok := tt.arraysConstant[key]; ok {
		return a
	}
	a := &ArrayType{Elem: elem, Size: size, SizeKind: ArrayConstantSize}
	tt.arraysConstant[key] = a
	return a
}

func (tt *TypeTable) ArrayUnsized(elem Type) *ArrayType {
	if a, ok := tt.arraysUnsized[elem]; ok {
		return a
	}
	a := &ArrayType{Elem: elem, SizeKind: ArrayUnsized}
	tt.arraysUnsized[elem] = a
	return a
}

func (tt *TypeTable) ArrayStar(elem Type) *ArrayType {
	if a, ok := tt.arraysStar[elem]; ok {
		return a
	}
	a := &ArrayType{Elem: elem, SizeKind: ArrayStarSize}
	tt.arraysStar[elem] = a
	return a
}

func (tt *TypeTable) ArrayVLA(elem Type, sizeExpr any) *ArrayType {
	// VLA 类型按 C99 的 variably-modified-type 语义不做驻留。
	return &ArrayType{Elem: elem, SizeExpr: sizeExpr, SizeKind: ArrayVLA}
}
