package runtime

import (
	"errors"
	"strings"
	"testing"

	"shinya.click/cvm/bytecode"
)

func TestLoadErrorWrapsCause(t *testing.T) {
	err := &LoadError{Reason: "missing entry", Cause: errors.New("entry=none")}
	if !strings.Contains(err.Error(), "missing entry") || !strings.Contains(err.Error(), "entry=none") {
		t.Fatalf("LoadError text = %q", err.Error())
	}
	if !errors.Is(err, err.Cause) {
		t.Fatalf("errors.Is did not unwrap cause")
	}
}

func TestTrapErrorIncludesExecutionContext(t *testing.T) {
	err := &TrapError{
		Reason:     "division by zero",
		Function:   "main",
		FunctionID: 0,
		PC:         3,
		Opcode:     bytecode.OpBinary,
	}
	got := err.Error()
	for _, want := range []string{"division by zero", "main", "pc=3", "OpBinary"} {
		if !strings.Contains(got, want) {
			t.Fatalf("TrapError missing %q: %s", want, got)
		}
	}
}

func TestTrapErrorIncludesUnnamedExecutionContext(t *testing.T) {
	err := &TrapError{
		Reason:     "invalid jump",
		FunctionID: 2,
		PC:         9,
		Opcode:     bytecode.OpJump,
	}
	got := err.Error()
	for _, want := range []string{"invalid jump", "fn#2", "pc=9", "OpJump"} {
		if !strings.Contains(got, want) {
			t.Fatalf("TrapError missing %q: %s", want, got)
		}
	}
}

func TestTrapErrorWrapsCause(t *testing.T) {
	err := &TrapError{
		Reason: "bad operand",
		Cause:  errors.New("operand=ptr"),
	}
	if !strings.Contains(err.Error(), "bad operand") || !strings.Contains(err.Error(), "operand=ptr") {
		t.Fatalf("TrapError text = %q", err.Error())
	}
	if !errors.Is(err, err.Cause) {
		t.Fatalf("errors.Is did not unwrap cause")
	}
}

func TestTrapErrorIncludesStack(t *testing.T) {
	err := &TrapError{
		Reason: "call failed",
		Stack:  []string{"main", "callee"},
	}
	got := err.Error()
	for _, want := range []string{"call failed", "stack=[main > callee]"} {
		if !strings.Contains(got, want) {
			t.Fatalf("TrapError missing %q: %s", want, got)
		}
	}
}

func TestExitStatusIsNormalResult(t *testing.T) {
	st := ExitStatus{Code: 7}
	if st.Code != 7 {
		t.Fatalf("exit code = %d, want 7", st.Code)
	}
}
