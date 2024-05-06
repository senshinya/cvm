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
	var sc *Scanner
	switch {
	case l.peek() == '\'':
		sc = CharacterLiteralScanner()
	case l.peek() == '"':
		sc = StringLiteralScanner()
	case util.IsLetter_(l.peek()):
		sc = IdentifierScanner()
	case l.peek() == '.' && util.IsDigit(l.lookForward(1)):
		sc = NumberLiteralScanner()
	case isPunctuatorPrefix(l.peek()):
		sc = PunctuatorScanner()
	case util.IsDigit(l.peek()):
		sc = NumberLiteralScanner()
	default:
		log.Panicf("Unrecognized token: %c", l.peek())
	}
	token := sc.scan(l)
	l.tokens = append(l.tokens, token)
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

func (l *Lexer) lookForward(i int) byte {
	forward := l.current + i
	if forward >= len(l.source) {
		return ' '
	}
	return l.source[forward]
}
