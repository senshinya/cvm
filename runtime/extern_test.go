package runtime

import (
	"bytes"
	"context"
	"math"
	"strings"
	"testing"

	"shinya.click/cvm/bytecode"
)

func TestDefaultExternRegistryHasExitAndAbort(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	for _, name := range []string{"exit", "abort", "__builtin_abort", "__builtin_va_start", "__builtin_va_end", "puts", "putchar", "putchar_unlocked", "getchar", "getchar_unlocked", "fputc", "fputc_unlocked", "putc", "putc_unlocked", "fputs", "fputs_unlocked", "fgetc", "getc", "getc_unlocked", "ungetc", "fgets", "fflush", "fclose", "ferror", "clearerr", "feof", "fwrite", "fread", "strcmp", "memcmp", "__builtin_malloc", "__builtin_calloc", "__builtin_strdup", "__builtin_object_size", "__builtin_dynamic_object_size", "__builtin_memcpy", "__builtin_memmove", "__builtin_mempcpy", "__builtin_memset", "__builtin_bzero", "__builtin___memcpy_chk", "__builtin___memmove_chk", "__builtin___mempcpy_chk", "__builtin___memset_chk", "__builtin_strlen", "__builtin_strchr", "__builtin_strstr", "__builtin_strcpy", "__builtin_stpcpy", "__builtin_strcat", "__builtin_strncpy", "__builtin_stpncpy", "__builtin_strncat", "__builtin___strcpy_chk", "__builtin___stpcpy_chk", "__builtin___strcat_chk", "__builtin___strncpy_chk", "__builtin___stpncpy_chk", "__builtin___strncat_chk", "__builtin_sprintf", "__builtin_snprintf", "__builtin_vsprintf", "__builtin_vsnprintf", "vsprintf", "vsnprintf", "__builtin___sprintf_chk", "__builtin___snprintf_chk", "__builtin___vsprintf_chk", "__builtin___vsnprintf_chk", "__builtin_printf", "__builtin_printf_unlocked", "printf", "printf_unlocked", "__builtin_fprintf", "__builtin_fprintf_unlocked", "fprintf", "fprintf_unlocked", "__builtin_vprintf", "vprintf", "__builtin_vfprintf", "vfprintf", "__builtin___printf_chk", "__builtin___fprintf_chk", "__builtin___vprintf_chk", "__builtin___vfprintf_chk", "feclearexcept", "fetestexcept"} {
		if _, ok := reg.Lookup(name); !ok {
			t.Fatalf("missing extern %s", name)
		}
	}
}

func TestPutsWritesCString(t *testing.T) {
	var out bytes.Buffer
	reg := DefaultExternRegistry(&out, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	addr := mustAllocBytes(t, mem, "string:0", []byte("hello\x00"), true, blockString)
	fn, _ := reg.Lookup("puts")
	ret, exit, err := fn(context.Background(), &ExternContext{Memory: mem}, []Value{ObjectAddrValue(addr)})
	if err != nil || exit != nil {
		t.Fatalf("puts ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if out.String() != "hello\n" {
		t.Fatalf("puts output = %q", out.String())
	}
}

func TestPutcharWritesByte(t *testing.T) {
	var out bytes.Buffer
	reg := DefaultExternRegistry(&out, nil)
	fn, ok := reg.Lookup("putchar")
	if !ok {
		t.Fatal("missing putchar extern")
	}
	ret, exit, err := fn(context.Background(), nil, []Value{IntValue(bytecode.TypeI32, 'A')})
	if err != nil || exit != nil {
		t.Fatalf("putchar ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeI32 || ret.Int != 'A' || out.String() != "A" {
		t.Fatalf("putchar ret=%#v output=%q, want i32 'A' and A", ret, out.String())
	}
}

func TestFputcWritesByteToHostHandle(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer
	reg := DefaultExternRegistry(&out, &errOut)
	mem := NewMemory(bytecode.DefaultTarget())
	stderr, ok := reg.LookupVariable("stderr", mem)
	if !ok {
		t.Fatal("missing stderr extern variable")
	}
	fn, ok := reg.Lookup("fputc")
	if !ok {
		t.Fatal("missing fputc extern")
	}
	ret, exit, err := fn(context.Background(), &ExternContext{Memory: mem}, []Value{IntValue(bytecode.TypeI32, 'B'), PtrValue(stderr)})
	if err != nil || exit != nil {
		t.Fatalf("fputc ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeI32 || ret.Int != 'B' || out.String() != "" || errOut.String() != "B" {
		t.Fatalf("fputc ret=%#v stdout=%q stderr=%q, want i32 'B' on stderr", ret, out.String(), errOut.String())
	}
}

func TestOutputCharacterAliasesWriteBytes(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer
	reg := DefaultExternRegistry(&out, &errOut)
	mem := NewMemory(bytecode.DefaultTarget())
	stderr, ok := reg.LookupVariable("stderr", mem)
	if !ok {
		t.Fatal("missing stderr extern variable")
	}
	for _, tt := range []struct {
		name string
		args []Value
		ch   rune
	}{
		{name: "putc", args: []Value{IntValue(bytecode.TypeI32, 'P'), PtrValue(stderr)}, ch: 'P'},
		{name: "putc_unlocked", args: []Value{IntValue(bytecode.TypeI32, 'U'), PtrValue(stderr)}, ch: 'U'},
		{name: "putchar_unlocked", args: []Value{IntValue(bytecode.TypeI32, 'H')}, ch: 'H'},
	} {
		fn, ok := reg.Lookup(tt.name)
		if !ok {
			t.Fatalf("missing %s extern", tt.name)
		}
		ret, exit, err := fn(context.Background(), &ExternContext{Memory: mem}, tt.args)
		if err != nil || exit != nil {
			t.Fatalf("%s ret=%#v exit=%#v err=%v", tt.name, ret, exit, err)
		}
		if ret.Type != bytecode.TypeI32 || ret.Int != uint64(tt.ch) {
			t.Fatalf("%s ret=%#v, want i32 %q", tt.name, ret, tt.ch)
		}
	}
	if out.String() != "H" || errOut.String() != "PU" {
		t.Fatalf("stdout=%q stderr=%q, want H and PU", out.String(), errOut.String())
	}
}

func TestGetcharEmptyInputReturnsEOF(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	fn, ok := reg.Lookup("getchar")
	if !ok {
		t.Fatal("missing getchar extern")
	}
	ret, exit, err := fn(context.Background(), nil, nil)
	if err != nil || exit != nil {
		t.Fatalf("getchar ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeI32 || int32(ret.Int) != -1 {
		t.Fatalf("getchar ret=%#v, want i32 -1", ret)
	}
}

func TestFgetcEmptyInputReturnsEOF(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	stdin, ok := reg.LookupVariable("stdin", mem)
	if !ok {
		t.Fatal("missing stdin extern variable")
	}
	fn, ok := reg.Lookup("fgetc")
	if !ok {
		t.Fatal("missing fgetc extern")
	}
	ret, exit, err := fn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(stdin)})
	if err != nil || exit != nil {
		t.Fatalf("fgetc ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeI32 || int32(ret.Int) != -1 {
		t.Fatalf("fgetc ret=%#v, want i32 -1", ret)
	}
}

func TestInputCharacterAliasesConsumePushback(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	stdin, ok := reg.LookupVariable("stdin", mem)
	if !ok {
		t.Fatal("missing stdin extern variable")
	}
	ungetcFn, ok := reg.Lookup("ungetc")
	if !ok {
		t.Fatal("missing ungetc extern")
	}
	for _, tt := range []struct {
		name string
		args []Value
		ch   rune
	}{
		{name: "getc", args: []Value{PtrValue(stdin)}, ch: 'G'},
		{name: "getc_unlocked", args: []Value{PtrValue(stdin)}, ch: 'U'},
		{name: "getchar_unlocked", args: nil, ch: 'H'},
	} {
		if _, exit, err := ungetcFn(context.Background(), &ExternContext{Memory: mem}, []Value{IntValue(bytecode.TypeI32, int64(tt.ch)), PtrValue(stdin)}); err != nil || exit != nil {
			t.Fatalf("ungetc %q exit=%#v err=%v", tt.ch, exit, err)
		}
		fn, ok := reg.Lookup(tt.name)
		if !ok {
			t.Fatalf("missing %s extern", tt.name)
		}
		ret, exit, err := fn(context.Background(), &ExternContext{Memory: mem}, tt.args)
		if err != nil || exit != nil {
			t.Fatalf("%s ret=%#v exit=%#v err=%v", tt.name, ret, exit, err)
		}
		if ret.Type != bytecode.TypeI32 || ret.Int != uint64(tt.ch) {
			t.Fatalf("%s ret=%#v, want i32 %q", tt.name, ret, tt.ch)
		}
	}
}

func TestFgetcRejectsUnknownHandle(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	fn, ok := reg.Lookup("fgetc")
	if !ok {
		t.Fatal("missing fgetc extern")
	}
	_, _, err := fn(context.Background(), nil, []Value{PtrValue(0xdeadbeef)})
	if err == nil || !strings.Contains(err.Error(), "unknown stream handle") {
		t.Fatalf("fgetc err = %v, want unknown stream handle", err)
	}
}

func TestUngetcPushesBackCharacterToFgetc(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	stdin, ok := reg.LookupVariable("stdin", mem)
	if !ok {
		t.Fatal("missing stdin extern variable")
	}
	ungetcFn, ok := reg.Lookup("ungetc")
	if !ok {
		t.Fatal("missing ungetc extern")
	}
	fgetcFn, ok := reg.Lookup("fgetc")
	if !ok {
		t.Fatal("missing fgetc extern")
	}
	ret, exit, err := ungetcFn(context.Background(), &ExternContext{Memory: mem}, []Value{IntValue(bytecode.TypeI32, 'Z'), PtrValue(stdin)})
	if err != nil || exit != nil {
		t.Fatalf("ungetc ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeI32 || ret.Int != 'Z' {
		t.Fatalf("ungetc ret=%#v, want i32 'Z'", ret)
	}
	ret, exit, err = fgetcFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(stdin)})
	if err != nil || exit != nil {
		t.Fatalf("fgetc ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeI32 || ret.Int != 'Z' {
		t.Fatalf("fgetc ret=%#v, want i32 'Z'", ret)
	}
	ret, exit, err = fgetcFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(stdin)})
	if err != nil || exit != nil {
		t.Fatalf("fgetc after pushback ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeI32 || int32(ret.Int) != -1 {
		t.Fatalf("fgetc after pushback ret=%#v, want EOF", ret)
	}
}

func TestUngetcPushbackFeedsGetchar(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	stdin, ok := reg.LookupVariable("stdin", mem)
	if !ok {
		t.Fatal("missing stdin extern variable")
	}
	ungetcFn, ok := reg.Lookup("ungetc")
	if !ok {
		t.Fatal("missing ungetc extern")
	}
	getcharFn, ok := reg.Lookup("getchar")
	if !ok {
		t.Fatal("missing getchar extern")
	}
	if _, exit, err := ungetcFn(context.Background(), &ExternContext{Memory: mem}, []Value{IntValue(bytecode.TypeI32, 'Q'), PtrValue(stdin)}); err != nil || exit != nil {
		t.Fatalf("ungetc exit=%#v err=%v", exit, err)
	}
	ret, exit, err := getcharFn(context.Background(), &ExternContext{Memory: mem}, nil)
	if err != nil || exit != nil {
		t.Fatalf("getchar ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeI32 || ret.Int != 'Q' {
		t.Fatalf("getchar ret=%#v, want i32 'Q'", ret)
	}
}

func TestUngetcEOFDoesNotPushBack(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	stdin, ok := reg.LookupVariable("stdin", mem)
	if !ok {
		t.Fatal("missing stdin extern variable")
	}
	ungetcFn, ok := reg.Lookup("ungetc")
	if !ok {
		t.Fatal("missing ungetc extern")
	}
	ret, exit, err := ungetcFn(context.Background(), &ExternContext{Memory: mem}, []Value{IntValue(bytecode.TypeI32, -1), PtrValue(stdin)})
	if err != nil || exit != nil {
		t.Fatalf("ungetc EOF ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeI32 || int32(ret.Int) != -1 {
		t.Fatalf("ungetc EOF ret=%#v, want EOF", ret)
	}
}

func TestCharacterInputUpdatesEOFStatus(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	stdin, ok := reg.LookupVariable("stdin", mem)
	if !ok {
		t.Fatal("missing stdin extern variable")
	}
	fgetcFn, ok := reg.Lookup("fgetc")
	if !ok {
		t.Fatal("missing fgetc extern")
	}
	feofFn, ok := reg.Lookup("feof")
	if !ok {
		t.Fatal("missing feof extern")
	}
	clearerrFn, ok := reg.Lookup("clearerr")
	if !ok {
		t.Fatal("missing clearerr extern")
	}
	ungetcFn, ok := reg.Lookup("ungetc")
	if !ok {
		t.Fatal("missing ungetc extern")
	}
	ret, exit, err := fgetcFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(stdin)})
	if err != nil || exit != nil {
		t.Fatalf("fgetc ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeI32 || int32(ret.Int) != -1 {
		t.Fatalf("fgetc ret=%#v, want EOF", ret)
	}
	ret, exit, err = feofFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(stdin)})
	if err != nil || exit != nil {
		t.Fatalf("feof ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeI32 || ret.Int != 1 {
		t.Fatalf("feof ret=%#v, want 1", ret)
	}
	if _, exit, err := clearerrFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(stdin)}); err != nil || exit != nil {
		t.Fatalf("clearerr exit=%#v err=%v", exit, err)
	}
	ret, exit, err = feofFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(stdin)})
	if err != nil || exit != nil {
		t.Fatalf("feof after clearerr ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeI32 || ret.Int != 0 {
		t.Fatalf("feof after clearerr ret=%#v, want 0", ret)
	}
	if _, exit, err := fgetcFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(stdin)}); err != nil || exit != nil {
		t.Fatalf("second fgetc exit=%#v err=%v", exit, err)
	}
	if _, exit, err := ungetcFn(context.Background(), &ExternContext{Memory: mem}, []Value{IntValue(bytecode.TypeI32, 'A'), PtrValue(stdin)}); err != nil || exit != nil {
		t.Fatalf("ungetc exit=%#v err=%v", exit, err)
	}
	ret, exit, err = feofFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(stdin)})
	if err != nil || exit != nil {
		t.Fatalf("feof after ungetc ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeI32 || ret.Int != 0 {
		t.Fatalf("feof after ungetc ret=%#v, want 0", ret)
	}
}

func TestFgetsEmptyInputReturnsNullAndLeavesBuffer(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	stdin, ok := reg.LookupVariable("stdin", mem)
	if !ok {
		t.Fatal("missing stdin extern variable")
	}
	buf := mustAllocBytes(t, mem, "fgets:buf", []byte("xy\x00"), false, blockLocal)
	fn, ok := reg.Lookup("fgets")
	if !ok {
		t.Fatal("missing fgets extern")
	}
	ret, exit, err := fn(context.Background(), &ExternContext{Memory: mem}, []Value{
		ObjectAddrValue(buf),
		IntValue(bytecode.TypeI32, 3),
		PtrValue(stdin),
	})
	if err != nil || exit != nil {
		t.Fatalf("fgets ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypePtr || ret.Int != 0 {
		t.Fatalf("fgets ret=%#v, want null pointer", ret)
	}
	got, err := mem.ReadCString(buf)
	if err != nil {
		t.Fatalf("ReadCString: %v", err)
	}
	if got != "xy" {
		t.Fatalf("fgets buffer = %q, want unchanged xy", got)
	}
}

func TestFgetsReadsUngetcPushback(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	stdin, ok := reg.LookupVariable("stdin", mem)
	if !ok {
		t.Fatal("missing stdin extern variable")
	}
	buf := mustAllocBytes(t, mem, "fgets:buf", []byte{0, 0, 0, 0}, false, blockLocal)
	ungetcFn, ok := reg.Lookup("ungetc")
	if !ok {
		t.Fatal("missing ungetc extern")
	}
	for _, ch := range []rune{'\n', 'i', 'H'} {
		if _, exit, err := ungetcFn(context.Background(), &ExternContext{Memory: mem}, []Value{IntValue(bytecode.TypeI32, int64(ch)), PtrValue(stdin)}); err != nil || exit != nil {
			t.Fatalf("ungetc %q exit=%#v err=%v", ch, exit, err)
		}
	}
	fn, ok := reg.Lookup("fgets")
	if !ok {
		t.Fatal("missing fgets extern")
	}
	ret, exit, err := fn(context.Background(), &ExternContext{Memory: mem}, []Value{
		ObjectAddrValue(buf),
		IntValue(bytecode.TypeI32, 4),
		PtrValue(stdin),
	})
	if err != nil || exit != nil {
		t.Fatalf("fgets ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypePtr || ret.Int != buf {
		t.Fatalf("fgets ret=%#v, want buffer pointer %#x", ret, buf)
	}
	got, err := mem.ReadCString(buf)
	if err != nil {
		t.Fatalf("ReadCString: %v", err)
	}
	if got != "Hi\n" {
		t.Fatalf("fgets buffer = %q, want Hi\\n", got)
	}
}

func TestFflushAcceptsHostHandleAndNull(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	stdout, ok := reg.LookupVariable("stdout", mem)
	if !ok {
		t.Fatal("missing stdout extern variable")
	}
	fn, ok := reg.Lookup("fflush")
	if !ok {
		t.Fatal("missing fflush extern")
	}
	for _, arg := range []Value{PtrValue(stdout), PtrValue(0)} {
		ret, exit, err := fn(context.Background(), &ExternContext{Memory: mem}, []Value{arg})
		if err != nil || exit != nil {
			t.Fatalf("fflush(%#v) ret=%#v exit=%#v err=%v", arg, ret, exit, err)
		}
		if ret.Type != bytecode.TypeI32 || ret.Int != 0 {
			t.Fatalf("fflush(%#v) = %#v, want i32 0", arg, ret)
		}
	}
}

func TestFcloseAcceptsHostHandle(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	stdout, ok := reg.LookupVariable("stdout", mem)
	if !ok {
		t.Fatal("missing stdout extern variable")
	}
	fn, ok := reg.Lookup("fclose")
	if !ok {
		t.Fatal("missing fclose extern")
	}
	ret, exit, err := fn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(stdout)})
	if err != nil || exit != nil {
		t.Fatalf("fclose ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeI32 || ret.Int != 0 {
		t.Fatalf("fclose ret=%#v, want i32 0", ret)
	}
}

func TestFcloseRejectsUnknownHandle(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	fn, ok := reg.Lookup("fclose")
	if !ok {
		t.Fatal("missing fclose extern")
	}
	_, _, err := fn(context.Background(), nil, []Value{PtrValue(0xdeadbeef)})
	if err == nil || !strings.Contains(err.Error(), "unknown stream handle") {
		t.Fatalf("fclose err = %v, want unknown stream handle", err)
	}
}

func TestStreamStatusExternsAcceptHostHandle(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	stdout, ok := reg.LookupVariable("stdout", mem)
	if !ok {
		t.Fatal("missing stdout extern variable")
	}

	for _, name := range []string{"ferror", "feof"} {
		fn, ok := reg.Lookup(name)
		if !ok {
			t.Fatalf("missing %s extern", name)
		}
		ret, exit, err := fn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(stdout)})
		if err != nil || exit != nil {
			t.Fatalf("%s ret=%#v exit=%#v err=%v", name, ret, exit, err)
		}
		if ret.Type != bytecode.TypeI32 || ret.Int != 0 {
			t.Fatalf("%s = %#v, want i32 0", name, ret)
		}
	}

	fn, ok := reg.Lookup("clearerr")
	if !ok {
		t.Fatal("missing clearerr extern")
	}
	ret, exit, err := fn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(stdout)})
	if err != nil || exit != nil {
		t.Fatalf("clearerr ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeVoid {
		t.Fatalf("clearerr ret=%#v, want void value", ret)
	}
}

func TestStreamStatusExternsRejectUnknownHandle(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	for _, name := range []string{"ferror", "feof", "clearerr"} {
		fn, ok := reg.Lookup(name)
		if !ok {
			t.Fatalf("missing %s extern", name)
		}
		_, _, err := fn(context.Background(), nil, []Value{PtrValue(0xdeadbeef)})
		if err == nil || !strings.Contains(err.Error(), "unknown stream handle") {
			t.Fatalf("%s err = %v, want unknown stream handle", name, err)
		}
	}
}

func TestFwriteWritesBytesToHostHandle(t *testing.T) {
	var out bytes.Buffer
	reg := DefaultExternRegistry(&out, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	stdout, ok := reg.LookupVariable("stdout", mem)
	if !ok {
		t.Fatal("missing stdout extern variable")
	}
	data := mustAllocBytes(t, mem, "fwrite:data", []byte("abcdef"), true, blockString)
	fn, ok := reg.Lookup("fwrite")
	if !ok {
		t.Fatal("missing fwrite extern")
	}
	ret, exit, err := fn(context.Background(), &ExternContext{Memory: mem}, []Value{
		ObjectAddrValue(data),
		UIntValue(bytecode.TypeU64, 2),
		UIntValue(bytecode.TypeU64, 3),
		PtrValue(stdout),
	})
	if err != nil || exit != nil {
		t.Fatalf("fwrite ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeU64 || ret.Int != 3 || out.String() != "abcdef" {
		t.Fatalf("fwrite ret=%#v output=%q, want u64 3 and abcdef", ret, out.String())
	}
}

func TestFwriteZeroSizeOrCountWritesNothing(t *testing.T) {
	var out bytes.Buffer
	reg := DefaultExternRegistry(&out, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	stdout, ok := reg.LookupVariable("stdout", mem)
	if !ok {
		t.Fatal("missing stdout extern variable")
	}
	fn, ok := reg.Lookup("fwrite")
	if !ok {
		t.Fatal("missing fwrite extern")
	}
	for _, args := range [][]Value{
		{ObjectAddrValue(0), UIntValue(bytecode.TypeU64, 0), UIntValue(bytecode.TypeU64, 3), PtrValue(stdout)},
		{ObjectAddrValue(0), UIntValue(bytecode.TypeU64, 3), UIntValue(bytecode.TypeU64, 0), PtrValue(stdout)},
	} {
		ret, exit, err := fn(context.Background(), &ExternContext{Memory: mem}, args)
		if err != nil || exit != nil {
			t.Fatalf("fwrite zero case ret=%#v exit=%#v err=%v", ret, exit, err)
		}
		if ret.Type != bytecode.TypeU64 || ret.Int != 0 {
			t.Fatalf("fwrite zero case ret=%#v, want u64 0", ret)
		}
	}
	if out.String() != "" {
		t.Fatalf("fwrite zero case output=%q, want empty", out.String())
	}
}

func TestFreadEmptyInputWritesNothing(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	stdin, ok := reg.LookupVariable("stdin", mem)
	if !ok {
		t.Fatal("missing stdin extern variable")
	}
	buf := mustAllocBytes(t, mem, "fread:buf", []byte("xy\x00"), false, blockLocal)
	fn, ok := reg.Lookup("fread")
	if !ok {
		t.Fatal("missing fread extern")
	}
	ret, exit, err := fn(context.Background(), &ExternContext{Memory: mem}, []Value{
		ObjectAddrValue(buf),
		UIntValue(bytecode.TypeU64, 1),
		UIntValue(bytecode.TypeU64, 2),
		PtrValue(stdin),
	})
	if err != nil || exit != nil {
		t.Fatalf("fread ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeU64 || ret.Int != 0 {
		t.Fatalf("fread ret=%#v, want u64 0", ret)
	}
	got, err := mem.ReadCString(buf)
	if err != nil {
		t.Fatalf("ReadCString: %v", err)
	}
	if got != "xy" {
		t.Fatalf("fread buffer = %q, want unchanged xy", got)
	}
}

func TestFreadReadsUngetcPushback(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	stdin, ok := reg.LookupVariable("stdin", mem)
	if !ok {
		t.Fatal("missing stdin extern variable")
	}
	buf := mustAllocBytes(t, mem, "fread:buf", []byte{0, 0, 0, 0}, false, blockLocal)
	ungetcFn, ok := reg.Lookup("ungetc")
	if !ok {
		t.Fatal("missing ungetc extern")
	}
	for _, ch := range []rune{'C', 'B', 'A'} {
		if _, exit, err := ungetcFn(context.Background(), &ExternContext{Memory: mem}, []Value{IntValue(bytecode.TypeI32, int64(ch)), PtrValue(stdin)}); err != nil || exit != nil {
			t.Fatalf("ungetc %q exit=%#v err=%v", ch, exit, err)
		}
	}
	fn, ok := reg.Lookup("fread")
	if !ok {
		t.Fatal("missing fread extern")
	}
	ret, exit, err := fn(context.Background(), &ExternContext{Memory: mem}, []Value{
		ObjectAddrValue(buf),
		UIntValue(bytecode.TypeU64, 1),
		UIntValue(bytecode.TypeU64, 3),
		PtrValue(stdin),
	})
	if err != nil || exit != nil {
		t.Fatalf("fread ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeU64 || ret.Int != 3 {
		t.Fatalf("fread ret=%#v, want u64 3", ret)
	}
	block, off, err := mem.rangeAccess(buf, 3, false)
	if err != nil {
		t.Fatalf("rangeAccess: %v", err)
	}
	if got := string(block.data[off : off+3]); got != "ABC" {
		t.Fatalf("fread bytes = %q, want ABC", got)
	}
}

func TestFputsWritesCStringToStderrHostHandle(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer
	reg := DefaultExternRegistry(&out, &errOut)
	mem := NewMemory(bytecode.DefaultTarget())
	addr := mustAllocBytes(t, mem, "string:0", []byte("hello\x00"), true, blockString)
	if _, ok := reg.LookupVariable("stdin", mem); !ok {
		t.Fatal("missing stdin extern variable")
	}
	stderr, ok := reg.LookupVariable("stderr", mem)
	if !ok {
		t.Fatal("missing stderr extern variable")
	}
	fn, _ := reg.Lookup("fputs")
	ret, exit, err := fn(context.Background(), &ExternContext{Memory: mem}, []Value{ObjectAddrValue(addr), PtrValue(stderr)})
	if err != nil || exit != nil {
		t.Fatalf("fputs ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if out.String() != "" {
		t.Fatalf("stdout output = %q, want empty", out.String())
	}
	if errOut.String() != "hello" {
		t.Fatalf("stderr output = %q", errOut.String())
	}
}

func TestFputsWritesCStringToLoadedStderrHostHandle(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer
	reg := DefaultExternRegistry(&out, &errOut)
	target := bytecode.DefaultTarget()
	mem := NewMemory(target)
	addr := mustAllocBytes(t, mem, "string:0", []byte("hello\x00"), true, blockString)
	stderrAddr, ok := reg.LookupVariable("stderr", mem)
	if !ok {
		t.Fatal("missing stderr extern variable")
	}
	loaded, err := mem.Load(stderrAddr, bytecode.TypePtr, target.PointerAlign)
	if err != nil {
		t.Fatalf("Load(stderr): %v", err)
	}
	fn, _ := reg.Lookup("fputs")
	ret, exit, err := fn(context.Background(), &ExternContext{Memory: mem}, []Value{ObjectAddrValue(addr), loaded})
	if err != nil || exit != nil {
		t.Fatalf("fputs ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if out.String() != "" {
		t.Fatalf("stdout output = %q, want empty", out.String())
	}
	if errOut.String() != "hello" {
		t.Fatalf("stderr output = %q", errOut.String())
	}
}

func TestFputsUnknownStreamHandleReturnsError(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer
	reg := DefaultExternRegistry(&out, &errOut)
	mem := NewMemory(bytecode.DefaultTarget())
	addr := mustAllocBytes(t, mem, "string:0", []byte("hello\x00"), true, blockString)
	fn, _ := reg.Lookup("fputs")
	_, _, err := fn(context.Background(), &ExternContext{Memory: mem}, []Value{ObjectAddrValue(addr), PtrValue(0xdeadbeef)})
	if err == nil || !strings.Contains(err.Error(), "unknown stream handle") {
		t.Fatalf("fputs err = %v, want unknown stream handle", err)
	}
	if out.String() != "" || errOut.String() != "" {
		t.Fatalf("stdout=%q stderr=%q, want no output", out.String(), errOut.String())
	}
}

func TestStrcmpComparesCStrings(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	fn, ok := reg.Lookup("strcmp")
	if !ok {
		t.Fatal("missing strcmp extern")
	}
	mem := NewMemory(bytecode.DefaultTarget())
	tests := []struct {
		name string
		a    string
		b    string
		want int64
	}{
		{name: "equal", a: "abc", b: "abc", want: 0},
		{name: "left less", a: "abc", b: "abd", want: -1},
		{name: "left greater", a: "abd", b: "abc", want: 1},
		{name: "prefix less", a: "ab", b: "abc", want: -1},
		{name: "prefix greater", a: "abc", b: "ab", want: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := mustAllocBytes(t, mem, "strcmp:a", []byte(tt.a+"\x00"), true, blockString)
			b := mustAllocBytes(t, mem, "strcmp:b", []byte(tt.b+"\x00"), true, blockString)
			ret, exit, err := fn(context.Background(), &ExternContext{Memory: mem}, []Value{ObjectAddrValue(a), ObjectAddrValue(b)})
			if err != nil || exit != nil {
				t.Fatalf("strcmp ret=%#v exit=%#v err=%v", ret, exit, err)
			}
			if ret.Type != bytecode.TypeI32 || int64(int32(ret.Int)) != tt.want {
				t.Fatalf("strcmp(%q, %q) = %#v, want i32 %d", tt.a, tt.b, ret, tt.want)
			}
		})
	}
}

func TestStrcmpRequiresMemory(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	fn, ok := reg.Lookup("strcmp")
	if !ok {
		t.Fatal("missing strcmp extern")
	}
	_, _, err := fn(context.Background(), nil, []Value{PtrValue(0), PtrValue(0)})
	if err == nil || !strings.Contains(err.Error(), "strcmp requires memory") {
		t.Fatalf("strcmp err = %v, want memory error", err)
	}
}

func TestMemcmpComparesBytes(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	fn, ok := reg.Lookup("memcmp")
	if !ok {
		t.Fatal("missing memcmp extern")
	}
	mem := NewMemory(bytecode.DefaultTarget())
	left := mustAllocBytes(t, mem, "memcmp:left", []byte{0, 1, 2, 3}, true, blockString)
	same := mustAllocBytes(t, mem, "memcmp:same", []byte{0, 1, 2, 9}, true, blockString)
	diff := mustAllocBytes(t, mem, "memcmp:diff", []byte{0, 1, 4, 3}, true, blockString)
	less := mustAllocBytes(t, mem, "memcmp:less", []byte{0, 1, 0, 3}, true, blockString)
	tests := []struct {
		name  string
		right uint64
		n     int64
		want  int64
	}{
		{name: "zero length", right: diff, n: 0, want: 0},
		{name: "equal prefix", right: same, n: 3, want: 0},
		{name: "left less", right: diff, n: 4, want: -1},
		{name: "left greater", right: less, n: 4, want: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ret, exit, err := fn(context.Background(), &ExternContext{Memory: mem}, []Value{ObjectAddrValue(left), ObjectAddrValue(tt.right), UIntValue(bytecode.TypeU64, uint64(tt.n))})
			if err != nil || exit != nil {
				t.Fatalf("memcmp ret=%#v exit=%#v err=%v", ret, exit, err)
			}
			if ret.Type != bytecode.TypeI32 || int64(int32(ret.Int)) != tt.want {
				t.Fatalf("memcmp = %#v, want i32 %d", ret, tt.want)
			}
		})
	}
}

func TestPrintfExternsWriteFormattedOutput(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer
	reg := DefaultExternRegistry(&out, &errOut)
	mem := NewMemory(bytecode.DefaultTarget())

	fmtAddr := mustAllocBytes(t, mem, "fmt:printf", []byte("x=%d %s\x00"), true, blockString)
	strAddr := mustAllocBytes(t, mem, "str:printf", []byte("ok\x00"), true, blockString)
	fn, ok := reg.Lookup("__builtin_printf")
	if !ok {
		t.Fatal("missing __builtin_printf extern")
	}
	ret, exit, callErr := fn(context.Background(), &ExternContext{Memory: mem, Stdout: &out, Stderr: &errOut}, []Value{
		ObjectAddrValue(fmtAddr),
		IntValue(bytecode.TypeI32, 7),
		ObjectAddrValue(strAddr),
	})
	if callErr != nil || exit != nil {
		t.Fatalf("__builtin_printf ret=%#v exit=%#v err=%v", ret, exit, callErr)
	}
	if ret.Type != bytecode.TypeI32 || ret.Int != 6 || out.String() != "x=7 ok" {
		t.Fatalf("__builtin_printf ret=%#v output=%q, want i32 6 and x=7 ok", ret, out.String())
	}

	stderrAddr, ok := reg.LookupVariable("stderr", mem)
	if !ok {
		t.Fatal("missing stderr extern variable")
	}
	fmtErrAddr := mustAllocBytes(t, mem, "fmt:fprintf", []byte("e=%u\x00"), true, blockString)
	fn, ok = reg.Lookup("__builtin_fprintf")
	if !ok {
		t.Fatal("missing __builtin_fprintf extern")
	}
	ret, exit, callErr = fn(context.Background(), &ExternContext{Memory: mem, Stdout: &out, Stderr: &errOut}, []Value{
		ObjectAddrValue(stderrAddr),
		ObjectAddrValue(fmtErrAddr),
		UIntValue(bytecode.TypeU32, 9),
	})
	if callErr != nil || exit != nil {
		t.Fatalf("__builtin_fprintf ret=%#v exit=%#v err=%v", ret, exit, callErr)
	}
	if ret.Type != bytecode.TypeI32 || ret.Int != 3 || errOut.String() != "e=9" {
		t.Fatalf("__builtin_fprintf ret=%#v stderr=%q, want i32 3 and e=9", ret, errOut.String())
	}
}

func TestCheckedPrintfExternsWriteFormattedOutput(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer
	reg := DefaultExternRegistry(&out, &errOut)
	mem := NewMemory(bytecode.DefaultTarget())

	fmtAddr := mustAllocBytes(t, mem, "fmt:printf_chk", []byte("x=%d %s\x00"), true, blockString)
	strAddr := mustAllocBytes(t, mem, "str:printf_chk", []byte("ok\x00"), true, blockString)
	fn, ok := reg.Lookup("__builtin___printf_chk")
	if !ok {
		t.Fatal("missing __builtin___printf_chk extern")
	}
	ret, exit, callErr := fn(context.Background(), &ExternContext{Memory: mem, Stdout: &out, Stderr: &errOut}, []Value{
		IntValue(bytecode.TypeI32, 0),
		ObjectAddrValue(fmtAddr),
		IntValue(bytecode.TypeI32, 7),
		ObjectAddrValue(strAddr),
	})
	if callErr != nil || exit != nil {
		t.Fatalf("__builtin___printf_chk ret=%#v exit=%#v err=%v", ret, exit, callErr)
	}
	if ret.Type != bytecode.TypeI32 || ret.Int != 6 || out.String() != "x=7 ok" {
		t.Fatalf("__builtin___printf_chk ret=%#v output=%q, want i32 6 and x=7 ok", ret, out.String())
	}

	stderrAddr, ok := reg.LookupVariable("stderr", mem)
	if !ok {
		t.Fatal("missing stderr extern variable")
	}
	fmtErrAddr := mustAllocBytes(t, mem, "fmt:fprintf_chk", []byte("e=%u\x00"), true, blockString)
	fn, ok = reg.Lookup("__builtin___fprintf_chk")
	if !ok {
		t.Fatal("missing __builtin___fprintf_chk extern")
	}
	ret, exit, callErr = fn(context.Background(), &ExternContext{Memory: mem, Stdout: &out, Stderr: &errOut}, []Value{
		ObjectAddrValue(stderrAddr),
		IntValue(bytecode.TypeI32, 0),
		ObjectAddrValue(fmtErrAddr),
		UIntValue(bytecode.TypeU32, 9),
	})
	if callErr != nil || exit != nil {
		t.Fatalf("__builtin___fprintf_chk ret=%#v exit=%#v err=%v", ret, exit, callErr)
	}
	if ret.Type != bytecode.TypeI32 || ret.Int != 3 || errOut.String() != "e=9" {
		t.Fatalf("__builtin___fprintf_chk ret=%#v stderr=%q, want i32 3 and e=9", ret, errOut.String())
	}
}

func TestVFormatExternsWriteLiteralOutput(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer
	reg := DefaultExternRegistry(&out, &errOut)
	mem := NewMemory(bytecode.DefaultTarget())

	fmtAddr := mustAllocBytes(t, mem, "fmt:vprintf", []byte("v%%\x00"), true, blockString)
	fn, ok := reg.Lookup("__builtin_vprintf")
	if !ok {
		t.Fatal("missing __builtin_vprintf extern")
	}
	ret, exit, callErr := fn(context.Background(), &ExternContext{Memory: mem, Stdout: &out, Stderr: &errOut}, []Value{
		ObjectAddrValue(fmtAddr),
		PtrValue(0),
	})
	if callErr != nil || exit != nil {
		t.Fatalf("__builtin_vprintf ret=%#v exit=%#v err=%v", ret, exit, callErr)
	}
	if ret.Type != bytecode.TypeI32 || ret.Int != 2 || out.String() != "v%" {
		t.Fatalf("__builtin_vprintf ret=%#v output=%q, want i32 2 and v%%", ret, out.String())
	}

	stderrAddr, ok := reg.LookupVariable("stderr", mem)
	if !ok {
		t.Fatal("missing stderr extern variable")
	}
	fmtErrAddr := mustAllocBytes(t, mem, "fmt:vfprintf", []byte("err\x00"), true, blockString)
	fn, ok = reg.Lookup("__builtin_vfprintf")
	if !ok {
		t.Fatal("missing __builtin_vfprintf extern")
	}
	ret, exit, callErr = fn(context.Background(), &ExternContext{Memory: mem, Stdout: &out, Stderr: &errOut}, []Value{
		ObjectAddrValue(stderrAddr),
		ObjectAddrValue(fmtErrAddr),
		PtrValue(0),
	})
	if callErr != nil || exit != nil {
		t.Fatalf("__builtin_vfprintf ret=%#v exit=%#v err=%v", ret, exit, callErr)
	}
	if ret.Type != bytecode.TypeI32 || ret.Int != 3 || errOut.String() != "err" {
		t.Fatalf("__builtin_vfprintf ret=%#v stderr=%q, want i32 3 and err", ret, errOut.String())
	}
}

func TestCheckedVFormatExternsWriteLiteralOutput(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer
	reg := DefaultExternRegistry(&out, &errOut)
	mem := NewMemory(bytecode.DefaultTarget())

	fmtAddr := mustAllocBytes(t, mem, "fmt:vprintf-chk", []byte("chk\x00"), true, blockString)
	fn, ok := reg.Lookup("__builtin___vprintf_chk")
	if !ok {
		t.Fatal("missing __builtin___vprintf_chk extern")
	}
	ret, exit, callErr := fn(context.Background(), &ExternContext{Memory: mem, Stdout: &out, Stderr: &errOut}, []Value{
		IntValue(bytecode.TypeI32, 0),
		ObjectAddrValue(fmtAddr),
		PtrValue(0),
	})
	if callErr != nil || exit != nil {
		t.Fatalf("__builtin___vprintf_chk ret=%#v exit=%#v err=%v", ret, exit, callErr)
	}
	if ret.Type != bytecode.TypeI32 || ret.Int != 3 || out.String() != "chk" {
		t.Fatalf("__builtin___vprintf_chk ret=%#v output=%q, want i32 3 and chk", ret, out.String())
	}

	stderrAddr, ok := reg.LookupVariable("stderr", mem)
	if !ok {
		t.Fatal("missing stderr extern variable")
	}
	fmtErrAddr := mustAllocBytes(t, mem, "fmt:vfprintf-chk", []byte("cerr\x00"), true, blockString)
	fn, ok = reg.Lookup("__builtin___vfprintf_chk")
	if !ok {
		t.Fatal("missing __builtin___vfprintf_chk extern")
	}
	ret, exit, callErr = fn(context.Background(), &ExternContext{Memory: mem, Stdout: &out, Stderr: &errOut}, []Value{
		ObjectAddrValue(stderrAddr),
		IntValue(bytecode.TypeI32, 0),
		ObjectAddrValue(fmtErrAddr),
		PtrValue(0),
	})
	if callErr != nil || exit != nil {
		t.Fatalf("__builtin___vfprintf_chk ret=%#v exit=%#v err=%v", ret, exit, callErr)
	}
	if ret.Type != bytecode.TypeI32 || ret.Int != 4 || errOut.String() != "cerr" {
		t.Fatalf("__builtin___vfprintf_chk ret=%#v stderr=%q, want i32 4 and cerr", ret, errOut.String())
	}
}

func TestCheckedVBufferFormatExternsWriteLiteralOutput(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	bufAddr := mustAlloc(t, mem, "buf:vformat-chk", 16, 1, false, blockLocal)
	fmtAddr := mustAllocBytes(t, mem, "fmt:vsprintf-chk", []byte("buf%%\x00"), true, blockString)

	fn, ok := reg.Lookup("__builtin___vsprintf_chk")
	if !ok {
		t.Fatal("missing __builtin___vsprintf_chk extern")
	}
	ret, exit, callErr := fn(context.Background(), &ExternContext{Memory: mem}, []Value{
		ObjectAddrValue(bufAddr),
		IntValue(bytecode.TypeI32, 0),
		UIntValue(bytecode.TypeU64, 16),
		ObjectAddrValue(fmtAddr),
		PtrValue(0),
	})
	if callErr != nil || exit != nil {
		t.Fatalf("__builtin___vsprintf_chk ret=%#v exit=%#v err=%v", ret, exit, callErr)
	}
	got, err := mem.ReadCString(bufAddr)
	if err != nil {
		t.Fatalf("ReadCString: %v", err)
	}
	if ret.Type != bytecode.TypeI32 || ret.Int != 4 || got != "buf%" {
		t.Fatalf("__builtin___vsprintf_chk ret=%#v output=%q, want i32 4 and buf%%", ret, got)
	}

	fmtNAddr := mustAllocBytes(t, mem, "fmt:vsnprintf-chk", []byte("abcdef\x00"), true, blockString)
	fn, ok = reg.Lookup("__builtin___vsnprintf_chk")
	if !ok {
		t.Fatal("missing __builtin___vsnprintf_chk extern")
	}
	ret, exit, callErr = fn(context.Background(), &ExternContext{Memory: mem}, []Value{
		ObjectAddrValue(bufAddr),
		UIntValue(bytecode.TypeU64, 4),
		IntValue(bytecode.TypeI32, 0),
		UIntValue(bytecode.TypeU64, 16),
		ObjectAddrValue(fmtNAddr),
		PtrValue(0),
	})
	if callErr != nil || exit != nil {
		t.Fatalf("__builtin___vsnprintf_chk ret=%#v exit=%#v err=%v", ret, exit, callErr)
	}
	got, err = mem.ReadCString(bufAddr)
	if err != nil {
		t.Fatalf("ReadCString: %v", err)
	}
	if ret.Type != bytecode.TypeI32 || ret.Int != 6 || got != "abc" {
		t.Fatalf("__builtin___vsnprintf_chk ret=%#v output=%q, want i32 6 and abc", ret, got)
	}
}

func TestFormatExternsSupportIntegerLengthModifiers(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	bufAddr := mustAlloc(t, mem, "buf:format-length", 32, 1, false, blockLocal)
	fmtAddr := mustAllocBytes(t, mem, "fmt:format-length", []byte("%ld %llu %zu\x00"), true, blockString)
	fn, ok := reg.Lookup("__builtin_sprintf")
	if !ok {
		t.Fatal("missing __builtin_sprintf extern")
	}
	ret, exit, callErr := fn(context.Background(), &ExternContext{Memory: mem}, []Value{
		ObjectAddrValue(bufAddr),
		ObjectAddrValue(fmtAddr),
		IntValue(bytecode.TypeI64, -7),
		UIntValue(bytecode.TypeU64, 42),
		UIntValue(bytecode.TypeU64, 3),
	})
	if callErr != nil || exit != nil {
		t.Fatalf("__builtin_sprintf ret=%#v exit=%#v err=%v", ret, exit, callErr)
	}
	got, err := mem.ReadCString(bufAddr)
	if err != nil {
		t.Fatalf("ReadCString: %v", err)
	}
	if ret.Type != bytecode.TypeI32 || ret.Int != 7 || got != "-7 42 3" {
		t.Fatalf("__builtin_sprintf ret=%#v output=%q, want i32 7 and -7 42 3", ret, got)
	}
}

func TestFormatExternsSupportIntegerRadixFormats(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	bufAddr := mustAlloc(t, mem, "buf:format-radix", 32, 1, false, blockLocal)
	fmtAddr := mustAllocBytes(t, mem, "fmt:format-radix", []byte("%x %X %o\x00"), true, blockString)
	fn, ok := reg.Lookup("__builtin_sprintf")
	if !ok {
		t.Fatal("missing __builtin_sprintf extern")
	}
	ret, exit, callErr := fn(context.Background(), &ExternContext{Memory: mem}, []Value{
		ObjectAddrValue(bufAddr),
		ObjectAddrValue(fmtAddr),
		UIntValue(bytecode.TypeU32, 48879),
		UIntValue(bytecode.TypeU32, 48879),
		UIntValue(bytecode.TypeU32, 511),
	})
	if callErr != nil || exit != nil {
		t.Fatalf("__builtin_sprintf ret=%#v exit=%#v err=%v", ret, exit, callErr)
	}
	got, err := mem.ReadCString(bufAddr)
	if err != nil {
		t.Fatalf("ReadCString: %v", err)
	}
	if ret.Type != bytecode.TypeI32 || ret.Int != 13 || got != "beef BEEF 777" {
		t.Fatalf("__builtin_sprintf ret=%#v output=%q, want i32 13 and beef BEEF 777", ret, got)
	}
}

func TestFormatExternsSupportPointerFormat(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	bufAddr := mustAlloc(t, mem, "buf:format-pointer", 32, 1, false, blockLocal)
	fmtAddr := mustAllocBytes(t, mem, "fmt:format-pointer", []byte("%p\x00"), true, blockString)
	fn, ok := reg.Lookup("__builtin_sprintf")
	if !ok {
		t.Fatal("missing __builtin_sprintf extern")
	}
	ret, exit, callErr := fn(context.Background(), &ExternContext{Memory: mem}, []Value{
		ObjectAddrValue(bufAddr),
		ObjectAddrValue(fmtAddr),
		PtrValue(0x1234),
	})
	if callErr != nil || exit != nil {
		t.Fatalf("__builtin_sprintf ret=%#v exit=%#v err=%v", ret, exit, callErr)
	}
	got, err := mem.ReadCString(bufAddr)
	if err != nil {
		t.Fatalf("ReadCString: %v", err)
	}
	if ret.Type != bytecode.TypeI32 || ret.Int != 6 || got != "0x1234" {
		t.Fatalf("__builtin_sprintf ret=%#v output=%q, want i32 6 and 0x1234", ret, got)
	}
}

func TestFormatExternsSupportBasicWidthFlags(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	bufAddr := mustAlloc(t, mem, "buf:format-width", 32, 1, false, blockLocal)
	fmtAddr := mustAllocBytes(t, mem, "fmt:format-width", []byte("%5d|%-4s|%05u\x00"), true, blockString)
	strAddr := mustAllocBytes(t, mem, "str:format-width", []byte("xy\x00"), true, blockString)
	fn, ok := reg.Lookup("__builtin_sprintf")
	if !ok {
		t.Fatal("missing __builtin_sprintf extern")
	}
	ret, exit, callErr := fn(context.Background(), &ExternContext{Memory: mem}, []Value{
		ObjectAddrValue(bufAddr),
		ObjectAddrValue(fmtAddr),
		IntValue(bytecode.TypeI32, 7),
		ObjectAddrValue(strAddr),
		UIntValue(bytecode.TypeU32, 42),
	})
	if callErr != nil || exit != nil {
		t.Fatalf("__builtin_sprintf ret=%#v exit=%#v err=%v", ret, exit, callErr)
	}
	got, err := mem.ReadCString(bufAddr)
	if err != nil {
		t.Fatalf("ReadCString: %v", err)
	}
	if ret.Type != bytecode.TypeI32 || ret.Int != 16 || got != "    7|xy  |00042" {
		t.Fatalf("__builtin_sprintf ret=%#v output=%q, want i32 16 and padded output", ret, got)
	}
}

func TestFormatExternsSupportPrecision(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	bufAddr := mustAlloc(t, mem, "buf:format-precision", 32, 1, false, blockLocal)
	fmtAddr := mustAllocBytes(t, mem, "fmt:format-precision", []byte("%.3s|%.4d|%7.4u\x00"), true, blockString)
	strAddr := mustAllocBytes(t, mem, "str:format-precision", []byte("abcdef\x00"), true, blockString)
	fn, ok := reg.Lookup("__builtin_sprintf")
	if !ok {
		t.Fatal("missing __builtin_sprintf extern")
	}
	ret, exit, callErr := fn(context.Background(), &ExternContext{Memory: mem}, []Value{
		ObjectAddrValue(bufAddr),
		ObjectAddrValue(fmtAddr),
		ObjectAddrValue(strAddr),
		IntValue(bytecode.TypeI32, 7),
		UIntValue(bytecode.TypeU32, 42),
	})
	if callErr != nil || exit != nil {
		t.Fatalf("__builtin_sprintf ret=%#v exit=%#v err=%v", ret, exit, callErr)
	}
	got, err := mem.ReadCString(bufAddr)
	if err != nil {
		t.Fatalf("ReadCString: %v", err)
	}
	if ret.Type != bytecode.TypeI32 || ret.Int != 16 || got != "abc|0007|   0042" {
		t.Fatalf("__builtin_sprintf ret=%#v output=%q, want i32 16 and precision output", ret, got)
	}
}

func TestFormatExternsSupportSignAndAlternateFlags(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	bufAddr := mustAlloc(t, mem, "buf:format-flags", 40, 1, false, blockLocal)
	fmtAddr := mustAllocBytes(t, mem, "fmt:format-flags", []byte("%+d|% d|%#x|%#X|%#o\x00"), true, blockString)
	fn, ok := reg.Lookup("__builtin_sprintf")
	if !ok {
		t.Fatal("missing __builtin_sprintf extern")
	}
	ret, exit, callErr := fn(context.Background(), &ExternContext{Memory: mem}, []Value{
		ObjectAddrValue(bufAddr),
		ObjectAddrValue(fmtAddr),
		IntValue(bytecode.TypeI32, 7),
		IntValue(bytecode.TypeI32, 7),
		UIntValue(bytecode.TypeU32, 48879),
		UIntValue(bytecode.TypeU32, 48879),
		UIntValue(bytecode.TypeU32, 511),
	})
	if callErr != nil || exit != nil {
		t.Fatalf("__builtin_sprintf ret=%#v exit=%#v err=%v", ret, exit, callErr)
	}
	got, err := mem.ReadCString(bufAddr)
	if err != nil {
		t.Fatalf("ReadCString: %v", err)
	}
	if ret.Type != bytecode.TypeI32 || ret.Int != 24 || got != "+7| 7|0xbeef|0XBEEF|0777" {
		t.Fatalf("__builtin_sprintf ret=%#v output=%q, want i32 24 and flag output", ret, got)
	}
}

func TestFormatExternsSupportFloatFormats(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	bufAddr := mustAlloc(t, mem, "buf:format-float", 40, 1, false, blockLocal)
	fmtAddr := mustAllocBytes(t, mem, "fmt:format-float", []byte("%f|%.2f|%.1e|%.1E\x00"), true, blockString)
	fn, ok := reg.Lookup("__builtin_sprintf")
	if !ok {
		t.Fatal("missing __builtin_sprintf extern")
	}
	ret, exit, callErr := fn(context.Background(), &ExternContext{Memory: mem}, []Value{
		ObjectAddrValue(bufAddr),
		ObjectAddrValue(fmtAddr),
		FloatValue(bytecode.TypeF64, 1.5),
		FloatValue(bytecode.TypeF64, 1.25),
		FloatValue(bytecode.TypeF64, 12),
		FloatValue(bytecode.TypeF64, 12),
	})
	if callErr != nil || exit != nil {
		t.Fatalf("__builtin_sprintf ret=%#v exit=%#v err=%v", ret, exit, callErr)
	}
	got, err := mem.ReadCString(bufAddr)
	if err != nil {
		t.Fatalf("ReadCString: %v", err)
	}
	if ret.Type != bytecode.TypeI32 || ret.Int != 29 || got != "1.500000|1.25|1.2e+01|1.2E+01" {
		t.Fatalf("__builtin_sprintf ret=%#v output=%q, want i32 29 and float output", ret, got)
	}
}

func TestFormatExternsSupportHexFloatFormats(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	bufAddr := mustAlloc(t, mem, "buf:format-hex-float", 32, 1, false, blockLocal)
	fmtAddr := mustAllocBytes(t, mem, "fmt:format-hex-float", []byte("%.1a|%.1A\x00"), true, blockString)
	fn, ok := reg.Lookup("__builtin_sprintf")
	if !ok {
		t.Fatal("missing __builtin_sprintf extern")
	}
	ret, exit, callErr := fn(context.Background(), &ExternContext{Memory: mem}, []Value{
		ObjectAddrValue(bufAddr),
		ObjectAddrValue(fmtAddr),
		FloatValue(bytecode.TypeF64, 1.5),
		FloatValue(bytecode.TypeF64, 1.5),
	})
	if callErr != nil || exit != nil {
		t.Fatalf("__builtin_sprintf ret=%#v exit=%#v err=%v", ret, exit, callErr)
	}
	got, err := mem.ReadCString(bufAddr)
	if err != nil {
		t.Fatalf("ReadCString: %v", err)
	}
	if ret.Type != bytecode.TypeI32 || ret.Int != 19 || got != "0x1.8p+00|0X1.8P+00" {
		t.Fatalf("__builtin_sprintf ret=%#v output=%q, want i32 19 and hex float output", ret, got)
	}
}

func TestFormatExternsSupportDynamicWidthAndPrecision(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	bufAddr := mustAlloc(t, mem, "buf:format-dynamic-width", 32, 1, false, blockLocal)
	fmtAddr := mustAllocBytes(t, mem, "fmt:format-dynamic-width", []byte("%*.*s|%0*d\x00"), true, blockString)
	strAddr := mustAllocBytes(t, mem, "str:format-dynamic-width", []byte("abcdef\x00"), true, blockString)
	fn, ok := reg.Lookup("__builtin_sprintf")
	if !ok {
		t.Fatal("missing __builtin_sprintf extern")
	}
	ret, exit, callErr := fn(context.Background(), &ExternContext{Memory: mem}, []Value{
		ObjectAddrValue(bufAddr),
		ObjectAddrValue(fmtAddr),
		IntValue(bytecode.TypeI32, 5),
		IntValue(bytecode.TypeI32, 3),
		ObjectAddrValue(strAddr),
		IntValue(bytecode.TypeI32, 4),
		IntValue(bytecode.TypeI32, 7),
	})
	if callErr != nil || exit != nil {
		t.Fatalf("__builtin_sprintf ret=%#v exit=%#v err=%v", ret, exit, callErr)
	}
	got, err := mem.ReadCString(bufAddr)
	if err != nil {
		t.Fatalf("ReadCString: %v", err)
	}
	if ret.Type != bytecode.TypeI32 || ret.Int != 10 || got != "  abc|0007" {
		t.Fatalf("__builtin_sprintf ret=%#v output=%q, want i32 10 and dynamic width output", ret, got)
	}
}

func TestFormatExternsSupportWriteCount(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	bufAddr := mustAlloc(t, mem, "buf:format-count", 16, 1, false, blockLocal)
	countAddr := mustAlloc(t, mem, "count:format-count", 4, 4, false, blockLocal)
	fmtAddr := mustAllocBytes(t, mem, "fmt:format-count", []byte("ab%ncd\x00"), true, blockString)
	fn, ok := reg.Lookup("__builtin_sprintf")
	if !ok {
		t.Fatal("missing __builtin_sprintf extern")
	}
	ret, exit, callErr := fn(context.Background(), &ExternContext{Memory: mem}, []Value{
		ObjectAddrValue(bufAddr),
		ObjectAddrValue(fmtAddr),
		PtrValue(countAddr),
	})
	if callErr != nil || exit != nil {
		t.Fatalf("__builtin_sprintf ret=%#v exit=%#v err=%v", ret, exit, callErr)
	}
	got, err := mem.ReadCString(bufAddr)
	if err != nil {
		t.Fatalf("ReadCString: %v", err)
	}
	count, err := mem.Load(countAddr, bytecode.TypeI32, 4)
	if err != nil {
		t.Fatalf("Load count: %v", err)
	}
	if ret.Type != bytecode.TypeI32 || ret.Int != 4 || got != "abcd" || int32(count.Int) != 2 {
		t.Fatalf("__builtin_sprintf ret=%#v output=%q count=%#v, want i32 4 abcd and count 2", ret, got, count)
	}
}

func TestFormatExternsSupportSizedWriteCount(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	bufAddr := mustAlloc(t, mem, "buf:format-sized-count", 16, 1, false, blockLocal)
	count8 := mustAlloc(t, mem, "count8:format-sized-count", 1, 1, false, blockLocal)
	count16 := mustAlloc(t, mem, "count16:format-sized-count", 2, 2, false, blockLocal)
	count64 := mustAlloc(t, mem, "count64:format-sized-count", 8, 8, false, blockLocal)
	fmtAddr := mustAllocBytes(t, mem, "fmt:format-sized-count", []byte("a%hhnb%hnc%lln\x00"), true, blockString)
	fn, ok := reg.Lookup("__builtin_sprintf")
	if !ok {
		t.Fatal("missing __builtin_sprintf extern")
	}
	ret, exit, callErr := fn(context.Background(), &ExternContext{Memory: mem}, []Value{
		ObjectAddrValue(bufAddr),
		ObjectAddrValue(fmtAddr),
		PtrValue(count8),
		PtrValue(count16),
		PtrValue(count64),
	})
	if callErr != nil || exit != nil {
		t.Fatalf("__builtin_sprintf ret=%#v exit=%#v err=%v", ret, exit, callErr)
	}
	got, err := mem.ReadCString(bufAddr)
	if err != nil {
		t.Fatalf("ReadCString: %v", err)
	}
	v8, err := mem.Load(count8, bytecode.TypeI8, 1)
	if err != nil {
		t.Fatalf("Load count8: %v", err)
	}
	v16, err := mem.Load(count16, bytecode.TypeI16, 2)
	if err != nil {
		t.Fatalf("Load count16: %v", err)
	}
	v64, err := mem.Load(count64, bytecode.TypeI64, 8)
	if err != nil {
		t.Fatalf("Load count64: %v", err)
	}
	if ret.Type != bytecode.TypeI32 || ret.Int != 3 || got != "abc" || v8.Int != 1 || v16.Int != 2 || v64.Int != 3 {
		t.Fatalf("__builtin_sprintf ret=%#v output=%q counts=%#v/%#v/%#v, want abc and 1/2/3", ret, got, v8, v16, v64)
	}
}

func TestFenvExternsAreNoOps(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	for _, name := range []string{"feclearexcept", "fetestexcept"} {
		t.Run(name, func(t *testing.T) {
			fn, ok := reg.Lookup(name)
			if !ok {
				t.Fatalf("missing %s extern", name)
			}
			ret, exit, err := fn(context.Background(), nil, []Value{IntValue(bytecode.TypeI32, 0)})
			if err != nil || exit != nil {
				t.Fatalf("%s ret=%#v exit=%#v err=%v", name, ret, exit, err)
			}
			if ret.Type != bytecode.TypeI32 || ret.Int != 0 {
				t.Fatalf("%s ret = %#v, want i32 0", name, ret)
			}
		})
	}
}

func TestMathClassificationExterns(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	tests := []struct {
		name string
		fn   string
		arg  Value
		want int64
	}{
		{name: "float subnormal", fn: "__cvm_fpclassifyf", arg: FloatValue(bytecode.TypeF32, float64(math.SmallestNonzeroFloat32)), want: 3},
		{name: "double subnormal", fn: "__cvm_fpclassify", arg: FloatValue(bytecode.TypeF64, math.SmallestNonzeroFloat64*2), want: 3},
		{name: "double normal", fn: "__cvm_fpclassify", arg: FloatValue(bytecode.TypeF64, minNormalFloat64), want: 2},
		{name: "double zero", fn: "__cvm_fpclassify", arg: FloatValue(bytecode.TypeF64, 0), want: 4},
		{name: "double nan", fn: "__cvm_isnan", arg: FloatValue(bytecode.TypeF64, math.NaN()), want: 1},
		{name: "double inf", fn: "__cvm_isinf", arg: FloatValue(bytecode.TypeF64, math.Inf(1)), want: 1},
		{name: "negative zero sign", fn: "__cvm_signbit", arg: FloatValue(bytecode.TypeF64, math.Copysign(0, -1)), want: 1},
		{name: "long double normal", fn: "__cvm_isnormall", arg: FloatValue(bytecode.TypeFLong, 1), want: 1},
		{name: "tgmath ilogb", fn: "__cvm_tgmath_ilogb", arg: FloatValue(bytecode.TypeF64, 8), want: 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn, ok := reg.Lookup(tt.fn)
			if !ok {
				t.Fatalf("missing %s extern", tt.fn)
			}
			ret, exit, err := fn(context.Background(), nil, []Value{tt.arg})
			if err != nil || exit != nil {
				t.Fatalf("%s ret=%#v exit=%#v err=%v", tt.fn, ret, exit, err)
			}
			if ret.Type != bytecode.TypeI32 || int64(int32(ret.Int)) != tt.want {
				t.Fatalf("%s ret = %#v, want i32 %d", tt.fn, ret, tt.want)
			}
		})
	}
}

func TestTgmathLongExterns(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	tests := []struct {
		name string
		fn   string
		arg  Value
		want int64
	}{
		{name: "lrint", fn: "__cvm_tgmath_lrint", arg: FloatValue(bytecode.TypeF64, 3), want: 3},
		{name: "lround", fn: "__cvm_tgmath_lround", arg: FloatValue(bytecode.TypeF64, 4), want: 4},
		{name: "llrint", fn: "__cvm_tgmath_llrint", arg: FloatValue(bytecode.TypeF64, 5), want: 5},
		{name: "llround", fn: "__cvm_tgmath_llround", arg: FloatValue(bytecode.TypeF64, 6), want: 6},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn, ok := reg.Lookup(tt.fn)
			if !ok {
				t.Fatalf("missing %s extern", tt.fn)
			}
			ret, exit, err := fn(context.Background(), nil, []Value{tt.arg})
			if err != nil || exit != nil {
				t.Fatalf("%s ret=%#v exit=%#v err=%v", tt.fn, ret, exit, err)
			}
			if ret.Type != bytecode.TypeI64 || int64(ret.Int) != tt.want {
				t.Fatalf("%s ret = %#v, want i64 %d", tt.fn, ret, tt.want)
			}
		})
	}
}

func TestTgmathFrexpExtern(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	fn, ok := reg.Lookup("__cvm_tgmath_frexp")
	if !ok {
		t.Fatal("missing __cvm_tgmath_frexp extern")
	}
	mem := NewMemory(bytecode.DefaultTarget())
	expAddr := mustAlloc(t, mem, "exp", 4, 4, false, blockLocal)
	ret, exit, err := fn(context.Background(), &ExternContext{Memory: mem}, []Value{
		FloatValue(bytecode.TypeF64, 8),
		PtrValue(expAddr),
	})
	if err != nil || exit != nil {
		t.Fatalf("__cvm_tgmath_frexp ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeF64 || ret.Float != 0.5 {
		t.Fatalf("__cvm_tgmath_frexp ret = %#v, want f64 0.5", ret)
	}
	exp, err := mem.Load(expAddr, bytecode.TypeI32, 4)
	if err != nil {
		t.Fatalf("Load exponent: %v", err)
	}
	if int64(exp.Int) != 4 {
		t.Fatalf("frexp exponent = %#v, want 4", exp)
	}
}

func TestTgmathRemquoExtern(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	fn, ok := reg.Lookup("__cvm_tgmath_remquo")
	if !ok {
		t.Fatal("missing __cvm_tgmath_remquo extern")
	}
	mem := NewMemory(bytecode.DefaultTarget())
	quoAddr := mustAlloc(t, mem, "quo", 4, 4, false, blockLocal)
	ret, exit, err := fn(context.Background(), &ExternContext{Memory: mem}, []Value{
		FloatValue(bytecode.TypeF64, 4),
		FloatValue(bytecode.TypeF64, 2),
		PtrValue(quoAddr),
	})
	if err != nil || exit != nil {
		t.Fatalf("__cvm_tgmath_remquo ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeF64 || ret.Float != 0 {
		t.Fatalf("__cvm_tgmath_remquo ret = %#v, want f64 0", ret)
	}
	quo, err := mem.Load(quoAddr, bytecode.TypeI32, 4)
	if err != nil {
		t.Fatalf("Load quotient: %v", err)
	}
	if int64(quo.Int) != 2 {
		t.Fatalf("remquo quotient = %#v, want 2", quo)
	}
}

func TestMathUnorderedExtern(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	fn, ok := reg.Lookup("__cvm_isunordered")
	if !ok {
		t.Fatal("missing __cvm_isunordered extern")
	}
	ret, exit, err := fn(context.Background(), nil, []Value{
		FloatValue(bytecode.TypeF64, math.NaN()),
		FloatValue(bytecode.TypeF64, 1),
	})
	if err != nil || exit != nil {
		t.Fatalf("__cvm_isunordered ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeI32 || ret.Int != 1 {
		t.Fatalf("__cvm_isunordered ret = %#v, want i32 1", ret)
	}
}

func TestTgmathFloatExterns(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	tests := []struct {
		name string
		fn   string
		args []Value
		want float64
	}{
		{name: "sin", fn: "__cvm_tgmath_sin", args: []Value{FloatValue(bytecode.TypeF64, 0)}, want: 0},
		{name: "exp", fn: "__cvm_tgmath_exp", args: []Value{FloatValue(bytecode.TypeF64, 0)}, want: 1},
		{name: "sqrt", fn: "__cvm_tgmath_sqrt", args: []Value{FloatValue(bytecode.TypeF64, 9)}, want: 3},
		{name: "fabs", fn: "__cvm_tgmath_fabs", args: []Value{FloatValue(bytecode.TypeF64, -3)}, want: 3},
		{name: "cos", fn: "__cvm_tgmath_cos", args: []Value{FloatValue(bytecode.TypeF64, 0)}, want: 1},
		{name: "tan", fn: "__cvm_tgmath_tan", args: []Value{FloatValue(bytecode.TypeF64, 0)}, want: 0},
		{name: "log", fn: "__cvm_tgmath_log", args: []Value{FloatValue(bytecode.TypeF64, 1)}, want: 0},
		{name: "sinh", fn: "__cvm_tgmath_sinh", args: []Value{FloatValue(bytecode.TypeF64, 0)}, want: 0},
		{name: "cosh", fn: "__cvm_tgmath_cosh", args: []Value{FloatValue(bytecode.TypeF64, 0)}, want: 1},
		{name: "tanh", fn: "__cvm_tgmath_tanh", args: []Value{FloatValue(bytecode.TypeF64, 0)}, want: 0},
		{name: "asin", fn: "__cvm_tgmath_asin", args: []Value{FloatValue(bytecode.TypeF64, 0)}, want: 0},
		{name: "acos", fn: "__cvm_tgmath_acos", args: []Value{FloatValue(bytecode.TypeF64, 1)}, want: 0},
		{name: "atan", fn: "__cvm_tgmath_atan", args: []Value{FloatValue(bytecode.TypeF64, 0)}, want: 0},
		{name: "asinh", fn: "__cvm_tgmath_asinh", args: []Value{FloatValue(bytecode.TypeF64, 0)}, want: 0},
		{name: "acosh", fn: "__cvm_tgmath_acosh", args: []Value{FloatValue(bytecode.TypeF64, 1)}, want: 0},
		{name: "atanh", fn: "__cvm_tgmath_atanh", args: []Value{FloatValue(bytecode.TypeF64, 0)}, want: 0},
		{name: "cbrt", fn: "__cvm_tgmath_cbrt", args: []Value{FloatValue(bytecode.TypeF64, 8)}, want: 2},
		{name: "ceil", fn: "__cvm_tgmath_ceil", args: []Value{FloatValue(bytecode.TypeF64, 1.25)}, want: 2},
		{name: "floor", fn: "__cvm_tgmath_floor", args: []Value{FloatValue(bytecode.TypeF64, 1.75)}, want: 1},
		{name: "trunc", fn: "__cvm_tgmath_trunc", args: []Value{FloatValue(bytecode.TypeF64, -1.75)}, want: -1},
		{name: "round", fn: "__cvm_tgmath_round", args: []Value{FloatValue(bytecode.TypeF64, -1.5)}, want: -2},
		{name: "exp2", fn: "__cvm_tgmath_exp2", args: []Value{FloatValue(bytecode.TypeF64, 3)}, want: 8},
		{name: "expm1", fn: "__cvm_tgmath_expm1", args: []Value{FloatValue(bytecode.TypeF64, 0)}, want: 0},
		{name: "log10", fn: "__cvm_tgmath_log10", args: []Value{FloatValue(bytecode.TypeF64, 1)}, want: 0},
		{name: "log1p", fn: "__cvm_tgmath_log1p", args: []Value{FloatValue(bytecode.TypeF64, 0)}, want: 0},
		{name: "log2", fn: "__cvm_tgmath_log2", args: []Value{FloatValue(bytecode.TypeF64, 8)}, want: 3},
		{name: "erf", fn: "__cvm_tgmath_erf", args: []Value{FloatValue(bytecode.TypeF64, 0)}, want: 0},
		{name: "erfc", fn: "__cvm_tgmath_erfc", args: []Value{FloatValue(bytecode.TypeF64, 0)}, want: 1},
		{name: "tgamma", fn: "__cvm_tgmath_tgamma", args: []Value{FloatValue(bytecode.TypeF64, 1)}, want: 1},
		{name: "lgamma", fn: "__cvm_tgmath_lgamma", args: []Value{FloatValue(bytecode.TypeF64, 1)}, want: 0},
		{name: "nearbyint", fn: "__cvm_tgmath_nearbyint", args: []Value{FloatValue(bytecode.TypeF64, 2)}, want: 2},
		{name: "rint", fn: "__cvm_tgmath_rint", args: []Value{FloatValue(bytecode.TypeF64, 3)}, want: 3},
		{name: "logb", fn: "__cvm_tgmath_logb", args: []Value{FloatValue(bytecode.TypeF64, 8)}, want: 3},
		{name: "atan2", fn: "__cvm_tgmath_atan2", args: []Value{FloatValue(bytecode.TypeF64, 0), FloatValue(bytecode.TypeF64, 1)}, want: 0},
		{name: "hypot", fn: "__cvm_tgmath_hypot", args: []Value{FloatValue(bytecode.TypeF64, 3), FloatValue(bytecode.TypeF64, 4)}, want: 5},
		{name: "fdim", fn: "__cvm_tgmath_fdim", args: []Value{FloatValue(bytecode.TypeF64, 5), FloatValue(bytecode.TypeF64, 3)}, want: 2},
		{name: "fmax", fn: "__cvm_tgmath_fmax", args: []Value{FloatValue(bytecode.TypeF64, 2), FloatValue(bytecode.TypeF64, 3)}, want: 3},
		{name: "fmin", fn: "__cvm_tgmath_fmin", args: []Value{FloatValue(bytecode.TypeF64, 2), FloatValue(bytecode.TypeF64, 3)}, want: 2},
		{name: "fmod", fn: "__cvm_tgmath_fmod", args: []Value{FloatValue(bytecode.TypeF64, 5), FloatValue(bytecode.TypeF64, 2)}, want: 1},
		{name: "remainder", fn: "__cvm_tgmath_remainder", args: []Value{FloatValue(bytecode.TypeF64, 4), FloatValue(bytecode.TypeF64, 2)}, want: 0},
		{name: "copysign", fn: "__cvm_tgmath_copysign", args: []Value{FloatValue(bytecode.TypeF64, 2), FloatValue(bytecode.TypeF64, -1)}, want: -2},
		{name: "nextafter", fn: "__cvm_tgmath_nextafter", args: []Value{FloatValue(bytecode.TypeF64, 1), FloatValue(bytecode.TypeF64, 1)}, want: 1},
		{name: "nexttoward", fn: "__cvm_tgmath_nexttoward", args: []Value{FloatValue(bytecode.TypeF64, 1), FloatValue(bytecode.TypeF64, 1)}, want: 1},
		{name: "scalbn", fn: "__cvm_tgmath_scalbn", args: []Value{FloatValue(bytecode.TypeF64, 2), IntValue(bytecode.TypeI32, 3)}, want: 16},
		{name: "scalbln", fn: "__cvm_tgmath_scalbln", args: []Value{FloatValue(bytecode.TypeF64, 3), IntValue(bytecode.TypeI64, 2)}, want: 12},
		{name: "ldexp", fn: "__cvm_tgmath_ldexp", args: []Value{FloatValue(bytecode.TypeF64, 2), IntValue(bytecode.TypeI32, 3)}, want: 16},
		{name: "fma", fn: "__cvm_tgmath_fma", args: []Value{FloatValue(bytecode.TypeF64, 2), FloatValue(bytecode.TypeF64, 3), FloatValue(bytecode.TypeF64, 4)}, want: 10},
		{name: "pow", fn: "__cvm_tgmath_pow", args: []Value{FloatValue(bytecode.TypeF64, 2), FloatValue(bytecode.TypeF64, 3)}, want: 8},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn, ok := reg.Lookup(tt.fn)
			if !ok {
				t.Fatalf("missing %s extern", tt.fn)
			}
			ret, exit, err := fn(context.Background(), nil, tt.args)
			if err != nil || exit != nil {
				t.Fatalf("%s ret=%#v exit=%#v err=%v", tt.fn, ret, exit, err)
			}
			if ret.Type != bytecode.TypeF64 || ret.Float != tt.want {
				t.Fatalf("%s ret = %#v, want f64 %v", tt.fn, ret, tt.want)
			}
		})
	}
}

func TestAbortReturnsTrap(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	for _, name := range []string{"abort", "__builtin_abort"} {
		t.Run(name, func(t *testing.T) {
			fn, ok := reg.Lookup(name)
			if !ok {
				t.Fatalf("missing %s extern", name)
			}
			_, _, err := fn(context.Background(), &ExternContext{Memory: NewMemory(bytecode.DefaultTarget())}, nil)
			if err == nil || !strings.Contains(err.Error(), "abort") {
				t.Fatalf("%s err = %v, want abort trap", name, err)
			}
		})
	}
}
