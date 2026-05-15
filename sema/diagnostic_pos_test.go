package sema

import (
	"errors"
	"strings"
	"testing"

	"shinya.click/cvm/common"
	"shinya.click/cvm/entity"
)

func analyzeError(t *testing.T, src string) *common.CvmError {
	t.Helper()
	_, err := Analyze(parseCandidates(t, src))
	if err == nil {
		t.Fatal("expected sema error")
	}
	var cvmErr *common.CvmError
	if !errors.As(err, &cvmErr) {
		t.Fatalf("expected CvmError, got %T %v", err, err)
	}
	return cvmErr
}

func TestDuplicateFieldDiagnosticUsesFieldLocation(t *testing.T) {
	err := analyzeError(t, "struct S {\n int a;\n int a;\n};")
	msg := firstMessageContaining(t, err, "duplicate member name")
	if msg.SourcePos.Line != 3 || msg.SourcePos.Column != 6 {
		t.Fatalf("duplicate field pos = %d:%d, want 3:6", msg.SourcePos.Line, msg.SourcePos.Column)
	}
}

func TestEnumeratorRedefinitionNoteUsesEnumeratorLocation(t *testing.T) {
	err := analyzeError(t, "enum { E };\nint E;")
	if len(err.Messages) < 2 {
		t.Fatalf("expected note message, got %#v", err.Messages)
	}
	note := err.Messages[1]
	if note.SourcePos.Line != 1 || note.SourcePos.Column != 7 {
		t.Fatalf("enumerator note pos = %d:%d, want 1:7", note.SourcePos.Line, note.SourcePos.Column)
	}
}

func TestTagConflictNoteUsesPreviousTagLocation(t *testing.T) {
	scope := NewScope(ScopeFile, nil)
	tag1 := &TagID{}
	tag2 := &TagID{}
	pos1 := entity.SourcePos{Line: 4, Column: 3}
	pos2 := entity.SourcePos{Line: 9, Column: 5}
	if err := scope.InsertTagChecked("T", &TagInfo{Tag: tag1, T: &StructType{Tag: tag1}, Pos: pos1}, pos1); err != nil {
		t.Fatal(err)
	}
	err := scope.InsertTagChecked("T", &TagInfo{Tag: tag2, T: &StructType{Tag: tag2}, Pos: pos2}, pos2)
	if err == nil {
		t.Fatal("expected conflicting tag error")
	}
	var cvmErr *common.CvmError
	if !errors.As(err, &cvmErr) || len(cvmErr.Messages) < 2 {
		t.Fatalf("expected CvmError with note, got %T %v", err, err)
	}
	note := cvmErr.Messages[1]
	if note.SourcePos.Line != 4 || note.SourcePos.Column != 3 {
		t.Fatalf("tag conflict note pos = %d:%d, want 4:3", note.SourcePos.Line, note.SourcePos.Column)
	}
}

func firstMessageContaining(t *testing.T, err *common.CvmError, text string) *common.CvmErrorMessages {
	t.Helper()
	for _, msg := range err.Messages {
		if strings.Contains(msg.CustomMessage, text) {
			return msg
		}
	}
	t.Fatalf("message containing %q not found in %#v", text, err.Messages)
	return nil
}
