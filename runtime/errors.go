package runtime

import (
	"fmt"
	"strings"

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
	Reason      string
	Function    string
	HasLocation bool
	FunctionID  int
	PC          int
	Opcode      bytecode.Opcode
	Stack       []string
	Cause       error
}

func (e *TrapError) Error() string {
	if e == nil {
		return "<nil>"
	}
	loc := ""
	if e.Function != "" || e.HasLocation {
		function := e.Function
		if function == "" {
			function = "fn"
		}
		loc = fmt.Sprintf(" in %s#%d pc=%d opcode=%s", function, e.FunctionID, e.PC, e.Opcode)
	}
	stack := ""
	if len(e.Stack) != 0 {
		stack = fmt.Sprintf(" stack=[%s]", strings.Join(e.Stack, " > "))
	}
	if e.Cause != nil {
		return fmt.Sprintf("runtime trap: %s%s%s: %v", e.Reason, loc, stack, e.Cause)
	}
	return fmt.Sprintf("runtime trap: %s%s%s", e.Reason, loc, stack)
}

func (e *TrapError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

type ExitStatus struct {
	Code       int
	skipAtexit bool
}
