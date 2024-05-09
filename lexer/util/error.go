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
		Stage:         "LEXER",
		ErrType:       typ,
		Line:          &l,
		StartColumn:   &sc,
		EndColumn:     &ec,
		CustomMessage: fmt.Sprintf(message, a...),
	}
}
