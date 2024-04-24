package common

import "fmt"

type TokenType string

const (
	// 字面量
	IDENTIFIER = "IDENTIFIER"
	STRING     = "STRING"
	CHARACTER  = "CHARACTER"
	NUMBER     = "NUMBER"

	// 关键字
	AUTO      = "AUTO"
	BREAK     = "BREAK"
	CASE      = "CASE"
	CHAR      = "CHAR"
	CONST     = "CONST"
	CONTINUE  = "CONTINUE"
	DEFAULT   = "DEFAULT"
	DO        = "DO"
	DOUBLE    = "DOUBLE"
	ELSE      = "ELSE"
	ENUM      = "ENUM"
	EXTERN    = "EXTERN"
	FLOAT     = "FLOAT"
	FOR       = "FOR"
	GOTO      = "GOTO"
	IF        = "IF"
	INT       = "INT"
	LONG      = "LONG"
	REGISTER  = "REGISTER"
	RESTRICT  = "RESTRICT"
	RETURN    = "RETURN"
	SHORT     = "SHORT"
	SIGNED    = "SIGNED"
	SIZEOF    = "SIZEOF"
	STATIC    = "STATIC"
	STRUCT    = "STRUCT"
	SWITCH    = "SWITCH"
	TYPEDEF   = "TYPEDEF"
	UNION     = "UNION"
	UNSIGNED  = "UNSIGNED"
	VOID      = "VOID"
	VOLATILE  = "VOLATILE"
	WHILE     = "WHILE"
	BOOL      = "_BOOL"
	COMPLEX   = "_COMPLEX"
	IMAGINARY = "_IMAGINARY"

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
