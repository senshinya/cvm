package sema

import "testing"

func TestBuiltinTypeKindString(t *testing.T) {
	bt := &BuiltinType{Kind: Int}
	if got := bt.String(); got != "int" {
		t.Fatalf("BuiltinType{Int}.String() = %q, want %q", got, "int")
	}
}
