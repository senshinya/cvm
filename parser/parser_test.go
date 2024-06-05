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

func TestTypeName(t *testing.T) {
	tokens, err := lexer.NewLexer("int a[sizeof(int (*const [])(unsigned int, ...))];").ScanTokens()
	if err != nil {
		panic(err)
	}
	NewParser(tokens).Parse()
}

func TestFunctionDeclaration1(t *testing.T) {
	tokens, err := lexer.NewLexer("int (*fpfi(int (*)(long), int))(int, ...);").ScanTokens()
	if err != nil {
		panic(err)
	}
	NewParser(tokens).Parse()
}

func TestFunctionDeclaration2(t *testing.T) {
	tokens, err := lexer.NewLexer("int f(void), *fip(), (*pfi)();").ScanTokens()
	if err != nil {
		panic(err)
	}
	NewParser(tokens).Parse()
}

func TestFunctionDeclaration3(t *testing.T) {
	tokens, err := lexer.NewLexer("int (*apfi[3])(int *x, int *y);").ScanTokens()
	if err != nil {
		panic(err)
	}
	NewParser(tokens).Parse()
}
