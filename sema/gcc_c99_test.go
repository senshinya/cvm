package sema

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"shinya.click/cvm/lexer"
	"shinya.click/cvm/parser"
)

func TestGCCC99AcceptSuite(t *testing.T) {
	runGCCC99Suite(t, filepath.Join("testdata", "gcc-c99", "accept"), true)
}

func TestGCCC99RejectSuite(t *testing.T) {
	runGCCC99Suite(t, filepath.Join("testdata", "gcc-c99", "reject"), false)
}

func runGCCC99Suite(t *testing.T, root string, wantAccept bool) {
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatalf("GCC C99 suite is not installed at %s: %v", root, err)
	}
	if len(entries) == 0 {
		t.Fatalf("GCC C99 suite is empty: %s", root)
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".c") {
			continue
		}
		path := filepath.Join(root, entry.Name())
		t.Run(entry.Name(), func(t *testing.T) {
			srcBytes, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			src := stripGCCDirectives(string(srcBytes))
			tokens, err := lexer.NewLexer(src).ScanTokens()
			if err != nil {
				if !wantAccept {
					return
				}
				t.Fatalf("lexer rejected GCC C99 case %s: %v", path, err)
			}
			candidates, err := parser.NewParser(tokens).Parse()
			if err != nil {
				if !wantAccept {
					return
				}
				t.Fatalf("parser rejected GCC C99 case %s: %v", path, err)
			}
			survivors, _ := PreFilter(candidates)
			if len(survivors) == 0 {
				if !wantAccept {
					return
				}
				t.Fatalf("PreFilter rejected every AST candidate for %s", path)
			}
			var lastErrs []string
			for _, cand := range survivors {
				r := NewSema().analyzeOne(cand)
				if wantAccept && len(r.Errors) == 0 {
					return
				}
				if !wantAccept && len(r.Errors) > 0 {
					return
				}
				lastErrs = lastErrs[:0]
				for _, e := range r.Errors {
					lastErrs = append(lastErrs, e.Error())
				}
			}
			if wantAccept {
				t.Fatalf("Sema rejected every candidate for %s:\n%s", path, strings.Join(lastErrs, "\n"))
			}
			t.Fatalf("Sema accepted every candidate for GCC reject case %s", path)
		})
	}
}

func stripGCCDirectives(src string) string {
	src = stripCComments(src)
	var out []string
	for _, line := range strings.Split(src, "\n") {
		trim := strings.TrimSpace(line)
		if strings.HasPrefix(trim, "/*") && strings.Contains(trim, "{ dg-") {
			continue
		}
		if strings.HasPrefix(trim, "//") && strings.Contains(trim, "{ dg-") {
			continue
		}
		out = append(out, line)
	}
	return strings.Join(out, "\n")
}

func stripCComments(src string) string {
	var b strings.Builder
	for i := 0; i < len(src); {
		if i+1 < len(src) && src[i] == '/' && src[i+1] == '*' {
			for i < len(src) && !(i+1 < len(src) && src[i] == '*' && src[i+1] == '/') {
				if src[i] == '\n' {
					b.WriteByte('\n')
				}
				i++
			}
			if i+1 < len(src) {
				i += 2
			}
			continue
		}
		b.WriteByte(src[i])
		i++
	}
	return b.String()
}
