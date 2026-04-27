package sema

import "testing"

func TestBuiltinTypeKindString(t *testing.T) {
	bt := &BuiltinType{Kind: Int}
	if got := bt.String(); got != "int" {
		t.Fatalf("BuiltinType{Int}.String() = %q, want %q", got, "int")
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
