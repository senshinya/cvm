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
	funcs       map[string]ExternFunc
	stdout      io.Writer
	stderr      io.Writer
	hostWriters map[uint64]io.Writer
}

func NewExternRegistry(stdout, stderr io.Writer) *ExternRegistry {
	if stdout == nil {
		stdout = os.Stdout
	}
	if stderr == nil {
		stderr = os.Stderr
	}
	return &ExternRegistry{
		funcs:       make(map[string]ExternFunc),
		stdout:      stdout,
		stderr:      stderr,
		hostWriters: make(map[uint64]io.Writer),
	}
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
		w, ok := r.lookupHostWriter(args[1].Int)
		if !ok {
			return Value{}, nil, fmt.Errorf("unknown stream handle %#x", args[1].Int)
		}
		if _, err := fmt.Fprint(w, s); err != nil {
			return Value{}, nil, err
		}
		return IntValue(bytecode.TypeI32, int64(len(s))), nil, nil
	})
	r.Register("strcmp", func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 2 {
			return Value{}, nil, fmt.Errorf("strcmp expects 2 arguments")
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("strcmp requires memory")
		}
		left, err := ec.Memory.ReadCString(args[0].Int)
		if err != nil {
			return Value{}, nil, err
		}
		right, err := ec.Memory.ReadCString(args[1].Int)
		if err != nil {
			return Value{}, nil, err
		}
		return IntValue(bytecode.TypeI32, int64(strcmpResult(left, right))), nil, nil
	})
	r.Register("memcmp", func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 3 {
			return Value{}, nil, fmt.Errorf("memcmp expects 3 arguments")
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("memcmp requires memory")
		}
		n := unsignedInt(args[2])
		if n == 0 {
			return IntValue(bytecode.TypeI32, 0), nil, nil
		}
		if n > uint64(maxInt()) {
			return Value{}, nil, fmt.Errorf("memcmp size %d exceeds int range", n)
		}
		leftBlock, leftOff, err := ec.Memory.rangeAccess(args[0].Int, int64(n), false)
		if err != nil {
			return Value{}, nil, err
		}
		rightBlock, rightOff, err := ec.Memory.rangeAccess(args[1].Int, int64(n), false)
		if err != nil {
			return Value{}, nil, err
		}
		left := leftBlock.data[leftOff : leftOff+int(n)]
		right := rightBlock.data[rightOff : rightOff+int(n)]
		return IntValue(bytecode.TypeI32, int64(memcmpResult(left, right))), nil, nil
	})
	return r
}

func strcmpResult(left, right string) int {
	for i := 0; i < len(left) && i < len(right); i++ {
		if left[i] < right[i] {
			return -1
		}
		if left[i] > right[i] {
			return 1
		}
	}
	if len(left) < len(right) {
		return -1
	}
	if len(left) > len(right) {
		return 1
	}
	return 0
}

func memcmpResult(left, right []byte) int {
	for i := 0; i < len(left) && i < len(right); i++ {
		if left[i] < right[i] {
			return -1
		}
		if left[i] > right[i] {
			return 1
		}
	}
	if len(left) < len(right) {
		return -1
	}
	if len(left) > len(right) {
		return 1
	}
	return 0
}

func (r *ExternRegistry) Register(name string, fn ExternFunc) {
	if fn == nil {
		delete(r.funcs, name)
		return
	}
	r.funcs[name] = fn
}

func (r *ExternRegistry) Lookup(name string) (ExternFunc, bool) {
	fn, ok := r.funcs[name]
	return fn, ok && fn != nil
}

func (r *ExternRegistry) LookupVariable(name string, mem *Memory) (uint64, bool) {
	addr, ok, err := r.LookupVariableAddr(name, mem)
	if err != nil {
		return 0, false
	}
	return addr, ok
}

func (r *ExternRegistry) LookupVariableAddr(name string, mem *Memory) (uint64, bool, error) {
	switch name {
	case "stdout":
		addr, err := r.allocHostWriter(name, mem, r.stdout)
		return addr, true, err
	case "stderr":
		addr, err := r.allocHostWriter(name, mem, r.stderr)
		return addr, true, err
	default:
		return 0, false, nil
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

func (r *ExternRegistry) lookupHostWriter(addr uint64) (io.Writer, bool) {
	w, ok := r.hostWriters[addr]
	return w, ok
}

func (r *ExternRegistry) allocHostWriter(name string, mem *Memory, w io.Writer) (uint64, error) {
	if mem == nil {
		return 0, fmt.Errorf("memory is nil")
	}
	addr, err := mem.TryAlloc("extern:"+name, mem.target.PointerSize, mem.target.PointerAlign, false, blockHostHandle)
	if err != nil {
		return 0, err
	}
	if err := mem.WritePointer(addr, addr); err != nil {
		return 0, err
	}
	if b, _, err := mem.rangeAccess(addr, mem.target.PointerSize, false); err == nil {
		b.readonly = true
	}
	r.hostWriters[addr] = w
	return addr, nil
}
