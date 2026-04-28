package preprocessor

import (
	"testing"

	"shinya.click/cvm/entity"
)

func TestBluePaintRecursiveSuppression(t *testing.T) {
	res, err := PreprocessSource("main.c", `
#define A A
int A;
`, Options{})
	if err != nil {
		t.Fatalf("PreprocessSource failed: %v", err)
	}
	if !hasIdentifier(res.Tokens, "A") {
		t.Fatalf("self-referential macro should leave disabled A token: %#v", res.Tokens)
	}
}

func TestMacroArgumentPrescan(t *testing.T) {
	res, err := PreprocessSource("main.c", `
#define F(x) x + x
#define ONE 1
int y = F(ONE);
`, Options{})
	if err != nil {
		t.Fatalf("PreprocessSource failed: %v", err)
	}
	count := 0
	for _, tok := range res.Tokens {
		if tok.Lexeme == "1" {
			count++
		}
	}
	if count != 2 {
		t.Fatalf("expanded 1 count = %d, want 2; tokens=%#v", count, res.Tokens)
	}
}

func TestGCCClassicDisabledTokenCase(t *testing.T) {
	res, err := PreprocessSource("main.c", `
#define foo(x) bar x
foo(foo) (2)
`, Options{})
	if err != nil {
		t.Fatalf("PreprocessSource failed: %v", err)
	}
	want := []string{"bar", "foo", "(", "2", ")"}
	got := nonEOFParserLexemes(res.Tokens)
	if !sameStrings(got, want) {
		t.Fatalf("lexemes = %#v, want %#v", got, want)
	}
}

func TestEmptyMacroPreservesFunctionLikeSpacing(t *testing.T) {
	res, err := PreprocessSource("main.c", `
#define EMPTY
#define F() 1
int x = F EMPTY ();
`, Options{})
	if err != nil {
		t.Fatalf("PreprocessSource failed: %v", err)
	}
	count := 0
	for _, tok := range res.Tokens {
		if tok.Lexeme == "1" {
			count++
		}
	}
	if count != 0 {
		t.Fatalf("F EMPTY () must not become F(); tokens=%#v", res.Tokens)
	}
}

func TestFunctionLikeMacroLookaheadPushback(t *testing.T) {
	res, err := PreprocessSource("main.c", `
#define F() 1
int a = F;
int b = F + 2;
`, Options{})
	if err != nil {
		t.Fatalf("PreprocessSource failed: %v", err)
	}
	if !hasIdentifier(res.Tokens, "F") {
		t.Fatalf("non-invoked function-like macro name should remain: %#v", res.Tokens)
	}
	if !hasLexeme(res.Tokens, "2") {
		t.Fatalf("lookahead must push back + and following tokens: %#v", res.Tokens)
	}
}

func TestBuiltinFileAndLineMacros(t *testing.T) {
	res, err := PreprocessSource("main.c", "#line 40 \"virt.c\"\nchar *f = __FILE__;\nint l = __LINE__;\n", Options{})
	if err != nil {
		t.Fatalf("PreprocessSource failed: %v", err)
	}
	if !hasString(res.Tokens, "\"virt.c\"") {
		t.Fatalf("__FILE__ did not use presumed file: %#v", res.Tokens)
	}
	if !hasLexeme(res.Tokens, "41") {
		t.Fatalf("__LINE__ did not use presumed line: %#v", res.Tokens)
	}
}

func hasLexeme(tokens []entity.Token, lexeme string) bool {
	for _, tok := range tokens {
		if tok.Lexeme == lexeme {
			return true
		}
	}
	return false
}

func nonEOFParserLexemes(tokens []entity.Token) []string {
	var out []string
	for _, tok := range tokens {
		if tok.Typ != entity.EOF {
			out = append(out, tok.Lexeme)
		}
	}
	return out
}
