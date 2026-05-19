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
	funcs          map[string]ExternFunc
	stdin          io.Reader
	stdout         io.Writer
	stderr         io.Writer
	hostWriters    map[uint64]io.Writer
	hostFDs        map[uint64]int32
	hostPushback   map[uint64][]byte
	hostEOF        map[uint64]bool
	hostError      map[uint64]bool
	hostClosed     map[uint64]bool
	hostFiles      map[uint64]*hostFile
	files          map[string][]byte
	env            map[string]string
	atexitHandlers []uint64
	stdinHandle    uint64
	staticStrings  map[*Memory]map[string]uint64
	staticVars     map[*Memory]map[string]uint64
	staticBlocks   map[*Memory]map[string]uint64
	strtokNext     map[*Memory]uint64
	randSeed       uint32
	tmpnamCounter  uint64
}

type hostFile struct {
	path       string
	data       []byte
	pos        int64
	readable   bool
	writable   bool
	appendMode bool
	updateMode bool
	lastOp     hostFileOp
}

type hostFileOp int

const (
	hostFileOpNone hostFileOp = iota
	hostFileOpRead
	hostFileOpWrite
)

type hostFileWriter struct {
	registry *ExternRegistry
	addr     uint64
}

func NewExternRegistry(stdout, stderr io.Writer) *ExternRegistry {
	return NewExternRegistryWithIO(nil, stdout, stderr)
}

func NewExternRegistryWithIO(stdin io.Reader, stdout, stderr io.Writer) *ExternRegistry {
	if stdout == nil {
		stdout = os.Stdout
	}
	if stderr == nil {
		stderr = os.Stderr
	}
	return &ExternRegistry{
		funcs:         make(map[string]ExternFunc),
		stdin:         stdin,
		stdout:        stdout,
		stderr:        stderr,
		hostWriters:   make(map[uint64]io.Writer),
		hostFDs:       make(map[uint64]int32),
		hostPushback:  make(map[uint64][]byte),
		hostEOF:       make(map[uint64]bool),
		hostError:     make(map[uint64]bool),
		hostClosed:    make(map[uint64]bool),
		hostFiles:     make(map[uint64]*hostFile),
		files:         make(map[string][]byte),
		env:           make(map[string]string),
		staticStrings: make(map[*Memory]map[string]uint64),
		staticVars:    make(map[*Memory]map[string]uint64),
		staticBlocks:  make(map[*Memory]map[string]uint64),
		strtokNext:    make(map[*Memory]uint64),
		randSeed:      1,
	}
}

func DefaultExternRegistry(stdout, stderr io.Writer) *ExternRegistry {
	return DefaultExternRegistryWithIO(nil, stdout, stderr)
}

func DefaultExternRegistryWithIO(stdin io.Reader, stdout, stderr io.Writer) *ExternRegistry {
	r := NewExternRegistryWithIO(stdin, stdout, stderr)
	r.Register("exit", exitExtern("exit"))
	r.Register("_Exit", exitExtern("_Exit"))
	r.Register("abort", abortExtern())
	r.Register("__builtin_abort", abortExtern())
	registerVaListExterns(r)
	r.Register("remove", removeExtern("remove", r))
	r.Register("rename", renameExtern("rename", r))
	r.Register("fopen", fopenExtern("fopen", r))
	r.Register("freopen", freopenExtern("freopen", r))
	r.Register("tmpfile", tmpfileExtern("tmpfile", r))
	r.Register("tmpnam", tmpnamExtern("tmpnam", r))
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
	r.Register("atof", atofExtern("atof", r))
	r.Register("strtol", strtoIntegerExtern("strtol", bytecode.TypeI64, true))
	r.Register("strtoul", strtoIntegerExtern("strtoul", bytecode.TypeU64, false))
	r.Register("strtoll", strtoIntegerExtern("strtoll", bytecode.TypeI64, true))
	r.Register("strtoull", strtoIntegerExtern("strtoull", bytecode.TypeU64, false))
	r.Register("strtod", strtoFloatExtern("strtod", bytecode.TypeF64, r))
	r.Register("strtof", strtoFloatExtern("strtof", bytecode.TypeF32, r))
	r.Register("strtold", strtoFloatExtern("strtold", bytecode.TypeFLong, r))
	r.Register("mblen", mblenExtern("mblen"))
	r.Register("mbtowc", mbtowcExtern("mbtowc"))
	r.Register("wctomb", wctombExtern("wctomb"))
	r.Register("mbstowcs", mbstowcsExtern("mbstowcs"))
	r.Register("wcstombs", wcstombsExtern("wcstombs"))
	r.Register("mbrlen", mbrlenExtern("mbrlen"))
	r.Register("mbrtowc", mbrtowcExtern("mbrtowc"))
	r.Register("wcrtomb", wcrtombExtern("wcrtomb"))
	r.Register("rand", randExtern("rand", r))
	r.Register("srand", srandExtern("srand", r))
	r.Register("getenv", getenvExtern("getenv", r))
	r.Register("system", systemExtern("system"))
	r.Register("atexit", atexitExtern("atexit", r))
	r.Register("setlocale", setlocaleExtern("setlocale", r))
	r.Register("localeconv", localeconvExtern("localeconv", r))
	r.Register("clock", clockExtern("clock"))
	r.Register("difftime", difftimeExtern("difftime"))
	r.Register("time", timeExtern("time"))
	registerCtypeClassificationExterns(r)
	registerCtypeCaseExterns(r)
	registerWideCtypeClassificationExterns(r)
	registerWideCtypeCaseExterns(r)
	registerWideCtypeDescriptorExterns(r)
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
	registerInputFormatExterns(r)
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

func (r *ExternRegistry) AddFile(path string, data []byte) {
	if r == nil {
		return
	}
	r.files[path] = append([]byte(nil), data...)
}

func (r *ExternRegistry) SetEnv(name, value string) {
	if r == nil {
		return
	}
	r.env[name] = value
}

func (r *ExternRegistry) registerAtexitHandler(addr uint64) {
	if r == nil {
		return
	}
	r.atexitHandlers = append(r.atexitHandlers, addr)
}

func (r *ExternRegistry) takeAtexitHandlers() []uint64 {
	if r == nil || len(r.atexitHandlers) == 0 {
		return nil
	}
	handlers := append([]uint64(nil), r.atexitHandlers...)
	r.atexitHandlers = nil
	return handlers
}

func (w hostFileWriter) Write(p []byte) (int, error) {
	if w.registry == nil {
		return 0, fmt.Errorf("file writer has no registry")
	}
	file := w.registry.hostFiles[w.addr]
	if file == nil {
		w.registry.hostError[w.addr] = true
		return 0, fmt.Errorf("unknown file handle %#x", w.addr)
	}
	if !file.writable {
		w.registry.hostError[w.addr] = true
		return 0, fmt.Errorf("file %q is not open for writing", file.path)
	}
	if file.updateMode && file.lastOp == hostFileOpRead && !w.registry.hostEOF[w.addr] {
		w.registry.hostError[w.addr] = true
		return 0, fmt.Errorf("file %q needs positioning before writing after reading", file.path)
	}
	if file.appendMode {
		file.pos = int64(len(file.data))
	}
	if file.pos < 0 {
		w.registry.hostError[w.addr] = true
		return 0, fmt.Errorf("file %q has negative offset", file.path)
	}
	end := file.pos + int64(len(p))
	if end < file.pos {
		w.registry.hostError[w.addr] = true
		return 0, fmt.Errorf("file %q write offset overflows", file.path)
	}
	if end > int64(maxInt()) {
		w.registry.hostError[w.addr] = true
		return 0, fmt.Errorf("file %q write offset exceeds host int range", file.path)
	}
	if end > int64(len(file.data)) {
		grown := make([]byte, int(end))
		copy(grown, file.data)
		file.data = grown
	}
	copy(file.data[file.pos:end], p)
	file.pos = end
	if file.path != "" {
		w.registry.files[file.path] = append([]byte(nil), file.data...)
	}
	file.lastOp = hostFileOpWrite
	delete(w.registry.hostEOF, w.addr)
	return len(p), nil
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

func removeExtern(name string, r *ExternRegistry) ExternFunc {
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
		path, err := ec.Memory.ReadCString(args[0].Int)
		if err != nil {
			return Value{}, nil, err
		}
		if _, ok := r.files[path]; ok {
			delete(r.files, path)
			return IntValue(bytecode.TypeI32, 0), nil, nil
		}
		return IntValue(bytecode.TypeI32, -1), nil, nil
	}
}

func renameExtern(name string, r *ExternRegistry) ExternFunc {
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
		oldPath, err := ec.Memory.ReadCString(args[0].Int)
		if err != nil {
			return Value{}, nil, err
		}
		newPath, err := ec.Memory.ReadCString(args[1].Int)
		if err != nil {
			return Value{}, nil, err
		}
		data, ok := r.files[oldPath]
		if ok {
			r.files[newPath] = append([]byte(nil), data...)
			delete(r.files, oldPath)
			return IntValue(bytecode.TypeI32, 0), nil, nil
		}
		return IntValue(bytecode.TypeI32, -1), nil, nil
	}
}

func fopenExtern(name string, r *ExternRegistry) ExternFunc {
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
		path, err := ec.Memory.ReadCString(args[0].Int)
		if err != nil {
			return Value{}, nil, err
		}
		mode, err := ec.Memory.ReadCString(args[1].Int)
		if err != nil {
			return Value{}, nil, err
		}
		readable := strings.HasPrefix(mode, "r") || strings.Contains(mode, "+")
		writable := strings.HasPrefix(mode, "w") || strings.HasPrefix(mode, "a") || strings.Contains(mode, "+")
		if !readable && !writable {
			return PtrValue(0), nil, nil
		}
		data, ok := r.files[path]
		if strings.HasPrefix(mode, "r") && !ok {
			return PtrValue(0), nil, nil
		}
		file := &hostFile{
			path:       path,
			data:       append([]byte(nil), data...),
			readable:   readable,
			writable:   writable,
			appendMode: strings.HasPrefix(mode, "a"),
			updateMode: strings.Contains(mode, "+"),
		}
		if strings.HasPrefix(mode, "w") {
			file.data = nil
		}
		if strings.HasPrefix(mode, "a") {
			file.pos = int64(len(file.data))
		}
		addr, err := r.allocHostWriter("file:"+path, ec.Memory, hostFileWriter{registry: r}, -1)
		if err != nil {
			return Value{}, nil, err
		}
		r.hostFiles[addr] = file
		r.hostWriters[addr] = hostFileWriter{registry: r, addr: addr}
		if writable {
			r.files[path] = append([]byte(nil), file.data...)
		}
		delete(r.hostEOF, addr)
		delete(r.hostError, addr)
		return PtrValue(addr), nil, nil
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
		path, err := ec.Memory.ReadCString(args[0].Int)
		if err != nil {
			return Value{}, nil, err
		}
		mode, err := ec.Memory.ReadCString(args[1].Int)
		if err != nil {
			return Value{}, nil, err
		}
		if _, ok := r.lookupHostWriter(args[2].Int); !ok {
			return Value{}, nil, fmt.Errorf("unknown stream handle %#x", args[2].Int)
		}
		readMode := strings.HasPrefix(mode, "r")
		writeMode := strings.HasPrefix(mode, "w")
		appendMode := strings.HasPrefix(mode, "a")
		if !readMode && !writeMode && !appendMode {
			return PtrValue(0), nil, nil
		}
		updateMode := strings.Contains(mode, "+")
		readable := readMode || updateMode
		writable := writeMode || appendMode || updateMode
		data, ok := r.files[path]
		if readMode && !ok {
			return PtrValue(0), nil, nil
		}
		file := &hostFile{
			path:       path,
			data:       append([]byte(nil), data...),
			readable:   readable,
			writable:   writable,
			appendMode: appendMode,
			updateMode: updateMode,
		}
		if writeMode {
			file.data = nil
		}
		if appendMode {
			file.pos = int64(len(file.data))
		}
		if old := r.hostFiles[args[2].Int]; old != nil && old.writable && old.path != "" {
			r.files[old.path] = append([]byte(nil), old.data...)
		}
		r.hostFiles[args[2].Int] = file
		r.hostWriters[args[2].Int] = hostFileWriter{registry: r, addr: args[2].Int}
		if writable {
			r.files[path] = append([]byte(nil), file.data...)
		}
		delete(r.hostPushback, args[2].Int)
		delete(r.hostEOF, args[2].Int)
		delete(r.hostError, args[2].Int)
		return PtrValue(args[2].Int), nil, nil
	}
}

func tmpfileExtern(name string, r *ExternRegistry) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 0 {
			return Value{}, nil, fmt.Errorf("%s expects 0 arguments", name)
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("%s requires memory", name)
		}
		file := &hostFile{readable: true, writable: true, updateMode: true}
		addr, err := r.allocHostWriter("tmpfile", ec.Memory, hostFileWriter{registry: r}, -1)
		if err != nil {
			return Value{}, nil, err
		}
		r.hostFiles[addr] = file
		r.hostWriters[addr] = hostFileWriter{registry: r, addr: addr}
		return PtrValue(addr), nil, nil
	}
}

func tmpnamExtern(name string, r *ExternRegistry) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 1 {
			return Value{}, nil, fmt.Errorf("%s expects 1 argument", name)
		}
		if !isPointerType(args[0].Type) {
			return Value{}, nil, fmt.Errorf("%s expects buffer pointer", name)
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("%s requires memory", name)
		}
		tmpName := "/tmp/cvm-tmp-0"
		if args[0].Int != 0 {
			tmpName = fmt.Sprintf("/tmp/cvm-tmp-%d", r.tmpnamCounter)
			r.tmpnamCounter++
			if err := writeMemoryBytes(ec.Memory, args[0].Int, append([]byte(tmpName), 0)); err != nil {
				return Value{}, nil, err
			}
			return PtrValue(args[0].Int), nil, nil
		}
		addr, err := r.staticCString(ec.Memory, "tmpnam", tmpName)
		if err != nil {
			return Value{}, nil, err
		}
		return PtrValue(addr), nil, nil
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
			r.hostError[args[1].Int] = true
			return IntValue(bytecode.TypeI32, -1), nil, nil
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
		if r.markHostReadErrorIfUnreadable(args[0].Int) {
			return IntValue(bytecode.TypeI32, -1), nil, nil
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
		if r.markHostReadErrorIfUnreadable(args[2].Int) {
			return PtrValue(0), nil, nil
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
			r.hostError[args[1].Int] = true
			return IntValue(bytecode.TypeI32, -1), nil, nil
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
			if file := r.hostFiles[args[0].Int]; file != nil && file.updateMode {
				file.lastOp = hostFileOpNone
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
		r.hostClosed[args[0].Int] = true
		delete(r.hostWriters, args[0].Int)
		delete(r.hostFDs, args[0].Int)
		delete(r.hostPushback, args[0].Int)
		delete(r.hostEOF, args[0].Int)
		delete(r.hostError, args[0].Int)
		if file := r.hostFiles[args[0].Int]; file != nil && file.writable && file.path != "" {
			r.files[file.path] = append([]byte(nil), file.data...)
		}
		delete(r.hostFiles, args[0].Int)
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
		if file := r.hostFiles[args[0].Int]; file != nil {
			offset := signedInt(args[1])
			whence := signedInt(args[2])
			var next int64
			switch whence {
			case 0:
				next = offset
			case 1:
				next = file.pos + offset
			case 2:
				next = int64(len(file.data)) + offset
			default:
				return IntValue(bytecode.TypeI32, -1), nil, nil
			}
			if next < 0 {
				return IntValue(bytecode.TypeI32, -1), nil, nil
			}
			file.pos = next
			if file.updateMode {
				file.lastOp = hostFileOpNone
			}
			delete(r.hostPushback, args[0].Int)
			delete(r.hostEOF, args[0].Int)
			return IntValue(bytecode.TypeI32, 0), nil, nil
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
		if file := r.hostFiles[args[0].Int]; file != nil {
			return IntValue(bytecode.TypeI64, file.pos), nil, nil
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
		if file := r.hostFiles[args[0].Int]; file != nil {
			file.pos = 0
			if file.updateMode {
				file.lastOp = hostFileOpNone
			}
		}
		delete(r.hostPushback, args[0].Int)
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
		file := r.hostFiles[args[0].Int]
		if file == nil {
			return IntValue(bytecode.TypeI32, -1), nil, nil
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("%s requires memory", name)
		}
		if _, _, err := ec.Memory.rangeAccess(args[1].Int, int64(valueSize(ec.Memory.target, bytecode.TypeI64)), true); err != nil {
			return Value{}, nil, err
		}
		if err := ec.Memory.Store(args[1].Int, bytecode.TypeI64, 8, IntValue(bytecode.TypeI64, file.pos)); err != nil {
			return Value{}, nil, err
		}
		return IntValue(bytecode.TypeI32, 0), nil, nil
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
		file := r.hostFiles[args[0].Int]
		if file == nil {
			return IntValue(bytecode.TypeI32, -1), nil, nil
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("%s requires memory", name)
		}
		if _, _, err := ec.Memory.rangeAccess(args[1].Int, int64(valueSize(ec.Memory.target, bytecode.TypeI64)), false); err != nil {
			return Value{}, nil, err
		}
		pos, err := ec.Memory.Load(args[1].Int, bytecode.TypeI64, 8)
		if err != nil {
			return Value{}, nil, err
		}
		next := signedInt(pos)
		if next < 0 {
			return IntValue(bytecode.TypeI32, -1), nil, nil
		}
		file.pos = next
		if file.updateMode {
			file.lastOp = hostFileOpNone
		}
		delete(r.hostPushback, args[0].Int)
		delete(r.hostEOF, args[0].Int)
		return IntValue(bytecode.TypeI32, 0), nil, nil
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
		mode := signedInt(args[2])
		if mode < 0 || mode > 2 {
			return IntValue(bytecode.TypeI32, 1), nil, nil
		}
		size := unsignedInt(args[3])
		if args[1].Int != 0 && size != 0 {
			if size > uint64(maxInt()) {
				return Value{}, nil, fmt.Errorf("%s buffer size exceeds host int range", name)
			}
			if ec == nil || ec.Memory == nil {
				return Value{}, nil, fmt.Errorf("%s requires memory", name)
			}
			if _, _, err := ec.Memory.rangeAccess(args[1].Int, int64(size), true); err != nil {
				return Value{}, nil, err
			}
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

func atofExtern(name string, r *ExternRegistry) ExternFunc {
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
		if parsed.rangeErr {
			if err := r.setErrno(ec.Memory, 34); err != nil {
				return Value{}, nil, err
			}
		}
		return FloatValue(bytecode.TypeF64, parsed.value), nil, nil
	}
}

func strtoFloatExtern(name string, ret bytecode.ValueType, r *ExternRegistry) ExternFunc {
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
		value, rangeErr := normalizeStrtoFloatResult(ret, parsed.value, parsed.rangeErr)
		if rangeErr {
			if err := r.setErrno(ec.Memory, 34); err != nil {
				return Value{}, nil, err
			}
		}
		return FloatValue(ret, value), nil, nil
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
		if ch >= 0x80 {
			return IntValue(bytecode.TypeI32, -1), nil, nil
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
		if ch >= 0x80 {
			return IntValue(bytecode.TypeI32, -1), nil, nil
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
		if wc < 0 || wc > 0x7f {
			return IntValue(bytecode.TypeI32, -1), nil, nil
		}
		if err := writeMemoryByte(ec.Memory, args[0].Int, byte(wc)); err != nil {
			return Value{}, nil, err
		}
		return IntValue(bytecode.TypeI32, 1), nil, nil
	}
}

func mbrlenExtern(name string) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 3 {
			return Value{}, nil, fmt.Errorf("%s expects 3 arguments", name)
		}
		if args[0].Int == 0 {
			return UIntValue(bytecode.TypeU64, 0), nil, nil
		}
		if !isPointerType(args[0].Type) || !isIntegerLike(args[1].Type) || (args[2].Int != 0 && !isPointerType(args[2].Type)) {
			return Value{}, nil, fmt.Errorf("%s expects string, length, and state pointer arguments", name)
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("%s requires memory", name)
		}
		if unsignedInt(args[1]) == 0 {
			return UIntValue(bytecode.TypeU64, math.MaxUint64-1), nil, nil
		}
		ch, err := readMemoryByte(ec.Memory, args[0].Int)
		if err != nil {
			return Value{}, nil, err
		}
		if ch >= 0x80 {
			return UIntValue(bytecode.TypeU64, math.MaxUint64), nil, nil
		}
		if ch == 0 {
			return UIntValue(bytecode.TypeU64, 0), nil, nil
		}
		return UIntValue(bytecode.TypeU64, 1), nil, nil
	}
}

func mbrtowcExtern(name string) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 4 {
			return Value{}, nil, fmt.Errorf("%s expects 4 arguments", name)
		}
		if args[1].Int == 0 {
			return UIntValue(bytecode.TypeU64, 0), nil, nil
		}
		if (args[0].Int != 0 && !isPointerType(args[0].Type)) || !isPointerType(args[1].Type) || !isIntegerLike(args[2].Type) || (args[3].Int != 0 && !isPointerType(args[3].Type)) {
			return Value{}, nil, fmt.Errorf("%s expects destination, string, length, and state pointer arguments", name)
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("%s requires memory", name)
		}
		if unsignedInt(args[2]) == 0 {
			return UIntValue(bytecode.TypeU64, math.MaxUint64-1), nil, nil
		}
		ch, err := readMemoryByte(ec.Memory, args[1].Int)
		if err != nil {
			return Value{}, nil, err
		}
		if ch >= 0x80 {
			return UIntValue(bytecode.TypeU64, math.MaxUint64), nil, nil
		}
		if args[0].Int != 0 {
			if err := ec.Memory.Store(args[0].Int, bytecode.TypeI32, 4, IntValue(bytecode.TypeI32, int64(ch))); err != nil {
				return Value{}, nil, err
			}
		}
		if ch == 0 {
			return UIntValue(bytecode.TypeU64, 0), nil, nil
		}
		return UIntValue(bytecode.TypeU64, 1), nil, nil
	}
}

func wcrtombExtern(name string) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 3 {
			return Value{}, nil, fmt.Errorf("%s expects 3 arguments", name)
		}
		if args[0].Int == 0 {
			return UIntValue(bytecode.TypeU64, 1), nil, nil
		}
		if !isPointerType(args[0].Type) || !isIntegerLike(args[1].Type) || (args[2].Int != 0 && !isPointerType(args[2].Type)) {
			return Value{}, nil, fmt.Errorf("%s expects destination, wide char, and state pointer arguments", name)
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("%s requires memory", name)
		}
		wc := signedInt(args[1])
		if wc < 0 || wc > 0x7f {
			return UIntValue(bytecode.TypeU64, math.MaxUint64), nil, nil
		}
		if err := writeMemoryByte(ec.Memory, args[0].Int, byte(wc)); err != nil {
			return Value{}, nil, err
		}
		return UIntValue(bytecode.TypeU64, 1), nil, nil
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
		for i := 0; i < len(src); i++ {
			if src[i] >= 0x80 {
				return UIntValue(bytecode.TypeU64, ^uint64(0)), nil, nil
			}
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
		for _, wc := range chars {
			if wc > 0x7f {
				return UIntValue(bytecode.TypeU64, ^uint64(0)), nil, nil
			}
		}
		if args[0].Int == 0 {
			return UIntValue(bytecode.TypeU64, uint64(len(chars))), nil, nil
		}
		n := unsignedInt(args[2])
		count := uint64(0)
		for count < n && count < uint64(len(chars)) {
			wc := chars[count]
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

func getenvExtern(name string, r *ExternRegistry) ExternFunc {
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
		key, err := ec.Memory.ReadCString(args[0].Int)
		if err != nil {
			return Value{}, nil, err
		}
		value, ok := r.env[key]
		if !ok {
			return PtrValue(0), nil, nil
		}
		addr, err := r.staticCString(ec.Memory, "getenv:"+key+"="+value, value)
		if err != nil {
			return Value{}, nil, err
		}
		return PtrValue(addr), nil, nil
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

func atexitExtern(name string, r *ExternRegistry) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 1 {
			return Value{}, nil, fmt.Errorf("%s expects 1 argument", name)
		}
		if !isPointerType(args[0].Type) {
			return Value{}, nil, fmt.Errorf("%s expects function pointer", name)
		}
		r.registerAtexitHandler(args[0].Int)
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
		return Value{}, &ExitStatus{Code: code, skipAtexit: name == "_Exit"}, nil
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

func localeconvExtern(name string, r *ExternRegistry) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 0 {
			return Value{}, nil, fmt.Errorf("%s expects 0 arguments", name)
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("%s requires memory", name)
		}
		addr, err := r.staticLocaleconv(ec.Memory)
		if err != nil {
			return Value{}, nil, err
		}
		return PtrValue(addr), nil, nil
	}
}

var localeconvPointerFields = []struct {
	name  string
	value string
}{
	{name: "decimal_point", value: "."},
	{name: "thousands_sep", value: ""},
	{name: "grouping", value: ""},
	{name: "int_curr_symbol", value: ""},
	{name: "currency_symbol", value: ""},
	{name: "mon_decimal_point", value: ""},
	{name: "mon_thousands_sep", value: ""},
	{name: "mon_grouping", value: ""},
	{name: "positive_sign", value: ""},
	{name: "negative_sign", value: ""},
}

var localeconvCharFields = []string{
	"int_frac_digits",
	"frac_digits",
	"p_cs_precedes",
	"p_sep_by_space",
	"n_cs_precedes",
	"n_sep_by_space",
	"p_sign_posn",
	"n_sign_posn",
	"int_p_cs_precedes",
	"int_p_sep_by_space",
	"int_n_cs_precedes",
	"int_n_sep_by_space",
	"int_p_sign_posn",
	"int_n_sign_posn",
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
	rangeErr  bool
}

func parseStrtoFloatString(s string) parsedStrtoFloat {
	start := 0
	for start < len(s) && isASCIIWhitespace(s[start]) {
		start++
	}
	if v, end, ok := parseStrtoFloatSpecial(s, start); ok {
		return parsedStrtoFloat{value: v, end: end, converted: true}
	}
	for end := len(s); end > start; end-- {
		token := s[start:end]
		v, err := strconv.ParseFloat(token, 64)
		if err == nil {
			return parsedStrtoFloat{
				value:     v,
				end:       end,
				converted: true,
				rangeErr:  v == 0 && floatTokenHasNonZeroSignificand(token),
			}
		}
		if numErr, ok := err.(*strconv.NumError); ok && numErr.Err == strconv.ErrRange {
			return parsedStrtoFloat{value: v, end: end, converted: true, rangeErr: true}
		}
	}
	return parsedStrtoFloat{}
}

func floatTokenHasNonZeroSignificand(token string) bool {
	i := 0
	if i < len(token) && (token[i] == '+' || token[i] == '-') {
		i++
	}
	hex := i+2 <= len(token) && token[i] == '0' && (token[i+1] == 'x' || token[i+1] == 'X')
	if hex {
		i += 2
	}
	for ; i < len(token); i++ {
		ch := token[i]
		if hex {
			if ch == 'p' || ch == 'P' {
				break
			}
		} else if ch == 'e' || ch == 'E' {
			break
		}
		if ch == '.' {
			continue
		}
		if digit := strtoDigit(ch); digit > 0 {
			return true
		}
	}
	return false
}

func normalizeStrtoFloatResult(ret bytecode.ValueType, value float64, rangeErr bool) (float64, bool) {
	if ret != bytecode.TypeF32 {
		return value, rangeErr
	}
	narrowed := float64(float32(value))
	if math.IsInf(narrowed, 0) && !math.IsInf(value, 0) {
		rangeErr = true
	}
	if narrowed == 0 && value != 0 {
		rangeErr = true
	}
	return narrowed, rangeErr
}

func parseStrtoFloatSpecial(s string, start int) (float64, int, bool) {
	i := start
	if i < len(s) && (s[i] == '+' || s[i] == '-') {
		i++
	}
	if i+3 > len(s) || !strings.EqualFold(s[i:i+3], "nan") {
		return 0, 0, false
	}
	end := i + 3
	if end < len(s) && s[end] == '(' {
		for j := end + 1; j < len(s); j++ {
			if s[j] == ')' {
				end = j + 1
				break
			}
		}
	}
	return math.NaN(), end, true
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
		if ch == -1 || !pred(byte(unsignedInt(args[0]))) {
			return IntValue(bytecode.TypeI32, 0), nil, nil
		}
		return IntValue(bytecode.TypeI32, 1), nil, nil
	}
}

func registerWideCtypeClassificationExterns(r *ExternRegistry) {
	r.Register("iswdigit", wideCtypeClassificationExtern("iswdigit", func(ch int64) bool {
		return ch >= '0' && ch <= '9'
	}))
	r.Register("iswalpha", wideCtypeClassificationExtern("iswalpha", func(ch int64) bool {
		return (ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z')
	}))
	r.Register("iswalnum", wideCtypeClassificationExtern("iswalnum", func(ch int64) bool {
		return (ch >= '0' && ch <= '9') || (ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z')
	}))
	r.Register("iswspace", wideCtypeClassificationExtern("iswspace", func(ch int64) bool {
		return ch >= 0 && ch <= 0x7f && isASCIIWhitespace(byte(ch))
	}))
	r.Register("iswlower", wideCtypeClassificationExtern("iswlower", func(ch int64) bool {
		return ch >= 'a' && ch <= 'z'
	}))
	r.Register("iswupper", wideCtypeClassificationExtern("iswupper", func(ch int64) bool {
		return ch >= 'A' && ch <= 'Z'
	}))
	r.Register("iswxdigit", wideCtypeClassificationExtern("iswxdigit", func(ch int64) bool {
		return (ch >= '0' && ch <= '9') || (ch >= 'A' && ch <= 'F') || (ch >= 'a' && ch <= 'f')
	}))
	r.Register("iswprint", wideCtypeClassificationExtern("iswprint", func(ch int64) bool {
		return ch >= 0x20 && ch <= 0x7e
	}))
	r.Register("iswblank", wideCtypeClassificationExtern("iswblank", func(ch int64) bool {
		return ch == ' ' || ch == '\t'
	}))
	r.Register("iswcntrl", wideCtypeClassificationExtern("iswcntrl", func(ch int64) bool {
		return (ch >= 0 && ch < 0x20) || ch == 0x7f
	}))
	r.Register("iswgraph", wideCtypeClassificationExtern("iswgraph", func(ch int64) bool {
		return ch >= 0x21 && ch <= 0x7e
	}))
	r.Register("iswpunct", wideCtypeClassificationExtern("iswpunct", func(ch int64) bool {
		return (ch >= 0x21 && ch <= 0x2f) || (ch >= 0x3a && ch <= 0x40) || (ch >= 0x5b && ch <= 0x60) || (ch >= 0x7b && ch <= 0x7e)
	}))
}

func wideCtypeClassificationExtern(name string, pred func(int64) bool) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 1 {
			return Value{}, nil, fmt.Errorf("%s expects 1 argument", name)
		}
		if !isIntegerLike(args[0].Type) {
			return Value{}, nil, fmt.Errorf("%s expects integer argument", name)
		}
		ch := signedInt(args[0])
		if ch == -1 || !pred(ch) {
			return IntValue(bytecode.TypeI32, 0), nil, nil
		}
		return IntValue(bytecode.TypeI32, 1), nil, nil
	}
}

func registerWideCtypeCaseExterns(r *ExternRegistry) {
	r.Register("towlower", wideCtypeCaseExtern("towlower", wideToLower))
	r.Register("towupper", wideCtypeCaseExtern("towupper", wideToUpper))
}

func wideCtypeCaseExtern(name string, convert func(int64) int64) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 1 {
			return Value{}, nil, fmt.Errorf("%s expects 1 argument", name)
		}
		if !isIntegerLike(args[0].Type) {
			return Value{}, nil, fmt.Errorf("%s expects integer argument", name)
		}
		ch := signedInt(args[0])
		if ch == -1 {
			return IntValue(bytecode.TypeI32, -1), nil, nil
		}
		return IntValue(bytecode.TypeI32, convert(ch)), nil, nil
	}
}

func wideToLower(ch int64) int64 {
	if ch >= 'A' && ch <= 'Z' {
		return ch + ('a' - 'A')
	}
	return ch
}

func wideToUpper(ch int64) int64 {
	if ch >= 'a' && ch <= 'z' {
		return ch - ('a' - 'A')
	}
	return ch
}

const (
	wideClassAlnum uint64 = iota + 1
	wideClassAlpha
	wideClassBlank
	wideClassCntrl
	wideClassDigit
	wideClassGraph
	wideClassLower
	wideClassPrint
	wideClassPunct
	wideClassSpace
	wideClassUpper
	wideClassXDigit
)

const (
	wideTransToLower uint64 = iota + 1
	wideTransToUpper
)

func registerWideCtypeDescriptorExterns(r *ExternRegistry) {
	r.Register("wctype", wideCtypeDescriptorExtern("wctype"))
	r.Register("iswctype", wideIswctypeExtern("iswctype"))
	r.Register("wctrans", wideTransDescriptorExtern("wctrans"))
	r.Register("towctrans", wideTowctransExtern("towctrans"))
}

func wideCtypeDescriptorExtern(name string) ExternFunc {
	classes := map[string]uint64{
		"alnum":  wideClassAlnum,
		"alpha":  wideClassAlpha,
		"blank":  wideClassBlank,
		"cntrl":  wideClassCntrl,
		"digit":  wideClassDigit,
		"graph":  wideClassGraph,
		"lower":  wideClassLower,
		"print":  wideClassPrint,
		"punct":  wideClassPunct,
		"space":  wideClassSpace,
		"upper":  wideClassUpper,
		"xdigit": wideClassXDigit,
	}
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
		return UIntValue(bytecode.TypeU64, classes[s]), nil, nil
	}
}

func wideIswctypeExtern(name string) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 2 {
			return Value{}, nil, fmt.Errorf("%s expects 2 arguments", name)
		}
		if !isIntegerLike(args[0].Type) || !isIntegerLike(args[1].Type) {
			return Value{}, nil, fmt.Errorf("%s expects integer arguments", name)
		}
		ch := signedInt(args[0])
		desc := unsignedInt(args[1])
		var ok bool
		switch desc {
		case wideClassAlnum:
			ok = (ch >= '0' && ch <= '9') || (ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z')
		case wideClassAlpha:
			ok = (ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z')
		case wideClassBlank:
			ok = ch == ' ' || ch == '\t'
		case wideClassCntrl:
			ok = (ch >= 0 && ch < 0x20) || ch == 0x7f
		case wideClassDigit:
			ok = ch >= '0' && ch <= '9'
		case wideClassGraph:
			ok = ch >= 0x21 && ch <= 0x7e
		case wideClassLower:
			ok = ch >= 'a' && ch <= 'z'
		case wideClassPrint:
			ok = ch >= 0x20 && ch <= 0x7e
		case wideClassPunct:
			ok = (ch >= 0x21 && ch <= 0x2f) || (ch >= 0x3a && ch <= 0x40) || (ch >= 0x5b && ch <= 0x60) || (ch >= 0x7b && ch <= 0x7e)
		case wideClassSpace:
			ok = ch >= 0 && ch <= 0x7f && isASCIIWhitespace(byte(ch))
		case wideClassUpper:
			ok = ch >= 'A' && ch <= 'Z'
		case wideClassXDigit:
			ok = (ch >= '0' && ch <= '9') || (ch >= 'A' && ch <= 'F') || (ch >= 'a' && ch <= 'f')
		}
		if ch == -1 || !ok {
			return IntValue(bytecode.TypeI32, 0), nil, nil
		}
		return IntValue(bytecode.TypeI32, 1), nil, nil
	}
}

func wideTransDescriptorExtern(name string) ExternFunc {
	transforms := map[string]uint64{
		"tolower": wideTransToLower,
		"toupper": wideTransToUpper,
	}
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
		return UIntValue(bytecode.TypeU64, transforms[s]), nil, nil
	}
}

func wideTowctransExtern(name string) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 2 {
			return Value{}, nil, fmt.Errorf("%s expects 2 arguments", name)
		}
		if !isIntegerLike(args[0].Type) || !isIntegerLike(args[1].Type) {
			return Value{}, nil, fmt.Errorf("%s expects integer arguments", name)
		}
		ch := signedInt(args[0])
		if ch == -1 {
			return IntValue(bytecode.TypeI32, -1), nil, nil
		}
		switch unsignedInt(args[1]) {
		case wideTransToLower:
			return IntValue(bytecode.TypeI32, wideToLower(ch)), nil, nil
		case wideTransToUpper:
			return IntValue(bytecode.TypeI32, wideToUpper(ch)), nil, nil
		default:
			return IntValue(bytecode.TypeI32, ch), nil, nil
		}
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
		if ch == -1 {
			return IntValue(bytecode.TypeI32, -1), nil, nil
		}
		return IntValue(bytecode.TypeI32, int64(convert(byte(unsignedInt(args[0]))))), nil, nil
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
			r.hostError[args[3].Int] = true
			return UIntValue(bytecode.TypeU64, 0), nil, nil
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
		if r.markHostReadErrorIfUnreadable(args[3].Int) {
			return UIntValue(bytecode.TypeU64, 0), nil, nil
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
		if (name == "ferror" || name == "ferror_unlocked") && r.hostError[args[0].Int] {
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
		r.hostError[args[0].Int] = false
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

func registerInputFormatExterns(r *ExternRegistry) {
	r.Register("scanf", scanfExtern("scanf", r))
	r.Register("fscanf", fscanfExtern("fscanf", r))
	r.Register("sscanf", sscanfExtern("sscanf"))
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
	registerTgmathModfExterns(r, "modf")
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
	r.Register("__builtin_nan", mathNanExtern("__builtin_nan", bytecode.TypeF64))
	r.Register("nanf", mathNanExtern("nanf", bytecode.TypeF32))
	r.Register("nan", mathNanExtern("nan", bytecode.TypeF64))
	r.Register("nanl", mathNanExtern("nanl", bytecode.TypeFLong))
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
	registerTgmathModfExterns(r, "__cvm_tgmath_modf")
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
		if n == 0 {
			return IntValue(bytecode.TypeI32, 0), nil, nil
		}
		for i := int64(0); i < n; i++ {
			leftAddr, err := addSignedOffset(args[0].Int, i)
			if err != nil {
				return Value{}, nil, err
			}
			rightAddr, err := addSignedOffset(args[1].Int, i)
			if err != nil {
				return Value{}, nil, err
			}
			left, err := readMemoryByte(ec.Memory, leftAddr)
			if err != nil {
				return Value{}, nil, err
			}
			right, err := readMemoryByte(ec.Memory, rightAddr)
			if err != nil {
				return Value{}, nil, err
			}
			if left < right {
				return IntValue(bytecode.TypeI32, -1), nil, nil
			}
			if left > right {
				return IntValue(bytecode.TypeI32, 1), nil, nil
			}
			if left == 0 {
				return IntValue(bytecode.TypeI32, 0), nil, nil
			}
		}
		return IntValue(bytecode.TypeI32, 0), nil, nil
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
		out, err := formatVaListCString(name, ec.Memory, args[1].Int, args[2].Int)
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
		out, err := formatVaListCString(name, ec.Memory, args[2].Int, args[3].Int)
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
		out, err := formatVaListCString(name, ec.Memory, args[3].Int, args[4].Int)
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
		out, err := formatVaListCString(name, ec.Memory, args[4].Int, args[5].Int)
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
		out, err := formatVaListCString(name, ec.Memory, args[0].Int, args[1].Int)
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
		out, err := formatVaListCString(name, ec.Memory, args[1].Int, args[2].Int)
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
		out, err := formatVaListCString(name, ec.Memory, args[1].Int, args[2].Int)
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
		out, err := formatVaListCString(name, ec.Memory, args[2].Int, args[3].Int)
		if err != nil {
			return Value{}, nil, err
		}
		if _, err := fmt.Fprint(w, out); err != nil {
			return Value{}, nil, err
		}
		return IntValue(bytecode.TypeI32, int64(len(out))), nil, nil
	}
}

func scanfExtern(name string, r *ExternRegistry) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) < 1 {
			return Value{}, nil, fmt.Errorf("%s expects at least 1 argument", name)
		}
		if !isPointerType(args[0].Type) {
			return Value{}, nil, fmt.Errorf("%s expects format string argument", name)
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("%s requires memory", name)
		}
		if r.stdinHandle == 0 {
			if _, ok := r.LookupVariable("stdin", ec.Memory); !ok {
				return Value{}, nil, fmt.Errorf("%s could not initialize stdin", name)
			}
		}
		n, err := scanHostStream(name, r, ec.Memory, r.stdinHandle, args[0].Int, args[1:])
		if err != nil {
			return Value{}, nil, err
		}
		return IntValue(bytecode.TypeI32, int64(n)), nil, nil
	}
}

func fscanfExtern(name string, r *ExternRegistry) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) < 2 {
			return Value{}, nil, fmt.Errorf("%s expects at least 2 arguments", name)
		}
		if !isPointerType(args[0].Type) || !isPointerType(args[1].Type) {
			return Value{}, nil, fmt.Errorf("%s expects stream and format string arguments", name)
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("%s requires memory", name)
		}
		if _, ok := r.lookupHostWriter(args[0].Int); !ok {
			return Value{}, nil, fmt.Errorf("unknown stream handle %#x", args[0].Int)
		}
		n, err := scanHostStream(name, r, ec.Memory, args[0].Int, args[1].Int, args[2:])
		if err != nil {
			return Value{}, nil, err
		}
		return IntValue(bytecode.TypeI32, int64(n)), nil, nil
	}
}

func sscanfExtern(name string) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) < 2 {
			return Value{}, nil, fmt.Errorf("%s expects at least 2 arguments", name)
		}
		if !isPointerType(args[0].Type) || !isPointerType(args[1].Type) {
			return Value{}, nil, fmt.Errorf("%s expects input and format string arguments", name)
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("%s requires memory", name)
		}
		n, err := scanCString(name, ec.Memory, args[0].Int, args[1].Int, args[2:])
		if err != nil {
			return Value{}, nil, err
		}
		return IntValue(bytecode.TypeI32, int64(n)), nil, nil
	}
}

func scanCString(name string, mem *Memory, inputAddr, formatAddr uint64, args []Value) (int, error) {
	input, err := mem.ReadCString(inputAddr)
	if err != nil {
		return 0, err
	}
	format, err := mem.ReadCString(formatAddr)
	if err != nil {
		return 0, err
	}
	assigned, _, inputFailure, err := scanString(name, mem, input, format, args)
	if inputFailure && assigned == 0 {
		return -1, err
	}
	return assigned, err
}

func scanHostStream(name string, r *ExternRegistry, mem *Memory, stream, formatAddr uint64, args []Value) (int, error) {
	format, err := mem.ReadCString(formatAddr)
	if err != nil {
		return 0, err
	}
	if r.markHostReadErrorIfUnreadable(stream) {
		return 0, nil
	}
	var input []byte
	for {
		ch, ok := r.readHostChar(stream)
		if !ok {
			break
		}
		input = append(input, ch)
	}
	assigned, consumed, inputFailure, err := scanString(name, mem, string(input), format, args)
	if consumed < len(input) {
		r.pushBackHostBytes(stream, input[consumed:])
	}
	if consumed == len(input) && len(input) == 0 {
		r.hostEOF[stream] = true
	}
	if inputFailure && assigned == 0 {
		return -1, err
	}
	return assigned, err
}

func (r *ExternRegistry) pushBackHostBytes(addr uint64, data []byte) {
	for i := len(data) - 1; i >= 0; i-- {
		r.hostPushback[addr] = append(r.hostPushback[addr], data[i])
	}
	if len(data) > 0 {
		r.hostEOF[addr] = false
	}
}

func (r *ExternRegistry) markHostReadErrorIfUnreadable(addr uint64) bool {
	if file := r.hostFiles[addr]; file != nil {
		if !file.readable {
			r.hostError[addr] = true
			return true
		}
		if file.updateMode && file.lastOp == hostFileOpWrite {
			r.hostError[addr] = true
			return true
		}
	}
	return false
}

func scanString(name string, mem *Memory, input, format string, args []Value) (int, int, bool, error) {
	inputIndex := 0
	argIndex := 0
	assigned := 0
	for formatIndex := 0; formatIndex < len(format); formatIndex++ {
		ch := format[formatIndex]
		if isASCIIWhitespace(ch) {
			for formatIndex+1 < len(format) && isASCIIWhitespace(format[formatIndex+1]) {
				formatIndex++
			}
			inputIndex = scanSkipWhitespace(input, inputIndex)
			continue
		}
		if ch != '%' {
			if inputIndex >= len(input) || input[inputIndex] != ch {
				return assigned, inputIndex, inputIndex >= len(input), nil
			}
			inputIndex++
			continue
		}
		formatIndex++
		if formatIndex >= len(format) {
			return 0, inputIndex, false, fmt.Errorf("%s has trailing %% in format", name)
		}
		if format[formatIndex] == '%' {
			if inputIndex >= len(input) || input[inputIndex] != '%' {
				return assigned, inputIndex, inputIndex >= len(input), nil
			}
			inputIndex++
			continue
		}
		suppress := false
		if format[formatIndex] == '*' {
			suppress = true
			formatIndex++
			if formatIndex >= len(format) {
				return 0, inputIndex, false, fmt.Errorf("%s has trailing %% in format", name)
			}
		}
		width := 0
		for formatIndex < len(format) && format[formatIndex] >= '0' && format[formatIndex] <= '9' {
			width = width*10 + int(format[formatIndex]-'0')
			formatIndex++
		}
		lengthMod := ""
		for formatIndex < len(format) {
			switch format[formatIndex] {
			case 'h', 'l':
				lengthMod = string(format[formatIndex])
				if formatIndex+1 < len(format) && format[formatIndex+1] == format[formatIndex] {
					lengthMod += string(format[formatIndex])
					formatIndex += 2
				} else {
					formatIndex++
				}
			case 'j', 'z', 't', 'L':
				lengthMod = string(format[formatIndex])
				formatIndex++
			default:
				goto verb
			}
		}
	verb:
		if formatIndex >= len(format) {
			return 0, inputIndex, false, fmt.Errorf("%s has trailing %% in format", name)
		}
		verb := format[formatIndex]
		if !suppress {
			if argIndex >= len(args) {
				return 0, inputIndex, false, fmt.Errorf("%s format needs more arguments", name)
			}
			if !isPointerType(args[argIndex].Type) {
				return 0, inputIndex, false, fmt.Errorf("%s %%%c expects pointer argument", name, verb)
			}
		}
		switch verb {
		case 'd', 'i', 'u', 'x', 'X', 'o':
			base := int64(10)
			if verb == 'i' {
				base = 0
			} else if verb == 'x' || verb == 'X' {
				base = 16
			} else if verb == 'o' {
				base = 8
			}
			inputIndex = scanSkipWhitespace(input, inputIndex)
			if inputIndex >= len(input) {
				return assigned, inputIndex, true, nil
			}
			token := scanLimit(input, inputIndex, width)
			parsed, err := parseStrtoIntegerString(name, token, base)
			if err != nil {
				return 0, inputIndex, false, err
			}
			if !parsed.converted {
				return assigned, inputIndex, false, nil
			}
			if !suppress {
				unsigned := verb == 'u' || verb == 'x' || verb == 'X' || verb == 'o'
				if err := scanStoreInteger(mem, args[argIndex].Int, lengthMod, unsigned, parsed); err != nil {
					return 0, inputIndex, false, err
				}
				argIndex++
				assigned++
			}
			inputIndex += parsed.end
		case 'f', 'F', 'e', 'E', 'g', 'G', 'a', 'A':
			inputIndex = scanSkipWhitespace(input, inputIndex)
			if inputIndex >= len(input) {
				return assigned, inputIndex, true, nil
			}
			token := scanLimit(input, inputIndex, width)
			parsed := parseStrtoFloatString(token)
			if !parsed.converted {
				return assigned, inputIndex, false, nil
			}
			if !suppress {
				if err := scanStoreFloat(mem, args[argIndex].Int, lengthMod, parsed.value); err != nil {
					return 0, inputIndex, false, err
				}
				argIndex++
				assigned++
			}
			inputIndex += parsed.end
		case 'p':
			if lengthMod != "" {
				return 0, inputIndex, false, fmt.Errorf("%s %%p does not support length modifier %q", name, lengthMod)
			}
			inputIndex = scanSkipWhitespace(input, inputIndex)
			if inputIndex >= len(input) {
				return assigned, inputIndex, true, nil
			}
			token := scanLimit(input, inputIndex, width)
			parsed, err := parseStrtoIntegerString(name, token, 16)
			if err != nil {
				return 0, inputIndex, false, err
			}
			if !parsed.converted {
				return assigned, inputIndex, false, nil
			}
			if parsed.neg {
				return 0, inputIndex, false, fmt.Errorf("%s %%p does not support negative pointer input", name)
			}
			if !suppress {
				if err := mem.WritePointer(args[argIndex].Int, parsed.value); err != nil {
					return 0, inputIndex, false, err
				}
				argIndex++
				assigned++
			}
			inputIndex += parsed.end
		case 'n':
			if width != 0 {
				return 0, inputIndex, false, fmt.Errorf("%s %%n does not support width", name)
			}
			if !suppress {
				countType, countAlign := scanIntegerType(lengthMod, false)
				if err := mem.Store(args[argIndex].Int, countType, countAlign, normalizeInt(IntValue(countType, int64(inputIndex)))); err != nil {
					return 0, inputIndex, false, err
				}
				argIndex++
			}
		case 's':
			inputIndex = scanSkipWhitespace(input, inputIndex)
			if inputIndex >= len(input) {
				return assigned, inputIndex, true, nil
			}
			end := inputIndex
			limit := len(input)
			if width > 0 && inputIndex+width < limit {
				limit = inputIndex + width
			}
			for end < limit && !isASCIIWhitespace(input[end]) {
				end++
			}
			if end == inputIndex {
				return assigned, inputIndex, false, nil
			}
			if !suppress {
				data := append([]byte(input[inputIndex:end]), 0)
				if err := writeMemoryBytes(mem, args[argIndex].Int, data); err != nil {
					return 0, inputIndex, false, err
				}
				argIndex++
				assigned++
			}
			inputIndex = end
		case 'c':
			count := width
			if count == 0 {
				count = 1
			}
			if inputIndex+count > len(input) {
				return assigned, inputIndex, true, nil
			}
			if !suppress {
				if err := writeMemoryBytes(mem, args[argIndex].Int, []byte(input[inputIndex:inputIndex+count])); err != nil {
					return 0, inputIndex, false, err
				}
				argIndex++
				assigned++
			}
			inputIndex += count
		case '[':
			matches, nextFormatIndex, err := parseScanSet(name, format, formatIndex)
			if err != nil {
				return 0, inputIndex, false, err
			}
			formatIndex = nextFormatIndex
			end := inputIndex
			limit := len(input)
			if width > 0 && inputIndex+width < limit {
				limit = inputIndex + width
			}
			for end < limit && matches(input[end]) {
				end++
			}
			if end == inputIndex {
				return assigned, inputIndex, inputIndex >= len(input), nil
			}
			if !suppress {
				data := append([]byte(input[inputIndex:end]), 0)
				if err := writeMemoryBytes(mem, args[argIndex].Int, data); err != nil {
					return 0, inputIndex, false, err
				}
				argIndex++
				assigned++
			}
			inputIndex = end
		default:
			return 0, inputIndex, false, fmt.Errorf("%s unsupported scan format %%%c", name, verb)
		}
	}
	return assigned, inputIndex, false, nil
}

func parseScanSet(name, format string, open int) (func(byte) bool, int, error) {
	i := open + 1
	if i >= len(format) {
		return nil, open, fmt.Errorf("%s has unterminated scan set", name)
	}
	negated := false
	if format[i] == '^' {
		negated = true
		i++
	}
	if i >= len(format) {
		return nil, open, fmt.Errorf("%s has unterminated scan set", name)
	}
	members := make(map[byte]bool)
	if format[i] == ']' {
		members[']'] = true
		i++
	}
	for i < len(format) && format[i] != ']' {
		start := format[i]
		if i+2 < len(format) && format[i+1] == '-' && format[i+2] != ']' {
			end := format[i+2]
			if start <= end {
				for ch := start; ch <= end; ch++ {
					members[ch] = true
				}
			} else {
				for ch := start; ch >= end; ch-- {
					members[ch] = true
					if ch == 0 {
						break
					}
				}
			}
			i += 3
			continue
		}
		members[start] = true
		i++
	}
	if i >= len(format) || format[i] != ']' {
		return nil, open, fmt.Errorf("%s has unterminated scan set", name)
	}
	return func(ch byte) bool {
		contained := members[ch]
		if negated {
			return !contained
		}
		return contained
	}, i, nil
}

func scanSkipWhitespace(s string, i int) int {
	for i < len(s) && isASCIIWhitespace(s[i]) {
		i++
	}
	return i
}

func scanLimit(s string, start, width int) string {
	if width <= 0 || start+width > len(s) {
		return s[start:]
	}
	return s[start : start+width]
}

func scanStoreInteger(mem *Memory, addr uint64, lengthMod string, unsigned bool, parsed parsedStrtoInteger) error {
	t, align := scanIntegerType(lengthMod, unsigned)
	if unsigned {
		v := parsed.value
		if parsed.neg {
			v = -v
		}
		return mem.Store(addr, t, align, normalizeInt(UIntValue(t, v)))
	}
	v := int64(parsed.value)
	if parsed.neg {
		v = -v
	}
	return mem.Store(addr, t, align, normalizeInt(IntValue(t, v)))
}

func scanStoreFloat(mem *Memory, addr uint64, lengthMod string, value float64) error {
	t, align, err := scanFloatType(lengthMod)
	if err != nil {
		return err
	}
	return mem.Store(addr, t, align, FloatValue(t, value))
}

func scanFloatType(lengthMod string) (bytecode.ValueType, int64, error) {
	switch lengthMod {
	case "":
		return bytecode.TypeF32, 4, nil
	case "l":
		return bytecode.TypeF64, 8, nil
	case "L":
		return bytecode.TypeFLong, 8, nil
	default:
		return bytecode.TypeVoid, 0, fmt.Errorf("unsupported scanf floating length modifier %q", lengthMod)
	}
}

func scanIntegerType(lengthMod string, unsigned bool) (bytecode.ValueType, int64) {
	switch lengthMod {
	case "hh":
		if unsigned {
			return bytecode.TypeU8, 1
		}
		return bytecode.TypeI8, 1
	case "h":
		if unsigned {
			return bytecode.TypeU16, 2
		}
		return bytecode.TypeI16, 2
	case "l", "ll", "j", "z", "t":
		if unsigned {
			return bytecode.TypeU64, 8
		}
		return bytecode.TypeI64, 8
	default:
		if unsigned {
			return bytecode.TypeU32, 4
		}
		return bytecode.TypeI32, 4
	}
}

const (
	memoryVaListTagI32  = 1
	memoryVaListTagU32  = 2
	memoryVaListTagI64  = 3
	memoryVaListTagU64  = 4
	memoryVaListTagPtr  = 5
	memoryVaListTagF64  = 6
	memoryVaListMaxArgs = 1024
)

func formatVaListCString(name string, mem *Memory, formatAddr, vaListAddr uint64) (string, error) {
	args, err := readMemoryVaList(name, mem, vaListAddr)
	if err != nil {
		return "", err
	}
	return formatCString(name, mem, formatAddr, args)
}

func readMemoryVaList(name string, mem *Memory, vaListAddr uint64) ([]Value, error) {
	if vaListAddr == 0 {
		return nil, nil
	}
	countValue, err := mem.Load(vaListAddr, bytecode.TypeU64, 8)
	if err != nil {
		return nil, err
	}
	count := countValue.Int
	if count > memoryVaListMaxArgs {
		return nil, fmt.Errorf("%s va_list count %d exceeds limit", name, count)
	}
	args := make([]Value, 0, int(count))
	for i := uint64(0); i < count; i++ {
		entry := vaListAddr + 8 + i*16
		tagValue, err := mem.Load(entry, bytecode.TypeU64, 8)
		if err != nil {
			return nil, err
		}
		payloadValue, err := mem.Load(entry+8, bytecode.TypeU64, 8)
		if err != nil {
			return nil, err
		}
		payload := payloadValue.Int
		switch tagValue.Int {
		case memoryVaListTagI32:
			args = append(args, IntValue(bytecode.TypeI32, int64(int32(payload))))
		case memoryVaListTagU32:
			args = append(args, UIntValue(bytecode.TypeU32, uint64(uint32(payload))))
		case memoryVaListTagI64:
			args = append(args, IntValue(bytecode.TypeI64, int64(payload)))
		case memoryVaListTagU64:
			args = append(args, UIntValue(bytecode.TypeU64, payload))
		case memoryVaListTagPtr:
			args = append(args, PtrValue(payload))
		case memoryVaListTagF64:
			args = append(args, FloatValue(bytecode.TypeF64, math.Float64frombits(payload)))
		default:
			return nil, fmt.Errorf("%s unsupported va_list tag %d at index %d", name, tagValue.Int, i)
		}
	}
	return args, nil
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

func registerTgmathModfExterns(r *ExternRegistry, base string) {
	r.Register(base+"f", mathModfExtern(base+"f", bytecode.TypeF32, 4))
	r.Register(base, mathModfExtern(base, bytecode.TypeF64, 8))
	r.Register(base+"l", mathModfExtern(base+"l", bytecode.TypeFLong, 16))
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

func mathNanExtern(name string, ret bytecode.ValueType) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 1 {
			return Value{}, nil, fmt.Errorf("%s expects 1 argument", name)
		}
		if !isPointerType(args[0].Type) {
			return Value{}, nil, fmt.Errorf("%s expects string argument", name)
		}
		return FloatValue(ret, math.NaN()), nil, nil
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

func mathModfExtern(name string, ret bytecode.ValueType, size int64) ExternFunc {
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
		intPart, frac := math.Modf(cvmFloat(args[0]))
		if err := ec.Memory.Store(args[1].Int, ret, size, FloatValue(ret, intPart)); err != nil {
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

func (r *ExternRegistry) staticBlock(mem *Memory, name string, size, align int64) (uint64, bool, error) {
	if mem == nil {
		return 0, false, fmt.Errorf("memory is nil")
	}
	if byName := r.staticBlocks[mem]; byName != nil {
		if addr, ok := byName[name]; ok {
			return addr, false, nil
		}
	}
	addr, err := mem.TryAlloc("extern:"+name, size, align, false, blockGlobal)
	if err != nil {
		return 0, false, err
	}
	if r.staticBlocks[mem] == nil {
		r.staticBlocks[mem] = make(map[string]uint64)
	}
	r.staticBlocks[mem][name] = addr
	return addr, true, nil
}

func (r *ExternRegistry) staticLocaleconv(mem *Memory) (uint64, error) {
	ptrSize := mem.target.PointerSize
	ptrAlign := mem.target.PointerAlign
	size := alignInt64(int64(len(localeconvPointerFields))*ptrSize+int64(len(localeconvCharFields)), ptrAlign)
	addr, created, err := r.staticBlock(mem, "localeconv", size, ptrAlign)
	if err != nil || !created {
		return addr, err
	}
	for i, field := range localeconvPointerFields {
		fieldAddr, err := r.staticCString(mem, "localeconv:"+field.name, field.value)
		if err != nil {
			return 0, err
		}
		if err := mem.Store(addr+uint64(int64(i)*ptrSize), bytecode.TypePtr, ptrAlign, PtrValue(fieldAddr)); err != nil {
			return 0, err
		}
	}
	charBase := addr + uint64(int64(len(localeconvPointerFields))*ptrSize)
	for i := range localeconvCharFields {
		if err := mem.Store(charBase+uint64(i), bytecode.TypeI8, 1, IntValue(bytecode.TypeI8, 127)); err != nil {
			return 0, err
		}
	}
	return addr, nil
}

func alignInt64(v, align int64) int64 {
	if align <= 1 {
		return v
	}
	if rem := v % align; rem != 0 {
		v += align - rem
	}
	return v
}

func (r *ExternRegistry) setErrno(mem *Memory, value int32) error {
	if r == nil {
		return nil
	}
	addr, err := r.staticI32Variable(mem, "errno", 0)
	if err != nil {
		return err
	}
	return mem.Store(addr, bytecode.TypeI32, 4, IntValue(bytecode.TypeI32, int64(value)))
}

func (r *ExternRegistry) lookupHostWriter(addr uint64) (io.Writer, bool) {
	if r.hostClosed[addr] {
		return nil, false
	}
	w, ok := r.hostWriters[addr]
	return w, ok
}

func (r *ExternRegistry) readHostChar(addr uint64) (byte, bool) {
	buf := r.hostPushback[addr]
	if len(buf) == 0 {
		if file := r.hostFiles[addr]; file != nil {
			if !file.readable {
				return 0, false
			}
			if file.pos < 0 || file.pos >= int64(len(file.data)) {
				return 0, false
			}
			ch := file.data[file.pos]
			file.pos++
			if file.updateMode {
				file.lastOp = hostFileOpRead
			}
			return ch, true
		}
		if addr == r.stdinHandle && r.stdin != nil {
			var one [1]byte
			n, err := r.stdin.Read(one[:])
			if n > 0 {
				return one[0], true
			}
			if err != nil {
				return 0, false
			}
		}
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
	delete(r.hostClosed, addr)
	delete(r.hostError, addr)
	return addr, nil
}
