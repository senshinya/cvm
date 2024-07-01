package common

import "fmt"

type CvmStage string

type ErrType string

const (
	ErrOpenFile            ErrType = "ErrOpenFile"
	ErrReadFile            ErrType = "ErrReadFile"
	ErrInvalidStateMachine ErrType = "ErrInvalidStateMachine"
)

type CvmError struct {
	Stage         CvmStage
	ErrType       ErrType
	SourceRange   *SourceRange
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
