package runtime

import (
	"fmt"
	"path"
	"strconv"
	"strings"
	"testing"
)

type gccExecCase struct {
	path     string
	exitCode int
	category string
	reason   string
}

func TestGCCExecutionManifestParses(t *testing.T) {
	content := "path\texit\tcategory\treason\n" +
		"sema/testdata/gcc-c99/accept/c99-main-1.c\t0\tarithmetic\treturns zero\n"
	cases := parseGCCExecManifest(t, content)
	if len(cases) != 1 {
		t.Fatalf("case count = %d, want 1", len(cases))
	}
	if got := cases[0]; got.path != "sema/testdata/gcc-c99/accept/c99-main-1.c" || got.exitCode != 0 || got.category != "arithmetic" || got.reason != "returns zero" {
		t.Fatalf("case = %#v", got)
	}
}

func TestGCCExecutionManifestRejectsEscapingPath(t *testing.T) {
	content := "path\texit\tcategory\treason\n" +
		"sema/testdata/gcc-c99/accept/../reject/foo.c\t0\tarithmetic\tescapes accept root\n"
	_, err := parseGCCExecManifestContent(content)
	if err == nil {
		t.Fatal("expected traversal path to be rejected")
	}
}

func parseGCCExecManifest(t *testing.T, content string) []gccExecCase {
	t.Helper()
	cases, err := parseGCCExecManifestContent(content)
	if err != nil {
		t.Fatal(err)
	}
	return cases
}

func parseGCCExecManifestContent(content string) ([]gccExecCase, error) {
	lines := strings.Split(content, "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "path\texit\tcategory\treason" {
		return nil, fmt.Errorf("manifest header is missing or malformed: %q", firstManifestLine(content))
	}
	var cases []gccExecCase
	seen := map[string]bool{}
	for lineNo, line := range lines[1:] {
		if strings.TrimSpace(line) == "" {
			continue
		}
		fields := strings.Split(line, "\t")
		if len(fields) != 4 {
			return nil, fmt.Errorf("manifest line %d malformed: %q", lineNo+2, line)
		}
		if seen[fields[0]] {
			return nil, fmt.Errorf("manifest line %d duplicates path %s", lineNo+2, fields[0])
		}
		seen[fields[0]] = true
		if !isAllowedGCCExecPath(fields[0]) {
			return nil, fmt.Errorf("manifest line %d path is outside supported GCC accept fixture roots: %s", lineNo+2, fields[0])
		}
		exitCode, err := strconv.Atoi(fields[1])
		if err != nil {
			return nil, fmt.Errorf("manifest line %d has invalid exit code %q: %v", lineNo+2, fields[1], err)
		}
		if strings.TrimSpace(fields[2]) == "" {
			return nil, fmt.Errorf("manifest line %d has empty category", lineNo+2)
		}
		if strings.TrimSpace(fields[3]) == "" {
			return nil, fmt.Errorf("manifest line %d has empty reason", lineNo+2)
		}
		cases = append(cases, gccExecCase{path: fields[0], exitCode: exitCode, category: fields[2], reason: fields[3]})
	}
	return cases, nil
}

func isAllowedGCCExecPath(manifestPath string) bool {
	if path.IsAbs(manifestPath) {
		return false
	}
	if cleaned := path.Clean(manifestPath); cleaned != manifestPath {
		return false
	}
	if !strings.HasSuffix(manifestPath, ".c") {
		return false
	}
	roots := []string{
		"sema/testdata/gcc-c99/accept/",
		"sema/testdata/gcc-c99-extra/accept/",
		"sema/testdata/gcc-c90-as-c99/accept/",
	}
	for _, root := range roots {
		if strings.HasPrefix(manifestPath, root) {
			return true
		}
	}
	return false
}

func firstManifestLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i]
	}
	return s
}
