package sema

import (
	"strings"
	"sync"
	"testing"

	"shinya.click/cvm/entity"
	"shinya.click/cvm/lexer"
	"shinya.click/cvm/parser"
)

func parseCandidates(t *testing.T, src string) []*entity.AstNode {
	t.Helper()
	tokens, err := lexer.NewLexer(src).ScanTokens()
	if err != nil {
		t.Fatal(err)
	}
	candidates, err := parser.NewParser(tokens).Parse()
	if err != nil {
		t.Fatal(err)
	}
	return candidates
}

func TestAggregator_SingleCandidatePass(t *testing.T) {
	prog, err := Analyze(parseCandidates(t, "int x = 5;"))
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
	prog, err := Analyze(parseCandidates(t, src))
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if prog == nil {
		t.Fatal("nil program")
	}
}

func TestAggregator_BestErrorPicked(t *testing.T) {
	_, err := Analyze(parseCandidates(t, "int main() { return undeclared_var; }"))
	if err == nil {
		t.Fatal("expected error for undeclared identifier")
	}
	if !strings.Contains(err.Error(), "undeclared_var") {
		t.Fatalf("expected undeclared identifier error, got %v", err)
	}
}

func TestAggregator_ConcurrencyStress(t *testing.T) {
	// 复用同一组候选树并发调用，确认聚合器没有跨候选共享 Sema 状态。
	candidates := parseCandidates(t, `typedef int a; int main() { int a; a; }`)
	if len(candidates) < 1 {
		t.Skip("parser returned no candidates; cannot stress")
	}

	const iters = 50
	var wg sync.WaitGroup
	wg.Add(iters)
	errs := make(chan error, iters)
	for i := 0; i < iters; i++ {
		go func() {
			defer wg.Done()
			_, err := Analyze(candidates)
			errs <- err
		}()
	}
	wg.Wait()
	close(errs)

	for err := range errs {
		if err != nil {
			t.Fatalf("unexpected concurrent Analyze error: %v", err)
		}
	}
}
