package sema

type TypeTable struct {
	builtins       [len(builtinNames)]*BuiltinType
	pointers       map[pointerKey]*PointerType
	arraysConstant map[arrayConstantKey]*ArrayType
	arraysUnsized  map[Type]*ArrayType
	arraysStar     map[Type]*ArrayType
	functions      map[functionKey]*FunctionType
}

func NewTypeTable() *TypeTable {
	tt := &TypeTable{
		pointers:       map[pointerKey]*PointerType{},
		arraysConstant: map[arrayConstantKey]*ArrayType{},
		arraysUnsized:  map[Type]*ArrayType{},
		arraysStar:     map[Type]*ArrayType{},
		functions:      map[functionKey]*FunctionType{},
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

type functionKey struct {
	ret      Type
	params   string
	variadic bool
	hasProto bool
}

func (tt *TypeTable) Function(ret Type, params []Type, variadic, hasProto bool) *FunctionType {
	key := functionKey{
		ret:      ret,
		params:   paramsKey(params),
		variadic: variadic,
		hasProto: hasProto,
	}
	if f, ok := tt.functions[key]; ok {
		return f
	}
	f := &FunctionType{
		Ret:      ret,
		Params:   append([]Type(nil), params...),
		Variadic: variadic,
		HasProto: hasProto,
	}
	tt.functions[key] = f
	return f
}

func paramsKey(params []Type) string {
	if len(params) == 0 {
		return ""
	}
	var b []byte
	for _, p := range params {
		ptr := uintptrOf(p)
		for i := 0; i < 8; i++ {
			b = append(b, byte(ptr>>(i*8)))
		}
		b = append(b, '|')
	}
	return string(b)
}
