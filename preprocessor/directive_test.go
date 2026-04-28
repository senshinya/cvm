package preprocessor

import (
	"testing"

	"shinya.click/cvm/entity"
)

func TestConditionalInclusionSelectsActiveBranch(t *testing.T) {
	res, err := PreprocessSource("main.c", `
#if 0
int bad = ;
#elif 1
int good;
#else
int also_bad = ;
#endif
`, Options{})
	if err != nil {
		t.Fatalf("PreprocessSource failed: %v", err)
	}
	if !hasIdentifier(res.Tokens, "good") {
		t.Fatalf("expected active branch token good in %#v", res.Tokens)
	}
	if hasIdentifier(res.Tokens, "bad") || hasIdentifier(res.Tokens, "also_bad") {
		t.Fatalf("inactive branch tokens leaked: %#v", res.Tokens)
	}
}

func TestLineDirectiveChangesDisplayedLocation(t *testing.T) {
	res, err := PreprocessSource("main.c", "#line 77 \"virt.c\"\nint x;\n", Options{})
	if err != nil {
		t.Fatalf("PreprocessSource failed: %v", err)
	}
	for _, tok := range res.Tokens {
		if tok.Typ == entity.INT {
			got := res.Sources.DisplayLocation(tok.SourceStart)
			if got.File != "virt.c" || got.Line != 77 {
				t.Fatalf("display location = %#v, want virt.c:77", got)
			}
			return
		}
	}
	t.Fatalf("INT token not found")
}

func hasIdentifier(tokens []entity.Token, name string) bool {
	for _, tok := range tokens {
		if tok.Typ == entity.IDENTIFIER && tok.Lexeme == name {
			return true
		}
	}
	return false
}
