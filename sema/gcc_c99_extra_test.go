package sema

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGCCC99ExtraExplicitStdSuite(t *testing.T) {
	t.Run("accept", func(t *testing.T) {
		runGCCC99Suite(t, filepath.Join("testdata", "gcc-c99-extra", "accept"), true)
	})
	t.Run("reject", func(t *testing.T) {
		runGCCC99Suite(t, filepath.Join("testdata", "gcc-c99-extra", "reject"), false)
	})
}

func TestGCCC99ExtraFixtureCoverage(t *testing.T) {
	manifest := filepath.Join("testdata", "gcc-c99-extra", "manifest.tsv")
	content, err := os.ReadFile(manifest)
	if err != nil {
		t.Fatalf("read GCC C99 extra manifest: %v", err)
	}
	if !strings.HasPrefix(string(content), "path\tstatus\texpected\treason") {
		t.Fatalf("manifest header is missing or malformed: %q", firstLine(string(content)))
	}
	accept := countCFiles(t, filepath.Join("testdata", "gcc-c99-extra", "accept"))
	reject := countCFiles(t, filepath.Join("testdata", "gcc-c99-extra", "reject"))
	importedAccept := countManifestStatus(string(content), "imported-accept")
	importedReject := countManifestStatus(string(content), "imported-reject")
	if importedAccept != accept || importedReject != reject {
		t.Fatalf("manifest import counts do not match directories: manifest accept=%d reject=%d, dirs accept=%d reject=%d", importedAccept, importedReject, accept, reject)
	}
	const minImported = 80
	if accept+reject < minImported {
		t.Fatalf("GCC C99 extra suite too small: accept=%d reject=%d total=%d, want >= %d", accept, reject, accept+reject, minImported)
	}
}

func TestGCCC99ExtraOnlyKnownSkipsRemain(t *testing.T) {
	manifest := filepath.Join("testdata", "gcc-c99-extra", "manifest.tsv")
	content, err := os.ReadFile(manifest)
	if err != nil {
		t.Fatalf("read GCC C99 extra manifest: %v", err)
	}
	allowed := map[string]bool{
		"requires preprocessor or system macro handling":              true,
		"requires GNU C extension support":                            true,
		"requires GCC target/runtime test harness":                    true,
		"requires GCC warning/diagnostic behavior":                    true,
		"requires C99 nested initializer/designator coverage":         true,
		"requires C99 variably modified type edge-case support":       true,
		"requires Unicode/UCN lexer support":                          true,
		"requires C99 incomplete return type diagnostics":             true,
		"requires C99 void lvalue/address diagnostics":                true,
		"requires C99 function prototype compatibility diagnostics":   true,
		"requires C99 inline definition diagnostics":                  true,
		"requires C99 incomplete-type pointer arithmetic diagnostics": true,
		"requires floating constant range diagnostics":                true,
	}
	for lineNo, line := range strings.Split(string(content), "\n") {
		if strings.TrimSpace(line) == "" || lineNo == 0 {
			continue
		}
		fields := strings.Split(line, "\t")
		if len(fields) != 4 {
			t.Fatalf("manifest line %d malformed: %q", lineNo+1, line)
		}
		if fields[1] != "skipped" {
			continue
		}
		if !allowed[fields[3]] {
			t.Fatalf("unknown GCC C99 extra skip at line %d: %s: %s", lineNo+1, fields[0], fields[3])
		}
	}
}
