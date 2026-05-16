package runtime

import (
	"fmt"

	"shinya.click/cvm/bytecode"
)

type LoadError struct {
	Reason string
	Cause  error
}

func (e *LoadError) Error() string {
	if e == nil {
		return "<nil>"
	}
	if e.Cause != nil {
		return fmt.Sprintf("runtime load error: %s: %v", e.Reason, e.Cause)
	}
	return fmt.Sprintf("runtime load error: %s", e.Reason)
}

func (e *LoadError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

type TrapError struct {
	Reason     string
	Function   string
	FunctionID int
	PC         int
	Opcode     bytecode.Opcode
	Stack      []string
	Cause      error
}

func (e *TrapError) Error() string {
	if e == nil {
		return "<nil>"
	}
	loc := ""
	if e.Function != "" {
		loc = fmt.Sprintf(" in %s#%d pc=%d opcode=%s", e.Function, e.FunctionID, e.PC, e.Opcode)
	}
	if e.Cause != nil {
		return fmt.Sprintf("runtime trap: %s%s: %v", e.Reason, loc, e.Cause)
	}
	return fmt.Sprintf("runtime trap: %s%s", e.Reason, loc)
}

func (e *TrapError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

type ExitStatus struct {
	Code int
}
