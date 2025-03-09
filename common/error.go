package common

import (
	"shinya.click/cvm/entity"
)

type MessageLevel int

const (
	MessageLevelNote MessageLevel = iota
	MessageLevelError
)

type CvmErrorMessages struct {
	Level         MessageLevel
	SourcePos     entity.SourcePos
	CustomMessage string
}

func NewErrorMessage(pos entity.SourcePos, customMessage string) *CvmErrorMessages {
	return &CvmErrorMessages{
		Level:         MessageLevelError,
		SourcePos:     pos,
		CustomMessage: customMessage,
	}
}

func NewNoteMessage(pos entity.SourcePos, customMessage string) *CvmErrorMessages {
	return &CvmErrorMessages{
		Level:         MessageLevelNote,
		SourcePos:     pos,
		CustomMessage: customMessage,
	}
}

type CvmError struct {
	Messages []*CvmErrorMessages
}

func (e *CvmError) Error() string {
	// just for implementing interface
	return ""
}

func NewCvmError(messages ...*CvmErrorMessages) *CvmError {
	return &CvmError{
		Messages: messages,
	}
}
