package runtime

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"shinya.click/cvm/bytecode"
	"shinya.click/cvm/codegen"
	"shinya.click/cvm/entity"
	"shinya.click/cvm/parser"
	"shinya.click/cvm/preprocessor"
	"shinya.click/cvm/sema"
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
	paths := []string{
		"sema/testdata/gcc-c99/accept/../reject/foo.c",
		"sema/testdata/gcc-c99/accept/..\\reject/foo.c",
	}
	for _, path := range paths {
		content := "path\texit\tcategory\treason\n" +
			path + "\t0\tarithmetic\tescapes accept root\n"
		_, err := parseGCCExecManifestContent(content)
		if err == nil {
			t.Fatalf("expected traversal path to be rejected: %s", path)
		}
	}
}

func TestGCCExecutionManifestRejectsInvalidRows(t *testing.T) {
	const validRow = "sema/testdata/gcc-c99/accept/c99-main-1.c\t0\tarithmetic\treturns zero\n"
	tests := []struct {
		name    string
		content string
		wantErr string
	}{
		{
			name:    "bad header",
			content: "path\texit\treason\n" + validRow,
			wantErr: "manifest header is missing or malformed",
		},
		{
			name:    "malformed row",
			content: "path\texit\tcategory\treason\nsema/testdata/gcc-c99/accept/c99-main-1.c\t0\tarithmetic\n",
			wantErr: "manifest line 2 malformed",
		},
		{
			name:    "duplicate path",
			content: "path\texit\tcategory\treason\n" + validRow + validRow,
			wantErr: "duplicates path",
		},
		{
			name:    "absolute path",
			content: "path\texit\tcategory\treason\n/sema/testdata/gcc-c99/accept/c99-main-1.c\t0\tarithmetic\tabsolute path\n",
			wantErr: "outside supported GCC accept fixture roots",
		},
		{
			name:    "unsupported root",
			content: "path\texit\tcategory\treason\nsema/testdata/gcc-c99/reject/not-accept.c\t0\tarithmetic\twrong root\n",
			wantErr: "outside supported GCC accept fixture roots",
		},
		{
			name:    "non C source",
			content: "path\texit\tcategory\treason\nsema/testdata/gcc-c99/accept/not-c.txt\t0\tarithmetic\tnot source\n",
			wantErr: "outside supported GCC accept fixture roots",
		},
		{
			name:    "invalid exit",
			content: "path\texit\tcategory\treason\nsema/testdata/gcc-c99/accept/c99-main-1.c\tzero\tarithmetic\tbad exit\n",
			wantErr: "invalid exit code",
		},
		{
			name:    "empty category",
			content: "path\texit\tcategory\treason\nsema/testdata/gcc-c99/accept/c99-main-1.c\t0\t\tmissing category\n",
			wantErr: "empty category",
		},
		{
			name:    "empty reason",
			content: "path\texit\tcategory\treason\nsema/testdata/gcc-c99/accept/c99-main-1.c\t0\tarithmetic\t \n",
			wantErr: "empty reason",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseGCCExecManifestContent(tt.content)
			if err == nil {
				t.Fatalf("parseGCCExecManifestContent succeeded, want error containing %q", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("error = %q, want substring %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestHasGCCRunDirectiveAcceptsBlockCommentRun(t *testing.T) {
	source := "/* { dg-do run } */\nint main(void) { return 0; }\n"
	if !hasGCCRunDirective(source) {
		t.Fatalf("hasGCCRunDirective(%q) = false, want true", source)
	}
}

func TestHasGCCRunDirectiveAcceptsTargetQualifiedRun(t *testing.T) {
	source := "/* { dg-do run { target *-*-solaris2* } } */\nint main(void) { return 0; }\n"
	if !hasGCCRunDirective(source) {
		t.Fatalf("hasGCCRunDirective(%q) = false, want true", source)
	}
}

func TestHasGCCRunDirectiveAcceptsC99RuntimeRequirement(t *testing.T) {
	source := "/* { dg-require-effective-target c99_runtime } */\nint main(void) { return 0; }\n"
	if !hasGCCRunDirective(source) {
		t.Fatalf("hasGCCRunDirective(%q) = false, want true", source)
	}
}

func TestHasGCCRunDirectiveRejectsNonRunCases(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name:   "compile only",
			source: "/* { dg-do compile } */\nint main(void) { return 0; }\n",
		},
		{
			name:   "no directive",
			source: "int main(void) { return 0; }\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if hasGCCRunDirective(tt.source) {
				t.Fatalf("hasGCCRunDirective(%q) = true, want false", tt.source)
			}
		})
	}
}

func TestGCCExecutionFixtures(t *testing.T) {
	content, err := os.ReadFile(filepath.Join("testdata", "gcc-exec", "manifest.tsv"))
	if err != nil {
		t.Fatalf("read GCC execution manifest: %v", err)
	}
	cases := parseGCCExecManifest(t, string(content))
	const minGCCExecCases = 6
	if len(cases) < minGCCExecCases {
		t.Fatalf("GCC execution suite too small: got %d cases, want >= %d", len(cases), minGCCExecCases)
	}
	for _, c := range cases {
		c := c
		t.Run(filepath.Base(c.path), func(t *testing.T) {
			sourcePath := filepath.Join("..", c.path)
			source, err := os.ReadFile(sourcePath)
			if err != nil {
				t.Fatalf("read fixture %s: %v", c.path, err)
			}
			if !hasGCCRunDirective(string(source)) {
				t.Fatalf("%s is not a GCC runtime fixture: missing { dg-do run } directive", c.path)
			}
			st := runGCCExecFixture(t, sourcePath, string(source))
			if st.Code != c.exitCode {
				t.Fatalf("%s exit code = %d, want %d", c.path, st.Code, c.exitCode)
			}
		})
	}
}

func TestTgmathSinFloatExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <tgmath.h>

float foo(float x)
{
  return sin(x);
}

int main(void)
{
  return foo(0.0f) == 0.0f ? 0 : 1;
}
`
	st := runGCCExecFixture(t, "tgmath-sin-float.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestTgmathExpPowExecuteThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <tgmath.h>

int main(void)
{
  if (exp(0.0) != 1.0)
    return 1;
  if (pow(2.0f, 3.0f) != 8.0f)
    return 2;
  return 0;
}
`
	st := runGCCExecFixture(t, "tgmath-exp-pow.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestTgmathComplexExpExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <tgmath.h>

int main(void)
{
  complex double z = __builtin_complex(0.0, 0.0);
  complex double r = exp(z);
  return __builtin_cabs(r) == 1.0 ? 0 : 1;
}
`
	st := runGCCExecFixture(t, "tgmath-complex-exp.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestTgmathComplexFloatExpExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <tgmath.h>

int main(void)
{
  complex float z = __builtin_complex(0.0f, 0.0f);
  complex float r = exp(z);
  return __builtin_cabsf(r) == 1.0f ? 0 : 1;
}
`
	st := runGCCExecFixture(t, "tgmath-complex-float-exp.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
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

func runGCCExecFixture(t *testing.T, path, source string) ExitStatus {
	t.Helper()
	const gccExecStepLimit = 100000

	src := stripGCCDirectives(source)
	pp, err := preprocessor.PreprocessSource(path, src, preprocessor.Options{})
	if err != nil {
		t.Fatalf("%s preprocess: %v", path, err)
	}
	if parserTokenCount(pp.Tokens) == 0 {
		t.Fatalf("%s preprocess produced no parser tokens", path)
	}
	candidates, err := parser.NewParser(pp.Tokens).Parse()
	if err != nil {
		t.Fatalf("%s parse: %v", path, err)
	}
	prog, err := sema.AnalyzeWithOptions(candidates, gccSemaOptions(source))
	if err != nil {
		t.Fatalf("%s sema: %v", path, err)
	}
	mod, err := codegen.Generate(prog)
	if err != nil {
		t.Fatalf("%s codegen: %v", path, err)
	}
	var encoded bytes.Buffer
	if err := bytecode.EncodeModule(&encoded, mod); err != nil {
		t.Fatalf("%s EncodeModule: %v", path, err)
	}
	p, err := Load(bytes.NewReader(encoded.Bytes()), LoadOptions{})
	if err != nil {
		t.Fatalf("%s Load: %v", path, err)
	}
	st, err := Run(context.Background(), p, RunOptions{StepLimit: gccExecStepLimit})
	if err != nil {
		t.Fatalf("%s Run: %v", path, err)
	}
	return st
}

func gccSemaOptions(src string) sema.SemaOptions {
	return sema.SemaOptions{
		PedanticErrors:                  gccPedanticErrors(src),
		GNUExtensions:                   gccGNUExtensions(src),
		Permissive:                      gccPermissive(src),
		WErrorDeclarationAfterStatement: gccWErrorDeclarationAfterStatement(src),
	}
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
	if strings.Contains(src, "empty initializer braces") && strings.Contains(src, "dg-warning") {
		return true
	}
	return false
}

func gccPermissive(src string) bool {
	for _, line := range strings.Split(src, "\n") {
		if strings.Contains(line, "dg-options") && strings.Contains(line, "-fpermissive") {
			return true
		}
	}
	return false
}

func gccWErrorDeclarationAfterStatement(src string) bool {
	for _, line := range strings.Split(src, "\n") {
		if strings.Contains(line, "dg-options") && strings.Contains(line, "-Werror=declaration-after-statement") {
			return true
		}
	}
	return false
}

func stripGCCDirectives(src string) string {
	var b strings.Builder
	inBlockComment := false
	for _, line := range strings.SplitAfter(src, "\n") {
		body := strings.TrimSuffix(line, "\n")
		newline := ""
		if strings.HasSuffix(line, "\n") {
			newline = "\n"
		}
		if isDejaGNULine(body) {
			if inBlockComment && strings.Contains(body, "*/") {
				b.WriteString("*/")
			}
			b.WriteString(newline)
			inBlockComment = updateBlockCommentState(body, inBlockComment)
			continue
		}
		b.WriteString(line)
		inBlockComment = updateBlockCommentState(body, inBlockComment)
	}
	return b.String()
}

func updateBlockCommentState(line string, inBlock bool) bool {
	for i := 0; i < len(line); i++ {
		if inBlock {
			if i+1 < len(line) && line[i] == '*' && line[i+1] == '/' {
				inBlock = false
				i++
			}
			continue
		}
		if i+1 < len(line) && line[i] == '/' && line[i+1] == '*' {
			inBlock = true
			i++
		}
	}
	return inBlock
}

func isDejaGNULine(line string) bool {
	trim := strings.TrimSpace(line)
	if strings.HasPrefix(trim, "/*") && strings.HasSuffix(trim, "*/") && strings.Contains(trim, "{ dg-") {
		return true
	}
	return strings.HasPrefix(trim, "//") && strings.Contains(trim, "{ dg-")
}

func hasGCCRunDirective(source string) bool {
	return strings.Contains(source, "{ dg-do run") || strings.Contains(source, "{ dg-require-effective-target c99_runtime")
}

func isAllowedGCCExecPath(manifestPath string) bool {
	if strings.Contains(manifestPath, "\\") {
		return false
	}
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
