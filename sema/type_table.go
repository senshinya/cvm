package sema

type TypeTable struct {
	builtins [len(builtinNames)]*BuiltinType
	pointers map[pointerKey]*PointerType
}

func NewTypeTable() *TypeTable {
	tt := &TypeTable{
		pointers: map[pointerKey]*PointerType{},
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
