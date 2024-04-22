package lexer

import (
	"fmt"
	"testing"
)

func TestCharacterLiteral(t *testing.T) {
	fmt.Println(NewLexer("'abc''\\n' '\\123''\\xAc6'\t  \n \n 'abc\\n\\123\\xAc6'").ScanTokens())
}

func TestIdentifier(t *testing.T) {
	fmt.Println(NewLexer("abc _12 ab_1 _i4 ").ScanTokens())
}

func TestAll(t *testing.T) {
	fmt.Println(NewLexer("'abc''\\n' '\\123''\\xAc6''abc\\n\\123\\xAc6'\nabc _12 ab_1 _i4 ").ScanTokens())
}
