package preprocessor

import "shinya.click/cvm/entity"

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
