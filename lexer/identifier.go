package lexer

import (
	"shinya.click/cvm/common"
	"shinya.click/cvm/entity"
)

var identifierStateTable = stateTable{
	"A": []Edge{{"letter_", "B"}},
	"B": []Edge{{"letter_", "B"}, {"digit", "B"}},
}

var identifierConditionTable = conditionTable{
	"letter_": common.IsLetter_,
	"digit":   common.IsDigit,
}

func identifierConstructor(s string, l, sc, ec int, _ state, _ interface{}) (entity.Token, error) {
	if tokenType, ok := keywords[s]; ok {
		return entity.NewToken(tokenType, s, nil, l, sc, ec), nil
	}
	return entity.NewToken(entity.IDENTIFIER, s, nil, l, sc, ec), nil
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

var keywords = map[string]entity.TokenType{
	"auto":       entity.AUTO,
	"break":      entity.BREAK,
	"case":       entity.CASE,
	"char":       entity.CHAR,
	"const":      entity.CONST,
	"continue":   entity.CONTINUE,
	"default":    entity.DEFAULT,
	"do":         entity.DO,
	"double":     entity.DOUBLE,
	"else":       entity.ELSE,
	"enum":       entity.ENUM,
	"extern":     entity.EXTERN,
	"float":      entity.FLOAT,
	"for":        entity.FOR,
	"goto":       entity.GOTO,
	"if":         entity.IF,
	"inline":     entity.INLINE,
	"int":        entity.INT,
	"long":       entity.LONG,
	"register":   entity.REGISTER,
	"restrict":   entity.RESTRICT,
	"return":     entity.RETURN,
	"short":      entity.SHORT,
	"signed":     entity.SIGNED,
	"sizeof":     entity.SIZEOF,
	"static":     entity.STATIC,
	"struct":     entity.STRUCT,
	"switch":     entity.SWITCH,
	"typedef":    entity.TYPEDEF,
	"union":      entity.UNION,
	"unsigned":   entity.UNSIGNED,
	"void":       entity.VOID,
	"volatile":   entity.VOLATILE,
	"while":      entity.WHILE,
	"_Bool":      entity.BOOL,
	"_Complex":   entity.COMPLEX,
	"_Imaginary": entity.IMAGINARY,
}
