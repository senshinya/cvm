package sema

import (
	"testing"

	"shinya.click/cvm/common"
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

func TestPreFilterC99SourceSmoke(t *testing.T) {
	tests := []struct {
		name string
		src  string
	}{
		{
			name: "typedef declaration",
			src: `typedef int T;
int main(void) {
    T *p;
    return 0;
}`,
		},
		{
			name: "block shadow typedef",
			src: `typedef int T;
int main(void) {
    int T;
    T * p;
    return 0;
}`,
		},
		{
			name: "function parameter shadows typedef",
			src: `typedef int T;
void f(int T) {
    T * p;
}`,
		},
		{
			name: "c99 for declaration",
			src: `int main(void) {
    for (int i = 0; i < 1; i = i + 1) {
        int j;
    }
    return 0;
}`,
		},
		{
			name: "undeclared function remains survivor",
			src: `int main(void) {
    scanf("%d", &x);
    return 0;
}`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			survivors, errs := parseAndPreFilter(t, tc.src)
			if len(survivors) == 0 {
				t.Fatalf("PreFilter returned no survivors; errs=%d", len(errs))
			}
		})
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

func parseAndPreFilter(t *testing.T, src string) ([]*entity.AstNode, []*common.CvmError) {
	t.Helper()
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
	return PreFilter(candidates)
}

func TestPreFilterCullsTypedefUsedAsPrimaryExpression(t *testing.T) {
	root := astSeq(
		typedefDecl("T", 1),
		primaryIdent("T", 2),
	)
	survivors, errs := PreFilter([]*entity.AstNode{root})
	if len(survivors) != 0 {
		t.Fatal("PreFilter kept candidate where typedef name was used as expression")
	}
	if len(errs) != 1 {
		t.Fatalf("PreFilter returned %d errors, want 1", len(errs))
	}
}

func TestPreFilterCullsTypedefNameWithoutVisibleTypedef(t *testing.T) {
	root := typedefName("T", 1)
	survivors, errs := PreFilter([]*entity.AstNode{root})
	if len(survivors) != 0 {
		t.Fatal("PreFilter kept typedef_name candidate for non-typedef identifier")
	}
	if len(errs) != 1 {
		t.Fatalf("PreFilter returned %d errors, want 1", len(errs))
	}
}

func TestPreFilterAllowsInnerOrdinaryIdentifierToShadowTypedef(t *testing.T) {
	root := astSeq(
		typedefDecl("T", 1),
		tokenNode(entity.LEFT_BRACES, "{", 2),
		varDecl("T", 3),
		primaryIdent("T", 4),
		tokenNode(entity.RIGHT_BRACES, "}", 5),
	)
	survivors, errs := PreFilter([]*entity.AstNode{root})
	if len(survivors) != 1 {
		t.Fatalf("PreFilter culled shadowing candidate; survivors=%d errs=%d", len(survivors), len(errs))
	}
}

func astSeq(children ...*entity.AstNode) *entity.AstNode {
	return &entity.AstNode{
		Typ:      parser.Program,
		Children: children,
	}
}

func typedefDecl(name string, line int) *entity.AstNode {
	return decl(name, line, true)
}

func varDecl(name string, line int) *entity.AstNode {
	return decl(name, line, false)
}

func decl(name string, line int, isTypedef bool) *entity.AstNode {
	tok := identToken(name, line)
	spec := &entity.AstNode{
		Typ:         parser.DeclarationSpecifiers,
		GLRLabel:    entity.GLRLabel{TypeDef: isTypedef, DeclaratorID: []*entity.Token{tok}},
		SourceRange: tok.SourceRange,
	}
	return &entity.AstNode{
		Typ:         parser.Declaration,
		Children:    []*entity.AstNode{spec},
		SourceRange: tok.SourceRange,
	}
}

func primaryIdent(name string, line int) *entity.AstNode {
	tok := identToken(name, line)
	return &entity.AstNode{
		Typ:        parser.PrimaryExpression,
		Production: entity.Production{Left: parser.PrimaryExpression, Index: 1},
		Children: []*entity.AstNode{
			{
				Typ:         entity.IDENTIFIER,
				Terminal:    tok,
				SourceRange: tok.SourceRange,
			},
		},
		SourceRange: tok.SourceRange,
	}
}

func typedefName(name string, line int) *entity.AstNode {
	tok := identToken(name, line)
	return &entity.AstNode{
		Typ:        parser.TypedefName,
		Production: entity.Production{Left: parser.TypedefName, Index: 1},
		Children: []*entity.AstNode{
			{
				Typ:         entity.IDENTIFIER,
				Terminal:    tok,
				SourceRange: tok.SourceRange,
			},
		},
		SourceRange: tok.SourceRange,
	}
}

func tokenNode(typ entity.TokenType, lexeme string, line int) *entity.AstNode {
	tok := &entity.Token{
		Typ:         typ,
		Lexeme:      lexeme,
		SourceRange: entity.SourceRange{SourceStart: entity.SourcePos{Line: line, Column: 1}},
	}
	return &entity.AstNode{Typ: typ, Terminal: tok, SourceRange: tok.SourceRange}
}

func identToken(name string, line int) *entity.Token {
	return &entity.Token{
		Typ:         entity.IDENTIFIER,
		Lexeme:      name,
		SourceRange: entity.SourceRange{SourceStart: entity.SourcePos{Line: line, Column: 1}},
	}
}
