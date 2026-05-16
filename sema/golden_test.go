package sema

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"shinya.click/cvm/lexer"
	"shinya.click/cvm/parser"
)

var updateGolden = flag.Bool("update", false, "regenerate sema testdata golden files")

func TestGoldenPass(t *testing.T) {
	matches, err := filepath.Glob(filepath.Join("testdata", "pass", "*.c"))
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) == 0 {
		t.Skip("no testdata/pass cases yet")
	}
	const minPassGoldenCases = 12
	if len(matches) < minPassGoldenCases {
		t.Fatalf("sema pass golden suite too small: got %d cases, want >= %d", len(matches), minPassGoldenCases)
	}
	for _, src := range matches {
		t.Run(filepath.Base(src), func(t *testing.T) {
			source, err := os.ReadFile(src)
			if err != nil {
				t.Fatal(err)
			}
			tokens, err := lexer.NewLexer(string(source)).ScanTokens()
			if err != nil {
				t.Fatalf("lex: %v", err)
			}
			candidates, err := parser.NewParser(tokens).Parse()
			if err != nil {
				t.Fatalf("parse: %v", err)
			}
			prog, err := Analyze(candidates)
			if err != nil {
				t.Fatalf("sema: %v", err)
			}
			if err := ValidateProgramInvariants(prog); err != nil {
				t.Fatalf("IR invariant: %v", err)
			}
			got := PrintProgram(prog)
			goldenPath := strings.TrimSuffix(src, ".c") + ".ir.golden"
			if *updateGolden {
				if err := os.WriteFile(goldenPath, []byte(got), 0644); err != nil {
					t.Fatal(err)
				}
				return
			}
			want, err := os.ReadFile(goldenPath)
			if err != nil {
				t.Fatalf("read golden: %v (run with -update to create)", err)
			}
			if got != string(want) {
				t.Errorf("golden mismatch for %s\n--- want ---\n%s\n--- got ---\n%s", src, want, got)
			}
		})
	}
}

func TestGoldenFail(t *testing.T) {
	matches, err := filepath.Glob(filepath.Join("testdata", "fail", "*.c"))
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) == 0 {
		t.Skip("no testdata/fail cases yet")
	}
	for _, src := range matches {
		t.Run(filepath.Base(src), func(t *testing.T) {
			source, err := os.ReadFile(src)
			if err != nil {
				t.Fatal(err)
			}
			goldenPath := strings.TrimSuffix(src, ".c") + ".err.golden"
			tokens, err := lexer.NewLexer(string(source)).ScanTokens()
			if err != nil {
				assertErrorMatches(t, err.Error(), goldenPath)
				return
			}
			candidates, err := parser.NewParser(tokens).Parse()
			if err != nil {
				assertErrorMatches(t, err.Error(), goldenPath)
				return
			}
			_, err = Analyze(candidates)
			if err == nil {
				t.Fatalf("expected error, got none")
			}
			assertErrorMatches(t, err.Error(), goldenPath)
		})
	}
}

func assertErrorMatches(t *testing.T, got, goldenPath string) {
	t.Helper()
	if *updateGolden {
		if err := os.WriteFile(goldenPath, []byte(got), 0644); err != nil {
			t.Fatal(err)
		}
		return
	}
	want, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("read golden: %v (run with -update to create)", err)
	}
	wantStr := strings.TrimSpace(string(want))
	if !strings.Contains(got, wantStr) {
		t.Errorf("error mismatch\n--- want substring ---\n%s\n--- got ---\n%s", wantStr, got)
	}
}
