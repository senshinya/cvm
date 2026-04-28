package sema

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"shinya.click/cvm/entity"
	"shinya.click/cvm/parser"
	"shinya.click/cvm/preprocessor"
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
		allowed := map[string]bool{
			"requires real system header":               true,
			"requires GCC pragma semantics":             true,
			"requires GNU preprocessor extension":       true,
			"requires target-specific runtime behavior": true,
		}
		if !allowed[reason] {
			t.Fatalf("unknown GCC C99 skip remains at line %d: %s: %s", lineNo+1, fields[0], reason)
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
			opts := SemaOptions{PedanticErrors: gccPedanticErrors(originalSrc), GNUExtensions: gccGNUExtensions(originalSrc)}
			src := stripGCCDirectives(originalSrc)
			if wantAccept {
				if gccDoPreprocessOnly(originalSrc) {
					if _, err := preprocessor.PreprocessSource(path, src, preprocessor.Options{}); err != nil {
						t.Fatalf("preprocessor rejected GCC C99 preprocess case %s: %v", path, err)
					}
					return
				}
				if err := preprocessParseAnalyze(t, path, src, opts); err != nil {
					t.Fatalf("preprocessor+parser+sema rejected GCC C99 case %s: %v", path, err)
				}
				return
			}
			if err := preprocessParseAnalyze(t, path, src, opts); err != nil {
				return
			}
			t.Fatalf("preprocessor+parser+sema accepted GCC reject case %s", path)
		})
	}
}

func preprocessParseAnalyze(t *testing.T, path string, src string, opts SemaOptions) error {
	t.Helper()
	pp, err := preprocessor.PreprocessSource(path, src, preprocessor.Options{})
	if err != nil {
		return err
	}
	if parserTokenCount(pp.Tokens) == 0 {
		return nil
	}
	candidates, err := parser.NewParser(pp.Tokens).Parse()
	if err != nil {
		return err
	}
	_, err = AnalyzeWithOptions(candidates, opts)
	return err
}

func parserTokenCount(tokens []entity.Token) int {
	count := 0
	for _, tok := range tokens {
		if tok.Typ != entity.EOF {
			count++
		}
	}
	return count
}

func gccDoPreprocessOnly(src string) bool {
	for _, line := range strings.Split(src, "\n") {
		if strings.Contains(line, "dg-do") && strings.Contains(line, "preprocess") {
			return true
		}
	}
	return false
}

func gccPedanticErrors(src string) bool {
	for _, line := range strings.Split(src, "\n") {
		if strings.Contains(line, "dg-options") && strings.Contains(line, "-pedantic-errors") {
			return true
		}
	}
	return false
}

func gccGNUExtensions(src string) bool {
	for _, line := range strings.Split(src, "\n") {
		if strings.Contains(line, "dg-options") && strings.Contains(line, "-std=gnu") {
			return true
		}
	}
	return false
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
	var b strings.Builder
	for _, line := range strings.SplitAfter(src, "\n") {
		body := strings.TrimSuffix(line, "\n")
		newline := ""
		if strings.HasSuffix(line, "\n") {
			newline = "\n"
		}
		if isDejaGNULine(body) {
			b.WriteString(newline)
			continue
		}
		b.WriteString(line)
	}
	return b.String()
}

func isDejaGNULine(line string) bool {
	trim := strings.TrimSpace(line)
	if strings.HasPrefix(trim, "/*") && strings.HasSuffix(trim, "*/") && strings.Contains(trim, "{ dg-") {
		return true
	}
	return strings.HasPrefix(trim, "//") && strings.Contains(trim, "{ dg-")
}
