# GCC Runtime Execution Fixtures Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a curated GCC-derived runtime execution gate that runs every currently supported deterministic GCC accept fixture through binary bytecode loading and asserts deterministic exit codes.

**Architecture:** Runtime gets its own manifest and runner under `runtime/testdata/gcc-exec` because execution fixtures have stricter suitability rules than codegen compile fixtures. The runner duplicates the small GCC fixture parsing/stripping option helpers from the codegen test package, then runs each selected C file through preprocessor, parser, sema, codegen, `bytecode.EncodeModule`, `runtime.Load`, and `runtime.Run`.

**Tech Stack:** Go `testing`, stdlib `bytes/os/path/filepath/strconv/strings/context`, existing `bytecode`, `codegen`, `entity`, `parser`, `preprocessor`, and `sema` packages.

---

## File Structure

- `runtime/testdata/gcc-exec/manifest.tsv`
  - New curated runtime execution manifest.
  - Header: `path	exit	category	reason`.
  - Contains only enabled execution fixtures; no skipped rows.

- `runtime/gcc_exec_test.go`
  - Extends the existing scaffold test.
  - Owns manifest parsing, path validation, source loading, GCC directive stripping, GCC sema options, compile-to-runtime execution, and exit-code assertion.
  - Keeps helper functions local to `runtime` tests to avoid depending on `codegen` package test internals.

No production runtime code should change for this plan. If a selected fixture exposes a Phase 1 runtime bug, stop after capturing the failure and create a focused follow-up fix task with a regression test.

---

### Task 1: Add Runtime Execution Manifest Parser

**Files:**
- Create: `runtime/testdata/gcc-exec/manifest.tsv`
- Modify: `runtime/gcc_exec_test.go`

- [ ] **Step 1: Write the initial manifest**

Create `runtime/testdata/gcc-exec/manifest.tsv` with only the header:

```tsv
path	exit	category	reason
```

- [ ] **Step 2: Replace the scaffold-only test with manifest parsing tests**

Modify `runtime/gcc_exec_test.go` so the imports and manifest parser look like this:

```go
package runtime

import (
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

func parseGCCExecManifest(t *testing.T, content string) []gccExecCase {
	t.Helper()
	lines := strings.Split(content, "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "path\texit\tcategory\treason" {
		t.Fatalf("manifest header is missing or malformed: %q", firstManifestLine(content))
	}
	var cases []gccExecCase
	seen := map[string]bool{}
	for lineNo, line := range lines[1:] {
		if strings.TrimSpace(line) == "" {
			continue
		}
		fields := strings.Split(line, "\t")
		if len(fields) != 4 {
			t.Fatalf("manifest line %d malformed: %q", lineNo+2, line)
		}
		if seen[fields[0]] {
			t.Fatalf("manifest line %d duplicates path %s", lineNo+2, fields[0])
		}
		seen[fields[0]] = true
		if !isAllowedGCCExecPath(fields[0]) {
			t.Fatalf("manifest line %d path is outside supported GCC accept fixture roots: %s", lineNo+2, fields[0])
		}
		exitCode, err := strconv.Atoi(fields[1])
		if err != nil {
			t.Fatalf("manifest line %d has invalid exit code %q: %v", lineNo+2, fields[1], err)
		}
		if strings.TrimSpace(fields[2]) == "" {
			t.Fatalf("manifest line %d has empty category", lineNo+2)
		}
		if strings.TrimSpace(fields[3]) == "" {
			t.Fatalf("manifest line %d has empty reason", lineNo+2)
		}
		cases = append(cases, gccExecCase{path: fields[0], exitCode: exitCode, category: fields[2], reason: fields[3]})
	}
	return cases
}
```

The import list includes `strconv` because `parseGCCExecManifest` uses `strconv.Atoi`.

- [ ] **Step 3: Add supported path validation**

Append this helper to `runtime/gcc_exec_test.go`:

```go
func isAllowedGCCExecPath(path string) bool {
	if !strings.HasSuffix(path, ".c") {
		return false
	}
	roots := []string{
		"sema/testdata/gcc-c99/accept/",
		"sema/testdata/gcc-c99-extra/accept/",
		"sema/testdata/gcc-c90-as-c99/accept/",
	}
	for _, root := range roots {
		if strings.HasPrefix(path, root) {
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
```

- [ ] **Step 4: Run the parser test**

Run:

```bash
go test ./runtime -run TestGCCExecutionManifestParses -count=1
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add runtime/gcc_exec_test.go runtime/testdata/gcc-exec/manifest.tsv
git commit -m "test(runtime): parse GCC execution manifest"
```

---

### Task 2: Add Compile-To-Runtime GCC Execution Runner

**Files:**
- Modify: `runtime/gcc_exec_test.go`

- [ ] **Step 1: Add the runner test skeleton**

Add this test to `runtime/gcc_exec_test.go`:

```go
func TestGCCExecutionFixtures(t *testing.T) {
	content, err := os.ReadFile(filepath.Join("testdata", "gcc-exec", "manifest.tsv"))
	if err != nil {
		t.Fatalf("read GCC execution manifest: %v", err)
	}
	cases := parseGCCExecManifest(t, string(content))
	const minGCCExecCases = 4
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
			st := runGCCExecFixture(t, sourcePath, string(source))
			if st.Code != c.exitCode {
				t.Fatalf("%s exit code = %d, want %d", c.path, st.Code, c.exitCode)
			}
		})
	}
}
```

Update imports to include:

```go
import (
	"bytes"
	"context"
	"os"
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
```

- [ ] **Step 2: Add compile and run helper**

Append this helper to `runtime/gcc_exec_test.go`:

```go
func runGCCExecFixture(t *testing.T, path, source string) ExitStatus {
	t.Helper()
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
	st, err := Run(context.Background(), p, RunOptions{})
	if err != nil {
		t.Fatalf("%s Run: %v", path, err)
	}
	return st
}
```

- [ ] **Step 3: Add GCC option and directive helpers**

Append these helpers to `runtime/gcc_exec_test.go`:

```go
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
```

- [ ] **Step 4: Run the empty manifest gate and verify it fails on minimum count**

Run:

```bash
go test ./runtime -run TestGCCExecutionFixtures -count=1
```

Expected: FAIL with `GCC execution suite too small: got 0 cases, want >= 4`.

- [ ] **Step 5: Commit the runner**

```bash
git add runtime/gcc_exec_test.go
git commit -m "test(runtime): add GCC execution runner"
```

---

### Task 3: Select And Add Initial GCC Execution Fixtures

**Files:**
- Modify: `runtime/testdata/gcc-exec/manifest.tsv`

- [ ] **Step 1: Run candidate probe from GCC accept fixtures**

Run this temporary probe to list current GCC accept fixtures that compile, encode, load, and run:

```bash
tmp_probe=$(mktemp)
cat > "$tmp_probe" <<'GOEOF'
package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"shinya.click/cvm/bytecode"
	"shinya.click/cvm/codegen"
	"shinya.click/cvm/entity"
	"shinya.click/cvm/parser"
	"shinya.click/cvm/preprocessor"
	cvmruntime "shinya.click/cvm/runtime"
	"shinya.click/cvm/sema"
)

func main() {
	var paths []string
	roots := []string{
		"sema/testdata/gcc-c99/accept",
		"sema/testdata/gcc-c99-extra/accept",
		"sema/testdata/gcc-c90-as-c99/accept",
	}
	for _, root := range roots {
		filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
			if err == nil && !d.IsDir() && strings.HasSuffix(path, ".c") {
				paths = append(paths, filepath.ToSlash(path))
			}
			return nil
		})
	}
	sort.Strings(paths)
	for _, path := range paths {
		source, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		st, err := run(path, string(source))
		if err == nil {
			fmt.Printf("%s\t%d\n", path, st.Code)
		}
	}
}

func run(path, source string) (cvmruntime.ExitStatus, error) {
	src := stripGCCDirectives(source)
	pp, err := preprocessor.PreprocessSource(path, src, preprocessor.Options{})
	if err != nil {
		return cvmruntime.ExitStatus{}, err
	}
	if parserTokenCount(pp.Tokens) == 0 {
		return cvmruntime.ExitStatus{}, fmt.Errorf("no parser tokens")
	}
	candidates, err := parser.NewParser(pp.Tokens).Parse()
	if err != nil {
		return cvmruntime.ExitStatus{}, err
	}
	prog, err := sema.AnalyzeWithOptions(candidates, gccSemaOptions(source))
	if err != nil {
		return cvmruntime.ExitStatus{}, err
	}
	mod, err := codegen.Generate(prog)
	if err != nil {
		return cvmruntime.ExitStatus{}, err
	}
	var encoded bytes.Buffer
	if err := bytecode.EncodeModule(&encoded, mod); err != nil {
		return cvmruntime.ExitStatus{}, err
	}
	p, err := cvmruntime.Load(bytes.NewReader(encoded.Bytes()), cvmruntime.LoadOptions{})
	if err != nil {
		return cvmruntime.ExitStatus{}, err
	}
	return cvmruntime.Run(context.Background(), p, cvmruntime.RunOptions{})
}

func gccSemaOptions(src string) sema.SemaOptions {
	return sema.SemaOptions{
		PedanticErrors:                  containsOption(src, "-pedantic-errors"),
		GNUExtensions:                   containsOption(src, "-std=gnu") || strings.Contains(src, "empty initializer braces"),
		Permissive:                      containsOption(src, "-fpermissive"),
		WErrorDeclarationAfterStatement: containsOption(src, "-Werror=declaration-after-statement"),
	}
}

func containsOption(src, opt string) bool {
	for _, line := range strings.Split(src, "\n") {
		if strings.Contains(line, "dg-options") && strings.Contains(line, opt) {
			return true
		}
	}
	return false
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
GOEOF
go run "$tmp_probe" | sort
rm -f "$tmp_probe"
```

Expected output:

```text
sema/testdata/gcc-c99-extra/accept/inline-10.c	1
sema/testdata/gcc-c99-extra/accept/overflow-2.c	1
sema/testdata/gcc-c99-extra/accept/signbit-sa.c	0
sema/testdata/gcc-c99/accept/c99-main-1.c	0
```

These four rows are the complete set of GCC accept fixtures that the current Phase 1 runtime can compile, encode, load, and run with deterministic exit status.

- [ ] **Step 2: Inspect candidate source files before adding them**

Run:

```bash
sed -n '1,180p' sema/testdata/gcc-c99/accept/c99-main-1.c
sed -n '1,180p' sema/testdata/gcc-c99-extra/accept/signbit-sa.c
sed -n '1,180p' sema/testdata/gcc-c99-extra/accept/inline-10.c
sed -n '1,180p' sema/testdata/gcc-c99-extra/accept/overflow-2.c
```

Confirm the return values match the source:

- `c99-main-1.c`: C99 implicit `return 0` from `main`.
- `signbit-sa.c`: `signbit` calls on positive constants combine to `0`.
- `inline-10.c`: inline `main` returns literal `1`.
- `overflow-2.c`: current integer overflow expression takes the true branch and returns `1`.

- [ ] **Step 3: Fill the manifest with the first curated batch**

Edit `runtime/testdata/gcc-exec/manifest.tsv` so it contains exactly these currently runnable entries:

```tsv
path	exit	category	reason
sema/testdata/gcc-c99/accept/c99-main-1.c	0	arithmetic	trivial main return exercises binary load and entry return
sema/testdata/gcc-c99-extra/accept/signbit-sa.c	0	builtin	constant signbit calls on positive values return zero
sema/testdata/gcc-c99-extra/accept/inline-10.c	1	function	inline main returns a deterministic nonzero status
sema/testdata/gcc-c99-extra/accept/overflow-2.c	1	arithmetic	overflow conversion branch returns deterministic nonzero status
```

Every row uses a repository-relative path from an allowed `accept/` root.

- [ ] **Step 4: Run the execution fixture gate**

Run:

```bash
go test ./runtime -run TestGCCExecutionFixtures -count=1
```

Expected: PASS. If one of these four fixtures fails at runtime, stop and capture the failing path and stage; do not broaden runtime behavior in this task.

- [ ] **Step 5: Run runtime package tests**

Run:

```bash
go test ./runtime -count=1
```

Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add runtime/testdata/gcc-exec/manifest.tsv
git commit -m "test(runtime): add GCC execution fixtures"
```

---

### Task 4: Review, Verification, And Documentation Tightening

**Files:**
- Modify: `runtime/testdata/gcc-exec/README.md`
- Modify: `runtime/gcc_exec_test.go`

- [ ] **Step 1: Update README with manifest rules**

Replace `runtime/testdata/gcc-exec/README.md` with:

```markdown
# GCC Execution Fixtures

This directory contains GCC-derived execution fixtures that are deterministic
under the cvm bytecode runtime.

`manifest.tsv` is the execution gate. Each row contains:

- `path`: repository-relative path to a GCC accept fixture;
- `exit`: expected exit code from `main`;
- `category`: compact coverage label;
- `reason`: why the fixture is deterministic and useful for runtime coverage.

Rules:

- Do not copy the compile-only GCC fixture set wholesale.
- Add execution fixtures only when the runtime implements the required bytecode
  and extern behavior.
- Do not add skipped rows to this manifest. Unsupported execution cases should
  stay out of the gate until their required runtime behavior exists.
- Prefer small cases that cover integer execution, control flow, calls, globals,
  arrays, local objects, and pointer arithmetic.
- Avoid hosted library requirements, diagnostics-only cases, target harness
  behavior, undefined behavior, bit-fields, varargs, floating-point execution,
  inline assembly, atomics, and unsupported GCC builtins.
```

- [ ] **Step 2: Run focused tests**

Run:

```bash
go test ./runtime -run 'TestGCCExecution|TestCompileAndRun' -count=1
```

Expected: PASS.

- [ ] **Step 3: Run compile gate**

Run:

```bash
go test ./codegen -run TestGCCBytecodeCompileSuite -count=1
```

Expected: PASS.

- [ ] **Step 4: Run full verification**

Run:

```bash
go test ./runtime -count=1
go test ./... -count=1
git diff --check
```

Expected: every command exits 0. `git diff --check` should produce no output.

- [ ] **Step 5: Review final diff**

Run:

```bash
git diff --stat HEAD~3..HEAD
git diff -- runtime/gcc_exec_test.go runtime/testdata/gcc-exec/README.md runtime/testdata/gcc-exec/manifest.tsv
```

Expected:

- the runner reads `manifest.tsv`;
- all listed fixture paths are under allowed GCC accept roots;
- the manifest contains no skipped rows;
- execution goes through `bytecode.EncodeModule`, `Load`, and `Run`;
- no production runtime files are changed unless a separate reviewed runtime bug fix was required.

- [ ] **Step 6: Commit**

```bash
git add runtime/gcc_exec_test.go runtime/testdata/gcc-exec/README.md
git commit -m "docs(runtime): document GCC execution gate"
```

If only verification ran and README already matches the final wording, skip this commit and state that no documentation changes were needed.
