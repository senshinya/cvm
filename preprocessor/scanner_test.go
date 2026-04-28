package preprocessor

import "testing"

func TestScannerCommentsSplicesAndFlags(t *testing.T) {
	sm := NewSourceManager()
	fileID := sm.AddFile("main.c", "int/**/x \\\n= 1;\n#define A 2\n")
	toks, err := scanFile(sm, fileID, "int/**/x \\\n= 1;\n#define A 2\n", Options{})
	if err != nil {
		t.Fatalf("scanFile failed: %v", err)
	}
	want := []string{"int", "x", "=", "1", ";", "\n", "#", "define", "A", "2", "\n"}
	if got := ppLexemes(toks); !sameStrings(got, want) {
		t.Fatalf("lexemes = %#v, want %#v", got, want)
	}
	if !toks[0].StartOfLine {
		t.Fatalf("first token should start a line")
	}
	if !toks[1].LeadingSpace {
		t.Fatalf("comment replacement should leave leading space before x")
	}
	if !toks[2].LeadingSpace {
		t.Fatalf("line splice should preserve leading space before =")
	}
}

func TestScannerTrigraphWhenC99(t *testing.T) {
	sm := NewSourceManager()
	fileID := sm.AddFile("main.c", "??=define A 1\nA\n")
	toks, err := scanFile(sm, fileID, "??=define A 1\nA\n", Options{Std: StandardC99})
	if err != nil {
		t.Fatalf("scanFile failed: %v", err)
	}
	want := []string{"#", "define", "A", "1", "\n", "A", "\n"}
	if got := ppLexemes(toks); !sameStrings(got, want) {
		t.Fatalf("lexemes = %#v, want %#v", got, want)
	}
	if !toks[0].NeedsCleaning {
		t.Fatalf("trigraph-produced token should need cleaning")
	}
}

func ppLexemes(toks []PPToken) []string {
	out := make([]string, 0, len(toks))
	for _, tok := range toks {
		out = append(out, tok.Lexeme)
	}
	return out
}

func sameStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
