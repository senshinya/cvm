package entity

import "fmt"

type TokenType string

const (
	// 字面量
	IDENTIFIER        TokenType = "IDENTIFIER"
	STRING            TokenType = "STRING"
	CHARACTER         TokenType = "CHARACTER"
	INTEGER_CONSTANT  TokenType = "INTEGER_CONSTANT"
	FLOATING_CONSTANT TokenType = "FLOATING_CONSTANT"

	// 关键字
	AUTO      TokenType = "AUTO"
	BREAK     TokenType = "BREAK"
	CASE      TokenType = "CASE"
	CHAR      TokenType = "CHAR"
	CONST     TokenType = "CONST"
	CONTINUE  TokenType = "CONTINUE"
	DEFAULT   TokenType = "DEFAULT"
	DO        TokenType = "DO"
	DOUBLE    TokenType = "DOUBLE"
	ELSE      TokenType = "ELSE"
	ENUM      TokenType = "ENUM"
	EXTERN    TokenType = "EXTERN"
	FLOAT     TokenType = "FLOAT"
	FOR       TokenType = "FOR"
	GOTO      TokenType = "GOTO"
	IF        TokenType = "IF"
	INLINE    TokenType = "INLINE"
	INT       TokenType = "INT"
	LONG      TokenType = "LONG"
	REGISTER  TokenType = "REGISTER"
	RESTRICT  TokenType = "RESTRICT"
	RETURN    TokenType = "RETURN"
	SHORT     TokenType = "SHORT"
	SIGNED    TokenType = "SIGNED"
	SIZEOF    TokenType = "SIZEOF"
	STATIC    TokenType = "STATIC"
	STRUCT    TokenType = "STRUCT"
	SWITCH    TokenType = "SWITCH"
	TYPEDEF   TokenType = "TYPEDEF"
	UNION     TokenType = "UNION"
	UNSIGNED  TokenType = "UNSIGNED"
	VOID      TokenType = "VOID"
	VOLATILE  TokenType = "VOLATILE"
	WHILE     TokenType = "WHILE"
	BOOL      TokenType = "_BOOL"
	COMPLEX   TokenType = "_COMPLEX"
	IMAGINARY TokenType = "_IMAGINARY"

	// 运算符
	LEFT_BRACKETS     TokenType = "LEFT_BRACKETS"
	RIGHT_BRACKETS    TokenType = "RIGHT_BRACKETS"
	LEFT_PARENTHESES  TokenType = "LEFT_PARENTHESES"
	RIGHT_PARENTHESES TokenType = "RIGHT_PARENTHESES"
	LEFT_BRACES       TokenType = "LEFT_BRACES"
	RIGHT_BRACES      TokenType = "RIGHT_BRACES"
	PERIOD            TokenType = "PERIOD"
	ARROW             TokenType = "ARROW"
	PLUS_PLUS         TokenType = "PLUS_PLUS"
	MINUS_MINUS       TokenType = "MINUS_MINUS"
	AND               TokenType = "AND"
	ASTERISK          TokenType = "ASTERISK"
	PLUS              TokenType = "PLUS"
	MINUS             TokenType = "MINUS"
	TILDE             TokenType = "TILDE"
	EXCLAMATION       TokenType = "EXCLAMATION"
	SLASH             TokenType = "SLASH"
	PERCENT           TokenType = "PERCENT"
	LEFT_SHIFT        TokenType = "LEFT_SHIFT"
	RIGHT_SHIFT       TokenType = "RIGHT_SHIFT"
	LESS              TokenType = "LESS"
	GREATER           TokenType = "GREATER"
	LESS_EQUAL        TokenType = "LESS_EQUAL"
	GREATER_EQUAL     TokenType = "GREATER_EQUAL"
	EQUAL_EQUAL       TokenType = "EQUAL_EQUAL"
	NOT_EQUAL         TokenType = "NOT_EQUAL"
	XOR               TokenType = "XOR"
	OR                TokenType = "OR"
	AND_AND           TokenType = "AND_AND"
	OR_OR             TokenType = "OR_OR"
	QUESTION          TokenType = "QUESTION"
	COLON             TokenType = "COLON"
	SEMICOLON         TokenType = "SEMICOLON"
	VARIADIC          TokenType = "VARIADIC"
	EQUAL             TokenType = "EQUAL"
	MULTIPLY_EQUAL    TokenType = "MULTIPLY_EQUAL"
	DIVIDE_EQUAL      TokenType = "DIVIDE_EQUAL"
	MOD_EQUAL         TokenType = "MOD_EQUAL"
	PLUS_EQUAL        TokenType = "PLUS_EQUAL"
	MINUS_EQUAL       TokenType = "MINUS_EQUAL"
	LEFT_SHIFT_EQUAL  TokenType = "LEFT_SHIFT_EQUAL"
	RIGHT_SHIFT_EQUAL TokenType = "RIGHT_SHIFT_EQUAL"
	AND_EQUAL         TokenType = "AND_EQUAL"
	XOR_EQUAL         TokenType = "XOR_EQUAL"
	OR_EQUAL          TokenType = "OR_EQUAL"
	COMMA             TokenType = "COMMA"
	SHARP             TokenType = "SHARP"
	SHARP_SHARP       TokenType = "SHARP_SHARP"

	EOF TokenType = "EOF"
)

func IsTerminalSymbol(s string) bool {
	_, ok := terminals[s]
	return ok
}

var terminals = map[string]struct{}{
	"IDENTIFIER":        {},
	"STRING":            {},
	"CHARACTER":         {},
	"INTEGER_CONSTANT":  {},
	"FLOATING_CONSTANT": {},
	"AUTO":              {},
	"BREAK":             {},
	"CASE":              {},
	"CHAR":              {},
	"CONST":             {},
	"CONTINUE":          {},
	"DEFAULT":           {},
	"DO":                {},
	"DOUBLE":            {},
	"ELSE":              {},
	"ENUM":              {},
	"EXTERN":            {},
	"FLOAT":             {},
	"FOR":               {},
	"GOTO":              {},
	"IF":                {},
	"INLINE":            {},
	"INT":               {},
	"LONG":              {},
	"REGISTER":          {},
	"RESTRICT":          {},
	"RETURN":            {},
	"SHORT":             {},
	"SIGNED":            {},
	"SIZEOF":            {},
	"STATIC":            {},
	"STRUCT":            {},
	"SWITCH":            {},
	"TYPEDEF":           {},
	"UNION":             {},
	"UNSIGNED":          {},
	"VOID":              {},
	"VOLATILE":          {},
	"WHILE":             {},
	"BOOL":              {},
	"COMPLEX":           {},
	"IMAGINARY":         {},
	"LEFT_BRACKETS":     {},
	"RIGHT_BRACKETS":    {},
	"LEFT_PARENTHESES":  {},
	"RIGHT_PARENTHESES": {},
	"LEFT_BRACES":       {},
	"RIGHT_BRACES":      {},
	"PERIOD":            {},
	"ARROW":             {},
	"PLUS_PLUS":         {},
	"MINUS_MINUS":       {},
	"AND":               {},
	"ASTERISK":          {},
	"PLUS":              {},
	"MINUS":             {},
	"TILDE":             {},
	"EXCLAMATION":       {},
	"SLASH":             {},
	"PERCENT":           {},
	"LEFT_SHIFT":        {},
	"RIGHT_SHIFT":       {},
	"LESS":              {},
	"GREATER":           {},
	"LESS_EQUAL":        {},
	"GREATER_EQUAL":     {},
	"EQUAL_EQUAL":       {},
	"NOT_EQUAL":         {},
	"XOR":               {},
	"OR":                {},
	"AND_AND":           {},
	"OR_OR":             {},
	"QUESTION":          {},
	"COLON":             {},
	"SEMICOLON":         {},
	"VARIADIC":          {},
	"EQUAL":             {},
	"MULTIPLY_EQUAL":    {},
	"DIVIDE_EQUAL":      {},
	"MOD_EQUAL":         {},
	"PLUS_EQUAL":        {},
	"MINUS_EQUAL":       {},
	"LEFT_SHIFT_EQUAL":  {},
	"RIGHT_SHIFT_EQUAL": {},
	"AND_EQUAL":         {},
	"XOR_EQUAL":         {},
	"OR_EQUAL":          {},
	"COMMA":             {},
	"SHARP":             {},
	"SHARP_SHARP":       {},
	"TYPE_NAME":         {},
	"EOF":               {},
}

type Token struct {
	Typ     TokenType
	Lexeme  string
	Literal any
	SourceRange
}

func NewToken(typ TokenType, lexeme string, literal any, line, sc, ec int) Token {
	return Token{
		Typ:     typ,
		Lexeme:  lexeme,
		Literal: literal,
		SourceRange: SourceRange{
			SourceStart: SourcePos{Line: line, Column: sc},
			SourceEnd:   SourcePos{Line: line, Column: ec},
		},
	}
}

func (t Token) String() string {
	return fmt.Sprintf("<%s %s>", t.Typ, t.Lexeme)
}

type SourcePos struct {
	Line   int
	Column int
}

type SourceRange struct {
	SourceStart SourcePos
	SourceEnd   SourcePos
}

func (r SourceRange) GetSourceRange() SourceRange {
	return r
}
