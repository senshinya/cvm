package lexer

import (
	"fmt"
	"testing"
)

func TestCharacterLiteral(t *testing.T) {
	fmt.Println(NewLexer("'abc''\\n' '\\123''\\xAc'\t  \n \n 'abc\\n\\123\\xAc'").ScanTokens())
}

func Test123(t *testing.T) {
	fmt.Println(NewLexer("'\\''").ScanTokens())
}

func TestIdentifier(t *testing.T) {
	fmt.Println(NewLexer("abc _12 ab_1 _i4 ").ScanTokens())
}

func TestStringLiteral(t *testing.T) {
	fmt.Println(NewLexer("'abc'\"abc\\n\\tCTAABC\\xAA\\123\"").ScanTokens())
}

func TestAll(t *testing.T) {
	fmt.Println(NewLexer("'abc''\\n' '\\123''\\xAc''abc\\n\\123\\xAc'\nabc _12 ab_1 _i4 ").ScanTokens())
}
