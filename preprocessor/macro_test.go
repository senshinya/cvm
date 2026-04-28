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

func TestEmptyVariadicMacroElidesComma(t *testing.T) {
	res, err := PreprocessSource("main.c", `
#define WRAP(fmt, ...) sink(fmt, __VA_ARGS__)
WRAP("x");
`, Options{})
	if err != nil {
		t.Fatalf("PreprocessSource failed: %v", err)
	}
	for i := 0; i+1 < len(res.Tokens); i++ {
		if res.Tokens[i].Typ == entity.COMMA && res.Tokens[i+1].Typ == entity.RIGHT_PARENTHESES {
			t.Fatalf("empty __VA_ARGS__ left a trailing comma: %#v", res.Tokens)
		}
	}
	if hasIdentifier(res.Tokens, "__VA_ARGS__") {
		t.Fatalf("__VA_ARGS__ was not substituted: %#v", res.Tokens)
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
