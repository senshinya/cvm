package runtime

import (
	"context"
	"fmt"
	"io"
	"os"

	"shinya.click/cvm/bytecode"
)

type ExternFunc func(context.Context, *ExternContext, []Value) (Value, *ExitStatus, error)

type ExternContext struct {
	Memory *Memory
	Stdout io.Writer
	Stderr io.Writer
}

type ExternRegistry struct {
	funcs  map[string]ExternFunc
	stdout io.Writer
	stderr io.Writer
}

func NewExternRegistry(stdout, stderr io.Writer) *ExternRegistry {
	if stdout == nil {
		stdout = os.Stdout
	}
	if stderr == nil {
		stderr = os.Stderr
	}
	return &ExternRegistry{funcs: make(map[string]ExternFunc), stdout: stdout, stderr: stderr}
}

func DefaultExternRegistry(stdout, stderr io.Writer) *ExternRegistry {
	r := NewExternRegistry(stdout, stderr)
	r.Register("exit", func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 1 {
			return Value{}, nil, fmt.Errorf("exit expects 1 argument")
		}
		code, err := args[0].ExitCode()
		if err != nil {
			return Value{}, nil, err
		}
		return Value{}, &ExitStatus{Code: code}, nil
	})
	r.Register("abort", func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		return Value{}, nil, &TrapError{Reason: "abort"}
	})
	r.Register("puts", func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 1 {
			return Value{}, nil, fmt.Errorf("puts expects 1 argument")
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("puts requires memory")
		}
		s, err := ec.Memory.ReadCString(args[0].Int)
		if err != nil {
			return Value{}, nil, err
		}
		if _, err := fmt.Fprintln(r.externStdout(ec), s); err != nil {
			return Value{}, nil, err
		}
		return IntValue(bytecode.TypeI32, int64(len(s)+1)), nil, nil
	})
	r.Register("fputs", func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 2 {
			return Value{}, nil, fmt.Errorf("fputs expects 2 arguments")
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("fputs requires memory")
		}
		s, err := ec.Memory.ReadCString(args[0].Int)
		if err != nil {
			return Value{}, nil, err
		}
		if _, err := fmt.Fprint(r.externStdout(ec), s); err != nil {
			return Value{}, nil, err
		}
		return IntValue(bytecode.TypeI32, int64(len(s))), nil, nil
	})
	return r
}

func (r *ExternRegistry) Register(name string, fn ExternFunc) {
	r.funcs[name] = fn
}

func (r *ExternRegistry) Lookup(name string) (ExternFunc, bool) {
	fn, ok := r.funcs[name]
	return fn, ok
}

func (r *ExternRegistry) LookupVariable(name string, mem *Memory) (uint64, bool) {
	switch name {
	case "stdout", "stderr":
		return mem.Alloc("extern:"+name, 8, 8, true, blockHostHandle), true
	default:
		return 0, false
	}
}

func (r *ExternRegistry) context(mem *Memory) *ExternContext {
	return &ExternContext{Memory: mem, Stdout: r.stdout, Stderr: r.stderr}
}

func (r *ExternRegistry) externStdout(ec *ExternContext) io.Writer {
	if ec != nil && ec.Stdout != nil {
		return ec.Stdout
	}
	return r.stdout
}
