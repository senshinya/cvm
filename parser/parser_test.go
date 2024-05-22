package parser

import (
	"shinya.click/cvm/lexer"
	"testing"
)

func TestTypeDefDeclaration(t *testing.T) {
	tokens, err := lexer.NewLexer("typedef int a,b,c;").ScanTokens()
	if err != nil {
		panic(err)
	}
	//NewParser(tokens).ConstructAST()
	tokens, err = lexer.NewLexer("int (*a(float))[1];").ScanTokens()
	if err != nil {
		panic(err)
	}
	NewParser(tokens).ConstructAST()
}
