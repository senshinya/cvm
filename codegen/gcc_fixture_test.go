package codegen

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"shinya.click/cvm/bytecode"
	"shinya.click/cvm/entity"
	"shinya.click/cvm/parser"
	"shinya.click/cvm/preprocessor"
	"shinya.click/cvm/sema"
)

func TestGCCBytecodeCompileSuite(t *testing.T) {
	manifest := filepath.Join("testdata", "gcc-bytecode-compile.tsv")
	content, err := os.ReadFile(manifest)
	if err != nil {
		t.Fatalf("read GCC bytecode manifest: %v", err)
	}
	cases := parseGCCBytecodeManifest(t, string(content))
	const minGCCBytecodeCases = 143
	if len(cases) < minGCCBytecodeCases {
		t.Fatalf("GCC bytecode compile suite too small: got %d cases, want >= %d", len(cases), minGCCBytecodeCases)
	}
	for _, c := range cases {
		c := c
		t.Run(filepath.Base(c.path), func(t *testing.T) {
			sourcePath := filepath.Join("..", c.path)
			source, err := os.ReadFile(sourcePath)
			if err != nil {
				t.Fatal(err)
			}
			mod := compileGCCBytecodeFixture(t, sourcePath, string(source))
			if err := bytecode.ValidateModule(mod); err != nil {
				t.Fatalf("validate bytecode: %v\n%s", err, bytecode.PrintModule(mod))
			}
		})
	}
}

func TestGCCBytecodeManifestCoversImportedAcceptFixtures(t *testing.T) {
	content, err := os.ReadFile(filepath.Join("testdata", "gcc-bytecode-compile.tsv"))
	if err != nil {
		t.Fatalf("read GCC bytecode manifest: %v", err)
	}
	cases := parseGCCBytecodeManifest(t, string(content))
	covered := map[string]bool{}
	for _, c := range cases {
		covered[c.path] = true
	}
	roots := []string{
		filepath.Join("..", "sema", "testdata", "gcc-c99", "accept"),
		filepath.Join("..", "sema", "testdata", "gcc-c99-extra", "accept"),
		filepath.Join("..", "sema", "testdata", "gcc-c90-as-c99", "accept"),
	}
	var missing []string
	for _, root := range roots {
		matches, err := filepath.Glob(filepath.Join(root, "*.c"))
		if err != nil {
			t.Fatalf("glob %s: %v", root, err)
		}
		for _, match := range matches {
			manifestPath := filepath.ToSlash(strings.TrimPrefix(match, filepath.Clean("..")+string(filepath.Separator)))
			if !covered[manifestPath] {
				missing = append(missing, manifestPath)
			}
		}
	}
	sort.Strings(missing)
	if len(missing) != 0 {
		t.Fatalf("GCC bytecode manifest is missing %d imported accept fixtures:\n%s", len(missing), strings.Join(missing, "\n"))
	}
}

func TestGCCTgmathFloatSinUsesFloatExtern(t *testing.T) {
	sourcePath := filepath.Join("..", "sema", "testdata", "gcc-c99", "accept", "c99-tgmath-2.c")
	source, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatal(err)
	}
	mod := compileGCCBytecodeFixture(t, sourcePath, string(source))
	if !moduleHasExtern(mod, "__cvm_tgmath_sinf") {
		t.Fatalf("c99-tgmath-2.c did not reference float tgmath extern; globals:\n%s", bytecode.PrintModule(mod))
	}
}

func TestGCCTgmathComplexExpUsesComplexExtern(t *testing.T) {
	sourcePath := filepath.Join("..", "sema", "testdata", "gcc-c99", "accept", "c99-tgmath-3.c")
	source, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatal(err)
	}
	mod := compileGCCBytecodeFixture(t, sourcePath, string(source))
	if !moduleHasExtern(mod, "__cvm_tgmath_cexp") {
		t.Fatalf("c99-tgmath-3.c did not reference complex tgmath extern; globals:\n%s", bytecode.PrintModule(mod))
	}
}

func TestGCCTgmathComplexPowFloatUsesComplexFloatExtern(t *testing.T) {
	sourcePath := filepath.Join("..", "sema", "testdata", "gcc-c99", "accept", "c99-tgmath-4.c")
	source, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatal(err)
	}
	mod := compileGCCBytecodeFixture(t, sourcePath, string(source))
	if !moduleHasExtern(mod, "__cvm_tgmath_cpowf") {
		t.Fatalf("c99-tgmath-4.c did not reference complex float tgmath extern; globals:\n%s", bytecode.PrintModule(mod))
	}
}

func TestGCCTgmathNexttowardUsesFirstArgumentRank(t *testing.T) {
	source := `#include <tgmath.h>
float f(void) {
  float x = 1.0f;
  long double y = 2.0L;
  return nexttoward(x, y);
}
`
	mod := compileGCCBytecodeFixture(t, "tgmath-nexttoward-rank.c", source)
	if !moduleHasExtern(mod, "__cvm_tgmath_nexttowardf") {
		t.Fatalf("nexttoward(float, long double) did not reference float extern; globals:\n%s", bytecode.PrintModule(mod))
	}
}

func TestGCCTgmathScalbnUsesFirstArgumentRank(t *testing.T) {
	source := `#include <tgmath.h>
float f(int n) {
  float x = 2.0f;
  return scalbn(x, n);
}
`
	mod := compileGCCBytecodeFixture(t, "tgmath-scalbn-rank.c", source)
	if !moduleHasExtern(mod, "__cvm_tgmath_scalbnf") {
		t.Fatalf("scalbn(float, int) did not reference float extern; globals:\n%s", bytecode.PrintModule(mod))
	}
}

func TestGCCTgmathIlogbUsesFloatExtern(t *testing.T) {
	source := `#include <tgmath.h>
int f(void) {
  float x = 8.0f;
  return ilogb(x);
}
`
	mod := compileGCCBytecodeFixture(t, "tgmath-ilogb-rank.c", source)
	if !moduleHasExtern(mod, "__cvm_tgmath_ilogbf") {
		t.Fatalf("ilogb(float) did not reference float extern; globals:\n%s", bytecode.PrintModule(mod))
	}
}

func TestGCCTgmathLongRoundingUsesFloatExtern(t *testing.T) {
	source := `#include <tgmath.h>
long f(void) {
  float x = 3.0f;
  return lrint(x);
}
`
	mod := compileGCCBytecodeFixture(t, "tgmath-lrint-rank.c", source)
	if !moduleHasExtern(mod, "__cvm_tgmath_lrintf") {
		t.Fatalf("lrint(float) did not reference float extern; globals:\n%s", bytecode.PrintModule(mod))
	}
}

type gccBytecodeCase struct {
	path   string
	reason string
}

func moduleHasExtern(mod *bytecode.Module, name string) bool {
	for _, g := range mod.Globals {
		if g.Kind == bytecode.GlobalExtern && g.Extern.Name == name {
			return true
		}
	}
	return false
}

func parseGCCBytecodeManifest(t *testing.T, content string) []gccBytecodeCase {
	t.Helper()
	lines := strings.Split(content, "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "path\treason" {
		t.Fatalf("manifest header is missing or malformed: %q", firstManifestLine(content))
	}
	var cases []gccBytecodeCase
	for lineNo, line := range lines[1:] {
		if strings.TrimSpace(line) == "" {
			continue
		}
		fields := strings.Split(line, "\t")
		if len(fields) != 2 {
			t.Fatalf("manifest line %d malformed: %q", lineNo+2, line)
		}
		if !strings.HasPrefix(fields[0], filepath.ToSlash(filepath.Join("sema", "testdata", "gcc-"))) {
			t.Fatalf("manifest line %d path is outside GCC fixture roots: %s", lineNo+2, fields[0])
		}
		if !strings.HasSuffix(fields[0], ".c") {
			t.Fatalf("manifest line %d path is not a C source: %s", lineNo+2, fields[0])
		}
		if strings.TrimSpace(fields[1]) == "" {
			t.Fatalf("manifest line %d has empty reason", lineNo+2)
		}
		cases = append(cases, gccBytecodeCase{path: fields[0], reason: fields[1]})
	}
	return cases
}

func compileGCCBytecodeFixture(t *testing.T, path, source string) *bytecode.Module {
	t.Helper()
	src := stripGCCDirectives(source)
	pp, err := preprocessor.PreprocessSource(path, src, preprocessor.Options{})
	if err != nil {
		t.Fatalf("preprocess: %v", err)
	}
	if parserTokenCount(pp.Tokens) == 0 {
		return bytecode.NewModule()
	}
	candidates, err := parser.NewParser(pp.Tokens).Parse()
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	prog, err := sema.AnalyzeWithOptions(candidates, gccSemaOptions(source))
	if err != nil {
		t.Fatalf("sema: %v", err)
	}
	mod, err := Generate(prog)
	if err != nil {
		t.Fatalf("codegen: %v", err)
	}
	return mod
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

func firstManifestLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i]
	}
	return s
}
