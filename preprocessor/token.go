package preprocessor

import (
	"strings"

	"shinya.click/cvm/entity"
	"shinya.click/cvm/lexer"
)

type PPTokenKind int

const (
	PPIdentifier PPTokenKind = iota
	PPNumber
	PPString
	PPCharacter
	PPPunctuator
	PPHeaderName
	PPPadding
	PPNewline
	PPEOF
)

type PPToken struct {
	Kind          PPTokenKind
	Lexeme        string
	Location      entity.SourcePos
	StartOfLine   bool
	LeadingSpace  bool
	DisableExpand bool
	NeedsCleaning bool
}

type TokenSource interface {
	Lex() (PPToken, error)
}

func convertToParserTokens(tokens []PPToken, sm *SourceManager) ([]entity.Token, error) {
	var out []entity.Token
	var last entity.SourcePos
	for _, tok := range tokens {
		if tok.Kind == PPNewline || tok.Kind == PPPadding {
			continue
		}
		last = tok.Location
		if tok.Kind == PPEOF {
			break
		}
		converted, err := convertOneToken(tok)
		if err != nil {
			return nil, err
		}
		converted.SourceStart = tok.Location
		converted.SourceEnd = tok.Location
		out = append(out, converted)
	}
	out = append(out, entity.Token{
		Typ:    entity.EOF,
		Lexeme: "",
		SourceRange: entity.SourceRange{
			SourceStart: last,
			SourceEnd:   last,
		},
	})
	return out, nil
}

func convertOneToken(tok PPToken) (entity.Token, error) {
	switch tok.Kind {
	case PPIdentifier:
		if typ, ok := keywordTokens[tok.Lexeme]; ok {
			return entity.Token{Typ: typ, Lexeme: tok.Lexeme}, nil
		}
		return entity.Token{Typ: entity.IDENTIFIER, Lexeme: tok.Lexeme}, nil
	case PPPunctuator:
		if isHash(tok) || isPaste(tok) {
			return entity.Token{}, ppError(tok.Location, "unexpected preprocessing token %q", tok.Lexeme)
		}
		if typ, ok := punctuatorTokens[tok.Lexeme]; ok {
			return entity.Token{Typ: typ, Lexeme: tok.Lexeme}, nil
		}
		return entity.Token{}, ppError(tok.Location, "unknown punctuator %q", tok.Lexeme)
	case PPNumber, PPString, PPCharacter:
		lexeme := tok.Lexeme
		if (tok.Kind == PPString || tok.Kind == PPCharacter) && strings.HasPrefix(lexeme, "L") {
			lexeme = strings.TrimPrefix(lexeme, "L")
		}
		scanned, err := lexer.NewLexer(lexeme).ScanTokens()
		if err != nil {
			return entity.Token{}, err
		}
		if len(scanned) == 0 {
			return entity.Token{}, ppError(tok.Location, "empty token conversion")
		}
		return entity.Token{Typ: scanned[0].Typ, Lexeme: lexeme, Literal: scanned[0].Literal}, nil
	case PPHeaderName:
		return entity.Token{}, ppError(tok.Location, "unexpected header name %q", tok.Lexeme)
	default:
		return entity.Token{}, ppError(tok.Location, "unexpected preprocessing token %q", tok.Lexeme)
	}
}

var keywordTokens = map[string]entity.TokenType{
	"auto":           entity.AUTO,
	"break":          entity.BREAK,
	"case":           entity.CASE,
	"char":           entity.CHAR,
	"const":          entity.CONST,
	"continue":       entity.CONTINUE,
	"default":        entity.DEFAULT,
	"do":             entity.DO,
	"double":         entity.DOUBLE,
	"else":           entity.ELSE,
	"enum":           entity.ENUM,
	"extern":         entity.EXTERN,
	"float":          entity.FLOAT,
	"for":            entity.FOR,
	"goto":           entity.GOTO,
	"if":             entity.IF,
	"inline":         entity.INLINE,
	"int":            entity.INT,
	"long":           entity.LONG,
	"register":       entity.REGISTER,
	"restrict":       entity.RESTRICT,
	"__restrict":     entity.RESTRICT,
	"__restrict__":   entity.RESTRICT,
	"return":         entity.RETURN,
	"short":          entity.SHORT,
	"signed":         entity.SIGNED,
	"sizeof":         entity.SIZEOF,
	"static":         entity.STATIC,
	"struct":         entity.STRUCT,
	"switch":         entity.SWITCH,
	"typedef":        entity.TYPEDEF,
	"__typeof":       entity.TYPEOF,
	"__typeof__":     entity.TYPEOF,
	"typeof":         entity.TYPEOF,
	"union":          entity.UNION,
	"unsigned":       entity.UNSIGNED,
	"void":           entity.VOID,
	"volatile":       entity.VOLATILE,
	"__volatile":     entity.VOLATILE,
	"__volatile__":   entity.VOLATILE,
	"while":          entity.WHILE,
	"_Bool":          entity.BOOL,
	"_Complex":       entity.COMPLEX,
	"__complex__":    entity.COMPLEX,
	"__complex":      entity.COMPLEX,
	"_Imaginary":     entity.IMAGINARY,
	"_Static_assert": entity.STATIC_ASSERT,
}

var punctuatorTokens = map[string]entity.TokenType{
	"[":   entity.LEFT_BRACKETS,
	"<:":  entity.LEFT_BRACKETS,
	"]":   entity.RIGHT_BRACKETS,
	":>":  entity.RIGHT_BRACKETS,
	"(":   entity.LEFT_PARENTHESES,
	")":   entity.RIGHT_PARENTHESES,
	"{":   entity.LEFT_BRACES,
	"<%":  entity.LEFT_BRACES,
	"}":   entity.RIGHT_BRACES,
	"%>":  entity.RIGHT_BRACES,
	".":   entity.PERIOD,
	"->":  entity.ARROW,
	"++":  entity.PLUS_PLUS,
	"--":  entity.MINUS_MINUS,
	"&":   entity.AND,
	"*":   entity.ASTERISK,
	"+":   entity.PLUS,
	"-":   entity.MINUS,
	"~":   entity.TILDE,
	"!":   entity.EXCLAMATION,
	"/":   entity.SLASH,
	"%":   entity.PERCENT,
	"<<":  entity.LEFT_SHIFT,
	">>":  entity.RIGHT_SHIFT,
	"<":   entity.LESS,
	">":   entity.GREATER,
	"<=":  entity.LESS_EQUAL,
	">=":  entity.GREATER_EQUAL,
	"==":  entity.EQUAL_EQUAL,
	"!=":  entity.NOT_EQUAL,
	"^":   entity.XOR,
	"|":   entity.OR,
	"&&":  entity.AND_AND,
	"||":  entity.OR_OR,
	"?":   entity.QUESTION,
	":":   entity.COLON,
	";":   entity.SEMICOLON,
	"...": entity.VARIADIC,
	"=":   entity.EQUAL,
	"*=":  entity.MULTIPLY_EQUAL,
	"/=":  entity.DIVIDE_EQUAL,
	"%=":  entity.MOD_EQUAL,
	"+=":  entity.PLUS_EQUAL,
	"-=":  entity.MINUS_EQUAL,
	"<<=": entity.LEFT_SHIFT_EQUAL,
	">>=": entity.RIGHT_SHIFT_EQUAL,
	"&=":  entity.AND_EQUAL,
	"^=":  entity.XOR_EQUAL,
	"|=":  entity.OR_EQUAL,
	",":   entity.COMMA,
}
