package lexer

import (
	"fmt"
	"testing"
)

func TestLexer(t *testing.T) {
	fmt.Println(NewLexer("'abc''\\n''\\123''\\xAc6''abc\\n\\123\\xAc6'").ScanTokens())
}
