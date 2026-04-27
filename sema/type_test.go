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
