package preprocessor

import "testing"

func TestPreprocessSourceReturnsEOFTokenAndSourceManager(t *testing.T) {
	res, err := PreprocessSource("main.c", "int x;\n", Options{})
	if err != nil {
		t.Fatalf("PreprocessSource failed: %v", err)
	}
	if res.Sources == nil {
		t.Fatalf("Sources is nil")
	}
	if len(res.Tokens) == 0 {
		t.Fatalf("Tokens is empty")
	}
	if got := res.Tokens[len(res.Tokens)-1].Lexeme; got != "" {
		t.Fatalf("EOF lexeme = %q, want empty", got)
	}
}
