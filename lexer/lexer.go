package lexer

import (
	"shinya.click/cvm/common"
	"shinya.click/cvm/lexer/util"
)

type Lexer struct {
	source  string
	tokens  []common.Token
	start   int
	current int
	sColumn int
	cColumn int
	line    int
}

func NewLexer(source string) *Lexer {
	return &Lexer{source: source, line: 1}
}

func (l *Lexer) ScanTokens() ([]common.Token, error) {
	for !l.isAtEnd() {
		if l.currentEmpty() {
			l.moveNext()
			continue
		}
		l.start = l.current
		l.sColumn = l.cColumn
		if err := l.scanToken(); err != nil {
			return nil, err
		}
	}

	l.tokens = append(l.tokens, common.NewToken(common.EOF, "", nil, l.line, l.cColumn, l.cColumn))
	return l.tokens, nil
}

func (l *Lexer) isAtEnd() bool {
	return l.current >= len(l.source)
}

func (l *Lexer) scanToken() error {
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
		return util.NewLexerError(util.ErrUnidentifiableToken, l.line, l.sColumn, l.cColumn, string(l.peek()))
	}
	token, err := sc.scan(l)
	if err != nil {
		return err
	}
	l.tokens = append(l.tokens, token)
	return nil
}

func (l *Lexer) peek() byte {
	return l.source[l.current]
}

func (l *Lexer) moveNext() {
	l.current++
	l.cColumn++
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
		l.cColumn = 0
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
