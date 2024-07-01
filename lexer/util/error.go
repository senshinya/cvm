package util

import (
	"fmt"
	"shinya.click/cvm/common"
)

const (
	ErrUnidentifiableToken common.ErrType = "ErrUnidentifiableToken"
	ErrInvalidCharacter    common.ErrType = "ErrInvalidCharacter"
)

func NewLexerError(typ common.ErrType, l, sc, ec int, message string, a ...any) common.CvmError {
	return common.CvmError{
		Stage:   "LEXER",
		ErrType: typ,
		SourceRange: &common.SourceRange{
			SourceStart: common.SourcePos{Line: l, Column: sc},
			SourceEnd:   common.SourcePos{Line: l, Column: ec},
		},
		CustomMessage: fmt.Sprintf(message, a...),
	}
}
