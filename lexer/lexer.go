package lexer

import (
	"log"
	"shinya.click/cvm/common"
	"shinya.click/cvm/lexer/util"
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
		if l.currentEmpty() {
			l.moveNext()
			continue
		}
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
	switch {
	case l.peek() == '\'':
		token := newCharacterLiteralScanner().scan(l)
		l.tokens = append(l.tokens, token)
	case l.peek() == '"':
		token := newStringLiteralScanner().scan(l)
		l.tokens = append(l.tokens, token)
	case util.IsLetter_(l.peek()):
		token := newIdentifierScanner().scan(l)
		l.tokens = append(l.tokens, token)
	case isPunctuatorPrefix(l.peek()):
		token := newPunctuatorScanner().scan(l)
		l.tokens = append(l.tokens, token)
	case util.IsDigit(l.peek()):
		token := newNumberLiteral().scan(l)
		l.tokens = append(l.tokens, token)
	default:
		log.Panicf("Unrecognized token: %c", l.peek())
	}
}

func (l *Lexer) peek() byte {
	return l.source[l.current]
}

func (l *Lexer) moveNext() {
	l.current++
}

func (l *Lexer) currentEmpty() bool {
	switch l.peek() {
	case ' ':
		fallthrough
	case '\r':
		fallthrough
	case '\t':
		return true
	case '\n':
		l.line++
		return true
	}
	return false
}
