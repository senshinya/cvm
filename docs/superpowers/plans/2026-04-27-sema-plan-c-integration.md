# Sema Plan C — Integration Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the temporary `compiler.go` glue with a full forest aggregator: PreFilter, then concurrent per-survivor Sema, then winner selection. Migrate the parser end-to-end tests to `sema/`. Add `testdata/pass` and `testdata/fail` golden infrastructure. Concurrency stress test with `-race` clean.

**Architecture:** `sema.Analyze(candidates []*entity.AstNode) (*Program, error)` becomes the single public entry. Each survivor runs in its own goroutine with its own `Sema` instance (fresh `TypeTable` / `SymbolTable` — no shared state). Aggregator picks one error-free result, reports ambiguity if more than one, falls back to best-error otherwise.

**Tech Stack:** Go stdlib `sync` + existing sema package from Plans A and B.

**Spec reference:** `docs/superpowers/specs/2026-04-27-sema-layer-design.md` § Sema Flow (Forest aggregation), § Testing.

**Prerequisites:** Plans A and B complete.

---

## File Structure

**Create:**
- `sema/aggregator.go` — `Analyze` (public forest entry), `pickBestErrorResult`, `AmbiguousParse` error
- `sema/aggregator_test.go` — concurrency + selection tests
- `sema/testdata/pass/<name>.c` + `<name>.ir.golden` — pass golden snapshots
- `sema/testdata/fail/<name>.c` + `<name>.err.golden` — fail golden snapshots
- `sema/printer.go` — IR pretty-printer for golden snapshots
- `sema/printer_test.go`
- `sema/golden_test.go` — discovers testdata/ and runs golden checks

**Modify:**
- `compiler.go` — call `sema.Analyze(candidates)` instead of the Plan A `PreFilter` glue
- `sema/sema.go` — make `Analyze` private (rename to `analyzeOne`), expose only the forest-level entry from `aggregator.go`
- migrate `TestSimpleFib`, `TestSqrt`, etc. from `parser/parser_test.go` to `sema/integration_test.go`

**Delete:**
- the migrated parser tests from `parser/parser_test.go` (keep only the parser-internal ones: TestDanglingElse / TestNoSpuriousFork / etc.)

---

## Task 1: Rename Sema.Analyze → Sema.analyzeOne; add error stubs

**Files:** Modify `sema/sema.go`, create `sema/aggregator.go` (stub)

- [ ] **Step 1: Rename**

In `sema/sema.go`, rename:
```go
func (s *Sema) Analyze(root *entity.AstNode) *SemaResult { ... }
```
to:
```go
func (s *Sema) analyzeOne(root *entity.AstNode) *SemaResult { ... }
```

Update all call sites in `sema/*_test.go`:
```bash
grep -l "NewSema().Analyze" sema/
```
For each file, replace `NewSema().Analyze(` → `NewSema().analyzeOne(`.

- [ ] **Step 2: Build**

Run: `go build ./...`
Expected: success.

- [ ] **Step 3: Run all tests to confirm rename didn't break anything**

Run: `go test ./sema/`
Expected: PASS (same set as end of Plan B).

- [ ] **Step 4: Add aggregator stub**

`sema/aggregator.go`:
```go
package sema

import (
	"fmt"
	"sync"

	"shinya.click/cvm/common"
	"shinya.click/cvm/entity"
)

// Analyze is the public forest entry: PreFilter → concurrent per-candidate
// Sema → winner selection. Returns the chosen Program or a best-effort error.
func Analyze(candidates []*entity.AstNode) (*Program, error) {
	survivors, prefilterErrs := PreFilter(candidates)
	if len(survivors) == 0 {
		if len(prefilterErrs) > 0 {
			return nil, prefilterErrs[0]
		}
		return nil, fmt.Errorf("no candidates remain after PreFilter")
	}

	results := make([]*SemaResult, len(survivors))
	var wg sync.WaitGroup
	for i, tree := range survivors {
		wg.Add(1)
		go func(i int, tree *entity.AstNode) {
			defer wg.Done()
			results[i] = NewSema().analyzeOne(tree)
		}(i, tree)
	}
	wg.Wait()

	var clean []*SemaResult
	for _, r := range results {
		if r != nil && len(r.Errors) == 0 {
			clean = append(clean, r)
		}
	}
	switch len(clean) {
	case 1:
		return clean[0].Program, nil
	case 0:
		best := pickBestErrorResult(results)
		if best != nil && len(best.Errors) > 0 {
			return nil, best.Errors[0]
		}
		if len(prefilterErrs) > 0 {
			return nil, prefilterErrs[0]
		}
		return nil, fmt.Errorf("no result and no errors recorded")
	default:
		return nil, ambiguousParse(clean)
	}
}

func pickBestErrorResult(results []*SemaResult) *SemaResult {
	var best *SemaResult
	for _, r := range results {
		if r == nil {
			continue
		}
		if best == nil {
			best = r
			continue
		}
		// Fewer errors first.
		if len(r.Errors) < len(best.Errors) {
			best = r
			continue
		}
		if len(r.Errors) > len(best.Errors) {
			continue
		}
		// Tie: prefer the one whose first error is at the deeper position.
		if len(r.Errors) > 0 && len(best.Errors) > 0 {
			if compareErrPos(r.Errors[0], best.Errors[0]) > 0 {
				best = r
			}
		}
	}
	return best
}

func compareErrPos(a, b *common.CvmError) int {
	if len(a.Messages) == 0 || len(b.Messages) == 0 {
		return 0
	}
	pa, pb := a.Messages[0].SourcePos, b.Messages[0].SourcePos
	if pa.Line != pb.Line {
		return pa.Line - pb.Line
	}
	return pa.Column - pb.Column
}

func ambiguousParse(_ []*SemaResult) error {
	return fmt.Errorf("ambiguous parse: multiple candidates type-check cleanly (this should not happen for valid C)")
}
```

- [ ] **Step 5: Build + commit**

Run: `go build ./...`
Expected: success.

```bash
git add sema/sema.go sema/aggregator.go sema/*_test.go
git commit -m "refactor(sema): rename Sema.Analyze → analyzeOne; add public Analyze stub"
```

---

## Task 2: Aggregator concurrency + selection tests

**Files:** Create `sema/aggregator_test.go`

- [ ] **Step 1: Failing test**

`sema/aggregator_test.go`:
```go
package sema

import (
	"shinya.click/cvm/lexer"
	"shinya.click/cvm/parser"
	"sync"
	"testing"
)

func parseAll(t *testing.T, src string) []interface{} {
	t.Helper()
	tokens, err := lexer.NewLexer(src).ScanTokens()
	if err != nil {
		t.Fatal(err)
	}
	candidates, err := parser.NewParser(tokens).Parse()
	if err != nil {
		t.Fatal(err)
	}
	out := make([]interface{}, len(candidates))
	for i, c := range candidates {
		out[i] = c
	}
	return out
}

func TestAggregator_SingleCandidatePass(t *testing.T) {
	tokens, _ := lexer.NewLexer("int x = 5;").ScanTokens()
	candidates, _ := parser.NewParser(tokens).Parse()
	prog, err := Analyze(candidates)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if prog == nil || len(prog.Globals) != 1 {
		t.Fatalf("expected 1 global, got %+v", prog)
	}
}

func TestAggregator_TypedefShadowProducesOne(t *testing.T) {
	src := `typedef int a;
int main() {
    int a;
    int b;
    a*b;
}`
	tokens, _ := lexer.NewLexer(src).ScanTokens()
	candidates, _ := parser.NewParser(tokens).Parse()
	prog, err := Analyze(candidates)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if prog == nil {
		t.Fatal("nil program")
	}
}

func TestAggregator_BestErrorPicked(t *testing.T) {
	// Reference an undeclared identifier. Aggregator should report that
	// error rather than returning nil.
	tokens, _ := lexer.NewLexer("int main() { undeclared_function(); }").ScanTokens()
	candidates, _ := parser.NewParser(tokens).Parse()
	_, err := Analyze(candidates)
	if err == nil {
		t.Fatal("expected error for undeclared identifier")
	}
}

func TestAggregator_ConcurrencyStress(t *testing.T) {
	// Construct a typedef shadow that produces ≥ 2 survivors.
	src := `typedef int a; int main() { int a; a; }`
	tokens, _ := lexer.NewLexer(src).ScanTokens()
	candidates, _ := parser.NewParser(tokens).Parse()
	if len(candidates) < 1 {
		t.Skip("parser returned no candidates; cannot stress")
	}
	const iters = 50
	var wg sync.WaitGroup
	wg.Add(iters)
	for i := 0; i < iters; i++ {
		go func() {
			defer wg.Done()
			Analyze(candidates)
		}()
	}
	wg.Wait()
}
```

- [ ] **Step 2: Run tests**

Run: `go test ./sema/ -run TestAggregator`
Expected: PASS for all (TestAggregator_BestErrorPicked validates the err-path).

- [ ] **Step 3: Run with -race**

Run: `go test ./sema/ -run TestAggregator -race -count=2`
Expected: PASS, no data races. If a race fires, the most likely culprit is shared state in PreFilter (the package-level audit map from Plan A is gone; the per-Sema TypeTable / SymbolTable should be isolated). Investigate before proceeding.

- [ ] **Step 4: Commit**

```bash
git add sema/aggregator_test.go
git commit -m "test(sema): aggregator concurrency + selection tests (-race clean)"
```

---

## Task 3: IR pretty-printer for golden snapshots

**Files:** Create `sema/printer.go`, `sema/printer_test.go`

- [ ] **Step 1: Failing test**

`sema/printer_test.go`:
```go
package sema

import (
	"shinya.click/cvm/lexer"
	"shinya.click/cvm/parser"
	"strings"
	"testing"
)

func TestPrintIR_Simple(t *testing.T) {
	tokens, _ := lexer.NewLexer("int x = 42;").ScanTokens()
	candidates, _ := parser.NewParser(tokens).Parse()
	prog, _ := Analyze(candidates)
	out := PrintProgram(prog)
	if !strings.Contains(out, "VarDecl") || !strings.Contains(out, "x") || !strings.Contains(out, "42") {
		t.Fatalf("printer output missing expected pieces:\n%s", out)
	}
}
```

- [ ] **Step 2: Run test**

Run: `go test ./sema/ -run TestPrintIR_Simple`
Expected: FAIL.

- [ ] **Step 3: Implement printer**

`sema/printer.go`:
```go
package sema

import (
	"fmt"
	"strings"
)

func PrintProgram(p *Program) string {
	if p == nil {
		return "<nil program>\n"
	}
	var b strings.Builder
	b.WriteString("Program\n")
	for _, d := range p.Globals {
		printDecl(&b, 1, d)
	}
	for _, f := range p.Funcs {
		printFuncDef(&b, 1, f)
	}
	return b.String()
}

func indent(b *strings.Builder, n int) {
	for i := 0; i < n; i++ {
		b.WriteString("  ")
	}
}

func printDecl(b *strings.Builder, depth int, d Decl) {
	indent(b, depth)
	switch x := d.(type) {
	case *VarDecl:
		fmt.Fprintf(b, "VarDecl name=%q type=%s storage=%v\n", x.Sym.Name, x.T, x.Storage)
		if x.Init != nil {
			indent(b, depth+1)
			b.WriteString("init:\n")
			printExpr(b, depth+2, x.Init)
		}
	case *FuncDecl:
		fmt.Fprintf(b, "FuncDecl name=%q type=%s\n", x.Sym.Name, x.T)
	case *TypedefDecl:
		fmt.Fprintf(b, "TypedefDecl name=%q type=%s\n", x.Sym.Name, x.T)
	case *TagDecl:
		fmt.Fprintf(b, "TagDecl type=%s\n", x.T)
	default:
		fmt.Fprintf(b, "%T\n", d)
	}
}

func printFuncDef(b *strings.Builder, depth int, f *FuncDef) {
	indent(b, depth)
	fmt.Fprintf(b, "FuncDef name=%q type=%s\n", f.Sym.Name, f.T)
	for _, p := range f.Params {
		indent(b, depth+1)
		fmt.Fprintf(b, "Param name=%q type=%s\n", p.Sym.Name, p.T)
	}
	if f.Body != nil {
		printStmt(b, depth+1, f.Body)
	}
}

func printStmt(b *strings.Builder, depth int, s Stmt) {
	if s == nil {
		return
	}
	indent(b, depth)
	switch x := s.(type) {
	case *Block:
		b.WriteString("Block\n")
		for _, it := range x.Items {
			printStmt(b, depth+1, it)
		}
	case *ExprStmt:
		b.WriteString("ExprStmt\n")
		printExpr(b, depth+1, x.Expr)
	case *DeclStmt:
		b.WriteString("DeclStmt\n")
		for _, d := range x.Decls {
			printDecl(b, depth+1, d)
		}
	case *IfStmt:
		b.WriteString("IfStmt\n")
		indent(b, depth+1)
		b.WriteString("cond:\n")
		printExpr(b, depth+2, x.Cond)
		indent(b, depth+1)
		b.WriteString("then:\n")
		printStmt(b, depth+2, x.Then)
		if x.Else != nil {
			indent(b, depth+1)
			b.WriteString("else:\n")
			printStmt(b, depth+2, x.Else)
		}
	case *WhileStmt:
		if x.DoWhile {
			b.WriteString("DoWhileStmt\n")
		} else {
			b.WriteString("WhileStmt\n")
		}
		printExpr(b, depth+1, x.Cond)
		printStmt(b, depth+1, x.Body)
	case *ForStmt:
		b.WriteString("ForStmt\n")
		printStmt(b, depth+1, x.Init)
		printExpr(b, depth+1, x.Cond)
		printExpr(b, depth+1, x.Post)
		printStmt(b, depth+1, x.Body)
	case *ReturnStmt:
		b.WriteString("ReturnStmt\n")
		if x.Value != nil {
			printExpr(b, depth+1, x.Value)
		}
	case *BreakStmt:
		b.WriteString("BreakStmt\n")
	case *ContinueStmt:
		b.WriteString("ContinueStmt\n")
	case *GotoStmt:
		fmt.Fprintf(b, "GotoStmt -> %q\n", labelName(x))
	case *LabeledStmt:
		fmt.Fprintf(b, "LabeledStmt %q\n", x.Name)
		printStmt(b, depth+1, x.Body)
	case *CaseStmt:
		fmt.Fprintf(b, "CaseStmt %d\n", x.Value)
		printStmt(b, depth+1, x.Body)
	case *DefaultStmt:
		b.WriteString("DefaultStmt\n")
		printStmt(b, depth+1, x.Body)
	case *SwitchStmt:
		b.WriteString("SwitchStmt\n")
		printExpr(b, depth+1, x.Cond)
		printStmt(b, depth+1, x.Body)
	case *EmptyStmt:
		b.WriteString("EmptyStmt\n")
	default:
		fmt.Fprintf(b, "%T\n", s)
	}
}

func labelName(g *GotoStmt) string {
	if g.Target != nil {
		return g.Target.Name
	}
	return g.Name
}

func printExpr(b *strings.Builder, depth int, e Expr) {
	if e == nil {
		return
	}
	indent(b, depth)
	switch x := e.(type) {
	case *IntLit:
		fmt.Fprintf(b, "IntLit %d : %s\n", x.Value, x.T)
	case *FloatLit:
		fmt.Fprintf(b, "FloatLit %v : %s\n", x.Value, x.T)
	case *CharLit:
		fmt.Fprintf(b, "CharLit %d : %s\n", x.Value, x.T)
	case *StringLit:
		fmt.Fprintf(b, "StringLit %q : %s\n", x.Value, x.T)
	case *VarRef:
		fmt.Fprintf(b, "VarRef %q : %s\n", x.Sym.Name, x.T)
	case *EnumRef:
		fmt.Fprintf(b, "EnumRef %q=%d : %s\n", x.Enumerator.Name, x.Enumerator.Value, x.T)
	case *BinOp:
		fmt.Fprintf(b, "BinOp %v : %s\n", x.Op, x.T)
		printExpr(b, depth+1, x.L)
		printExpr(b, depth+1, x.R)
	case *UnOp:
		fmt.Fprintf(b, "UnOp %v : %s\n", x.Op, x.T)
		printExpr(b, depth+1, x.X)
	case *AssignExpr:
		fmt.Fprintf(b, "AssignExpr : %s\n", x.T)
		printExpr(b, depth+1, x.L)
		printExpr(b, depth+1, x.R)
	case *CompoundAssign:
		fmt.Fprintf(b, "CompoundAssign %v : %s\n", x.Op, x.T)
		printExpr(b, depth+1, x.L)
		printExpr(b, depth+1, x.R)
	case *CallExpr:
		fmt.Fprintf(b, "CallExpr : %s\n", x.T)
		printExpr(b, depth+1, x.Callee)
		for _, a := range x.Args {
			printExpr(b, depth+1, a)
		}
	case *MemberExpr:
		arrow := "."
		if x.Arrow {
			arrow = "->"
		}
		field := ""
		if x.Field != nil {
			field = x.Field.Name
		}
		fmt.Fprintf(b, "MemberExpr %s%s : %s\n", arrow, field, x.T)
		printExpr(b, depth+1, x.Base)
	case *IndexExpr:
		fmt.Fprintf(b, "IndexExpr : %s\n", x.T)
		printExpr(b, depth+1, x.Base)
		printExpr(b, depth+1, x.Index)
	case *CondExpr:
		fmt.Fprintf(b, "CondExpr : %s\n", x.T)
		printExpr(b, depth+1, x.Cond)
		printExpr(b, depth+1, x.Then)
		printExpr(b, depth+1, x.Else)
	case *SizeofExpr:
		fmt.Fprintf(b, "SizeofExpr : %s\n", x.T)
		if x.Operand.Type != nil {
			indent(b, depth+1)
			fmt.Fprintf(b, "type: %s\n", x.Operand.Type)
		}
		if x.Operand.Expr != nil {
			printExpr(b, depth+1, x.Operand.Expr)
		}
	case *CommaExpr:
		fmt.Fprintf(b, "CommaExpr : %s\n", x.T)
		printExpr(b, depth+1, x.L)
		printExpr(b, depth+1, x.R)
	case *ImplicitCast:
		fmt.Fprintf(b, "ImplicitCast %v %s -> %s\n", x.Kind, x.From, x.To)
		printExpr(b, depth+1, x.X)
	case *ExplicitCast:
		fmt.Fprintf(b, "ExplicitCast -> %s\n", x.To)
		printExpr(b, depth+1, x.X)
	case *AddrConst:
		fmt.Fprintf(b, "AddrConst &%s + %d : %s\n", x.Sym.Name, x.Offset, x.T)
	case *InitList:
		fmt.Fprintf(b, "InitList : %s\n", x.T)
		for _, el := range x.Elems {
			printExpr(b, depth+1, el.Value)
		}
	case *CompoundLit:
		fmt.Fprintf(b, "CompoundLit : %s\n", x.T)
		printExpr(b, depth+1, x.Init)
	default:
		fmt.Fprintf(b, "%T : %s\n", e, e.GetType())
	}
}
```

- [ ] **Step 4: Run test**

Run: `go test ./sema/ -run TestPrintIR_Simple`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add sema/printer.go sema/printer_test.go
git commit -m "feat(sema): IR pretty-printer for golden snapshots"
```

---

## Task 4: testdata/pass golden snapshot infrastructure

**Files:** Create `sema/testdata/pass/simple_var.c`, `sema/testdata/pass/simple_var.ir.golden`, `sema/golden_test.go`

- [ ] **Step 1: Add a starter pass case**

`sema/testdata/pass/simple_var.c`:
```c
int x = 42;
```

Generate the golden file by running the printer manually first (Step 4 captures it).

- [ ] **Step 2: Add golden_test.go**

`sema/golden_test.go`:
```go
package sema

import (
	"flag"
	"os"
	"path/filepath"
	"shinya.click/cvm/lexer"
	"shinya.click/cvm/parser"
	"strings"
	"testing"
)

var update = flag.Bool("update", false, "regenerate testdata/*.golden files")

func TestGoldenPass(t *testing.T) {
	matches, err := filepath.Glob("testdata/pass/*.c")
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) == 0 {
		t.Skip("no testdata/pass cases yet")
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
			got := PrintProgram(prog)
			goldenPath := strings.TrimSuffix(src, ".c") + ".ir.golden"
			if *update {
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
	matches, err := filepath.Glob("testdata/fail/*.c")
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
			tokens, err := lexer.NewLexer(string(source)).ScanTokens()
			if err != nil {
				goldenPath := strings.TrimSuffix(src, ".c") + ".err.golden"
				assertErrorMatches(t, err.Error(), goldenPath)
				return
			}
			candidates, perr := parser.NewParser(tokens).Parse()
			if perr != nil {
				goldenPath := strings.TrimSuffix(src, ".c") + ".err.golden"
				assertErrorMatches(t, perr.Error(), goldenPath)
				return
			}
			_, serr := Analyze(candidates)
			if serr == nil {
				t.Fatalf("expected error, got none")
			}
			goldenPath := strings.TrimSuffix(src, ".c") + ".err.golden"
			assertErrorMatches(t, serr.Error(), goldenPath)
		})
	}
}

func assertErrorMatches(t *testing.T, got, goldenPath string) {
	t.Helper()
	if *update {
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
```

- [ ] **Step 3: Generate the starter golden**

Run: `go test ./sema/ -run TestGoldenPass -update`
Expected: PASS, creates `simple_var.ir.golden`.

- [ ] **Step 4: Run without -update to confirm regression check works**

Run: `go test ./sema/ -run TestGoldenPass`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add sema/testdata sema/golden_test.go
git commit -m "test(sema): testdata/pass + testdata/fail golden snapshot infrastructure"
```

---

## Task 5: Add golden cases for diverse programs

**Files:** Create more `sema/testdata/pass/*.c` and corresponding goldens

- [ ] **Step 1: Add cases**

Create files:

`sema/testdata/pass/typedef.c`:
```c
typedef int Int32;
Int32 x;
```

`sema/testdata/pass/struct_member.c`:
```c
struct Point { int x; int y; };
struct Point p;
```

`sema/testdata/pass/function_decl_call.c`:
```c
int add(int a, int b);
int sum = add(1, 2);
```

`sema/testdata/pass/control_flow.c`:
```c
int main(void) {
    int x;
    if (x) return 1;
    while (x > 0) x = x - 1;
    for (int i = 0; i < 10; i = i + 1) {}
    return 0;
}
```

`sema/testdata/pass/pointer_arith.c`:
```c
int a[3];
int *p = a;
int v = *(p + 1);
```

`sema/testdata/pass/enum_case.c`:
```c
enum Color { RED, GREEN, BLUE };
int main(void) {
    int c = GREEN;
    switch (c) {
        case RED: return 0;
        case GREEN: return 1;
        case BLUE: return 2;
    }
    return -1;
}
```

- [ ] **Step 2: Generate goldens**

Run: `go test ./sema/ -run TestGoldenPass -update`
Expected: PASS.

- [ ] **Step 3: Inspect each golden file**

Run: `cat sema/testdata/pass/typedef.ir.golden`
Confirm the output looks structurally sane (matches the input semantics). Repeat for each.

- [ ] **Step 4: Commit goldens**

```bash
git add sema/testdata/pass/
git commit -m "test(sema): golden snapshots for typedef/struct/call/control-flow/pointer/enum"
```

- [ ] **Step 5: Re-run without -update**

Run: `go test ./sema/ -run TestGoldenPass`
Expected: PASS.

---

## Task 6: Add golden fail cases

**Files:** Create `sema/testdata/fail/*.c` + `*.err.golden`

- [ ] **Step 1: Add fail cases**

`sema/testdata/fail/undeclared.c`:
```c
int main(void) {
    return undeclared_var;
}
```

`sema/testdata/fail/type_mismatch.c`:
```c
int main(void) {
    int *p;
    p = 3.14;
    return 0;
}
```

`sema/testdata/fail/redefinition.c`:
```c
int x;
typedef int x;
```

`sema/testdata/fail/case_not_constant.c`:
```c
int main(void) {
    int x;
    switch (x) {
        case x + 1: return 0;
    }
}
```

- [ ] **Step 2: Generate goldens**

Run: `go test ./sema/ -run TestGoldenFail -update`
Expected: PASS.

- [ ] **Step 3: Inspect goldens — should each contain a meaningful error message**

Run: `cat sema/testdata/fail/undeclared.err.golden`
Confirm it includes "undeclared". Repeat for others. If the message is misleading, fix the corresponding error in `sema/error.go` first, then re-update.

- [ ] **Step 4: Trim goldens to substring match (optional)**

Edit each `.err.golden` to keep only a discriminating substring (e.g. `use of undeclared identifier 'undeclared_var'`). The test uses substring match.

- [ ] **Step 5: Commit**

```bash
git add sema/testdata/fail/
git commit -m "test(sema): golden snapshots for undeclared/type-mismatch/redef/case-not-const"
```

---

## Task 7: Wire compiler.go to use sema.Analyze

**Files:** Modify `compiler.go`

- [ ] **Step 1: Inspect current Plan-A glue**

Run: `cat compiler.go`

Current `RunSource` calls `sema.PreFilter` only.

- [ ] **Step 2: Replace with full Analyze**

In `compiler.go`, replace the body of `RunSource`:
```go
func (c *Compiler) RunSource(source string) {
	if c.FileName == "" {
		c.FileName = "main.c"
	}
	c.Source = source
	c.Lines = strings.Split(source, "\n")
	tokens, err := lexer.NewLexer(c.Source).ScanTokens()
	if err != nil {
		c.handleError(err)
		return
	}
	candidates, err := parser.NewParser(tokens).Parse()
	if err != nil {
		c.handleError(err)
		return
	}
	prog, err := sema.Analyze(candidates)
	if err != nil {
		c.handleError(err)
		return
	}
	_ = prog
	// Codegen wired in a future plan.
}
```

- [ ] **Step 3: Build**

Run: `go build ./...`
Expected: success.

- [ ] **Step 4: Run all tests**

Run: `go test ./... -count=1`
Expected: same PASS set as end of Plan B (parser tests still pass; sema tests including aggregator + golden pass; compiler_test passes if any).

- [ ] **Step 5: Commit**

```bash
git add compiler.go
git commit -m "refactor(compiler): use sema.Analyze instead of bare PreFilter"
```

---

## Task 8: Migrate parser end-to-end tests to sema

**Files:** Modify `parser/parser_test.go`, `sema/integration_test.go`

- [ ] **Step 1: Identify parser tests that exercise the full pipeline**

Run: `grep -l "Parse()" parser/parser_test.go | head -1`

Tests like `TestSimpleFib`, `TestSqrt` are end-to-end. Tests like `TestDanglingElse`, `TestNoSpuriousFork`, `TestCandidateCountBoundedOnShadow`, `TestBestErrorPosition` are parser-internal.

End-to-end candidates to move:
- `TestSimpleFib`
- `TestSqrt`
- `TestStructEnumDef`
- `TestComplexTypeDef`
- `TestTypeDefDeclaration`
- `TestTypeName`
- `TestFunctionDeclaration1` / `2` / `3`
- `TestFuncDef`
- `TestTwoFunc`
- `TestSizeOf`
- `TestAmbiguous`

Keep parser-internal:
- `TestDanglingElse`
- `TestNoSpuriousFork`
- `TestCandidateCountBoundedOnShadow`
- `TestBestErrorPosition`

- [ ] **Step 2: Move to sema/integration_test.go**

Append each migrated test to `sema/integration_test.go`, replacing the test body's parser-only assertion with `Analyze(candidates)` and an error check.

Example template:
```go
func TestE2E_SimpleFib(t *testing.T) {
	src := `int Fibon1(int n){
    if (n == 1 || n == 2){
        return 1;
    } else{
        return Fibon1(n - 1) + Fibon1(n - 2);
    }
}
int scanf(const char *, ...);
int printf(const char *, ...);
int main(){
    int n = 0;
    int ret = 0;
    scanf("%d", &n);
    ret = Fibon1(n);
    printf("ret=%d", ret);
    return 0;
}`
	tokens, err := lexer.NewLexer(src).ScanTokens()
	if err != nil {
		t.Fatal(err)
	}
	candidates, err := parser.NewParser(tokens).Parse()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Analyze(candidates); err != nil {
		t.Fatalf("sema: %v", err)
	}
}
```

Note: `scanf`/`printf` declarations are now part of the source (no preprocessor / libc). This makes the test self-contained.

- [ ] **Step 3: Delete migrated tests from parser/parser_test.go**

Remove `TestSimpleFib` etc. from `parser/parser_test.go`.

- [ ] **Step 4: Run all tests**

Run: `go test ./... -count=1`
Expected: PASS for all (parser-internal kept tests + sema migrated tests + sema units + sema goldens + sema integration).

- [ ] **Step 5: Commit**

```bash
git add parser/parser_test.go sema/integration_test.go
git commit -m "test: migrate end-to-end parser tests to sema package"
```

---

## Task 9: Final cleanup pass

**Files:** none (cleanup only)

- [ ] **Step 1: gofmt sweep**

Run: `gofmt -l ./...`
Expected: no output. If any, run `gofmt -w ./...` and commit:
```bash
gofmt -w ./...
git add -u
git commit -m "style: gofmt sweep at end of Plan C"
```

- [ ] **Step 2: go vet sweep**

Run: `go vet ./...`
Expected: clean.

- [ ] **Step 3: Race-clean test sweep**

Run: `go test ./... -race -count=2`
Expected: PASS, no races.

- [ ] **Step 4: Confirm package shapes**

Run: `find sema -type f -name "*.go" | sort`

Expected files:
- `sema/aggregator.go`, `sema/aggregator_test.go`
- `sema/cast.go`, `sema/cast_test.go`
- `sema/const.go`, `sema/const_test.go`
- `sema/decl.go`, `sema/decl_test.go`
- `sema/doc.go`
- `sema/error.go`
- `sema/expr.go`, `sema/expr_test.go`
- `sema/golden_test.go`
- `sema/init.go`, `sema/init_test.go`
- `sema/integration_test.go`
- `sema/ir.go`, `sema/ir_test.go`
- `sema/op.go`
- `sema/prefilter.go`, `sema/prefilter_test.go`
- `sema/printer.go`, `sema/printer_test.go`
- `sema/scope.go`, `sema/scope_test.go`
- `sema/sema.go`, `sema/sema_test.go`
- `sema/spec.go`, `sema/spec_test.go`
- `sema/stmt.go`, `sema/stmt_test.go`
- `sema/test_util_test.go`
- `sema/type.go`, `sema/type_table.go`, `sema/type_uintptr.go`, `sema/type_test.go`

Run: `find sema/testdata -type f | sort`

Expected:
- multiple `pass/*.c` and `pass/*.ir.golden`
- multiple `fail/*.c` and `fail/*.err.golden`

- [ ] **Step 5: Final commit if any drift**

If gofmt / vet / race produced changes, commit them. Otherwise nothing more to commit.

---

## Plan C Acceptance

After Task 9:
- `sema.Analyze(candidates)` is the single public forest entry; concurrent per-survivor execution; race-clean.
- `compiler.go` uses `sema.Analyze` instead of bare PreFilter.
- IR pretty-printer powers golden snapshots in `sema/testdata/pass/`.
- Error golden snapshots in `sema/testdata/fail/` capture meaningful diagnostics.
- All previous parser end-to-end tests now run as sema integration tests with self-contained source (no preprocessor needed).
- Parser package retains only its own internal tests (dangling-else / fork-pruning / best-error / shadow bound).
- `go test ./... -race` clean.

The complete Sema layer is now ready for Codegen as the next plan.
