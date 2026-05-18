package runtime

import (
	"context"
	"fmt"
	"io"
	"math"
	"math/cmplx"
	"os"
	"strconv"
	"strings"

	"shinya.click/cvm/bytecode"
)

type ExternFunc func(context.Context, *ExternContext, []Value) (Value, *ExitStatus, error)

type ExternContext struct {
	Memory *Memory
	Stdout io.Writer
	Stderr io.Writer
}

type ExternRegistry struct {
	funcs         map[string]ExternFunc
	stdout        io.Writer
	stderr        io.Writer
	hostWriters   map[uint64]io.Writer
	hostFDs       map[uint64]int32
	hostPushback  map[uint64][]byte
	hostEOF       map[uint64]bool
	stdinHandle   uint64
	staticStrings map[*Memory]map[string]uint64
	staticVars    map[*Memory]map[string]uint64
	strtokNext    map[*Memory]uint64
	randSeed      uint32
}

func NewExternRegistry(stdout, stderr io.Writer) *ExternRegistry {
	if stdout == nil {
		stdout = os.Stdout
	}
	if stderr == nil {
		stderr = os.Stderr
	}
	return &ExternRegistry{
		funcs:         make(map[string]ExternFunc),
		stdout:        stdout,
		stderr:        stderr,
		hostWriters:   make(map[uint64]io.Writer),
		hostFDs:       make(map[uint64]int32),
		hostPushback:  make(map[uint64][]byte),
		hostEOF:       make(map[uint64]bool),
		staticStrings: make(map[*Memory]map[string]uint64),
		staticVars:    make(map[*Memory]map[string]uint64),
		strtokNext:    make(map[*Memory]uint64),
		randSeed:      1,
	}
}

func DefaultExternRegistry(stdout, stderr io.Writer) *ExternRegistry {
	r := NewExternRegistry(stdout, stderr)
	r.Register("exit", exitExtern("exit"))
	r.Register("_Exit", exitExtern("_Exit"))
	r.Register("abort", abortExtern())
	r.Register("__builtin_abort", abortExtern())
	registerVaListExterns(r)
	r.Register("remove", removeExtern("remove"))
	r.Register("rename", renameExtern("rename"))
	r.Register("fopen", fopenExtern("fopen"))
	r.Register("freopen", freopenExtern("freopen", r))
	r.Register("tmpfile", tmpfileExtern("tmpfile"))
	r.Register("tmpnam", tmpnamExtern("tmpnam"))
	for _, name := range []string{"puts", "puts_unlocked"} {
		r.Register(name, putsExtern(name, r))
	}
	for _, name := range []string{"putchar", "putchar_unlocked"} {
		r.Register(name, putcharExtern(name, r))
	}
	for _, name := range []string{"getchar", "getchar_unlocked"} {
		r.Register(name, getcharExtern(name, r))
	}
	for _, name := range []string{"fputc", "fputc_unlocked", "putc", "putc_unlocked"} {
		r.Register(name, fputcExtern(name, r))
	}
	for _, name := range []string{"fgetc", "fgetc_unlocked", "getc", "getc_unlocked"} {
		r.Register(name, fgetcExtern(name, r))
	}
	r.Register("ungetc", ungetcExtern("ungetc", r))
	for _, name := range []string{"fgets", "fgets_unlocked"} {
		r.Register(name, fgetsExtern(name, r))
	}
	for _, name := range []string{"fputs", "fputs_unlocked"} {
		r.Register(name, fputsExtern(name, r))
	}
	r.Register("perror", perrorExtern("perror", r))
	for _, name := range []string{"fflush", "fflush_unlocked"} {
		r.Register(name, fflushExtern(name, r))
	}
	r.Register("fclose", fcloseExtern("fclose", r))
	for _, name := range []string{"fileno", "fileno_unlocked"} {
		r.Register(name, filenoExtern(name, r))
	}
	r.Register("fseek", fseekExtern("fseek", r))
	r.Register("ftell", ftellExtern("ftell", r))
	r.Register("rewind", rewindExtern("rewind", r))
	r.Register("fgetpos", fgetposExtern("fgetpos", r))
	r.Register("fsetpos", fsetposExtern("fsetpos", r))
	r.Register("setbuf", setbufExtern("setbuf", r))
	r.Register("setvbuf", setvbufExtern("setvbuf", r))
	r.Register("flockfile", streamLockExtern("flockfile", r, false))
	r.Register("ftrylockfile", streamLockExtern("ftrylockfile", r, true))
	r.Register("funlockfile", streamLockExtern("funlockfile", r, false))
	for _, name := range []string{"fwrite", "fwrite_unlocked"} {
		r.Register(name, fwriteExtern(name, r))
	}
	for _, name := range []string{"fread", "fread_unlocked"} {
		r.Register(name, freadExtern(name, r))
	}
	for _, name := range []string{"ferror", "ferror_unlocked", "feof", "feof_unlocked"} {
		r.Register(name, streamStatusExtern(name, r))
	}
	for _, name := range []string{"clearerr", "clearerr_unlocked"} {
		r.Register(name, clearerrExtern(name, r))
	}
	r.Register("abs", signedAbsExtern("abs", bytecode.TypeI32))
	r.Register("labs", signedAbsExtern("labs", bytecode.TypeI64))
	r.Register("llabs", signedAbsExtern("llabs", bytecode.TypeI64))
	r.Register("div", signedDivExtern("div", bytecode.TypeI32, 4, 4))
	r.Register("ldiv", signedDivExtern("ldiv", bytecode.TypeI64, 8, 8))
	r.Register("lldiv", signedDivExtern("lldiv", bytecode.TypeI64, 8, 8))
	r.Register("atoi", atoiExtern("atoi", bytecode.TypeI32))
	r.Register("atol", atoiExtern("atol", bytecode.TypeI64))
	r.Register("atoll", atoiExtern("atoll", bytecode.TypeI64))
	r.Register("atof", atofExtern("atof"))
	r.Register("strtol", strtoIntegerExtern("strtol", bytecode.TypeI64, true))
	r.Register("strtoul", strtoIntegerExtern("strtoul", bytecode.TypeU64, false))
	r.Register("strtoll", strtoIntegerExtern("strtoll", bytecode.TypeI64, true))
	r.Register("strtoull", strtoIntegerExtern("strtoull", bytecode.TypeU64, false))
	r.Register("strtod", strtoFloatExtern("strtod", bytecode.TypeF64))
	r.Register("strtof", strtoFloatExtern("strtof", bytecode.TypeF32))
	r.Register("strtold", strtoFloatExtern("strtold", bytecode.TypeFLong))
	r.Register("mblen", mblenExtern("mblen"))
	r.Register("mbtowc", mbtowcExtern("mbtowc"))
	r.Register("wctomb", wctombExtern("wctomb"))
	r.Register("mbstowcs", mbstowcsExtern("mbstowcs"))
	r.Register("wcstombs", wcstombsExtern("wcstombs"))
	r.Register("rand", randExtern("rand", r))
	r.Register("srand", srandExtern("srand", r))
	r.Register("getenv", getenvExtern("getenv"))
	r.Register("system", systemExtern("system"))
	r.Register("atexit", atexitExtern("atexit"))
	r.Register("setlocale", setlocaleExtern("setlocale", r))
	r.Register("clock", clockExtern("clock"))
	r.Register("difftime", difftimeExtern("difftime"))
	r.Register("time", timeExtern("time"))
	registerCtypeClassificationExterns(r)
	registerCtypeCaseExterns(r)
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
	r.Register("memcmp", memoryCompareExtern("memcmp"))
	r.Register("bcmp", memoryCompareExtern("bcmp"))
	registerAllocationExterns(r)
	registerMemoryExterns(r)
	registerOutputFormatExterns(r)
	r.Register("feclearexcept", func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 1 {
			return Value{}, nil, fmt.Errorf("feclearexcept expects 1 argument")
		}
		return IntValue(bytecode.TypeI32, 0), nil, nil
	})
	r.Register("fetestexcept", func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 1 {
			return Value{}, nil, fmt.Errorf("fetestexcept expects 1 argument")
		}
		return IntValue(bytecode.TypeI32, 0), nil, nil
	})
	registerMathExterns(r)
	return r
}

func registerVaListExterns(r *ExternRegistry) {
	r.Register("__builtin_va_start", func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 2 {
			return Value{}, nil, fmt.Errorf("__builtin_va_start expects 2 arguments")
		}
		return Value{}, nil, nil
	})
	r.Register("__builtin_va_end", func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 1 {
			return Value{}, nil, fmt.Errorf("__builtin_va_end expects 1 argument")
		}
		return Value{}, nil, nil
	})
}

func removeExtern(name string) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 1 {
			return Value{}, nil, fmt.Errorf("%s expects 1 argument", name)
		}
		if !isPointerType(args[0].Type) {
			return Value{}, nil, fmt.Errorf("%s expects path pointer", name)
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("%s requires memory", name)
		}
		if _, err := ec.Memory.ReadCString(args[0].Int); err != nil {
			return Value{}, nil, err
		}
		return IntValue(bytecode.TypeI32, -1), nil, nil
	}
}

func renameExtern(name string) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 2 {
			return Value{}, nil, fmt.Errorf("%s expects 2 arguments", name)
		}
		if !isPointerType(args[0].Type) || !isPointerType(args[1].Type) {
			return Value{}, nil, fmt.Errorf("%s expects path pointers", name)
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("%s requires memory", name)
		}
		if _, err := ec.Memory.ReadCString(args[0].Int); err != nil {
			return Value{}, nil, err
		}
		if _, err := ec.Memory.ReadCString(args[1].Int); err != nil {
			return Value{}, nil, err
		}
		return IntValue(bytecode.TypeI32, -1), nil, nil
	}
}

func fopenExtern(name string) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 2 {
			return Value{}, nil, fmt.Errorf("%s expects 2 arguments", name)
		}
		if !isPointerType(args[0].Type) || !isPointerType(args[1].Type) {
			return Value{}, nil, fmt.Errorf("%s expects path and mode pointers", name)
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("%s requires memory", name)
		}
		if _, err := ec.Memory.ReadCString(args[0].Int); err != nil {
			return Value{}, nil, err
		}
		if _, err := ec.Memory.ReadCString(args[1].Int); err != nil {
			return Value{}, nil, err
		}
		return PtrValue(0), nil, nil
	}
}

func freopenExtern(name string, r *ExternRegistry) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 3 {
			return Value{}, nil, fmt.Errorf("%s expects 3 arguments", name)
		}
		if !isPointerType(args[0].Type) || !isPointerType(args[1].Type) || !isPointerType(args[2].Type) {
			return Value{}, nil, fmt.Errorf("%s expects path, mode, and stream pointers", name)
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("%s requires memory", name)
		}
		if _, err := ec.Memory.ReadCString(args[0].Int); err != nil {
			return Value{}, nil, err
		}
		if _, err := ec.Memory.ReadCString(args[1].Int); err != nil {
			return Value{}, nil, err
		}
		if _, ok := r.lookupHostWriter(args[2].Int); !ok {
			return Value{}, nil, fmt.Errorf("unknown stream handle %#x", args[2].Int)
		}
		return PtrValue(0), nil, nil
	}
}

func tmpfileExtern(name string) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 0 {
			return Value{}, nil, fmt.Errorf("%s expects 0 arguments", name)
		}
		return PtrValue(0), nil, nil
	}
}

func tmpnamExtern(name string) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 1 {
			return Value{}, nil, fmt.Errorf("%s expects 1 argument", name)
		}
		if !isPointerType(args[0].Type) {
			return Value{}, nil, fmt.Errorf("%s expects buffer pointer", name)
		}
		if args[0].Int != 0 {
			if ec == nil || ec.Memory == nil {
				return Value{}, nil, fmt.Errorf("%s requires memory", name)
			}
			if err := writeMemoryByte(ec.Memory, args[0].Int, 0); err != nil {
				return Value{}, nil, err
			}
		}
		return PtrValue(0), nil, nil
	}
}

func putsExtern(name string, r *ExternRegistry) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 1 {
			return Value{}, nil, fmt.Errorf("%s expects 1 argument", name)
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("%s requires memory", name)
		}
		s, err := ec.Memory.ReadCString(args[0].Int)
		if err != nil {
			return Value{}, nil, err
		}
		if _, err := fmt.Fprintln(r.externStdout(ec), s); err != nil {
			return Value{}, nil, err
		}
		return IntValue(bytecode.TypeI32, int64(len(s)+1)), nil, nil
	}
}

func putcharExtern(name string, r *ExternRegistry) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 1 {
			return Value{}, nil, fmt.Errorf("%s expects 1 argument", name)
		}
		if !isIntegerLike(args[0].Type) {
			return Value{}, nil, fmt.Errorf("%s expects integer argument", name)
		}
		ch := byte(args[0].Int)
		if _, err := r.externStdout(ec).Write([]byte{ch}); err != nil {
			return Value{}, nil, err
		}
		return IntValue(bytecode.TypeI32, int64(ch)), nil, nil
	}
}

func getcharExtern(name string, r *ExternRegistry) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 0 {
			return Value{}, nil, fmt.Errorf("%s expects 0 arguments", name)
		}
		if ch, ok := r.readHostChar(r.stdinHandle); ok {
			return IntValue(bytecode.TypeI32, int64(ch)), nil, nil
		}
		if r.stdinHandle != 0 {
			r.hostEOF[r.stdinHandle] = true
		}
		return IntValue(bytecode.TypeI32, -1), nil, nil
	}
}

func fputcExtern(name string, r *ExternRegistry) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 2 {
			return Value{}, nil, fmt.Errorf("%s expects 2 arguments", name)
		}
		if !isIntegerLike(args[0].Type) || !isPointerType(args[1].Type) {
			return Value{}, nil, fmt.Errorf("%s expects character and stream arguments", name)
		}
		w, ok := r.lookupHostWriter(args[1].Int)
		if !ok {
			return Value{}, nil, fmt.Errorf("unknown stream handle %#x", args[1].Int)
		}
		ch := byte(args[0].Int)
		if _, err := w.Write([]byte{ch}); err != nil {
			return Value{}, nil, err
		}
		return IntValue(bytecode.TypeI32, int64(ch)), nil, nil
	}
}

func fgetcExtern(name string, r *ExternRegistry) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 1 {
			return Value{}, nil, fmt.Errorf("%s expects 1 argument", name)
		}
		if !isPointerType(args[0].Type) {
			return Value{}, nil, fmt.Errorf("%s expects stream pointer", name)
		}
		if _, ok := r.lookupHostWriter(args[0].Int); !ok {
			return Value{}, nil, fmt.Errorf("unknown stream handle %#x", args[0].Int)
		}
		if ch, ok := r.readHostChar(args[0].Int); ok {
			return IntValue(bytecode.TypeI32, int64(ch)), nil, nil
		}
		r.hostEOF[args[0].Int] = true
		return IntValue(bytecode.TypeI32, -1), nil, nil
	}
}

func ungetcExtern(name string, r *ExternRegistry) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 2 {
			return Value{}, nil, fmt.Errorf("%s expects 2 arguments", name)
		}
		if !isIntegerLike(args[0].Type) || !isPointerType(args[1].Type) {
			return Value{}, nil, fmt.Errorf("%s expects character and stream arguments", name)
		}
		ch := int32(args[0].Int)
		if ch == -1 {
			return IntValue(bytecode.TypeI32, -1), nil, nil
		}
		if _, ok := r.lookupHostWriter(args[1].Int); !ok {
			return Value{}, nil, fmt.Errorf("unknown stream handle %#x", args[1].Int)
		}
		b := byte(ch)
		r.hostPushback[args[1].Int] = append(r.hostPushback[args[1].Int], b)
		r.hostEOF[args[1].Int] = false
		return IntValue(bytecode.TypeI32, int64(b)), nil, nil
	}
}

func fgetsExtern(name string, r *ExternRegistry) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 3 {
			return Value{}, nil, fmt.Errorf("%s expects 3 arguments", name)
		}
		if !isPointerType(args[0].Type) || !isIntegerLike(args[1].Type) || !isPointerType(args[2].Type) {
			return Value{}, nil, fmt.Errorf("%s expects buffer, size, and stream arguments", name)
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("%s requires memory", name)
		}
		if _, ok := r.lookupHostWriter(args[2].Int); !ok {
			return Value{}, nil, fmt.Errorf("unknown stream handle %#x", args[2].Int)
		}
		n := int32(args[1].Int)
		if n <= 1 {
			return PtrValue(0), nil, nil
		}
		buf := make([]byte, 0, n-1)
		for len(buf) < int(n)-1 {
			ch, ok := r.readHostChar(args[2].Int)
			if !ok {
				break
			}
			buf = append(buf, ch)
			if ch == '\n' {
				break
			}
		}
		if len(buf) == 0 {
			r.hostEOF[args[2].Int] = true
			return PtrValue(0), nil, nil
		}
		block, off, err := ec.Memory.rangeAccess(args[0].Int, int64(len(buf)+1), true)
		if err != nil {
			return Value{}, nil, err
		}
		copy(block.data[off:], buf)
		block.data[off+len(buf)] = 0
		return PtrValue(args[0].Int), nil, nil
	}
}

func fputsExtern(name string, r *ExternRegistry) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 2 {
			return Value{}, nil, fmt.Errorf("%s expects 2 arguments", name)
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("%s requires memory", name)
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
	}
}

func perrorExtern(name string, r *ExternRegistry) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 1 {
			return Value{}, nil, fmt.Errorf("%s expects 1 argument", name)
		}
		if args[0].Int != 0 && !isPointerType(args[0].Type) {
			return Value{}, nil, fmt.Errorf("%s expects string pointer", name)
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("%s requires memory", name)
		}
		prefix := ""
		if args[0].Int != 0 {
			s, err := ec.Memory.ReadCString(args[0].Int)
			if err != nil {
				return Value{}, nil, err
			}
			prefix = s
		}
		msg := "error\n"
		if prefix != "" {
			msg = prefix + ": " + msg
		}
		if _, err := fmt.Fprint(r.externStderr(ec), msg); err != nil {
			return Value{}, nil, err
		}
		return Value{}, nil, nil
	}
}

func fflushExtern(name string, r *ExternRegistry) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 1 {
			return Value{}, nil, fmt.Errorf("%s expects 1 argument", name)
		}
		if args[0].Int != 0 && !isPointerType(args[0].Type) {
			return Value{}, nil, fmt.Errorf("%s expects stream pointer", name)
		}
		if args[0].Int != 0 {
			if _, ok := r.lookupHostWriter(args[0].Int); !ok {
				return Value{}, nil, fmt.Errorf("unknown stream handle %#x", args[0].Int)
			}
		}
		return IntValue(bytecode.TypeI32, 0), nil, nil
	}
}

func fcloseExtern(name string, r *ExternRegistry) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 1 {
			return Value{}, nil, fmt.Errorf("%s expects 1 argument", name)
		}
		if !isPointerType(args[0].Type) {
			return Value{}, nil, fmt.Errorf("%s expects stream pointer", name)
		}
		if _, ok := r.lookupHostWriter(args[0].Int); !ok {
			return Value{}, nil, fmt.Errorf("unknown stream handle %#x", args[0].Int)
		}
		return IntValue(bytecode.TypeI32, 0), nil, nil
	}
}

func filenoExtern(name string, r *ExternRegistry) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 1 {
			return Value{}, nil, fmt.Errorf("%s expects 1 argument", name)
		}
		if !isPointerType(args[0].Type) {
			return Value{}, nil, fmt.Errorf("%s expects stream pointer", name)
		}
		fd, ok := r.hostFDs[args[0].Int]
		if !ok {
			return Value{}, nil, fmt.Errorf("unknown stream handle %#x", args[0].Int)
		}
		return IntValue(bytecode.TypeI32, int64(fd)), nil, nil
	}
}

func fseekExtern(name string, r *ExternRegistry) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 3 {
			return Value{}, nil, fmt.Errorf("%s expects 3 arguments", name)
		}
		if !isPointerType(args[0].Type) || !isIntegerLike(args[1].Type) || !isIntegerLike(args[2].Type) {
			return Value{}, nil, fmt.Errorf("%s expects stream, offset, and whence arguments", name)
		}
		if _, ok := r.lookupHostWriter(args[0].Int); !ok {
			return Value{}, nil, fmt.Errorf("unknown stream handle %#x", args[0].Int)
		}
		return IntValue(bytecode.TypeI32, -1), nil, nil
	}
}

func ftellExtern(name string, r *ExternRegistry) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 1 {
			return Value{}, nil, fmt.Errorf("%s expects 1 argument", name)
		}
		if !isPointerType(args[0].Type) {
			return Value{}, nil, fmt.Errorf("%s expects stream pointer", name)
		}
		if _, ok := r.lookupHostWriter(args[0].Int); !ok {
			return Value{}, nil, fmt.Errorf("unknown stream handle %#x", args[0].Int)
		}
		return IntValue(bytecode.TypeI64, -1), nil, nil
	}
}

func rewindExtern(name string, r *ExternRegistry) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 1 {
			return Value{}, nil, fmt.Errorf("%s expects 1 argument", name)
		}
		if !isPointerType(args[0].Type) {
			return Value{}, nil, fmt.Errorf("%s expects stream pointer", name)
		}
		if _, ok := r.lookupHostWriter(args[0].Int); !ok {
			return Value{}, nil, fmt.Errorf("unknown stream handle %#x", args[0].Int)
		}
		delete(r.hostEOF, args[0].Int)
		return Value{}, nil, nil
	}
}

func fgetposExtern(name string, r *ExternRegistry) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 2 {
			return Value{}, nil, fmt.Errorf("%s expects 2 arguments", name)
		}
		if !isPointerType(args[0].Type) || !isPointerType(args[1].Type) {
			return Value{}, nil, fmt.Errorf("%s expects stream and position pointers", name)
		}
		if _, ok := r.lookupHostWriter(args[0].Int); !ok {
			return Value{}, nil, fmt.Errorf("unknown stream handle %#x", args[0].Int)
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("%s requires memory", name)
		}
		if _, _, err := ec.Memory.rangeAccess(args[1].Int, int64(valueSize(ec.Memory.target, bytecode.TypeI64)), true); err != nil {
			return Value{}, nil, err
		}
		return IntValue(bytecode.TypeI32, -1), nil, nil
	}
}

func fsetposExtern(name string, r *ExternRegistry) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 2 {
			return Value{}, nil, fmt.Errorf("%s expects 2 arguments", name)
		}
		if !isPointerType(args[0].Type) || !isPointerType(args[1].Type) {
			return Value{}, nil, fmt.Errorf("%s expects stream and position pointers", name)
		}
		if _, ok := r.lookupHostWriter(args[0].Int); !ok {
			return Value{}, nil, fmt.Errorf("unknown stream handle %#x", args[0].Int)
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("%s requires memory", name)
		}
		if _, _, err := ec.Memory.rangeAccess(args[1].Int, int64(valueSize(ec.Memory.target, bytecode.TypeI64)), false); err != nil {
			return Value{}, nil, err
		}
		return IntValue(bytecode.TypeI32, -1), nil, nil
	}
}

func setbufExtern(name string, r *ExternRegistry) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 2 {
			return Value{}, nil, fmt.Errorf("%s expects 2 arguments", name)
		}
		if !isPointerType(args[0].Type) || (args[1].Int != 0 && !isPointerType(args[1].Type)) {
			return Value{}, nil, fmt.Errorf("%s expects stream and buffer pointers", name)
		}
		if _, ok := r.lookupHostWriter(args[0].Int); !ok {
			return Value{}, nil, fmt.Errorf("unknown stream handle %#x", args[0].Int)
		}
		return Value{}, nil, nil
	}
}

func setvbufExtern(name string, r *ExternRegistry) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 4 {
			return Value{}, nil, fmt.Errorf("%s expects 4 arguments", name)
		}
		if !isPointerType(args[0].Type) || (args[1].Int != 0 && !isPointerType(args[1].Type)) || !isIntegerLike(args[2].Type) || !isIntegerLike(args[3].Type) {
			return Value{}, nil, fmt.Errorf("%s expects stream, buffer, mode, and size arguments", name)
		}
		if _, ok := r.lookupHostWriter(args[0].Int); !ok {
			return Value{}, nil, fmt.Errorf("unknown stream handle %#x", args[0].Int)
		}
		return IntValue(bytecode.TypeI32, 0), nil, nil
	}
}

func streamLockExtern(name string, r *ExternRegistry, returnsInt bool) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 1 {
			return Value{}, nil, fmt.Errorf("%s expects 1 argument", name)
		}
		if !isPointerType(args[0].Type) {
			return Value{}, nil, fmt.Errorf("%s expects stream pointer", name)
		}
		if _, ok := r.lookupHostWriter(args[0].Int); !ok {
			return Value{}, nil, fmt.Errorf("unknown stream handle %#x", args[0].Int)
		}
		if returnsInt {
			return IntValue(bytecode.TypeI32, 0), nil, nil
		}
		return Value{}, nil, nil
	}
}

func signedAbsExtern(name string, ret bytecode.ValueType) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 1 {
			return Value{}, nil, fmt.Errorf("%s expects 1 argument", name)
		}
		if !isIntegerLike(args[0].Type) {
			return Value{}, nil, fmt.Errorf("%s expects integer argument", name)
		}
		v := signedInt(args[0])
		if v < 0 {
			v = -v
		}
		return normalizeInt(IntValue(ret, v)), nil, nil
	}
}

func signedDivExtern(name string, typ bytecode.ValueType, fieldSize, fieldAlign int64) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 2 {
			return Value{}, nil, fmt.Errorf("%s expects 2 arguments", name)
		}
		if !isIntegerLike(args[0].Type) || !isIntegerLike(args[1].Type) {
			return Value{}, nil, fmt.Errorf("%s expects integer arguments", name)
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("%s requires memory", name)
		}
		numer := signedInt(args[0])
		denom := signedInt(args[1])
		if denom == 0 {
			return Value{}, nil, fmt.Errorf("%s division by zero", name)
		}
		addr, err := ec.Memory.TryAlloc(name+":result", fieldSize*2, fieldAlign, false, blockGlobal)
		if err != nil {
			return Value{}, nil, err
		}
		if err := ec.Memory.Store(addr, typ, fieldAlign, IntValue(typ, numer/denom)); err != nil {
			return Value{}, nil, err
		}
		if err := ec.Memory.Store(addr+uint64(fieldSize), typ, fieldAlign, IntValue(typ, numer%denom)); err != nil {
			return Value{}, nil, err
		}
		return ObjectAddrValue(addr), nil, nil
	}
}

func atoiExtern(name string, ret bytecode.ValueType) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 1 {
			return Value{}, nil, fmt.Errorf("%s expects 1 argument", name)
		}
		if !isPointerType(args[0].Type) {
			return Value{}, nil, fmt.Errorf("%s expects string pointer", name)
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("%s requires memory", name)
		}
		s, err := ec.Memory.ReadCString(args[0].Int)
		if err != nil {
			return Value{}, nil, err
		}
		return normalizeInt(IntValue(ret, parseAtoiString(s))), nil, nil
	}
}

func parseAtoiString(s string) int64 {
	i := 0
	for i < len(s) && isASCIIWhitespace(s[i]) {
		i++
	}
	neg := false
	if i < len(s) {
		switch s[i] {
		case '-':
			neg = true
			i++
		case '+':
			i++
		}
	}
	var v int64
	for i < len(s) && s[i] >= '0' && s[i] <= '9' {
		v = v*10 + int64(s[i]-'0')
		i++
	}
	if neg {
		return -v
	}
	return v
}

func strtoIntegerExtern(name string, ret bytecode.ValueType, signed bool) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 3 {
			return Value{}, nil, fmt.Errorf("%s expects 3 arguments", name)
		}
		if !isPointerType(args[0].Type) || (args[1].Int != 0 && !isPointerType(args[1].Type)) || !isIntegerLike(args[2].Type) {
			return Value{}, nil, fmt.Errorf("%s expects string, end pointer, and base arguments", name)
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("%s requires memory", name)
		}
		s, err := ec.Memory.ReadCString(args[0].Int)
		if err != nil {
			return Value{}, nil, err
		}
		parsed, err := parseStrtoIntegerString(name, s, signedInt(args[2]))
		if err != nil {
			return Value{}, nil, err
		}
		end := args[0].Int
		if parsed.converted {
			end, err = addSignedOffset(args[0].Int, int64(parsed.end))
			if err != nil {
				return Value{}, nil, err
			}
		}
		if args[1].Int != 0 {
			if err := ec.Memory.WritePointer(args[1].Int, end); err != nil {
				return Value{}, nil, err
			}
		}
		if signed {
			v := int64(parsed.value)
			if parsed.neg {
				v = -v
			}
			return normalizeInt(IntValue(ret, v)), nil, nil
		}
		v := parsed.value
		if parsed.neg {
			v = -v
		}
		return normalizeInt(UIntValue(ret, v)), nil, nil
	}
}

type parsedStrtoInteger struct {
	value     uint64
	neg       bool
	end       int
	converted bool
}

func parseStrtoIntegerString(name, s string, base int64) (parsedStrtoInteger, error) {
	if base != 0 && base != 8 && base != 10 && base != 16 {
		return parsedStrtoInteger{}, fmt.Errorf("%s unsupported base %d", name, base)
	}
	i := 0
	for i < len(s) && isASCIIWhitespace(s[i]) {
		i++
	}
	neg := false
	if i < len(s) {
		switch s[i] {
		case '-':
			neg = true
			i++
		case '+':
			i++
		}
	}
	digitStart := i
	actualBase := base
	if actualBase == 0 {
		actualBase = 10
		if i < len(s) && s[i] == '0' {
			actualBase = 8
			if i+1 < len(s) && (s[i+1] == 'x' || s[i+1] == 'X') {
				actualBase = 16
				i += 2
				digitStart = i
			}
		}
	} else if actualBase == 16 && i+1 < len(s) && s[i] == '0' && (s[i+1] == 'x' || s[i+1] == 'X') {
		i += 2
		digitStart = i
	}
	var v uint64
	for i < len(s) {
		digit := strtoDigit(s[i])
		if digit < 0 || int64(digit) >= actualBase {
			break
		}
		v = v*uint64(actualBase) + uint64(digit)
		i++
	}
	if i == digitStart {
		return parsedStrtoInteger{neg: neg}, nil
	}
	return parsedStrtoInteger{value: v, neg: neg, end: i, converted: true}, nil
}

func strtoDigit(ch byte) int {
	switch {
	case ch >= '0' && ch <= '9':
		return int(ch - '0')
	case ch >= 'a' && ch <= 'f':
		return int(ch-'a') + 10
	case ch >= 'A' && ch <= 'F':
		return int(ch-'A') + 10
	default:
		return -1
	}
}

func atofExtern(name string) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 1 {
			return Value{}, nil, fmt.Errorf("%s expects 1 argument", name)
		}
		if !isPointerType(args[0].Type) {
			return Value{}, nil, fmt.Errorf("%s expects string pointer", name)
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("%s requires memory", name)
		}
		s, err := ec.Memory.ReadCString(args[0].Int)
		if err != nil {
			return Value{}, nil, err
		}
		parsed := parseStrtoFloatString(s)
		return FloatValue(bytecode.TypeF64, parsed.value), nil, nil
	}
}

func strtoFloatExtern(name string, ret bytecode.ValueType) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 2 {
			return Value{}, nil, fmt.Errorf("%s expects 2 arguments", name)
		}
		if !isPointerType(args[0].Type) || (args[1].Int != 0 && !isPointerType(args[1].Type)) {
			return Value{}, nil, fmt.Errorf("%s expects string and end pointer arguments", name)
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("%s requires memory", name)
		}
		s, err := ec.Memory.ReadCString(args[0].Int)
		if err != nil {
			return Value{}, nil, err
		}
		parsed := parseStrtoFloatString(s)
		end := args[0].Int
		if parsed.converted {
			end, err = addSignedOffset(args[0].Int, int64(parsed.end))
			if err != nil {
				return Value{}, nil, err
			}
		}
		if args[1].Int != 0 {
			if err := ec.Memory.WritePointer(args[1].Int, end); err != nil {
				return Value{}, nil, err
			}
		}
		return FloatValue(ret, parsed.value), nil, nil
	}
}

func mblenExtern(name string) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 2 {
			return Value{}, nil, fmt.Errorf("%s expects 2 arguments", name)
		}
		if args[0].Int == 0 {
			return IntValue(bytecode.TypeI32, 0), nil, nil
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("%s requires memory", name)
		}
		if unsignedInt(args[1]) == 0 {
			return IntValue(bytecode.TypeI32, -1), nil, nil
		}
		ch, err := readMemoryByte(ec.Memory, args[0].Int)
		if err != nil {
			return Value{}, nil, err
		}
		if ch == 0 {
			return IntValue(bytecode.TypeI32, 0), nil, nil
		}
		return IntValue(bytecode.TypeI32, 1), nil, nil
	}
}

func mbtowcExtern(name string) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 3 {
			return Value{}, nil, fmt.Errorf("%s expects 3 arguments", name)
		}
		if args[1].Int == 0 {
			return IntValue(bytecode.TypeI32, 0), nil, nil
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("%s requires memory", name)
		}
		if unsignedInt(args[2]) == 0 {
			return IntValue(bytecode.TypeI32, -1), nil, nil
		}
		ch, err := readMemoryByte(ec.Memory, args[1].Int)
		if err != nil {
			return Value{}, nil, err
		}
		if args[0].Int != 0 {
			if err := ec.Memory.Store(args[0].Int, bytecode.TypeI32, 4, IntValue(bytecode.TypeI32, int64(ch))); err != nil {
				return Value{}, nil, err
			}
		}
		if ch == 0 {
			return IntValue(bytecode.TypeI32, 0), nil, nil
		}
		return IntValue(bytecode.TypeI32, 1), nil, nil
	}
}

func wctombExtern(name string) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 2 {
			return Value{}, nil, fmt.Errorf("%s expects 2 arguments", name)
		}
		if args[0].Int == 0 {
			return IntValue(bytecode.TypeI32, 0), nil, nil
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("%s requires memory", name)
		}
		wc := signedInt(args[1])
		if wc < 0 || wc > 255 {
			return IntValue(bytecode.TypeI32, -1), nil, nil
		}
		if err := writeMemoryByte(ec.Memory, args[0].Int, byte(wc)); err != nil {
			return Value{}, nil, err
		}
		return IntValue(bytecode.TypeI32, 1), nil, nil
	}
}

func mbstowcsExtern(name string) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 3 {
			return Value{}, nil, fmt.Errorf("%s expects 3 arguments", name)
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("%s requires memory", name)
		}
		src, err := ec.Memory.ReadCString(args[1].Int)
		if err != nil {
			return Value{}, nil, err
		}
		if args[0].Int == 0 {
			return UIntValue(bytecode.TypeU64, uint64(len(src))), nil, nil
		}
		n := unsignedInt(args[2])
		count := uint64(0)
		for count < n && count < uint64(len(src)) {
			if err := ec.Memory.Store(args[0].Int+count*4, bytecode.TypeI32, 4, IntValue(bytecode.TypeI32, int64(src[count]))); err != nil {
				return Value{}, nil, err
			}
			count++
		}
		if count < n {
			if err := ec.Memory.Store(args[0].Int+count*4, bytecode.TypeI32, 4, IntValue(bytecode.TypeI32, 0)); err != nil {
				return Value{}, nil, err
			}
		}
		return UIntValue(bytecode.TypeU64, count), nil, nil
	}
}

func wcstombsExtern(name string) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 3 {
			return Value{}, nil, fmt.Errorf("%s expects 3 arguments", name)
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("%s requires memory", name)
		}
		chars, err := readWideCString(ec.Memory, args[1].Int)
		if err != nil {
			return Value{}, nil, err
		}
		if args[0].Int == 0 {
			return UIntValue(bytecode.TypeU64, uint64(len(chars))), nil, nil
		}
		n := unsignedInt(args[2])
		count := uint64(0)
		for count < n && count < uint64(len(chars)) {
			wc := chars[count]
			if wc > 255 {
				return UIntValue(bytecode.TypeU64, ^uint64(0)), nil, nil
			}
			if err := writeMemoryByte(ec.Memory, args[0].Int+count, byte(wc)); err != nil {
				return Value{}, nil, err
			}
			count++
		}
		if count < n {
			if err := writeMemoryByte(ec.Memory, args[0].Int+count, 0); err != nil {
				return Value{}, nil, err
			}
		}
		return UIntValue(bytecode.TypeU64, count), nil, nil
	}
}

func randExtern(name string, r *ExternRegistry) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 0 {
			return Value{}, nil, fmt.Errorf("%s expects 0 arguments", name)
		}
		r.randSeed = r.randSeed*1103515245 + 12345
		return IntValue(bytecode.TypeI32, int64((r.randSeed/65536)%32768)), nil, nil
	}
}

func srandExtern(name string, r *ExternRegistry) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 1 {
			return Value{}, nil, fmt.Errorf("%s expects 1 argument", name)
		}
		if !isIntegerLike(args[0].Type) {
			return Value{}, nil, fmt.Errorf("%s expects seed argument", name)
		}
		r.randSeed = uint32(unsignedInt(args[0]))
		return Value{}, nil, nil
	}
}

func getenvExtern(name string) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 1 {
			return Value{}, nil, fmt.Errorf("%s expects 1 argument", name)
		}
		if !isPointerType(args[0].Type) {
			return Value{}, nil, fmt.Errorf("%s expects string pointer", name)
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("%s requires memory", name)
		}
		if _, err := ec.Memory.ReadCString(args[0].Int); err != nil {
			return Value{}, nil, err
		}
		return PtrValue(0), nil, nil
	}
}

func systemExtern(name string) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 1 {
			return Value{}, nil, fmt.Errorf("%s expects 1 argument", name)
		}
		if args[0].Int != 0 && !isPointerType(args[0].Type) {
			return Value{}, nil, fmt.Errorf("%s expects command pointer", name)
		}
		if args[0].Int == 0 {
			return IntValue(bytecode.TypeI32, 0), nil, nil
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("%s requires memory", name)
		}
		if _, err := ec.Memory.ReadCString(args[0].Int); err != nil {
			return Value{}, nil, err
		}
		return IntValue(bytecode.TypeI32, -1), nil, nil
	}
}

func atexitExtern(name string) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 1 {
			return Value{}, nil, fmt.Errorf("%s expects 1 argument", name)
		}
		if !isPointerType(args[0].Type) {
			return Value{}, nil, fmt.Errorf("%s expects function pointer", name)
		}
		return IntValue(bytecode.TypeI32, 0), nil, nil
	}
}

func exitExtern(name string) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 1 {
			return Value{}, nil, fmt.Errorf("%s expects 1 argument", name)
		}
		code, err := args[0].ExitCode()
		if err != nil {
			return Value{}, nil, err
		}
		return Value{}, &ExitStatus{Code: code}, nil
	}
}

func setlocaleExtern(name string, r *ExternRegistry) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 2 {
			return Value{}, nil, fmt.Errorf("%s expects 2 arguments", name)
		}
		if !isIntegerLike(args[0].Type) {
			return Value{}, nil, fmt.Errorf("%s expects category argument", name)
		}
		category := signedInt(args[0])
		if category < 0 || category > 5 {
			return PtrValue(0), nil, nil
		}
		if !isPointerType(args[1].Type) {
			return Value{}, nil, fmt.Errorf("%s expects locale pointer", name)
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("%s requires memory", name)
		}
		locale := "C"
		if args[1].Int != 0 {
			var err error
			locale, err = ec.Memory.ReadCString(args[1].Int)
			if err != nil {
				return Value{}, nil, err
			}
		}
		if locale != "" && locale != "C" {
			return PtrValue(0), nil, nil
		}
		addr, err := r.staticCString(ec.Memory, name, "C")
		if err != nil {
			return Value{}, nil, err
		}
		return PtrValue(addr), nil, nil
	}
}

func clockExtern(name string) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 0 {
			return Value{}, nil, fmt.Errorf("%s expects 0 arguments", name)
		}
		return IntValue(bytecode.TypeI64, 0), nil, nil
	}
}

func difftimeExtern(name string) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 2 {
			return Value{}, nil, fmt.Errorf("%s expects 2 arguments", name)
		}
		if !isIntegerLike(args[0].Type) || !isIntegerLike(args[1].Type) {
			return Value{}, nil, fmt.Errorf("%s expects time arguments", name)
		}
		return FloatValue(bytecode.TypeF64, float64(signedInt(args[0])-signedInt(args[1]))), nil, nil
	}
}

func timeExtern(name string) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 1 {
			return Value{}, nil, fmt.Errorf("%s expects 1 argument", name)
		}
		if !isPointerType(args[0].Type) {
			return Value{}, nil, fmt.Errorf("%s expects result pointer", name)
		}
		now := IntValue(bytecode.TypeI64, 0)
		if args[0].Int != 0 {
			if ec == nil || ec.Memory == nil {
				return Value{}, nil, fmt.Errorf("%s requires memory", name)
			}
			if err := ec.Memory.Store(args[0].Int, bytecode.TypeI64, 8, now); err != nil {
				return Value{}, nil, err
			}
		}
		return now, nil, nil
	}
}

type parsedStrtoFloat struct {
	value     float64
	end       int
	converted bool
}

func parseStrtoFloatString(s string) parsedStrtoFloat {
	start := 0
	for start < len(s) && isASCIIWhitespace(s[start]) {
		start++
	}
	for end := len(s); end > start; end-- {
		v, err := strconv.ParseFloat(s[start:end], 64)
		if err == nil {
			return parsedStrtoFloat{value: v, end: end, converted: true}
		}
	}
	return parsedStrtoFloat{}
}

func isASCIIWhitespace(ch byte) bool {
	switch ch {
	case ' ', '\f', '\n', '\r', '\t', '\v':
		return true
	default:
		return false
	}
}

func registerCtypeClassificationExterns(r *ExternRegistry) {
	r.Register("isdigit", ctypeClassificationExtern("isdigit", func(ch byte) bool {
		return ch >= '0' && ch <= '9'
	}))
	r.Register("isalpha", ctypeClassificationExtern("isalpha", func(ch byte) bool {
		return (ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z')
	}))
	r.Register("isalnum", ctypeClassificationExtern("isalnum", func(ch byte) bool {
		return (ch >= '0' && ch <= '9') || (ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z')
	}))
	r.Register("isspace", ctypeClassificationExtern("isspace", isASCIIWhitespace))
	r.Register("islower", ctypeClassificationExtern("islower", func(ch byte) bool {
		return ch >= 'a' && ch <= 'z'
	}))
	r.Register("isupper", ctypeClassificationExtern("isupper", func(ch byte) bool {
		return ch >= 'A' && ch <= 'Z'
	}))
	r.Register("isxdigit", ctypeClassificationExtern("isxdigit", func(ch byte) bool {
		return (ch >= '0' && ch <= '9') || (ch >= 'A' && ch <= 'F') || (ch >= 'a' && ch <= 'f')
	}))
	r.Register("isprint", ctypeClassificationExtern("isprint", func(ch byte) bool {
		return ch >= 0x20 && ch <= 0x7e
	}))
	r.Register("isblank", ctypeClassificationExtern("isblank", func(ch byte) bool {
		return ch == ' ' || ch == '\t'
	}))
	r.Register("iscntrl", ctypeClassificationExtern("iscntrl", func(ch byte) bool {
		return ch < 0x20 || ch == 0x7f
	}))
	r.Register("isgraph", ctypeClassificationExtern("isgraph", func(ch byte) bool {
		return ch >= 0x21 && ch <= 0x7e
	}))
	r.Register("ispunct", ctypeClassificationExtern("ispunct", func(ch byte) bool {
		return (ch >= 0x21 && ch <= 0x2f) || (ch >= 0x3a && ch <= 0x40) || (ch >= 0x5b && ch <= 0x60) || (ch >= 0x7b && ch <= 0x7e)
	}))
}

func ctypeClassificationExtern(name string, pred func(byte) bool) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 1 {
			return Value{}, nil, fmt.Errorf("%s expects 1 argument", name)
		}
		if !isIntegerLike(args[0].Type) {
			return Value{}, nil, fmt.Errorf("%s expects integer argument", name)
		}
		ch := signedInt(args[0])
		if ch < 0 || ch > 255 || !pred(byte(ch)) {
			return IntValue(bytecode.TypeI32, 0), nil, nil
		}
		return IntValue(bytecode.TypeI32, 1), nil, nil
	}
}

func registerCtypeCaseExterns(r *ExternRegistry) {
	r.Register("tolower", ctypeCaseExtern("tolower", func(ch byte) byte {
		if ch >= 'A' && ch <= 'Z' {
			return ch + ('a' - 'A')
		}
		return ch
	}))
	r.Register("toupper", ctypeCaseExtern("toupper", func(ch byte) byte {
		if ch >= 'a' && ch <= 'z' {
			return ch - ('a' - 'A')
		}
		return ch
	}))
}

func ctypeCaseExtern(name string, convert func(byte) byte) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 1 {
			return Value{}, nil, fmt.Errorf("%s expects 1 argument", name)
		}
		if !isIntegerLike(args[0].Type) {
			return Value{}, nil, fmt.Errorf("%s expects integer argument", name)
		}
		ch := signedInt(args[0])
		if ch < 0 || ch > 255 {
			return IntValue(bytecode.TypeI32, ch), nil, nil
		}
		return IntValue(bytecode.TypeI32, int64(convert(byte(ch)))), nil, nil
	}
}

func fwriteExtern(name string, r *ExternRegistry) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 4 {
			return Value{}, nil, fmt.Errorf("%s expects 4 arguments", name)
		}
		if !isPointerType(args[0].Type) || !isIntegerLike(args[1].Type) || !isIntegerLike(args[2].Type) || !isPointerType(args[3].Type) {
			return Value{}, nil, fmt.Errorf("%s expects buffer, size, count, and stream arguments", name)
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("%s requires memory", name)
		}
		w, ok := r.lookupHostWriter(args[3].Int)
		if !ok {
			return Value{}, nil, fmt.Errorf("unknown stream handle %#x", args[3].Int)
		}
		size, err := memorySizeArg(name, args[1])
		if err != nil {
			return Value{}, nil, err
		}
		count, err := memorySizeArg(name, args[2])
		if err != nil {
			return Value{}, nil, err
		}
		if size == 0 || count == 0 {
			return UIntValue(bytecode.TypeU64, 0), nil, nil
		}
		if size > int64(maxInt())/count {
			return Value{}, nil, fmt.Errorf("%s byte count overflows", name)
		}
		total := size * count
		block, off, err := ec.Memory.rangeAccess(args[0].Int, total, false)
		if err != nil {
			return Value{}, nil, err
		}
		if _, err := w.Write(block.data[off : off+int(total)]); err != nil {
			return Value{}, nil, err
		}
		return UIntValue(bytecode.TypeU64, uint64(count)), nil, nil
	}
}

func freadExtern(name string, r *ExternRegistry) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 4 {
			return Value{}, nil, fmt.Errorf("%s expects 4 arguments", name)
		}
		if !isPointerType(args[0].Type) || !isIntegerLike(args[1].Type) || !isIntegerLike(args[2].Type) || !isPointerType(args[3].Type) {
			return Value{}, nil, fmt.Errorf("%s expects buffer, size, count, and stream arguments", name)
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("%s requires memory", name)
		}
		if _, ok := r.lookupHostWriter(args[3].Int); !ok {
			return Value{}, nil, fmt.Errorf("unknown stream handle %#x", args[3].Int)
		}
		size, err := memorySizeArg(name, args[1])
		if err != nil {
			return Value{}, nil, err
		}
		count, err := memorySizeArg(name, args[2])
		if err != nil {
			return Value{}, nil, err
		}
		if size == 0 || count == 0 {
			return UIntValue(bytecode.TypeU64, 0), nil, nil
		}
		if size > int64(maxInt())/count {
			return Value{}, nil, fmt.Errorf("%s byte count overflows", name)
		}
		total := size * count
		block, off, err := ec.Memory.rangeAccess(args[0].Int, total, true)
		if err != nil {
			return Value{}, nil, err
		}
		read := int64(0)
		for read < total {
			ch, ok := r.readHostChar(args[3].Int)
			if !ok {
				break
			}
			block.data[off+int(read)] = ch
			read++
		}
		if read < total {
			r.hostEOF[args[3].Int] = true
		}
		return UIntValue(bytecode.TypeU64, uint64(read/size)), nil, nil
	}
}

func streamStatusExtern(name string, r *ExternRegistry) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 1 {
			return Value{}, nil, fmt.Errorf("%s expects 1 argument", name)
		}
		if !isPointerType(args[0].Type) {
			return Value{}, nil, fmt.Errorf("%s expects stream pointer", name)
		}
		if _, ok := r.lookupHostWriter(args[0].Int); !ok {
			return Value{}, nil, fmt.Errorf("unknown stream handle %#x", args[0].Int)
		}
		if (name == "feof" || name == "feof_unlocked") && r.hostEOF[args[0].Int] {
			return IntValue(bytecode.TypeI32, 1), nil, nil
		}
		return IntValue(bytecode.TypeI32, 0), nil, nil
	}
}

func clearerrExtern(name string, r *ExternRegistry) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 1 {
			return Value{}, nil, fmt.Errorf("%s expects 1 argument", name)
		}
		if !isPointerType(args[0].Type) {
			return Value{}, nil, fmt.Errorf("%s expects stream pointer", name)
		}
		if _, ok := r.lookupHostWriter(args[0].Int); !ok {
			return Value{}, nil, fmt.Errorf("unknown stream handle %#x", args[0].Int)
		}
		r.hostEOF[args[0].Int] = false
		return Value{}, nil, nil
	}
}

func registerAllocationExterns(r *ExternRegistry) {
	for _, name := range []string{"__builtin_malloc", "malloc"} {
		r.Register(name, mallocExtern(name))
	}
	for _, name := range []string{"__builtin_calloc", "calloc"} {
		r.Register(name, callocExtern(name))
	}
	r.Register("realloc", reallocExtern("realloc"))
	for _, name := range []string{"__builtin_strdup", "strdup"} {
		r.Register(name, strdupExtern(name))
	}
	r.Register("strndup", stringNDupExtern("strndup"))
	r.Register("free", freeExtern("free"))
	r.Register("__builtin_object_size", objectSizeExtern("__builtin_object_size"))
	r.Register("__builtin_dynamic_object_size", objectSizeExtern("__builtin_dynamic_object_size"))
}

func registerMemoryExterns(r *ExternRegistry) {
	for _, name := range []string{"__builtin_memcpy", "memcpy", "__builtin_memmove", "memmove"} {
		r.Register(name, memoryCopyExtern(name, false))
	}
	for _, name := range []string{"__builtin_mempcpy", "mempcpy"} {
		r.Register(name, memoryCopyExtern(name, true))
	}
	r.Register("memccpy", memoryCharCopyExtern("memccpy"))
	r.Register("bcopy", memoryBcopyExtern("bcopy"))
	for _, name := range []string{"__builtin_memset", "memset"} {
		r.Register(name, memorySetExtern(name))
	}
	for _, name := range []string{"__builtin_bzero", "bzero"} {
		r.Register(name, memoryBzeroExtern(name))
	}
	for _, name := range []string{"__builtin___memcpy_chk", "__builtin___memmove_chk"} {
		r.Register(name, memoryCheckedCopyExtern(name, false))
	}
	r.Register("__builtin___mempcpy_chk", memoryCheckedCopyExtern("__builtin___mempcpy_chk", true))
	r.Register("__builtin___memset_chk", memoryCheckedSetExtern("__builtin___memset_chk"))
	for _, name := range []string{"__builtin_strlen", "strlen"} {
		r.Register(name, stringLengthExtern(name))
	}
	r.Register("strnlen", stringNLengthExtern("strnlen"))
	r.Register("strerror", stringErrorExtern("strerror", r))
	for _, name := range []string{"__builtin_strchr", "strchr"} {
		r.Register(name, stringCharSearchExtern(name))
	}
	r.Register("strrchr", stringReverseCharSearchExtern("strrchr"))
	for _, name := range []string{"__builtin_strstr", "strstr"} {
		r.Register(name, stringSearchExtern(name))
	}
	r.Register("strpbrk", stringSetSearchExtern("strpbrk"))
	r.Register("strspn", stringSpanExtern("strspn", true))
	r.Register("strcspn", stringSpanExtern("strcspn", false))
	r.Register("strtok", stringTokenExtern("strtok", r))
	r.Register("strxfrm", stringTransformExtern("strxfrm"))
	r.Register("strncmp", stringNCompareExtern("strncmp"))
	r.Register("strcoll", stringCollateExtern("strcoll"))
	r.Register("memchr", memoryCharSearchExtern("memchr"))
	for _, name := range []string{"__builtin_strcpy", "strcpy"} {
		r.Register(name, stringCopyExtern(name, false))
	}
	for _, name := range []string{"__builtin_stpcpy", "stpcpy"} {
		r.Register(name, stringCopyExtern(name, true))
	}
	for _, name := range []string{"__builtin_strcat", "strcat"} {
		r.Register(name, stringConcatExtern(name))
	}
	for _, name := range []string{"__builtin_strncpy", "strncpy"} {
		r.Register(name, stringNCopyExtern(name, false))
	}
	for _, name := range []string{"__builtin_stpncpy", "stpncpy"} {
		r.Register(name, stringNCopyExtern(name, true))
	}
	for _, name := range []string{"__builtin_strncat", "strncat"} {
		r.Register(name, stringNConcatExtern(name))
	}
	for _, name := range []string{"__builtin___strcpy_chk", "__builtin___stpcpy_chk"} {
		r.Register(name, stringCheckedCopyExtern(name, name == "__builtin___stpcpy_chk"))
	}
	r.Register("__builtin___strcat_chk", stringCheckedConcatExtern("__builtin___strcat_chk"))
	for _, name := range []string{"__builtin___strncpy_chk", "__builtin___stpncpy_chk"} {
		r.Register(name, stringCheckedNCopyExtern(name, name == "__builtin___stpncpy_chk"))
	}
	r.Register("__builtin___strncat_chk", stringCheckedNConcatExtern("__builtin___strncat_chk"))
	for _, name := range []string{"__builtin_sprintf", "sprintf"} {
		r.Register(name, sprintfExtern(name))
	}
	for _, name := range []string{"__builtin_snprintf", "snprintf"} {
		r.Register(name, snprintfExtern(name))
	}
	for _, name := range []string{"__builtin_vsprintf", "vsprintf"} {
		r.Register(name, vsprintfExtern(name))
	}
	for _, name := range []string{"__builtin_vsnprintf", "vsnprintf"} {
		r.Register(name, vsnprintfExtern(name))
	}
	r.Register("__builtin___sprintf_chk", sprintfCheckedExtern("__builtin___sprintf_chk"))
	r.Register("__builtin___snprintf_chk", snprintfCheckedExtern("__builtin___snprintf_chk"))
	r.Register("__builtin___vsprintf_chk", vsprintfCheckedExtern("__builtin___vsprintf_chk"))
	r.Register("__builtin___vsnprintf_chk", vsnprintfCheckedExtern("__builtin___vsnprintf_chk"))
}

func registerOutputFormatExterns(r *ExternRegistry) {
	for _, name := range []string{"__builtin_printf", "__builtin_printf_unlocked", "printf", "printf_unlocked"} {
		r.Register(name, printfExtern(name, r))
	}
	for _, name := range []string{"__builtin_fprintf", "__builtin_fprintf_unlocked", "fprintf", "fprintf_unlocked"} {
		r.Register(name, fprintfExtern(name, r))
	}
	for _, name := range []string{"__builtin_vprintf", "vprintf", "vprintf_unlocked"} {
		r.Register(name, vprintfExtern(name, r))
	}
	for _, name := range []string{"__builtin_vfprintf", "vfprintf", "vfprintf_unlocked"} {
		r.Register(name, vfprintfExtern(name, r))
	}
	r.Register("__builtin___printf_chk", printfCheckedExtern("__builtin___printf_chk", r))
	r.Register("__builtin___fprintf_chk", fprintfCheckedExtern("__builtin___fprintf_chk", r))
	r.Register("__builtin___vprintf_chk", vprintfCheckedExtern("__builtin___vprintf_chk", r))
	r.Register("__builtin___vfprintf_chk", vfprintfCheckedExtern("__builtin___vfprintf_chk", r))
}

func registerMathExterns(r *ExternRegistry) {
	for _, suffix := range []string{"f", "", "l"} {
		r.Register("__cvm_fpclassify"+suffix, mathUnaryExtern("__cvm_fpclassify"+suffix, func(v Value) int64 {
			return cvmFPClassify(v)
		}))
		r.Register("__cvm_isfinite"+suffix, mathUnaryExtern("__cvm_isfinite"+suffix, func(v Value) int64 {
			class := cvmFPClassify(v)
			return boolInt(class != fpClassNaN && class != fpClassInfinite)
		}))
		r.Register("__cvm_isinf"+suffix, mathUnaryExtern("__cvm_isinf"+suffix, func(v Value) int64 {
			return boolInt(cvmFPClassify(v) == fpClassInfinite)
		}))
		r.Register("__cvm_isnan"+suffix, mathUnaryExtern("__cvm_isnan"+suffix, func(v Value) int64 {
			return boolInt(cvmFPClassify(v) == fpClassNaN)
		}))
		r.Register("__cvm_isnormal"+suffix, mathUnaryExtern("__cvm_isnormal"+suffix, func(v Value) int64 {
			return boolInt(cvmFPClassify(v) == fpClassNormal)
		}))
		r.Register("__cvm_signbit"+suffix, mathUnaryExtern("__cvm_signbit"+suffix, func(v Value) int64 {
			return boolInt(math.Signbit(cvmFloat(v)))
		}))
	}
	registerTgmathRealExterns(r, "__cvm_tgmath_sin", math.Sin)
	registerTgmathRealExterns(r, "__cvm_tgmath_exp", math.Exp)
	r.Register("fabsf", mathUnaryFloatExtern("fabsf", bytecode.TypeF32, math.Abs))
	r.Register("fabs", mathUnaryFloatExtern("fabs", bytecode.TypeF64, math.Abs))
	r.Register("fabsl", mathUnaryFloatExtern("fabsl", bytecode.TypeFLong, math.Abs))
	r.Register("sqrtf", mathUnaryFloatExtern("sqrtf", bytecode.TypeF32, math.Sqrt))
	r.Register("sqrt", mathUnaryFloatExtern("sqrt", bytecode.TypeF64, math.Sqrt))
	r.Register("sqrtl", mathUnaryFloatExtern("sqrtl", bytecode.TypeFLong, math.Sqrt))
	registerTgmathRealExterns(r, "sin", math.Sin)
	registerTgmathRealExterns(r, "cos", math.Cos)
	registerTgmathRealExterns(r, "tan", math.Tan)
	registerTgmathRealExterns(r, "sinh", math.Sinh)
	registerTgmathRealExterns(r, "cosh", math.Cosh)
	registerTgmathRealExterns(r, "tanh", math.Tanh)
	registerTgmathRealExterns(r, "asin", math.Asin)
	registerTgmathRealExterns(r, "acos", math.Acos)
	registerTgmathRealExterns(r, "atan", math.Atan)
	registerTgmathRealExterns(r, "asinh", math.Asinh)
	registerTgmathRealExterns(r, "acosh", math.Acosh)
	registerTgmathRealExterns(r, "atanh", math.Atanh)
	registerTgmathRealExterns(r, "cbrt", math.Cbrt)
	registerTgmathRealExterns(r, "erf", math.Erf)
	registerTgmathRealExterns(r, "erfc", math.Erfc)
	registerTgmathRealExterns(r, "tgamma", math.Gamma)
	registerTgmathRealExterns(r, "lgamma", func(v float64) float64 {
		r, _ := math.Lgamma(v)
		return r
	})
	registerTgmathRealExterns(r, "exp", math.Exp)
	registerTgmathRealExterns(r, "exp2", math.Exp2)
	registerTgmathRealExterns(r, "expm1", math.Expm1)
	registerTgmathRealExterns(r, "log", math.Log)
	registerTgmathRealExterns(r, "log10", math.Log10)
	registerTgmathRealExterns(r, "log1p", math.Log1p)
	registerTgmathRealExterns(r, "log2", math.Log2)
	registerTgmathRealExterns(r, "ceil", math.Ceil)
	registerTgmathRealExterns(r, "floor", math.Floor)
	registerTgmathRealExterns(r, "trunc", math.Trunc)
	registerTgmathRealExterns(r, "round", math.Round)
	registerTgmathRealExterns(r, "nearbyint", math.RoundToEven)
	registerTgmathRealExterns(r, "rint", math.RoundToEven)
	registerTgmathRealExterns(r, "logb", math.Logb)
	registerTgmathIntExterns(r, "ilogb", math.Ilogb)
	registerTgmathLongExterns(r, "lrint", math.RoundToEven)
	registerTgmathLongExterns(r, "lround", math.Round)
	registerTgmathLongExterns(r, "llrint", math.RoundToEven)
	registerTgmathLongExterns(r, "llround", math.Round)
	registerTgmathRealIntBinaryExterns(r, "scalbn", math.Ldexp)
	registerTgmathRealIntBinaryExterns(r, "scalbln", math.Ldexp)
	registerTgmathRealIntBinaryExterns(r, "ldexp", math.Ldexp)
	registerTgmathFrexpExterns(r, "frexp")
	registerTgmathRemquoExterns(r, "remquo")
	registerTgmathRealBinaryExterns(r, "pow", math.Pow)
	registerTgmathRealBinaryExterns(r, "atan2", math.Atan2)
	registerTgmathRealBinaryExterns(r, "hypot", math.Hypot)
	registerTgmathRealBinaryExterns(r, "fdim", math.Dim)
	registerTgmathRealBinaryExterns(r, "fmax", math.Max)
	registerTgmathRealBinaryExterns(r, "fmin", math.Min)
	registerTgmathRealBinaryExterns(r, "fmod", math.Mod)
	registerTgmathRealBinaryExterns(r, "remainder", math.Remainder)
	registerTgmathRealBinaryExterns(r, "copysign", math.Copysign)
	registerTgmathRealBinaryExterns(r, "nextafter", math.Nextafter)
	registerTgmathRealBinaryExterns(r, "nexttoward", math.Nextafter)
	registerTgmathRealTernaryExterns(r, "fma", math.FMA)
	registerTgmathRealExterns(r, "__cvm_tgmath_sqrt", math.Sqrt)
	registerTgmathRealExterns(r, "__cvm_tgmath_fabs", math.Abs)
	registerTgmathRealExterns(r, "__cvm_tgmath_cos", math.Cos)
	registerTgmathRealExterns(r, "__cvm_tgmath_tan", math.Tan)
	registerTgmathRealExterns(r, "__cvm_tgmath_log", math.Log)
	registerTgmathRealExterns(r, "__cvm_tgmath_sinh", math.Sinh)
	registerTgmathRealExterns(r, "__cvm_tgmath_cosh", math.Cosh)
	registerTgmathRealExterns(r, "__cvm_tgmath_tanh", math.Tanh)
	registerTgmathRealExterns(r, "__cvm_tgmath_asin", math.Asin)
	registerTgmathRealExterns(r, "__cvm_tgmath_acos", math.Acos)
	registerTgmathRealExterns(r, "__cvm_tgmath_atan", math.Atan)
	registerTgmathRealExterns(r, "__cvm_tgmath_asinh", math.Asinh)
	registerTgmathRealExterns(r, "__cvm_tgmath_acosh", math.Acosh)
	registerTgmathRealExterns(r, "__cvm_tgmath_atanh", math.Atanh)
	registerTgmathRealExterns(r, "__cvm_tgmath_cbrt", math.Cbrt)
	registerTgmathRealExterns(r, "__cvm_tgmath_ceil", math.Ceil)
	registerTgmathRealExterns(r, "__cvm_tgmath_floor", math.Floor)
	registerTgmathRealExterns(r, "__cvm_tgmath_trunc", math.Trunc)
	registerTgmathRealExterns(r, "__cvm_tgmath_round", math.Round)
	registerTgmathRealExterns(r, "__cvm_tgmath_exp2", math.Exp2)
	registerTgmathRealExterns(r, "__cvm_tgmath_expm1", math.Expm1)
	registerTgmathRealExterns(r, "__cvm_tgmath_log10", math.Log10)
	registerTgmathRealExterns(r, "__cvm_tgmath_log1p", math.Log1p)
	registerTgmathRealExterns(r, "__cvm_tgmath_log2", math.Log2)
	registerTgmathRealExterns(r, "__cvm_tgmath_erf", math.Erf)
	registerTgmathRealExterns(r, "__cvm_tgmath_erfc", math.Erfc)
	registerTgmathRealExterns(r, "__cvm_tgmath_tgamma", math.Gamma)
	registerTgmathRealExterns(r, "__cvm_tgmath_lgamma", func(v float64) float64 {
		r, _ := math.Lgamma(v)
		return r
	})
	registerTgmathRealExterns(r, "__cvm_tgmath_nearbyint", math.RoundToEven)
	registerTgmathRealExterns(r, "__cvm_tgmath_rint", math.RoundToEven)
	registerTgmathRealExterns(r, "__cvm_tgmath_logb", math.Logb)
	registerTgmathIntExterns(r, "__cvm_tgmath_ilogb", math.Ilogb)
	registerTgmathLongExterns(r, "__cvm_tgmath_lrint", math.RoundToEven)
	registerTgmathLongExterns(r, "__cvm_tgmath_lround", math.Round)
	registerTgmathLongExterns(r, "__cvm_tgmath_llrint", math.RoundToEven)
	registerTgmathLongExterns(r, "__cvm_tgmath_llround", math.Round)
	registerTgmathRealBinaryExterns(r, "__cvm_tgmath_pow", math.Pow)
	r.Register("__builtin_pow", mathBinaryFloatExtern("__builtin_pow", bytecode.TypeF64, math.Pow))
	r.Register("__builtin_huge_valf", mathConstantFloatExtern("__builtin_huge_valf", bytecode.TypeF32, math.Inf(1)))
	r.Register("__builtin_huge_val", mathConstantFloatExtern("__builtin_huge_val", bytecode.TypeF64, math.Inf(1)))
	r.Register("__builtin_huge_vall", mathConstantFloatExtern("__builtin_huge_vall", bytecode.TypeFLong, math.Inf(1)))
	r.Register("__builtin_nan", mathNanExtern("__builtin_nan"))
	r.Register("nan", mathNanExtern("nan"))
	registerTgmathRealBinaryExterns(r, "__cvm_tgmath_atan2", math.Atan2)
	registerTgmathRealBinaryExterns(r, "__cvm_tgmath_hypot", math.Hypot)
	registerTgmathRealBinaryExterns(r, "__cvm_tgmath_fdim", math.Dim)
	registerTgmathRealBinaryExterns(r, "__cvm_tgmath_fmax", math.Max)
	registerTgmathRealBinaryExterns(r, "__cvm_tgmath_fmin", math.Min)
	registerTgmathRealBinaryExterns(r, "__cvm_tgmath_fmod", math.Mod)
	registerTgmathRealBinaryExterns(r, "__cvm_tgmath_remainder", math.Remainder)
	registerTgmathRealBinaryExterns(r, "__cvm_tgmath_copysign", math.Copysign)
	registerTgmathRealBinaryExterns(r, "__cvm_tgmath_nextafter", math.Nextafter)
	registerTgmathRealBinaryExterns(r, "__cvm_tgmath_nexttoward", math.Nextafter)
	registerTgmathRealIntBinaryExterns(r, "__cvm_tgmath_scalbn", math.Ldexp)
	registerTgmathRealIntBinaryExterns(r, "__cvm_tgmath_scalbln", math.Ldexp)
	registerTgmathRealIntBinaryExterns(r, "__cvm_tgmath_ldexp", math.Ldexp)
	registerTgmathFrexpExterns(r, "__cvm_tgmath_frexp")
	registerTgmathRemquoExterns(r, "__cvm_tgmath_remquo")
	registerTgmathComplexRealExterns(r, "__cvm_tgmath_carg", func(z complex128) float64 { return math.Atan2(imag(z), real(z)) })
	registerTgmathComplexRealExterns(r, "__cvm_tgmath_cimag", func(z complex128) float64 { return imag(z) })
	registerTgmathComplexRealExterns(r, "__cvm_tgmath_creal", func(z complex128) float64 { return real(z) })
	registerTgmathComplexRealExterns(r, "carg", func(z complex128) float64 { return math.Atan2(imag(z), real(z)) })
	registerTgmathComplexRealExterns(r, "cimag", func(z complex128) float64 { return imag(z) })
	registerTgmathComplexRealExterns(r, "creal", func(z complex128) float64 { return real(z) })
	registerTgmathRealTernaryExterns(r, "__cvm_tgmath_fma", math.FMA)
	registerTgmathComplexExterns(r, "conj", cmplx.Conj)
	registerTgmathComplexExterns(r, "cproj", cvmComplexProject)
	registerTgmathComplexExterns(r, "csin", cmplx.Sin)
	registerTgmathComplexExterns(r, "ccos", cmplx.Cos)
	registerTgmathComplexExterns(r, "ctan", cmplx.Tan)
	registerTgmathComplexExterns(r, "csinh", cmplx.Sinh)
	registerTgmathComplexExterns(r, "ccosh", cmplx.Cosh)
	registerTgmathComplexExterns(r, "ctanh", cmplx.Tanh)
	registerTgmathComplexExterns(r, "casin", cmplx.Asin)
	registerTgmathComplexExterns(r, "cacos", cmplx.Acos)
	registerTgmathComplexExterns(r, "catan", cmplx.Atan)
	registerTgmathComplexExterns(r, "casinh", cmplx.Asinh)
	registerTgmathComplexExterns(r, "cacosh", cmplx.Acosh)
	registerTgmathComplexExterns(r, "catanh", cmplx.Atanh)
	registerTgmathComplexExterns(r, "cexp", cmplx.Exp)
	registerTgmathComplexExterns(r, "clog", cmplx.Log)
	registerTgmathComplexExterns(r, "csqrt", cmplx.Sqrt)
	registerTgmathComplexBinaryExterns(r, "cpow", cmplx.Pow)
	registerTgmathComplexExterns(r, "__cvm_tgmath_conj", cmplx.Conj)
	registerTgmathComplexExterns(r, "__cvm_tgmath_cproj", cvmComplexProject)
	registerTgmathComplexExterns(r, "__cvm_tgmath_csin", cmplx.Sin)
	registerTgmathComplexExterns(r, "__cvm_tgmath_cexp", cmplx.Exp)
	registerTgmathComplexExterns(r, "__cvm_tgmath_csqrt", cmplx.Sqrt)
	registerTgmathComplexExterns(r, "__cvm_tgmath_ccos", cmplx.Cos)
	registerTgmathComplexExterns(r, "__cvm_tgmath_ctan", cmplx.Tan)
	registerTgmathComplexExterns(r, "__cvm_tgmath_clog", cmplx.Log)
	registerTgmathComplexExterns(r, "__cvm_tgmath_csinh", cmplx.Sinh)
	registerTgmathComplexExterns(r, "__cvm_tgmath_ccosh", cmplx.Cosh)
	registerTgmathComplexExterns(r, "__cvm_tgmath_ctanh", cmplx.Tanh)
	registerTgmathComplexExterns(r, "__cvm_tgmath_casin", cmplx.Asin)
	registerTgmathComplexExterns(r, "__cvm_tgmath_cacos", cmplx.Acos)
	registerTgmathComplexExterns(r, "__cvm_tgmath_catan", cmplx.Atan)
	registerTgmathComplexExterns(r, "__cvm_tgmath_casinh", cmplx.Asinh)
	registerTgmathComplexExterns(r, "__cvm_tgmath_cacosh", cmplx.Acosh)
	registerTgmathComplexExterns(r, "__cvm_tgmath_catanh", cmplx.Atanh)
	registerTgmathComplexBinaryExterns(r, "__cvm_tgmath_cpow", cmplx.Pow)
	r.Register("cabsf", complexAbsExtern("cabsf", bytecode.TypeF32, 4))
	r.Register("cabs", complexAbsExtern("cabs", bytecode.TypeF64, 8))
	r.Register("cabsl", complexAbsExtern("cabsl", bytecode.TypeFLong, 16))
	r.Register("__builtin_cabsf", complexAbsExtern("__builtin_cabsf", bytecode.TypeF32, 4))
	r.Register("__builtin_cabs", complexAbsExtern("__builtin_cabs", bytecode.TypeF64, 8))
	r.Register("__builtin_cabsl", complexAbsExtern("__builtin_cabsl", bytecode.TypeFLong, 16))
	r.Register("__cvm_isunordered", func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 2 {
			return Value{}, nil, fmt.Errorf("__cvm_isunordered expects 2 arguments")
		}
		return IntValue(bytecode.TypeI32, boolInt(math.IsNaN(cvmFloat(args[0])) || math.IsNaN(cvmFloat(args[1])))), nil, nil
	})
}

func abortExtern() ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		return Value{}, nil, &TrapError{Reason: "abort"}
	}
}

func mallocExtern(name string) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 1 {
			return Value{}, nil, fmt.Errorf("%s expects 1 argument", name)
		}
		if !isIntegerLike(args[0].Type) {
			return Value{}, nil, fmt.Errorf("%s expects size argument", name)
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("%s requires memory", name)
		}
		size, err := memorySizeArg(name, args[0])
		if err != nil {
			return Value{}, nil, err
		}
		addr, err := ec.Memory.TryAlloc("extern:"+name, nonzeroAllocSize(size), ec.Memory.target.PointerAlign, false, blockGlobal)
		if err != nil {
			return Value{}, nil, err
		}
		return PtrValue(addr), nil, nil
	}
}

func callocExtern(name string) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 2 {
			return Value{}, nil, fmt.Errorf("%s expects 2 arguments", name)
		}
		if !isIntegerLike(args[0].Type) || !isIntegerLike(args[1].Type) {
			return Value{}, nil, fmt.Errorf("%s expects count and size arguments", name)
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("%s requires memory", name)
		}
		count := unsignedInt(args[0])
		size := unsignedInt(args[1])
		if count != 0 && size > uint64(maxInt())/count {
			return Value{}, nil, fmt.Errorf("%s allocation size overflows", name)
		}
		total := count * size
		if total > uint64(maxInt()) {
			return Value{}, nil, fmt.Errorf("%s allocation size %d exceeds int range", name, total)
		}
		addr, err := ec.Memory.TryAlloc("extern:"+name, nonzeroAllocSize(int64(total)), ec.Memory.target.PointerAlign, false, blockGlobal)
		if err != nil {
			return Value{}, nil, err
		}
		return PtrValue(addr), nil, nil
	}
}

func reallocExtern(name string) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 2 {
			return Value{}, nil, fmt.Errorf("%s expects 2 arguments", name)
		}
		if args[0].Int != 0 && !isPointerType(args[0].Type) {
			return Value{}, nil, fmt.Errorf("%s expects pointer argument", name)
		}
		if !isIntegerLike(args[1].Type) {
			return Value{}, nil, fmt.Errorf("%s expects size argument", name)
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("%s requires memory", name)
		}
		size, err := memorySizeArg(name, args[1])
		if err != nil {
			return Value{}, nil, err
		}
		if args[0].Int == 0 {
			addr, err := ec.Memory.TryAlloc("extern:"+name, nonzeroAllocSize(size), ec.Memory.target.PointerAlign, false, blockGlobal)
			if err != nil {
				return Value{}, nil, err
			}
			return PtrValue(addr), nil, nil
		}
		oldBlock, oldOff, err := ec.Memory.rangeAccess(args[0].Int, 1, false)
		if err != nil {
			return Value{}, nil, err
		}
		if oldOff != 0 || oldBlock.kind != blockGlobal {
			return Value{}, nil, fmt.Errorf("%s expects pointer returned by allocation extern", name)
		}
		if size == 0 {
			if err := ec.Memory.Free(args[0].Int, blockGlobal); err != nil {
				return Value{}, nil, err
			}
			return PtrValue(0), nil, nil
		}
		addr, err := ec.Memory.TryAlloc("extern:"+name, nonzeroAllocSize(size), ec.Memory.target.PointerAlign, false, blockGlobal)
		if err != nil {
			return Value{}, nil, err
		}
		copySize := int64(len(oldBlock.data))
		if size < copySize {
			copySize = size
		}
		if copySize > 0 {
			newBlock, newOff, err := ec.Memory.rangeAccess(addr, copySize, true)
			if err != nil {
				return Value{}, nil, err
			}
			copy(newBlock.data[newOff:newOff+int(copySize)], oldBlock.data[:int(copySize)])
		}
		if err := ec.Memory.Free(args[0].Int, blockGlobal); err != nil {
			return Value{}, nil, err
		}
		return PtrValue(addr), nil, nil
	}
}

func strdupExtern(name string) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 1 {
			return Value{}, nil, fmt.Errorf("%s expects 1 argument", name)
		}
		if !isPointerType(args[0].Type) {
			return Value{}, nil, fmt.Errorf("%s expects string argument", name)
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("%s requires memory", name)
		}
		src, err := ec.Memory.ReadCString(args[0].Int)
		if err != nil {
			return Value{}, nil, err
		}
		data := append([]byte(src), 0)
		addr, err := ec.Memory.TryAlloc("extern:"+name, int64(len(data)), 1, false, blockGlobal)
		if err != nil {
			return Value{}, nil, err
		}
		if err := writeMemoryBytes(ec.Memory, addr, data); err != nil {
			return Value{}, nil, err
		}
		return PtrValue(addr), nil, nil
	}
}

func stringNDupExtern(name string) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 2 {
			return Value{}, nil, fmt.Errorf("%s expects 2 arguments", name)
		}
		if !isPointerType(args[0].Type) || !isIntegerLike(args[1].Type) {
			return Value{}, nil, fmt.Errorf("%s expects string and size arguments", name)
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("%s requires memory", name)
		}
		n, err := memorySizeArg(name, args[1])
		if err != nil {
			return Value{}, nil, err
		}
		var data []byte
		for i := int64(0); i < n; i++ {
			addr, err := addSignedOffset(args[0].Int, i)
			if err != nil {
				return Value{}, nil, err
			}
			block, off, err := ec.Memory.rangeAccess(addr, 1, false)
			if err != nil {
				return Value{}, nil, err
			}
			if block.data[off] == 0 {
				break
			}
			data = append(data, block.data[off])
		}
		data = append(data, 0)
		addr, err := ec.Memory.TryAlloc("extern:"+name, int64(len(data)), 1, false, blockGlobal)
		if err != nil {
			return Value{}, nil, err
		}
		if err := writeMemoryBytes(ec.Memory, addr, data); err != nil {
			return Value{}, nil, err
		}
		return PtrValue(addr), nil, nil
	}
}

func freeExtern(name string) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 1 {
			return Value{}, nil, fmt.Errorf("%s expects 1 argument", name)
		}
		if args[0].Int != 0 && !isPointerType(args[0].Type) {
			return Value{}, nil, fmt.Errorf("%s expects pointer argument", name)
		}
		if args[0].Int == 0 {
			return Value{}, nil, nil
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("%s requires memory", name)
		}
		if err := ec.Memory.Free(args[0].Int, blockGlobal); err != nil {
			return Value{}, nil, err
		}
		return Value{}, nil, nil
	}
}

func nonzeroAllocSize(size int64) int64 {
	if size == 0 {
		return 1
	}
	return size
}

func objectSizeExtern(name string) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 2 {
			return Value{}, nil, fmt.Errorf("%s expects 2 arguments", name)
		}
		if !isPointerType(args[0].Type) || !isIntegerLike(args[1].Type) {
			return Value{}, nil, fmt.Errorf("%s expects pointer and type arguments", name)
		}
		switch signedInt(args[1]) & 3 {
		case 2, 3:
			return UIntValue(bytecode.TypeU64, 0), nil, nil
		default:
			return UIntValue(bytecode.TypeU64, ^uint64(0)), nil, nil
		}
	}
}

func memoryCompareExtern(name string) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 3 {
			return Value{}, nil, fmt.Errorf("%s expects 3 arguments", name)
		}
		if !isPointerType(args[0].Type) || !isPointerType(args[1].Type) || !isIntegerLike(args[2].Type) {
			return Value{}, nil, fmt.Errorf("%s expects pointer, pointer, and size arguments", name)
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("%s requires memory", name)
		}
		size, err := memorySizeArg(name, args[2])
		if err != nil {
			return Value{}, nil, err
		}
		if size == 0 {
			return IntValue(bytecode.TypeI32, 0), nil, nil
		}
		leftBlock, leftOff, err := ec.Memory.rangeAccess(args[0].Int, size, false)
		if err != nil {
			return Value{}, nil, err
		}
		rightBlock, rightOff, err := ec.Memory.rangeAccess(args[1].Int, size, false)
		if err != nil {
			return Value{}, nil, err
		}
		left := leftBlock.data[leftOff : leftOff+int(size)]
		right := rightBlock.data[rightOff : rightOff+int(size)]
		return IntValue(bytecode.TypeI32, int64(memcmpResult(left, right))), nil, nil
	}
}

func memoryCopyExtern(name string, returnEnd bool) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 3 {
			return Value{}, nil, fmt.Errorf("%s expects 3 arguments", name)
		}
		if !isPointerType(args[0].Type) || !isPointerType(args[1].Type) || !isIntegerLike(args[2].Type) {
			return Value{}, nil, fmt.Errorf("%s expects pointer, pointer, and size arguments", name)
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("%s requires memory", name)
		}
		size, err := memorySizeArg(name, args[2])
		if err != nil {
			return Value{}, nil, err
		}
		if err := ec.Memory.Copy(args[0].Int, args[1].Int, size); err != nil {
			return Value{}, nil, err
		}
		addr := args[0].Int
		if returnEnd {
			addr += uint64(size)
		}
		return PtrValue(addr), nil, nil
	}
}

func memoryCharCopyExtern(name string) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 4 {
			return Value{}, nil, fmt.Errorf("%s expects 4 arguments", name)
		}
		if !isPointerType(args[0].Type) || !isPointerType(args[1].Type) || !isIntegerLike(args[2].Type) || !isIntegerLike(args[3].Type) {
			return Value{}, nil, fmt.Errorf("%s expects pointer, pointer, integer, and size arguments", name)
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("%s requires memory", name)
		}
		size, err := memorySizeArg(name, args[3])
		if err != nil {
			return Value{}, nil, err
		}
		if size == 0 {
			return PtrValue(0), nil, nil
		}
		dstBlock, dstOff, err := ec.Memory.rangeAccess(args[0].Int, size, true)
		if err != nil {
			return Value{}, nil, err
		}
		srcBlock, srcOff, err := ec.Memory.rangeAccess(args[1].Int, size, false)
		if err != nil {
			return Value{}, nil, err
		}
		needle := byte(args[2].Int)
		for i := 0; i < int(size); i++ {
			ch := srcBlock.data[srcOff+i]
			dstBlock.data[dstOff+i] = ch
			if ch == needle {
				return PtrValue(args[0].Int + uint64(i) + 1), nil, nil
			}
		}
		return PtrValue(0), nil, nil
	}
}

func memoryBcopyExtern(name string) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 3 {
			return Value{}, nil, fmt.Errorf("%s expects 3 arguments", name)
		}
		if !isPointerType(args[0].Type) || !isPointerType(args[1].Type) || !isIntegerLike(args[2].Type) {
			return Value{}, nil, fmt.Errorf("%s expects source, destination, and size arguments", name)
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("%s requires memory", name)
		}
		size, err := memorySizeArg(name, args[2])
		if err != nil {
			return Value{}, nil, err
		}
		if err := ec.Memory.Copy(args[1].Int, args[0].Int, size); err != nil {
			return Value{}, nil, err
		}
		return Value{}, nil, nil
	}
}

func memorySetExtern(name string) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 3 {
			return Value{}, nil, fmt.Errorf("%s expects 3 arguments", name)
		}
		if !isPointerType(args[0].Type) || !isIntegerLike(args[1].Type) || !isIntegerLike(args[2].Type) {
			return Value{}, nil, fmt.Errorf("%s expects pointer, integer, and size arguments", name)
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("%s requires memory", name)
		}
		size, err := memorySizeArg(name, args[2])
		if err != nil {
			return Value{}, nil, err
		}
		if err := ec.Memory.Set(args[0].Int, byte(unsignedInt(args[1])), size); err != nil {
			return Value{}, nil, err
		}
		return PtrValue(args[0].Int), nil, nil
	}
}

func memoryBzeroExtern(name string) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 2 {
			return Value{}, nil, fmt.Errorf("%s expects 2 arguments", name)
		}
		if !isPointerType(args[0].Type) || !isIntegerLike(args[1].Type) {
			return Value{}, nil, fmt.Errorf("%s expects pointer and size arguments", name)
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("%s requires memory", name)
		}
		size, err := memorySizeArg(name, args[1])
		if err != nil {
			return Value{}, nil, err
		}
		if err := ec.Memory.Set(args[0].Int, 0, size); err != nil {
			return Value{}, nil, err
		}
		return Value{}, nil, nil
	}
}

func memoryCheckedCopyExtern(name string, returnEnd bool) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 4 {
			return Value{}, nil, fmt.Errorf("%s expects 4 arguments", name)
		}
		if !isPointerType(args[0].Type) || !isPointerType(args[1].Type) || !isIntegerLike(args[2].Type) || !isIntegerLike(args[3].Type) {
			return Value{}, nil, fmt.Errorf("%s expects pointer, pointer, size, and object-size arguments", name)
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("%s requires memory", name)
		}
		size, err := memorySizeArg(name, args[2])
		if err != nil {
			return Value{}, nil, err
		}
		if err := checkObjectSize(name, uint64(size), args[3]); err != nil {
			return Value{}, nil, err
		}
		if err := ec.Memory.Copy(args[0].Int, args[1].Int, size); err != nil {
			return Value{}, nil, err
		}
		addr := args[0].Int
		if returnEnd {
			addr += uint64(size)
		}
		return PtrValue(addr), nil, nil
	}
}

func memoryCheckedSetExtern(name string) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 4 {
			return Value{}, nil, fmt.Errorf("%s expects 4 arguments", name)
		}
		if !isPointerType(args[0].Type) || !isIntegerLike(args[1].Type) || !isIntegerLike(args[2].Type) || !isIntegerLike(args[3].Type) {
			return Value{}, nil, fmt.Errorf("%s expects pointer, integer, size, and object-size arguments", name)
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("%s requires memory", name)
		}
		size, err := memorySizeArg(name, args[2])
		if err != nil {
			return Value{}, nil, err
		}
		if err := checkObjectSize(name, uint64(size), args[3]); err != nil {
			return Value{}, nil, err
		}
		if err := ec.Memory.Set(args[0].Int, byte(unsignedInt(args[1])), size); err != nil {
			return Value{}, nil, err
		}
		return PtrValue(args[0].Int), nil, nil
	}
}

func memorySizeArg(name string, arg Value) (int64, error) {
	size := unsignedInt(arg)
	if size > uint64(maxInt()) {
		return 0, fmt.Errorf("%s size %d exceeds int range", name, size)
	}
	return int64(size), nil
}

func checkObjectSize(name string, size uint64, objectSize Value) error {
	limit := unsignedInt(objectSize)
	if limit != ^uint64(0) && size > limit {
		return &TrapError{Reason: fmt.Sprintf("%s object size %d is smaller than operation size %d", name, limit, size)}
	}
	return nil
}

func stringLengthExtern(name string) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 1 {
			return Value{}, nil, fmt.Errorf("%s expects 1 argument", name)
		}
		if !isPointerType(args[0].Type) {
			return Value{}, nil, fmt.Errorf("%s expects string argument", name)
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("%s requires memory", name)
		}
		s, err := ec.Memory.ReadCString(args[0].Int)
		if err != nil {
			return Value{}, nil, err
		}
		return UIntValue(bytecode.TypeU64, uint64(len(s))), nil, nil
	}
}

func stringNLengthExtern(name string) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 2 {
			return Value{}, nil, fmt.Errorf("%s expects 2 arguments", name)
		}
		if !isPointerType(args[0].Type) || !isIntegerLike(args[1].Type) {
			return Value{}, nil, fmt.Errorf("%s expects string and size arguments", name)
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("%s requires memory", name)
		}
		n, err := memorySizeArg(name, args[1])
		if err != nil {
			return Value{}, nil, err
		}
		if n == 0 {
			return UIntValue(bytecode.TypeU64, 0), nil, nil
		}
		for i := int64(0); i < n; i++ {
			addr, err := addSignedOffset(args[0].Int, i)
			if err != nil {
				return Value{}, nil, err
			}
			block, off, err := ec.Memory.rangeAccess(addr, 1, false)
			if err != nil {
				return Value{}, nil, err
			}
			if block.data[off] == 0 {
				return UIntValue(bytecode.TypeU64, uint64(i)), nil, nil
			}
		}
		return UIntValue(bytecode.TypeU64, uint64(n)), nil, nil
	}
}

func stringErrorExtern(name string, r *ExternRegistry) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 1 {
			return Value{}, nil, fmt.Errorf("%s expects 1 argument", name)
		}
		if !isIntegerLike(args[0].Type) {
			return Value{}, nil, fmt.Errorf("%s expects integer argument", name)
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("%s requires memory", name)
		}
		addr, err := r.staticCString(ec.Memory, name, "error")
		if err != nil {
			return Value{}, nil, err
		}
		return PtrValue(addr), nil, nil
	}
}

func stringCharSearchExtern(name string) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 2 {
			return Value{}, nil, fmt.Errorf("%s expects 2 arguments", name)
		}
		if !isPointerType(args[0].Type) || !isIntegerLike(args[1].Type) {
			return Value{}, nil, fmt.Errorf("%s expects string and integer arguments", name)
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("%s requires memory", name)
		}
		s, err := ec.Memory.ReadCString(args[0].Int)
		if err != nil {
			return Value{}, nil, err
		}
		ch := byte(unsignedInt(args[1]))
		if ch == 0 {
			return PtrValue(args[0].Int + uint64(len(s))), nil, nil
		}
		idx := strings.IndexByte(s, ch)
		if idx < 0 {
			return PtrValue(0), nil, nil
		}
		return PtrValue(args[0].Int + uint64(idx)), nil, nil
	}
}

func stringReverseCharSearchExtern(name string) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 2 {
			return Value{}, nil, fmt.Errorf("%s expects 2 arguments", name)
		}
		if !isPointerType(args[0].Type) || !isIntegerLike(args[1].Type) {
			return Value{}, nil, fmt.Errorf("%s expects string and integer arguments", name)
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("%s requires memory", name)
		}
		s, err := ec.Memory.ReadCString(args[0].Int)
		if err != nil {
			return Value{}, nil, err
		}
		ch := byte(unsignedInt(args[1]))
		if ch == 0 {
			addr, err := addSignedOffset(args[0].Int, int64(len(s)))
			if err != nil {
				return Value{}, nil, err
			}
			return PtrValue(addr), nil, nil
		}
		idx := strings.LastIndexByte(s, ch)
		if idx < 0 {
			return PtrValue(0), nil, nil
		}
		addr, err := addSignedOffset(args[0].Int, int64(idx))
		if err != nil {
			return Value{}, nil, err
		}
		return PtrValue(addr), nil, nil
	}
}

func stringSearchExtern(name string) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 2 {
			return Value{}, nil, fmt.Errorf("%s expects 2 arguments", name)
		}
		if !isPointerType(args[0].Type) || !isPointerType(args[1].Type) {
			return Value{}, nil, fmt.Errorf("%s expects string arguments", name)
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("%s requires memory", name)
		}
		haystack, err := ec.Memory.ReadCString(args[0].Int)
		if err != nil {
			return Value{}, nil, err
		}
		needle, err := ec.Memory.ReadCString(args[1].Int)
		if err != nil {
			return Value{}, nil, err
		}
		idx := strings.Index(haystack, needle)
		if idx < 0 {
			return PtrValue(0), nil, nil
		}
		return PtrValue(args[0].Int + uint64(idx)), nil, nil
	}
}

func stringSetSearchExtern(name string) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 2 {
			return Value{}, nil, fmt.Errorf("%s expects 2 arguments", name)
		}
		if !isPointerType(args[0].Type) || !isPointerType(args[1].Type) {
			return Value{}, nil, fmt.Errorf("%s expects string arguments", name)
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("%s requires memory", name)
		}
		s, err := ec.Memory.ReadCString(args[0].Int)
		if err != nil {
			return Value{}, nil, err
		}
		accept, err := ec.Memory.ReadCString(args[1].Int)
		if err != nil {
			return Value{}, nil, err
		}
		for i := 0; i < len(s); i++ {
			if strings.IndexByte(accept, s[i]) < 0 {
				continue
			}
			addr, err := addSignedOffset(args[0].Int, int64(i))
			if err != nil {
				return Value{}, nil, err
			}
			return PtrValue(addr), nil, nil
		}
		return PtrValue(0), nil, nil
	}
}

func stringSpanExtern(name string, acceptMatch bool) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 2 {
			return Value{}, nil, fmt.Errorf("%s expects 2 arguments", name)
		}
		if !isPointerType(args[0].Type) || !isPointerType(args[1].Type) {
			return Value{}, nil, fmt.Errorf("%s expects string arguments", name)
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("%s requires memory", name)
		}
		s, err := ec.Memory.ReadCString(args[0].Int)
		if err != nil {
			return Value{}, nil, err
		}
		set, err := ec.Memory.ReadCString(args[1].Int)
		if err != nil {
			return Value{}, nil, err
		}
		for i := 0; i < len(s); i++ {
			matches := strings.IndexByte(set, s[i]) >= 0
			if matches != acceptMatch {
				return UIntValue(bytecode.TypeU64, uint64(i)), nil, nil
			}
		}
		return UIntValue(bytecode.TypeU64, uint64(len(s))), nil, nil
	}
}

func stringTokenExtern(name string, r *ExternRegistry) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 2 {
			return Value{}, nil, fmt.Errorf("%s expects 2 arguments", name)
		}
		if !isPointerType(args[0].Type) || !isPointerType(args[1].Type) {
			return Value{}, nil, fmt.Errorf("%s expects string arguments", name)
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("%s requires memory", name)
		}
		delims, err := ec.Memory.ReadCString(args[1].Int)
		if err != nil {
			return Value{}, nil, err
		}
		start := args[0].Int
		if start == 0 {
			start = r.strtokNext[ec.Memory]
		}
		if start == 0 {
			return PtrValue(0), nil, nil
		}

		addr := start
		for {
			ch, err := readMemoryByte(ec.Memory, addr)
			if err != nil {
				return Value{}, nil, err
			}
			if ch == 0 {
				r.strtokNext[ec.Memory] = 0
				return PtrValue(0), nil, nil
			}
			if strings.IndexByte(delims, ch) < 0 {
				break
			}
			addr, err = addSignedOffset(addr, 1)
			if err != nil {
				return Value{}, nil, err
			}
		}

		tokenStart := addr
		for {
			ch, err := readMemoryByte(ec.Memory, addr)
			if err != nil {
				return Value{}, nil, err
			}
			if ch == 0 {
				r.strtokNext[ec.Memory] = 0
				return PtrValue(tokenStart), nil, nil
			}
			if strings.IndexByte(delims, ch) >= 0 {
				if err := writeMemoryByte(ec.Memory, addr, 0); err != nil {
					return Value{}, nil, err
				}
				next, err := addSignedOffset(addr, 1)
				if err != nil {
					return Value{}, nil, err
				}
				r.strtokNext[ec.Memory] = next
				return PtrValue(tokenStart), nil, nil
			}
			addr, err = addSignedOffset(addr, 1)
			if err != nil {
				return Value{}, nil, err
			}
		}
	}
}

func stringTransformExtern(name string) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 3 {
			return Value{}, nil, fmt.Errorf("%s expects 3 arguments", name)
		}
		if !isPointerType(args[0].Type) || !isPointerType(args[1].Type) || !isIntegerLike(args[2].Type) {
			return Value{}, nil, fmt.Errorf("%s expects destination, source, and size arguments", name)
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("%s requires memory", name)
		}
		src, err := ec.Memory.ReadCString(args[1].Int)
		if err != nil {
			return Value{}, nil, err
		}
		n, err := memorySizeArg(name, args[2])
		if err != nil {
			return Value{}, nil, err
		}
		if n > 0 {
			limit := int(n) - 1
			if limit > len(src) {
				limit = len(src)
			}
			out := make([]byte, limit+1)
			copy(out, src[:limit])
			if err := writeMemoryBytes(ec.Memory, args[0].Int, out); err != nil {
				return Value{}, nil, err
			}
		}
		return UIntValue(bytecode.TypeU64, uint64(len(src))), nil, nil
	}
}

func stringNCompareExtern(name string) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 3 {
			return Value{}, nil, fmt.Errorf("%s expects 3 arguments", name)
		}
		if !isPointerType(args[0].Type) || !isPointerType(args[1].Type) || !isIntegerLike(args[2].Type) {
			return Value{}, nil, fmt.Errorf("%s expects two strings and size argument", name)
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("%s requires memory", name)
		}
		n, err := memorySizeArg(name, args[2])
		if err != nil {
			return Value{}, nil, err
		}
		left, err := ec.Memory.ReadCString(args[0].Int)
		if err != nil {
			return Value{}, nil, err
		}
		right, err := ec.Memory.ReadCString(args[1].Int)
		if err != nil {
			return Value{}, nil, err
		}
		return IntValue(bytecode.TypeI32, int64(strncmpResult(left, right, n))), nil, nil
	}
}

func memoryCharSearchExtern(name string) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 3 {
			return Value{}, nil, fmt.Errorf("%s expects 3 arguments", name)
		}
		if !isPointerType(args[0].Type) || !isIntegerLike(args[1].Type) || !isIntegerLike(args[2].Type) {
			return Value{}, nil, fmt.Errorf("%s expects memory, character, and size arguments", name)
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("%s requires memory", name)
		}
		n, err := memorySizeArg(name, args[2])
		if err != nil {
			return Value{}, nil, err
		}
		if n == 0 {
			return PtrValue(0), nil, nil
		}
		block, off, err := ec.Memory.rangeAccess(args[0].Int, n, false)
		if err != nil {
			return Value{}, nil, err
		}
		needle := byte(args[1].Int)
		for i, b := range block.data[off : off+int(n)] {
			if b == needle {
				addr, err := addSignedOffset(args[0].Int, int64(i))
				if err != nil {
					return Value{}, nil, err
				}
				return PtrValue(addr), nil, nil
			}
		}
		return PtrValue(0), nil, nil
	}
}

func stringCollateExtern(name string) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 2 {
			return Value{}, nil, fmt.Errorf("%s expects 2 arguments", name)
		}
		if !isPointerType(args[0].Type) || !isPointerType(args[1].Type) {
			return Value{}, nil, fmt.Errorf("%s expects string arguments", name)
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("%s requires memory", name)
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
	}
}

func stringCopyExtern(name string, returnEnd bool) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 2 {
			return Value{}, nil, fmt.Errorf("%s expects 2 arguments", name)
		}
		if !isPointerType(args[0].Type) || !isPointerType(args[1].Type) {
			return Value{}, nil, fmt.Errorf("%s expects string arguments", name)
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("%s requires memory", name)
		}
		src, err := ec.Memory.ReadCString(args[1].Int)
		if err != nil {
			return Value{}, nil, err
		}
		data := append([]byte(src), 0)
		if err := writeMemoryBytes(ec.Memory, args[0].Int, data); err != nil {
			return Value{}, nil, err
		}
		addr := args[0].Int
		if returnEnd {
			addr, err = addSignedOffset(args[0].Int, int64(len(src)))
			if err != nil {
				return Value{}, nil, err
			}
		}
		return PtrValue(addr), nil, nil
	}
}

func stringConcatExtern(name string) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 2 {
			return Value{}, nil, fmt.Errorf("%s expects 2 arguments", name)
		}
		if !isPointerType(args[0].Type) || !isPointerType(args[1].Type) {
			return Value{}, nil, fmt.Errorf("%s expects string arguments", name)
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("%s requires memory", name)
		}
		dst, err := ec.Memory.ReadCString(args[0].Int)
		if err != nil {
			return Value{}, nil, err
		}
		src, err := ec.Memory.ReadCString(args[1].Int)
		if err != nil {
			return Value{}, nil, err
		}
		start, err := addSignedOffset(args[0].Int, int64(len(dst)))
		if err != nil {
			return Value{}, nil, err
		}
		if err := writeMemoryBytes(ec.Memory, start, append([]byte(src), 0)); err != nil {
			return Value{}, nil, err
		}
		return PtrValue(args[0].Int), nil, nil
	}
}

func stringNCopyExtern(name string, returnEnd bool) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 3 {
			return Value{}, nil, fmt.Errorf("%s expects 3 arguments", name)
		}
		if !isPointerType(args[0].Type) || !isPointerType(args[1].Type) || !isIntegerLike(args[2].Type) {
			return Value{}, nil, fmt.Errorf("%s expects string, string, and size arguments", name)
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("%s requires memory", name)
		}
		size, err := memorySizeArg(name, args[2])
		if err != nil {
			return Value{}, nil, err
		}
		src, err := ec.Memory.ReadCString(args[1].Int)
		if err != nil {
			return Value{}, nil, err
		}
		data := make([]byte, int(size))
		copied := copy(data, []byte(src))
		if err := writeMemoryBytes(ec.Memory, args[0].Int, data); err != nil {
			return Value{}, nil, err
		}
		addr := args[0].Int
		if returnEnd {
			offset := size
			if copied < int(size) {
				offset = int64(copied)
			}
			addr, err = addSignedOffset(args[0].Int, offset)
			if err != nil {
				return Value{}, nil, err
			}
		}
		return PtrValue(addr), nil, nil
	}
}

func stringNConcatExtern(name string) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 3 {
			return Value{}, nil, fmt.Errorf("%s expects 3 arguments", name)
		}
		if !isPointerType(args[0].Type) || !isPointerType(args[1].Type) || !isIntegerLike(args[2].Type) {
			return Value{}, nil, fmt.Errorf("%s expects string, string, and size arguments", name)
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("%s requires memory", name)
		}
		size, err := memorySizeArg(name, args[2])
		if err != nil {
			return Value{}, nil, err
		}
		dst, err := ec.Memory.ReadCString(args[0].Int)
		if err != nil {
			return Value{}, nil, err
		}
		src, err := ec.Memory.ReadCString(args[1].Int)
		if err != nil {
			return Value{}, nil, err
		}
		count := len(src)
		if int64(count) > size {
			count = int(size)
		}
		start, err := addSignedOffset(args[0].Int, int64(len(dst)))
		if err != nil {
			return Value{}, nil, err
		}
		data := append([]byte(src[:count]), 0)
		if err := writeMemoryBytes(ec.Memory, start, data); err != nil {
			return Value{}, nil, err
		}
		return PtrValue(args[0].Int), nil, nil
	}
}

func stringCheckedCopyExtern(name string, returnEnd bool) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 3 {
			return Value{}, nil, fmt.Errorf("%s expects 3 arguments", name)
		}
		if !isPointerType(args[0].Type) || !isPointerType(args[1].Type) || !isIntegerLike(args[2].Type) {
			return Value{}, nil, fmt.Errorf("%s expects string, string, and object-size arguments", name)
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("%s requires memory", name)
		}
		src, err := ec.Memory.ReadCString(args[1].Int)
		if err != nil {
			return Value{}, nil, err
		}
		if err := checkObjectSize(name, uint64(len(src)+1), args[2]); err != nil {
			return Value{}, nil, err
		}
		if err := writeMemoryBytes(ec.Memory, args[0].Int, append([]byte(src), 0)); err != nil {
			return Value{}, nil, err
		}
		addr := args[0].Int
		if returnEnd {
			addr, err = addSignedOffset(args[0].Int, int64(len(src)))
			if err != nil {
				return Value{}, nil, err
			}
		}
		return PtrValue(addr), nil, nil
	}
}

func stringCheckedConcatExtern(name string) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 3 {
			return Value{}, nil, fmt.Errorf("%s expects 3 arguments", name)
		}
		if !isPointerType(args[0].Type) || !isPointerType(args[1].Type) || !isIntegerLike(args[2].Type) {
			return Value{}, nil, fmt.Errorf("%s expects string, string, and object-size arguments", name)
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("%s requires memory", name)
		}
		dst, err := ec.Memory.ReadCString(args[0].Int)
		if err != nil {
			return Value{}, nil, err
		}
		src, err := ec.Memory.ReadCString(args[1].Int)
		if err != nil {
			return Value{}, nil, err
		}
		if err := checkObjectSize(name, uint64(len(dst)+len(src)+1), args[2]); err != nil {
			return Value{}, nil, err
		}
		start, err := addSignedOffset(args[0].Int, int64(len(dst)))
		if err != nil {
			return Value{}, nil, err
		}
		if err := writeMemoryBytes(ec.Memory, start, append([]byte(src), 0)); err != nil {
			return Value{}, nil, err
		}
		return PtrValue(args[0].Int), nil, nil
	}
}

func stringCheckedNCopyExtern(name string, returnEnd bool) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 4 {
			return Value{}, nil, fmt.Errorf("%s expects 4 arguments", name)
		}
		if !isPointerType(args[0].Type) || !isPointerType(args[1].Type) || !isIntegerLike(args[2].Type) || !isIntegerLike(args[3].Type) {
			return Value{}, nil, fmt.Errorf("%s expects string, string, size, and object-size arguments", name)
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("%s requires memory", name)
		}
		size, err := memorySizeArg(name, args[2])
		if err != nil {
			return Value{}, nil, err
		}
		if err := checkObjectSize(name, uint64(size), args[3]); err != nil {
			return Value{}, nil, err
		}
		src, err := ec.Memory.ReadCString(args[1].Int)
		if err != nil {
			return Value{}, nil, err
		}
		data := make([]byte, int(size))
		copied := copy(data, []byte(src))
		if err := writeMemoryBytes(ec.Memory, args[0].Int, data); err != nil {
			return Value{}, nil, err
		}
		addr := args[0].Int
		if returnEnd {
			offset := size
			if copied < int(size) {
				offset = int64(copied)
			}
			addr, err = addSignedOffset(args[0].Int, offset)
			if err != nil {
				return Value{}, nil, err
			}
		}
		return PtrValue(addr), nil, nil
	}
}

func stringCheckedNConcatExtern(name string) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 4 {
			return Value{}, nil, fmt.Errorf("%s expects 4 arguments", name)
		}
		if !isPointerType(args[0].Type) || !isPointerType(args[1].Type) || !isIntegerLike(args[2].Type) || !isIntegerLike(args[3].Type) {
			return Value{}, nil, fmt.Errorf("%s expects string, string, size, and object-size arguments", name)
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("%s requires memory", name)
		}
		size, err := memorySizeArg(name, args[2])
		if err != nil {
			return Value{}, nil, err
		}
		dst, err := ec.Memory.ReadCString(args[0].Int)
		if err != nil {
			return Value{}, nil, err
		}
		src, err := ec.Memory.ReadCString(args[1].Int)
		if err != nil {
			return Value{}, nil, err
		}
		count := len(src)
		if int64(count) > size {
			count = int(size)
		}
		if err := checkObjectSize(name, uint64(len(dst)+count+1), args[3]); err != nil {
			return Value{}, nil, err
		}
		start, err := addSignedOffset(args[0].Int, int64(len(dst)))
		if err != nil {
			return Value{}, nil, err
		}
		if err := writeMemoryBytes(ec.Memory, start, append([]byte(src[:count]), 0)); err != nil {
			return Value{}, nil, err
		}
		return PtrValue(args[0].Int), nil, nil
	}
}

func writeMemoryBytes(mem *Memory, addr uint64, data []byte) error {
	if len(data) == 0 {
		return nil
	}
	block, off, err := mem.rangeAccess(addr, int64(len(data)), true)
	if err != nil {
		return err
	}
	copy(block.data[off:off+len(data)], data)
	return nil
}

func readMemoryByte(mem *Memory, addr uint64) (byte, error) {
	block, off, err := mem.rangeAccess(addr, 1, false)
	if err != nil {
		return 0, err
	}
	return block.data[off], nil
}

func writeMemoryByte(mem *Memory, addr uint64, value byte) error {
	block, off, err := mem.rangeAccess(addr, 1, true)
	if err != nil {
		return err
	}
	block.data[off] = value
	return nil
}

func readWideCString(mem *Memory, addr uint64) ([]uint64, error) {
	var out []uint64
	for {
		ch, err := mem.Load(addr+uint64(len(out))*4, bytecode.TypeI32, 4)
		if err != nil {
			return nil, err
		}
		if ch.Int == 0 {
			return out, nil
		}
		out = append(out, unsignedInt(ch))
	}
}

func sprintfExtern(name string) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) < 2 {
			return Value{}, nil, fmt.Errorf("%s expects at least 2 arguments", name)
		}
		if !isPointerType(args[0].Type) || !isPointerType(args[1].Type) {
			return Value{}, nil, fmt.Errorf("%s expects destination and format pointers", name)
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("%s requires memory", name)
		}
		out, err := formatCString(name, ec.Memory, args[1].Int, args[2:])
		if err != nil {
			return Value{}, nil, err
		}
		if err := writeMemoryBytes(ec.Memory, args[0].Int, append([]byte(out), 0)); err != nil {
			return Value{}, nil, err
		}
		return IntValue(bytecode.TypeI32, int64(len(out))), nil, nil
	}
}

func snprintfExtern(name string) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) < 3 {
			return Value{}, nil, fmt.Errorf("%s expects at least 3 arguments", name)
		}
		if !isPointerType(args[0].Type) || !isIntegerLike(args[1].Type) || !isPointerType(args[2].Type) {
			return Value{}, nil, fmt.Errorf("%s expects destination, size, and format arguments", name)
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("%s requires memory", name)
		}
		out, err := formatCString(name, ec.Memory, args[2].Int, args[3:])
		if err != nil {
			return Value{}, nil, err
		}
		size, err := memorySizeArg(name, args[1])
		if err != nil {
			return Value{}, nil, err
		}
		if size > 0 {
			n := len(out)
			if n > int(size)-1 {
				n = int(size) - 1
			}
			data := make([]byte, n+1)
			copy(data, out[:n])
			if err := writeMemoryBytes(ec.Memory, args[0].Int, data); err != nil {
				return Value{}, nil, err
			}
		}
		return IntValue(bytecode.TypeI32, int64(len(out))), nil, nil
	}
}

func vsprintfExtern(name string) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 3 {
			return Value{}, nil, fmt.Errorf("%s expects 3 arguments", name)
		}
		if !isPointerType(args[0].Type) || !isPointerType(args[1].Type) || !isPointerType(args[2].Type) {
			return Value{}, nil, fmt.Errorf("%s expects destination, format, and va_list pointers", name)
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("%s requires memory", name)
		}
		out, err := formatCString(name, ec.Memory, args[1].Int, nil)
		if err != nil {
			return Value{}, nil, err
		}
		if err := writeMemoryBytes(ec.Memory, args[0].Int, append([]byte(out), 0)); err != nil {
			return Value{}, nil, err
		}
		return IntValue(bytecode.TypeI32, int64(len(out))), nil, nil
	}
}

func vsnprintfExtern(name string) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 4 {
			return Value{}, nil, fmt.Errorf("%s expects 4 arguments", name)
		}
		if !isPointerType(args[0].Type) || !isIntegerLike(args[1].Type) || !isPointerType(args[2].Type) || !isPointerType(args[3].Type) {
			return Value{}, nil, fmt.Errorf("%s expects destination, size, format, and va_list arguments", name)
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("%s requires memory", name)
		}
		out, err := formatCString(name, ec.Memory, args[2].Int, nil)
		if err != nil {
			return Value{}, nil, err
		}
		size, err := memorySizeArg(name, args[1])
		if err != nil {
			return Value{}, nil, err
		}
		if size > 0 {
			n := len(out)
			if n > int(size)-1 {
				n = int(size) - 1
			}
			data := make([]byte, n+1)
			copy(data, out[:n])
			if err := writeMemoryBytes(ec.Memory, args[0].Int, data); err != nil {
				return Value{}, nil, err
			}
		}
		return IntValue(bytecode.TypeI32, int64(len(out))), nil, nil
	}
}

func sprintfCheckedExtern(name string) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) < 4 {
			return Value{}, nil, fmt.Errorf("%s expects at least 4 arguments", name)
		}
		if !isPointerType(args[0].Type) || !isIntegerLike(args[1].Type) || !isIntegerLike(args[2].Type) || !isPointerType(args[3].Type) {
			return Value{}, nil, fmt.Errorf("%s expects destination, flag, object-size, and format arguments", name)
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("%s requires memory", name)
		}
		out, err := formatCString(name, ec.Memory, args[3].Int, args[4:])
		if err != nil {
			return Value{}, nil, err
		}
		if err := checkObjectSize(name, uint64(len(out)+1), args[2]); err != nil {
			return Value{}, nil, err
		}
		if err := writeMemoryBytes(ec.Memory, args[0].Int, append([]byte(out), 0)); err != nil {
			return Value{}, nil, err
		}
		return IntValue(bytecode.TypeI32, int64(len(out))), nil, nil
	}
}

func snprintfCheckedExtern(name string) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) < 5 {
			return Value{}, nil, fmt.Errorf("%s expects at least 5 arguments", name)
		}
		if !isPointerType(args[0].Type) || !isIntegerLike(args[1].Type) || !isIntegerLike(args[2].Type) || !isIntegerLike(args[3].Type) || !isPointerType(args[4].Type) {
			return Value{}, nil, fmt.Errorf("%s expects destination, size, flag, object-size, and format arguments", name)
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("%s requires memory", name)
		}
		size, err := memorySizeArg(name, args[1])
		if err != nil {
			return Value{}, nil, err
		}
		if err := checkObjectSize(name, uint64(size), args[3]); err != nil {
			return Value{}, nil, err
		}
		out, err := formatCString(name, ec.Memory, args[4].Int, args[5:])
		if err != nil {
			return Value{}, nil, err
		}
		if size > 0 {
			n := len(out)
			if n > int(size)-1 {
				n = int(size) - 1
			}
			data := make([]byte, n+1)
			copy(data, out[:n])
			if err := writeMemoryBytes(ec.Memory, args[0].Int, data); err != nil {
				return Value{}, nil, err
			}
		}
		return IntValue(bytecode.TypeI32, int64(len(out))), nil, nil
	}
}

func vsprintfCheckedExtern(name string) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 5 {
			return Value{}, nil, fmt.Errorf("%s expects 5 arguments", name)
		}
		if !isPointerType(args[0].Type) || !isIntegerLike(args[1].Type) || !isIntegerLike(args[2].Type) || !isPointerType(args[3].Type) || !isPointerType(args[4].Type) {
			return Value{}, nil, fmt.Errorf("%s expects destination, flag, object-size, format, and va_list arguments", name)
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("%s requires memory", name)
		}
		out, err := formatCString(name, ec.Memory, args[3].Int, nil)
		if err != nil {
			return Value{}, nil, err
		}
		if err := checkObjectSize(name, uint64(len(out)+1), args[2]); err != nil {
			return Value{}, nil, err
		}
		if err := writeMemoryBytes(ec.Memory, args[0].Int, append([]byte(out), 0)); err != nil {
			return Value{}, nil, err
		}
		return IntValue(bytecode.TypeI32, int64(len(out))), nil, nil
	}
}

func vsnprintfCheckedExtern(name string) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 6 {
			return Value{}, nil, fmt.Errorf("%s expects 6 arguments", name)
		}
		if !isPointerType(args[0].Type) || !isIntegerLike(args[1].Type) || !isIntegerLike(args[2].Type) || !isIntegerLike(args[3].Type) || !isPointerType(args[4].Type) || !isPointerType(args[5].Type) {
			return Value{}, nil, fmt.Errorf("%s expects destination, size, flag, object-size, format, and va_list arguments", name)
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("%s requires memory", name)
		}
		size, err := memorySizeArg(name, args[1])
		if err != nil {
			return Value{}, nil, err
		}
		if err := checkObjectSize(name, uint64(size), args[3]); err != nil {
			return Value{}, nil, err
		}
		out, err := formatCString(name, ec.Memory, args[4].Int, nil)
		if err != nil {
			return Value{}, nil, err
		}
		if size > 0 {
			n := len(out)
			if n > int(size)-1 {
				n = int(size) - 1
			}
			data := make([]byte, n+1)
			copy(data, out[:n])
			if err := writeMemoryBytes(ec.Memory, args[0].Int, data); err != nil {
				return Value{}, nil, err
			}
		}
		return IntValue(bytecode.TypeI32, int64(len(out))), nil, nil
	}
}

func printfExtern(name string, r *ExternRegistry) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) < 1 {
			return Value{}, nil, fmt.Errorf("%s expects at least 1 argument", name)
		}
		if !isPointerType(args[0].Type) {
			return Value{}, nil, fmt.Errorf("%s expects format pointer", name)
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("%s requires memory", name)
		}
		out, err := formatCString(name, ec.Memory, args[0].Int, args[1:])
		if err != nil {
			return Value{}, nil, err
		}
		if _, err := fmt.Fprint(r.externStdout(ec), out); err != nil {
			return Value{}, nil, err
		}
		return IntValue(bytecode.TypeI32, int64(len(out))), nil, nil
	}
}

func vprintfExtern(name string, r *ExternRegistry) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 2 {
			return Value{}, nil, fmt.Errorf("%s expects 2 arguments", name)
		}
		if !isPointerType(args[0].Type) || !isPointerType(args[1].Type) {
			return Value{}, nil, fmt.Errorf("%s expects format and va_list pointers", name)
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("%s requires memory", name)
		}
		out, err := formatCString(name, ec.Memory, args[0].Int, nil)
		if err != nil {
			return Value{}, nil, err
		}
		if _, err := fmt.Fprint(r.externStdout(ec), out); err != nil {
			return Value{}, nil, err
		}
		return IntValue(bytecode.TypeI32, int64(len(out))), nil, nil
	}
}

func fprintfExtern(name string, r *ExternRegistry) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) < 2 {
			return Value{}, nil, fmt.Errorf("%s expects at least 2 arguments", name)
		}
		if !isPointerType(args[0].Type) || !isPointerType(args[1].Type) {
			return Value{}, nil, fmt.Errorf("%s expects stream and format pointers", name)
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("%s requires memory", name)
		}
		w, ok := r.lookupHostWriter(args[0].Int)
		if !ok {
			return Value{}, nil, fmt.Errorf("unknown stream handle %#x", args[0].Int)
		}
		out, err := formatCString(name, ec.Memory, args[1].Int, args[2:])
		if err != nil {
			return Value{}, nil, err
		}
		if _, err := fmt.Fprint(w, out); err != nil {
			return Value{}, nil, err
		}
		return IntValue(bytecode.TypeI32, int64(len(out))), nil, nil
	}
}

func vfprintfExtern(name string, r *ExternRegistry) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 3 {
			return Value{}, nil, fmt.Errorf("%s expects 3 arguments", name)
		}
		if !isPointerType(args[0].Type) || !isPointerType(args[1].Type) || !isPointerType(args[2].Type) {
			return Value{}, nil, fmt.Errorf("%s expects stream, format, and va_list pointers", name)
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("%s requires memory", name)
		}
		w, ok := r.lookupHostWriter(args[0].Int)
		if !ok {
			return Value{}, nil, fmt.Errorf("unknown stream handle %#x", args[0].Int)
		}
		out, err := formatCString(name, ec.Memory, args[1].Int, nil)
		if err != nil {
			return Value{}, nil, err
		}
		if _, err := fmt.Fprint(w, out); err != nil {
			return Value{}, nil, err
		}
		return IntValue(bytecode.TypeI32, int64(len(out))), nil, nil
	}
}

func printfCheckedExtern(name string, r *ExternRegistry) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) < 2 {
			return Value{}, nil, fmt.Errorf("%s expects at least 2 arguments", name)
		}
		if !isIntegerLike(args[0].Type) || !isPointerType(args[1].Type) {
			return Value{}, nil, fmt.Errorf("%s expects flag and format arguments", name)
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("%s requires memory", name)
		}
		out, err := formatCString(name, ec.Memory, args[1].Int, args[2:])
		if err != nil {
			return Value{}, nil, err
		}
		if _, err := fmt.Fprint(r.externStdout(ec), out); err != nil {
			return Value{}, nil, err
		}
		return IntValue(bytecode.TypeI32, int64(len(out))), nil, nil
	}
}

func vprintfCheckedExtern(name string, r *ExternRegistry) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 3 {
			return Value{}, nil, fmt.Errorf("%s expects 3 arguments", name)
		}
		if !isIntegerLike(args[0].Type) || !isPointerType(args[1].Type) || !isPointerType(args[2].Type) {
			return Value{}, nil, fmt.Errorf("%s expects flag, format, and va_list arguments", name)
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("%s requires memory", name)
		}
		out, err := formatCString(name, ec.Memory, args[1].Int, nil)
		if err != nil {
			return Value{}, nil, err
		}
		if _, err := fmt.Fprint(r.externStdout(ec), out); err != nil {
			return Value{}, nil, err
		}
		return IntValue(bytecode.TypeI32, int64(len(out))), nil, nil
	}
}

func fprintfCheckedExtern(name string, r *ExternRegistry) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) < 3 {
			return Value{}, nil, fmt.Errorf("%s expects at least 3 arguments", name)
		}
		if !isPointerType(args[0].Type) || !isIntegerLike(args[1].Type) || !isPointerType(args[2].Type) {
			return Value{}, nil, fmt.Errorf("%s expects stream, flag, and format arguments", name)
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("%s requires memory", name)
		}
		w, ok := r.lookupHostWriter(args[0].Int)
		if !ok {
			return Value{}, nil, fmt.Errorf("unknown stream handle %#x", args[0].Int)
		}
		out, err := formatCString(name, ec.Memory, args[2].Int, args[3:])
		if err != nil {
			return Value{}, nil, err
		}
		if _, err := fmt.Fprint(w, out); err != nil {
			return Value{}, nil, err
		}
		return IntValue(bytecode.TypeI32, int64(len(out))), nil, nil
	}
}

func vfprintfCheckedExtern(name string, r *ExternRegistry) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 4 {
			return Value{}, nil, fmt.Errorf("%s expects 4 arguments", name)
		}
		if !isPointerType(args[0].Type) || !isIntegerLike(args[1].Type) || !isPointerType(args[2].Type) || !isPointerType(args[3].Type) {
			return Value{}, nil, fmt.Errorf("%s expects stream, flag, format, and va_list arguments", name)
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("%s requires memory", name)
		}
		w, ok := r.lookupHostWriter(args[0].Int)
		if !ok {
			return Value{}, nil, fmt.Errorf("unknown stream handle %#x", args[0].Int)
		}
		out, err := formatCString(name, ec.Memory, args[2].Int, nil)
		if err != nil {
			return Value{}, nil, err
		}
		if _, err := fmt.Fprint(w, out); err != nil {
			return Value{}, nil, err
		}
		return IntValue(bytecode.TypeI32, int64(len(out))), nil, nil
	}
}

func formatCString(name string, mem *Memory, formatAddr uint64, args []Value) (string, error) {
	format, err := mem.ReadCString(formatAddr)
	if err != nil {
		return "", err
	}
	var out strings.Builder
	argIndex := 0
	for i := 0; i < len(format); i++ {
		if format[i] != '%' {
			out.WriteByte(format[i])
			continue
		}
		i++
		if i >= len(format) {
			return "", fmt.Errorf("%s has trailing %% in format", name)
		}
		if format[i] == '%' {
			out.WriteByte('%')
			continue
		}
		leftAlign := false
		zeroPad := false
		showSign := false
		leadingSpace := false
		alternate := false
		for i < len(format) {
			switch format[i] {
			case '-':
				leftAlign = true
				i++
			case '0':
				zeroPad = true
				i++
			case '+':
				showSign = true
				i++
			case ' ':
				leadingSpace = true
				i++
			case '#':
				alternate = true
				i++
			default:
				goto width
			}
		}
	width:
		width := 0
		if i < len(format) && format[i] == '*' {
			if argIndex >= len(args) {
				return "", fmt.Errorf("%s format needs more arguments", name)
			}
			dynamicWidth, err := formatIntArg(name, "dynamic width", args[argIndex])
			if err != nil {
				return "", err
			}
			argIndex++
			if dynamicWidth < 0 {
				leftAlign = true
				width = -dynamicWidth
			} else {
				width = dynamicWidth
			}
			i++
		} else {
			for i < len(format) && format[i] >= '0' && format[i] <= '9' {
				width = width*10 + int(format[i]-'0')
				i++
			}
		}
		precision := -1
		if i < len(format) && format[i] == '.' {
			i++
			if i < len(format) && format[i] == '*' {
				if argIndex >= len(args) {
					return "", fmt.Errorf("%s format needs more arguments", name)
				}
				dynamicPrecision, err := formatIntArg(name, "dynamic precision", args[argIndex])
				if err != nil {
					return "", err
				}
				argIndex++
				if dynamicPrecision >= 0 {
					precision = dynamicPrecision
				}
				i++
			} else {
				precision = 0
				for i < len(format) && format[i] >= '0' && format[i] <= '9' {
					precision = precision*10 + int(format[i]-'0')
					i++
				}
			}
		}
		lengthMod := ""
		for i < len(format) {
			switch format[i] {
			case 'h', 'l':
				lengthMod = string(format[i])
				if i+1 < len(format) && format[i+1] == format[i] {
					lengthMod += string(format[i])
					i += 2
				} else {
					i++
				}
			case 'j', 'z', 't', 'L':
				lengthMod = string(format[i])
				i++
			default:
				goto verb
			}
		}
	verb:
		if i >= len(format) {
			return "", fmt.Errorf("%s has trailing %% in format", name)
		}
		if argIndex >= len(args) {
			return "", fmt.Errorf("%s format needs more arguments", name)
		}
		arg := args[argIndex]
		argIndex++
		piece := ""
		integerFormat := false
		switch format[i] {
		case 's':
			if !isPointerType(arg.Type) {
				return "", fmt.Errorf("%s %%s expects pointer argument", name)
			}
			s, err := mem.ReadCString(arg.Int)
			if err != nil {
				return "", err
			}
			piece = s
			if precision >= 0 && len(piece) > precision {
				piece = piece[:precision]
			}
		case 'd', 'i':
			if !isIntegerLike(arg.Type) {
				return "", fmt.Errorf("%s %%%c expects integer argument", name, format[i])
			}
			piece = strconv.FormatInt(signedInt(arg), 10)
			if !strings.HasPrefix(piece, "-") {
				if showSign {
					piece = "+" + piece
				} else if leadingSpace {
					piece = " " + piece
				}
			}
			integerFormat = true
		case 'u':
			if !isIntegerLike(arg.Type) {
				return "", fmt.Errorf("%s %%u expects integer argument", name)
			}
			piece = strconv.FormatUint(unsignedInt(arg), 10)
			integerFormat = true
		case 'x':
			if !isIntegerLike(arg.Type) {
				return "", fmt.Errorf("%s %%x expects integer argument", name)
			}
			piece = strconv.FormatUint(unsignedInt(arg), 16)
			if alternate && piece != "0" {
				piece = "0x" + piece
			}
			integerFormat = true
		case 'X':
			if !isIntegerLike(arg.Type) {
				return "", fmt.Errorf("%s %%X expects integer argument", name)
			}
			piece = strings.ToUpper(strconv.FormatUint(unsignedInt(arg), 16))
			if alternate && piece != "0" {
				piece = "0X" + piece
			}
			integerFormat = true
		case 'o':
			if !isIntegerLike(arg.Type) {
				return "", fmt.Errorf("%s %%o expects integer argument", name)
			}
			piece = strconv.FormatUint(unsignedInt(arg), 8)
			if alternate && !strings.HasPrefix(piece, "0") {
				piece = "0" + piece
			}
			integerFormat = true
		case 'p':
			if !isPointerType(arg.Type) {
				return "", fmt.Errorf("%s %%p expects pointer argument", name)
			}
			if precision >= 0 {
				return "", fmt.Errorf("%s %%p does not support precision", name)
			}
			piece = "0x" + strconv.FormatUint(arg.Int, 16)
		case 'c':
			if !isIntegerLike(arg.Type) {
				return "", fmt.Errorf("%s %%c expects integer argument", name)
			}
			if precision >= 0 {
				return "", fmt.Errorf("%s %%c does not support precision", name)
			}
			piece = string([]byte{byte(unsignedInt(arg))})
		case 'n':
			if !isPointerType(arg.Type) {
				return "", fmt.Errorf("%s %%n expects pointer argument", name)
			}
			if leftAlign || zeroPad || showSign || leadingSpace || alternate || width != 0 || precision >= 0 {
				return "", fmt.Errorf("%s %%n does not support flags, width, or precision", name)
			}
			countType, countAlign := writeCountType(lengthMod)
			if err := mem.Store(arg.Int, countType, countAlign, IntValue(countType, int64(out.Len()))); err != nil {
				return "", err
			}
			continue
		case 'f', 'F', 'e', 'E', 'g', 'G', 'a', 'A':
			if !isFloatLike(arg.Type) {
				return "", fmt.Errorf("%s %%%c expects floating argument", name, format[i])
			}
			verb := format[i]
			if verb == 'F' {
				verb = 'f'
			}
			if verb == 'a' {
				verb = 'x'
			}
			if verb == 'A' {
				verb = 'X'
			}
			floatPrecision := precision
			if floatPrecision < 0 {
				floatPrecision = 6
			}
			piece = strconv.FormatFloat(cvmFloat(arg), verb, floatPrecision, floatFormatBits(arg.Type))
			if format[i] == 'F' {
				piece = strings.ToUpper(piece)
			}
			if !strings.HasPrefix(piece, "-") && !strings.HasPrefix(piece, "N") && !strings.HasPrefix(piece, "I") {
				if showSign {
					piece = "+" + piece
				} else if leadingSpace {
					piece = " " + piece
				}
			}
		default:
			return "", fmt.Errorf("%s unsupported format %%%c", name, format[i])
		}
		if integerFormat && precision >= 0 {
			piece = applyIntegerPrecision(piece, precision)
			zeroPad = false
		}
		writeFormattedPiece(&out, piece, width, leftAlign, zeroPad)
	}
	return out.String(), nil
}

func applyIntegerPrecision(piece string, precision int) string {
	prefix, digits := splitNumericPrefix(piece)
	if precision == 0 && digits == "0" {
		digits = ""
	}
	if len(digits) < precision {
		digits = strings.Repeat("0", precision-len(digits)) + digits
	}
	return prefix + digits
}

func formatIntArg(name, what string, v Value) (int, error) {
	if !isIntegerLike(v.Type) {
		return 0, fmt.Errorf("%s %s expects integer argument", name, what)
	}
	n := signedInt(v)
	if n > int64(maxInt()) || n < -int64(maxInt()) {
		return 0, fmt.Errorf("%s %s %d exceeds int range", name, what, n)
	}
	return int(n), nil
}

func writeCountType(lengthMod string) (bytecode.ValueType, int64) {
	switch lengthMod {
	case "hh":
		return bytecode.TypeI8, 1
	case "h":
		return bytecode.TypeI16, 2
	case "l", "ll", "j", "z", "t":
		return bytecode.TypeI64, 8
	default:
		return bytecode.TypeI32, 4
	}
}

func isFloatLike(t bytecode.ValueType) bool {
	switch t {
	case bytecode.TypeF32, bytecode.TypeF64, bytecode.TypeFLong:
		return true
	default:
		return false
	}
}

func floatFormatBits(t bytecode.ValueType) int {
	if t == bytecode.TypeF32 {
		return 32
	}
	return 64
}

func writeFormattedPiece(out *strings.Builder, piece string, width int, leftAlign, zeroPad bool) {
	pad := width - len(piece)
	if pad <= 0 {
		out.WriteString(piece)
		return
	}
	padByte := byte(' ')
	if zeroPad && !leftAlign {
		padByte = '0'
	}
	if leftAlign {
		out.WriteString(piece)
		writeRepeatedByte(out, padByte, pad)
		return
	}
	if padByte == '0' {
		prefix, digits := splitNumericPrefix(piece)
		if prefix != "" {
			out.WriteString(prefix)
			writeRepeatedByte(out, padByte, pad)
			out.WriteString(digits)
			return
		}
	}
	writeRepeatedByte(out, padByte, pad)
	out.WriteString(piece)
}

func splitNumericPrefix(piece string) (string, string) {
	for _, prefix := range []string{"-", "+", " ", "0x", "0X"} {
		if strings.HasPrefix(piece, prefix) {
			return prefix, piece[len(prefix):]
		}
	}
	return "", piece
}

func writeRepeatedByte(out *strings.Builder, b byte, n int) {
	for i := 0; i < n; i++ {
		out.WriteByte(b)
	}
}

func complexAbsExtern(name string, realType bytecode.ValueType, realSize uint64) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 1 {
			return Value{}, nil, fmt.Errorf("%s expects 1 argument", name)
		}
		z, err := loadComplexArg(name, ec, args[0], realType, realSize)
		if err != nil {
			return Value{}, nil, err
		}
		return FloatValue(realType, math.Hypot(real(z), imag(z))), nil, nil
	}
}

func registerTgmathComplexExterns(r *ExternRegistry, base string, fn func(complex128) complex128) {
	r.Register(base+"f", complexUnaryExtern(base+"f", bytecode.TypeF32, 4, fn))
	r.Register(base, complexUnaryExtern(base, bytecode.TypeF64, 8, fn))
	r.Register(base+"l", complexUnaryExtern(base+"l", bytecode.TypeFLong, 16, fn))
}

func registerTgmathComplexBinaryExterns(r *ExternRegistry, base string, fn func(complex128, complex128) complex128) {
	r.Register(base+"f", complexBinaryExtern(base+"f", bytecode.TypeF32, 4, fn))
	r.Register(base, complexBinaryExtern(base, bytecode.TypeF64, 8, fn))
	r.Register(base+"l", complexBinaryExtern(base+"l", bytecode.TypeFLong, 16, fn))
}

func registerTgmathComplexRealExterns(r *ExternRegistry, base string, fn func(complex128) float64) {
	r.Register(base+"f", complexRealExtern(base+"f", bytecode.TypeF32, 4, fn))
	r.Register(base, complexRealExtern(base, bytecode.TypeF64, 8, fn))
	r.Register(base+"l", complexRealExtern(base+"l", bytecode.TypeFLong, 16, fn))
}

func complexUnaryExtern(name string, realType bytecode.ValueType, realSize uint64, fn func(complex128) complex128) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 1 {
			return Value{}, nil, fmt.Errorf("%s expects 1 argument", name)
		}
		z, err := loadComplexArg(name, ec, args[0], realType, realSize)
		if err != nil {
			return Value{}, nil, err
		}
		addr, err := storeComplexResult(name, ec, realType, realSize, fn(z))
		if err != nil {
			return Value{}, nil, err
		}
		return ObjectAddrValue(addr), nil, nil
	}
}

func complexRealExtern(name string, realType bytecode.ValueType, realSize uint64, fn func(complex128) float64) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 1 {
			return Value{}, nil, fmt.Errorf("%s expects 1 argument", name)
		}
		z, err := loadComplexArg(name, ec, args[0], realType, realSize)
		if err != nil {
			return Value{}, nil, err
		}
		return FloatValue(realType, fn(z)), nil, nil
	}
}

func complexBinaryExtern(name string, realType bytecode.ValueType, realSize uint64, fn func(complex128, complex128) complex128) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 2 {
			return Value{}, nil, fmt.Errorf("%s expects 2 arguments", name)
		}
		left, err := loadComplexArg(name, ec, args[0], realType, realSize)
		if err != nil {
			return Value{}, nil, err
		}
		right, err := loadComplexOrRealArg(name, ec, args[1], realType, realSize)
		if err != nil {
			return Value{}, nil, err
		}
		addr, err := storeComplexResult(name, ec, realType, realSize, fn(left, right))
		if err != nil {
			return Value{}, nil, err
		}
		return ObjectAddrValue(addr), nil, nil
	}
}

func loadComplexArg(name string, ec *ExternContext, arg Value, realType bytecode.ValueType, realSize uint64) (complex128, error) {
	if arg.Type != bytecode.TypeObjectAddr {
		return 0, fmt.Errorf("%s expects complex object address", name)
	}
	if ec == nil || ec.Memory == nil {
		return 0, fmt.Errorf("%s requires memory", name)
	}
	if arg.Int > math.MaxUint64-realSize {
		return 0, fmt.Errorf("%s complex imaginary component address overflows", name)
	}
	realPart, err := ec.Memory.Load(arg.Int, realType, 1)
	if err != nil {
		return 0, err
	}
	imagPart, err := ec.Memory.Load(arg.Int+realSize, realType, 1)
	if err != nil {
		return 0, err
	}
	return complex(cvmFloat(realPart), cvmFloat(imagPart)), nil
}

func loadComplexOrRealArg(name string, ec *ExternContext, arg Value, realType bytecode.ValueType, realSize uint64) (complex128, error) {
	if arg.Type == bytecode.TypeObjectAddr {
		return loadComplexArg(name, ec, arg, realType, realSize)
	}
	if !isFloatType(arg.Type) {
		return 0, fmt.Errorf("%s expects complex or floating argument", name)
	}
	return complex(cvmFloat(arg), 0), nil
}

func storeComplexResult(name string, ec *ExternContext, realType bytecode.ValueType, realSize uint64, z complex128) (uint64, error) {
	if ec == nil || ec.Memory == nil {
		return 0, fmt.Errorf("%s requires memory", name)
	}
	if realSize > math.MaxInt64/2 {
		return 0, fmt.Errorf("%s complex result size overflows", name)
	}
	addr, err := ec.Memory.TryAlloc("extern:"+name+":result", int64(realSize*2), int64(realSize), false, blockGlobal)
	if err != nil {
		return 0, err
	}
	if err := ec.Memory.Store(addr, realType, int64(realSize), FloatValue(realType, real(z))); err != nil {
		return 0, err
	}
	if err := ec.Memory.Store(addr+realSize, realType, int64(realSize), FloatValue(realType, imag(z))); err != nil {
		return 0, err
	}
	return addr, nil
}

func cvmComplexProject(z complex128) complex128 {
	if math.IsInf(real(z), 0) || math.IsInf(imag(z), 0) {
		return complex(math.Inf(1), math.Copysign(0, imag(z)))
	}
	return z
}

func registerTgmathRealExterns(r *ExternRegistry, base string, fn func(float64) float64) {
	r.Register(base+"f", mathUnaryFloatExtern(base+"f", bytecode.TypeF32, fn))
	r.Register(base, mathUnaryFloatExtern(base, bytecode.TypeF64, fn))
	r.Register(base+"l", mathUnaryFloatExtern(base+"l", bytecode.TypeFLong, fn))
}

func registerTgmathIntExterns(r *ExternRegistry, base string, fn func(float64) int) {
	r.Register(base+"f", mathUnaryIntExtern(base+"f", fn))
	r.Register(base, mathUnaryIntExtern(base, fn))
	r.Register(base+"l", mathUnaryIntExtern(base+"l", fn))
}

func registerTgmathLongExterns(r *ExternRegistry, base string, fn func(float64) float64) {
	r.Register(base+"f", mathUnaryLongExtern(base+"f", fn))
	r.Register(base, mathUnaryLongExtern(base, fn))
	r.Register(base+"l", mathUnaryLongExtern(base+"l", fn))
}

func registerTgmathRealBinaryExterns(r *ExternRegistry, base string, fn func(float64, float64) float64) {
	r.Register(base+"f", mathBinaryFloatExtern(base+"f", bytecode.TypeF32, fn))
	r.Register(base, mathBinaryFloatExtern(base, bytecode.TypeF64, fn))
	r.Register(base+"l", mathBinaryFloatExtern(base+"l", bytecode.TypeFLong, fn))
}

func registerTgmathRealIntBinaryExterns(r *ExternRegistry, base string, fn func(float64, int) float64) {
	r.Register(base+"f", mathFloatIntExtern(base+"f", bytecode.TypeF32, fn))
	r.Register(base, mathFloatIntExtern(base, bytecode.TypeF64, fn))
	r.Register(base+"l", mathFloatIntExtern(base+"l", bytecode.TypeFLong, fn))
}

func registerTgmathFrexpExterns(r *ExternRegistry, base string) {
	r.Register(base+"f", mathFrexpExtern(base+"f", bytecode.TypeF32))
	r.Register(base, mathFrexpExtern(base, bytecode.TypeF64))
	r.Register(base+"l", mathFrexpExtern(base+"l", bytecode.TypeFLong))
}

func registerTgmathRemquoExterns(r *ExternRegistry, base string) {
	r.Register(base+"f", mathRemquoExtern(base+"f", bytecode.TypeF32))
	r.Register(base, mathRemquoExtern(base, bytecode.TypeF64))
	r.Register(base+"l", mathRemquoExtern(base+"l", bytecode.TypeFLong))
}

func registerTgmathRealTernaryExterns(r *ExternRegistry, base string, fn func(float64, float64, float64) float64) {
	r.Register(base+"f", mathTernaryFloatExtern(base+"f", bytecode.TypeF32, fn))
	r.Register(base, mathTernaryFloatExtern(base, bytecode.TypeF64, fn))
	r.Register(base+"l", mathTernaryFloatExtern(base+"l", bytecode.TypeFLong, fn))
}

func mathUnaryFloatExtern(name string, ret bytecode.ValueType, fn func(float64) float64) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 1 {
			return Value{}, nil, fmt.Errorf("%s expects 1 argument", name)
		}
		if !isFloatType(args[0].Type) {
			return Value{}, nil, fmt.Errorf("%s expects floating argument", name)
		}
		return floatResult(ret, fn(cvmFloat(args[0]))), nil, nil
	}
}

func mathUnaryIntExtern(name string, fn func(float64) int) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 1 {
			return Value{}, nil, fmt.Errorf("%s expects 1 argument", name)
		}
		if !isFloatType(args[0].Type) {
			return Value{}, nil, fmt.Errorf("%s expects floating argument", name)
		}
		return IntValue(bytecode.TypeI32, int64(fn(cvmFloat(args[0])))), nil, nil
	}
}

func mathUnaryLongExtern(name string, fn func(float64) float64) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 1 {
			return Value{}, nil, fmt.Errorf("%s expects 1 argument", name)
		}
		if !isFloatType(args[0].Type) {
			return Value{}, nil, fmt.Errorf("%s expects floating argument", name)
		}
		return IntValue(bytecode.TypeI64, int64(fn(cvmFloat(args[0])))), nil, nil
	}
}

func mathConstantFloatExtern(name string, ret bytecode.ValueType, value float64) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 0 {
			return Value{}, nil, fmt.Errorf("%s expects 0 arguments", name)
		}
		return floatResult(ret, value), nil, nil
	}
}

func mathNanExtern(name string) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 1 {
			return Value{}, nil, fmt.Errorf("%s expects 1 argument", name)
		}
		if !isPointerType(args[0].Type) {
			return Value{}, nil, fmt.Errorf("%s expects string argument", name)
		}
		return FloatValue(bytecode.TypeF64, math.NaN()), nil, nil
	}
}

func mathBinaryFloatExtern(name string, ret bytecode.ValueType, fn func(float64, float64) float64) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 2 {
			return Value{}, nil, fmt.Errorf("%s expects 2 arguments", name)
		}
		if !isFloatType(args[0].Type) || !isFloatType(args[1].Type) {
			return Value{}, nil, fmt.Errorf("%s expects floating arguments", name)
		}
		return floatResult(ret, fn(cvmFloat(args[0]), cvmFloat(args[1]))), nil, nil
	}
}

func mathFloatIntExtern(name string, ret bytecode.ValueType, fn func(float64, int) float64) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 2 {
			return Value{}, nil, fmt.Errorf("%s expects 2 arguments", name)
		}
		if !isFloatType(args[0].Type) || !isIntegerLike(args[1].Type) {
			return Value{}, nil, fmt.Errorf("%s expects floating and integer arguments", name)
		}
		exp := signedInt(args[1])
		if int64(int(exp)) != exp {
			return Value{}, nil, fmt.Errorf("%s exponent %d exceeds int range", name, exp)
		}
		return floatResult(ret, fn(cvmFloat(args[0]), int(exp))), nil, nil
	}
}

func mathFrexpExtern(name string, ret bytecode.ValueType) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 2 {
			return Value{}, nil, fmt.Errorf("%s expects 2 arguments", name)
		}
		if !isFloatType(args[0].Type) || !isPointerType(args[1].Type) {
			return Value{}, nil, fmt.Errorf("%s expects floating and pointer arguments", name)
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("%s requires memory", name)
		}
		frac, exp := math.Frexp(cvmFloat(args[0]))
		if err := ec.Memory.Store(args[1].Int, bytecode.TypeI32, 4, IntValue(bytecode.TypeI32, int64(exp))); err != nil {
			return Value{}, nil, err
		}
		return floatResult(ret, frac), nil, nil
	}
}

func mathRemquoExtern(name string, ret bytecode.ValueType) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 3 {
			return Value{}, nil, fmt.Errorf("%s expects 3 arguments", name)
		}
		if !isFloatType(args[0].Type) || !isFloatType(args[1].Type) || !isPointerType(args[2].Type) {
			return Value{}, nil, fmt.Errorf("%s expects floating, floating, and pointer arguments", name)
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("%s requires memory", name)
		}
		x, y := cvmFloat(args[0]), cvmFloat(args[1])
		quo := int64(math.RoundToEven(x / y))
		if err := ec.Memory.Store(args[2].Int, bytecode.TypeI32, 4, IntValue(bytecode.TypeI32, quo)); err != nil {
			return Value{}, nil, err
		}
		return floatResult(ret, math.Remainder(x, y)), nil, nil
	}
}

func mathTernaryFloatExtern(name string, ret bytecode.ValueType, fn func(float64, float64, float64) float64) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 3 {
			return Value{}, nil, fmt.Errorf("%s expects 3 arguments", name)
		}
		if !isFloatType(args[0].Type) || !isFloatType(args[1].Type) || !isFloatType(args[2].Type) {
			return Value{}, nil, fmt.Errorf("%s expects floating arguments", name)
		}
		return floatResult(ret, fn(cvmFloat(args[0]), cvmFloat(args[1]), cvmFloat(args[2]))), nil, nil
	}
}

func mathUnaryExtern(name string, pred func(Value) int64) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 1 {
			return Value{}, nil, fmt.Errorf("%s expects 1 argument", name)
		}
		if !isFloatType(args[0].Type) {
			return Value{}, nil, fmt.Errorf("%s expects floating argument", name)
		}
		return IntValue(bytecode.TypeI32, pred(args[0])), nil, nil
	}
}

const (
	fpClassNaN       = 0
	fpClassInfinite  = 1
	fpClassNormal    = 2
	fpClassSubnormal = 3
	fpClassZero      = 4

	minNormalFloat32 = 1.1754943508222875e-38
	minNormalFloat64 = 2.2250738585072014e-308
)

func cvmFPClassify(v Value) int64 {
	f := cvmFloat(v)
	switch {
	case math.IsNaN(f):
		return fpClassNaN
	case math.IsInf(f, 0):
		return fpClassInfinite
	case f == 0:
		return fpClassZero
	}
	minNormal := minNormalFloat64
	if v.Type == bytecode.TypeF32 {
		minNormal = minNormalFloat32
	}
	if math.Abs(f) < minNormal {
		return fpClassSubnormal
	}
	return fpClassNormal
}

func cvmFloat(v Value) float64 {
	if v.Type == bytecode.TypeF32 {
		return float64(float32(v.Float))
	}
	return v.Float
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

func strncmpResult(left, right string, n int64) int {
	for i := int64(0); i < n; i++ {
		var lb, rb byte
		if i < int64(len(left)) {
			lb = left[i]
		}
		if i < int64(len(right)) {
			rb = right[i]
		}
		if lb < rb {
			return -1
		}
		if lb > rb {
			return 1
		}
		if lb == 0 {
			return 0
		}
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
	case "stdin":
		addr, err := r.allocHostWriter(name, mem, io.Discard, 0)
		if err == nil {
			r.stdinHandle = addr
		}
		return addr, true, err
	case "stdout":
		addr, err := r.allocHostWriter(name, mem, r.stdout, 1)
		return addr, true, err
	case "stderr":
		addr, err := r.allocHostWriter(name, mem, r.stderr, 2)
		return addr, true, err
	case "errno":
		addr, err := r.staticI32Variable(mem, name, 0)
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

func (r *ExternRegistry) externStderr(ec *ExternContext) io.Writer {
	if ec != nil && ec.Stderr != nil {
		return ec.Stderr
	}
	return r.stderr
}

func (r *ExternRegistry) staticCString(mem *Memory, name, value string) (uint64, error) {
	if mem == nil {
		return 0, fmt.Errorf("memory is nil")
	}
	if byName := r.staticStrings[mem]; byName != nil {
		if addr, ok := byName[name]; ok {
			return addr, nil
		}
	}
	data := append([]byte(value), 0)
	addr, err := mem.TryAlloc("extern:"+name, int64(len(data)), 1, false, blockString)
	if err != nil {
		return 0, err
	}
	if err := writeMemoryBytes(mem, addr, data); err != nil {
		return 0, err
	}
	if r.staticStrings[mem] == nil {
		r.staticStrings[mem] = make(map[string]uint64)
	}
	r.staticStrings[mem][name] = addr
	return addr, nil
}

func (r *ExternRegistry) staticI32Variable(mem *Memory, name string, initial int32) (uint64, error) {
	if mem == nil {
		return 0, fmt.Errorf("memory is nil")
	}
	if byName := r.staticVars[mem]; byName != nil {
		if addr, ok := byName[name]; ok {
			return addr, nil
		}
	}
	addr, err := mem.TryAlloc("extern:"+name, 4, 4, false, blockGlobal)
	if err != nil {
		return 0, err
	}
	if err := mem.Store(addr, bytecode.TypeI32, 4, IntValue(bytecode.TypeI32, int64(initial))); err != nil {
		return 0, err
	}
	if r.staticVars[mem] == nil {
		r.staticVars[mem] = make(map[string]uint64)
	}
	r.staticVars[mem][name] = addr
	return addr, nil
}

func (r *ExternRegistry) lookupHostWriter(addr uint64) (io.Writer, bool) {
	w, ok := r.hostWriters[addr]
	return w, ok
}

func (r *ExternRegistry) readHostChar(addr uint64) (byte, bool) {
	buf := r.hostPushback[addr]
	if len(buf) == 0 {
		return 0, false
	}
	ch := buf[len(buf)-1]
	buf = buf[:len(buf)-1]
	if len(buf) == 0 {
		delete(r.hostPushback, addr)
	} else {
		r.hostPushback[addr] = buf
	}
	return ch, true
}

func (r *ExternRegistry) allocHostWriter(name string, mem *Memory, w io.Writer, fd int32) (uint64, error) {
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
	r.hostFDs[addr] = fd
	return addr, nil
}
