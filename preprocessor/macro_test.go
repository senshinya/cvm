package preprocessor

import (
	"testing"

	"shinya.click/cvm/entity"
)

func TestObjectAndFunctionLikeMacros(t *testing.T) {
	res, err := PreprocessSource("main.c", `
#define T int
#define ADD(a, b) ((a) + (b))
T x = ADD(1, 2);
`, Options{})
	if err != nil {
		t.Fatalf("PreprocessSource failed: %v", err)
	}
	want := []entity.TokenType{entity.INT, entity.IDENTIFIER, entity.EQUAL, entity.LEFT_PARENTHESES, entity.LEFT_PARENTHESES, entity.INTEGER_CONSTANT}
	for i, typ := range want {
		if res.Tokens[i].Typ != typ {
			t.Fatalf("token[%d] = %s, want %s; tokens=%#v", i, res.Tokens[i].Typ, typ, res.Tokens)
		}
	}
}

func TestStringifyPasteAndVariadicMacros(t *testing.T) {
	res, err := PreprocessSource("main.c", `
#define STR(x) #x
#define CAT(a, b) a ## b
#define FIRST(fmt, ...) fmt
char *s = STR(hello world);
int CAT(ab, c) = 3;
char *f = FIRST("x", 1, 2);
`, Options{})
	if err != nil {
		t.Fatalf("PreprocessSource failed: %v", err)
	}
	if !hasIdentifier(res.Tokens, "abc") {
		t.Fatalf("token paste did not produce abc: %#v", res.Tokens)
	}
	if !hasString(res.Tokens, "\"hello world\"") {
		t.Fatalf("stringification did not produce expected string token: %#v", res.Tokens)
	}
}

func hasString(tokens []entity.Token, lexeme string) bool {
	for _, tok := range tokens {
		if tok.Typ == entity.STRING && tok.Lexeme == lexeme {
			return true
		}
	}
	return false
}
