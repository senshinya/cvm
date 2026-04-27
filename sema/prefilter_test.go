package sema

import (
	"testing"

	"shinya.click/cvm/entity"
	"shinya.click/cvm/lexer"
	"shinya.click/cvm/parser"
)

func TestPreFilterTypedefShadow(t *testing.T) {
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
	candidates, err := parser.NewParser(tokens).Parse()
	if err != nil {
		t.Fatal(err)
	}
	if len(candidates) == 0 {
		t.Fatal("parser returned no candidates")
	}
	survivors, _ := PreFilter(candidates)
	if len(survivors) == 0 {
		t.Fatal("PreFilter culled all candidates")
	}
}

func TestPreFilterDoesNotCullUndeclared(t *testing.T) {
	ident := &entity.Token{
		Typ:         entity.IDENTIFIER,
		Lexeme:      "scanf",
		SourceRange: entity.SourceRange{SourceStart: entity.SourcePos{Line: 1, Column: 14}},
	}
	root := &entity.AstNode{
		Typ:        parser.PrimaryExpression,
		Production: entity.Production{Left: parser.PrimaryExpression, Index: 1},
		Children: []*entity.AstNode{
			{
				Typ:      entity.IDENTIFIER,
				Terminal: ident,
			},
		},
	}
	survivors, _ := PreFilter([]*entity.AstNode{root})
	if len(survivors) == 0 {
		t.Fatal("PreFilter incorrectly culled undeclared-identifier candidate")
	}
}
