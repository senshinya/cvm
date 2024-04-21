package common

import "fmt"

type TokenType string

const (
	// 单字符 token
	LEFT_PAREN  TokenType = "LEFT_PAREN"
	RIGHT_PAREN           = "RIGHT_PAREN"
	LEFT_BRACE            = "LEFT_BRACE"
	RIGHT_BRACE           = "RIGHT_BRACE"
	COMMA                 = "COMMA"
	DOT                   = "DOT"
	MINUS                 = "MINUS"
	PLUS                  = "PLUS"
	SEMICOLON             = "SEMICOLON"
	SLASH                 = "SLASH"
	STAR                  = "STAR"

	// 单/双字符 token
	BANG          = "BANG"
	BANG_EQUAL    = "BANG_EQUAL"
	EQUAL         = "EQUAL"
	EQUAL_EQUAL   = "EQUAL_EQUAL"
	GREATER       = "GREATER"
	GREATER_EQUAL = "GREATER_EQUAL"
	LESS          = "LESS"
	LESS_EQUAL    = "LESS_EQUAL"

	// 字面量
	IDENTIFIER = "IDENTIFIER"
	STRING     = "STRING"
	CHARACTER  = "CHARACTER"
	NUMBER     = "NUMBER"

	// 关键字
	AUTO       = "AUTO"
	BREAK      = "BREAK"
	CASE       = "CASE"
	CHAR       = "CHAR"
	CONST      = "CONST"
	CONTINUE   = "CONTINUE"
	DEFAULT    = "DEFAULT"
	DO         = "DO"
	DOUBLE     = "DOUBLE"
	ELSE       = "ELSE"
	ENUM       = "ENUM"
	EXTERN     = "EXTERN"
	FLOAT      = "FLOAT"
	FOR        = "FOR"
	GOTO       = "GOTO"
	IF         = "IF"
	INT        = "INT"
	LONG       = "LONG"
	REGISTER   = "REGISTER"
	RESTRICT   = "RESTRICT"
	RETURN     = "RETURN"
	SHORT      = "SHORT"
	SIGNED     = "SIGNED"
	SIZEOF     = "SIZEOF"
	STATIC     = "STATIC"
	STRUCT     = "STRUCT"
	SWITCH     = "SWITCH"
	TYPEDEF    = "TYPEDEF"
	UNION      = "UNION"
	UNSIGNED   = "UNSIGNED"
	VOID       = "VOID"
	VOLATILE   = "VOLATILE"
	WHILE      = "WHILE"
	_BOOL      = "_BOOL"
	_COMPLEX   = "_COMPLEX"
	_IMAGINARY = "_IMAGINARY"

	EOF = "EOF"
)

type Token struct {
	Typ     TokenType
	Lexeme  string
	Literal any
	Line    int
}

func NewToken(typ TokenType, lexeme string, literal any, line int) Token {
	return Token{
		Typ:     typ,
		Lexeme:  lexeme,
		Literal: literal,
		Line:    line,
	}
}

func (t Token) string() string {
	return fmt.Sprintf("%s %s %v", t.Typ, t.Lexeme, t.Literal)
}
