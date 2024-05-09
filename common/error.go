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
	Stage                  CvmStage
	ErrType                ErrType
	Line                   *int
	StartColumn, EndColumn *int
	CustomMessage          string
}

func NewVmError(typ ErrType, message string, a ...any) CvmError {
	return CvmError{
		Stage:         "VM",
		ErrType:       typ,
		CustomMessage: fmt.Sprintf(message, a...),
	}
}

func (e CvmError) Error() string {
	if e.Line == nil && e.StartColumn == nil {
		return fmt.Sprintf("Stage %s - %s %s", e.Stage, e.ErrType, e.CustomMessage)
	}
	if e.Line != nil && e.StartColumn == nil {
		return fmt.Sprintf("Stage %s - %s at Line %d - %s", e.Stage, e.ErrType, *e.Line, e.CustomMessage)
	}
	return fmt.Sprintf("Stage %s - %s from %d:%d to %d:%d - %s", e.Stage, e.ErrType, *e.Line, *e.StartColumn, *e.Line, *e.EndColumn, e.CustomMessage)
}
