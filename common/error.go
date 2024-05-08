package common

type CvmStage string

type ErrType string

type CvmError struct {
	Stage                  CvmStage
	ErrType                ErrType
	Line                   int64
	StartColumn, EndColumn int64
	CustomMessage          string
}
