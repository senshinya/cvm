package parser

import (
	"errors"
	"shinya.click/cvm/common"
	"shinya.click/cvm/entity"
	"shinya.click/cvm/lexer"
	"testing"
)

func TestDanglingElse(t *testing.T) {
	src := `void f() {
	int a, b, x, y;
	if (a) if (b) x; else y;
}`
	candidates := mustParse(t, src)
	outer := findFirst(candidates[0], SelectionStatement)
	if outer == nil {
		t.Fatal("no selection_statement in tree")
	}
	if !outer.ReducedBy(SelectionStatement, 1) {
		t.Fatalf("outer if reduced by %s, want production 1 (no ELSE bound)", outer.Production)
	}
	innerStmt := outer.Children[len(outer.Children)-1]
	inner := findFirst(innerStmt, SelectionStatement)
	if inner == nil {
		t.Fatal("no inner selection_statement")
	}
	if !inner.ReducedBy(SelectionStatement, 2) {
		t.Fatalf("inner if reduced by %s, want production 2 (ELSE bound to inner)", inner.Production)
	}
}

func TestNoSpuriousFork(t *testing.T) {
	cases := []struct {
		name string
		src  string
	}{
		{"plain function", `int main() { int x; x = 1; return 0; }`},
		{"complex typedef", `typedef int (*(*a)(void))(void);`},
		{"function decl with typedef param", `typedef int Int; void f(Int x);`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tokens, err := lexer.NewLexer(tc.src).ScanTokens()
			if err != nil {
				t.Fatal(err)
			}
			p := NewParser(tokens)
			if _, err := p.Parse(); err != nil {
				t.Fatal(err)
			}
			if p.ForkCount != 0 {
				t.Fatalf("ForkCount = %d, want 0 — typedef pruning should resolve all conflicts here", p.ForkCount)
			}
		})
	}
}

func TestCandidateCountBoundedOnShadow(t *testing.T) {
	// Outer typedef shadowed by inner variable. The parser cannot resolve this
	// without proper scoping, so it must keep both branches alive — but that
	// should not blow up to many candidates.
	src := `typedef int a;
int main() {
	int a;
	int b;
	a*b;
}`
	tokens, err := lexer.NewLexer(src).ScanTokens()
	if err != nil {
		t.Fatal(err)
	}
	candidates, err := NewParser(tokens).Parse()
	if err != nil {
		t.Fatal(err)
	}
	if len(candidates) == 0 {
		t.Fatal("expected at least one candidate")
	}
	if len(candidates) > 4 {
		t.Fatalf("len(candidates) = %d, want <= 4 — fork pruning should keep this bounded", len(candidates))
	}
}

func TestBestErrorPosition(t *testing.T) {
	// Lines 1-3 parse fine; line 4 contains a `;` that breaks the function
	// declaration mid-parameter-list.
	src := `int a;
int b;
int c;
int f(int ;);`
	tokens, err := lexer.NewLexer(src).ScanTokens()
	if err != nil {
		t.Fatal(err)
	}
	_, err = NewParser(tokens).Parse()
	if err == nil {
		t.Fatal("expected parse error")
	}
	var cvmErr *common.CvmError
	if !errors.As(err, &cvmErr) || len(cvmErr.Messages) == 0 {
		t.Fatalf("expected CvmError with messages, got %T: %v", err, err)
	}
	pos := cvmErr.Messages[0].SourcePos
	if pos.Line < 4 {
		t.Fatalf("error reported at line %d, want >= 4 — best-error heuristic should track the deepest branch", pos.Line)
	}
}

func mustParse(t *testing.T, src string) []*entity.AstNode {
	t.Helper()
	tokens, err := lexer.NewLexer(src).ScanTokens()
	if err != nil {
		t.Fatal(err)
	}
	candidates, err := NewParser(tokens).Parse()
	if err != nil {
		t.Fatal(err)
	}
	if len(candidates) == 0 {
		t.Fatal("no candidates returned")
	}
	return candidates
}

func findFirst(node *entity.AstNode, typ entity.TokenType) *entity.AstNode {
	if node == nil {
		return nil
	}
	if node.Typ == typ {
		return node
	}
	for _, child := range node.Children {
		if found := findFirst(child, typ); found != nil {
			return found
		}
	}
	return nil
}
