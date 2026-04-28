package preprocessor

import (
	"testing"

	"shinya.click/cvm/entity"
)

func TestConvertExpandedTokensToParserTokens(t *testing.T) {
	res, err := PreprocessSource("main.c", "int x = 1;\n", Options{})
	if err != nil {
		t.Fatalf("PreprocessSource failed: %v", err)
	}
	want := []entity.TokenType{entity.INT, entity.IDENTIFIER, entity.EQUAL, entity.INTEGER_CONSTANT, entity.SEMICOLON, entity.EOF}
	if len(res.Tokens) != len(want) {
		t.Fatalf("token count = %d, want %d: %#v", len(res.Tokens), len(want), res.Tokens)
	}
	for i, typ := range want {
		if res.Tokens[i].Typ != typ {
			t.Fatalf("token[%d] type = %s, want %s", i, res.Tokens[i].Typ, typ)
		}
		if res.Tokens[i].SourceStart.LocationID == 0 && typ != entity.EOF {
			t.Fatalf("token[%d] missing source location: %#v", i, res.Tokens[i])
		}
	}
}
