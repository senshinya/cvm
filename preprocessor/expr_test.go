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
