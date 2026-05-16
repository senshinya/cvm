package sema

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGCCC90AsC99ExplicitStdSuite(t *testing.T) {
	t.Run("accept", func(t *testing.T) {
		runGCCC99Suite(t, filepath.Join("testdata", "gcc-c90-as-c99", "accept"), true)
	})
	t.Run("reject", func(t *testing.T) {
		runGCCC99Suite(t, filepath.Join("testdata", "gcc-c90-as-c99", "reject"), false)
	})
}

func TestGCCC90AsC99FixtureCoverage(t *testing.T) {
	manifest := filepath.Join("testdata", "gcc-c90-as-c99", "manifest.tsv")
	content, err := os.ReadFile(manifest)
	if err != nil {
		t.Fatalf("read GCC C90-as-C99 manifest: %v", err)
	}
	if !strings.HasPrefix(string(content), "path\tstatus\texpected\treason") {
		t.Fatalf("manifest header is missing or malformed: %q", firstLine(string(content)))
	}
	accept := countCFiles(t, filepath.Join("testdata", "gcc-c90-as-c99", "accept"))
	reject := countCFiles(t, filepath.Join("testdata", "gcc-c90-as-c99", "reject"))
	importedAccept := countManifestStatus(string(content), "imported-accept")
	importedReject := countManifestStatus(string(content), "imported-reject")
	if importedAccept != accept || importedReject != reject {
		t.Fatalf("manifest import counts do not match directories: manifest accept=%d reject=%d, dirs accept=%d reject=%d", importedAccept, importedReject, accept, reject)
	}
	const minImported = 101
	if accept+reject < minImported {
		t.Fatalf("GCC C90-as-C99 suite too small: accept=%d reject=%d total=%d, want >= %d", accept, reject, accept+reject, minImported)
	}
}

func TestGCCC90AsC99OnlyKnownSkipsRemain(t *testing.T) {
	manifest := filepath.Join("testdata", "gcc-c90-as-c99", "manifest.tsv")
	content, err := os.ReadFile(manifest)
	if err != nil {
		t.Fatalf("read GCC C90-as-C99 manifest: %v", err)
	}
	allowed := map[string]bool{
		"requires GNU C extension support":           true,
		"requires GCC target/runtime test harness":   true,
		"requires GCC warning/diagnostic behavior":   true,
		"requires GCC builtin support":               true,
		"requires GCC pragma semantics":              true,
		"requires GCC builtin namespace diagnostics": true,
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
			t.Fatalf("unknown GCC C90-as-C99 skip at line %d: %s: %s", lineNo+1, fields[0], fields[3])
		}
		if fields[3] == "requires preprocessor or system macro handling" {
			t.Fatalf("obsolete GCC C90-as-C99 preprocessor skip remains at line %d: %s", lineNo+1, fields[0])
		}
	}
}
