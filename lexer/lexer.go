package lexer

import (
	"log"
	"shinya.click/cvm/common"
)

type Lexer struct {
	source  string
	tokens  []common.Token
	start   int
	current int
	line    int
}

func NewLexer(source string) *Lexer {
	return &Lexer{source: source, line: 1}
}

func (l *Lexer) ScanTokens() []common.Token {
	for !l.isAtEnd() {
		l.start = l.current
		l.scanToken()
	}

	l.tokens = append(l.tokens, common.NewToken(common.EOF, "", nil, l.line))
	return l.tokens
}

func (l *Lexer) isAtEnd() bool {
	return l.current >= len(l.source)
}

func (l *Lexer) scanToken() {
	switch l.getCurrent() {
	case '\'':
		token := newCharacterLiteralScanner().scan(l)
		l.tokens = append(l.tokens, token)
	default:
		log.Panicf("Unrecognized token: %c", l.getCurrent())
	}
}

func (l *Lexer) getCurrent() byte {
	return l.source[l.current]
}

func (l *Lexer) moveNext() {
	l.current++
}

func (l *Lexer) addToken(typ common.TokenType) {
	l.addTokenWithLiteral(typ, nil)
}

func (l *Lexer) addTokenWithLiteral(typ common.TokenType, literal any) {
	text := l.source[l.start:l.current]
	l.tokens = append(l.tokens, common.NewToken(typ, text, literal, l.line))
}
