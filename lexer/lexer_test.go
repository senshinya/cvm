package lexer

import (
	"fmt"
	"testing"
)

func TestCharacterLiteral(t *testing.T) {
	tokens, err := NewLexer("'abc''\\n' '\\123''\\xAc'\t  \n \n 'abc\\n\\123\\xAc'").ScanTokens()
	if err != nil {
		panic(err)
	}
	fmt.Println(tokens)
}

func Test123(t *testing.T) {
	tokens, err := NewLexer("'\\''").ScanTokens()
	if err != nil {
		panic(err)
	}
	fmt.Println(tokens)
}

func TestIdentifier(t *testing.T) {
	tokens, err := NewLexer("abc _12 ab_1 _i4 ").ScanTokens()
	if err != nil {
		panic(err)
	}
	fmt.Println(tokens)
}

func TestStringLiteral(t *testing.T) {
	tokens, err := NewLexer("'abc'\"abc\\n\\tCTAABC\\xAA\\123\"").ScanTokens()
	if err != nil {
		panic(err)
	}
	fmt.Println(tokens)
}

func TestNumberLiteral(t *testing.T) {
	tokens, err := NewLexer(`42
1234567890
0777
0
0x1F
0XABCDEF
0x123456789ABCDEF
42U
42L
42UL
42LU
42LL
42ULL
3.14159
0.1
.123
123.
1e10
1E+10
1e-10
3.14e-10
0x1.1p10
0X1.FFP+10
0x0.1p-10
0x1p10
3.14f
3.14F
3.14l
3.14L
.1e10
2.5e-3
.1E+10
0X.FFP+10
0x1.8p-3
0x.1p10
`).ScanTokens()
	if err != nil {
		panic(err)
	}
	fmt.Println(tokens)
}

func TestNumberLiteral1(t *testing.T) {
	tokens, err := NewLexer(`3.14L
3.14L`).ScanTokens()
	if err != nil {
		panic(err)
	}
	fmt.Println(tokens)
}

func TestAll(t *testing.T) {
	tokens, err := NewLexer("'abc''\\n' '\\123''\\xAc''abc\\n\\123\\xAc'\nabc _12 ab_1 _i4 ").ScanTokens()
	if err != nil {
		panic(err)
	}
	fmt.Println(tokens)
}

func TestIds(t *testing.T) {
	tokens, err := NewLexer("a,b,c").ScanTokens()
	if err != nil {
		panic(err)
	}
	fmt.Println(tokens)
}
