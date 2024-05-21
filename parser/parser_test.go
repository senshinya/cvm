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
	tokens, err = lexer.NewLexer("typedef void fv(int), (*pfv)(int);").ScanTokens()
	if err != nil {
		panic(err)
	}
	NewParser(tokens).ConstructAST()
}
