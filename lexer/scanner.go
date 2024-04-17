package lexer

import (
	"shinya.click/cvm/common"
	"shinya.click/cvm/util"
)

type Scanner struct {
	source  string
	tokens  []common.Token
	start   int
	current int
	line    int
}

func NewScanner(source string) *Scanner {
	return &Scanner{source: source, line: 1}
}

func (s *Scanner) ScanTokens() []common.Token {
	for !s.isAtEnd() {
		s.start = s.current
		s.scanToken()
	}

	s.tokens = append(s.tokens, common.NewToken(common.EOF, "", nil, s.line))
	return s.tokens
}

func (s *Scanner) isAtEnd() bool {
	return s.current >= len(s.source)
}

var singleCharToken = map[byte]common.TokenType{
	'(': common.LEFT_PAREN,
	')': common.RIGHT_PAREN,
	'{': common.LEFT_BRACE,
	'}': common.RIGHT_BRACE,
	',': common.COMMA,
	'.': common.DOT,
	'-': common.MINUS,
	'+': common.PLUS,
	';': common.SEMICOLON,
	'*': common.STAR,
}

func (s *Scanner) scanToken() {
	c := s.advance()
	switch c {
	case '(':
		s.addToken(common.LEFT_PAREN)
	case ')':
		s.addToken(common.RIGHT_PAREN)
	case '{':
		s.addToken(common.LEFT_BRACE)
	case '}':
		s.addToken(common.RIGHT_BRACE)
	case ',':
		s.addToken(common.COMMA)
	case '.':
		s.addToken(common.DOT)
	case '-':
		s.addToken(common.MINUS)
	case '+':
		s.addToken(common.PLUS)
	case ';':
		s.addToken(common.SEMICOLON)
	case '*':
		s.addToken(common.STAR)
	default:
		util.Error(s.line, "Unexpected character.")
	}
}

func (s *Scanner) advance() byte {
	s.current++
	return s.source[s.current-1]
}

func (s *Scanner) addToken(typ common.TokenType) {
	s.addTokenWithLiteral(typ, nil)
}

func (s *Scanner) addTokenWithLiteral(typ common.TokenType, literal any) {
	text := s.source[s.start:s.current]
	s.tokens = append(s.tokens, common.NewToken(typ, text, literal, s.line))
}
