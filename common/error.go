package common

import (
	"fmt"
	"github.com/hyphennn/glambda/gslice"
	"shinya.click/cvm/entity"
	"strings"
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
	return strings.Join(gslice.Map(e.Messages, func(m *CvmErrorMessages) string {
		return fmt.Sprintf("%d:%d: %s", m.SourcePos.Line, m.SourcePos.Column, m.CustomMessage)
	}), "\n")
}

func NewCvmError(messages ...*CvmErrorMessages) *CvmError {
	return &CvmError{
		Messages: messages,
	}
}
