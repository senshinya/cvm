package parser

import (
	"fmt"
	"shinya.click/cvm/common"
)

const (
	ErrInvalidSpecifiers common.ErrType = "ErrInvalidSpecifiers"
)

func NewParserError(typ common.ErrType, sourceRange common.SourceRange, message string, a ...any) common.CvmError {
	return common.CvmError{
		Stage:         "PARSER",
		ErrType:       typ,
		SourceRange:   &sourceRange,
		CustomMessage: fmt.Sprintf(message, a...),
	}
}
