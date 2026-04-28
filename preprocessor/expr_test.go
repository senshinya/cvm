package preprocessor

import "testing"

func TestIfExpressionDefinedAndIdentifiers(t *testing.T) {
	pp := newPreprocessor("main.c", "", Options{})
	pp.macros.DefineObject("A", []PPToken{{Kind: PPNumber, Lexeme: "4"}})
	got, err := pp.evalIfExpression([]PPToken{
		{Kind: PPIdentifier, Lexeme: "defined"},
		{Kind: PPPunctuator, Lexeme: "("},
		{Kind: PPIdentifier, Lexeme: "A"},
		{Kind: PPPunctuator, Lexeme: ")"},
		{Kind: PPPunctuator, Lexeme: "&&"},
		{Kind: PPIdentifier, Lexeme: "MISSING"},
		{Kind: PPPunctuator, Lexeme: "=="},
		{Kind: PPNumber, Lexeme: "0"},
	})
	if err != nil {
		t.Fatalf("evalIfExpression failed: %v", err)
	}
	if got == 0 {
		t.Fatalf("expression evaluated false, want true")
	}
}

func TestIfExpressionUnsignedBoundary(t *testing.T) {
	pp := newPreprocessor("main.c", "", Options{})
	got, err := pp.evalIfExpression([]PPToken{
		{Kind: PPNumber, Lexeme: "18446744073709551615UL"},
		{Kind: PPPunctuator, Lexeme: ">"},
		{Kind: PPNumber, Lexeme: "0"},
	})
	if err != nil {
		t.Fatalf("evalIfExpression failed: %v", err)
	}
	if got == 0 {
		t.Fatalf("unsigned boundary expression evaluated false, want true")
	}
}

func TestIfExpressionExpandsNestedObjectMacros(t *testing.T) {
	pp := newPreprocessor("main.c", "", Options{})
	pp.macros.DefineObject("A", []PPToken{{Kind: PPIdentifier, Lexeme: "B"}})
	pp.macros.DefineObject("B", []PPToken{{Kind: PPIdentifier, Lexeme: "C"}})
	pp.macros.DefineObject("C", []PPToken{{Kind: PPNumber, Lexeme: "7"}})
	got, err := pp.evalIfExpression([]PPToken{
		{Kind: PPIdentifier, Lexeme: "A"},
		{Kind: PPPunctuator, Lexeme: "=="},
		{Kind: PPNumber, Lexeme: "7"},
	})
	if err != nil {
		t.Fatalf("evalIfExpression failed: %v", err)
	}
	if got == 0 {
		t.Fatalf("nested object-like macro expression evaluated false, want true")
	}
}
