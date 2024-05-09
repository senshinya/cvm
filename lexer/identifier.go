package lexer

import (
	"shinya.click/cvm/common"
	"shinya.click/cvm/lexer/util"
)

var identifierStateTable = stateTable{
	"A": []Edge{{"letter_", "B"}},
	"B": []Edge{{"letter_", "B"}, {"digit", "B"}},
}

var identifierConditionTable = conditionTable{
	"letter_": util.IsLetter_,
	"digit":   util.IsDigit,
}

func identifierConstructor(s string, l, sc, ec int, _ state, _ interface{}) (common.Token, error) {
	if tokenType, ok := keywords[s]; ok {
		return common.NewToken(tokenType, s, nil, l, sc, ec), nil
	}
	return common.NewToken(common.IDENTIFIER, s, nil, l, sc, ec), nil
}

var identifierScanner *Scanner

func init() {
	identifierScanner = newIdentifierScanner()
}

func IdentifierScanner() *Scanner {
	return identifierScanner
}

func newIdentifierScanner() *Scanner {
	return NewScannerBuilder("Identifier").
		StateTable(identifierStateTable).
		ConditionTable(identifierConditionTable).
		TokenConstructor(identifierConstructor).
		StartState("A").
		EndState([]state{"B"}).
		Build()
}

var keywords = map[string]common.TokenType{
	"auto":       common.AUTO,
	"break":      common.BREAK,
	"case":       common.CASE,
	"char":       common.CHAR,
	"const":      common.CONST,
	"continue":   common.CONTINUE,
	"default":    common.DEFAULT,
	"do":         common.DO,
	"double":     common.DOUBLE,
	"else":       common.ELSE,
	"enum":       common.ENUM,
	"extern":     common.EXTERN,
	"float":      common.FLOAT,
	"for":        common.FOR,
	"goto":       common.GOTO,
	"if":         common.IF,
	"int":        common.INT,
	"long":       common.LONG,
	"register":   common.REGISTER,
	"restrict":   common.RESTRICT,
	"return":     common.RETURN,
	"short":      common.SHORT,
	"signed":     common.SIGNED,
	"sizeof":     common.SIZEOF,
	"static":     common.STATIC,
	"struct":     common.STRUCT,
	"switch":     common.SWITCH,
	"typedef":    common.TYPEDEF,
	"union":      common.UNION,
	"unsigned":   common.UNSIGNED,
	"void":       common.VOID,
	"volatile":   common.VOLATILE,
	"while":      common.WHILE,
	"_Bool":      common.BOOL,
	"_Complex":   common.COMPLEX,
	"_Imaginary": common.IMAGINARY,
}
