package common

import "fmt"

type TokenType string

const (
	// 字面量
	IDENTIFIER        = "IDENTIFIER"
	STRING            = "STRING"
	CHARACTER         = "CHARACTER"
	INTEGER_CONSTANT  = "INTEGER_CONSTANT"
	FLOATING_CONSTANT = "FLOATING_CONSTANT"

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

	// 运算符
	LEFT_BRACKETS     = "LEFT_BRACKETS"
	RIGHT_BRACKETS    = "RIGHT_BRACKETS"
	LEFT_PARENTHESES  = "LEFT_PARENTHESES"
	RIGHT_PARENTHESES = "RIGHT_PARENTHESES"
	LEFT_BRACES       = "LEFT_BRACES"
	RIGHT_BRACES      = "RIGHT_BRACES"
	PERIOD            = "PERIOD"
	ARROW             = "ARROW"
	PLUS_PLUS         = "PLUS_PLUS"
	MINUS_MINUS       = "MINUS_MINUS"
	AND               = "AND"
	ASTERISK          = "ASTERISK"
	PLUS              = "PLUS"
	MINUS             = "MINUS"
	TILDE             = "TILDE"
	EXCLAMATION       = "EXCLAMATION"
	SLASH             = "SLASH"
	PERCENT           = "PERCENT"
	LEFT_SHIFT        = "LEFT_SHIFT"
	RIGHT_SHIFT       = "RIGHT_SHIFT"
	LESS              = "LESS"
	GREATER           = "GREATER"
	LESS_EQUAL        = "LESS_EQUAL"
	GREATER_EQUAL     = "GREATER_EQUAL"
	EQUAL_EQUAL       = "EQUAL_EQUAL"
	NOT_EQUAL         = "NOT_EQUAL"
	XOR               = "XOR"
	OR                = "OR"
	AND_AND           = "AND_AND"
	OR_OR             = "OR_OR"
	QUESTION          = "QUESTION"
	COLON             = "COLON"
	SEMICOLON         = "SEMICOLON"
	VARIADIC          = "VARIADIC"
	EQUAL             = "EQUAL"
	MULTIPLY_EQUAL    = "MULTIPLY_EQUAL"
	DIVIDE_EQUAL      = "DIVIDE_EQUAL"
	MOD_EQUAL         = "MOD_EQUAL"
	PLUS_EQUAL        = "PLUS_EQUAL"
	MINUS_EQUAL       = "MINUS_EQUAL"
	LEFT_SHIFT_EQUAL  = "LEFT_SHIFT_EQUAL"
	RIGHT_SHIFT_EQUAL = "RIGHT_SHIFT_EQUAL"
	AND_EQUAL         = "AND_EQUAL"
	XOR_EQUAL         = "XOR_EQUAL"
	OR_EQUAL          = "OR_EQUAL"
	COMMA             = "COMMA"
	SHARP             = "SHARP"
	SHARP_SHARP       = "SHARP_SHARP"

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
