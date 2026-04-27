package sema

import (
	"testing"

	"shinya.click/cvm/entity"
)

func TestCastBasics(t *testing.T) {
	tt := NewTypeTable()
	intT := tt.Builtin(Int)
	s := NewSema()
	lvalue := &VarRef{Sym: &Symbol{Name: "x", Kind: SymVar, T: intT}, T: intT, Range: entity.SourceRange{}}
	out := s.castLValueToRValue(lvalue)
	if ic, ok := out.(*ImplicitCast); !ok || ic.Kind != LValueToRValue || ic.X != lvalue {
		t.Fatalf("expected LValueToRValue, got %T", out)
	}
	arr := &VarRef{Sym: &Symbol{Name: "a"}, T: tt.ArrayConstant(intT, 3)}
	out = s.castArrayDecay(arr)
	if ic, ok := out.(*ImplicitCast); !ok || ic.Kind != ArrayDecay {
		t.Fatalf("expected ArrayDecay, got %T", out)
	}
}

func TestCastIntegerPromotionAndUAC(t *testing.T) {
	tt := NewTypeTable()
	s := NewSema()
	charT := s.Types.Builtin(Char)
	out := s.castIntegerPromotion(&IntLit{Value: 1, T: charT})
	if ic, ok := out.(*ImplicitCast); !ok || ic.Kind != IntegerPromotion || ic.To != s.Types.Builtin(Int) {
		t.Fatalf("expected char -> int promotion, got %T", out)
	}
	l := &IntLit{Value: 1, T: s.Types.Builtin(Int)}
	r := &FloatLit{Value: 2, T: s.Types.Builtin(Double)}
	lOut, rOut, common := s.castUsualArithmetic(l, r)
	if common != s.Types.Builtin(Double) || lOut == l || rOut != r {
		t.Fatalf("bad UAC: %T %T %v", lOut, rOut, common)
	}
	_ = tt
}

func TestCastBoolAndNullPointer(t *testing.T) {
	s := NewSema()
	intT := s.Types.Builtin(Int)
	ptrT := s.Types.Pointer(intT)
	b := s.castBoolConversion(&IntLit{Value: 5, T: intT})
	if ic, ok := b.(*ImplicitCast); !ok || ic.Kind != BoolConversion {
		t.Fatalf("expected bool conversion, got %T", b)
	}
	n := s.castNullPointerConstant(&IntLit{Value: 0, T: intT}, ptrT)
	if ic, ok := n.(*ImplicitCast); !ok || ic.Kind != NullPointerConstant {
		t.Fatalf("expected null pointer conversion, got %T", n)
	}
}
