package common

import (
	"fmt"
	"shinya.click/cvm/entity"
)

type CvmStage string

type ErrType string

const (
	ErrOpenFile            ErrType = "ErrOpenFile"
	ErrReadFile            ErrType = "ErrReadFile"
	ErrInvalidStateMachine ErrType = "ErrInvalidStateMachine"

	ErrUnidentifiableToken ErrType = "ErrUnidentifiableToken"
	ErrInvalidCharacter    ErrType = "ErrInvalidCharacter"
)

type CvmError struct {
	Stage         CvmStage
	ErrType       ErrType
	SourceRange   *entity.SourceRange
	CustomMessage string
}

func NewInitError(typ ErrType, message string, a ...any) CvmError {
	return CvmError{
		Stage:         "Init",
		ErrType:       typ,
		CustomMessage: fmt.Sprintf(message, a...),
	}
}

func (e CvmError) Error() string {
	if e.SourceRange == nil {
		return fmt.Sprintf("Stage %s - %s %s", e.Stage, e.ErrType, e.CustomMessage)
	}
	return fmt.Sprintf("Stage %s - %s from %d:%d to %d:%d - %s", e.Stage, e.ErrType, e.SourceRange.SourceStart.Line, e.SourceRange.SourceStart.Column, e.SourceRange.SourceEnd.Line, e.SourceRange.SourceEnd.Column, e.CustomMessage)
}

func NewLexerError(typ ErrType, l, sc, ec int, message string, a ...any) CvmError {
	return CvmError{
		Stage:   "LEXER",
		ErrType: typ,
		SourceRange: &entity.SourceRange{
			SourceStart: entity.SourcePos{Line: l, Column: sc},
			SourceEnd:   entity.SourcePos{Line: l, Column: ec},
		},
		CustomMessage: fmt.Sprintf(message, a...),
	}
}

const (
	ErrSymbolRedefinition   ErrType = "ErrSymbolRedefinition"
	ErrSymbolKindMismatch   ErrType = "ErrSymbolKindMismatch"
	ErrSymbolNotFound       ErrType = "ErrSymbolNotFound"
	ErrInvalidTypeSpecifier ErrType = "ErrInvalidTypeSpecifier"
)

func NewParserError(typ ErrType, sourceRange entity.SourceRange, message string, a ...any) CvmError {
	return CvmError{
		Stage:         "PARSER",
		ErrType:       typ,
		SourceRange:   &sourceRange,
		CustomMessage: fmt.Sprintf(message, a...),
	}
}
