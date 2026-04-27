package sema

type TypeTable struct {
	builtins [len(builtinNames)]*BuiltinType
}

func NewTypeTable() *TypeTable {
	tt := &TypeTable{}
	for k := Void; int(k) < len(builtinNames); k++ {
		tt.builtins[k] = &BuiltinType{Kind: k}
	}
	return tt
}

func (tt *TypeTable) Builtin(k BuiltinKind) *BuiltinType {
	return tt.builtins[k]
}
