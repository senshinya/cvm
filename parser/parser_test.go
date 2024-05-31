package parser

import (
	"shinya.click/cvm/lexer"
	"testing"
)

func TestTypeDefDeclaration(t *testing.T) {
	tokens, err := lexer.NewLexer("volatile int (*const a(float))[2*3];").ScanTokens()
	if err != nil {
		panic(err)
	}
	NewParser(tokens).Parse()
	//
	//tokens, err := lexer.NewLexer("volatile ABC abc;").ScanTokens()
	//if err != nil {
	//	panic(err)
	//}
	//NewParser(tokens).Parse()
}
