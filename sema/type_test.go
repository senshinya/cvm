package sema

import "testing"

func TestBuiltinTypeKindString(t *testing.T) {
	bt := &BuiltinType{Kind: Int}
	if got := bt.String(); got != "int" {
		t.Fatalf("BuiltinType{Int}.String() = %q, want %q", got, "int")
	}
}

func TestBuiltinTypeAllC99Names(t *testing.T) {
	tests := map[BuiltinKind]string{
		Void:              "void",
		Bool:              "_Bool",
		Char:              "char",
		SChar:             "signed char",
		UChar:             "unsigned char",
		Short:             "short",
		UShort:            "unsigned short",
		Int:               "int",
		UInt:              "unsigned int",
		Long:              "long",
		ULong:             "unsigned long",
		LongLong:          "long long",
		ULongLong:         "unsigned long long",
		Float:             "float",
		Double:            "double",
		LongDouble:        "long double",
		FloatComplex:      "float _Complex",
		DoubleComplex:     "double _Complex",
		LongDoubleComplex: "long double _Complex",
	}
	tt := NewTypeTable()
	for kind, want := range tests {
		if got := tt.Builtin(kind).String(); got != want {
			t.Fatalf("Builtin(%v).String() = %q, want %q", kind, got, want)
		}
	}
}

func TestTypeTableBuiltinSingleton(t *testing.T) {
	tt := NewTypeTable()
	a := tt.Builtin(Int)
	b := tt.Builtin(Int)
	if a != b {
		t.Fatalf("Builtin(Int) returned distinct pointers; expected interning")
	}
	c := tt.Builtin(UInt)
	if a == c {
		t.Fatalf("Int and UInt returned same pointer")
	}
}

func TestTypeTableInstancesDoNotShareInternedTypes(t *testing.T) {
	tt1 := NewTypeTable()
	tt2 := NewTypeTable()
	if tt1.Builtin(Int) == tt2.Builtin(Int) {
		t.Fatalf("distinct TypeTable instances shared builtin singleton")
	}
	if tt1.Pointer(tt1.Builtin(Int)) == tt2.Pointer(tt2.Builtin(Int)) {
		t.Fatalf("distinct TypeTable instances shared pointer singleton")
	}
}

func TestPointerTypeInterning(t *testing.T) {
	tt := NewTypeTable()
	intT := tt.Builtin(Int)
	p1 := tt.Pointer(intT)
	p2 := tt.Pointer(intT)
	if p1 != p2 {
		t.Fatalf("Pointer(int) interning failed: %p vs %p", p1, p2)
	}
	pp := tt.Pointer(p1)
	if pp == p1 {
		t.Fatalf("Pointer(int*) collided with Pointer(int)")
	}
	if got := pp.String(); got != "int**" {
		t.Fatalf("Pointer(int*).String() = %q, want %q", got, "int**")
	}
}

func TestPointerTypeUsesQualifiedIdentity(t *testing.T) {
	tt := NewTypeTable()
	intT := tt.Builtin(Int)
	constInt := tt.Qualified(intT, true, false, false)
	if tt.Pointer(intT) == tt.Pointer(constInt) {
		t.Fatalf("pointer to int collided with pointer to const int")
	}
}

func TestArrayTypeConstantInterning(t *testing.T) {
	tt := NewTypeTable()
	intT := tt.Builtin(Int)
	a1 := tt.ArrayConstant(intT, 5)
	a2 := tt.ArrayConstant(intT, 5)
	if a1 != a2 {
		t.Fatalf("ArrayConstant(int, 5) interning failed")
	}
	a3 := tt.ArrayConstant(intT, 6)
	if a1 == a3 {
		t.Fatalf("Different sizes collided")
	}
}

func TestArrayTypeStarInterningAndDistinctness(t *testing.T) {
	tt := NewTypeTable()
	intT := tt.Builtin(Int)
	star1 := tt.ArrayStar(intT)
	star2 := tt.ArrayStar(intT)
	if star1 != star2 {
		t.Fatalf("ArrayStar(int) interning failed")
	}
	if star1 == tt.ArrayUnsized(intT) {
		t.Fatalf("ArrayStar(int) collided with ArrayUnsized(int)")
	}
	if got := star1.String(); got != "int[*]" {
		t.Fatalf("ArrayStar(int).String() = %q, want %q", got, "int[*]")
	}
}

func TestArrayTypeUnsizedInterning(t *testing.T) {
	tt := NewTypeTable()
	intT := tt.Builtin(Int)
	a1 := tt.ArrayUnsized(intT)
	a2 := tt.ArrayUnsized(intT)
	if a1 != a2 {
		t.Fatalf("ArrayUnsized(int) interning failed")
	}
}

func TestArrayTypeVLANotInterned(t *testing.T) {
	tt := NewTypeTable()
	intT := tt.Builtin(Int)
	a1 := tt.ArrayVLA(intT, nil)
	a2 := tt.ArrayVLA(intT, nil)
	if a1 == a2 {
		t.Fatalf("VLA arrays must NOT be interned")
	}
}

func TestFunctionTypeInterning(t *testing.T) {
	tt := NewTypeTable()
	intT := tt.Builtin(Int)
	doubleT := tt.Builtin(Double)
	f1 := tt.Function(intT, []Type{intT, doubleT}, false, true)
	f2 := tt.Function(intT, []Type{intT, doubleT}, false, true)
	if f1 != f2 {
		t.Fatalf("identical function types not interned")
	}
	f3 := tt.Function(intT, []Type{intT, doubleT}, true, true)
	if f1 == f3 {
		t.Fatalf("variadic flag did not differentiate")
	}
	f4 := tt.Function(intT, nil, false, false)
	f5 := tt.Function(intT, nil, false, true)
	if f4 == f5 {
		t.Fatalf("HasProto flag did not differentiate")
	}
}

func TestFunctionTypeInterningDifferentiatesRetAndParamOrder(t *testing.T) {
	tt := NewTypeTable()
	intT := tt.Builtin(Int)
	doubleT := tt.Builtin(Double)
	if tt.Function(intT, []Type{intT, doubleT}, false, true) == tt.Function(doubleT, []Type{intT, doubleT}, false, true) {
		t.Fatalf("return type did not differentiate function type")
	}
	if tt.Function(intT, []Type{intT, doubleT}, false, true) == tt.Function(intT, []Type{doubleT, intT}, false, true) {
		t.Fatalf("parameter order did not differentiate function type")
	}
}

func TestFunctionTypeCopiesParamSlice(t *testing.T) {
	tt := NewTypeTable()
	intT := tt.Builtin(Int)
	doubleT := tt.Builtin(Double)
	params := []Type{intT}
	fn := tt.Function(intT, params, false, true)
	params[0] = doubleT
	if fn.Params[0] != intT {
		t.Fatalf("Function stored caller-owned params slice")
	}
}

func TestQualTypeInterning(t *testing.T) {
	tt := NewTypeTable()
	intT := tt.Builtin(Int)
	c1 := tt.Qualified(intT, true, false, false)
	c2 := tt.Qualified(intT, true, false, false)
	if c1 != c2 {
		t.Fatalf("const int interning failed")
	}
	cv := tt.Qualified(intT, true, true, false)
	if c1 == cv {
		t.Fatalf("different qualifier sets collided")
	}
	if got := c1.String(); got != "const int" {
		t.Fatalf("String() = %q, want %q", got, "const int")
	}
}

func TestQualTypeAllC99QualifierCombinations(t *testing.T) {
	tt := NewTypeTable()
	intT := tt.Builtin(Int)
	tests := []struct {
		name                        string
		isConst, isVolatile, isRest bool
		want                        string
	}{
		{name: "const", isConst: true, want: "const int"},
		{name: "volatile", isVolatile: true, want: "volatile int"},
		{name: "restrict", isRest: true, want: "restrict int"},
		{name: "all", isConst: true, isVolatile: true, isRest: true, want: "const volatile restrict int"},
	}
	seen := map[*QualType]string{}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			q := tt.Qualified(intT, tc.isConst, tc.isVolatile, tc.isRest)
			if prev, ok := seen[q]; ok {
				t.Fatalf("%s collided with %s", tc.name, prev)
			}
			seen[q] = tc.name
			if got := q.String(); got != tc.want {
				t.Fatalf("String() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestStructTypeForwardCompletion(t *testing.T) {
	tt := NewTypeTable()
	tag := tt.NewTagID()
	st := tt.Struct(tag)
	if st.Complete {
		t.Fatalf("forward struct should be incomplete")
	}
	pst := tt.Pointer(st)
	intT := tt.Builtin(Int)
	tt.CompleteStruct(st, []*Field{{Name: "x", T: intT}})
	if !st.Complete {
		t.Fatalf("struct still incomplete after CompleteStruct")
	}
	if pst.Pointee != st {
		t.Fatalf("pointer's pointee no longer points to completed struct (lost identity)")
	}
	if len(st.Fields) != 1 || st.Fields[0].Name != "x" {
		t.Fatalf("fields not populated: %+v", st.Fields)
	}
}

func TestStructUnionEnumTagIdentityIsNominal(t *testing.T) {
	tt := NewTypeTable()
	tag1 := tt.NewTagID()
	tag2 := tt.NewTagID()
	if tag1 == tag2 {
		t.Fatalf("NewTagID returned identical pointers")
	}
	if tt.Struct(tag1) != tt.Struct(tag1) {
		t.Fatalf("Struct did not reuse the same TagID")
	}
	if tt.Struct(tag1) == tt.Struct(tag2) {
		t.Fatalf("different struct tags produced same type")
	}
	if tt.Union(tag1) != tt.Union(tag1) {
		t.Fatalf("Union did not reuse the same TagID")
	}
	if tt.Enum(tag1) != tt.Enum(tag1) {
		t.Fatalf("Enum did not reuse the same TagID")
	}
}

func TestUnionAndEnum(t *testing.T) {
	tt := NewTypeTable()

	uTag := tt.NewTagID()
	u := tt.Union(uTag)
	if u.Complete {
		t.Fatalf("forward union should be incomplete")
	}
	intT := tt.Builtin(Int)
	tt.CompleteUnion(u, []*Field{{Name: "i", T: intT}})
	if !u.Complete || len(u.Fields) != 1 {
		t.Fatalf("CompleteUnion failed: %+v", u)
	}

	eTag := tt.NewTagID()
	e := tt.Enum(eTag)
	tt.CompleteEnum(e, intT, []*Enumerator{{Name: "RED", Value: 0}})
	if !e.Complete || e.Underlying != intT || len(e.Enumerators) != 1 {
		t.Fatalf("CompleteEnum failed: %+v", e)
	}
}

func TestErrorTypeSingleton(t *testing.T) {
	if !IsError(ErrorTypeSingleton) {
		t.Fatalf("IsError(ErrorTypeSingleton) = false, want true")
	}
	tt := NewTypeTable()
	if IsError(tt.Builtin(Int)) {
		t.Fatalf("IsError(int) = true, want false")
	}
}
