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

func TestGCCC99FixtureCoverage(t *testing.T) {
	manifest := filepath.Join("testdata", "gcc-c99", "manifest.tsv")
	content, err := os.ReadFile(manifest)
	if err != nil {
		t.Fatalf("read GCC C99 manifest: %v", err)
	}
	if !strings.HasPrefix(string(content), "path\tstatus\texpected\treason") {
		t.Fatalf("manifest header is missing or malformed: %q", firstLine(string(content)))
	}

	accept := countCFiles(t, filepath.Join("testdata", "gcc-c99", "accept"))
	reject := countCFiles(t, filepath.Join("testdata", "gcc-c99", "reject"))
	importedAccept := countManifestStatus(string(content), "imported-accept")
	importedReject := countManifestStatus(string(content), "imported-reject")
	if importedAccept != accept || importedReject != reject {
		t.Fatalf("manifest import counts do not match directories: manifest accept=%d reject=%d, dirs accept=%d reject=%d", importedAccept, importedReject, accept, reject)
	}
	const minImported = 40
	if accept+reject < minImported {
		t.Fatalf("GCC C99 suite too small: accept=%d reject=%d total=%d, want >= %d", accept, reject, accept+reject, minImported)
	}
}

func TestGCCC99OnlyPreprocessorSkipsRemain(t *testing.T) {
	manifest := filepath.Join("testdata", "gcc-c99", "manifest.tsv")
	content, err := os.ReadFile(manifest)
	if err != nil {
		t.Fatalf("read GCC C99 manifest: %v", err)
	}
	for lineNo, line := range strings.Split(string(content), "\n") {
		if strings.TrimSpace(line) == "" || lineNo == 0 {
			continue
		}
		fields := strings.Split(line, "\t")
		if len(fields) != 4 {
			t.Fatalf("manifest line %d malformed: %q", lineNo+1, line)
		}
		status, reason := fields[1], fields[3]
		if status != "skipped" {
			continue
		}
		if reason != "requires preprocessor or system macro handling" {
			t.Fatalf("non-preprocessor GCC C99 skip remains at line %d: %s: %s", lineNo+1, fields[0], reason)
		}
	}
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
			originalSrc := string(srcBytes)
			opts := SemaOptions{PedanticErrors: gccPedanticErrors(originalSrc)}
			src := stripGCCDirectives(originalSrc)
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
			if wantAccept {
				if _, err := AnalyzeWithOptions(candidates, opts); err != nil {
					t.Fatalf("lexer+parser+sema rejected GCC C99 case %s: %v", path, err)
				}
				return
			}
			if _, err := AnalyzeWithOptions(candidates, opts); err != nil {
				return
			}
			t.Fatalf("lexer+parser+sema accepted GCC reject case %s", path)
		})
	}
}

func gccPedanticErrors(src string) bool {
	return strings.Contains(src, "-pedantic-errors")
}

func countCFiles(t *testing.T, root string) int {
	t.Helper()
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".c") {
			count++
		}
	}
	return count
}

func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i]
	}
	return s
}

func countManifestStatus(content, status string) int {
	count := 0
	for _, line := range strings.Split(content, "\n") {
		fields := strings.Split(line, "\t")
		if len(fields) >= 2 && fields[1] == status {
			count++
		}
	}
	return count
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
