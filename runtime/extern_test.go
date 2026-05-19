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
	for _, name := range []string{"exit", "_Exit", "abort", "__builtin_abort", "__builtin_va_start", "__builtin_va_end", "remove", "rename", "fopen", "freopen", "tmpfile", "tmpnam", "fseek", "ftell", "rewind", "fgetpos", "fsetpos", "puts", "puts_unlocked", "putchar", "putchar_unlocked", "getchar", "getchar_unlocked", "fputc", "fputc_unlocked", "putc", "putc_unlocked", "fputs", "fputs_unlocked", "fgetc", "fgetc_unlocked", "getc", "getc_unlocked", "ungetc", "fgets", "fgets_unlocked", "fwide", "fputwc", "putwc", "putwchar", "fgetwc", "getwc", "getwchar", "ungetwc", "fputws", "fgetws", "fflush", "fflush_unlocked", "fclose", "fileno", "fileno_unlocked", "setbuf", "setvbuf", "flockfile", "ftrylockfile", "funlockfile", "ferror", "ferror_unlocked", "clearerr", "clearerr_unlocked", "feof", "feof_unlocked", "fwrite", "fwrite_unlocked", "fread", "fread_unlocked", "perror", "abs", "labs", "llabs", "div", "ldiv", "lldiv", "atoi", "atol", "atoll", "atof", "strtol", "strtoul", "strtoll", "strtoull", "strtod", "strtof", "strtold", "mblen", "mbtowc", "wctomb", "mbstowcs", "wcstombs", "mbrlen", "mbrtowc", "wcrtomb", "mbsrtowcs", "wcsrtombs", "rand", "srand", "getenv", "system", "atexit", "setlocale", "localeconv", "clock", "difftime", "time", "nan", "nanf", "nanl", "fabs", "fabsf", "fabsl", "sqrt", "sqrtf", "sqrtl", "sin", "sinf", "sinl", "cos", "cosf", "cosl", "tan", "tanf", "tanl", "sinh", "sinhf", "sinhl", "cosh", "coshf", "coshl", "tanh", "tanhf", "tanhl", "asin", "asinf", "asinl", "acos", "acosf", "acosl", "atan", "atanf", "atanl", "asinh", "asinhf", "asinhl", "acosh", "acoshf", "acoshl", "atanh", "atanhf", "atanhl", "cbrt", "cbrtf", "cbrtl", "erf", "erff", "erfl", "erfc", "erfcf", "erfcl", "tgamma", "tgammaf", "tgammal", "lgamma", "lgammaf", "lgammal", "exp", "expf", "expl", "exp2", "exp2f", "exp2l", "expm1", "expm1f", "expm1l", "log", "logf", "logl", "log10", "log10f", "log10l", "log1p", "log1pf", "log1pl", "log2", "log2f", "log2l", "ceil", "ceilf", "ceill", "floor", "floorf", "floorl", "trunc", "truncf", "truncl", "round", "roundf", "roundl", "nearbyint", "nearbyintf", "nearbyintl", "rint", "rintf", "rintl", "logb", "logbf", "logbl", "ilogb", "ilogbf", "ilogbl", "lrint", "lrintf", "lrintl", "lround", "lroundf", "lroundl", "llrint", "llrintf", "llrintl", "llround", "llroundf", "llroundl", "scalbn", "scalbnf", "scalbnl", "scalbln", "scalblnf", "scalblnl", "ldexp", "ldexpf", "ldexpl", "frexp", "frexpf", "frexpl", "modf", "modff", "modfl", "remquo", "remquof", "remquol", "pow", "powf", "powl", "atan2", "atan2f", "atan2l", "hypot", "hypotf", "hypotl", "fdim", "fdimf", "fdiml", "fmax", "fmaxf", "fmaxl", "fmin", "fminf", "fminl", "fmod", "fmodf", "fmodl", "remainder", "remainderf", "remainderl", "copysign", "copysignf", "copysignl", "fma", "fmaf", "fmal", "nextafter", "nextafterf", "nextafterl", "nexttoward", "nexttowardf", "nexttowardl", "cabs", "cabsf", "cabsl", "creal", "crealf", "creall", "cimag", "cimagf", "cimagl", "carg", "cargf", "cargl", "conj", "conjf", "conjl", "cproj", "cprojf", "cprojl", "csin", "csinf", "csinl", "ccos", "ccosf", "ccosl", "ctan", "ctanf", "ctanl", "csinh", "csinhf", "csinhl", "ccosh", "ccoshf", "ccoshl", "ctanh", "ctanhf", "ctanhl", "casin", "casinf", "casinl", "cacos", "cacosf", "cacosl", "catan", "catanf", "catanl", "casinh", "casinhf", "casinhl", "cacosh", "cacoshf", "cacoshl", "catanh", "catanhf", "catanhl", "cexp", "cexpf", "cexpl", "clog", "clogf", "clogl", "csqrt", "csqrtf", "csqrtl", "cpow", "cpowf", "cpowl", "isdigit", "isalpha", "isalnum", "isspace", "islower", "isupper", "isxdigit", "isprint", "iswdigit", "iswalpha", "iswalnum", "iswspace", "iswlower", "iswupper", "iswxdigit", "iswprint", "iswblank", "iswcntrl", "iswgraph", "iswpunct", "towlower", "towupper", "wctype", "iswctype", "wctrans", "towctrans", "isblank", "iscntrl", "isgraph", "ispunct", "tolower", "toupper", "strcmp", "memcmp", "bcmp", "strncmp", "strcoll", "memchr", "wcslen", "wcscmp", "wcsncmp", "wcscoll", "wcsxfrm", "wcstok", "wcschr", "wcsrchr", "wcsstr", "wcspbrk", "wcsspn", "wcscspn", "wcscpy", "wcsncpy", "wcscat", "wcsncat", "wmemchr", "wmemcmp", "wmemcpy", "wmemmove", "wmemset", "strrchr", "strpbrk", "strspn", "strcspn", "strtok", "strxfrm", "strnlen", "strerror", "__builtin_malloc", "malloc", "__builtin_calloc", "calloc", "realloc", "__builtin_strdup", "strdup", "strndup", "free", "__builtin_object_size", "__builtin_dynamic_object_size", "__builtin_memcpy", "memcpy", "__builtin_memmove", "memmove", "__builtin_mempcpy", "mempcpy", "memccpy", "bcopy", "__builtin_memset", "memset", "__builtin_bzero", "bzero", "__builtin___memcpy_chk", "__builtin___memmove_chk", "__builtin___mempcpy_chk", "__builtin___memset_chk", "__builtin_strlen", "strlen", "__builtin_strchr", "strchr", "__builtin_strstr", "strstr", "__builtin_strcpy", "strcpy", "__builtin_stpcpy", "stpcpy", "__builtin_strcat", "strcat", "__builtin_strncpy", "strncpy", "__builtin_stpncpy", "stpncpy", "__builtin_strncat", "strncat", "__builtin___strcpy_chk", "__builtin___stpcpy_chk", "__builtin___strcat_chk", "__builtin___strncpy_chk", "__builtin___stpncpy_chk", "__builtin___strncat_chk", "__builtin_sprintf", "__builtin_snprintf", "__builtin_vsprintf", "__builtin_vsnprintf", "sprintf", "snprintf", "vsprintf", "vsnprintf", "__builtin___sprintf_chk", "__builtin___snprintf_chk", "__builtin___vsprintf_chk", "__builtin___vsnprintf_chk", "__builtin_printf", "__builtin_printf_unlocked", "printf", "printf_unlocked", "__builtin_fprintf", "__builtin_fprintf_unlocked", "fprintf", "fprintf_unlocked", "__builtin_vprintf", "vprintf", "vprintf_unlocked", "__builtin_vfprintf", "vfprintf", "vfprintf_unlocked", "__builtin___printf_chk", "__builtin___fprintf_chk", "__builtin___vprintf_chk", "__builtin___vfprintf_chk", "scanf", "fscanf", "sscanf", "feclearexcept", "fetestexcept"} {
		if _, ok := reg.Lookup(name); !ok {
			t.Fatalf("missing extern %s", name)
		}
	}
	for _, name := range []string{"wprintf", "fwprintf", "swprintf", "vwprintf", "vfwprintf", "vswprintf"} {
		if _, ok := reg.Lookup(name); !ok {
			t.Fatalf("missing extern %s", name)
		}
	}
	for _, name := range []string{"wscanf", "fwscanf", "swscanf"} {
		if _, ok := reg.Lookup(name); !ok {
			t.Fatalf("missing extern %s", name)
		}
	}
}

func TestStdlibProcessTerminationExterns(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	exitFn, ok := reg.Lookup("exit")
	if !ok {
		t.Fatal("missing exit extern")
	}
	ret, exit, err := exitFn(context.Background(), nil, []Value{IntValue(bytecode.TypeI32, 23)})
	if err != nil || exit == nil {
		t.Fatalf("exit ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if exit.Code != 23 {
		t.Fatalf("exit code = %d, want 23", exit.Code)
	}

	immediateExitFn, ok := reg.Lookup("_Exit")
	if !ok {
		t.Fatal("missing _Exit extern")
	}
	ret, exit, err = immediateExitFn(context.Background(), nil, []Value{IntValue(bytecode.TypeI32, 42)})
	if err != nil || exit == nil {
		t.Fatalf("_Exit ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if exit.Code != 42 {
		t.Fatalf("_Exit code = %d, want 42", exit.Code)
	}

	abortFn, ok := reg.Lookup("abort")
	if !ok {
		t.Fatal("missing abort extern")
	}
	_, exit, err = abortFn(context.Background(), nil, nil)
	if exit != nil {
		t.Fatalf("abort exit=%#v, want nil exit", exit)
	}
	if err == nil || !strings.Contains(err.Error(), "abort") {
		t.Fatalf("abort err=%v, want abort trap", err)
	}
}

func TestStdioFileOperationStubs(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	oldPath := mustAllocBytes(t, mem, "stdio:old-path", []byte("old.txt\x00"), true, blockString)
	newPath := mustAllocBytes(t, mem, "stdio:new-path", []byte("new.txt\x00"), true, blockString)

	removeFn, ok := reg.Lookup("remove")
	if !ok {
		t.Fatal("missing remove extern")
	}
	ret, exit, err := removeFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(oldPath)})
	if err != nil || exit != nil {
		t.Fatalf("remove ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeI32 || signedInt(ret) != -1 {
		t.Fatalf("remove ret=%#v, want i32 -1", ret)
	}

	renameFn, ok := reg.Lookup("rename")
	if !ok {
		t.Fatal("missing rename extern")
	}
	ret, exit, err = renameFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(oldPath), PtrValue(newPath)})
	if err != nil || exit != nil {
		t.Fatalf("rename ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeI32 || signedInt(ret) != -1 {
		t.Fatalf("rename ret=%#v, want i32 -1", ret)
	}
}

func TestConfiguredFileRemoveRename(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	reg.AddFile("old.txt", []byte("ABC"))
	mem := NewMemory(bytecode.DefaultTarget())
	oldPath := mustAllocBytes(t, mem, "stdio:configured-old-path", []byte("old.txt\x00"), true, blockString)
	newPath := mustAllocBytes(t, mem, "stdio:configured-new-path", []byte("new.txt\x00"), true, blockString)
	mode := mustAllocBytes(t, mem, "stdio:configured-mode", []byte("r\x00"), true, blockString)

	renameFn, ok := reg.Lookup("rename")
	if !ok {
		t.Fatal("missing rename extern")
	}
	removeFn, ok := reg.Lookup("remove")
	if !ok {
		t.Fatal("missing remove extern")
	}
	fopenFn, ok := reg.Lookup("fopen")
	if !ok {
		t.Fatal("missing fopen extern")
	}

	ret, exit, err := renameFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(oldPath), PtrValue(newPath)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeI32 || ret.Int != 0 {
		t.Fatalf("rename ret=%#v exit=%#v err=%v, want 0", ret, exit, err)
	}
	ret, exit, err = fopenFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(oldPath), PtrValue(mode)})
	if err != nil || exit != nil || ret.Type != bytecode.TypePtr || ret.Int != 0 {
		t.Fatalf("fopen old ret=%#v exit=%#v err=%v, want null", ret, exit, err)
	}
	ret, exit, err = fopenFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(newPath), PtrValue(mode)})
	if err != nil || exit != nil || ret.Type != bytecode.TypePtr || ret.Int == 0 {
		t.Fatalf("fopen new ret=%#v exit=%#v err=%v, want handle", ret, exit, err)
	}
	ret, exit, err = removeFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(newPath)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeI32 || ret.Int != 0 {
		t.Fatalf("remove ret=%#v exit=%#v err=%v, want 0", ret, exit, err)
	}
	ret, exit, err = fopenFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(newPath), PtrValue(mode)})
	if err != nil || exit != nil || ret.Type != bytecode.TypePtr || ret.Int != 0 {
		t.Fatalf("fopen removed ret=%#v exit=%#v err=%v, want null", ret, exit, err)
	}
}

func TestStdioOpenMissingFileStubs(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	path := mustAllocBytes(t, mem, "stdio:path", []byte("missing.txt\x00"), true, blockString)
	mode := mustAllocBytes(t, mem, "stdio:mode", []byte("r\x00"), true, blockString)

	fopenFn, ok := reg.Lookup("fopen")
	if !ok {
		t.Fatal("missing fopen extern")
	}
	ret, exit, err := fopenFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(path), PtrValue(mode)})
	if err != nil || exit != nil {
		t.Fatalf("fopen ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypePtr || ret.Int != 0 {
		t.Fatalf("fopen ret=%#v, want null", ret)
	}

	freopenFn, ok := reg.Lookup("freopen")
	if !ok {
		t.Fatal("missing freopen extern")
	}
	stdout, ok := reg.LookupVariable("stdout", mem)
	if !ok {
		t.Fatal("missing stdout extern variable")
	}
	ret, exit, err = freopenFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(path), PtrValue(mode), PtrValue(stdout)})
	if err != nil || exit != nil {
		t.Fatalf("freopen ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypePtr || ret.Int != 0 {
		t.Fatalf("freopen ret=%#v, want null", ret)
	}
}

func TestFreopenConfiguredReadFile(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	reg.AddFile("old.txt", []byte("OLD"))
	reg.AddFile("new.txt", []byte("NEW"))
	mem := NewMemory(bytecode.DefaultTarget())
	oldPath := mustAllocBytes(t, mem, "stdio:freopen-old-path", []byte("old.txt\x00"), true, blockString)
	newPath := mustAllocBytes(t, mem, "stdio:freopen-new-path", []byte("new.txt\x00"), true, blockString)
	mode := mustAllocBytes(t, mem, "stdio:freopen-read-mode", []byte("r\x00"), true, blockString)

	fopenFn, ok := reg.Lookup("fopen")
	if !ok {
		t.Fatal("missing fopen extern")
	}
	freopenFn, ok := reg.Lookup("freopen")
	if !ok {
		t.Fatal("missing freopen extern")
	}
	fgetcFn, ok := reg.Lookup("fgetc")
	if !ok {
		t.Fatal("missing fgetc extern")
	}

	ret, exit, err := fopenFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(oldPath), PtrValue(mode)})
	if err != nil || exit != nil || ret.Type != bytecode.TypePtr || ret.Int == 0 {
		t.Fatalf("fopen ret=%#v exit=%#v err=%v, want file handle", ret, exit, err)
	}
	file := ret.Int
	ret, exit, err = fgetcFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(file)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeI32 || ret.Int != 'O' {
		t.Fatalf("fgetc old ret=%#v exit=%#v err=%v, want O", ret, exit, err)
	}
	ret, exit, err = freopenFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(newPath), PtrValue(mode), PtrValue(file)})
	if err != nil || exit != nil || ret.Type != bytecode.TypePtr || ret.Int != file {
		t.Fatalf("freopen ret=%#v exit=%#v err=%v, want same file handle", ret, exit, err)
	}
	ret, exit, err = fgetcFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(file)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeI32 || ret.Int != 'N' {
		t.Fatalf("fgetc reopened ret=%#v exit=%#v err=%v, want N", ret, exit, err)
	}
}

func TestFreopenConfiguredWriteFile(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	reg.AddFile("old.txt", []byte("OLD"))
	reg.AddFile("out.txt", []byte("stale"))
	mem := NewMemory(bytecode.DefaultTarget())
	oldPath := mustAllocBytes(t, mem, "stdio:freopen-write-old-path", []byte("old.txt\x00"), true, blockString)
	outPath := mustAllocBytes(t, mem, "stdio:freopen-write-out-path", []byte("out.txt\x00"), true, blockString)
	readMode := mustAllocBytes(t, mem, "stdio:freopen-write-read-mode", []byte("r\x00"), true, blockString)
	writeMode := mustAllocBytes(t, mem, "stdio:freopen-write-mode", []byte("w\x00"), true, blockString)
	data := mustAllocBytes(t, mem, "stdio:freopen-write-data", []byte("XY"), true, blockString)

	fopenFn, ok := reg.Lookup("fopen")
	if !ok {
		t.Fatal("missing fopen extern")
	}
	freopenFn, ok := reg.Lookup("freopen")
	if !ok {
		t.Fatal("missing freopen extern")
	}
	fwriteFn, ok := reg.Lookup("fwrite")
	if !ok {
		t.Fatal("missing fwrite extern")
	}

	ret, exit, err := fopenFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(oldPath), PtrValue(readMode)})
	if err != nil || exit != nil || ret.Type != bytecode.TypePtr || ret.Int == 0 {
		t.Fatalf("fopen ret=%#v exit=%#v err=%v, want file handle", ret, exit, err)
	}
	file := ret.Int
	ret, exit, err = freopenFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(outPath), PtrValue(writeMode), PtrValue(file)})
	if err != nil || exit != nil || ret.Type != bytecode.TypePtr || ret.Int != file {
		t.Fatalf("freopen ret=%#v exit=%#v err=%v, want same file handle", ret, exit, err)
	}
	if got := string(reg.files["out.txt"]); got != "" {
		t.Fatalf("freopen truncated out.txt to %q, want empty", got)
	}
	ret, exit, err = fwriteFn(context.Background(), &ExternContext{Memory: mem}, []Value{
		ObjectAddrValue(data),
		UIntValue(bytecode.TypeU64, 1),
		UIntValue(bytecode.TypeU64, 2),
		PtrValue(file),
	})
	if err != nil || exit != nil || ret.Type != bytecode.TypeU64 || ret.Int != 2 {
		t.Fatalf("fwrite ret=%#v exit=%#v err=%v, want 2", ret, exit, err)
	}
	if got := string(reg.files["out.txt"]); got != "XY" {
		t.Fatalf("written out.txt = %q, want XY", got)
	}
}

func TestFreopenConfiguredAppendFile(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	reg.AddFile("old.txt", []byte("OLD"))
	reg.AddFile("log.txt", []byte("AB"))
	reg.AddFile("plus.txt", []byte("AB"))
	mem := NewMemory(bytecode.DefaultTarget())
	oldPath := mustAllocBytes(t, mem, "stdio:freopen-append-old-path", []byte("old.txt\x00"), true, blockString)
	logPath := mustAllocBytes(t, mem, "stdio:freopen-append-log-path", []byte("log.txt\x00"), true, blockString)
	plusPath := mustAllocBytes(t, mem, "stdio:freopen-append-plus-path", []byte("plus.txt\x00"), true, blockString)
	readMode := mustAllocBytes(t, mem, "stdio:freopen-append-read-mode", []byte("r\x00"), true, blockString)
	appendMode := mustAllocBytes(t, mem, "stdio:freopen-append-mode", []byte("a\x00"), true, blockString)
	appendPlusMode := mustAllocBytes(t, mem, "stdio:freopen-append-plus-mode", []byte("a+\x00"), true, blockString)

	fopenFn, ok := reg.Lookup("fopen")
	if !ok {
		t.Fatal("missing fopen extern")
	}
	freopenFn, ok := reg.Lookup("freopen")
	if !ok {
		t.Fatal("missing freopen extern")
	}
	fseekFn, ok := reg.Lookup("fseek")
	if !ok {
		t.Fatal("missing fseek extern")
	}
	fputcFn, ok := reg.Lookup("fputc")
	if !ok {
		t.Fatal("missing fputc extern")
	}

	ret, exit, err := fopenFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(oldPath), PtrValue(readMode)})
	if err != nil || exit != nil || ret.Type != bytecode.TypePtr || ret.Int == 0 {
		t.Fatalf("fopen ret=%#v exit=%#v err=%v, want file handle", ret, exit, err)
	}
	file := ret.Int
	ret, exit, err = freopenFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(logPath), PtrValue(appendMode), PtrValue(file)})
	if err != nil || exit != nil || ret.Type != bytecode.TypePtr || ret.Int != file {
		t.Fatalf("freopen append ret=%#v exit=%#v err=%v, want same file handle", ret, exit, err)
	}
	ret, exit, err = fseekFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(file), IntValue(bytecode.TypeI64, 0), IntValue(bytecode.TypeI32, 0)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeI32 || ret.Int != 0 {
		t.Fatalf("fseek append ret=%#v exit=%#v err=%v, want 0", ret, exit, err)
	}
	ret, exit, err = fputcFn(context.Background(), &ExternContext{Memory: mem}, []Value{IntValue(bytecode.TypeI32, 'C'), PtrValue(file)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeI32 || ret.Int != 'C' {
		t.Fatalf("fputc append ret=%#v exit=%#v err=%v, want C", ret, exit, err)
	}
	if got := string(reg.files["log.txt"]); got != "ABC" {
		t.Fatalf("append log.txt = %q, want ABC", got)
	}

	ret, exit, err = freopenFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(plusPath), PtrValue(appendPlusMode), PtrValue(file)})
	if err != nil || exit != nil || ret.Type != bytecode.TypePtr || ret.Int != file {
		t.Fatalf("freopen append plus ret=%#v exit=%#v err=%v, want same file handle", ret, exit, err)
	}
	ret, exit, err = fseekFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(file), IntValue(bytecode.TypeI64, 0), IntValue(bytecode.TypeI32, 0)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeI32 || ret.Int != 0 {
		t.Fatalf("fseek append plus ret=%#v exit=%#v err=%v, want 0", ret, exit, err)
	}
	ret, exit, err = fputcFn(context.Background(), &ExternContext{Memory: mem}, []Value{IntValue(bytecode.TypeI32, 'Z'), PtrValue(file)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeI32 || ret.Int != 'Z' {
		t.Fatalf("fputc append plus ret=%#v exit=%#v err=%v, want Z", ret, exit, err)
	}
	if got := string(reg.files["plus.txt"]); got != "ABZ" {
		t.Fatalf("append plus.txt = %q, want ABZ", got)
	}
}

func TestFreopenFailurePreservesStream(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	reg.AddFile("data.txt", []byte("ABC"))
	mem := NewMemory(bytecode.DefaultTarget())
	path := mustAllocBytes(t, mem, "stdio:freopen-fail-path", []byte("data.txt\x00"), true, blockString)
	missingPath := mustAllocBytes(t, mem, "stdio:freopen-fail-missing-path", []byte("missing.txt\x00"), true, blockString)
	readMode := mustAllocBytes(t, mem, "stdio:freopen-fail-read-mode", []byte("r\x00"), true, blockString)
	invalidMode := mustAllocBytes(t, mem, "stdio:freopen-fail-invalid-mode", []byte("x+\x00"), true, blockString)

	fopenFn, ok := reg.Lookup("fopen")
	if !ok {
		t.Fatal("missing fopen extern")
	}
	freopenFn, ok := reg.Lookup("freopen")
	if !ok {
		t.Fatal("missing freopen extern")
	}
	fgetcFn, ok := reg.Lookup("fgetc")
	if !ok {
		t.Fatal("missing fgetc extern")
	}

	ret, exit, err := fopenFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(path), PtrValue(readMode)})
	if err != nil || exit != nil || ret.Type != bytecode.TypePtr || ret.Int == 0 {
		t.Fatalf("fopen ret=%#v exit=%#v err=%v, want file handle", ret, exit, err)
	}
	file := ret.Int
	ret, exit, err = fgetcFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(file)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeI32 || ret.Int != 'A' {
		t.Fatalf("fgetc initial ret=%#v exit=%#v err=%v, want A", ret, exit, err)
	}
	ret, exit, err = freopenFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(missingPath), PtrValue(readMode), PtrValue(file)})
	if err != nil || exit != nil || ret.Type != bytecode.TypePtr || ret.Int != 0 {
		t.Fatalf("freopen missing ret=%#v exit=%#v err=%v, want null", ret, exit, err)
	}
	ret, exit, err = fgetcFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(file)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeI32 || ret.Int != 'B' {
		t.Fatalf("fgetc after missing ret=%#v exit=%#v err=%v, want B", ret, exit, err)
	}
	ret, exit, err = freopenFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(path), PtrValue(invalidMode), PtrValue(file)})
	if err != nil || exit != nil || ret.Type != bytecode.TypePtr || ret.Int != 0 {
		t.Fatalf("freopen invalid ret=%#v exit=%#v err=%v, want null", ret, exit, err)
	}
	ret, exit, err = fgetcFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(file)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeI32 || ret.Int != 'C' {
		t.Fatalf("fgetc after invalid ret=%#v exit=%#v err=%v, want C", ret, exit, err)
	}
}

func TestTmpfileReadWrite(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	data := mustAllocBytes(t, mem, "stdio:tmpfile-data", []byte("AB"), true, blockString)
	buf := mustAllocBytes(t, mem, "stdio:tmpfile-buf", []byte{0, 0}, false, blockLocal)
	tmpfileFn, ok := reg.Lookup("tmpfile")
	if !ok {
		t.Fatal("missing tmpfile extern")
	}
	fwriteFn, ok := reg.Lookup("fwrite")
	if !ok {
		t.Fatal("missing fwrite extern")
	}
	rewindFn, ok := reg.Lookup("rewind")
	if !ok {
		t.Fatal("missing rewind extern")
	}
	freadFn, ok := reg.Lookup("fread")
	if !ok {
		t.Fatal("missing fread extern")
	}

	ret, exit, err := tmpfileFn(context.Background(), &ExternContext{Memory: mem}, nil)
	if err != nil || exit != nil || ret.Type != bytecode.TypePtr || ret.Int == 0 {
		t.Fatalf("tmpfile ret=%#v exit=%#v err=%v, want handle", ret, exit, err)
	}
	file := ret.Int
	ret, exit, err = fwriteFn(context.Background(), &ExternContext{Memory: mem}, []Value{
		ObjectAddrValue(data),
		UIntValue(bytecode.TypeU64, 1),
		UIntValue(bytecode.TypeU64, 2),
		PtrValue(file),
	})
	if err != nil || exit != nil || ret.Type != bytecode.TypeU64 || ret.Int != 2 {
		t.Fatalf("fwrite ret=%#v exit=%#v err=%v, want 2", ret, exit, err)
	}
	if _, exit, err = rewindFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(file)}); err != nil || exit != nil {
		t.Fatalf("rewind exit=%#v err=%v", exit, err)
	}
	ret, exit, err = freadFn(context.Background(), &ExternContext{Memory: mem}, []Value{
		ObjectAddrValue(buf),
		UIntValue(bytecode.TypeU64, 1),
		UIntValue(bytecode.TypeU64, 2),
		PtrValue(file),
	})
	if err != nil || exit != nil || ret.Type != bytecode.TypeU64 || ret.Int != 2 {
		t.Fatalf("fread ret=%#v exit=%#v err=%v, want 2", ret, exit, err)
	}
	block, off, err := mem.rangeAccess(buf, 2, false)
	if err != nil {
		t.Fatalf("rangeAccess buf: %v", err)
	}
	if got := string(block.data[off : off+2]); got != "AB" {
		t.Fatalf("tmpfile bytes = %q, want AB", got)
	}
}

func TestFopenReadsConfiguredFile(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	reg.AddFile("data.txt", []byte("ABC"))
	mem := NewMemory(bytecode.DefaultTarget())
	path := mustAllocBytes(t, mem, "stdio:file-path", []byte("data.txt\x00"), true, blockString)
	mode := mustAllocBytes(t, mem, "stdio:file-mode", []byte("r\x00"), true, blockString)
	buf := mustAllocBytes(t, mem, "stdio:file-buf", []byte{0, 0}, false, blockLocal)

	fopenFn, ok := reg.Lookup("fopen")
	if !ok {
		t.Fatal("missing fopen extern")
	}
	fgetcFn, ok := reg.Lookup("fgetc")
	if !ok {
		t.Fatal("missing fgetc extern")
	}
	freadFn, ok := reg.Lookup("fread")
	if !ok {
		t.Fatal("missing fread extern")
	}

	ret, exit, err := fopenFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(path), PtrValue(mode)})
	if err != nil || exit != nil {
		t.Fatalf("fopen ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypePtr || ret.Int == 0 {
		t.Fatalf("fopen ret=%#v, want file handle", ret)
	}
	file := ret.Int

	ret, exit, err = fgetcFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(file)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeI32 || ret.Int != 'A' {
		t.Fatalf("fgetc ret=%#v exit=%#v err=%v, want 'A'", ret, exit, err)
	}
	ret, exit, err = freadFn(context.Background(), &ExternContext{Memory: mem}, []Value{
		ObjectAddrValue(buf),
		UIntValue(bytecode.TypeU64, 1),
		UIntValue(bytecode.TypeU64, 2),
		PtrValue(file),
	})
	if err != nil || exit != nil || ret.Type != bytecode.TypeU64 || ret.Int != 2 {
		t.Fatalf("fread ret=%#v exit=%#v err=%v, want 2", ret, exit, err)
	}
	block, off, err := mem.rangeAccess(buf, 2, false)
	if err != nil {
		t.Fatalf("rangeAccess buf: %v", err)
	}
	if got := string(block.data[off : off+2]); got != "BC" {
		t.Fatalf("fread bytes = %q, want BC", got)
	}
}

func TestConfiguredFilePositioning(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	reg.AddFile("data.txt", []byte("ABCDE"))
	mem := NewMemory(bytecode.DefaultTarget())
	path := mustAllocBytes(t, mem, "stdio:file-pos-path", []byte("data.txt\x00"), true, blockString)
	mode := mustAllocBytes(t, mem, "stdio:file-pos-mode", []byte("r\x00"), true, blockString)

	fopenFn, ok := reg.Lookup("fopen")
	if !ok {
		t.Fatal("missing fopen extern")
	}
	fseekFn, ok := reg.Lookup("fseek")
	if !ok {
		t.Fatal("missing fseek extern")
	}
	ftellFn, ok := reg.Lookup("ftell")
	if !ok {
		t.Fatal("missing ftell extern")
	}
	fgetposFn, ok := reg.Lookup("fgetpos")
	if !ok {
		t.Fatal("missing fgetpos extern")
	}
	fsetposFn, ok := reg.Lookup("fsetpos")
	if !ok {
		t.Fatal("missing fsetpos extern")
	}
	rewindFn, ok := reg.Lookup("rewind")
	if !ok {
		t.Fatal("missing rewind extern")
	}
	fgetcFn, ok := reg.Lookup("fgetc")
	if !ok {
		t.Fatal("missing fgetc extern")
	}

	ret, exit, err := fopenFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(path), PtrValue(mode)})
	if err != nil || exit != nil || ret.Type != bytecode.TypePtr || ret.Int == 0 {
		t.Fatalf("fopen ret=%#v exit=%#v err=%v, want file handle", ret, exit, err)
	}
	file := ret.Int

	ret, exit, err = fseekFn(context.Background(), &ExternContext{Memory: mem}, []Value{
		PtrValue(file),
		IntValue(bytecode.TypeI64, 2),
		IntValue(bytecode.TypeI32, 0),
	})
	if err != nil || exit != nil || ret.Type != bytecode.TypeI32 || ret.Int != 0 {
		t.Fatalf("fseek set ret=%#v exit=%#v err=%v, want 0", ret, exit, err)
	}
	ret, exit, err = ftellFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(file)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeI64 || ret.Int != 2 {
		t.Fatalf("ftell ret=%#v exit=%#v err=%v, want 2", ret, exit, err)
	}
	posAddr := mustAlloc(t, mem, "stdio:file-pos-fpos", 8, 8, false, blockLocal)
	ret, exit, err = fgetposFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(file), PtrValue(posAddr)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeI32 || ret.Int != 0 {
		t.Fatalf("fgetpos ret=%#v exit=%#v err=%v, want 0", ret, exit, err)
	}
	pos, err := mem.Load(posAddr, bytecode.TypeI64, 8)
	if err != nil || pos.Int != 2 {
		t.Fatalf("fgetpos stored=%#v err=%v, want 2", pos, err)
	}
	ret, exit, err = fgetcFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(file)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeI32 || ret.Int != 'C' {
		t.Fatalf("fgetc ret=%#v exit=%#v err=%v, want C", ret, exit, err)
	}
	ret, exit, err = fsetposFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(file), PtrValue(posAddr)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeI32 || ret.Int != 0 {
		t.Fatalf("fsetpos ret=%#v exit=%#v err=%v, want 0", ret, exit, err)
	}
	ret, exit, err = fgetcFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(file)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeI32 || ret.Int != 'C' {
		t.Fatalf("fgetc after fsetpos ret=%#v exit=%#v err=%v, want C", ret, exit, err)
	}
	ret, exit, err = fseekFn(context.Background(), &ExternContext{Memory: mem}, []Value{
		PtrValue(file),
		IntValue(bytecode.TypeI64, -1),
		IntValue(bytecode.TypeI32, 2),
	})
	if err != nil || exit != nil || ret.Type != bytecode.TypeI32 || ret.Int != 0 {
		t.Fatalf("fseek end ret=%#v exit=%#v err=%v, want 0", ret, exit, err)
	}
	ret, exit, err = fgetcFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(file)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeI32 || ret.Int != 'E' {
		t.Fatalf("fgetc after SEEK_END ret=%#v exit=%#v err=%v, want E", ret, exit, err)
	}
	ret, exit, err = rewindFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(file)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeVoid {
		t.Fatalf("rewind ret=%#v exit=%#v err=%v, want void", ret, exit, err)
	}
	ret, exit, err = fgetcFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(file)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeI32 || ret.Int != 'A' {
		t.Fatalf("fgetc after rewind ret=%#v exit=%#v err=%v, want A", ret, exit, err)
	}
}

func TestFopenWritesConfiguredFile(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	path := mustAllocBytes(t, mem, "stdio:file-write-path", []byte("out.txt\x00"), true, blockString)
	writeMode := mustAllocBytes(t, mem, "stdio:file-write-mode", []byte("w\x00"), true, blockString)
	readMode := mustAllocBytes(t, mem, "stdio:file-read-mode", []byte("r\x00"), true, blockString)
	data := mustAllocBytes(t, mem, "stdio:file-write-data", []byte("AB"), true, blockString)
	buf := mustAllocBytes(t, mem, "stdio:file-read-buf", []byte{0, 0, 0}, false, blockLocal)

	fopenFn, ok := reg.Lookup("fopen")
	if !ok {
		t.Fatal("missing fopen extern")
	}
	fwriteFn, ok := reg.Lookup("fwrite")
	if !ok {
		t.Fatal("missing fwrite extern")
	}
	fputcFn, ok := reg.Lookup("fputc")
	if !ok {
		t.Fatal("missing fputc extern")
	}
	fcloseFn, ok := reg.Lookup("fclose")
	if !ok {
		t.Fatal("missing fclose extern")
	}
	freadFn, ok := reg.Lookup("fread")
	if !ok {
		t.Fatal("missing fread extern")
	}

	ret, exit, err := fopenFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(path), PtrValue(writeMode)})
	if err != nil || exit != nil || ret.Type != bytecode.TypePtr || ret.Int == 0 {
		t.Fatalf("fopen write ret=%#v exit=%#v err=%v, want file handle", ret, exit, err)
	}
	file := ret.Int
	ret, exit, err = fwriteFn(context.Background(), &ExternContext{Memory: mem}, []Value{
		ObjectAddrValue(data),
		UIntValue(bytecode.TypeU64, 1),
		UIntValue(bytecode.TypeU64, 2),
		PtrValue(file),
	})
	if err != nil || exit != nil || ret.Type != bytecode.TypeU64 || ret.Int != 2 {
		t.Fatalf("fwrite ret=%#v exit=%#v err=%v, want 2", ret, exit, err)
	}
	ret, exit, err = fputcFn(context.Background(), &ExternContext{Memory: mem}, []Value{
		IntValue(bytecode.TypeI32, 'C'),
		PtrValue(file),
	})
	if err != nil || exit != nil || ret.Type != bytecode.TypeI32 || ret.Int != 'C' {
		t.Fatalf("fputc ret=%#v exit=%#v err=%v, want C", ret, exit, err)
	}
	if _, exit, err = fcloseFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(file)}); err != nil || exit != nil {
		t.Fatalf("fclose write exit=%#v err=%v", exit, err)
	}

	ret, exit, err = fopenFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(path), PtrValue(readMode)})
	if err != nil || exit != nil || ret.Type != bytecode.TypePtr || ret.Int == 0 {
		t.Fatalf("fopen read ret=%#v exit=%#v err=%v, want file handle", ret, exit, err)
	}
	ret, exit, err = freadFn(context.Background(), &ExternContext{Memory: mem}, []Value{
		ObjectAddrValue(buf),
		UIntValue(bytecode.TypeU64, 1),
		UIntValue(bytecode.TypeU64, 3),
		PtrValue(ret.Int),
	})
	if err != nil || exit != nil || ret.Type != bytecode.TypeU64 || ret.Int != 3 {
		t.Fatalf("fread ret=%#v exit=%#v err=%v, want 3", ret, exit, err)
	}
	block, off, err := mem.rangeAccess(buf, 3, false)
	if err != nil {
		t.Fatalf("rangeAccess buf: %v", err)
	}
	if got := string(block.data[off : off+3]); got != "ABC" {
		t.Fatalf("file bytes = %q, want ABC", got)
	}
}

func TestFopenAppendsConfiguredFile(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	reg.AddFile("log.txt", []byte("AB"))
	mem := NewMemory(bytecode.DefaultTarget())
	path := mustAllocBytes(t, mem, "stdio:file-append-path", []byte("log.txt\x00"), true, blockString)
	appendMode := mustAllocBytes(t, mem, "stdio:file-append-mode", []byte("a\x00"), true, blockString)
	readMode := mustAllocBytes(t, mem, "stdio:file-append-read-mode", []byte("r\x00"), true, blockString)
	buf := mustAllocBytes(t, mem, "stdio:file-append-buf", []byte{0, 0, 0}, false, blockLocal)

	fopenFn, ok := reg.Lookup("fopen")
	if !ok {
		t.Fatal("missing fopen extern")
	}
	fputcFn, ok := reg.Lookup("fputc")
	if !ok {
		t.Fatal("missing fputc extern")
	}
	fcloseFn, ok := reg.Lookup("fclose")
	if !ok {
		t.Fatal("missing fclose extern")
	}
	freadFn, ok := reg.Lookup("fread")
	if !ok {
		t.Fatal("missing fread extern")
	}

	ret, exit, err := fopenFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(path), PtrValue(appendMode)})
	if err != nil || exit != nil || ret.Type != bytecode.TypePtr || ret.Int == 0 {
		t.Fatalf("fopen append ret=%#v exit=%#v err=%v, want file handle", ret, exit, err)
	}
	file := ret.Int
	ret, exit, err = fputcFn(context.Background(), &ExternContext{Memory: mem}, []Value{
		IntValue(bytecode.TypeI32, 'C'),
		PtrValue(file),
	})
	if err != nil || exit != nil || ret.Type != bytecode.TypeI32 || ret.Int != 'C' {
		t.Fatalf("fputc ret=%#v exit=%#v err=%v, want C", ret, exit, err)
	}
	if _, exit, err = fcloseFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(file)}); err != nil || exit != nil {
		t.Fatalf("fclose append exit=%#v err=%v", exit, err)
	}
	ret, exit, err = fopenFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(path), PtrValue(readMode)})
	if err != nil || exit != nil || ret.Type != bytecode.TypePtr || ret.Int == 0 {
		t.Fatalf("fopen read ret=%#v exit=%#v err=%v, want file handle", ret, exit, err)
	}
	ret, exit, err = freadFn(context.Background(), &ExternContext{Memory: mem}, []Value{
		ObjectAddrValue(buf),
		UIntValue(bytecode.TypeU64, 1),
		UIntValue(bytecode.TypeU64, 3),
		PtrValue(ret.Int),
	})
	if err != nil || exit != nil || ret.Type != bytecode.TypeU64 || ret.Int != 3 {
		t.Fatalf("fread ret=%#v exit=%#v err=%v, want 3", ret, exit, err)
	}
	block, off, err := mem.rangeAccess(buf, 3, false)
	if err != nil {
		t.Fatalf("rangeAccess buf: %v", err)
	}
	if got := string(block.data[off : off+3]); got != "ABC" {
		t.Fatalf("append bytes = %q, want ABC", got)
	}
}

func TestFopenAppendModeWritesAtEndAfterSeek(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	reg.AddFile("log.txt", []byte("AB"))
	mem := NewMemory(bytecode.DefaultTarget())
	path := mustAllocBytes(t, mem, "stdio:file-append-seek-path", []byte("log.txt\x00"), true, blockString)
	appendMode := mustAllocBytes(t, mem, "stdio:file-append-seek-mode", []byte("a+\x00"), true, blockString)
	readMode := mustAllocBytes(t, mem, "stdio:file-append-seek-read-mode", []byte("r\x00"), true, blockString)
	buf := mustAllocBytes(t, mem, "stdio:file-append-seek-buf", []byte{0, 0, 0}, false, blockLocal)

	fopenFn, ok := reg.Lookup("fopen")
	if !ok {
		t.Fatal("missing fopen extern")
	}
	fseekFn, ok := reg.Lookup("fseek")
	if !ok {
		t.Fatal("missing fseek extern")
	}
	fputcFn, ok := reg.Lookup("fputc")
	if !ok {
		t.Fatal("missing fputc extern")
	}
	fcloseFn, ok := reg.Lookup("fclose")
	if !ok {
		t.Fatal("missing fclose extern")
	}
	freadFn, ok := reg.Lookup("fread")
	if !ok {
		t.Fatal("missing fread extern")
	}

	ret, exit, err := fopenFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(path), PtrValue(appendMode)})
	if err != nil || exit != nil || ret.Type != bytecode.TypePtr || ret.Int == 0 {
		t.Fatalf("fopen append ret=%#v exit=%#v err=%v, want file handle", ret, exit, err)
	}
	file := ret.Int
	ret, exit, err = fseekFn(context.Background(), &ExternContext{Memory: mem}, []Value{
		PtrValue(file),
		IntValue(bytecode.TypeI64, 0),
		IntValue(bytecode.TypeI32, 0),
	})
	if err != nil || exit != nil || ret.Type != bytecode.TypeI32 || ret.Int != 0 {
		t.Fatalf("fseek ret=%#v exit=%#v err=%v, want 0", ret, exit, err)
	}
	ret, exit, err = fputcFn(context.Background(), &ExternContext{Memory: mem}, []Value{IntValue(bytecode.TypeI32, 'Z'), PtrValue(file)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeI32 || ret.Int != 'Z' {
		t.Fatalf("fputc ret=%#v exit=%#v err=%v, want Z", ret, exit, err)
	}
	if _, exit, err = fcloseFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(file)}); err != nil || exit != nil {
		t.Fatalf("fclose append exit=%#v err=%v", exit, err)
	}

	ret, exit, err = fopenFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(path), PtrValue(readMode)})
	if err != nil || exit != nil || ret.Type != bytecode.TypePtr || ret.Int == 0 {
		t.Fatalf("fopen read ret=%#v exit=%#v err=%v, want file handle", ret, exit, err)
	}
	ret, exit, err = freadFn(context.Background(), &ExternContext{Memory: mem}, []Value{
		ObjectAddrValue(buf),
		UIntValue(bytecode.TypeU64, 1),
		UIntValue(bytecode.TypeU64, 3),
		PtrValue(ret.Int),
	})
	if err != nil || exit != nil || ret.Type != bytecode.TypeU64 || ret.Int != 3 {
		t.Fatalf("fread ret=%#v exit=%#v err=%v, want 3", ret, exit, err)
	}
	block, off, err := mem.rangeAccess(buf, 3, false)
	if err != nil {
		t.Fatalf("rangeAccess buf: %v", err)
	}
	if got := string(block.data[off : off+3]); got != "ABZ" {
		t.Fatalf("append after seek bytes = %q, want ABZ", got)
	}
}

func TestFopenUpdateModeRequiresDirectionBarriers(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	reg.AddFile("data.txt", []byte("AB"))
	reg.AddFile("eof.txt", []byte("A"))
	mem := NewMemory(bytecode.DefaultTarget())
	dataPath := mustAllocBytes(t, mem, "stdio:update-data-path", []byte("data.txt\x00"), true, blockString)
	eofPath := mustAllocBytes(t, mem, "stdio:update-eof-path", []byte("eof.txt\x00"), true, blockString)
	writePath := mustAllocBytes(t, mem, "stdio:update-write-path", []byte("write.txt\x00"), true, blockString)
	readWriteMode := mustAllocBytes(t, mem, "stdio:update-rplus-mode", []byte("r+\x00"), true, blockString)
	writeReadMode := mustAllocBytes(t, mem, "stdio:update-wplus-mode", []byte("w+\x00"), true, blockString)
	readMode := mustAllocBytes(t, mem, "stdio:update-read-mode", []byte("r\x00"), true, blockString)
	buf := mustAllocBytes(t, mem, "stdio:update-buf", []byte{0, 0}, false, blockLocal)
	z := mustAllocBytes(t, mem, "stdio:update-z", []byte("Z"), true, blockString)

	fopenFn, ok := reg.Lookup("fopen")
	if !ok {
		t.Fatal("missing fopen extern")
	}
	freadFn, ok := reg.Lookup("fread")
	if !ok {
		t.Fatal("missing fread extern")
	}
	fwriteFn, ok := reg.Lookup("fwrite")
	if !ok {
		t.Fatal("missing fwrite extern")
	}
	ferrorFn, ok := reg.Lookup("ferror")
	if !ok {
		t.Fatal("missing ferror extern")
	}
	fcloseFn, ok := reg.Lookup("fclose")
	if !ok {
		t.Fatal("missing fclose extern")
	}

	ret, exit, err := fopenFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(dataPath), PtrValue(readWriteMode)})
	if err != nil || exit != nil || ret.Type != bytecode.TypePtr || ret.Int == 0 {
		t.Fatalf("fopen r+ ret=%#v exit=%#v err=%v, want handle", ret, exit, err)
	}
	dataFile := ret.Int
	ret, exit, err = freadFn(context.Background(), &ExternContext{Memory: mem}, []Value{ObjectAddrValue(buf), UIntValue(bytecode.TypeU64, 1), UIntValue(bytecode.TypeU64, 1), PtrValue(dataFile)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeU64 || ret.Int != 1 {
		t.Fatalf("initial fread ret=%#v exit=%#v err=%v, want 1", ret, exit, err)
	}
	ret, exit, err = fwriteFn(context.Background(), &ExternContext{Memory: mem}, []Value{ObjectAddrValue(z), UIntValue(bytecode.TypeU64, 1), UIntValue(bytecode.TypeU64, 1), PtrValue(dataFile)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeU64 || ret.Int != 0 {
		t.Fatalf("write after read ret=%#v exit=%#v err=%v, want 0", ret, exit, err)
	}
	ret, exit, err = ferrorFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(dataFile)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeI32 || ret.Int == 0 {
		t.Fatalf("ferror after invalid write ret=%#v exit=%#v err=%v, want nonzero", ret, exit, err)
	}
	if _, exit, err = fcloseFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(dataFile)}); err != nil || exit != nil {
		t.Fatalf("fclose data exit=%#v err=%v", exit, err)
	}

	ret, exit, err = fopenFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(writePath), PtrValue(writeReadMode)})
	if err != nil || exit != nil || ret.Type != bytecode.TypePtr || ret.Int == 0 {
		t.Fatalf("fopen w+ ret=%#v exit=%#v err=%v, want handle", ret, exit, err)
	}
	writeFile := ret.Int
	ret, exit, err = fwriteFn(context.Background(), &ExternContext{Memory: mem}, []Value{ObjectAddrValue(z), UIntValue(bytecode.TypeU64, 1), UIntValue(bytecode.TypeU64, 1), PtrValue(writeFile)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeU64 || ret.Int != 1 {
		t.Fatalf("initial fwrite ret=%#v exit=%#v err=%v, want 1", ret, exit, err)
	}
	ret, exit, err = freadFn(context.Background(), &ExternContext{Memory: mem}, []Value{ObjectAddrValue(buf), UIntValue(bytecode.TypeU64, 1), UIntValue(bytecode.TypeU64, 1), PtrValue(writeFile)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeU64 || ret.Int != 0 {
		t.Fatalf("read after write ret=%#v exit=%#v err=%v, want 0", ret, exit, err)
	}
	ret, exit, err = ferrorFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(writeFile)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeI32 || ret.Int == 0 {
		t.Fatalf("ferror after invalid read ret=%#v exit=%#v err=%v, want nonzero", ret, exit, err)
	}
	if _, exit, err = fcloseFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(writeFile)}); err != nil || exit != nil {
		t.Fatalf("fclose write exit=%#v err=%v", exit, err)
	}

	ret, exit, err = fopenFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(eofPath), PtrValue(readWriteMode)})
	if err != nil || exit != nil || ret.Type != bytecode.TypePtr || ret.Int == 0 {
		t.Fatalf("fopen eof r+ ret=%#v exit=%#v err=%v, want handle", ret, exit, err)
	}
	eofFile := ret.Int
	ret, exit, err = freadFn(context.Background(), &ExternContext{Memory: mem}, []Value{ObjectAddrValue(buf), UIntValue(bytecode.TypeU64, 1), UIntValue(bytecode.TypeU64, 2), PtrValue(eofFile)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeU64 || ret.Int != 1 {
		t.Fatalf("eof fread ret=%#v exit=%#v err=%v, want 1", ret, exit, err)
	}
	ret, exit, err = fwriteFn(context.Background(), &ExternContext{Memory: mem}, []Value{ObjectAddrValue(z), UIntValue(bytecode.TypeU64, 1), UIntValue(bytecode.TypeU64, 1), PtrValue(eofFile)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeU64 || ret.Int != 1 {
		t.Fatalf("write after read EOF ret=%#v exit=%#v err=%v, want 1", ret, exit, err)
	}
	if _, exit, err = fcloseFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(eofFile)}); err != nil || exit != nil {
		t.Fatalf("fclose eof exit=%#v err=%v", exit, err)
	}
	ret, exit, err = fopenFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(eofPath), PtrValue(readMode)})
	if err != nil || exit != nil || ret.Type != bytecode.TypePtr || ret.Int == 0 {
		t.Fatalf("reopen eof ret=%#v exit=%#v err=%v, want handle", ret, exit, err)
	}
	ret, exit, err = freadFn(context.Background(), &ExternContext{Memory: mem}, []Value{ObjectAddrValue(buf), UIntValue(bytecode.TypeU64, 1), UIntValue(bytecode.TypeU64, 2), PtrValue(ret.Int)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeU64 || ret.Int != 2 {
		t.Fatalf("read eof file ret=%#v exit=%#v err=%v, want 2", ret, exit, err)
	}
	block, off, err := mem.rangeAccess(buf, 2, false)
	if err != nil {
		t.Fatalf("rangeAccess buf: %v", err)
	}
	if got := string(block.data[off : off+2]); got != "AZ" {
		t.Fatalf("eof file bytes = %q, want AZ", got)
	}
}

func TestTmpnamNullReturnsStaticName(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	fn, ok := reg.Lookup("tmpnam")
	if !ok {
		t.Fatal("missing tmpnam extern")
	}
	ret, exit, err := fn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(0)})
	if err != nil || exit != nil {
		t.Fatalf("tmpnam ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypePtr || ret.Int == 0 {
		t.Fatalf("tmpnam ret=%#v, want non-null pointer", ret)
	}
	got, err := mem.ReadCString(ret.Int)
	if err != nil {
		t.Fatalf("tmpnam string: %v", err)
	}
	if got != "/tmp/cvm-tmp-0" {
		t.Fatalf("tmpnam string = %q, want /tmp/cvm-tmp-0", got)
	}
	ret2, exit, err := fn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(0)})
	if err != nil || exit != nil {
		t.Fatalf("tmpnam second ret=%#v exit=%#v err=%v", ret2, exit, err)
	}
	if ret2.Type != bytecode.TypePtr || ret2.Int != ret.Int {
		t.Fatalf("tmpnam second ret=%#v, want same static pointer %#x", ret2, ret.Int)
	}
}

func TestTmpnamWritesCallerBuffer(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	buf := mustAlloc(t, mem, "stdio:tmpnam-buffer", 20, 1, false, blockLocal)
	fn, ok := reg.Lookup("tmpnam")
	if !ok {
		t.Fatal("missing tmpnam extern")
	}

	ret, exit, err := fn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(buf)})
	if err != nil || exit != nil {
		t.Fatalf("tmpnam ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypePtr || ret.Int != buf {
		t.Fatalf("tmpnam ret=%#v, want buffer pointer %#x", ret, buf)
	}
	got, err := mem.ReadCString(buf)
	if err != nil {
		t.Fatalf("tmpnam buffer string: %v", err)
	}
	if got != "/tmp/cvm-tmp-0" {
		t.Fatalf("tmpnam buffer = %q, want /tmp/cvm-tmp-0", got)
	}
}

func TestTmpnamCallerBufferSequence(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	first := mustAlloc(t, mem, "stdio:tmpnam-sequence-first", 20, 1, false, blockLocal)
	second := mustAlloc(t, mem, "stdio:tmpnam-sequence-second", 20, 1, false, blockLocal)
	fn, ok := reg.Lookup("tmpnam")
	if !ok {
		t.Fatal("missing tmpnam extern")
	}

	ret, exit, err := fn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(first)})
	if err != nil || exit != nil || ret.Type != bytecode.TypePtr || ret.Int != first {
		t.Fatalf("first tmpnam ret=%#v exit=%#v err=%v, want first buffer", ret, exit, err)
	}
	ret, exit, err = fn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(second)})
	if err != nil || exit != nil || ret.Type != bytecode.TypePtr || ret.Int != second {
		t.Fatalf("second tmpnam ret=%#v exit=%#v err=%v, want second buffer", ret, exit, err)
	}
	firstName, err := mem.ReadCString(first)
	if err != nil {
		t.Fatalf("first tmpnam string: %v", err)
	}
	secondName, err := mem.ReadCString(second)
	if err != nil {
		t.Fatalf("second tmpnam string: %v", err)
	}
	if firstName != "/tmp/cvm-tmp-0" || secondName != "/tmp/cvm-tmp-1" {
		t.Fatalf("tmpnam names = %q, %q; want sequence 0, 1", firstName, secondName)
	}
}

func TestStdioPositionStubs(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	stdout, ok := reg.LookupVariable("stdout", mem)
	if !ok {
		t.Fatal("missing stdout extern variable")
	}
	stdin, ok := reg.LookupVariable("stdin", mem)
	if !ok {
		t.Fatal("missing stdin extern variable")
	}

	fseekFn, ok := reg.Lookup("fseek")
	if !ok {
		t.Fatal("missing fseek extern")
	}
	ret, exit, err := fseekFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(stdout), IntValue(bytecode.TypeI64, 0), IntValue(bytecode.TypeI32, 0)})
	if err != nil || exit != nil {
		t.Fatalf("fseek ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeI32 || signedInt(ret) != -1 {
		t.Fatalf("fseek ret=%#v, want i32 -1", ret)
	}

	ftellFn, ok := reg.Lookup("ftell")
	if !ok {
		t.Fatal("missing ftell extern")
	}
	ret, exit, err = ftellFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(stdout)})
	if err != nil || exit != nil {
		t.Fatalf("ftell ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeI64 || signedInt(ret) != -1 {
		t.Fatalf("ftell ret=%#v, want i64 -1", ret)
	}

	rewindFn, ok := reg.Lookup("rewind")
	if !ok {
		t.Fatal("missing rewind extern")
	}
	ret, exit, err = rewindFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(stdin)})
	if err != nil || exit != nil {
		t.Fatalf("rewind ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeVoid {
		t.Fatalf("rewind ret=%#v, want void", ret)
	}

	pos := mustAllocBytes(t, mem, "stdio:fpos", make([]byte, 8), false, blockLocal)
	fgetposFn, ok := reg.Lookup("fgetpos")
	if !ok {
		t.Fatal("missing fgetpos extern")
	}
	ret, exit, err = fgetposFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(stdout), PtrValue(pos)})
	if err != nil || exit != nil {
		t.Fatalf("fgetpos ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeI32 || signedInt(ret) != -1 {
		t.Fatalf("fgetpos ret=%#v, want i32 -1", ret)
	}
	ret, exit, err = fgetposFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(stdout), PtrValue(0xdeadbeef)})
	if err != nil || exit != nil {
		t.Fatalf("fgetpos invalid stdout ptr ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeI32 || signedInt(ret) != -1 {
		t.Fatalf("fgetpos invalid stdout ptr ret=%#v, want i32 -1", ret)
	}

	fsetposFn, ok := reg.Lookup("fsetpos")
	if !ok {
		t.Fatal("missing fsetpos extern")
	}
	ret, exit, err = fsetposFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(stdout), PtrValue(pos)})
	if err != nil || exit != nil {
		t.Fatalf("fsetpos ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeI32 || signedInt(ret) != -1 {
		t.Fatalf("fsetpos ret=%#v, want i32 -1", ret)
	}
	ret, exit, err = fsetposFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(stdout), PtrValue(0xdeadbeef)})
	if err != nil || exit != nil {
		t.Fatalf("fsetpos invalid stdout ptr ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeI32 || signedInt(ret) != -1 {
		t.Fatalf("fsetpos invalid stdout ptr ret=%#v, want i32 -1", ret)
	}
}

func TestConfiguredFilePositionErrors(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	reg.AddFile("data.txt", []byte("ABCDE"))
	mem := NewMemory(bytecode.DefaultTarget())
	path := mustAllocBytes(t, mem, "stdio:file-pos-error-path", []byte("data.txt\x00"), true, blockString)
	mode := mustAllocBytes(t, mem, "stdio:file-pos-error-mode", []byte("r\x00"), true, blockString)
	pos := mustAlloc(t, mem, "stdio:file-pos-error-fpos", 8, 8, false, blockLocal)

	fopenFn, ok := reg.Lookup("fopen")
	if !ok {
		t.Fatal("missing fopen extern")
	}
	fseekFn, ok := reg.Lookup("fseek")
	if !ok {
		t.Fatal("missing fseek extern")
	}
	ftellFn, ok := reg.Lookup("ftell")
	if !ok {
		t.Fatal("missing ftell extern")
	}
	fgetposFn, ok := reg.Lookup("fgetpos")
	if !ok {
		t.Fatal("missing fgetpos extern")
	}
	fsetposFn, ok := reg.Lookup("fsetpos")
	if !ok {
		t.Fatal("missing fsetpos extern")
	}

	ret, exit, err := fopenFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(path), PtrValue(mode)})
	if err != nil || exit != nil || ret.Type != bytecode.TypePtr || ret.Int == 0 {
		t.Fatalf("fopen ret=%#v exit=%#v err=%v, want file handle", ret, exit, err)
	}
	file := ret.Int
	ret, exit, err = fseekFn(context.Background(), &ExternContext{Memory: mem}, []Value{
		PtrValue(file),
		IntValue(bytecode.TypeI64, 2),
		IntValue(bytecode.TypeI32, 0),
	})
	if err != nil || exit != nil || ret.Type != bytecode.TypeI32 || ret.Int != 0 {
		t.Fatalf("fseek ret=%#v exit=%#v err=%v, want 0", ret, exit, err)
	}

	ret, exit, err = fgetposFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(file), PtrValue(0xdeadbeef)})
	if err == nil || !strings.Contains(err.Error(), "invalid memory access") || exit != nil {
		t.Fatalf("fgetpos invalid ptr ret=%#v exit=%#v err=%v, want invalid memory access", ret, exit, err)
	}
	ret, exit, err = fsetposFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(file), PtrValue(0xdeadbeef)})
	if err == nil || !strings.Contains(err.Error(), "invalid memory access") || exit != nil {
		t.Fatalf("fsetpos invalid ptr ret=%#v exit=%#v err=%v, want invalid memory access", ret, exit, err)
	}

	if err := mem.Store(pos, bytecode.TypeI64, 8, IntValue(bytecode.TypeI64, -1)); err != nil {
		t.Fatalf("store negative fpos: %v", err)
	}
	ret, exit, err = fsetposFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(file), PtrValue(pos)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeI32 || signedInt(ret) != -1 {
		t.Fatalf("fsetpos negative ret=%#v exit=%#v err=%v, want -1", ret, exit, err)
	}
	ret, exit, err = ftellFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(file)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeI64 || ret.Int != 2 {
		t.Fatalf("ftell after negative fsetpos ret=%#v exit=%#v err=%v, want 2", ret, exit, err)
	}
}

func TestPlainMathUnaryExterns(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	tests := []struct {
		name string
		arg  Value
		want bytecode.ValueType
		val  float64
	}{
		{name: "fabsf", arg: FloatValue(bytecode.TypeF32, -2), want: bytecode.TypeF32, val: 2},
		{name: "fabs", arg: FloatValue(bytecode.TypeF64, -3), want: bytecode.TypeF64, val: 3},
		{name: "fabsl", arg: FloatValue(bytecode.TypeFLong, -4), want: bytecode.TypeFLong, val: 4},
		{name: "sqrtf", arg: FloatValue(bytecode.TypeF32, 4), want: bytecode.TypeF32, val: 2},
		{name: "sqrt", arg: FloatValue(bytecode.TypeF64, 9), want: bytecode.TypeF64, val: 3},
		{name: "sqrtl", arg: FloatValue(bytecode.TypeFLong, 16), want: bytecode.TypeFLong, val: 4},
		{name: "sinf", arg: FloatValue(bytecode.TypeF32, 0), want: bytecode.TypeF32, val: 0},
		{name: "sin", arg: FloatValue(bytecode.TypeF64, 0), want: bytecode.TypeF64, val: 0},
		{name: "sinl", arg: FloatValue(bytecode.TypeFLong, 0), want: bytecode.TypeFLong, val: 0},
		{name: "cosf", arg: FloatValue(bytecode.TypeF32, 0), want: bytecode.TypeF32, val: 1},
		{name: "cos", arg: FloatValue(bytecode.TypeF64, 0), want: bytecode.TypeF64, val: 1},
		{name: "cosl", arg: FloatValue(bytecode.TypeFLong, 0), want: bytecode.TypeFLong, val: 1},
		{name: "tanf", arg: FloatValue(bytecode.TypeF32, 0), want: bytecode.TypeF32, val: 0},
		{name: "tan", arg: FloatValue(bytecode.TypeF64, 0), want: bytecode.TypeF64, val: 0},
		{name: "tanl", arg: FloatValue(bytecode.TypeFLong, 0), want: bytecode.TypeFLong, val: 0},
		{name: "sinhf", arg: FloatValue(bytecode.TypeF32, 0), want: bytecode.TypeF32, val: 0},
		{name: "sinh", arg: FloatValue(bytecode.TypeF64, 0), want: bytecode.TypeF64, val: 0},
		{name: "sinhl", arg: FloatValue(bytecode.TypeFLong, 0), want: bytecode.TypeFLong, val: 0},
		{name: "coshf", arg: FloatValue(bytecode.TypeF32, 0), want: bytecode.TypeF32, val: 1},
		{name: "cosh", arg: FloatValue(bytecode.TypeF64, 0), want: bytecode.TypeF64, val: 1},
		{name: "coshl", arg: FloatValue(bytecode.TypeFLong, 0), want: bytecode.TypeFLong, val: 1},
		{name: "tanhf", arg: FloatValue(bytecode.TypeF32, 0), want: bytecode.TypeF32, val: 0},
		{name: "tanh", arg: FloatValue(bytecode.TypeF64, 0), want: bytecode.TypeF64, val: 0},
		{name: "tanhl", arg: FloatValue(bytecode.TypeFLong, 0), want: bytecode.TypeFLong, val: 0},
		{name: "asinf", arg: FloatValue(bytecode.TypeF32, 0), want: bytecode.TypeF32, val: 0},
		{name: "asin", arg: FloatValue(bytecode.TypeF64, 0), want: bytecode.TypeF64, val: 0},
		{name: "asinl", arg: FloatValue(bytecode.TypeFLong, 0), want: bytecode.TypeFLong, val: 0},
		{name: "acosf", arg: FloatValue(bytecode.TypeF32, 1), want: bytecode.TypeF32, val: 0},
		{name: "acos", arg: FloatValue(bytecode.TypeF64, 1), want: bytecode.TypeF64, val: 0},
		{name: "acosl", arg: FloatValue(bytecode.TypeFLong, 1), want: bytecode.TypeFLong, val: 0},
		{name: "atanf", arg: FloatValue(bytecode.TypeF32, 0), want: bytecode.TypeF32, val: 0},
		{name: "atan", arg: FloatValue(bytecode.TypeF64, 0), want: bytecode.TypeF64, val: 0},
		{name: "atanl", arg: FloatValue(bytecode.TypeFLong, 0), want: bytecode.TypeFLong, val: 0},
		{name: "asinhf", arg: FloatValue(bytecode.TypeF32, 0), want: bytecode.TypeF32, val: 0},
		{name: "asinh", arg: FloatValue(bytecode.TypeF64, 0), want: bytecode.TypeF64, val: 0},
		{name: "asinhl", arg: FloatValue(bytecode.TypeFLong, 0), want: bytecode.TypeFLong, val: 0},
		{name: "acoshf", arg: FloatValue(bytecode.TypeF32, 1), want: bytecode.TypeF32, val: 0},
		{name: "acosh", arg: FloatValue(bytecode.TypeF64, 1), want: bytecode.TypeF64, val: 0},
		{name: "acoshl", arg: FloatValue(bytecode.TypeFLong, 1), want: bytecode.TypeFLong, val: 0},
		{name: "atanhf", arg: FloatValue(bytecode.TypeF32, 0), want: bytecode.TypeF32, val: 0},
		{name: "atanh", arg: FloatValue(bytecode.TypeF64, 0), want: bytecode.TypeF64, val: 0},
		{name: "atanhl", arg: FloatValue(bytecode.TypeFLong, 0), want: bytecode.TypeFLong, val: 0},
		{name: "cbrtf", arg: FloatValue(bytecode.TypeF32, 8), want: bytecode.TypeF32, val: 2},
		{name: "cbrt", arg: FloatValue(bytecode.TypeF64, 8), want: bytecode.TypeF64, val: 2},
		{name: "cbrtl", arg: FloatValue(bytecode.TypeFLong, 8), want: bytecode.TypeFLong, val: 2},
		{name: "erff", arg: FloatValue(bytecode.TypeF32, 0), want: bytecode.TypeF32, val: 0},
		{name: "erf", arg: FloatValue(bytecode.TypeF64, 0), want: bytecode.TypeF64, val: 0},
		{name: "erfl", arg: FloatValue(bytecode.TypeFLong, 0), want: bytecode.TypeFLong, val: 0},
		{name: "erfcf", arg: FloatValue(bytecode.TypeF32, 0), want: bytecode.TypeF32, val: 1},
		{name: "erfc", arg: FloatValue(bytecode.TypeF64, 0), want: bytecode.TypeF64, val: 1},
		{name: "erfcl", arg: FloatValue(bytecode.TypeFLong, 0), want: bytecode.TypeFLong, val: 1},
		{name: "tgammaf", arg: FloatValue(bytecode.TypeF32, 1), want: bytecode.TypeF32, val: 1},
		{name: "tgamma", arg: FloatValue(bytecode.TypeF64, 1), want: bytecode.TypeF64, val: 1},
		{name: "tgammal", arg: FloatValue(bytecode.TypeFLong, 1), want: bytecode.TypeFLong, val: 1},
		{name: "lgammaf", arg: FloatValue(bytecode.TypeF32, 1), want: bytecode.TypeF32, val: 0},
		{name: "lgamma", arg: FloatValue(bytecode.TypeF64, 1), want: bytecode.TypeF64, val: 0},
		{name: "lgammal", arg: FloatValue(bytecode.TypeFLong, 1), want: bytecode.TypeFLong, val: 0},
		{name: "expf", arg: FloatValue(bytecode.TypeF32, 0), want: bytecode.TypeF32, val: 1},
		{name: "exp", arg: FloatValue(bytecode.TypeF64, 0), want: bytecode.TypeF64, val: 1},
		{name: "expl", arg: FloatValue(bytecode.TypeFLong, 0), want: bytecode.TypeFLong, val: 1},
		{name: "exp2f", arg: FloatValue(bytecode.TypeF32, 3), want: bytecode.TypeF32, val: 8},
		{name: "exp2", arg: FloatValue(bytecode.TypeF64, 3), want: bytecode.TypeF64, val: 8},
		{name: "exp2l", arg: FloatValue(bytecode.TypeFLong, 3), want: bytecode.TypeFLong, val: 8},
		{name: "expm1f", arg: FloatValue(bytecode.TypeF32, 0), want: bytecode.TypeF32, val: 0},
		{name: "expm1", arg: FloatValue(bytecode.TypeF64, 0), want: bytecode.TypeF64, val: 0},
		{name: "expm1l", arg: FloatValue(bytecode.TypeFLong, 0), want: bytecode.TypeFLong, val: 0},
		{name: "logf", arg: FloatValue(bytecode.TypeF32, 1), want: bytecode.TypeF32, val: 0},
		{name: "log", arg: FloatValue(bytecode.TypeF64, 1), want: bytecode.TypeF64, val: 0},
		{name: "logl", arg: FloatValue(bytecode.TypeFLong, 1), want: bytecode.TypeFLong, val: 0},
		{name: "log10f", arg: FloatValue(bytecode.TypeF32, 100), want: bytecode.TypeF32, val: 2},
		{name: "log10", arg: FloatValue(bytecode.TypeF64, 100), want: bytecode.TypeF64, val: 2},
		{name: "log10l", arg: FloatValue(bytecode.TypeFLong, 100), want: bytecode.TypeFLong, val: 2},
		{name: "log1pf", arg: FloatValue(bytecode.TypeF32, 0), want: bytecode.TypeF32, val: 0},
		{name: "log1p", arg: FloatValue(bytecode.TypeF64, 0), want: bytecode.TypeF64, val: 0},
		{name: "log1pl", arg: FloatValue(bytecode.TypeFLong, 0), want: bytecode.TypeFLong, val: 0},
		{name: "log2f", arg: FloatValue(bytecode.TypeF32, 8), want: bytecode.TypeF32, val: 3},
		{name: "log2", arg: FloatValue(bytecode.TypeF64, 8), want: bytecode.TypeF64, val: 3},
		{name: "log2l", arg: FloatValue(bytecode.TypeFLong, 8), want: bytecode.TypeFLong, val: 3},
		{name: "nearbyintf", arg: FloatValue(bytecode.TypeF32, 1.5), want: bytecode.TypeF32, val: 2},
		{name: "nearbyint", arg: FloatValue(bytecode.TypeF64, 1.5), want: bytecode.TypeF64, val: 2},
		{name: "nearbyintl", arg: FloatValue(bytecode.TypeFLong, 1.5), want: bytecode.TypeFLong, val: 2},
		{name: "rintf", arg: FloatValue(bytecode.TypeF32, 1.5), want: bytecode.TypeF32, val: 2},
		{name: "rint", arg: FloatValue(bytecode.TypeF64, 1.5), want: bytecode.TypeF64, val: 2},
		{name: "rintl", arg: FloatValue(bytecode.TypeFLong, 1.5), want: bytecode.TypeFLong, val: 2},
		{name: "logbf", arg: FloatValue(bytecode.TypeF32, 8), want: bytecode.TypeF32, val: 3},
		{name: "logb", arg: FloatValue(bytecode.TypeF64, 8), want: bytecode.TypeF64, val: 3},
		{name: "logbl", arg: FloatValue(bytecode.TypeFLong, 8), want: bytecode.TypeFLong, val: 3},
		{name: "ceilf", arg: FloatValue(bytecode.TypeF32, 1.25), want: bytecode.TypeF32, val: 2},
		{name: "ceil", arg: FloatValue(bytecode.TypeF64, 1.25), want: bytecode.TypeF64, val: 2},
		{name: "ceill", arg: FloatValue(bytecode.TypeFLong, 1.25), want: bytecode.TypeFLong, val: 2},
		{name: "floorf", arg: FloatValue(bytecode.TypeF32, 1.75), want: bytecode.TypeF32, val: 1},
		{name: "floor", arg: FloatValue(bytecode.TypeF64, 1.75), want: bytecode.TypeF64, val: 1},
		{name: "floorl", arg: FloatValue(bytecode.TypeFLong, 1.75), want: bytecode.TypeFLong, val: 1},
		{name: "truncf", arg: FloatValue(bytecode.TypeF32, -1.75), want: bytecode.TypeF32, val: -1},
		{name: "trunc", arg: FloatValue(bytecode.TypeF64, -1.75), want: bytecode.TypeF64, val: -1},
		{name: "truncl", arg: FloatValue(bytecode.TypeFLong, -1.75), want: bytecode.TypeFLong, val: -1},
		{name: "roundf", arg: FloatValue(bytecode.TypeF32, 1.5), want: bytecode.TypeF32, val: 2},
		{name: "round", arg: FloatValue(bytecode.TypeF64, 1.5), want: bytecode.TypeF64, val: 2},
		{name: "roundl", arg: FloatValue(bytecode.TypeFLong, 1.5), want: bytecode.TypeFLong, val: 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn, ok := reg.Lookup(tt.name)
			if !ok {
				t.Fatalf("missing %s extern", tt.name)
			}
			ret, exit, err := fn(context.Background(), nil, []Value{tt.arg})
			if err != nil || exit != nil {
				t.Fatalf("%s ret=%#v exit=%#v err=%v", tt.name, ret, exit, err)
			}
			if ret.Type != tt.want || ret.Float != tt.val {
				t.Fatalf("%s ret=%#v, want %s %v", tt.name, ret, tt.want, tt.val)
			}
		})
	}
}

func TestPlainMathBinaryExterns(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	tests := []struct {
		name string
		args []Value
		want bytecode.ValueType
		val  float64
	}{
		{name: "powf", args: []Value{FloatValue(bytecode.TypeF32, 2), FloatValue(bytecode.TypeF32, 3)}, want: bytecode.TypeF32, val: 8},
		{name: "pow", args: []Value{FloatValue(bytecode.TypeF64, 2), FloatValue(bytecode.TypeF64, 3)}, want: bytecode.TypeF64, val: 8},
		{name: "powl", args: []Value{FloatValue(bytecode.TypeFLong, 2), FloatValue(bytecode.TypeFLong, 3)}, want: bytecode.TypeFLong, val: 8},
		{name: "atan2f", args: []Value{FloatValue(bytecode.TypeF32, 0), FloatValue(bytecode.TypeF32, 1)}, want: bytecode.TypeF32, val: 0},
		{name: "atan2", args: []Value{FloatValue(bytecode.TypeF64, 0), FloatValue(bytecode.TypeF64, 1)}, want: bytecode.TypeF64, val: 0},
		{name: "atan2l", args: []Value{FloatValue(bytecode.TypeFLong, 0), FloatValue(bytecode.TypeFLong, 1)}, want: bytecode.TypeFLong, val: 0},
		{name: "hypotf", args: []Value{FloatValue(bytecode.TypeF32, 3), FloatValue(bytecode.TypeF32, 4)}, want: bytecode.TypeF32, val: 5},
		{name: "hypot", args: []Value{FloatValue(bytecode.TypeF64, 3), FloatValue(bytecode.TypeF64, 4)}, want: bytecode.TypeF64, val: 5},
		{name: "hypotl", args: []Value{FloatValue(bytecode.TypeFLong, 3), FloatValue(bytecode.TypeFLong, 4)}, want: bytecode.TypeFLong, val: 5},
		{name: "fdimf", args: []Value{FloatValue(bytecode.TypeF32, 5), FloatValue(bytecode.TypeF32, 3)}, want: bytecode.TypeF32, val: 2},
		{name: "fdim", args: []Value{FloatValue(bytecode.TypeF64, 5), FloatValue(bytecode.TypeF64, 3)}, want: bytecode.TypeF64, val: 2},
		{name: "fdiml", args: []Value{FloatValue(bytecode.TypeFLong, 5), FloatValue(bytecode.TypeFLong, 3)}, want: bytecode.TypeFLong, val: 2},
		{name: "fmaxf", args: []Value{FloatValue(bytecode.TypeF32, 2), FloatValue(bytecode.TypeF32, 3)}, want: bytecode.TypeF32, val: 3},
		{name: "fmax", args: []Value{FloatValue(bytecode.TypeF64, 2), FloatValue(bytecode.TypeF64, 3)}, want: bytecode.TypeF64, val: 3},
		{name: "fmaxl", args: []Value{FloatValue(bytecode.TypeFLong, 2), FloatValue(bytecode.TypeFLong, 3)}, want: bytecode.TypeFLong, val: 3},
		{name: "fminf", args: []Value{FloatValue(bytecode.TypeF32, 2), FloatValue(bytecode.TypeF32, 3)}, want: bytecode.TypeF32, val: 2},
		{name: "fmin", args: []Value{FloatValue(bytecode.TypeF64, 2), FloatValue(bytecode.TypeF64, 3)}, want: bytecode.TypeF64, val: 2},
		{name: "fminl", args: []Value{FloatValue(bytecode.TypeFLong, 2), FloatValue(bytecode.TypeFLong, 3)}, want: bytecode.TypeFLong, val: 2},
		{name: "fmodf", args: []Value{FloatValue(bytecode.TypeF32, 5), FloatValue(bytecode.TypeF32, 2)}, want: bytecode.TypeF32, val: 1},
		{name: "fmod", args: []Value{FloatValue(bytecode.TypeF64, 5), FloatValue(bytecode.TypeF64, 2)}, want: bytecode.TypeF64, val: 1},
		{name: "fmodl", args: []Value{FloatValue(bytecode.TypeFLong, 5), FloatValue(bytecode.TypeFLong, 2)}, want: bytecode.TypeFLong, val: 1},
		{name: "remainderf", args: []Value{FloatValue(bytecode.TypeF32, 4), FloatValue(bytecode.TypeF32, 2)}, want: bytecode.TypeF32, val: 0},
		{name: "remainder", args: []Value{FloatValue(bytecode.TypeF64, 4), FloatValue(bytecode.TypeF64, 2)}, want: bytecode.TypeF64, val: 0},
		{name: "remainderl", args: []Value{FloatValue(bytecode.TypeFLong, 4), FloatValue(bytecode.TypeFLong, 2)}, want: bytecode.TypeFLong, val: 0},
		{name: "copysignf", args: []Value{FloatValue(bytecode.TypeF32, 2), FloatValue(bytecode.TypeF32, -1)}, want: bytecode.TypeF32, val: -2},
		{name: "copysign", args: []Value{FloatValue(bytecode.TypeF64, 2), FloatValue(bytecode.TypeF64, -1)}, want: bytecode.TypeF64, val: -2},
		{name: "copysignl", args: []Value{FloatValue(bytecode.TypeFLong, 2), FloatValue(bytecode.TypeFLong, -1)}, want: bytecode.TypeFLong, val: -2},
		{name: "nextafterf", args: []Value{FloatValue(bytecode.TypeF32, 1), FloatValue(bytecode.TypeF32, 1)}, want: bytecode.TypeF32, val: 1},
		{name: "nextafter", args: []Value{FloatValue(bytecode.TypeF64, 1), FloatValue(bytecode.TypeF64, 1)}, want: bytecode.TypeF64, val: 1},
		{name: "nextafterl", args: []Value{FloatValue(bytecode.TypeFLong, 1), FloatValue(bytecode.TypeFLong, 1)}, want: bytecode.TypeFLong, val: 1},
		{name: "nexttowardf", args: []Value{FloatValue(bytecode.TypeF32, 1), FloatValue(bytecode.TypeFLong, 1)}, want: bytecode.TypeF32, val: 1},
		{name: "nexttoward", args: []Value{FloatValue(bytecode.TypeF64, 1), FloatValue(bytecode.TypeFLong, 1)}, want: bytecode.TypeF64, val: 1},
		{name: "nexttowardl", args: []Value{FloatValue(bytecode.TypeFLong, 1), FloatValue(bytecode.TypeFLong, 1)}, want: bytecode.TypeFLong, val: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn, ok := reg.Lookup(tt.name)
			if !ok {
				t.Fatalf("missing %s extern", tt.name)
			}
			ret, exit, err := fn(context.Background(), nil, tt.args)
			if err != nil || exit != nil {
				t.Fatalf("%s ret=%#v exit=%#v err=%v", tt.name, ret, exit, err)
			}
			if ret.Type != tt.want || ret.Float != tt.val {
				t.Fatalf("%s ret=%#v, want %s %v", tt.name, ret, tt.want, tt.val)
			}
		})
	}
}

func TestPlainMathTernaryExterns(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	tests := []struct {
		name string
		args []Value
		want bytecode.ValueType
		val  float64
	}{
		{name: "fmaf", args: []Value{FloatValue(bytecode.TypeF32, 2), FloatValue(bytecode.TypeF32, 3), FloatValue(bytecode.TypeF32, 4)}, want: bytecode.TypeF32, val: 10},
		{name: "fma", args: []Value{FloatValue(bytecode.TypeF64, 2), FloatValue(bytecode.TypeF64, 3), FloatValue(bytecode.TypeF64, 4)}, want: bytecode.TypeF64, val: 10},
		{name: "fmal", args: []Value{FloatValue(bytecode.TypeFLong, 2), FloatValue(bytecode.TypeFLong, 3), FloatValue(bytecode.TypeFLong, -4)}, want: bytecode.TypeFLong, val: 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn, ok := reg.Lookup(tt.name)
			if !ok {
				t.Fatalf("missing %s extern", tt.name)
			}
			ret, exit, err := fn(context.Background(), nil, tt.args)
			if err != nil || exit != nil {
				t.Fatalf("%s ret=%#v exit=%#v err=%v", tt.name, ret, exit, err)
			}
			if ret.Type != tt.want || ret.Float != tt.val {
				t.Fatalf("%s ret=%#v, want %s %v", tt.name, ret, tt.want, tt.val)
			}
		})
	}
}

func TestPlainMathNanExterns(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	tests := []struct {
		name string
		want bytecode.ValueType
	}{
		{name: "nanf", want: bytecode.TypeF32},
		{name: "nan", want: bytecode.TypeF64},
		{name: "nanl", want: bytecode.TypeFLong},
	}
	payloads := []string{"", "123", "payload"}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn, ok := reg.Lookup(tt.name)
			if !ok {
				t.Fatalf("missing %s extern", tt.name)
			}
			for _, payload := range payloads {
				addr := mustAllocBytes(t, mem, "nan:"+tt.name+":"+payload, []byte(payload+"\x00"), true, blockString)
				ret, exit, err := fn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(addr)})
				if err != nil || exit != nil {
					t.Fatalf("%s(%q) ret=%#v exit=%#v err=%v", tt.name, payload, ret, exit, err)
				}
				if ret.Type != tt.want || !math.IsNaN(ret.Float) {
					t.Fatalf("%s(%q) ret=%#v, want %s NaN", tt.name, payload, ret, tt.want)
				}
			}
		})
	}
}

func TestPlainMathModfExterns(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	tests := []struct {
		name string
		arg  Value
		want bytecode.ValueType
		size int64
	}{
		{name: "modff", arg: FloatValue(bytecode.TypeF32, 3.25), want: bytecode.TypeF32, size: 4},
		{name: "modf", arg: FloatValue(bytecode.TypeF64, 3.25), want: bytecode.TypeF64, size: 8},
		{name: "modfl", arg: FloatValue(bytecode.TypeFLong, 3.25), want: bytecode.TypeFLong, size: 16},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn, ok := reg.Lookup(tt.name)
			if !ok {
				t.Fatalf("missing %s extern", tt.name)
			}
			mem := NewMemory(bytecode.DefaultTarget())
			intPartAddr := mustAlloc(t, mem, "intpart", tt.size, tt.size, false, blockLocal)
			ret, exit, err := fn(context.Background(), &ExternContext{Memory: mem}, []Value{
				tt.arg,
				PtrValue(intPartAddr),
			})
			if err != nil || exit != nil {
				t.Fatalf("%s ret=%#v exit=%#v err=%v", tt.name, ret, exit, err)
			}
			if ret.Type != tt.want || ret.Float != 0.25 {
				t.Fatalf("%s ret = %#v, want %s 0.25", tt.name, ret, tt.want)
			}
			intPart, err := mem.Load(intPartAddr, tt.want, tt.size)
			if err != nil {
				t.Fatalf("Load integer part: %v", err)
			}
			if intPart.Float != 3 {
				t.Fatalf("modf integer part = %#v, want 3", intPart)
			}
		})
	}
}

func TestPutsWritesCString(t *testing.T) {
	for _, name := range []string{"puts", "puts_unlocked"} {
		var out bytes.Buffer
		reg := DefaultExternRegistry(&out, nil)
		mem := NewMemory(bytecode.DefaultTarget())
		addr := mustAllocBytes(t, mem, "string:0", []byte("hello\x00"), true, blockString)
		fn, ok := reg.Lookup(name)
		if !ok {
			t.Fatalf("missing %s extern", name)
		}
		ret, exit, err := fn(context.Background(), &ExternContext{Memory: mem}, []Value{ObjectAddrValue(addr)})
		if err != nil || exit != nil {
			t.Fatalf("%s ret=%#v exit=%#v err=%v", name, ret, exit, err)
		}
		if ret.Type != bytecode.TypeI32 || ret.Int != 6 {
			t.Fatalf("%s ret=%#v, want i32 6", name, ret)
		}
		if out.String() != "hello\n" {
			t.Fatalf("%s output = %q", name, out.String())
		}
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

func TestFwideTracksStreamOrientation(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer
	reg := DefaultExternRegistry(&out, &errOut)
	mem := NewMemory(bytecode.DefaultTarget())
	stdout, ok := reg.LookupVariable("stdout", mem)
	if !ok {
		t.Fatal("missing stdout extern variable")
	}
	stderr, ok := reg.LookupVariable("stderr", mem)
	if !ok {
		t.Fatal("missing stderr extern variable")
	}
	stdin, ok := reg.LookupVariable("stdin", mem)
	if !ok {
		t.Fatal("missing stdin extern variable")
	}
	fwideFn, ok := reg.Lookup("fwide")
	if !ok {
		t.Fatal("missing fwide extern")
	}
	fputcFn, ok := reg.Lookup("fputc")
	if !ok {
		t.Fatal("missing fputc extern")
	}
	fgetcFn, ok := reg.Lookup("fgetc")
	if !ok {
		t.Fatal("missing fgetc extern")
	}
	fcloseFn, ok := reg.Lookup("fclose")
	if !ok {
		t.Fatal("missing fclose extern")
	}

	ret, exit, err := fwideFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(stdout), IntValue(bytecode.TypeI32, 0)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeI32 || signedInt(ret) != 0 {
		t.Fatalf("fwide query ret=%#v exit=%#v err=%v, want 0", ret, exit, err)
	}
	ret, exit, err = fwideFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(stdout), IntValue(bytecode.TypeI32, 1)})
	if err != nil || exit != nil || signedInt(ret) <= 0 {
		t.Fatalf("fwide wide ret=%#v exit=%#v err=%v, want positive", ret, exit, err)
	}
	ret, exit, err = fwideFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(stdout), IntValue(bytecode.TypeI32, -1)})
	if err != nil || exit != nil || signedInt(ret) <= 0 {
		t.Fatalf("fwide wide lock ret=%#v exit=%#v err=%v, want still positive", ret, exit, err)
	}

	ret, exit, err = fwideFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(stderr), IntValue(bytecode.TypeI32, -1)})
	if err != nil || exit != nil || signedInt(ret) >= 0 {
		t.Fatalf("fwide byte ret=%#v exit=%#v err=%v, want negative", ret, exit, err)
	}
	ret, exit, err = fwideFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(stderr), IntValue(bytecode.TypeI32, 1)})
	if err != nil || exit != nil || signedInt(ret) >= 0 {
		t.Fatalf("fwide byte lock ret=%#v exit=%#v err=%v, want still negative", ret, exit, err)
	}

	ret, exit, err = fgetcFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(stdin)})
	if err != nil || exit != nil || int32(ret.Int) != -1 {
		t.Fatalf("fgetc ret=%#v exit=%#v err=%v, want EOF", ret, exit, err)
	}
	ret, exit, err = fwideFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(stdin), IntValue(bytecode.TypeI32, 0)})
	if err != nil || exit != nil || signedInt(ret) >= 0 {
		t.Fatalf("fwide after fgetc ret=%#v exit=%#v err=%v, want negative", ret, exit, err)
	}

	ret, exit, err = fputcFn(context.Background(), &ExternContext{Memory: mem}, []Value{IntValue(bytecode.TypeI32, 'x'), PtrValue(stderr)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeI32 || ret.Int != 'x' {
		t.Fatalf("fputc ret=%#v exit=%#v err=%v, want x", ret, exit, err)
	}
	ret, exit, err = fwideFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(0xfeed), IntValue(bytecode.TypeI32, 0)})
	if err == nil || exit != nil {
		t.Fatalf("fwide unknown ret=%#v exit=%#v err=%v, want error", ret, exit, err)
	}
	ret, exit, err = fcloseFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(stdin)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeI32 || ret.Int != 0 {
		t.Fatalf("fclose ret=%#v exit=%#v err=%v, want 0", ret, exit, err)
	}
	ret, exit, err = fwideFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(stdin), IntValue(bytecode.TypeI32, 0)})
	if err == nil || exit != nil {
		t.Fatalf("fwide closed ret=%#v exit=%#v err=%v, want error", ret, exit, err)
	}
}

func TestFputwcWritesWideCharToHostHandle(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer
	reg := DefaultExternRegistry(&out, &errOut)
	mem := NewMemory(bytecode.DefaultTarget())
	stdout, ok := reg.LookupVariable("stdout", mem)
	if !ok {
		t.Fatal("missing stdout extern variable")
	}
	stderr, ok := reg.LookupVariable("stderr", mem)
	if !ok {
		t.Fatal("missing stderr extern variable")
	}
	fputwcFn, ok := reg.Lookup("fputwc")
	if !ok {
		t.Fatal("missing fputwc extern")
	}
	fwideFn, ok := reg.Lookup("fwide")
	if !ok {
		t.Fatal("missing fwide extern")
	}
	ferrorFn, ok := reg.Lookup("ferror")
	if !ok {
		t.Fatal("missing ferror extern")
	}
	fcloseFn, ok := reg.Lookup("fclose")
	if !ok {
		t.Fatal("missing fclose extern")
	}

	ret, exit, err := fputwcFn(context.Background(), &ExternContext{Memory: mem}, []Value{IntValue(bytecode.TypeI32, 'A'), PtrValue(stdout)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeI32 || ret.Int != 'A' || out.String() != "A" {
		t.Fatalf("fputwc ascii ret=%#v exit=%#v err=%v output=%q, want A", ret, exit, err, out.String())
	}
	ret, exit, err = fwideFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(stdout), IntValue(bytecode.TypeI32, 0)})
	if err != nil || exit != nil || signedInt(ret) <= 0 {
		t.Fatalf("fwide after fputwc ret=%#v exit=%#v err=%v, want positive", ret, exit, err)
	}
	ret, exit, err = fputwcFn(context.Background(), &ExternContext{Memory: mem}, []Value{IntValue(bytecode.TypeI32, 0), PtrValue(stdout)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeI32 || ret.Int != 0 || !bytes.Equal(out.Bytes(), []byte{'A', 0}) {
		t.Fatalf("fputwc nul ret=%#v exit=%#v err=%v output=%v, want A NUL", ret, exit, err, out.Bytes())
	}
	ret, exit, err = fputwcFn(context.Background(), &ExternContext{Memory: mem}, []Value{IntValue(bytecode.TypeI32, 0x80), PtrValue(stdout)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeI32 || signedInt(ret) != -1 || !bytes.Equal(out.Bytes(), []byte{'A', 0}) {
		t.Fatalf("fputwc high ret=%#v exit=%#v err=%v output=%v, want WEOF and unchanged output", ret, exit, err, out.Bytes())
	}
	ret, exit, err = ferrorFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(stdout)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeI32 || ret.Int != 1 {
		t.Fatalf("ferror ret=%#v exit=%#v err=%v, want 1 after high wchar", ret, exit, err)
	}

	ret, exit, err = fwideFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(stderr), IntValue(bytecode.TypeI32, -1)})
	if err != nil || exit != nil || signedInt(ret) >= 0 {
		t.Fatalf("fwide byte ret=%#v exit=%#v err=%v, want negative", ret, exit, err)
	}
	ret, exit, err = fputwcFn(context.Background(), &ExternContext{Memory: mem}, []Value{IntValue(bytecode.TypeI32, 'B'), PtrValue(stderr)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeI32 || signedInt(ret) != -1 || errOut.String() != "" {
		t.Fatalf("fputwc byte-oriented ret=%#v exit=%#v err=%v stderr=%q, want WEOF and no write", ret, exit, err, errOut.String())
	}
	ret, exit, err = fcloseFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(stderr)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeI32 || ret.Int != 0 {
		t.Fatalf("fclose ret=%#v exit=%#v err=%v, want 0", ret, exit, err)
	}
	ret, exit, err = fputwcFn(context.Background(), &ExternContext{Memory: mem}, []Value{IntValue(bytecode.TypeI32, 'C'), PtrValue(stderr)})
	if err == nil || exit != nil {
		t.Fatalf("fputwc closed ret=%#v exit=%#v err=%v, want error", ret, exit, err)
	}
}

func TestWideOutputAliasesWriteWideChars(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer
	reg := DefaultExternRegistry(&out, &errOut)
	mem := NewMemory(bytecode.DefaultTarget())
	stdout, ok := reg.LookupVariable("stdout", mem)
	if !ok {
		t.Fatal("missing stdout extern variable")
	}
	stderr, ok := reg.LookupVariable("stderr", mem)
	if !ok {
		t.Fatal("missing stderr extern variable")
	}
	putwcFn, ok := reg.Lookup("putwc")
	if !ok {
		t.Fatal("missing putwc extern")
	}
	putwcharFn, ok := reg.Lookup("putwchar")
	if !ok {
		t.Fatal("missing putwchar extern")
	}
	fwideFn, ok := reg.Lookup("fwide")
	if !ok {
		t.Fatal("missing fwide extern")
	}

	ret, exit, err := putwcFn(context.Background(), &ExternContext{Memory: mem}, []Value{IntValue(bytecode.TypeI32, 'Q'), PtrValue(stderr)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeI32 || ret.Int != 'Q' || errOut.String() != "Q" {
		t.Fatalf("putwc ret=%#v exit=%#v err=%v stderr=%q, want Q", ret, exit, err, errOut.String())
	}
	ret, exit, err = fwideFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(stderr), IntValue(bytecode.TypeI32, 0)})
	if err != nil || exit != nil || signedInt(ret) <= 0 {
		t.Fatalf("fwide stderr ret=%#v exit=%#v err=%v, want positive", ret, exit, err)
	}
	ret, exit, err = putwcharFn(context.Background(), &ExternContext{Memory: mem}, []Value{IntValue(bytecode.TypeI32, 'R')})
	if err != nil || exit != nil || ret.Type != bytecode.TypeI32 || ret.Int != 'R' || out.String() != "R" {
		t.Fatalf("putwchar ret=%#v exit=%#v err=%v stdout=%q, want R", ret, exit, err, out.String())
	}
	ret, exit, err = fwideFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(stdout), IntValue(bytecode.TypeI32, 0)})
	if err != nil || exit != nil || signedInt(ret) <= 0 {
		t.Fatalf("fwide stdout ret=%#v exit=%#v err=%v, want positive", ret, exit, err)
	}
}

func TestFgetwcReadsWideCharFromHostHandle(t *testing.T) {
	reg := DefaultExternRegistryWithIO(strings.NewReader("A\x00"), nil, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	stdin, ok := reg.LookupVariable("stdin", mem)
	if !ok {
		t.Fatal("missing stdin extern variable")
	}
	fgetwcFn, ok := reg.Lookup("fgetwc")
	if !ok {
		t.Fatal("missing fgetwc extern")
	}
	fwideFn, ok := reg.Lookup("fwide")
	if !ok {
		t.Fatal("missing fwide extern")
	}

	ret, exit, err := fgetwcFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(stdin)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeI32 || ret.Int != 'A' {
		t.Fatalf("fgetwc ascii ret=%#v exit=%#v err=%v, want A", ret, exit, err)
	}
	ret, exit, err = fwideFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(stdin), IntValue(bytecode.TypeI32, 0)})
	if err != nil || exit != nil || signedInt(ret) <= 0 {
		t.Fatalf("fwide after fgetwc ret=%#v exit=%#v err=%v, want positive", ret, exit, err)
	}
	ret, exit, err = fgetwcFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(stdin)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeI32 || ret.Int != 0 {
		t.Fatalf("fgetwc nul ret=%#v exit=%#v err=%v, want 0", ret, exit, err)
	}
	ret, exit, err = fgetwcFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(stdin)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeI32 || signedInt(ret) != -1 {
		t.Fatalf("fgetwc eof ret=%#v exit=%#v err=%v, want WEOF", ret, exit, err)
	}
}

func TestFgetwcRejectsInvalidAndByteOrientedStreams(t *testing.T) {
	reg := DefaultExternRegistryWithIO(bytes.NewReader([]byte{0x80, 'B'}), nil, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	stdin, ok := reg.LookupVariable("stdin", mem)
	if !ok {
		t.Fatal("missing stdin extern variable")
	}
	fgetwcFn, ok := reg.Lookup("fgetwc")
	if !ok {
		t.Fatal("missing fgetwc extern")
	}
	ferrorFn, ok := reg.Lookup("ferror")
	if !ok {
		t.Fatal("missing ferror extern")
	}

	ret, exit, err := fgetwcFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(stdin)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeI32 || signedInt(ret) != -1 {
		t.Fatalf("fgetwc high byte ret=%#v exit=%#v err=%v, want WEOF", ret, exit, err)
	}
	ret, exit, err = ferrorFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(stdin)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeI32 || ret.Int != 1 {
		t.Fatalf("ferror ret=%#v exit=%#v err=%v, want 1 after high byte", ret, exit, err)
	}

	byteReg := DefaultExternRegistryWithIO(strings.NewReader("C"), nil, nil)
	byteMem := NewMemory(bytecode.DefaultTarget())
	byteStdin, ok := byteReg.LookupVariable("stdin", byteMem)
	if !ok {
		t.Fatal("missing byte stdin extern variable")
	}
	byteFwideFn, ok := byteReg.Lookup("fwide")
	if !ok {
		t.Fatal("missing byte fwide extern")
	}
	byteFgetwcFn, ok := byteReg.Lookup("fgetwc")
	if !ok {
		t.Fatal("missing byte fgetwc extern")
	}
	ret, exit, err = byteFwideFn(context.Background(), &ExternContext{Memory: byteMem}, []Value{PtrValue(byteStdin), IntValue(bytecode.TypeI32, -1)})
	if err != nil || exit != nil || signedInt(ret) >= 0 {
		t.Fatalf("byte fwide ret=%#v exit=%#v err=%v, want negative", ret, exit, err)
	}
	ret, exit, err = byteFgetwcFn(context.Background(), &ExternContext{Memory: byteMem}, []Value{PtrValue(byteStdin)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeI32 || signedInt(ret) != -1 {
		t.Fatalf("byte fgetwc ret=%#v exit=%#v err=%v, want WEOF", ret, exit, err)
	}

	closeFn, ok := byteReg.Lookup("fclose")
	if !ok {
		t.Fatal("missing fclose extern")
	}
	ret, exit, err = closeFn(context.Background(), &ExternContext{Memory: byteMem}, []Value{PtrValue(byteStdin)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeI32 || ret.Int != 0 {
		t.Fatalf("fclose ret=%#v exit=%#v err=%v, want 0", ret, exit, err)
	}
	ret, exit, err = byteFgetwcFn(context.Background(), &ExternContext{Memory: byteMem}, []Value{PtrValue(byteStdin)})
	if err == nil || exit != nil {
		t.Fatalf("closed fgetwc ret=%#v exit=%#v err=%v, want error", ret, exit, err)
	}
}

func TestWideInputAliasesReadConfiguredStdin(t *testing.T) {
	reg := DefaultExternRegistryWithIO(strings.NewReader("XY"), nil, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	stdin, ok := reg.LookupVariable("stdin", mem)
	if !ok {
		t.Fatal("missing stdin extern variable")
	}
	getwcFn, ok := reg.Lookup("getwc")
	if !ok {
		t.Fatal("missing getwc extern")
	}
	getwcharFn, ok := reg.Lookup("getwchar")
	if !ok {
		t.Fatal("missing getwchar extern")
	}
	fwideFn, ok := reg.Lookup("fwide")
	if !ok {
		t.Fatal("missing fwide extern")
	}

	ret, exit, err := getwcFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(stdin)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeI32 || ret.Int != 'X' {
		t.Fatalf("getwc ret=%#v exit=%#v err=%v, want X", ret, exit, err)
	}
	ret, exit, err = getwcharFn(context.Background(), &ExternContext{Memory: mem}, nil)
	if err != nil || exit != nil || ret.Type != bytecode.TypeI32 || ret.Int != 'Y' {
		t.Fatalf("getwchar ret=%#v exit=%#v err=%v, want Y", ret, exit, err)
	}
	ret, exit, err = fwideFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(stdin), IntValue(bytecode.TypeI32, 0)})
	if err != nil || exit != nil || signedInt(ret) <= 0 {
		t.Fatalf("fwide stdin ret=%#v exit=%#v err=%v, want positive", ret, exit, err)
	}
}

func TestUngetwcPushesWideCharBack(t *testing.T) {
	reg := DefaultExternRegistryWithIO(strings.NewReader("A"), nil, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	stdin, ok := reg.LookupVariable("stdin", mem)
	if !ok {
		t.Fatal("missing stdin extern variable")
	}
	fgetwcFn, ok := reg.Lookup("fgetwc")
	if !ok {
		t.Fatal("missing fgetwc extern")
	}
	ungetwcFn, ok := reg.Lookup("ungetwc")
	if !ok {
		t.Fatal("missing ungetwc extern")
	}
	feofFn, ok := reg.Lookup("feof")
	if !ok {
		t.Fatal("missing feof extern")
	}

	ret, exit, err := fgetwcFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(stdin)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeI32 || ret.Int != 'A' {
		t.Fatalf("fgetwc ret=%#v exit=%#v err=%v, want A", ret, exit, err)
	}
	ret, exit, err = fgetwcFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(stdin)})
	if err != nil || exit != nil || signedInt(ret) != -1 {
		t.Fatalf("fgetwc eof ret=%#v exit=%#v err=%v, want WEOF", ret, exit, err)
	}
	ret, exit, err = feofFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(stdin)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeI32 || ret.Int != 1 {
		t.Fatalf("feof before ungetwc ret=%#v exit=%#v err=%v, want 1", ret, exit, err)
	}
	ret, exit, err = ungetwcFn(context.Background(), &ExternContext{Memory: mem}, []Value{IntValue(bytecode.TypeI32, 'Z'), PtrValue(stdin)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeI32 || ret.Int != 'Z' {
		t.Fatalf("ungetwc ret=%#v exit=%#v err=%v, want Z", ret, exit, err)
	}
	ret, exit, err = feofFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(stdin)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeI32 || ret.Int != 0 {
		t.Fatalf("feof after ungetwc ret=%#v exit=%#v err=%v, want 0", ret, exit, err)
	}
	ret, exit, err = fgetwcFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(stdin)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeI32 || ret.Int != 'Z' {
		t.Fatalf("fgetwc pushback ret=%#v exit=%#v err=%v, want Z", ret, exit, err)
	}
}

func TestUngetwcRejectsEOFAndInvalidWideChars(t *testing.T) {
	reg := DefaultExternRegistryWithIO(strings.NewReader("B"), nil, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	stdin, ok := reg.LookupVariable("stdin", mem)
	if !ok {
		t.Fatal("missing stdin extern variable")
	}
	ungetwcFn, ok := reg.Lookup("ungetwc")
	if !ok {
		t.Fatal("missing ungetwc extern")
	}
	fgetwcFn, ok := reg.Lookup("fgetwc")
	if !ok {
		t.Fatal("missing fgetwc extern")
	}
	ferrorFn, ok := reg.Lookup("ferror")
	if !ok {
		t.Fatal("missing ferror extern")
	}

	ret, exit, err := ungetwcFn(context.Background(), &ExternContext{Memory: mem}, []Value{IntValue(bytecode.TypeI32, -1), PtrValue(stdin)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeI32 || signedInt(ret) != -1 {
		t.Fatalf("ungetwc WEOF ret=%#v exit=%#v err=%v, want WEOF", ret, exit, err)
	}
	ret, exit, err = fgetwcFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(stdin)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeI32 || ret.Int != 'B' {
		t.Fatalf("fgetwc after WEOF ret=%#v exit=%#v err=%v, want original B", ret, exit, err)
	}
	ret, exit, err = ungetwcFn(context.Background(), &ExternContext{Memory: mem}, []Value{IntValue(bytecode.TypeI32, 0x80), PtrValue(stdin)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeI32 || signedInt(ret) != -1 {
		t.Fatalf("ungetwc high ret=%#v exit=%#v err=%v, want WEOF", ret, exit, err)
	}
	ret, exit, err = ferrorFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(stdin)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeI32 || ret.Int != 1 {
		t.Fatalf("ferror ret=%#v exit=%#v err=%v, want 1 after high wide char", ret, exit, err)
	}
}

func TestFputwsWritesWideStringToHostHandle(t *testing.T) {
	var out bytes.Buffer
	reg := DefaultExternRegistry(&out, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	stdout, ok := reg.LookupVariable("stdout", mem)
	if !ok {
		t.Fatal("missing stdout extern variable")
	}
	fputwsFn, ok := reg.Lookup("fputws")
	if !ok {
		t.Fatal("missing fputws extern")
	}
	fwideFn, ok := reg.Lookup("fwide")
	if !ok {
		t.Fatal("missing fwide extern")
	}
	makeWide := func(name string, chars ...uint32) uint64 {
		t.Helper()
		addr, err := mem.TryAlloc(name, int64(len(chars)*4), 4, false, blockLocal)
		if err != nil {
			t.Fatalf("alloc %s: %v", name, err)
		}
		for i, ch := range chars {
			if err := storeWideChar(mem, addr+uint64(i*4), ch); err != nil {
				t.Fatalf("store %s[%d]: %v", name, i, err)
			}
		}
		return addr
	}

	text := makeWide("fputws:text", 'H', 'i', 0)
	ret, exit, err := fputwsFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(text), PtrValue(stdout)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeI32 || ret.Int != 2 || out.String() != "Hi" {
		t.Fatalf("fputws text ret=%#v exit=%#v err=%v output=%q, want 2 and Hi", ret, exit, err, out.String())
	}
	ret, exit, err = fwideFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(stdout), IntValue(bytecode.TypeI32, 0)})
	if err != nil || exit != nil || signedInt(ret) <= 0 {
		t.Fatalf("fwide after fputws ret=%#v exit=%#v err=%v, want positive", ret, exit, err)
	}
	empty := makeWide("fputws:empty", 0)
	ret, exit, err = fputwsFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(empty), PtrValue(stdout)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeI32 || ret.Int != 0 || out.String() != "Hi" {
		t.Fatalf("fputws empty ret=%#v exit=%#v err=%v output=%q, want 0 and unchanged", ret, exit, err, out.String())
	}
	high := makeWide("fputws:high", '!', 0x80, 'x', 0)
	ret, exit, err = fputwsFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(high), PtrValue(stdout)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeI32 || signedInt(ret) != -1 || out.String() != "Hi!" {
		t.Fatalf("fputws high ret=%#v exit=%#v err=%v output=%q, want WEOF after prefix", ret, exit, err, out.String())
	}
}

func TestFputwsRejectsClosedStream(t *testing.T) {
	var out bytes.Buffer
	reg := DefaultExternRegistry(&out, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	stdout, ok := reg.LookupVariable("stdout", mem)
	if !ok {
		t.Fatal("missing stdout extern variable")
	}
	text := mustAllocBytes(t, mem, "fputws:closed-text", make([]byte, 8), false, blockLocal)
	if err := storeWideChar(mem, text, 'C'); err != nil {
		t.Fatalf("store closed text[0]: %v", err)
	}
	if err := storeWideChar(mem, text+4, 0); err != nil {
		t.Fatalf("store closed text[1]: %v", err)
	}
	fcloseFn, ok := reg.Lookup("fclose")
	if !ok {
		t.Fatal("missing fclose extern")
	}
	fputwsFn, ok := reg.Lookup("fputws")
	if !ok {
		t.Fatal("missing fputws extern")
	}
	ret, exit, err := fcloseFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(stdout)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeI32 || ret.Int != 0 {
		t.Fatalf("fclose ret=%#v exit=%#v err=%v, want 0", ret, exit, err)
	}
	ret, exit, err = fputwsFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(text), PtrValue(stdout)})
	if err == nil || exit != nil {
		t.Fatalf("fputws closed ret=%#v exit=%#v err=%v, want error", ret, exit, err)
	}
}

func TestFgetwsReadsWideLineFromHostHandle(t *testing.T) {
	reg := DefaultExternRegistryWithIO(strings.NewReader("ab\ncd"), nil, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	stdin, ok := reg.LookupVariable("stdin", mem)
	if !ok {
		t.Fatal("missing stdin extern variable")
	}
	buf := mustAllocBytes(t, mem, "fgetws:buf", make([]byte, 16), false, blockLocal)
	fgetwsFn, ok := reg.Lookup("fgetws")
	if !ok {
		t.Fatal("missing fgetws extern")
	}

	ret, exit, err := fgetwsFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(buf), IntValue(bytecode.TypeI32, 4), PtrValue(stdin)})
	if err != nil || exit != nil || ret.Type != bytecode.TypePtr || ret.Int != buf {
		t.Fatalf("fgetws ret=%#v exit=%#v err=%v, want buffer", ret, exit, err)
	}
	for i, want := range []uint32{'a', 'b', '\n', 0} {
		got, err := loadWideChar(mem, buf+uint64(i*4))
		if err != nil || got != want {
			t.Fatalf("fgetws line[%d]=%#x err=%v, want %#x", i, got, err, want)
		}
	}

	ret, exit, err = fgetwsFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(buf), IntValue(bytecode.TypeI32, 3), PtrValue(stdin)})
	if err != nil || exit != nil || ret.Type != bytecode.TypePtr || ret.Int != buf {
		t.Fatalf("fgetws width ret=%#v exit=%#v err=%v, want buffer", ret, exit, err)
	}
	for i, want := range []uint32{'c', 'd', 0} {
		got, err := loadWideChar(mem, buf+uint64(i*4))
		if err != nil || got != want {
			t.Fatalf("fgetws width[%d]=%#x err=%v, want %#x", i, got, err, want)
		}
	}
}

func TestFgetwsEOFAndInvalidInput(t *testing.T) {
	reg := DefaultExternRegistryWithIO(strings.NewReader("x"), nil, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	stdin, ok := reg.LookupVariable("stdin", mem)
	if !ok {
		t.Fatal("missing stdin extern variable")
	}
	buf := mustAllocBytes(t, mem, "fgetws:eof-buf", make([]byte, 8), false, blockLocal)
	fgetwsFn, ok := reg.Lookup("fgetws")
	if !ok {
		t.Fatal("missing fgetws extern")
	}
	feofFn, ok := reg.Lookup("feof")
	if !ok {
		t.Fatal("missing feof extern")
	}

	ret, exit, err := fgetwsFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(buf), IntValue(bytecode.TypeI32, 3), PtrValue(stdin)})
	if err != nil || exit != nil || ret.Type != bytecode.TypePtr || ret.Int != buf {
		t.Fatalf("fgetws data before EOF ret=%#v exit=%#v err=%v, want buffer", ret, exit, err)
	}
	got, err := loadWideChar(mem, buf)
	if err != nil || got != 'x' {
		t.Fatalf("fgetws data stored %#x err=%v, want x", got, err)
	}
	ret, exit, err = fgetwsFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(buf), IntValue(bytecode.TypeI32, 3), PtrValue(stdin)})
	if err != nil || exit != nil || ret.Type != bytecode.TypePtr || ret.Int != 0 {
		t.Fatalf("fgetws EOF ret=%#v exit=%#v err=%v, want null", ret, exit, err)
	}
	ret, exit, err = feofFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(stdin)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeI32 || ret.Int != 1 {
		t.Fatalf("feof ret=%#v exit=%#v err=%v, want 1", ret, exit, err)
	}

	highReg := DefaultExternRegistryWithIO(bytes.NewReader([]byte{0x80}), nil, nil)
	highMem := NewMemory(bytecode.DefaultTarget())
	highStdin, ok := highReg.LookupVariable("stdin", highMem)
	if !ok {
		t.Fatal("missing high stdin extern variable")
	}
	highBuf := mustAllocBytes(t, highMem, "fgetws:high-buf", make([]byte, 8), false, blockLocal)
	highFgetwsFn, ok := highReg.Lookup("fgetws")
	if !ok {
		t.Fatal("missing high fgetws extern")
	}
	ret, exit, err = highFgetwsFn(context.Background(), &ExternContext{Memory: highMem}, []Value{PtrValue(highBuf), IntValue(bytecode.TypeI32, 3), PtrValue(highStdin)})
	if err != nil || exit != nil || ret.Type != bytecode.TypePtr || ret.Int != 0 {
		t.Fatalf("fgetws high ret=%#v exit=%#v err=%v, want null", ret, exit, err)
	}
}

func TestWideFormatCStringBridge(t *testing.T) {
	mem := NewMemory(bytecode.DefaultTarget())
	makeWide := func(name string, chars ...uint32) uint64 {
		t.Helper()
		addr, err := mem.TryAlloc(name, int64(len(chars)*4), 4, false, blockLocal)
		if err != nil {
			t.Fatalf("alloc %s: %v", name, err)
		}
		for i, ch := range chars {
			if err := storeWideChar(mem, addr+uint64(i*4), ch); err != nil {
				t.Fatalf("store %s[%d]: %v", name, i, err)
			}
		}
		return addr
	}
	format := makeWide("wide-format:ascii", '%', '0', '4', 'd', ' ', '%', 's', 0)
	got, err := wideFormatCString("test", mem, format)
	if err != nil || got != "%04d %s" {
		t.Fatalf("wideFormatCString ascii = %q err=%v, want %%04d %%s", got, err)
	}
	high := makeWide("wide-format:high", '%', 'd', 0x80, 0)
	got, err = wideFormatCString("test", mem, high)
	if err == nil || got != "" || !strings.Contains(err.Error(), "unsupported") {
		t.Fatalf("wideFormatCString high = %q err=%v, want unsupported error", got, err)
	}
}

func TestWideScannerBridge(t *testing.T) {
	mem := NewMemory(bytecode.DefaultTarget())
	makeWide := func(name string, chars ...uint32) uint64 {
		t.Helper()
		addr, err := mem.TryAlloc(name, int64(len(chars)*4), 4, false, blockLocal)
		if err != nil {
			t.Fatalf("alloc %s: %v", name, err)
		}
		for i, ch := range chars {
			if err := storeWideChar(mem, addr+uint64(i*4), ch); err != nil {
				t.Fatalf("store %s[%d]: %v", name, i, err)
			}
		}
		return addr
	}
	readWide := func(addr uint64) string {
		t.Helper()
		chars, err := readWideCString(mem, addr)
		if err != nil {
			t.Fatalf("read wide string: %v", err)
		}
		var b strings.Builder
		for _, ch := range chars {
			b.WriteByte(byte(ch))
		}
		return b.String()
	}

	input := makeWide("wide-scan:input", ' ', '-', '1', '2', ' ', '3', '4', ' ', 'h', 'e', 'l', 'l', 'o', ' ', 'Z', 'A', ' ', 's', 'k', 'i', 'p', ' ', '!', 0)
	format := makeWide("wide-scan:format", '%', 'd', ' ', '%', 'u', ' ', '%', '5', 'l', 's', ' ', '%', '2', 'l', 'c', ' ', '%', '*', 's', ' ', '%', 'l', 'c', 0)
	signedAddr := mustAlloc(t, mem, "wide-scan:signed", 4, 4, false, blockLocal)
	unsignedAddr := mustAlloc(t, mem, "wide-scan:unsigned", 4, 4, false, blockLocal)
	wordAddr := mustAllocBytes(t, mem, "wide-scan:word", make([]byte, 24), false, blockLocal)
	charsAddr := mustAllocBytes(t, mem, "wide-scan:chars", make([]byte, 8), false, blockLocal)
	bangAddr := mustAllocBytes(t, mem, "wide-scan:bang", make([]byte, 4), false, blockLocal)

	n, err := scanWideCString("wide-scan", mem, input, format, []Value{
		PtrValue(signedAddr),
		PtrValue(unsignedAddr),
		PtrValue(wordAddr),
		PtrValue(charsAddr),
		PtrValue(bangAddr),
	})
	if err != nil || n != 5 {
		t.Fatalf("scanWideCString assigned=%d err=%v, want 5", n, err)
	}
	signed, err := mem.Load(signedAddr, bytecode.TypeI32, 4)
	if err != nil || int32(signed.Int) != -12 {
		t.Fatalf("signed=%#v err=%v, want -12", signed, err)
	}
	unsigned, err := mem.Load(unsignedAddr, bytecode.TypeU32, 4)
	if err != nil || uint32(unsigned.Int) != 34 {
		t.Fatalf("unsigned=%#v err=%v, want 34", unsigned, err)
	}
	if got := readWide(wordAddr); got != "hello" {
		t.Fatalf("wide word = %q, want hello", got)
	}
	for i, want := range []uint32{'Z', 'A'} {
		got, err := loadWideChar(mem, charsAddr+uint64(i*4))
		if err != nil || got != want {
			t.Fatalf("wide char[%d]=%#x err=%v, want %#x", i, got, err, want)
		}
	}
	got, err := loadWideChar(mem, bangAddr)
	if err != nil || got != '!' {
		t.Fatalf("wide bang=%#x err=%v, want !", got, err)
	}

	empty := makeWide("wide-scan:empty-input", 0)
	intFormat := makeWide("wide-scan:int-format", '%', 'd', 0)
	n, err = scanWideCString("wide-scan", mem, empty, intFormat, []Value{PtrValue(signedAddr)})
	if err != nil || n != -1 {
		t.Fatalf("scanWideCString empty assigned=%d err=%v, want -1", n, err)
	}
	wordInput := makeWide("wide-scan:word-input", 'w', 'o', 'r', 'd', 0)
	n, err = scanWideCString("wide-scan", mem, wordInput, intFormat, []Value{PtrValue(signedAddr)})
	if err != nil || n != 0 {
		t.Fatalf("scanWideCString mismatch assigned=%d err=%v, want 0", n, err)
	}
	highInput := makeWide("wide-scan:high-input", 0x80, 0)
	_, err = scanWideCString("wide-scan", mem, highInput, intFormat, []Value{PtrValue(signedAddr)})
	if err == nil || !strings.Contains(err.Error(), "unsupported") {
		t.Fatalf("scanWideCString high input err=%v, want unsupported", err)
	}
}

func TestSwscanfExternScansWideInput(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	fn, ok := reg.Lookup("swscanf")
	if !ok {
		t.Fatal("missing swscanf extern")
	}
	makeWide := func(name string, chars ...uint32) uint64 {
		t.Helper()
		addr, err := mem.TryAlloc(name, int64(len(chars)*4), 4, false, blockLocal)
		if err != nil {
			t.Fatalf("alloc %s: %v", name, err)
		}
		for i, ch := range chars {
			if err := storeWideChar(mem, addr+uint64(i*4), ch); err != nil {
				t.Fatalf("store %s[%d]: %v", name, i, err)
			}
		}
		return addr
	}
	readWide := func(addr uint64) string {
		t.Helper()
		chars, err := readWideCString(mem, addr)
		if err != nil {
			t.Fatalf("read wide string: %v", err)
		}
		var b strings.Builder
		for _, ch := range chars {
			b.WriteByte(byte(ch))
		}
		return b.String()
	}

	input := makeWide("swscanf:input", ' ', '-', '1', '2', ' ', 'h', 'e', 'l', 'l', 'o', ' ', 'Z', 'A', ' ', '!', 0)
	format := makeWide("swscanf:format", '%', 'd', ' ', '%', '5', 'l', 's', ' ', '%', '2', 'l', 'c', ' ', '%', 'l', 'c', '%', 'n', 0)
	valueAddr := mustAlloc(t, mem, "swscanf:value", 4, 4, false, blockLocal)
	wordAddr := mustAllocBytes(t, mem, "swscanf:word", make([]byte, 24), false, blockLocal)
	charsAddr := mustAllocBytes(t, mem, "swscanf:chars", make([]byte, 8), false, blockLocal)
	bangAddr := mustAllocBytes(t, mem, "swscanf:bang", make([]byte, 4), false, blockLocal)
	countAddr := mustAlloc(t, mem, "swscanf:count", 4, 4, false, blockLocal)

	ret, exit, err := fn(context.Background(), &ExternContext{Memory: mem}, []Value{
		ObjectAddrValue(input),
		ObjectAddrValue(format),
		PtrValue(valueAddr),
		PtrValue(wordAddr),
		PtrValue(charsAddr),
		PtrValue(bangAddr),
		PtrValue(countAddr),
	})
	if err != nil || exit != nil {
		t.Fatalf("swscanf ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeI32 || ret.Int != 4 {
		t.Fatalf("swscanf ret=%#v, want i32 4", ret)
	}
	value, err := mem.Load(valueAddr, bytecode.TypeI32, 4)
	if err != nil || int32(value.Int) != -12 {
		t.Fatalf("value=%#v err=%v, want -12", value, err)
	}
	if got := readWide(wordAddr); got != "hello" {
		t.Fatalf("wide word = %q, want hello", got)
	}
	for i, want := range []uint32{'Z', 'A'} {
		got, err := loadWideChar(mem, charsAddr+uint64(i*4))
		if err != nil || got != want {
			t.Fatalf("wide chars[%d]=%#x err=%v, want %#x", i, got, err, want)
		}
	}
	bang, err := loadWideChar(mem, bangAddr)
	if err != nil || bang != '!' {
		t.Fatalf("bang=%#x err=%v, want !", bang, err)
	}
	count, err := mem.Load(countAddr, bytecode.TypeI32, 4)
	if err != nil || count.Int != 15 {
		t.Fatalf("count=%#v err=%v, want 15", count, err)
	}

	empty := makeWide("swscanf:empty-input", 0)
	intFormat := makeWide("swscanf:int-format", '%', 'd', 0)
	ret, exit, err = fn(context.Background(), &ExternContext{Memory: mem}, []Value{ObjectAddrValue(empty), ObjectAddrValue(intFormat), PtrValue(valueAddr)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeI32 || int64(int32(ret.Int)) != -1 {
		t.Fatalf("swscanf empty ret=%#v exit=%#v err=%v, want -1", ret, exit, err)
	}
	wordInput := makeWide("swscanf:word-input", 'w', 'o', 'r', 'd', 0)
	ret, exit, err = fn(context.Background(), &ExternContext{Memory: mem}, []Value{ObjectAddrValue(wordInput), ObjectAddrValue(intFormat), PtrValue(valueAddr)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeI32 || ret.Int != 0 {
		t.Fatalf("swscanf mismatch ret=%#v exit=%#v err=%v, want 0", ret, exit, err)
	}
}

func TestWscanfExternScansStdinAndPreservesUnreadInput(t *testing.T) {
	reg := DefaultExternRegistryWithIO(strings.NewReader("41 tail"), nil, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	fn, ok := reg.Lookup("wscanf")
	if !ok {
		t.Fatal("missing wscanf extern")
	}
	makeWide := func(name string, chars ...uint32) uint64 {
		t.Helper()
		addr, err := mem.TryAlloc(name, int64(len(chars)*4), 4, false, blockLocal)
		if err != nil {
			t.Fatalf("alloc %s: %v", name, err)
		}
		for i, ch := range chars {
			if err := storeWideChar(mem, addr+uint64(i*4), ch); err != nil {
				t.Fatalf("store %s[%d]: %v", name, i, err)
			}
		}
		return addr
	}
	format := makeWide("wscanf:format", '%', 'd', 0)
	valueAddr := mustAlloc(t, mem, "wscanf:value", 4, 4, false, blockLocal)

	ret, exit, err := fn(context.Background(), &ExternContext{Memory: mem}, []Value{
		ObjectAddrValue(format),
		PtrValue(valueAddr),
	})
	if err != nil || exit != nil {
		t.Fatalf("wscanf ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeI32 || ret.Int != 1 {
		t.Fatalf("wscanf ret=%#v, want i32 1", ret)
	}
	value, err := mem.Load(valueAddr, bytecode.TypeI32, 4)
	if err != nil || value.Int != 41 {
		t.Fatalf("value=%#v err=%v, want 41", value, err)
	}
	if got := reg.hostOrientationResult(reg.stdinHandle); got <= 0 {
		t.Fatalf("stdin orientation = %d, want wide", got)
	}
	ch, ok := reg.readHostChar(reg.stdinHandle)
	if !ok || ch != ' ' {
		t.Fatalf("next stdin char=%q ok=%v, want space", ch, ok)
	}
}

func TestFwscanfExternScansConfiguredFileAndTracksStreamState(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	reg.AddFile("wide.txt", []byte("52 tail"))
	reg.AddFile("word.txt", []byte("word"))
	reg.AddFile("empty.txt", nil)
	mem := NewMemory(bytecode.DefaultTarget())
	makeWide := func(name string, chars ...uint32) uint64 {
		t.Helper()
		addr, err := mem.TryAlloc(name, int64(len(chars)*4), 4, false, blockLocal)
		if err != nil {
			t.Fatalf("alloc %s: %v", name, err)
		}
		for i, ch := range chars {
			if err := storeWideChar(mem, addr+uint64(i*4), ch); err != nil {
				t.Fatalf("store %s[%d]: %v", name, i, err)
			}
		}
		return addr
	}
	openFile := func(name string) uint64 {
		t.Helper()
		path := mustAllocBytes(t, mem, "fwscanf:path:"+name, append([]byte(name), 0), true, blockString)
		mode := mustAllocBytes(t, mem, "fwscanf:mode:"+name, []byte("r\x00"), true, blockString)
		fopenFn, ok := reg.Lookup("fopen")
		if !ok {
			t.Fatal("missing fopen extern")
		}
		ret, exit, err := fopenFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(path), PtrValue(mode)})
		if err != nil || exit != nil || ret.Type != bytecode.TypePtr || ret.Int == 0 {
			t.Fatalf("fopen %s ret=%#v exit=%#v err=%v, want file handle", name, ret, exit, err)
		}
		return ret.Int
	}
	format := makeWide("fwscanf:format", '%', 'd', 0)
	valueAddr := mustAlloc(t, mem, "fwscanf:value", 4, 4, false, blockLocal)
	fn, ok := reg.Lookup("fwscanf")
	if !ok {
		t.Fatal("missing fwscanf extern")
	}

	file := openFile("wide.txt")
	ret, exit, err := fn(context.Background(), &ExternContext{Memory: mem}, []Value{
		PtrValue(file),
		ObjectAddrValue(format),
		PtrValue(valueAddr),
	})
	if err != nil || exit != nil {
		t.Fatalf("fwscanf ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeI32 || ret.Int != 1 {
		t.Fatalf("fwscanf ret=%#v, want i32 1", ret)
	}
	value, err := mem.Load(valueAddr, bytecode.TypeI32, 4)
	if err != nil || value.Int != 52 {
		t.Fatalf("value=%#v err=%v, want 52", value, err)
	}
	if got := reg.hostOrientationResult(file); got <= 0 {
		t.Fatalf("file orientation = %d, want wide", got)
	}
	ch, ok := reg.readHostChar(file)
	if !ok || ch != ' ' {
		t.Fatalf("next file char=%q ok=%v, want space", ch, ok)
	}

	wordFile := openFile("word.txt")
	ret, exit, err = fn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(wordFile), ObjectAddrValue(format), PtrValue(valueAddr)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeI32 || ret.Int != 0 {
		t.Fatalf("fwscanf mismatch ret=%#v exit=%#v err=%v, want 0", ret, exit, err)
	}
	ch, ok = reg.readHostChar(wordFile)
	if !ok || ch != 'w' {
		t.Fatalf("mismatch next file char=%q ok=%v, want w", ch, ok)
	}

	emptyFile := openFile("empty.txt")
	ret, exit, err = fn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(emptyFile), ObjectAddrValue(format), PtrValue(valueAddr)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeI32 || int64(int32(ret.Int)) != -1 {
		t.Fatalf("fwscanf empty ret=%#v exit=%#v err=%v, want -1", ret, exit, err)
	}
	if !reg.hostEOF[emptyFile] {
		t.Fatalf("empty file EOF flag not set")
	}
}

func TestSwprintfWritesWideFormattedOutput(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	buf := mustAllocBytes(t, mem, "swprintf:buf", make([]byte, 64), false, blockLocal)
	str := mustAllocBytes(t, mem, "swprintf:str", []byte("abcdef\x00"), true, blockString)
	makeWide := func(name string, chars ...uint32) uint64 {
		t.Helper()
		addr, err := mem.TryAlloc(name, int64(len(chars)*4), 4, false, blockLocal)
		if err != nil {
			t.Fatalf("alloc %s: %v", name, err)
		}
		for i, ch := range chars {
			if err := storeWideChar(mem, addr+uint64(i*4), ch); err != nil {
				t.Fatalf("store %s[%d]: %v", name, i, err)
			}
		}
		return addr
	}
	readWide := func(addr uint64) string {
		t.Helper()
		chars, err := readWideCString(mem, addr)
		if err != nil {
			t.Fatalf("read wide string: %v", err)
		}
		var b strings.Builder
		for _, ch := range chars {
			b.WriteByte(byte(ch))
		}
		return b.String()
	}
	format := makeWide("swprintf:fmt", '%', '0', '4', 'd', ' ', '%', '.', '2', 's', ' ', '%', 'c', ' ', '%', '%', 0)
	fn, ok := reg.Lookup("swprintf")
	if !ok {
		t.Fatal("missing swprintf extern")
	}
	ret, exit, err := fn(context.Background(), &ExternContext{Memory: mem}, []Value{
		PtrValue(buf),
		UIntValue(bytecode.TypeU64, 16),
		PtrValue(format),
		IntValue(bytecode.TypeI32, 7),
		PtrValue(str),
		IntValue(bytecode.TypeI32, 'Z'),
	})
	if err != nil || exit != nil || ret.Type != bytecode.TypeI32 || ret.Int != 11 {
		t.Fatalf("swprintf ret=%#v exit=%#v err=%v, want 11", ret, exit, err)
	}
	if got := readWide(buf); got != "0007 ab Z %" {
		t.Fatalf("swprintf output = %q, want formatted output", got)
	}
}

func TestSwprintfSupportsWideCharacterFormat(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	buf := mustAllocBytes(t, mem, "swprintf:wide-char-buf", make([]byte, 32), false, blockLocal)
	makeWide := func(name string, chars ...uint32) uint64 {
		t.Helper()
		addr, err := mem.TryAlloc(name, int64(len(chars)*4), 4, false, blockLocal)
		if err != nil {
			t.Fatalf("alloc %s: %v", name, err)
		}
		for i, ch := range chars {
			if err := storeWideChar(mem, addr+uint64(i*4), ch); err != nil {
				t.Fatalf("store %s[%d]: %v", name, i, err)
			}
		}
		return addr
	}
	readWide := func(addr uint64) string {
		t.Helper()
		chars, err := readWideCString(mem, addr)
		if err != nil {
			t.Fatalf("read wide string: %v", err)
		}
		var b strings.Builder
		for _, ch := range chars {
			b.WriteByte(byte(ch))
		}
		return b.String()
	}
	format := makeWide("swprintf:wide-char-fmt", '%', 'c', ' ', '%', 'l', 'c', 0)
	fn, ok := reg.Lookup("swprintf")
	if !ok {
		t.Fatal("missing swprintf extern")
	}
	ret, exit, err := fn(context.Background(), &ExternContext{Memory: mem}, []Value{
		PtrValue(buf),
		UIntValue(bytecode.TypeU64, 8),
		PtrValue(format),
		IntValue(bytecode.TypeI32, 'A'),
		IntValue(bytecode.TypeI32, 'Z'),
	})
	if err != nil || exit != nil || ret.Type != bytecode.TypeI32 || ret.Int != 3 {
		t.Fatalf("swprintf ret=%#v exit=%#v err=%v, want 3", ret, exit, err)
	}
	if got := readWide(buf); got != "A Z" {
		t.Fatalf("swprintf output = %q, want A Z", got)
	}
}

func TestSwprintfSupportsWideStringFormat(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	buf := mustAllocBytes(t, mem, "swprintf:wide-string-buf", make([]byte, 64), false, blockLocal)
	makeWide := func(name string, chars ...uint32) uint64 {
		t.Helper()
		addr, err := mem.TryAlloc(name, int64(len(chars)*4), 4, false, blockLocal)
		if err != nil {
			t.Fatalf("alloc %s: %v", name, err)
		}
		for i, ch := range chars {
			if err := storeWideChar(mem, addr+uint64(i*4), ch); err != nil {
				t.Fatalf("store %s[%d]: %v", name, i, err)
			}
		}
		return addr
	}
	readWide := func(addr uint64) string {
		t.Helper()
		chars, err := readWideCString(mem, addr)
		if err != nil {
			t.Fatalf("read wide string: %v", err)
		}
		var b strings.Builder
		for _, ch := range chars {
			b.WriteByte(byte(ch))
		}
		return b.String()
	}
	format := makeWide("swprintf:wide-string-fmt", '%', 's', ' ', '%', '.', '3', 'l', 's', ' ', '%', 'l', 's', 0)
	narrow := mustAllocBytes(t, mem, "swprintf:wide-string-narrow", []byte("ok\x00"), true, blockString)
	wide := makeWide("swprintf:wide-string-wide", 'a', 'b', 'c', 'd', 'e', 'f', 0)
	fn, ok := reg.Lookup("swprintf")
	if !ok {
		t.Fatal("missing swprintf extern")
	}
	ret, exit, err := fn(context.Background(), &ExternContext{Memory: mem}, []Value{
		PtrValue(buf),
		UIntValue(bytecode.TypeU64, 16),
		PtrValue(format),
		PtrValue(narrow),
		PtrValue(wide),
		PtrValue(wide),
	})
	if err != nil || exit != nil || ret.Type != bytecode.TypeI32 || ret.Int != 13 {
		t.Fatalf("swprintf ret=%#v exit=%#v err=%v, want 13", ret, exit, err)
	}
	if got := readWide(buf); got != "ok abc abcdef" {
		t.Fatalf("swprintf output = %q, want ok abc abcdef", got)
	}
}

func TestSwprintfTruncatesAndRejectsHighFormat(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	buf := mustAllocBytes(t, mem, "swprintf:trunc-buf", make([]byte, 24), false, blockLocal)
	makeWide := func(name string, chars ...uint32) uint64 {
		t.Helper()
		addr, err := mem.TryAlloc(name, int64(len(chars)*4), 4, false, blockLocal)
		if err != nil {
			t.Fatalf("alloc %s: %v", name, err)
		}
		for i, ch := range chars {
			if err := storeWideChar(mem, addr+uint64(i*4), ch); err != nil {
				t.Fatalf("store %s[%d]: %v", name, i, err)
			}
		}
		return addr
	}
	fn, ok := reg.Lookup("swprintf")
	if !ok {
		t.Fatal("missing swprintf extern")
	}
	format := makeWide("swprintf:trunc-fmt", 'a', 'b', 'c', 'd', 'e', 0)
	ret, exit, err := fn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(buf), UIntValue(bytecode.TypeU64, 4), PtrValue(format)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeI32 || signedInt(ret) != -1 {
		t.Fatalf("swprintf trunc ret=%#v exit=%#v err=%v, want -1", ret, exit, err)
	}
	for i, want := range []uint32{'a', 'b', 'c', 0} {
		got, err := loadWideChar(mem, buf+uint64(i*4))
		if err != nil || got != want {
			t.Fatalf("swprintf trunc[%d]=%#x err=%v, want %#x", i, got, err, want)
		}
	}
	high := makeWide("swprintf:high-fmt", '%', 'd', 0x80, 0)
	ret, exit, err = fn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(buf), UIntValue(bytecode.TypeU64, 4), PtrValue(high), IntValue(bytecode.TypeI32, 1)})
	if err == nil || exit != nil {
		t.Fatalf("swprintf high ret=%#v exit=%#v err=%v, want error", ret, exit, err)
	}
}

func TestFwprintfWritesToStream(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer
	reg := DefaultExternRegistry(&out, &errOut)
	mem := NewMemory(bytecode.DefaultTarget())
	stderr, ok := reg.LookupVariable("stderr", mem)
	if !ok {
		t.Fatal("missing stderr extern variable")
	}
	makeWide := func(name string, chars ...uint32) uint64 {
		t.Helper()
		addr, err := mem.TryAlloc(name, int64(len(chars)*4), 4, false, blockLocal)
		if err != nil {
			t.Fatalf("alloc %s: %v", name, err)
		}
		for i, ch := range chars {
			if err := storeWideChar(mem, addr+uint64(i*4), ch); err != nil {
				t.Fatalf("store %s[%d]: %v", name, i, err)
			}
		}
		return addr
	}
	format := makeWide("fwprintf:fmt", 'e', '=', '%', 'd', 0)
	fn, ok := reg.Lookup("fwprintf")
	if !ok {
		t.Fatal("missing fwprintf extern")
	}
	ret, exit, err := fn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(stderr), PtrValue(format), IntValue(bytecode.TypeI32, 9)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeI32 || ret.Int != 3 || errOut.String() != "e=9" || out.String() != "" {
		t.Fatalf("fwprintf ret=%#v exit=%#v err=%v stdout=%q stderr=%q, want e=9", ret, exit, err, out.String(), errOut.String())
	}
	fcloseFn, ok := reg.Lookup("fclose")
	if !ok {
		t.Fatal("missing fclose extern")
	}
	ret, exit, err = fcloseFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(stderr)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeI32 || ret.Int != 0 {
		t.Fatalf("fclose ret=%#v exit=%#v err=%v, want 0", ret, exit, err)
	}
	ret, exit, err = fn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(stderr), PtrValue(format), IntValue(bytecode.TypeI32, 1)})
	if err == nil || exit != nil {
		t.Fatalf("fwprintf closed ret=%#v exit=%#v err=%v, want error", ret, exit, err)
	}
}

func TestWideVPrintfFunctionsReadMemoryVaList(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer
	reg := DefaultExternRegistry(&out, &errOut)
	mem := NewMemory(bytecode.DefaultTarget())
	stderr, ok := reg.LookupVariable("stderr", mem)
	if !ok {
		t.Fatal("missing stderr extern variable")
	}
	buf := mustAlloc(t, mem, "buf:wide-vformat-va-list", 48, 4, false, blockLocal)
	str := mustAllocBytes(t, mem, "str:wide-vformat-va-list", []byte("ok\x00"), true, blockString)
	makeWide := func(name string, chars ...uint32) uint64 {
		t.Helper()
		addr, err := mem.TryAlloc(name, int64(len(chars)*4), 4, false, blockLocal)
		if err != nil {
			t.Fatalf("alloc %s: %v", name, err)
		}
		for i, ch := range chars {
			if err := storeWideChar(mem, addr+uint64(i*4), ch); err != nil {
				t.Fatalf("store %s[%d]: %v", name, i, err)
			}
		}
		return addr
	}
	readWide := func(addr uint64) string {
		t.Helper()
		chars, err := readWideCString(mem, addr)
		if err != nil {
			t.Fatalf("read wide string: %v", err)
		}
		var b strings.Builder
		for _, ch := range chars {
			b.WriteByte(byte(ch))
		}
		return b.String()
	}
	format := makeWide("wide-vformat:fmt", '%', 'd', ' ', '%', 's', 0)
	ap := mustAllocVaList(t, mem, "ap:wide-vformat", IntValue(bytecode.TypeI32, 42), ObjectAddrValue(str))

	vswprintfFn, ok := reg.Lookup("vswprintf")
	if !ok {
		t.Fatal("missing vswprintf extern")
	}
	ret, exit, err := vswprintfFn(context.Background(), &ExternContext{Memory: mem}, []Value{
		ObjectAddrValue(buf),
		UIntValue(bytecode.TypeU64, 16),
		ObjectAddrValue(format),
		PtrValue(ap),
	})
	if err != nil || exit != nil || ret.Type != bytecode.TypeI32 || ret.Int != 5 || readWide(buf) != "42 ok" {
		t.Fatalf("vswprintf ret=%#v exit=%#v err=%v output=%q, want 42 ok", ret, exit, err, readWide(buf))
	}

	vwprintfFn, ok := reg.Lookup("vwprintf")
	if !ok {
		t.Fatal("missing vwprintf extern")
	}
	ret, exit, err = vwprintfFn(context.Background(), &ExternContext{Memory: mem, Stdout: &out}, []Value{
		ObjectAddrValue(format),
		PtrValue(ap),
	})
	if err != nil || exit != nil || ret.Type != bytecode.TypeI32 || ret.Int != 5 || out.String() != "42 ok" {
		t.Fatalf("vwprintf ret=%#v exit=%#v err=%v stdout=%q, want 42 ok", ret, exit, err, out.String())
	}

	vfwprintfFn, ok := reg.Lookup("vfwprintf")
	if !ok {
		t.Fatal("missing vfwprintf extern")
	}
	ret, exit, err = vfwprintfFn(context.Background(), &ExternContext{Memory: mem}, []Value{
		PtrValue(stderr),
		ObjectAddrValue(format),
		PtrValue(ap),
	})
	if err != nil || exit != nil || ret.Type != bytecode.TypeI32 || ret.Int != 5 || errOut.String() != "42 ok" {
		t.Fatalf("vfwprintf ret=%#v exit=%#v err=%v stderr=%q, want 42 ok", ret, exit, err, errOut.String())
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

func TestStdioReadsConfiguredStdin(t *testing.T) {
	reg := DefaultExternRegistryWithIO(strings.NewReader("ABC\nrest"), nil, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	stdin, ok := reg.LookupVariable("stdin", mem)
	if !ok {
		t.Fatal("missing stdin extern variable")
	}
	fgetcFn, ok := reg.Lookup("fgetc")
	if !ok {
		t.Fatal("missing fgetc extern")
	}
	getcharFn, ok := reg.Lookup("getchar")
	if !ok {
		t.Fatal("missing getchar extern")
	}
	fgetsFn, ok := reg.Lookup("fgets")
	if !ok {
		t.Fatal("missing fgets extern")
	}
	freadFn, ok := reg.Lookup("fread")
	if !ok {
		t.Fatal("missing fread extern")
	}

	ret, exit, err := fgetcFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(stdin)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeI32 || ret.Int != 'A' {
		t.Fatalf("fgetc ret=%#v exit=%#v err=%v, want 'A'", ret, exit, err)
	}
	ret, exit, err = getcharFn(context.Background(), &ExternContext{Memory: mem}, nil)
	if err != nil || exit != nil || ret.Type != bytecode.TypeI32 || ret.Int != 'B' {
		t.Fatalf("getchar ret=%#v exit=%#v err=%v, want 'B'", ret, exit, err)
	}

	line := mustAllocBytes(t, mem, "stdin:line", []byte{0, 0, 0, 0}, false, blockLocal)
	ret, exit, err = fgetsFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(line), IntValue(bytecode.TypeI32, 4), PtrValue(stdin)})
	if err != nil || exit != nil || ret.Type != bytecode.TypePtr || ret.Int != line {
		t.Fatalf("fgets ret=%#v exit=%#v err=%v, want buffer", ret, exit, err)
	}
	got, err := mem.ReadCString(line)
	if err != nil {
		t.Fatalf("ReadCString line: %v", err)
	}
	if got != "C\n" {
		t.Fatalf("fgets line = %q, want C newline", got)
	}

	block := mustAllocBytes(t, mem, "stdin:block", []byte{0, 0, 0}, false, blockLocal)
	ret, exit, err = freadFn(context.Background(), &ExternContext{Memory: mem}, []Value{
		PtrValue(block),
		UIntValue(bytecode.TypeU64, 1),
		UIntValue(bytecode.TypeU64, 3),
		PtrValue(stdin),
	})
	if err != nil || exit != nil || ret.Type != bytecode.TypeU64 || ret.Int != 3 {
		t.Fatalf("fread ret=%#v exit=%#v err=%v, want 3", ret, exit, err)
	}
	bytesBlock, off, err := mem.rangeAccess(block, 3, false)
	if err != nil {
		t.Fatalf("rangeAccess block: %v", err)
	}
	if got := string(bytesBlock.data[off : off+3]); got != "res" {
		t.Fatalf("fread bytes = %q, want res", got)
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
		{name: "fgetc_unlocked", args: []Value{PtrValue(stdin)}, ch: 'F'},
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

func TestUnlockedStreamStatusAliases(t *testing.T) {
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
	feofFn, ok := reg.Lookup("feof_unlocked")
	if !ok {
		t.Fatal("missing feof_unlocked extern")
	}
	clearerrFn, ok := reg.Lookup("clearerr_unlocked")
	if !ok {
		t.Fatal("missing clearerr_unlocked extern")
	}
	ferrorFn, ok := reg.Lookup("ferror_unlocked")
	if !ok {
		t.Fatal("missing ferror_unlocked extern")
	}
	if _, exit, err := fgetcFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(stdin)}); err != nil || exit != nil {
		t.Fatalf("fgetc exit=%#v err=%v", exit, err)
	}
	ret, exit, err := feofFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(stdin)})
	if err != nil || exit != nil {
		t.Fatalf("feof_unlocked ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeI32 || ret.Int != 1 {
		t.Fatalf("feof_unlocked ret=%#v, want 1", ret)
	}
	ret, exit, err = ferrorFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(stdin)})
	if err != nil || exit != nil {
		t.Fatalf("ferror_unlocked ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeI32 || ret.Int != 0 {
		t.Fatalf("ferror_unlocked ret=%#v, want 0", ret)
	}
	if _, exit, err := clearerrFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(stdin)}); err != nil || exit != nil {
		t.Fatalf("clearerr_unlocked exit=%#v err=%v", exit, err)
	}
	ret, exit, err = feofFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(stdin)})
	if err != nil || exit != nil {
		t.Fatalf("feof_unlocked after clearerr ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeI32 || ret.Int != 0 {
		t.Fatalf("feof_unlocked after clearerr ret=%#v, want 0", ret)
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

func TestUnlockedBlockIOAliases(t *testing.T) {
	var out bytes.Buffer
	reg := DefaultExternRegistry(&out, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	stdin, ok := reg.LookupVariable("stdin", mem)
	if !ok {
		t.Fatal("missing stdin extern variable")
	}
	stdout, ok := reg.LookupVariable("stdout", mem)
	if !ok {
		t.Fatal("missing stdout extern variable")
	}
	ungetcFn, ok := reg.Lookup("ungetc")
	if !ok {
		t.Fatal("missing ungetc extern")
	}
	for _, ch := range []rune{'Y', 'X'} {
		if _, exit, err := ungetcFn(context.Background(), &ExternContext{Memory: mem}, []Value{IntValue(bytecode.TypeI32, int64(ch)), PtrValue(stdin)}); err != nil || exit != nil {
			t.Fatalf("ungetc %q exit=%#v err=%v", ch, exit, err)
		}
	}
	buf := mustAllocBytes(t, mem, "unlocked:buf", []byte{0, 0, 0}, false, blockLocal)
	freadFn, ok := reg.Lookup("fread_unlocked")
	if !ok {
		t.Fatal("missing fread_unlocked extern")
	}
	ret, exit, err := freadFn(context.Background(), &ExternContext{Memory: mem}, []Value{
		ObjectAddrValue(buf),
		UIntValue(bytecode.TypeU64, 1),
		UIntValue(bytecode.TypeU64, 2),
		PtrValue(stdin),
	})
	if err != nil || exit != nil {
		t.Fatalf("fread_unlocked ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeU64 || ret.Int != 2 {
		t.Fatalf("fread_unlocked ret=%#v, want u64 2", ret)
	}
	fwriteFn, ok := reg.Lookup("fwrite_unlocked")
	if !ok {
		t.Fatal("missing fwrite_unlocked extern")
	}
	ret, exit, err = fwriteFn(context.Background(), &ExternContext{Memory: mem}, []Value{
		ObjectAddrValue(buf),
		UIntValue(bytecode.TypeU64, 1),
		UIntValue(bytecode.TypeU64, 2),
		PtrValue(stdout),
	})
	if err != nil || exit != nil {
		t.Fatalf("fwrite_unlocked ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeU64 || ret.Int != 2 || out.String() != "XY" {
		t.Fatalf("fwrite_unlocked ret=%#v output=%q, want u64 2 and XY", ret, out.String())
	}
	for _, ch := range []rune{'\n', 'Z'} {
		if _, exit, err := ungetcFn(context.Background(), &ExternContext{Memory: mem}, []Value{IntValue(bytecode.TypeI32, int64(ch)), PtrValue(stdin)}); err != nil || exit != nil {
			t.Fatalf("ungetc %q exit=%#v err=%v", ch, exit, err)
		}
	}
	fgetsFn, ok := reg.Lookup("fgets_unlocked")
	if !ok {
		t.Fatal("missing fgets_unlocked extern")
	}
	ret, exit, err = fgetsFn(context.Background(), &ExternContext{Memory: mem}, []Value{
		ObjectAddrValue(buf),
		IntValue(bytecode.TypeI32, 3),
		PtrValue(stdin),
	})
	if err != nil || exit != nil {
		t.Fatalf("fgets_unlocked ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypePtr || ret.Int != buf {
		t.Fatalf("fgets_unlocked ret=%#v, want buffer pointer %#x", ret, buf)
	}
	got, err := mem.ReadCString(buf)
	if err != nil {
		t.Fatalf("ReadCString: %v", err)
	}
	if got != "Z\n" {
		t.Fatalf("fgets_unlocked buffer=%q, want Z\\n", got)
	}
}

func TestFflushAcceptsHostHandleAndNull(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	stdout, ok := reg.LookupVariable("stdout", mem)
	if !ok {
		t.Fatal("missing stdout extern variable")
	}
	for _, name := range []string{"fflush", "fflush_unlocked"} {
		fn, ok := reg.Lookup(name)
		if !ok {
			t.Fatalf("missing %s extern", name)
		}
		for _, arg := range []Value{PtrValue(stdout), PtrValue(0)} {
			ret, exit, err := fn(context.Background(), &ExternContext{Memory: mem}, []Value{arg})
			if err != nil || exit != nil {
				t.Fatalf("%s(%#v) ret=%#v exit=%#v err=%v", name, arg, ret, exit, err)
			}
			if ret.Type != bytecode.TypeI32 || ret.Int != 0 {
				t.Fatalf("%s(%#v) = %#v, want i32 0", name, arg, ret)
			}
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

func TestFcloseRejectsLaterStreamUse(t *testing.T) {
	var out bytes.Buffer
	reg := DefaultExternRegistry(&out, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	stdout, ok := reg.LookupVariable("stdout", mem)
	if !ok {
		t.Fatal("missing stdout extern variable")
	}
	fcloseFn, ok := reg.Lookup("fclose")
	if !ok {
		t.Fatal("missing fclose extern")
	}
	fputcFn, ok := reg.Lookup("fputc")
	if !ok {
		t.Fatal("missing fputc extern")
	}
	ret, exit, err := fcloseFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(stdout)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeI32 || ret.Int != 0 {
		t.Fatalf("fclose ret=%#v exit=%#v err=%v, want i32 0", ret, exit, err)
	}
	_, _, err = fputcFn(context.Background(), &ExternContext{Memory: mem}, []Value{
		IntValue(bytecode.TypeI32, 'x'),
		PtrValue(stdout),
	})
	if err == nil || !strings.Contains(err.Error(), "stream handle") {
		t.Fatalf("fputc after fclose err = %v, want stream handle error", err)
	}
	if out.String() != "" {
		t.Fatalf("stdout after closed fputc = %q, want empty", out.String())
	}
}

func TestStreamWriteErrorSetsErrorIndicator(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	reg.AddFile("readonly.txt", []byte("AB"))
	mem := NewMemory(bytecode.DefaultTarget())
	path := mustAllocBytes(t, mem, "stdio:error-path", []byte("readonly.txt\x00"), true, blockString)
	mode := mustAllocBytes(t, mem, "stdio:error-mode", []byte("r\x00"), true, blockString)

	fopenFn, ok := reg.Lookup("fopen")
	if !ok {
		t.Fatal("missing fopen extern")
	}
	ret, exit, err := fopenFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(path), PtrValue(mode)})
	if err != nil || exit != nil || ret.Type != bytecode.TypePtr || ret.Int == 0 {
		t.Fatalf("fopen ret=%#v exit=%#v err=%v, want file handle", ret, exit, err)
	}
	file := ret.Int
	fputcFn, ok := reg.Lookup("fputc")
	if !ok {
		t.Fatal("missing fputc extern")
	}
	ret, exit, err = fputcFn(context.Background(), &ExternContext{Memory: mem}, []Value{IntValue(bytecode.TypeI32, 'X'), PtrValue(file)})
	if err != nil || exit != nil {
		t.Fatalf("fputc ret=%#v exit=%#v err=%v, want EOF without trap", ret, exit, err)
	}
	if ret.Type != bytecode.TypeI32 || int32(ret.Int) != -1 {
		t.Fatalf("fputc ret=%#v, want EOF", ret)
	}
	ferrorFn, ok := reg.Lookup("ferror")
	if !ok {
		t.Fatal("missing ferror extern")
	}
	ret, exit, err = ferrorFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(file)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeI32 || ret.Int != 1 {
		t.Fatalf("ferror ret=%#v exit=%#v err=%v, want 1", ret, exit, err)
	}
	clearerrFn, ok := reg.Lookup("clearerr")
	if !ok {
		t.Fatal("missing clearerr extern")
	}
	if _, exit, err := clearerrFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(file)}); err != nil || exit != nil {
		t.Fatalf("clearerr exit=%#v err=%v", exit, err)
	}
	ret, exit, err = ferrorFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(file)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeI32 || ret.Int != 0 {
		t.Fatalf("ferror after clearerr ret=%#v exit=%#v err=%v, want 0", ret, exit, err)
	}
}

func TestStreamReadErrorSetsErrorIndicator(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	path := mustAllocBytes(t, mem, "stdio:read-error-path", []byte("writeonly.txt\x00"), true, blockString)
	mode := mustAllocBytes(t, mem, "stdio:read-error-mode", []byte("w\x00"), true, blockString)

	fopenFn, ok := reg.Lookup("fopen")
	if !ok {
		t.Fatal("missing fopen extern")
	}
	ret, exit, err := fopenFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(path), PtrValue(mode)})
	if err != nil || exit != nil || ret.Type != bytecode.TypePtr || ret.Int == 0 {
		t.Fatalf("fopen ret=%#v exit=%#v err=%v, want file handle", ret, exit, err)
	}
	file := ret.Int
	fgetcFn, ok := reg.Lookup("fgetc")
	if !ok {
		t.Fatal("missing fgetc extern")
	}
	ret, exit, err = fgetcFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(file)})
	if err != nil || exit != nil {
		t.Fatalf("fgetc ret=%#v exit=%#v err=%v, want EOF without trap", ret, exit, err)
	}
	if ret.Type != bytecode.TypeI32 || int32(ret.Int) != -1 {
		t.Fatalf("fgetc ret=%#v, want EOF", ret)
	}
	ferrorFn, ok := reg.Lookup("ferror")
	if !ok {
		t.Fatal("missing ferror extern")
	}
	ret, exit, err = ferrorFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(file)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeI32 || ret.Int != 1 {
		t.Fatalf("ferror ret=%#v exit=%#v err=%v, want 1", ret, exit, err)
	}
	feofFn, ok := reg.Lookup("feof")
	if !ok {
		t.Fatal("missing feof extern")
	}
	ret, exit, err = feofFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(file)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeI32 || ret.Int != 0 {
		t.Fatalf("feof ret=%#v exit=%#v err=%v, want 0", ret, exit, err)
	}
}

func TestFilenoReturnsStandardStreamDescriptors(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	streams := []struct {
		variable string
		fn       string
		fd       int64
	}{
		{variable: "stdin", fn: "fileno", fd: 0},
		{variable: "stdout", fn: "fileno", fd: 1},
		{variable: "stderr", fn: "fileno_unlocked", fd: 2},
	}
	for _, tt := range streams {
		stream, ok := reg.LookupVariable(tt.variable, mem)
		if !ok {
			t.Fatalf("missing %s extern variable", tt.variable)
		}
		fn, ok := reg.Lookup(tt.fn)
		if !ok {
			t.Fatalf("missing %s extern", tt.fn)
		}
		ret, exit, err := fn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(stream)})
		if err != nil || exit != nil {
			t.Fatalf("%s(%s) ret=%#v exit=%#v err=%v", tt.fn, tt.variable, ret, exit, err)
		}
		if ret.Type != bytecode.TypeI32 || ret.Int != uint64(tt.fd) {
			t.Fatalf("%s(%s) = %#v, want i32 %d", tt.fn, tt.variable, ret, tt.fd)
		}
	}
}

func TestStdioBufferControls(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	buf := mustAlloc(t, mem, "stdio:setvbuf-buffer", 8, 1, false, blockLocal)
	stdout, ok := reg.LookupVariable("stdout", mem)
	if !ok {
		t.Fatal("missing stdout extern variable")
	}
	setbufFn, ok := reg.Lookup("setbuf")
	if !ok {
		t.Fatal("missing setbuf extern")
	}
	if ret, exit, err := setbufFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(stdout), PtrValue(0)}); err != nil || exit != nil || ret.Type != 0 {
		t.Fatalf("setbuf ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret, exit, err := setbufFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(stdout), PtrValue(buf)}); err != nil || exit != nil || ret.Type != 0 {
		t.Fatalf("setbuf buffer ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	setvbufFn, ok := reg.Lookup("setvbuf")
	if !ok {
		t.Fatal("missing setvbuf extern")
	}
	for _, mode := range []int64{0, 1, 2} {
		ret, exit, err := setvbufFn(context.Background(), &ExternContext{Memory: mem}, []Value{
			PtrValue(stdout),
			PtrValue(0),
			IntValue(bytecode.TypeI32, mode),
			UIntValue(bytecode.TypeU64, 0),
		})
		if err != nil || exit != nil {
			t.Fatalf("setvbuf mode %d ret=%#v exit=%#v err=%v", mode, ret, exit, err)
		}
		if ret.Type != bytecode.TypeI32 || ret.Int != 0 {
			t.Fatalf("setvbuf mode %d ret=%#v, want i32 0", mode, ret)
		}
	}
	ret, exit, err := setvbufFn(context.Background(), &ExternContext{Memory: mem}, []Value{
		PtrValue(stdout),
		PtrValue(0),
		IntValue(bytecode.TypeI32, 99),
		UIntValue(bytecode.TypeU64, 0),
	})
	if err != nil || exit != nil {
		t.Fatalf("setvbuf invalid mode ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeI32 || ret.Int == 0 {
		t.Fatalf("setvbuf invalid mode ret=%#v, want nonzero i32", ret)
	}
	ret, exit, err = setvbufFn(context.Background(), &ExternContext{Memory: mem}, []Value{
		PtrValue(stdout),
		PtrValue(0),
		IntValue(bytecode.TypeI32, 0),
		UIntValue(bytecode.TypeU64, 8),
	})
	if err != nil || exit != nil || ret.Type != bytecode.TypeI32 || ret.Int != 0 {
		t.Fatalf("setvbuf null buffer nonzero size ret=%#v exit=%#v err=%v, want 0", ret, exit, err)
	}
	ret, exit, err = setvbufFn(context.Background(), &ExternContext{Memory: mem}, []Value{
		PtrValue(stdout),
		PtrValue(buf),
		IntValue(bytecode.TypeI32, 0),
		UIntValue(bytecode.TypeU64, 8),
	})
	if err != nil || exit != nil || ret.Type != bytecode.TypeI32 || ret.Int != 0 {
		t.Fatalf("setvbuf valid buffer ret=%#v exit=%#v err=%v, want 0", ret, exit, err)
	}
	ret, exit, err = setvbufFn(context.Background(), &ExternContext{Memory: mem}, []Value{
		PtrValue(stdout),
		PtrValue(buf + 4),
		IntValue(bytecode.TypeI32, 0),
		UIntValue(bytecode.TypeU64, 8),
	})
	if err == nil || !strings.Contains(err.Error(), "invalid memory access") || exit != nil {
		t.Fatalf("setvbuf invalid buffer ret=%#v exit=%#v err=%v, want invalid memory access", ret, exit, err)
	}
	tmpfileFn, ok := reg.Lookup("tmpfile")
	if !ok {
		t.Fatal("missing tmpfile extern")
	}
	fcloseFn, ok := reg.Lookup("fclose")
	if !ok {
		t.Fatal("missing fclose extern")
	}
	ret, exit, err = tmpfileFn(context.Background(), &ExternContext{Memory: mem}, nil)
	if err != nil || exit != nil || ret.Type != bytecode.TypePtr || ret.Int == 0 {
		t.Fatalf("tmpfile ret=%#v exit=%#v err=%v, want file handle", ret, exit, err)
	}
	closed := ret.Int
	if _, exit, err = fcloseFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(closed)}); err != nil || exit != nil {
		t.Fatalf("fclose exit=%#v err=%v", exit, err)
	}
	ret, exit, err = setvbufFn(context.Background(), &ExternContext{Memory: mem}, []Value{
		PtrValue(closed),
		PtrValue(0),
		IntValue(bytecode.TypeI32, 0),
		UIntValue(bytecode.TypeU64, 0),
	})
	if err == nil || !strings.Contains(err.Error(), "unknown stream handle") || exit != nil {
		t.Fatalf("setvbuf closed stream ret=%#v exit=%#v err=%v, want unknown stream handle", ret, exit, err)
	}
}

func TestStdioStreamLockControls(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	stdout, ok := reg.LookupVariable("stdout", mem)
	if !ok {
		t.Fatal("missing stdout extern variable")
	}
	for _, name := range []string{"flockfile", "funlockfile"} {
		fn, ok := reg.Lookup(name)
		if !ok {
			t.Fatalf("missing %s extern", name)
		}
		ret, exit, err := fn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(stdout)})
		if err != nil || exit != nil || ret.Type != 0 {
			t.Fatalf("%s ret=%#v exit=%#v err=%v", name, ret, exit, err)
		}
	}
	fn, ok := reg.Lookup("ftrylockfile")
	if !ok {
		t.Fatal("missing ftrylockfile extern")
	}
	ret, exit, err := fn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(stdout)})
	if err != nil || exit != nil {
		t.Fatalf("ftrylockfile ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeI32 || ret.Int != 0 {
		t.Fatalf("ftrylockfile ret=%#v, want i32 0", ret)
	}
}

func TestStdlibAbsExterns(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	tests := []struct {
		name string
		arg  Value
		typ  bytecode.ValueType
		want int64
	}{
		{name: "abs", arg: IntValue(bytecode.TypeI32, -7), typ: bytecode.TypeI32, want: 7},
		{name: "abs", arg: IntValue(bytecode.TypeI32, 5), typ: bytecode.TypeI32, want: 5},
		{name: "labs", arg: IntValue(bytecode.TypeI64, -9), typ: bytecode.TypeI64, want: 9},
		{name: "llabs", arg: IntValue(bytecode.TypeI64, -11), typ: bytecode.TypeI64, want: 11},
	}
	for _, tt := range tests {
		fn, ok := reg.Lookup(tt.name)
		if !ok {
			t.Fatalf("missing %s extern", tt.name)
		}
		ret, exit, err := fn(context.Background(), nil, []Value{tt.arg})
		if err != nil || exit != nil {
			t.Fatalf("%s ret=%#v exit=%#v err=%v", tt.name, ret, exit, err)
		}
		if ret.Type != tt.typ || int64(ret.Int) != tt.want {
			t.Fatalf("%s ret=%#v, want %s %d", tt.name, ret, tt.typ, tt.want)
		}
	}
}

func TestStdlibAtoiExterns(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	tests := []struct {
		name  string
		input string
		typ   bytecode.ValueType
		want  int64
	}{
		{name: "atoi", input: " \t-42tail", typ: bytecode.TypeI32, want: -42},
		{name: "atoi", input: "+17", typ: bytecode.TypeI32, want: 17},
		{name: "atol", input: "\n12345x", typ: bytecode.TypeI64, want: 12345},
		{name: "atoll", input: "  -9876543210", typ: bytecode.TypeI64, want: -9876543210},
		{name: "atoll", input: "nondigit", typ: bytecode.TypeI64, want: 0},
	}
	for _, tt := range tests {
		fn, ok := reg.Lookup(tt.name)
		if !ok {
			t.Fatalf("missing %s extern", tt.name)
		}
		addr := mustAllocBytes(t, mem, "stdlib:"+tt.name+":"+tt.input, []byte(tt.input+"\x00"), true, blockString)
		ret, exit, err := fn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(addr)})
		if err != nil || exit != nil {
			t.Fatalf("%s(%q) ret=%#v exit=%#v err=%v", tt.name, tt.input, ret, exit, err)
		}
		if ret.Type != tt.typ || signedInt(ret) != tt.want {
			t.Fatalf("%s(%q) ret=%#v, want %s %d", tt.name, tt.input, ret, tt.typ, tt.want)
		}
	}
}

func TestStdlibMultibyteExterns(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	target := bytecode.DefaultTarget()
	mem := NewMemory(target)
	text := mustAllocBytes(t, mem, "stdlib:mb:text", []byte("Az\x00"), true, blockString)
	high := mustAllocBytes(t, mem, "stdlib:mb:high", []byte{0x80, 0}, true, blockString)
	wide, err := mem.TryAlloc("stdlib:mb:wide", 16, 4, false, blockLocal)
	if err != nil {
		t.Fatalf("alloc wide: %v", err)
	}
	out := mustAllocBytes(t, mem, "stdlib:mb:out", make([]byte, 4), false, blockLocal)

	mblenFn, ok := reg.Lookup("mblen")
	if !ok {
		t.Fatal("missing mblen extern")
	}
	ret, exit, err := mblenFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(text), UIntValue(bytecode.TypeU64, 3)})
	if err != nil || exit != nil {
		t.Fatalf("mblen ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeI32 || ret.Int != 1 {
		t.Fatalf("mblen ret=%#v, want i32 1", ret)
	}
	ret, exit, err = mblenFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(text + 2), UIntValue(bytecode.TypeU64, 1)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeI32 || ret.Int != 0 {
		t.Fatalf("mblen nul ret=%#v exit=%#v err=%v, want i32 0", ret, exit, err)
	}
	ret, exit, err = mblenFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(0), UIntValue(bytecode.TypeU64, 0)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeI32 || ret.Int != 0 {
		t.Fatalf("mblen reset ret=%#v exit=%#v err=%v, want i32 0", ret, exit, err)
	}
	ret, exit, err = mblenFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(text), UIntValue(bytecode.TypeU64, 0)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeI32 || signedInt(ret) != -1 {
		t.Fatalf("mblen zero n ret=%#v exit=%#v err=%v, want i32 -1", ret, exit, err)
	}
	ret, exit, err = mblenFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(high), UIntValue(bytecode.TypeU64, 1)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeI32 || signedInt(ret) != -1 {
		t.Fatalf("mblen high-bit ret=%#v exit=%#v err=%v, want i32 -1", ret, exit, err)
	}

	mbrlenFn, ok := reg.Lookup("mbrlen")
	if !ok {
		t.Fatal("missing mbrlen extern")
	}
	ret, exit, err = mbrlenFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(text), UIntValue(bytecode.TypeU64, 3), PtrValue(wide)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeU64 || ret.Int != 1 {
		t.Fatalf("mbrlen ret=%#v exit=%#v err=%v, want u64 1", ret, exit, err)
	}
	ret, exit, err = mbrlenFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(text + 2), UIntValue(bytecode.TypeU64, 1), PtrValue(wide)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeU64 || ret.Int != 0 {
		t.Fatalf("mbrlen nul ret=%#v exit=%#v err=%v, want u64 0", ret, exit, err)
	}
	ret, exit, err = mbrlenFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(0), UIntValue(bytecode.TypeU64, 0), PtrValue(wide)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeU64 || ret.Int != 0 {
		t.Fatalf("mbrlen reset ret=%#v exit=%#v err=%v, want u64 0", ret, exit, err)
	}
	ret, exit, err = mbrlenFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(text), UIntValue(bytecode.TypeU64, 0), PtrValue(wide)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeU64 || ret.Int != math.MaxUint64-1 {
		t.Fatalf("mbrlen zero n ret=%#v exit=%#v err=%v, want u64 -2", ret, exit, err)
	}
	ret, exit, err = mbrlenFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(high), UIntValue(bytecode.TypeU64, 1), PtrValue(wide)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeU64 || ret.Int != math.MaxUint64 {
		t.Fatalf("mbrlen high-bit ret=%#v exit=%#v err=%v, want u64 -1", ret, exit, err)
	}

	mbrtowcFn, ok := reg.Lookup("mbrtowc")
	if !ok {
		t.Fatal("missing mbrtowc extern")
	}
	ret, exit, err = mbrtowcFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(wide), PtrValue(text), UIntValue(bytecode.TypeU64, 3), PtrValue(wide + 8)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeU64 || ret.Int != 1 {
		t.Fatalf("mbrtowc ret=%#v exit=%#v err=%v, want u64 1", ret, exit, err)
	}
	got, err := mem.Load(wide, bytecode.TypeI32, 4)
	if err != nil || got.Int != 'A' {
		t.Fatalf("mbrtowc stored %#v err=%v, want 'A'", got, err)
	}
	ret, exit, err = mbrtowcFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(0), PtrValue(text), UIntValue(bytecode.TypeU64, 3), PtrValue(wide + 8)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeU64 || ret.Int != 1 {
		t.Fatalf("mbrtowc null dest ret=%#v exit=%#v err=%v, want u64 1", ret, exit, err)
	}
	ret, exit, err = mbrtowcFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(0), PtrValue(0), UIntValue(bytecode.TypeU64, 0), PtrValue(wide + 8)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeU64 || ret.Int != 0 {
		t.Fatalf("mbrtowc null source ret=%#v exit=%#v err=%v, want u64 0", ret, exit, err)
	}
	ret, exit, err = mbrtowcFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(wide), PtrValue(text), UIntValue(bytecode.TypeU64, 0), PtrValue(wide + 8)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeU64 || ret.Int != math.MaxUint64-1 {
		t.Fatalf("mbrtowc zero n ret=%#v exit=%#v err=%v, want u64 -2", ret, exit, err)
	}
	ret, exit, err = mbrtowcFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(wide), PtrValue(text + 2), UIntValue(bytecode.TypeU64, 1), PtrValue(wide + 8)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeU64 || ret.Int != 0 {
		t.Fatalf("mbrtowc nul ret=%#v exit=%#v err=%v, want u64 0", ret, exit, err)
	}
	got, err = mem.Load(wide, bytecode.TypeI32, 4)
	if err != nil || got.Int != 0 {
		t.Fatalf("mbrtowc nul stored %#v err=%v, want 0", got, err)
	}
	if err := mem.Store(wide, bytecode.TypeI32, 4, IntValue(bytecode.TypeI32, 'Q')); err != nil {
		t.Fatalf("store mbrtowc sentinel: %v", err)
	}
	ret, exit, err = mbrtowcFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(wide), PtrValue(high), UIntValue(bytecode.TypeU64, 1), PtrValue(wide + 8)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeU64 || ret.Int != math.MaxUint64 {
		t.Fatalf("mbrtowc high-bit ret=%#v exit=%#v err=%v, want u64 -1", ret, exit, err)
	}
	got, err = mem.Load(wide, bytecode.TypeI32, 4)
	if err != nil || got.Int != 'Q' {
		t.Fatalf("mbrtowc high-bit stored %#v err=%v, want sentinel", got, err)
	}

	wcrtombFn, ok := reg.Lookup("wcrtomb")
	if !ok {
		t.Fatal("missing wcrtomb extern")
	}
	ret, exit, err = wcrtombFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(out), IntValue(bytecode.TypeI32, 'R'), PtrValue(wide + 8)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeU64 || ret.Int != 1 {
		t.Fatalf("wcrtomb ret=%#v exit=%#v err=%v, want u64 1", ret, exit, err)
	}
	storedByte, err := mem.Load(out, bytecode.TypeI8, 1)
	if err != nil || storedByte.Int != 'R' {
		t.Fatalf("wcrtomb stored %#v err=%v, want 'R'", storedByte, err)
	}
	ret, exit, err = wcrtombFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(out), IntValue(bytecode.TypeI32, 0), PtrValue(wide + 8)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeU64 || ret.Int != 1 {
		t.Fatalf("wcrtomb nul ret=%#v exit=%#v err=%v, want u64 1", ret, exit, err)
	}
	storedByte, err = mem.Load(out, bytecode.TypeI8, 1)
	if err != nil || storedByte.Int != 0 {
		t.Fatalf("wcrtomb nul stored %#v err=%v, want 0", storedByte, err)
	}
	ret, exit, err = wcrtombFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(0), IntValue(bytecode.TypeI32, 'R'), PtrValue(wide + 8)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeU64 || ret.Int != 1 {
		t.Fatalf("wcrtomb reset ret=%#v exit=%#v err=%v, want u64 1", ret, exit, err)
	}
	if err := mem.Store(out, bytecode.TypeI8, 1, IntValue(bytecode.TypeI8, '!')); err != nil {
		t.Fatalf("store wcrtomb sentinel: %v", err)
	}
	ret, exit, err = wcrtombFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(out), IntValue(bytecode.TypeI32, 0x80), PtrValue(wide + 8)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeU64 || ret.Int != math.MaxUint64 {
		t.Fatalf("wcrtomb high wchar ret=%#v exit=%#v err=%v, want u64 -1", ret, exit, err)
	}
	storedByte, err = mem.Load(out, bytecode.TypeI8, 1)
	if err != nil || storedByte.Int != '!' {
		t.Fatalf("wcrtomb high wchar stored %#v err=%v, want sentinel", storedByte, err)
	}

	mbtowcFn, ok := reg.Lookup("mbtowc")
	if !ok {
		t.Fatal("missing mbtowc extern")
	}
	ret, exit, err = mbtowcFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(wide), PtrValue(text), UIntValue(bytecode.TypeU64, 3)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeI32 || ret.Int != 1 {
		t.Fatalf("mbtowc ret=%#v exit=%#v err=%v, want i32 1", ret, exit, err)
	}
	got, err = mem.Load(wide, bytecode.TypeI32, 4)
	if err != nil || got.Int != 'A' {
		t.Fatalf("mbtowc stored %#v err=%v, want 'A'", got, err)
	}
	ret, exit, err = mbtowcFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(0), PtrValue(text), UIntValue(bytecode.TypeU64, 3)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeI32 || ret.Int != 1 {
		t.Fatalf("mbtowc null dest ret=%#v exit=%#v err=%v, want i32 1", ret, exit, err)
	}
	ret, exit, err = mbtowcFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(0), PtrValue(0), UIntValue(bytecode.TypeU64, 0)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeI32 || ret.Int != 0 {
		t.Fatalf("mbtowc null source ret=%#v exit=%#v err=%v, want i32 0", ret, exit, err)
	}
	ret, exit, err = mbtowcFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(wide), PtrValue(text), UIntValue(bytecode.TypeU64, 0)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeI32 || signedInt(ret) != -1 {
		t.Fatalf("mbtowc zero n ret=%#v exit=%#v err=%v, want i32 -1", ret, exit, err)
	}
	ret, exit, err = mbtowcFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(wide), PtrValue(text + 2), UIntValue(bytecode.TypeU64, 1)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeI32 || ret.Int != 0 {
		t.Fatalf("mbtowc nul ret=%#v exit=%#v err=%v, want i32 0", ret, exit, err)
	}
	got, err = mem.Load(wide, bytecode.TypeI32, 4)
	if err != nil || got.Int != 0 {
		t.Fatalf("mbtowc nul stored %#v err=%v, want 0", got, err)
	}
	if err := mem.Store(wide, bytecode.TypeI32, 4, IntValue(bytecode.TypeI32, 0x1234)); err != nil {
		t.Fatalf("store wide sentinel: %v", err)
	}
	ret, exit, err = mbtowcFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(wide), PtrValue(high), UIntValue(bytecode.TypeU64, 1)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeI32 || signedInt(ret) != -1 {
		t.Fatalf("mbtowc high-bit ret=%#v exit=%#v err=%v, want i32 -1", ret, exit, err)
	}
	got, err = mem.Load(wide, bytecode.TypeI32, 4)
	if err != nil || got.Int != 0x1234 {
		t.Fatalf("mbtowc high-bit stored %#v err=%v, want sentinel", got, err)
	}

	wctombFn, ok := reg.Lookup("wctomb")
	if !ok {
		t.Fatal("missing wctomb extern")
	}
	ret, exit, err = wctombFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(out), IntValue(bytecode.TypeI32, 'Q')})
	if err != nil || exit != nil || ret.Type != bytecode.TypeI32 || ret.Int != 1 {
		t.Fatalf("wctomb ret=%#v exit=%#v err=%v, want i32 1", ret, exit, err)
	}
	ch, err := readMemoryByte(mem, out)
	if err != nil || ch != 'Q' {
		t.Fatalf("wctomb wrote %q err=%v, want Q", ch, err)
	}
	ret, exit, err = wctombFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(out), IntValue(bytecode.TypeI32, 0)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeI32 || ret.Int != 1 {
		t.Fatalf("wctomb nul ret=%#v exit=%#v err=%v, want i32 1", ret, exit, err)
	}
	ch, err = readMemoryByte(mem, out)
	if err != nil || ch != 0 {
		t.Fatalf("wctomb nul wrote %#x err=%v, want 0", ch, err)
	}
	if err := writeMemoryByte(mem, out, 'X'); err != nil {
		t.Fatalf("store wctomb sentinel: %v", err)
	}
	ret, exit, err = wctombFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(out), IntValue(bytecode.TypeI32, 0x80)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeI32 || signedInt(ret) != -1 {
		t.Fatalf("wctomb high wchar ret=%#v exit=%#v err=%v, want i32 -1", ret, exit, err)
	}
	ch, err = readMemoryByte(mem, out)
	if err != nil || ch != 'X' {
		t.Fatalf("wctomb high wchar wrote %q err=%v, want sentinel", ch, err)
	}
	ret, exit, err = wctombFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(0), IntValue(bytecode.TypeI32, 0)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeI32 || ret.Int != 0 {
		t.Fatalf("wctomb reset ret=%#v exit=%#v err=%v, want i32 0", ret, exit, err)
	}

	mbstowcsFn, ok := reg.Lookup("mbstowcs")
	if !ok {
		t.Fatal("missing mbstowcs extern")
	}
	ret, exit, err = mbstowcsFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(wide), PtrValue(text), UIntValue(bytecode.TypeU64, 4)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeU64 || ret.Int != 2 {
		t.Fatalf("mbstowcs ret=%#v exit=%#v err=%v, want u64 2", ret, exit, err)
	}
	got, err = mem.Load(wide+4, bytecode.TypeI32, 4)
	if err != nil || got.Int != 'z' {
		t.Fatalf("mbstowcs second stored %#v err=%v, want 'z'", got, err)
	}
	got, err = mem.Load(wide+8, bytecode.TypeI32, 4)
	if err != nil || got.Int != 0 {
		t.Fatalf("mbstowcs terminator stored %#v err=%v, want 0", got, err)
	}
	if err := mem.Store(wide+4, bytecode.TypeI32, 4, IntValue(bytecode.TypeI32, 'X')); err != nil {
		t.Fatalf("store mbstowcs trunc sentinel: %v", err)
	}
	ret, exit, err = mbstowcsFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(wide), PtrValue(text), UIntValue(bytecode.TypeU64, 1)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeU64 || ret.Int != 1 {
		t.Fatalf("mbstowcs trunc ret=%#v exit=%#v err=%v, want u64 1", ret, exit, err)
	}
	got, err = mem.Load(wide+4, bytecode.TypeI32, 4)
	if err != nil || got.Int != 'X' {
		t.Fatalf("mbstowcs trunc sentinel %#v err=%v, want unchanged", got, err)
	}
	ret, exit, err = mbstowcsFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(0), PtrValue(text), UIntValue(bytecode.TypeU64, 0)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeU64 || ret.Int != 2 {
		t.Fatalf("mbstowcs length ret=%#v exit=%#v err=%v, want u64 2", ret, exit, err)
	}
	ret, exit, err = mbstowcsFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(wide), PtrValue(high), UIntValue(bytecode.TypeU64, 4)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeU64 || signedInt(ret) != -1 {
		t.Fatalf("mbstowcs high-bit ret=%#v exit=%#v err=%v, want u64 -1", ret, exit, err)
	}

	srcSlot, err := mem.TryAlloc("stdlib:mb:srcslot", target.PointerSize, target.PointerAlign, false, blockLocal)
	if err != nil {
		t.Fatalf("alloc source slot: %v", err)
	}
	mbsrtowcsFn, ok := reg.Lookup("mbsrtowcs")
	if !ok {
		t.Fatal("missing mbsrtowcs extern")
	}
	if err := mem.WritePointer(srcSlot, text); err != nil {
		t.Fatalf("store mbsrtowcs source: %v", err)
	}
	ret, exit, err = mbsrtowcsFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(wide), PtrValue(srcSlot), UIntValue(bytecode.TypeU64, 4), PtrValue(wide + 8)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeU64 || ret.Int != 2 {
		t.Fatalf("mbsrtowcs ret=%#v exit=%#v err=%v, want u64 2", ret, exit, err)
	}
	got, err = mem.Load(wide+4, bytecode.TypeI32, 4)
	if err != nil || got.Int != 'z' {
		t.Fatalf("mbsrtowcs second stored %#v err=%v, want 'z'", got, err)
	}
	loadedPtr, err := mem.Load(srcSlot, bytecode.TypePtr, target.PointerAlign)
	if err != nil || loadedPtr.Int != 0 {
		t.Fatalf("mbsrtowcs source after complete=%#v err=%v, want null", loadedPtr, err)
	}
	if err := mem.WritePointer(srcSlot, text); err != nil {
		t.Fatalf("store mbsrtowcs trunc source: %v", err)
	}
	if err := mem.Store(wide+4, bytecode.TypeI32, 4, IntValue(bytecode.TypeI32, 'X')); err != nil {
		t.Fatalf("store mbsrtowcs trunc sentinel: %v", err)
	}
	ret, exit, err = mbsrtowcsFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(wide), PtrValue(srcSlot), UIntValue(bytecode.TypeU64, 1), PtrValue(wide + 8)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeU64 || ret.Int != 1 {
		t.Fatalf("mbsrtowcs trunc ret=%#v exit=%#v err=%v, want u64 1", ret, exit, err)
	}
	loadedPtr, err = mem.Load(srcSlot, bytecode.TypePtr, target.PointerAlign)
	if err != nil || loadedPtr.Int != text+1 {
		t.Fatalf("mbsrtowcs source after trunc=%#v err=%v, want %#x", loadedPtr, err, text+1)
	}
	got, err = mem.Load(wide+4, bytecode.TypeI32, 4)
	if err != nil || got.Int != 'X' {
		t.Fatalf("mbsrtowcs trunc sentinel %#v err=%v, want unchanged", got, err)
	}
	if err := mem.WritePointer(srcSlot, text); err != nil {
		t.Fatalf("store mbsrtowcs length source: %v", err)
	}
	ret, exit, err = mbsrtowcsFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(0), PtrValue(srcSlot), UIntValue(bytecode.TypeU64, 0), PtrValue(wide + 8)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeU64 || ret.Int != 2 {
		t.Fatalf("mbsrtowcs length ret=%#v exit=%#v err=%v, want u64 2", ret, exit, err)
	}
	loadedPtr, err = mem.Load(srcSlot, bytecode.TypePtr, target.PointerAlign)
	if err != nil || loadedPtr.Int != text {
		t.Fatalf("mbsrtowcs source after length=%#v err=%v, want unchanged", loadedPtr, err)
	}
	if err := mem.WritePointer(srcSlot, high); err != nil {
		t.Fatalf("store mbsrtowcs high source: %v", err)
	}
	ret, exit, err = mbsrtowcsFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(wide), PtrValue(srcSlot), UIntValue(bytecode.TypeU64, 4), PtrValue(wide + 8)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeU64 || ret.Int != math.MaxUint64 {
		t.Fatalf("mbsrtowcs high-bit ret=%#v exit=%#v err=%v, want u64 -1", ret, exit, err)
	}

	if err := mem.Store(wide, bytecode.TypeI32, 4, IntValue(bytecode.TypeI32, 'o')); err != nil {
		t.Fatalf("store wide[0]: %v", err)
	}
	if err := mem.Store(wide+4, bytecode.TypeI32, 4, IntValue(bytecode.TypeI32, 'k')); err != nil {
		t.Fatalf("store wide[1]: %v", err)
	}
	if err := mem.Store(wide+8, bytecode.TypeI32, 4, IntValue(bytecode.TypeI32, 0)); err != nil {
		t.Fatalf("store wide[2]: %v", err)
	}
	wcstombsFn, ok := reg.Lookup("wcstombs")
	if !ok {
		t.Fatal("missing wcstombs extern")
	}
	ret, exit, err = wcstombsFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(out), PtrValue(wide), UIntValue(bytecode.TypeU64, 4)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeU64 || ret.Int != 2 {
		t.Fatalf("wcstombs ret=%#v exit=%#v err=%v, want u64 2", ret, exit, err)
	}
	gotString, err := mem.ReadCString(out)
	if err != nil || gotString != "ok" {
		t.Fatalf("wcstombs string %q err=%v, want ok", gotString, err)
	}
	if err := writeMemoryByte(mem, out+1, 'X'); err != nil {
		t.Fatalf("store wcstombs trunc sentinel: %v", err)
	}
	ret, exit, err = wcstombsFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(out), PtrValue(wide), UIntValue(bytecode.TypeU64, 1)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeU64 || ret.Int != 1 {
		t.Fatalf("wcstombs trunc ret=%#v exit=%#v err=%v, want u64 1", ret, exit, err)
	}
	ch, err = readMemoryByte(mem, out+1)
	if err != nil || ch != 'X' {
		t.Fatalf("wcstombs trunc sentinel %q err=%v, want unchanged", ch, err)
	}
	ret, exit, err = wcstombsFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(0), PtrValue(wide), UIntValue(bytecode.TypeU64, 0)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeU64 || ret.Int != 2 {
		t.Fatalf("wcstombs length ret=%#v exit=%#v err=%v, want u64 2", ret, exit, err)
	}
	if err := mem.Store(wide+4, bytecode.TypeI32, 4, IntValue(bytecode.TypeI32, 0x80)); err != nil {
		t.Fatalf("store high wide: %v", err)
	}
	if err := writeMemoryByte(mem, out, 'X'); err != nil {
		t.Fatalf("store wcstombs high sentinel: %v", err)
	}
	ret, exit, err = wcstombsFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(out), PtrValue(wide), UIntValue(bytecode.TypeU64, 4)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeU64 || signedInt(ret) != -1 {
		t.Fatalf("wcstombs high wchar ret=%#v exit=%#v err=%v, want u64 -1", ret, exit, err)
	}
	ch, err = readMemoryByte(mem, out)
	if err != nil || ch != 'X' {
		t.Fatalf("wcstombs high wchar wrote %q err=%v, want sentinel", ch, err)
	}

	if err := mem.Store(wide, bytecode.TypeI32, 4, IntValue(bytecode.TypeI32, 'o')); err != nil {
		t.Fatalf("store wcsrtombs wide[0]: %v", err)
	}
	if err := mem.Store(wide+4, bytecode.TypeI32, 4, IntValue(bytecode.TypeI32, 'k')); err != nil {
		t.Fatalf("store wcsrtombs wide[1]: %v", err)
	}
	if err := mem.Store(wide+8, bytecode.TypeI32, 4, IntValue(bytecode.TypeI32, 0)); err != nil {
		t.Fatalf("store wcsrtombs wide[2]: %v", err)
	}
	wcsrtombsFn, ok := reg.Lookup("wcsrtombs")
	if !ok {
		t.Fatal("missing wcsrtombs extern")
	}
	if err := mem.WritePointer(srcSlot, wide); err != nil {
		t.Fatalf("store wcsrtombs source: %v", err)
	}
	ret, exit, err = wcsrtombsFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(out), PtrValue(srcSlot), UIntValue(bytecode.TypeU64, 4), PtrValue(wide + 12)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeU64 || ret.Int != 2 {
		t.Fatalf("wcsrtombs ret=%#v exit=%#v err=%v, want u64 2", ret, exit, err)
	}
	gotString, err = mem.ReadCString(out)
	if err != nil || gotString != "ok" {
		t.Fatalf("wcsrtombs string %q err=%v, want ok", gotString, err)
	}
	loadedPtr, err = mem.Load(srcSlot, bytecode.TypePtr, target.PointerAlign)
	if err != nil || loadedPtr.Int != 0 {
		t.Fatalf("wcsrtombs source after complete=%#v err=%v, want null", loadedPtr, err)
	}
	if err := mem.WritePointer(srcSlot, wide); err != nil {
		t.Fatalf("store wcsrtombs trunc source: %v", err)
	}
	if err := writeMemoryByte(mem, out+1, 'X'); err != nil {
		t.Fatalf("store wcsrtombs trunc sentinel: %v", err)
	}
	ret, exit, err = wcsrtombsFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(out), PtrValue(srcSlot), UIntValue(bytecode.TypeU64, 1), PtrValue(wide + 12)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeU64 || ret.Int != 1 {
		t.Fatalf("wcsrtombs trunc ret=%#v exit=%#v err=%v, want u64 1", ret, exit, err)
	}
	loadedPtr, err = mem.Load(srcSlot, bytecode.TypePtr, target.PointerAlign)
	if err != nil || loadedPtr.Int != wide+4 {
		t.Fatalf("wcsrtombs source after trunc=%#v err=%v, want %#x", loadedPtr, err, wide+4)
	}
	ch, err = readMemoryByte(mem, out+1)
	if err != nil || ch != 'X' {
		t.Fatalf("wcsrtombs trunc sentinel %q err=%v, want unchanged", ch, err)
	}
	if err := mem.WritePointer(srcSlot, wide); err != nil {
		t.Fatalf("store wcsrtombs length source: %v", err)
	}
	ret, exit, err = wcsrtombsFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(0), PtrValue(srcSlot), UIntValue(bytecode.TypeU64, 0), PtrValue(wide + 12)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeU64 || ret.Int != 2 {
		t.Fatalf("wcsrtombs length ret=%#v exit=%#v err=%v, want u64 2", ret, exit, err)
	}
	loadedPtr, err = mem.Load(srcSlot, bytecode.TypePtr, target.PointerAlign)
	if err != nil || loadedPtr.Int != wide {
		t.Fatalf("wcsrtombs source after length=%#v err=%v, want unchanged", loadedPtr, err)
	}
	if err := mem.Store(wide+4, bytecode.TypeI32, 4, IntValue(bytecode.TypeI32, 0x80)); err != nil {
		t.Fatalf("store wcsrtombs high wide: %v", err)
	}
	if err := mem.WritePointer(srcSlot, wide); err != nil {
		t.Fatalf("store wcsrtombs high source: %v", err)
	}
	ret, exit, err = wcsrtombsFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(out), PtrValue(srcSlot), UIntValue(bytecode.TypeU64, 4), PtrValue(wide + 12)})
	if err != nil || exit != nil || ret.Type != bytecode.TypeU64 || ret.Int != math.MaxUint64 {
		t.Fatalf("wcsrtombs high wchar ret=%#v exit=%#v err=%v, want u64 -1", ret, exit, err)
	}
}

func TestStdlibStrtolExterns(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	target := bytecode.DefaultTarget()
	mem := NewMemory(target)
	endptr, err := mem.TryAlloc("strtol:endptr", target.PointerSize, target.PointerAlign, false, blockGlobal)
	if err != nil {
		t.Fatalf("alloc endptr: %v", err)
	}

	strtolFn, ok := reg.Lookup("strtol")
	if !ok {
		t.Fatal("missing strtol extern")
	}
	signedText := mustAllocBytes(t, mem, "strtol:signed", []byte(" \t-0x2a tail\x00"), true, blockString)
	ret, exit, err := strtolFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(signedText), PtrValue(endptr), IntValue(bytecode.TypeI32, 0)})
	if err != nil || exit != nil {
		t.Fatalf("strtol ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeI64 || signedInt(ret) != -42 {
		t.Fatalf("strtol ret=%#v, want i64 -42", ret)
	}
	loadedEnd, err := mem.Load(endptr, bytecode.TypePtr, target.PointerAlign)
	if err != nil {
		t.Fatalf("load strtol endptr: %v", err)
	}
	if loadedEnd.Int != signedText+7 {
		t.Fatalf("strtol endptr=%#x, want %#x", loadedEnd.Int, signedText+7)
	}

	decimalText := mustAllocBytes(t, mem, "strtol:decimal", []byte("  +123abc\x00"), true, blockString)
	ret, exit, err = strtolFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(decimalText), PtrValue(endptr), IntValue(bytecode.TypeI32, 0)})
	if err != nil || exit != nil {
		t.Fatalf("strtol decimal ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeI64 || signedInt(ret) != 123 {
		t.Fatalf("strtol decimal ret=%#v, want i64 123", ret)
	}
	loadedEnd, err = mem.Load(endptr, bytecode.TypePtr, target.PointerAlign)
	if err != nil {
		t.Fatalf("load strtol decimal endptr: %v", err)
	}
	if loadedEnd.Int != decimalText+6 {
		t.Fatalf("strtol decimal endptr=%#x, want %#x", loadedEnd.Int, decimalText+6)
	}

	octalSignedText := mustAllocBytes(t, mem, "strtol:octal", []byte("0755,tail\x00"), true, blockString)
	ret, exit, err = strtolFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(octalSignedText), PtrValue(endptr), IntValue(bytecode.TypeI32, 0)})
	if err != nil || exit != nil {
		t.Fatalf("strtol octal ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeI64 || signedInt(ret) != 493 {
		t.Fatalf("strtol octal ret=%#v, want i64 493", ret)
	}
	loadedEnd, err = mem.Load(endptr, bytecode.TypePtr, target.PointerAlign)
	if err != nil {
		t.Fatalf("load strtol octal endptr: %v", err)
	}
	if loadedEnd.Int != octalSignedText+4 {
		t.Fatalf("strtol octal endptr=%#x, want %#x", loadedEnd.Int, octalSignedText+4)
	}

	strtoulFn, ok := reg.Lookup("strtoul")
	if !ok {
		t.Fatal("missing strtoul extern")
	}
	octalText := mustAllocBytes(t, mem, "strtoul:octal", []byte("077z\x00"), true, blockString)
	ret, exit, err = strtoulFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(octalText), PtrValue(endptr), IntValue(bytecode.TypeI32, 0)})
	if err != nil || exit != nil {
		t.Fatalf("strtoul octal ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeU64 || ret.Int != 63 {
		t.Fatalf("strtoul octal ret=%#v, want u64 63", ret)
	}
	loadedEnd, err = mem.Load(endptr, bytecode.TypePtr, target.PointerAlign)
	if err != nil {
		t.Fatalf("load strtoul endptr: %v", err)
	}
	if loadedEnd.Int != octalText+3 {
		t.Fatalf("strtoul endptr=%#x, want %#x", loadedEnd.Int, octalText+3)
	}

	hexText := mustAllocBytes(t, mem, "strtoul:hex", []byte("7f!\x00"), true, blockString)
	ret, exit, err = strtoulFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(hexText), PtrValue(0), IntValue(bytecode.TypeI32, 16)})
	if err != nil || exit != nil {
		t.Fatalf("strtoul hex ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeU64 || ret.Int != 127 {
		t.Fatalf("strtoul hex ret=%#v, want u64 127", ret)
	}

	negativeUnsignedText := mustAllocBytes(t, mem, "strtoul:negative", []byte("-42xyz\x00"), true, blockString)
	ret, exit, err = strtoulFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(negativeUnsignedText), PtrValue(endptr), IntValue(bytecode.TypeI32, 10)})
	if err != nil || exit != nil {
		t.Fatalf("strtoul negative ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeU64 || ret.Int != ^uint64(41) {
		t.Fatalf("strtoul negative ret=%#v, want u64 -42", ret)
	}
	loadedEnd, err = mem.Load(endptr, bytecode.TypePtr, target.PointerAlign)
	if err != nil {
		t.Fatalf("load strtoul negative endptr: %v", err)
	}
	if loadedEnd.Int != negativeUnsignedText+3 {
		t.Fatalf("strtoul negative endptr=%#x, want %#x", loadedEnd.Int, negativeUnsignedText+3)
	}

	unsignedNoneText := mustAllocBytes(t, mem, "strtoul:none", []byte("nope\x00"), true, blockString)
	ret, exit, err = strtoulFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(unsignedNoneText), PtrValue(endptr), IntValue(bytecode.TypeI32, 10)})
	if err != nil || exit != nil {
		t.Fatalf("strtoul none ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeU64 || ret.Int != 0 {
		t.Fatalf("strtoul none ret=%#v, want u64 0", ret)
	}
	loadedEnd, err = mem.Load(endptr, bytecode.TypePtr, target.PointerAlign)
	if err != nil {
		t.Fatalf("load strtoul none endptr: %v", err)
	}
	if loadedEnd.Int != unsignedNoneText {
		t.Fatalf("strtoul none endptr=%#x, want %#x", loadedEnd.Int, unsignedNoneText)
	}

	noneText := mustAllocBytes(t, mem, "strtol:none", []byte("xyz\x00"), true, blockString)
	ret, exit, err = strtolFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(noneText), PtrValue(endptr), IntValue(bytecode.TypeI32, 10)})
	if err != nil || exit != nil {
		t.Fatalf("strtol none ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeI64 || ret.Int != 0 {
		t.Fatalf("strtol none ret=%#v, want i64 0", ret)
	}
	loadedEnd, err = mem.Load(endptr, bytecode.TypePtr, target.PointerAlign)
	if err != nil {
		t.Fatalf("load strtol none endptr: %v", err)
	}
	if loadedEnd.Int != noneText {
		t.Fatalf("strtol none endptr=%#x, want %#x", loadedEnd.Int, noneText)
	}
}

func TestStdlibStrtollExterns(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	target := bytecode.DefaultTarget()
	mem := NewMemory(target)
	endptr, err := mem.TryAlloc("strtoll:endptr", target.PointerSize, target.PointerAlign, false, blockGlobal)
	if err != nil {
		t.Fatalf("alloc endptr: %v", err)
	}

	strtollFn, ok := reg.Lookup("strtoll")
	if !ok {
		t.Fatal("missing strtoll extern")
	}
	signedText := mustAllocBytes(t, mem, "strtoll:signed", []byte("-9876543210x\x00"), true, blockString)
	ret, exit, err := strtollFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(signedText), PtrValue(endptr), IntValue(bytecode.TypeI32, 10)})
	if err != nil || exit != nil {
		t.Fatalf("strtoll ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeI64 || signedInt(ret) != -9876543210 {
		t.Fatalf("strtoll ret=%#v, want i64 -9876543210", ret)
	}
	loadedEnd, err := mem.Load(endptr, bytecode.TypePtr, target.PointerAlign)
	if err != nil {
		t.Fatalf("load strtoll endptr: %v", err)
	}
	if loadedEnd.Int != signedText+11 {
		t.Fatalf("strtoll endptr=%#x, want %#x", loadedEnd.Int, signedText+11)
	}

	maxSignedText := mustAllocBytes(t, mem, "strtoll:max", []byte("9223372036854775807;\x00"), true, blockString)
	ret, exit, err = strtollFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(maxSignedText), PtrValue(endptr), IntValue(bytecode.TypeI32, 10)})
	if err != nil || exit != nil {
		t.Fatalf("strtoll max ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeI64 || signedInt(ret) != 9223372036854775807 {
		t.Fatalf("strtoll max ret=%#v, want i64 max", ret)
	}
	loadedEnd, err = mem.Load(endptr, bytecode.TypePtr, target.PointerAlign)
	if err != nil {
		t.Fatalf("load strtoll max endptr: %v", err)
	}
	if loadedEnd.Int != maxSignedText+19 {
		t.Fatalf("strtoll max endptr=%#x, want %#x", loadedEnd.Int, maxSignedText+19)
	}

	strtoullFn, ok := reg.Lookup("strtoull")
	if !ok {
		t.Fatal("missing strtoull extern")
	}
	unsignedText := mustAllocBytes(t, mem, "strtoull:unsigned", []byte("0X100000000z\x00"), true, blockString)
	ret, exit, err = strtoullFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(unsignedText), PtrValue(endptr), IntValue(bytecode.TypeI32, 0)})
	if err != nil || exit != nil {
		t.Fatalf("strtoull ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeU64 || ret.Int != 4294967296 {
		t.Fatalf("strtoull ret=%#v, want u64 4294967296", ret)
	}
	loadedEnd, err = mem.Load(endptr, bytecode.TypePtr, target.PointerAlign)
	if err != nil {
		t.Fatalf("load strtoull endptr: %v", err)
	}
	if loadedEnd.Int != unsignedText+11 {
		t.Fatalf("strtoull endptr=%#x, want %#x", loadedEnd.Int, unsignedText+11)
	}

	maxUnsignedText := mustAllocBytes(t, mem, "strtoull:max", []byte("18446744073709551615!\x00"), true, blockString)
	ret, exit, err = strtoullFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(maxUnsignedText), PtrValue(endptr), IntValue(bytecode.TypeI32, 10)})
	if err != nil || exit != nil {
		t.Fatalf("strtoull max ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeU64 || ret.Int != ^uint64(0) {
		t.Fatalf("strtoull max ret=%#v, want u64 max", ret)
	}
	loadedEnd, err = mem.Load(endptr, bytecode.TypePtr, target.PointerAlign)
	if err != nil {
		t.Fatalf("load strtoull max endptr: %v", err)
	}
	if loadedEnd.Int != maxUnsignedText+20 {
		t.Fatalf("strtoull max endptr=%#x, want %#x", loadedEnd.Int, maxUnsignedText+20)
	}
}

func TestStdlibFloatParserExterns(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	target := bytecode.DefaultTarget()
	mem := NewMemory(target)
	endptr, err := mem.TryAlloc("strtod:endptr", target.PointerSize, target.PointerAlign, false, blockGlobal)
	if err != nil {
		t.Fatalf("alloc endptr: %v", err)
	}

	atofFn, ok := reg.Lookup("atof")
	if !ok {
		t.Fatal("missing atof extern")
	}
	atofText := mustAllocBytes(t, mem, "atof:decimal", []byte(" \t3.25tail\x00"), true, blockString)
	ret, exit, err := atofFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(atofText)})
	if err != nil || exit != nil {
		t.Fatalf("atof ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeF64 || ret.Float != 3.25 {
		t.Fatalf("atof ret=%#v, want f64 3.25", ret)
	}

	atofExponentText := mustAllocBytes(t, mem, "atof:exponent", []byte("-1.25e2x\x00"), true, blockString)
	ret, exit, err = atofFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(atofExponentText)})
	if err != nil || exit != nil {
		t.Fatalf("atof exponent ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeF64 || ret.Float != -125 {
		t.Fatalf("atof exponent ret=%#v, want f64 -125", ret)
	}

	atofNoneText := mustAllocBytes(t, mem, "atof:none", []byte("word\x00"), true, blockString)
	ret, exit, err = atofFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(atofNoneText)})
	if err != nil || exit != nil {
		t.Fatalf("atof none ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeF64 || ret.Float != 0 {
		t.Fatalf("atof none ret=%#v, want f64 0", ret)
	}

	strtodFn, ok := reg.Lookup("strtod")
	if !ok {
		t.Fatal("missing strtod extern")
	}
	errnoAddr, ok := reg.LookupVariable("errno", mem)
	if !ok {
		t.Fatal("missing errno extern variable")
	}
	if err := mem.Store(errnoAddr, bytecode.TypeI32, 4, IntValue(bytecode.TypeI32, 77)); err != nil {
		t.Fatalf("store errno before strtod: %v", err)
	}
	ret, exit, err = atofFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(atofNoneText)})
	if err != nil || exit != nil {
		t.Fatalf("atof errno-preserving none ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeF64 || ret.Float != 0 {
		t.Fatalf("atof errno-preserving none ret=%#v, want f64 0", ret)
	}
	errnoValue, err := mem.Load(errnoAddr, bytecode.TypeI32, 4)
	if err != nil {
		t.Fatalf("load errno after atof none: %v", err)
	}
	if signedInt(errnoValue) != 77 {
		t.Fatalf("errno after atof none=%#v, want 77", errnoValue)
	}
	if err := mem.Store(errnoAddr, bytecode.TypeI32, 4, IntValue(bytecode.TypeI32, 0)); err != nil {
		t.Fatalf("store errno before atof overflow: %v", err)
	}
	atofOverflowText := mustAllocBytes(t, mem, "atof:overflow", []byte("1e309tail\x00"), true, blockString)
	ret, exit, err = atofFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(atofOverflowText)})
	if err != nil || exit != nil {
		t.Fatalf("atof overflow ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeF64 || !math.IsInf(ret.Float, 1) {
		t.Fatalf("atof overflow ret=%#v, want +inf", ret)
	}
	errnoValue, err = mem.Load(errnoAddr, bytecode.TypeI32, 4)
	if err != nil {
		t.Fatalf("load errno after atof overflow: %v", err)
	}
	if signedInt(errnoValue) != 34 {
		t.Fatalf("errno after atof overflow=%#v, want ERANGE", errnoValue)
	}
	if err := mem.Store(errnoAddr, bytecode.TypeI32, 4, IntValue(bytecode.TypeI32, 0)); err != nil {
		t.Fatalf("store errno before atof underflow: %v", err)
	}
	atofUnderflowText := mustAllocBytes(t, mem, "atof:underflow", []byte("-1e-400tail\x00"), true, blockString)
	ret, exit, err = atofFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(atofUnderflowText)})
	if err != nil || exit != nil {
		t.Fatalf("atof underflow ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeF64 || ret.Float != 0 || !math.Signbit(ret.Float) {
		t.Fatalf("atof underflow ret=%#v, want -0", ret)
	}
	errnoValue, err = mem.Load(errnoAddr, bytecode.TypeI32, 4)
	if err != nil {
		t.Fatalf("load errno after atof underflow: %v", err)
	}
	if signedInt(errnoValue) != 34 {
		t.Fatalf("errno after atof underflow=%#v, want ERANGE", errnoValue)
	}
	if err := mem.Store(errnoAddr, bytecode.TypeI32, 4, IntValue(bytecode.TypeI32, 77)); err != nil {
		t.Fatalf("restore errno before strtod: %v", err)
	}
	decimalText := mustAllocBytes(t, mem, "strtod:decimal", []byte(" -12.5e1x\x00"), true, blockString)
	ret, exit, err = strtodFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(decimalText), PtrValue(endptr)})
	if err != nil || exit != nil {
		t.Fatalf("strtod decimal ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeF64 || ret.Float != -125 {
		t.Fatalf("strtod decimal ret=%#v, want f64 -125", ret)
	}
	loadedEnd, err := mem.Load(endptr, bytecode.TypePtr, target.PointerAlign)
	if err != nil {
		t.Fatalf("load strtod decimal endptr: %v", err)
	}
	if loadedEnd.Int != decimalText+8 {
		t.Fatalf("strtod decimal endptr=%#x, want %#x", loadedEnd.Int, decimalText+8)
	}
	errnoValue, err = mem.Load(errnoAddr, bytecode.TypeI32, 4)
	if err != nil {
		t.Fatalf("load errno after strtod success: %v", err)
	}
	if signedInt(errnoValue) != 77 {
		t.Fatalf("errno after strtod success=%#v, want 77", errnoValue)
	}

	positiveExponentText := mustAllocBytes(t, mem, "strtod:positive-exponent", []byte(" +6.25e-1;\x00"), true, blockString)
	ret, exit, err = strtodFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(positiveExponentText), PtrValue(endptr)})
	if err != nil || exit != nil {
		t.Fatalf("strtod positive exponent ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeF64 || ret.Float != 0.625 {
		t.Fatalf("strtod positive exponent ret=%#v, want f64 0.625", ret)
	}
	loadedEnd, err = mem.Load(endptr, bytecode.TypePtr, target.PointerAlign)
	if err != nil {
		t.Fatalf("load strtod positive exponent endptr: %v", err)
	}
	if loadedEnd.Int != positiveExponentText+9 {
		t.Fatalf("strtod positive exponent endptr=%#x, want %#x", loadedEnd.Int, positiveExponentText+9)
	}

	hexText := mustAllocBytes(t, mem, "strtod:hex", []byte("0x1.8p+2!\x00"), true, blockString)
	ret, exit, err = strtodFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(hexText), PtrValue(endptr)})
	if err != nil || exit != nil {
		t.Fatalf("strtod hex ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeF64 || ret.Float != 6 {
		t.Fatalf("strtod hex ret=%#v, want f64 6", ret)
	}
	loadedEnd, err = mem.Load(endptr, bytecode.TypePtr, target.PointerAlign)
	if err != nil {
		t.Fatalf("load strtod hex endptr: %v", err)
	}
	if loadedEnd.Int != hexText+8 {
		t.Fatalf("strtod hex endptr=%#x, want %#x", loadedEnd.Int, hexText+8)
	}

	hexOneText := mustAllocBytes(t, mem, "strtod:hex-one", []byte("0x1p0:\x00"), true, blockString)
	ret, exit, err = strtodFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(hexOneText), PtrValue(endptr)})
	if err != nil || exit != nil {
		t.Fatalf("strtod hex one ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeF64 || ret.Float != 1 {
		t.Fatalf("strtod hex one ret=%#v, want f64 1", ret)
	}
	loadedEnd, err = mem.Load(endptr, bytecode.TypePtr, target.PointerAlign)
	if err != nil {
		t.Fatalf("load strtod hex one endptr: %v", err)
	}
	if loadedEnd.Int != hexOneText+5 {
		t.Fatalf("strtod hex one endptr=%#x, want %#x", loadedEnd.Int, hexOneText+5)
	}

	hexFractionText := mustAllocBytes(t, mem, "strtod:hex-fraction", []byte("0x1.fp-1?\x00"), true, blockString)
	ret, exit, err = strtodFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(hexFractionText), PtrValue(endptr)})
	if err != nil || exit != nil {
		t.Fatalf("strtod hex fraction ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeF64 || ret.Float != 0.96875 {
		t.Fatalf("strtod hex fraction ret=%#v, want f64 0.96875", ret)
	}
	loadedEnd, err = mem.Load(endptr, bytecode.TypePtr, target.PointerAlign)
	if err != nil {
		t.Fatalf("load strtod hex fraction endptr: %v", err)
	}
	if loadedEnd.Int != hexFractionText+8 {
		t.Fatalf("strtod hex fraction endptr=%#x, want %#x", loadedEnd.Int, hexFractionText+8)
	}

	infText := mustAllocBytes(t, mem, "strtod:inf", []byte("inf!\x00"), true, blockString)
	ret, exit, err = strtodFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(infText), PtrValue(endptr)})
	if err != nil || exit != nil {
		t.Fatalf("strtod inf ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeF64 || !math.IsInf(ret.Float, 1) {
		t.Fatalf("strtod inf ret=%#v, want +inf", ret)
	}
	loadedEnd, err = mem.Load(endptr, bytecode.TypePtr, target.PointerAlign)
	if err != nil {
		t.Fatalf("load strtod inf endptr: %v", err)
	}
	if loadedEnd.Int != infText+3 {
		t.Fatalf("strtod inf endptr=%#x, want %#x", loadedEnd.Int, infText+3)
	}

	positiveInfinityText := mustAllocBytes(t, mem, "strtod:positive-infinity", []byte("+infinity;\x00"), true, blockString)
	ret, exit, err = strtodFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(positiveInfinityText), PtrValue(endptr)})
	if err != nil || exit != nil {
		t.Fatalf("strtod positive infinity ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeF64 || !math.IsInf(ret.Float, 1) {
		t.Fatalf("strtod positive infinity ret=%#v, want +inf", ret)
	}
	loadedEnd, err = mem.Load(endptr, bytecode.TypePtr, target.PointerAlign)
	if err != nil {
		t.Fatalf("load strtod positive infinity endptr: %v", err)
	}
	if loadedEnd.Int != positiveInfinityText+9 {
		t.Fatalf("strtod positive infinity endptr=%#x, want %#x", loadedEnd.Int, positiveInfinityText+9)
	}

	negativeInfText := mustAllocBytes(t, mem, "strtod:negative-inf", []byte("-INF?\x00"), true, blockString)
	ret, exit, err = strtodFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(negativeInfText), PtrValue(endptr)})
	if err != nil || exit != nil {
		t.Fatalf("strtod negative inf ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeF64 || !math.IsInf(ret.Float, -1) {
		t.Fatalf("strtod negative inf ret=%#v, want -inf", ret)
	}
	loadedEnd, err = mem.Load(endptr, bytecode.TypePtr, target.PointerAlign)
	if err != nil {
		t.Fatalf("load strtod negative inf endptr: %v", err)
	}
	if loadedEnd.Int != negativeInfText+4 {
		t.Fatalf("strtod negative inf endptr=%#x, want %#x", loadedEnd.Int, negativeInfText+4)
	}

	nanText := mustAllocBytes(t, mem, "strtod:nan", []byte("nan!\x00"), true, blockString)
	ret, exit, err = strtodFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(nanText), PtrValue(endptr)})
	if err != nil || exit != nil {
		t.Fatalf("strtod nan ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeF64 || !math.IsNaN(ret.Float) {
		t.Fatalf("strtod nan ret=%#v, want NaN", ret)
	}
	loadedEnd, err = mem.Load(endptr, bytecode.TypePtr, target.PointerAlign)
	if err != nil {
		t.Fatalf("load strtod nan endptr: %v", err)
	}
	if loadedEnd.Int != nanText+3 {
		t.Fatalf("strtod nan endptr=%#x, want %#x", loadedEnd.Int, nanText+3)
	}

	upperNanText := mustAllocBytes(t, mem, "strtod:upper-nan", []byte("NAN;\x00"), true, blockString)
	ret, exit, err = strtodFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(upperNanText), PtrValue(endptr)})
	if err != nil || exit != nil {
		t.Fatalf("strtod upper nan ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeF64 || !math.IsNaN(ret.Float) {
		t.Fatalf("strtod upper nan ret=%#v, want NaN", ret)
	}
	loadedEnd, err = mem.Load(endptr, bytecode.TypePtr, target.PointerAlign)
	if err != nil {
		t.Fatalf("load strtod upper nan endptr: %v", err)
	}
	if loadedEnd.Int != upperNanText+3 {
		t.Fatalf("strtod upper nan endptr=%#x, want %#x", loadedEnd.Int, upperNanText+3)
	}

	nanPayloadText := mustAllocBytes(t, mem, "strtod:nan-payload", []byte("nan(payload)?\x00"), true, blockString)
	ret, exit, err = strtodFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(nanPayloadText), PtrValue(endptr)})
	if err != nil || exit != nil {
		t.Fatalf("strtod nan payload ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeF64 || !math.IsNaN(ret.Float) {
		t.Fatalf("strtod nan payload ret=%#v, want NaN", ret)
	}
	loadedEnd, err = mem.Load(endptr, bytecode.TypePtr, target.PointerAlign)
	if err != nil {
		t.Fatalf("load strtod nan payload endptr: %v", err)
	}
	if loadedEnd.Int != nanPayloadText+12 {
		t.Fatalf("strtod nan payload endptr=%#x, want %#x", loadedEnd.Int, nanPayloadText+12)
	}

	noneText := mustAllocBytes(t, mem, "strtod:none", []byte("word\x00"), true, blockString)
	ret, exit, err = strtodFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(noneText), PtrValue(endptr)})
	if err != nil || exit != nil {
		t.Fatalf("strtod none ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeF64 || ret.Float != 0 {
		t.Fatalf("strtod none ret=%#v, want f64 0", ret)
	}
	loadedEnd, err = mem.Load(endptr, bytecode.TypePtr, target.PointerAlign)
	if err != nil {
		t.Fatalf("load strtod none endptr: %v", err)
	}
	if loadedEnd.Int != noneText {
		t.Fatalf("strtod none endptr=%#x, want %#x", loadedEnd.Int, noneText)
	}
	errnoValue, err = mem.Load(errnoAddr, bytecode.TypeI32, 4)
	if err != nil {
		t.Fatalf("load errno after strtod none: %v", err)
	}
	if signedInt(errnoValue) != 77 {
		t.Fatalf("errno after strtod none=%#v, want 77", errnoValue)
	}

	if err := mem.Store(errnoAddr, bytecode.TypeI32, 4, IntValue(bytecode.TypeI32, 0)); err != nil {
		t.Fatalf("store errno before strtod positive overflow: %v", err)
	}
	overflowText := mustAllocBytes(t, mem, "strtod:overflow", []byte("1e309!\x00"), true, blockString)
	ret, exit, err = strtodFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(overflowText), PtrValue(endptr)})
	if err != nil || exit != nil {
		t.Fatalf("strtod overflow ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeF64 || !math.IsInf(ret.Float, 1) {
		t.Fatalf("strtod overflow ret=%#v, want +inf", ret)
	}
	loadedEnd, err = mem.Load(endptr, bytecode.TypePtr, target.PointerAlign)
	if err != nil {
		t.Fatalf("load strtod overflow endptr: %v", err)
	}
	if loadedEnd.Int != overflowText+5 {
		t.Fatalf("strtod overflow endptr=%#x, want %#x", loadedEnd.Int, overflowText+5)
	}
	errnoValue, err = mem.Load(errnoAddr, bytecode.TypeI32, 4)
	if err != nil {
		t.Fatalf("load errno after strtod overflow: %v", err)
	}
	if signedInt(errnoValue) != 34 {
		t.Fatalf("errno after strtod overflow=%#v, want ERANGE", errnoValue)
	}

	if err := mem.Store(errnoAddr, bytecode.TypeI32, 4, IntValue(bytecode.TypeI32, 0)); err != nil {
		t.Fatalf("store errno before strtod negative overflow: %v", err)
	}
	negativeOverflowText := mustAllocBytes(t, mem, "strtod:negative-overflow", []byte("-1e309?\x00"), true, blockString)
	ret, exit, err = strtodFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(negativeOverflowText), PtrValue(endptr)})
	if err != nil || exit != nil {
		t.Fatalf("strtod negative overflow ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeF64 || !math.IsInf(ret.Float, -1) {
		t.Fatalf("strtod negative overflow ret=%#v, want -inf", ret)
	}
	loadedEnd, err = mem.Load(endptr, bytecode.TypePtr, target.PointerAlign)
	if err != nil {
		t.Fatalf("load strtod negative overflow endptr: %v", err)
	}
	if loadedEnd.Int != negativeOverflowText+6 {
		t.Fatalf("strtod negative overflow endptr=%#x, want %#x", loadedEnd.Int, negativeOverflowText+6)
	}
	errnoValue, err = mem.Load(errnoAddr, bytecode.TypeI32, 4)
	if err != nil {
		t.Fatalf("load errno after strtod negative overflow: %v", err)
	}
	if signedInt(errnoValue) != 34 {
		t.Fatalf("errno after strtod negative overflow=%#v, want ERANGE", errnoValue)
	}

	if err := mem.Store(errnoAddr, bytecode.TypeI32, 4, IntValue(bytecode.TypeI32, 0)); err != nil {
		t.Fatalf("store errno before strtod hex overflow: %v", err)
	}
	hexOverflowText := mustAllocBytes(t, mem, "strtod:hex-overflow", []byte("0x1p+2048!\x00"), true, blockString)
	ret, exit, err = strtodFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(hexOverflowText), PtrValue(endptr)})
	if err != nil || exit != nil {
		t.Fatalf("strtod hex overflow ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeF64 || !math.IsInf(ret.Float, 1) {
		t.Fatalf("strtod hex overflow ret=%#v, want +inf", ret)
	}
	loadedEnd, err = mem.Load(endptr, bytecode.TypePtr, target.PointerAlign)
	if err != nil {
		t.Fatalf("load strtod hex overflow endptr: %v", err)
	}
	if loadedEnd.Int != hexOverflowText+9 {
		t.Fatalf("strtod hex overflow endptr=%#x, want %#x", loadedEnd.Int, hexOverflowText+9)
	}
	errnoValue, err = mem.Load(errnoAddr, bytecode.TypeI32, 4)
	if err != nil {
		t.Fatalf("load errno after strtod hex overflow: %v", err)
	}
	if signedInt(errnoValue) != 34 {
		t.Fatalf("errno after strtod hex overflow=%#v, want ERANGE", errnoValue)
	}

	if err := mem.Store(errnoAddr, bytecode.TypeI32, 4, IntValue(bytecode.TypeI32, 0)); err != nil {
		t.Fatalf("store errno before strtod negative hex overflow: %v", err)
	}
	negativeHexOverflowText := mustAllocBytes(t, mem, "strtod:negative-hex-overflow", []byte("-0x1p+2048?\x00"), true, blockString)
	ret, exit, err = strtodFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(negativeHexOverflowText), PtrValue(endptr)})
	if err != nil || exit != nil {
		t.Fatalf("strtod negative hex overflow ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeF64 || !math.IsInf(ret.Float, -1) {
		t.Fatalf("strtod negative hex overflow ret=%#v, want -inf", ret)
	}
	loadedEnd, err = mem.Load(endptr, bytecode.TypePtr, target.PointerAlign)
	if err != nil {
		t.Fatalf("load strtod negative hex overflow endptr: %v", err)
	}
	if loadedEnd.Int != negativeHexOverflowText+10 {
		t.Fatalf("strtod negative hex overflow endptr=%#x, want %#x", loadedEnd.Int, negativeHexOverflowText+10)
	}
	errnoValue, err = mem.Load(errnoAddr, bytecode.TypeI32, 4)
	if err != nil {
		t.Fatalf("load errno after strtod negative hex overflow: %v", err)
	}
	if signedInt(errnoValue) != 34 {
		t.Fatalf("errno after strtod negative hex overflow=%#v, want ERANGE", errnoValue)
	}

	if err := mem.Store(errnoAddr, bytecode.TypeI32, 4, IntValue(bytecode.TypeI32, 0)); err != nil {
		t.Fatalf("store errno before strtod underflow: %v", err)
	}
	underflowText := mustAllocBytes(t, mem, "strtod:underflow", []byte("1e-400!\x00"), true, blockString)
	ret, exit, err = strtodFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(underflowText), PtrValue(endptr)})
	if err != nil || exit != nil {
		t.Fatalf("strtod underflow ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeF64 || ret.Float != 0 || math.Signbit(ret.Float) {
		t.Fatalf("strtod underflow ret=%#v, want +0", ret)
	}
	loadedEnd, err = mem.Load(endptr, bytecode.TypePtr, target.PointerAlign)
	if err != nil {
		t.Fatalf("load strtod underflow endptr: %v", err)
	}
	if loadedEnd.Int != underflowText+6 {
		t.Fatalf("strtod underflow endptr=%#x, want %#x", loadedEnd.Int, underflowText+6)
	}
	errnoValue, err = mem.Load(errnoAddr, bytecode.TypeI32, 4)
	if err != nil {
		t.Fatalf("load errno after strtod underflow: %v", err)
	}
	if signedInt(errnoValue) != 34 {
		t.Fatalf("errno after strtod underflow=%#v, want ERANGE", errnoValue)
	}

	if err := mem.Store(errnoAddr, bytecode.TypeI32, 4, IntValue(bytecode.TypeI32, 0)); err != nil {
		t.Fatalf("store errno before strtod negative underflow: %v", err)
	}
	negativeUnderflowText := mustAllocBytes(t, mem, "strtod:negative-underflow", []byte("-1e-400?\x00"), true, blockString)
	ret, exit, err = strtodFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(negativeUnderflowText), PtrValue(endptr)})
	if err != nil || exit != nil {
		t.Fatalf("strtod negative underflow ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeF64 || ret.Float != 0 || !math.Signbit(ret.Float) {
		t.Fatalf("strtod negative underflow ret=%#v, want -0", ret)
	}
	loadedEnd, err = mem.Load(endptr, bytecode.TypePtr, target.PointerAlign)
	if err != nil {
		t.Fatalf("load strtod negative underflow endptr: %v", err)
	}
	if loadedEnd.Int != negativeUnderflowText+7 {
		t.Fatalf("strtod negative underflow endptr=%#x, want %#x", loadedEnd.Int, negativeUnderflowText+7)
	}
	errnoValue, err = mem.Load(errnoAddr, bytecode.TypeI32, 4)
	if err != nil {
		t.Fatalf("load errno after strtod negative underflow: %v", err)
	}
	if signedInt(errnoValue) != 34 {
		t.Fatalf("errno after strtod negative underflow=%#v, want ERANGE", errnoValue)
	}

	if err := mem.Store(errnoAddr, bytecode.TypeI32, 4, IntValue(bytecode.TypeI32, 0)); err != nil {
		t.Fatalf("store errno before strtod subnormal: %v", err)
	}
	subnormalText := mustAllocBytes(t, mem, "strtod:subnormal", []byte("5e-324!\x00"), true, blockString)
	ret, exit, err = strtodFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(subnormalText), PtrValue(endptr)})
	if err != nil || exit != nil {
		t.Fatalf("strtod subnormal ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeF64 || ret.Float != 5e-324 {
		t.Fatalf("strtod subnormal ret=%#v, want smallest subnormal", ret)
	}
	loadedEnd, err = mem.Load(endptr, bytecode.TypePtr, target.PointerAlign)
	if err != nil {
		t.Fatalf("load strtod subnormal endptr: %v", err)
	}
	if loadedEnd.Int != subnormalText+6 {
		t.Fatalf("strtod subnormal endptr=%#x, want %#x", loadedEnd.Int, subnormalText+6)
	}
	errnoValue, err = mem.Load(errnoAddr, bytecode.TypeI32, 4)
	if err != nil {
		t.Fatalf("load errno after strtod subnormal: %v", err)
	}
	if signedInt(errnoValue) != 0 {
		t.Fatalf("errno after strtod subnormal=%#v, want 0", errnoValue)
	}

	if err := mem.Store(errnoAddr, bytecode.TypeI32, 4, IntValue(bytecode.TypeI32, 0)); err != nil {
		t.Fatalf("store errno before strtod hex underflow: %v", err)
	}
	hexUnderflowText := mustAllocBytes(t, mem, "strtod:hex-underflow", []byte("0x1p-20000!\x00"), true, blockString)
	ret, exit, err = strtodFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(hexUnderflowText), PtrValue(endptr)})
	if err != nil || exit != nil {
		t.Fatalf("strtod hex underflow ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeF64 || ret.Float != 0 || math.Signbit(ret.Float) {
		t.Fatalf("strtod hex underflow ret=%#v, want +0", ret)
	}
	loadedEnd, err = mem.Load(endptr, bytecode.TypePtr, target.PointerAlign)
	if err != nil {
		t.Fatalf("load strtod hex underflow endptr: %v", err)
	}
	if loadedEnd.Int != hexUnderflowText+10 {
		t.Fatalf("strtod hex underflow endptr=%#x, want %#x", loadedEnd.Int, hexUnderflowText+10)
	}
	errnoValue, err = mem.Load(errnoAddr, bytecode.TypeI32, 4)
	if err != nil {
		t.Fatalf("load errno after strtod hex underflow: %v", err)
	}
	if signedInt(errnoValue) != 34 {
		t.Fatalf("errno after strtod hex underflow=%#v, want ERANGE", errnoValue)
	}

	if err := mem.Store(errnoAddr, bytecode.TypeI32, 4, IntValue(bytecode.TypeI32, 0)); err != nil {
		t.Fatalf("store errno before strtod negative hex underflow: %v", err)
	}
	negativeHexUnderflowText := mustAllocBytes(t, mem, "strtod:negative-hex-underflow", []byte("-0x1p-20000?\x00"), true, blockString)
	ret, exit, err = strtodFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(negativeHexUnderflowText), PtrValue(endptr)})
	if err != nil || exit != nil {
		t.Fatalf("strtod negative hex underflow ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeF64 || ret.Float != 0 || !math.Signbit(ret.Float) {
		t.Fatalf("strtod negative hex underflow ret=%#v, want -0", ret)
	}
	loadedEnd, err = mem.Load(endptr, bytecode.TypePtr, target.PointerAlign)
	if err != nil {
		t.Fatalf("load strtod negative hex underflow endptr: %v", err)
	}
	if loadedEnd.Int != negativeHexUnderflowText+11 {
		t.Fatalf("strtod negative hex underflow endptr=%#x, want %#x", loadedEnd.Int, negativeHexUnderflowText+11)
	}
	errnoValue, err = mem.Load(errnoAddr, bytecode.TypeI32, 4)
	if err != nil {
		t.Fatalf("load errno after strtod negative hex underflow: %v", err)
	}
	if signedInt(errnoValue) != 34 {
		t.Fatalf("errno after strtod negative hex underflow=%#v, want ERANGE", errnoValue)
	}

	if err := mem.Store(errnoAddr, bytecode.TypeI32, 4, IntValue(bytecode.TypeI32, 0)); err != nil {
		t.Fatalf("store errno before strtod hex subnormal: %v", err)
	}
	hexSubnormalText := mustAllocBytes(t, mem, "strtod:hex-subnormal", []byte("0x1p-1074!\x00"), true, blockString)
	ret, exit, err = strtodFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(hexSubnormalText), PtrValue(endptr)})
	if err != nil || exit != nil {
		t.Fatalf("strtod hex subnormal ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeF64 || ret.Float != 5e-324 {
		t.Fatalf("strtod hex subnormal ret=%#v, want smallest subnormal", ret)
	}
	loadedEnd, err = mem.Load(endptr, bytecode.TypePtr, target.PointerAlign)
	if err != nil {
		t.Fatalf("load strtod hex subnormal endptr: %v", err)
	}
	if loadedEnd.Int != hexSubnormalText+9 {
		t.Fatalf("strtod hex subnormal endptr=%#x, want %#x", loadedEnd.Int, hexSubnormalText+9)
	}
	errnoValue, err = mem.Load(errnoAddr, bytecode.TypeI32, 4)
	if err != nil {
		t.Fatalf("load errno after strtod hex subnormal: %v", err)
	}
	if signedInt(errnoValue) != 0 {
		t.Fatalf("errno after strtod hex subnormal=%#v, want 0", errnoValue)
	}
}

func TestStdlibMoreFloatParserExterns(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	target := bytecode.DefaultTarget()
	mem := NewMemory(target)
	endptr, err := mem.TryAlloc("strtof:endptr", target.PointerSize, target.PointerAlign, false, blockGlobal)
	if err != nil {
		t.Fatalf("alloc endptr: %v", err)
	}

	strtofFn, ok := reg.Lookup("strtof")
	if !ok {
		t.Fatal("missing strtof extern")
	}
	floatText := mustAllocBytes(t, mem, "strtof:decimal", []byte(" 1.5!\x00"), true, blockString)
	ret, exit, err := strtofFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(floatText), PtrValue(endptr)})
	if err != nil || exit != nil {
		t.Fatalf("strtof ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeF32 || ret.Float != 1.5 {
		t.Fatalf("strtof ret=%#v, want f32 1.5", ret)
	}
	loadedEnd, err := mem.Load(endptr, bytecode.TypePtr, target.PointerAlign)
	if err != nil {
		t.Fatalf("load strtof endptr: %v", err)
	}
	if loadedEnd.Int != floatText+4 {
		t.Fatalf("strtof endptr=%#x, want %#x", loadedEnd.Int, floatText+4)
	}

	errnoAddr, ok := reg.LookupVariable("errno", mem)
	if !ok {
		t.Fatal("missing errno extern variable")
	}
	if err := mem.Store(errnoAddr, bytecode.TypeI32, 4, IntValue(bytecode.TypeI32, 0)); err != nil {
		t.Fatalf("store errno before strtof overflow: %v", err)
	}
	floatOverflowText := mustAllocBytes(t, mem, "strtof:overflow", []byte("1e39!\x00"), true, blockString)
	ret, exit, err = strtofFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(floatOverflowText), PtrValue(endptr)})
	if err != nil || exit != nil {
		t.Fatalf("strtof overflow ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeF32 || !math.IsInf(ret.Float, 1) {
		t.Fatalf("strtof overflow ret=%#v, want f32 +inf", ret)
	}
	loadedEnd, err = mem.Load(endptr, bytecode.TypePtr, target.PointerAlign)
	if err != nil {
		t.Fatalf("load strtof overflow endptr: %v", err)
	}
	if loadedEnd.Int != floatOverflowText+4 {
		t.Fatalf("strtof overflow endptr=%#x, want %#x", loadedEnd.Int, floatOverflowText+4)
	}
	errnoValue, err := mem.Load(errnoAddr, bytecode.TypeI32, 4)
	if err != nil {
		t.Fatalf("load errno after strtof overflow: %v", err)
	}
	if signedInt(errnoValue) != 34 {
		t.Fatalf("errno after strtof overflow=%#v, want ERANGE", errnoValue)
	}

	if err := mem.Store(errnoAddr, bytecode.TypeI32, 4, IntValue(bytecode.TypeI32, 0)); err != nil {
		t.Fatalf("store errno before strtof negative overflow: %v", err)
	}
	negativeFloatOverflowText := mustAllocBytes(t, mem, "strtof:negative-overflow", []byte("-1e39?\x00"), true, blockString)
	ret, exit, err = strtofFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(negativeFloatOverflowText), PtrValue(endptr)})
	if err != nil || exit != nil {
		t.Fatalf("strtof negative overflow ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeF32 || !math.IsInf(ret.Float, -1) {
		t.Fatalf("strtof negative overflow ret=%#v, want f32 -inf", ret)
	}
	loadedEnd, err = mem.Load(endptr, bytecode.TypePtr, target.PointerAlign)
	if err != nil {
		t.Fatalf("load strtof negative overflow endptr: %v", err)
	}
	if loadedEnd.Int != negativeFloatOverflowText+5 {
		t.Fatalf("strtof negative overflow endptr=%#x, want %#x", loadedEnd.Int, negativeFloatOverflowText+5)
	}
	errnoValue, err = mem.Load(errnoAddr, bytecode.TypeI32, 4)
	if err != nil {
		t.Fatalf("load errno after strtof negative overflow: %v", err)
	}
	if signedInt(errnoValue) != 34 {
		t.Fatalf("errno after strtof negative overflow=%#v, want ERANGE", errnoValue)
	}

	if err := mem.Store(errnoAddr, bytecode.TypeI32, 4, IntValue(bytecode.TypeI32, 0)); err != nil {
		t.Fatalf("store errno before strtof underflow: %v", err)
	}
	floatUnderflowText := mustAllocBytes(t, mem, "strtof:underflow", []byte("1e-50!\x00"), true, blockString)
	ret, exit, err = strtofFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(floatUnderflowText), PtrValue(endptr)})
	if err != nil || exit != nil {
		t.Fatalf("strtof underflow ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeF32 || ret.Float != 0 || math.Signbit(ret.Float) {
		t.Fatalf("strtof underflow ret=%#v, want f32 +0", ret)
	}
	loadedEnd, err = mem.Load(endptr, bytecode.TypePtr, target.PointerAlign)
	if err != nil {
		t.Fatalf("load strtof underflow endptr: %v", err)
	}
	if loadedEnd.Int != floatUnderflowText+5 {
		t.Fatalf("strtof underflow endptr=%#x, want %#x", loadedEnd.Int, floatUnderflowText+5)
	}
	errnoValue, err = mem.Load(errnoAddr, bytecode.TypeI32, 4)
	if err != nil {
		t.Fatalf("load errno after strtof underflow: %v", err)
	}
	if signedInt(errnoValue) != 34 {
		t.Fatalf("errno after strtof underflow=%#v, want ERANGE", errnoValue)
	}

	if err := mem.Store(errnoAddr, bytecode.TypeI32, 4, IntValue(bytecode.TypeI32, 0)); err != nil {
		t.Fatalf("store errno before strtof negative underflow: %v", err)
	}
	negativeFloatUnderflowText := mustAllocBytes(t, mem, "strtof:negative-underflow", []byte("-1e-50?\x00"), true, blockString)
	ret, exit, err = strtofFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(negativeFloatUnderflowText), PtrValue(endptr)})
	if err != nil || exit != nil {
		t.Fatalf("strtof negative underflow ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeF32 || ret.Float != 0 || !math.Signbit(ret.Float) {
		t.Fatalf("strtof negative underflow ret=%#v, want f32 -0", ret)
	}
	loadedEnd, err = mem.Load(endptr, bytecode.TypePtr, target.PointerAlign)
	if err != nil {
		t.Fatalf("load strtof negative underflow endptr: %v", err)
	}
	if loadedEnd.Int != negativeFloatUnderflowText+6 {
		t.Fatalf("strtof negative underflow endptr=%#x, want %#x", loadedEnd.Int, negativeFloatUnderflowText+6)
	}
	errnoValue, err = mem.Load(errnoAddr, bytecode.TypeI32, 4)
	if err != nil {
		t.Fatalf("load errno after strtof negative underflow: %v", err)
	}
	if signedInt(errnoValue) != 34 {
		t.Fatalf("errno after strtof negative underflow=%#v, want ERANGE", errnoValue)
	}

	if err := mem.Store(errnoAddr, bytecode.TypeI32, 4, IntValue(bytecode.TypeI32, 0)); err != nil {
		t.Fatalf("store errno before strtof subnormal: %v", err)
	}
	floatSubnormalText := mustAllocBytes(t, mem, "strtof:subnormal", []byte("1e-45!\x00"), true, blockString)
	ret, exit, err = strtofFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(floatSubnormalText), PtrValue(endptr)})
	if err != nil || exit != nil {
		t.Fatalf("strtof subnormal ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeF32 || ret.Float != float64(float32(math.SmallestNonzeroFloat32)) {
		t.Fatalf("strtof subnormal ret=%#v, want smallest float32 subnormal", ret)
	}
	loadedEnd, err = mem.Load(endptr, bytecode.TypePtr, target.PointerAlign)
	if err != nil {
		t.Fatalf("load strtof subnormal endptr: %v", err)
	}
	if loadedEnd.Int != floatSubnormalText+5 {
		t.Fatalf("strtof subnormal endptr=%#x, want %#x", loadedEnd.Int, floatSubnormalText+5)
	}
	errnoValue, err = mem.Load(errnoAddr, bytecode.TypeI32, 4)
	if err != nil {
		t.Fatalf("load errno after strtof subnormal: %v", err)
	}
	if signedInt(errnoValue) != 0 {
		t.Fatalf("errno after strtof subnormal=%#v, want 0", errnoValue)
	}

	strtoldFn, ok := reg.Lookup("strtold")
	if !ok {
		t.Fatal("missing strtold extern")
	}
	longText := mustAllocBytes(t, mem, "strtold:hex", []byte("-0x1.4p+2z\x00"), true, blockString)
	ret, exit, err = strtoldFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(longText), PtrValue(endptr)})
	if err != nil || exit != nil {
		t.Fatalf("strtold ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeFLong || ret.Float != -5 {
		t.Fatalf("strtold ret=%#v, want flong -5", ret)
	}
	loadedEnd, err = mem.Load(endptr, bytecode.TypePtr, target.PointerAlign)
	if err != nil {
		t.Fatalf("load strtold endptr: %v", err)
	}
	if loadedEnd.Int != longText+9 {
		t.Fatalf("strtold endptr=%#x, want %#x", loadedEnd.Int, longText+9)
	}

	if err := mem.Store(errnoAddr, bytecode.TypeI32, 4, IntValue(bytecode.TypeI32, 0)); err != nil {
		t.Fatalf("store errno before strtold overflow: %v", err)
	}
	longOverflowText := mustAllocBytes(t, mem, "strtold:overflow", []byte("1e309!\x00"), true, blockString)
	ret, exit, err = strtoldFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(longOverflowText), PtrValue(endptr)})
	if err != nil || exit != nil {
		t.Fatalf("strtold overflow ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeFLong || !math.IsInf(ret.Float, 1) {
		t.Fatalf("strtold overflow ret=%#v, want flong +inf", ret)
	}
	loadedEnd, err = mem.Load(endptr, bytecode.TypePtr, target.PointerAlign)
	if err != nil {
		t.Fatalf("load strtold overflow endptr: %v", err)
	}
	if loadedEnd.Int != longOverflowText+5 {
		t.Fatalf("strtold overflow endptr=%#x, want %#x", loadedEnd.Int, longOverflowText+5)
	}
	errnoValue, err = mem.Load(errnoAddr, bytecode.TypeI32, 4)
	if err != nil {
		t.Fatalf("load errno after strtold overflow: %v", err)
	}
	if signedInt(errnoValue) != 34 {
		t.Fatalf("errno after strtold overflow=%#v, want ERANGE", errnoValue)
	}

	if err := mem.Store(errnoAddr, bytecode.TypeI32, 4, IntValue(bytecode.TypeI32, 0)); err != nil {
		t.Fatalf("store errno before strtold negative overflow: %v", err)
	}
	negativeLongOverflowText := mustAllocBytes(t, mem, "strtold:negative-overflow", []byte("-1e309?\x00"), true, blockString)
	ret, exit, err = strtoldFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(negativeLongOverflowText), PtrValue(endptr)})
	if err != nil || exit != nil {
		t.Fatalf("strtold negative overflow ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeFLong || !math.IsInf(ret.Float, -1) {
		t.Fatalf("strtold negative overflow ret=%#v, want flong -inf", ret)
	}
	loadedEnd, err = mem.Load(endptr, bytecode.TypePtr, target.PointerAlign)
	if err != nil {
		t.Fatalf("load strtold negative overflow endptr: %v", err)
	}
	if loadedEnd.Int != negativeLongOverflowText+6 {
		t.Fatalf("strtold negative overflow endptr=%#x, want %#x", loadedEnd.Int, negativeLongOverflowText+6)
	}
	errnoValue, err = mem.Load(errnoAddr, bytecode.TypeI32, 4)
	if err != nil {
		t.Fatalf("load errno after strtold negative overflow: %v", err)
	}
	if signedInt(errnoValue) != 34 {
		t.Fatalf("errno after strtold negative overflow=%#v, want ERANGE", errnoValue)
	}

	if err := mem.Store(errnoAddr, bytecode.TypeI32, 4, IntValue(bytecode.TypeI32, 0)); err != nil {
		t.Fatalf("store errno before strtold underflow: %v", err)
	}
	longUnderflowText := mustAllocBytes(t, mem, "strtold:underflow", []byte("1e-400!\x00"), true, blockString)
	ret, exit, err = strtoldFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(longUnderflowText), PtrValue(endptr)})
	if err != nil || exit != nil {
		t.Fatalf("strtold underflow ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeFLong || ret.Float != 0 || math.Signbit(ret.Float) {
		t.Fatalf("strtold underflow ret=%#v, want flong +0", ret)
	}
	loadedEnd, err = mem.Load(endptr, bytecode.TypePtr, target.PointerAlign)
	if err != nil {
		t.Fatalf("load strtold underflow endptr: %v", err)
	}
	if loadedEnd.Int != longUnderflowText+6 {
		t.Fatalf("strtold underflow endptr=%#x, want %#x", loadedEnd.Int, longUnderflowText+6)
	}
	errnoValue, err = mem.Load(errnoAddr, bytecode.TypeI32, 4)
	if err != nil {
		t.Fatalf("load errno after strtold underflow: %v", err)
	}
	if signedInt(errnoValue) != 34 {
		t.Fatalf("errno after strtold underflow=%#v, want ERANGE", errnoValue)
	}

	if err := mem.Store(errnoAddr, bytecode.TypeI32, 4, IntValue(bytecode.TypeI32, 0)); err != nil {
		t.Fatalf("store errno before strtold negative underflow: %v", err)
	}
	negativeLongUnderflowText := mustAllocBytes(t, mem, "strtold:negative-underflow", []byte("-1e-400?\x00"), true, blockString)
	ret, exit, err = strtoldFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(negativeLongUnderflowText), PtrValue(endptr)})
	if err != nil || exit != nil {
		t.Fatalf("strtold negative underflow ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeFLong || ret.Float != 0 || !math.Signbit(ret.Float) {
		t.Fatalf("strtold negative underflow ret=%#v, want flong -0", ret)
	}
	loadedEnd, err = mem.Load(endptr, bytecode.TypePtr, target.PointerAlign)
	if err != nil {
		t.Fatalf("load strtold negative underflow endptr: %v", err)
	}
	if loadedEnd.Int != negativeLongUnderflowText+7 {
		t.Fatalf("strtold negative underflow endptr=%#x, want %#x", loadedEnd.Int, negativeLongUnderflowText+7)
	}
	errnoValue, err = mem.Load(errnoAddr, bytecode.TypeI32, 4)
	if err != nil {
		t.Fatalf("load errno after strtold negative underflow: %v", err)
	}
	if signedInt(errnoValue) != 34 {
		t.Fatalf("errno after strtold negative underflow=%#v, want ERANGE", errnoValue)
	}

	if err := mem.Store(errnoAddr, bytecode.TypeI32, 4, IntValue(bytecode.TypeI32, 0)); err != nil {
		t.Fatalf("store errno before strtold subnormal: %v", err)
	}
	longSubnormalText := mustAllocBytes(t, mem, "strtold:subnormal", []byte("5e-324!\x00"), true, blockString)
	ret, exit, err = strtoldFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(longSubnormalText), PtrValue(endptr)})
	if err != nil || exit != nil {
		t.Fatalf("strtold subnormal ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeFLong || ret.Float != 5e-324 {
		t.Fatalf("strtold subnormal ret=%#v, want smallest binary64 subnormal", ret)
	}
	loadedEnd, err = mem.Load(endptr, bytecode.TypePtr, target.PointerAlign)
	if err != nil {
		t.Fatalf("load strtold subnormal endptr: %v", err)
	}
	if loadedEnd.Int != longSubnormalText+6 {
		t.Fatalf("strtold subnormal endptr=%#x, want %#x", loadedEnd.Int, longSubnormalText+6)
	}
	errnoValue, err = mem.Load(errnoAddr, bytecode.TypeI32, 4)
	if err != nil {
		t.Fatalf("load errno after strtold subnormal: %v", err)
	}
	if signedInt(errnoValue) != 0 {
		t.Fatalf("errno after strtold subnormal=%#v, want 0", errnoValue)
	}
}

func TestCtypeClassificationExterns(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	tests := []struct {
		name string
		ch   int64
		want bool
	}{
		{name: "isdigit", ch: '5', want: true},
		{name: "isdigit", ch: 'x', want: false},
		{name: "isalpha", ch: 'Z', want: true},
		{name: "isalnum", ch: '7', want: true},
		{name: "isspace", ch: '\n', want: true},
		{name: "islower", ch: 'q', want: true},
		{name: "isupper", ch: 'Q', want: true},
		{name: "isxdigit", ch: 'f', want: true},
		{name: "isxdigit", ch: 'G', want: false},
		{name: "isprint", ch: ' ', want: true},
		{name: "isprint", ch: 0x7f, want: false},
		{name: "isalpha", ch: -1, want: false},
		{name: "isalpha", ch: 0x141, want: true},
		{name: "isdigit", ch: 0x135, want: true},
	}
	for _, tt := range tests {
		fn, ok := reg.Lookup(tt.name)
		if !ok {
			t.Fatalf("missing %s extern", tt.name)
		}
		ret, exit, err := fn(context.Background(), nil, []Value{IntValue(bytecode.TypeI32, tt.ch)})
		if err != nil || exit != nil {
			t.Fatalf("%s(%d) ret=%#v exit=%#v err=%v", tt.name, tt.ch, ret, exit, err)
		}
		got := ret.Int != 0
		if ret.Type != bytecode.TypeI32 || got != tt.want {
			t.Fatalf("%s(%d) ret=%#v, want bool %v", tt.name, tt.ch, ret, tt.want)
		}
	}
}

func TestWideCtypeClassificationExterns(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	tests := []struct {
		name string
		ch   int64
		want bool
	}{
		{name: "iswdigit", ch: '5', want: true},
		{name: "iswdigit", ch: 'x', want: false},
		{name: "iswalpha", ch: 'Z', want: true},
		{name: "iswalnum", ch: '7', want: true},
		{name: "iswspace", ch: '\n', want: true},
		{name: "iswlower", ch: 'q', want: true},
		{name: "iswupper", ch: 'Q', want: true},
		{name: "iswxdigit", ch: 'f', want: true},
		{name: "iswxdigit", ch: 'G', want: false},
		{name: "iswprint", ch: ' ', want: true},
		{name: "iswprint", ch: 0x7f, want: false},
		{name: "iswalpha", ch: -1, want: false},
		{name: "iswalpha", ch: 0x141, want: false},
		{name: "iswdigit", ch: 0x135, want: false},
	}
	for _, tt := range tests {
		fn, ok := reg.Lookup(tt.name)
		if !ok {
			t.Fatalf("missing %s extern", tt.name)
		}
		ret, exit, err := fn(context.Background(), nil, []Value{IntValue(bytecode.TypeI32, tt.ch)})
		if err != nil || exit != nil {
			t.Fatalf("%s(%d) ret=%#v exit=%#v err=%v", tt.name, tt.ch, ret, exit, err)
		}
		got := ret.Int != 0
		if ret.Type != bytecode.TypeI32 || got != tt.want {
			t.Fatalf("%s(%d) ret=%#v, want bool %v", tt.name, tt.ch, ret, tt.want)
		}
	}
}

func TestCtypeExtraClassificationExterns(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	tests := []struct {
		name string
		ch   int64
		want bool
	}{
		{name: "isblank", ch: ' ', want: true},
		{name: "isblank", ch: '\t', want: true},
		{name: "isblank", ch: '\n', want: false},
		{name: "iscntrl", ch: 0x1f, want: true},
		{name: "iscntrl", ch: 0x7f, want: true},
		{name: "iscntrl", ch: 'A', want: false},
		{name: "isgraph", ch: '!', want: true},
		{name: "isgraph", ch: ' ', want: false},
		{name: "ispunct", ch: '!', want: true},
		{name: "ispunct", ch: 'A', want: false},
		{name: "ispunct", ch: ' ', want: false},
	}
	for _, tt := range tests {
		fn, ok := reg.Lookup(tt.name)
		if !ok {
			t.Fatalf("missing %s extern", tt.name)
		}
		ret, exit, err := fn(context.Background(), nil, []Value{IntValue(bytecode.TypeI32, tt.ch)})
		if err != nil || exit != nil {
			t.Fatalf("%s(%d) ret=%#v exit=%#v err=%v", tt.name, tt.ch, ret, exit, err)
		}
		got := ret.Int != 0
		if ret.Type != bytecode.TypeI32 || got != tt.want {
			t.Fatalf("%s(%d) ret=%#v, want bool %v", tt.name, tt.ch, ret, tt.want)
		}
	}
}

func TestWideCtypeExtraClassificationExterns(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	tests := []struct {
		name string
		ch   int64
		want bool
	}{
		{name: "iswblank", ch: ' ', want: true},
		{name: "iswblank", ch: '\t', want: true},
		{name: "iswblank", ch: '\n', want: false},
		{name: "iswcntrl", ch: 0x1f, want: true},
		{name: "iswcntrl", ch: 0x7f, want: true},
		{name: "iswcntrl", ch: 'A', want: false},
		{name: "iswgraph", ch: '!', want: true},
		{name: "iswgraph", ch: ' ', want: false},
		{name: "iswpunct", ch: '!', want: true},
		{name: "iswpunct", ch: 'A', want: false},
		{name: "iswpunct", ch: ' ', want: false},
		{name: "iswpunct", ch: 0x121, want: false},
	}
	for _, tt := range tests {
		fn, ok := reg.Lookup(tt.name)
		if !ok {
			t.Fatalf("missing %s extern", tt.name)
		}
		ret, exit, err := fn(context.Background(), nil, []Value{IntValue(bytecode.TypeI32, tt.ch)})
		if err != nil || exit != nil {
			t.Fatalf("%s(%d) ret=%#v exit=%#v err=%v", tt.name, tt.ch, ret, exit, err)
		}
		got := ret.Int != 0
		if ret.Type != bytecode.TypeI32 || got != tt.want {
			t.Fatalf("%s(%d) ret=%#v, want bool %v", tt.name, tt.ch, ret, tt.want)
		}
	}
}

func TestCtypeCaseConversionExterns(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	tests := []struct {
		name string
		ch   int64
		want int64
	}{
		{name: "tolower", ch: 'A', want: 'a'},
		{name: "tolower", ch: 'z', want: 'z'},
		{name: "tolower", ch: '!', want: '!'},
		{name: "tolower", ch: -1, want: -1},
		{name: "tolower", ch: 0x141, want: 'a'},
		{name: "toupper", ch: 'q', want: 'Q'},
		{name: "toupper", ch: 'Z', want: 'Z'},
		{name: "toupper", ch: '!', want: '!'},
		{name: "toupper", ch: -1, want: -1},
		{name: "toupper", ch: 0x171, want: 'Q'},
	}
	for _, tt := range tests {
		fn, ok := reg.Lookup(tt.name)
		if !ok {
			t.Fatalf("missing %s extern", tt.name)
		}
		ret, exit, err := fn(context.Background(), nil, []Value{IntValue(bytecode.TypeI32, tt.ch)})
		if err != nil || exit != nil {
			t.Fatalf("%s(%d) ret=%#v exit=%#v err=%v", tt.name, tt.ch, ret, exit, err)
		}
		if ret.Type != bytecode.TypeI32 || signedInt(ret) != tt.want {
			t.Fatalf("%s(%d) ret=%#v, want i32 %d", tt.name, tt.ch, ret, tt.want)
		}
	}
}

func TestWideCtypeCaseConversionExterns(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	tests := []struct {
		name string
		ch   int64
		want int64
	}{
		{name: "towlower", ch: 'A', want: 'a'},
		{name: "towlower", ch: 'z', want: 'z'},
		{name: "towlower", ch: '!', want: '!'},
		{name: "towlower", ch: -1, want: -1},
		{name: "towlower", ch: 0x141, want: 0x141},
		{name: "towupper", ch: 'q', want: 'Q'},
		{name: "towupper", ch: 'Z', want: 'Z'},
		{name: "towupper", ch: '!', want: '!'},
		{name: "towupper", ch: -1, want: -1},
		{name: "towupper", ch: 0x171, want: 0x171},
	}
	for _, tt := range tests {
		fn, ok := reg.Lookup(tt.name)
		if !ok {
			t.Fatalf("missing %s extern", tt.name)
		}
		ret, exit, err := fn(context.Background(), nil, []Value{IntValue(bytecode.TypeI32, tt.ch)})
		if err != nil || exit != nil {
			t.Fatalf("%s(%d) ret=%#v exit=%#v err=%v", tt.name, tt.ch, ret, exit, err)
		}
		if ret.Type != bytecode.TypeI32 || signedInt(ret) != tt.want {
			t.Fatalf("%s(%d) ret=%#v, want i32 %d", tt.name, tt.ch, ret, tt.want)
		}
	}
}

func TestWideCtypeDescriptorExterns(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	wctypeFn, ok := reg.Lookup("wctype")
	if !ok {
		t.Fatal("missing wctype extern")
	}
	iswctypeFn, ok := reg.Lookup("iswctype")
	if !ok {
		t.Fatal("missing iswctype extern")
	}
	alphaName := mustAllocBytes(t, mem, "wctype:alpha", []byte("alpha\x00"), true, blockString)
	digitName := mustAllocBytes(t, mem, "wctype:digit", []byte("digit\x00"), true, blockString)
	unknownName := mustAllocBytes(t, mem, "wctype:unknown", []byte("emoji\x00"), true, blockString)

	alpha, exit, err := wctypeFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(alphaName)})
	if err != nil || exit != nil || alpha.Type != bytecode.TypeU64 || alpha.Int == 0 {
		t.Fatalf("wctype alpha ret=%#v exit=%#v err=%v, want nonzero descriptor", alpha, exit, err)
	}
	digit, exit, err := wctypeFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(digitName)})
	if err != nil || exit != nil || digit.Type != bytecode.TypeU64 || digit.Int == 0 {
		t.Fatalf("wctype digit ret=%#v exit=%#v err=%v, want nonzero descriptor", digit, exit, err)
	}
	unknown, exit, err := wctypeFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(unknownName)})
	if err != nil || exit != nil || unknown.Type != bytecode.TypeU64 || unknown.Int != 0 {
		t.Fatalf("wctype unknown ret=%#v exit=%#v err=%v, want zero descriptor", unknown, exit, err)
	}

	tests := []struct {
		ch   int64
		desc Value
		want bool
	}{
		{ch: 'A', desc: alpha, want: true},
		{ch: '!', desc: alpha, want: false},
		{ch: 0x141, desc: alpha, want: false},
		{ch: -1, desc: alpha, want: false},
		{ch: '7', desc: digit, want: true},
		{ch: 'x', desc: digit, want: false},
		{ch: 'A', desc: unknown, want: false},
	}
	for _, tt := range tests {
		ret, exit, err := iswctypeFn(context.Background(), &ExternContext{Memory: mem}, []Value{IntValue(bytecode.TypeI32, tt.ch), tt.desc})
		if err != nil || exit != nil {
			t.Fatalf("iswctype(%d, %#v) ret=%#v exit=%#v err=%v", tt.ch, tt.desc, ret, exit, err)
		}
		got := ret.Int != 0
		if ret.Type != bytecode.TypeI32 || got != tt.want {
			t.Fatalf("iswctype(%d, %#v) ret=%#v, want %v", tt.ch, tt.desc, ret, tt.want)
		}
	}
}

func TestWideCtypeTransDescriptorExterns(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	wctransFn, ok := reg.Lookup("wctrans")
	if !ok {
		t.Fatal("missing wctrans extern")
	}
	towctransFn, ok := reg.Lookup("towctrans")
	if !ok {
		t.Fatal("missing towctrans extern")
	}
	lowerName := mustAllocBytes(t, mem, "wctrans:tolower", []byte("tolower\x00"), true, blockString)
	upperName := mustAllocBytes(t, mem, "wctrans:toupper", []byte("toupper\x00"), true, blockString)
	unknownName := mustAllocBytes(t, mem, "wctrans:unknown", []byte("swapcase\x00"), true, blockString)

	lower, exit, err := wctransFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(lowerName)})
	if err != nil || exit != nil || lower.Type != bytecode.TypeU64 || lower.Int == 0 {
		t.Fatalf("wctrans lower ret=%#v exit=%#v err=%v, want nonzero descriptor", lower, exit, err)
	}
	upper, exit, err := wctransFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(upperName)})
	if err != nil || exit != nil || upper.Type != bytecode.TypeU64 || upper.Int == 0 {
		t.Fatalf("wctrans upper ret=%#v exit=%#v err=%v, want nonzero descriptor", upper, exit, err)
	}
	unknown, exit, err := wctransFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(unknownName)})
	if err != nil || exit != nil || unknown.Type != bytecode.TypeU64 || unknown.Int != 0 {
		t.Fatalf("wctrans unknown ret=%#v exit=%#v err=%v, want zero descriptor", unknown, exit, err)
	}

	tests := []struct {
		ch   int64
		desc Value
		want int64
	}{
		{ch: 'A', desc: lower, want: 'a'},
		{ch: 'z', desc: lower, want: 'z'},
		{ch: 'q', desc: upper, want: 'Q'},
		{ch: 'Z', desc: upper, want: 'Z'},
		{ch: '!', desc: upper, want: '!'},
		{ch: -1, desc: lower, want: -1},
		{ch: 0x141, desc: lower, want: 0x141},
		{ch: 'A', desc: unknown, want: 'A'},
	}
	for _, tt := range tests {
		ret, exit, err := towctransFn(context.Background(), &ExternContext{Memory: mem}, []Value{IntValue(bytecode.TypeI32, tt.ch), tt.desc})
		if err != nil || exit != nil {
			t.Fatalf("towctrans(%d, %#v) ret=%#v exit=%#v err=%v", tt.ch, tt.desc, ret, exit, err)
		}
		if ret.Type != bytecode.TypeI32 || signedInt(ret) != tt.want {
			t.Fatalf("towctrans(%d, %#v) ret=%#v, want %d", tt.ch, tt.desc, ret, tt.want)
		}
	}
}

func TestPlainMemoryOperationExterns(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	buf := mustAllocBytes(t, mem, "plain-memory:buf", []byte("abcdef\x00"), false, blockLocal)
	memsetFn, ok := reg.Lookup("memset")
	if !ok {
		t.Fatal("missing memset extern")
	}
	ret, exit, err := memsetFn(context.Background(), &ExternContext{Memory: mem}, []Value{
		PtrValue(0xdeadbeef),
		IntValue(bytecode.TypeI32, 0x182),
		UIntValue(bytecode.TypeU64, 0),
	})
	if err != nil || exit != nil {
		t.Fatalf("memset zero ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypePtr || ret.Int != 0xdeadbeef {
		t.Fatalf("memset zero ret=%#v, want original pointer", ret)
	}
	ret, exit, err = memsetFn(context.Background(), &ExternContext{Memory: mem}, []Value{
		ObjectAddrValue(buf + 4),
		IntValue(bytecode.TypeI32, 0x182),
		UIntValue(bytecode.TypeU64, 2),
	})
	if err != nil || exit != nil {
		t.Fatalf("memset masked ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypePtr || ret.Int != buf+4 {
		t.Fatalf("memset masked ret=%#v, want pointer %#x", ret, buf+4)
	}
	block, off, err := mem.rangeAccess(buf, 7, false)
	if err != nil {
		t.Fatalf("read memset buffer: %v", err)
	}
	if got := []byte(block.data[off+4 : off+6]); !bytes.Equal(got, []byte{0x82, 0x82}) {
		t.Fatalf("memset masked bytes=%v, want [130 130]", got)
	}

	fn, ok := reg.Lookup("bzero")
	if !ok {
		t.Fatal("missing bzero extern")
	}
	ret, exit, err = fn(context.Background(), &ExternContext{Memory: mem}, []Value{
		PtrValue(0xdeadbeef),
		UIntValue(bytecode.TypeU64, 0),
	})
	if err != nil || exit != nil {
		t.Fatalf("bzero zero ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeVoid {
		t.Fatalf("bzero zero ret=%#v, want void value", ret)
	}
	ret, exit, err = fn(context.Background(), &ExternContext{Memory: mem}, []Value{
		ObjectAddrValue(buf + 2),
		UIntValue(bytecode.TypeU64, 2),
	})
	if err != nil || exit != nil {
		t.Fatalf("bzero ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeVoid {
		t.Fatalf("bzero ret=%#v, want void value", ret)
	}
	got, err := mem.ReadCString(buf)
	if err != nil {
		t.Fatalf("ReadCString: %v", err)
	}
	if got != "ab" {
		t.Fatalf("bzero buffer prefix = %q, want ab", got)
	}
}

func TestMemoryCharCopyExtern(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	src := mustAllocBytes(t, mem, "memccpy:src", []byte("abczef\x00"), true, blockString)
	dst := mustAllocBytes(t, mem, "memccpy:dst", []byte("xxxxxx\x00"), false, blockLocal)
	fn, ok := reg.Lookup("memccpy")
	if !ok {
		t.Fatal("missing memccpy extern")
	}

	ret, exit, err := fn(context.Background(), &ExternContext{Memory: mem}, []Value{
		PtrValue(dst),
		PtrValue(src),
		IntValue(bytecode.TypeI32, 'z'),
		UIntValue(bytecode.TypeU64, 6),
	})
	if err != nil || exit != nil {
		t.Fatalf("memccpy hit ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypePtr || ret.Int != dst+4 {
		t.Fatalf("memccpy hit ret=%#v, want pointer %#x", ret, dst+4)
	}
	block, off, err := mem.rangeAccess(dst, 6, false)
	if err != nil {
		t.Fatalf("read memccpy dst: %v", err)
	}
	if got := string(block.data[off : off+6]); got != "abczxx" {
		t.Fatalf("memccpy dst after hit = %q, want abczxx", got)
	}
	if err := mem.Store(dst+3, bytecode.TypeI8, 1, IntValue(bytecode.TypeI8, 'Z')); err != nil {
		t.Fatalf("mark memccpy dst: %v", err)
	}

	ret, exit, err = fn(context.Background(), &ExternContext{Memory: mem}, []Value{
		PtrValue(dst),
		PtrValue(src),
		IntValue(bytecode.TypeI32, 'q'),
		UIntValue(bytecode.TypeU64, 3),
	})
	if err != nil || exit != nil {
		t.Fatalf("memccpy miss ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypePtr || ret.Int != 0 {
		t.Fatalf("memccpy miss ret=%#v, want null", ret)
	}
	block, off, err = mem.rangeAccess(dst, 6, false)
	if err != nil {
		t.Fatalf("read memccpy miss dst: %v", err)
	}
	if got := string(block.data[off : off+6]); got != "abcZxx" {
		t.Fatalf("memccpy dst after miss = %q, want abcZxx", got)
	}

	highSrc := mustAllocBytes(t, mem, "memccpy:high-src", []byte{1, 0x82, 3}, true, blockString)
	highDst := mustAllocBytes(t, mem, "memccpy:high-dst", []byte{9, 9, 9}, false, blockLocal)
	ret, exit, err = fn(context.Background(), &ExternContext{Memory: mem}, []Value{
		PtrValue(highDst),
		PtrValue(highSrc),
		IntValue(bytecode.TypeI32, 0x182),
		UIntValue(bytecode.TypeU64, 3),
	})
	if err != nil || exit != nil {
		t.Fatalf("memccpy masked ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypePtr || ret.Int != highDst+2 {
		t.Fatalf("memccpy masked ret=%#v, want pointer %#x", ret, highDst+2)
	}
	block, off, err = mem.rangeAccess(highDst, 3, false)
	if err != nil {
		t.Fatalf("read memccpy masked dst: %v", err)
	}
	if got := []byte(block.data[off : off+3]); !bytes.Equal(got, []byte{1, 0x82, 9}) {
		t.Fatalf("memccpy masked dst after hit = %v, want [1 130 9]", got)
	}
}

func TestStringsBSDMemoryExterns(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	left := mustAllocBytes(t, mem, "bcmp:left", []byte("abcd\x00"), true, blockString)
	right := mustAllocBytes(t, mem, "bcmp:right", []byte("abce\x00"), true, blockString)
	dst := mustAllocBytes(t, mem, "bcopy:dst", []byte("xxxx\x00"), false, blockGlobal)

	bcmpFn, ok := reg.Lookup("bcmp")
	if !ok {
		t.Fatal("missing bcmp extern")
	}
	ret, exit, err := bcmpFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(left), PtrValue(right), UIntValue(bytecode.TypeU64, 4)})
	if err != nil || exit != nil {
		t.Fatalf("bcmp ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeI32 || signedInt(ret) >= 0 {
		t.Fatalf("bcmp ret=%#v, want negative i32", ret)
	}

	bcopyFn, ok := reg.Lookup("bcopy")
	if !ok {
		t.Fatal("missing bcopy extern")
	}
	ret, exit, err = bcopyFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(left), PtrValue(dst), UIntValue(bytecode.TypeU64, 4)})
	if err != nil || exit != nil {
		t.Fatalf("bcopy ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeVoid {
		t.Fatalf("bcopy ret=%#v, want void", ret)
	}
	got, err := mem.ReadCString(dst)
	if err != nil {
		t.Fatalf("ReadCString(dst): %v", err)
	}
	if got != "abcd" {
		t.Fatalf("bcopy dst=%q, want abcd", got)
	}
	overlapForward := mustAllocBytes(t, mem, "bcopy:overlap-forward", []byte("abcdef\x00"), false, blockGlobal)
	ret, exit, err = bcopyFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(overlapForward), PtrValue(overlapForward + 1), UIntValue(bytecode.TypeU64, 5)})
	if err != nil || exit != nil {
		t.Fatalf("bcopy overlap forward ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	got, err = mem.ReadCString(overlapForward)
	if err != nil || got != "aabcde" {
		t.Fatalf("bcopy overlap forward dst=%q err=%v, want aabcde", got, err)
	}
	overlapBackward := mustAllocBytes(t, mem, "bcopy:overlap-backward", []byte("abcdef\x00"), false, blockGlobal)
	ret, exit, err = bcopyFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(overlapBackward + 1), PtrValue(overlapBackward), UIntValue(bytecode.TypeU64, 5)})
	if err != nil || exit != nil {
		t.Fatalf("bcopy overlap backward ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	got, err = mem.ReadCString(overlapBackward)
	if err != nil || got != "bcdeff" {
		t.Fatalf("bcopy overlap backward dst=%q err=%v, want bcdeff", got, err)
	}

	bzeroFn, ok := reg.Lookup("bzero")
	if !ok {
		t.Fatal("missing bzero extern")
	}
	ret, exit, err = bzeroFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(dst + 2), UIntValue(bytecode.TypeU64, 2)})
	if err != nil || exit != nil {
		t.Fatalf("bzero ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	zero, err := mem.Load(dst+2, bytecode.TypeI8, 1)
	if err != nil {
		t.Fatalf("load bzero dst[2]: %v", err)
	}
	if zero.Int != 0 {
		t.Fatalf("bzero dst[2]=%#v, want zero", zero)
	}
}

func TestPlainStringWriteExterns(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	for _, name := range []string{"strcpy", "stpcpy", "strcat", "strncpy", "stpncpy", "strncat"} {
		if _, ok := reg.Lookup(name); !ok {
			t.Fatalf("missing %s extern", name)
		}
	}
	mem := NewMemory(bytecode.DefaultTarget())
	buf := mustAllocBytes(t, mem, "strcpy:buf", make([]byte, 8), false, blockLocal)
	src := mustAllocBytes(t, mem, "strcpy:src", []byte("ab\x00"), true, blockString)
	strcpyFn, _ := reg.Lookup("strcpy")
	ret, exit, err := strcpyFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(buf), PtrValue(src)})
	if err != nil || exit != nil {
		t.Fatalf("strcpy ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypePtr || ret.Int != buf {
		t.Fatalf("strcpy ret=%#v, want pointer %#x", ret, buf)
	}
	got, err := mem.ReadCString(buf)
	if err != nil || got != "ab" {
		t.Fatalf("strcpy dst=%q err=%v, want ab", got, err)
	}
	nul, err := mem.Load(buf+2, bytecode.TypeI8, 1)
	if err != nil || nul.Int != 0 {
		t.Fatalf("strcpy terminator=%#v err=%v, want zero", nul, err)
	}

	stpcpyFn, _ := reg.Lookup("stpcpy")
	src2 := mustAllocBytes(t, mem, "stpcpy:src", []byte("cd\x00"), true, blockString)
	ret, exit, err = stpcpyFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(buf + 2), PtrValue(src2)})
	if err != nil || exit != nil {
		t.Fatalf("stpcpy ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypePtr || ret.Int != buf+4 {
		t.Fatalf("stpcpy ret=%#v, want pointer %#x", ret, buf+4)
	}
	got, err = mem.ReadCString(buf)
	if err != nil || got != "abcd" {
		t.Fatalf("stpcpy dst=%q err=%v, want abcd", got, err)
	}
	nul, err = mem.Load(buf+4, bytecode.TypeI8, 1)
	if err != nil || nul.Int != 0 {
		t.Fatalf("stpcpy terminator=%#v err=%v, want zero", nul, err)
	}

	strncpyFn, _ := reg.Lookup("strncpy")
	copyBuf := mustAllocBytes(t, mem, "strncpy:buf", []byte("??????"), false, blockLocal)
	shortSrc := mustAllocBytes(t, mem, "strncpy:short", []byte("xy\x00"), true, blockString)
	ret, exit, err = strncpyFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(copyBuf), PtrValue(shortSrc), UIntValue(bytecode.TypeU64, 4)})
	if err != nil || exit != nil {
		t.Fatalf("strncpy ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypePtr || ret.Int != copyBuf {
		t.Fatalf("strncpy ret=%#v, want pointer %#x", ret, copyBuf)
	}
	block, off, err := mem.rangeAccess(copyBuf, 6, false)
	if err != nil {
		t.Fatalf("read strncpy dst: %v", err)
	}
	if got := []byte(block.data[off : off+6]); !bytes.Equal(got, []byte{'x', 'y', 0, 0, '?', '?'}) {
		t.Fatalf("strncpy dst=%v, want [120 121 0 0 63 63]", got)
	}

	stpncpyFn, _ := reg.Lookup("stpncpy")
	truncSrc := mustAllocBytes(t, mem, "stpncpy:trunc", []byte("pqrs\x00"), true, blockString)
	if err := mem.Store(copyBuf+2, bytecode.TypeI8, 1, IntValue(bytecode.TypeI8, '!')); err != nil {
		t.Fatalf("mark stpncpy trunc sentinel: %v", err)
	}
	ret, exit, err = stpncpyFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(copyBuf), PtrValue(truncSrc), UIntValue(bytecode.TypeU64, 2)})
	if err != nil || exit != nil {
		t.Fatalf("stpncpy trunc ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypePtr || ret.Int != copyBuf+2 {
		t.Fatalf("stpncpy trunc ret=%#v, want pointer %#x", ret, copyBuf+2)
	}
	block, off, err = mem.rangeAccess(copyBuf, 4, false)
	if err != nil {
		t.Fatalf("read stpncpy trunc dst: %v", err)
	}
	if got := []byte(block.data[off : off+4]); !bytes.Equal(got, []byte{'p', 'q', '!', 0}) {
		t.Fatalf("stpncpy trunc dst=%v, want [112 113 33 0]", got)
	}

	padSrc := mustAllocBytes(t, mem, "stpncpy:pad", []byte("a\x00"), true, blockString)
	ret, exit, err = stpncpyFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(copyBuf), PtrValue(padSrc), UIntValue(bytecode.TypeU64, 3)})
	if err != nil || exit != nil {
		t.Fatalf("stpncpy pad ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypePtr || ret.Int != copyBuf+1 {
		t.Fatalf("stpncpy pad ret=%#v, want pointer %#x", ret, copyBuf+1)
	}
	block, off, err = mem.rangeAccess(copyBuf, 3, false)
	if err != nil {
		t.Fatalf("read stpncpy pad dst: %v", err)
	}
	if got := []byte(block.data[off : off+3]); !bytes.Equal(got, []byte{'a', 0, 0}) {
		t.Fatalf("stpncpy pad dst=%v, want [97 0 0]", got)
	}

	catBuf := mustAllocBytes(t, mem, "strcat:buf", []byte("ab\x00?????"), false, blockLocal)
	catSrc := mustAllocBytes(t, mem, "strcat:src", []byte("cd\x00"), true, blockString)
	strcatFn, _ := reg.Lookup("strcat")
	ret, exit, err = strcatFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(catBuf), PtrValue(catSrc)})
	if err != nil || exit != nil {
		t.Fatalf("strcat ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypePtr || ret.Int != catBuf {
		t.Fatalf("strcat ret=%#v, want pointer %#x", ret, catBuf)
	}
	got, err = mem.ReadCString(catBuf)
	if err != nil || got != "abcd" {
		t.Fatalf("strcat dst=%q err=%v, want abcd", got, err)
	}
	nul, err = mem.Load(catBuf+4, bytecode.TypeI8, 1)
	if err != nil || nul.Int != 0 {
		t.Fatalf("strcat terminator=%#v err=%v, want zero", nul, err)
	}

	strncatFn, _ := reg.Lookup("strncat")
	longSrc := mustAllocBytes(t, mem, "strncat:src", []byte("efgh\x00"), true, blockString)
	ret, exit, err = strncatFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(catBuf), PtrValue(longSrc), UIntValue(bytecode.TypeU64, 2)})
	if err != nil || exit != nil {
		t.Fatalf("strncat ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypePtr || ret.Int != catBuf {
		t.Fatalf("strncat ret=%#v, want pointer %#x", ret, catBuf)
	}
	got, err = mem.ReadCString(catBuf)
	if err != nil || got != "abcdef" {
		t.Fatalf("strncat dst=%q err=%v, want abcdef", got, err)
	}
	ret, exit, err = strncatFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(catBuf), PtrValue(longSrc), UIntValue(bytecode.TypeU64, 0)})
	if err != nil || exit != nil {
		t.Fatalf("strncat zero ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypePtr || ret.Int != catBuf {
		t.Fatalf("strncat zero ret=%#v, want pointer %#x", ret, catBuf)
	}
	got, err = mem.ReadCString(catBuf)
	if err != nil || got != "abcdef" {
		t.Fatalf("strncat zero dst=%q err=%v, want abcdef", got, err)
	}
}

func TestCheckedStringMemoryExterns(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	dst := mustAllocBytes(t, mem, "checked:dst", make([]byte, 16), false, blockLocal)
	src := mustAllocBytes(t, mem, "checked:src", []byte("abcd\x00"), true, blockString)
	wantObjectSizeErr := func(name string, err error) {
		t.Helper()
		if err == nil || !strings.Contains(err.Error(), "object size") {
			t.Fatalf("%s err=%v, want object size error", name, err)
		}
	}

	memcpyFn, ok := reg.Lookup("__builtin___memcpy_chk")
	if !ok {
		t.Fatal("missing __builtin___memcpy_chk extern")
	}
	ret, exit, err := memcpyFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(dst), PtrValue(src), UIntValue(bytecode.TypeU64, 5), UIntValue(bytecode.TypeU64, 5)})
	if err != nil || exit != nil || ret.Type != bytecode.TypePtr || ret.Int != dst {
		t.Fatalf("memcpy_chk boundary ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	_, _, err = memcpyFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(dst), PtrValue(src), UIntValue(bytecode.TypeU64, 5), UIntValue(bytecode.TypeU64, 4)})
	wantObjectSizeErr("memcpy_chk overflow", err)

	memsetFn, ok := reg.Lookup("__builtin___memset_chk")
	if !ok {
		t.Fatal("missing __builtin___memset_chk extern")
	}
	ret, exit, err = memsetFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(dst + 5), IntValue(bytecode.TypeI32, 'z'), UIntValue(bytecode.TypeU64, 2), UIntValue(bytecode.TypeU64, 2)})
	if err != nil || exit != nil || ret.Type != bytecode.TypePtr || ret.Int != dst+5 {
		t.Fatalf("memset_chk boundary ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	_, _, err = memsetFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(dst + 5), IntValue(bytecode.TypeI32, 'z'), UIntValue(bytecode.TypeU64, 3), UIntValue(bytecode.TypeU64, 2)})
	wantObjectSizeErr("memset_chk overflow", err)

	strcpyFn, ok := reg.Lookup("__builtin___strcpy_chk")
	if !ok {
		t.Fatal("missing __builtin___strcpy_chk extern")
	}
	shortSrc := mustAllocBytes(t, mem, "checked:short-src", []byte("ab\x00"), true, blockString)
	ret, exit, err = strcpyFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(dst), PtrValue(shortSrc), UIntValue(bytecode.TypeU64, 3)})
	if err != nil || exit != nil || ret.Type != bytecode.TypePtr || ret.Int != dst {
		t.Fatalf("strcpy_chk boundary ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	_, _, err = strcpyFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(dst), PtrValue(shortSrc), UIntValue(bytecode.TypeU64, 2)})
	wantObjectSizeErr("strcpy_chk overflow", err)

	strncatFn, ok := reg.Lookup("__builtin___strncat_chk")
	if !ok {
		t.Fatal("missing __builtin___strncat_chk extern")
	}
	tail := mustAllocBytes(t, mem, "checked:tail", []byte("cdxx\x00"), true, blockString)
	ret, exit, err = strncatFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(dst), PtrValue(tail), UIntValue(bytecode.TypeU64, 2), UIntValue(bytecode.TypeU64, 5)})
	if err != nil || exit != nil || ret.Type != bytecode.TypePtr || ret.Int != dst {
		t.Fatalf("strncat_chk boundary ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	_, _, err = strncatFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(dst), PtrValue(tail), UIntValue(bytecode.TypeU64, 1), UIntValue(bytecode.TypeU64, 4)})
	wantObjectSizeErr("strncat_chk overflow", err)
}

func TestPlainAllocationExterns(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	mallocFn, ok := reg.Lookup("malloc")
	if !ok {
		t.Fatal("missing malloc extern")
	}
	ret, exit, err := mallocFn(context.Background(), &ExternContext{Memory: mem}, []Value{UIntValue(bytecode.TypeU64, 4)})
	if err != nil || exit != nil {
		t.Fatalf("malloc ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypePtr || ret.Int == 0 {
		t.Fatalf("malloc ret=%#v, want nonzero pointer", ret)
	}
	if err := mem.Store(ret.Int, bytecode.TypeI8, 1, IntValue(bytecode.TypeI8, 'm')); err != nil {
		t.Fatalf("Store malloc byte: %v", err)
	}
	freeFn, ok := reg.Lookup("free")
	if !ok {
		t.Fatal("missing free extern")
	}
	freeRet, exit, err := freeFn(context.Background(), &ExternContext{Memory: mem}, []Value{ret})
	if err != nil || exit != nil {
		t.Fatalf("free ret=%#v exit=%#v err=%v", freeRet, exit, err)
	}
	if freeRet.Type != bytecode.TypeVoid {
		t.Fatalf("free ret=%#v, want void value", freeRet)
	}
	if err := mem.Store(ret.Int, bytecode.TypeI8, 1, IntValue(bytecode.TypeI8, 'x')); err == nil || !strings.Contains(err.Error(), "use after free") {
		t.Fatalf("Store after free err=%v, want use after free", err)
	}

	callocFn, ok := reg.Lookup("calloc")
	if !ok {
		t.Fatal("missing calloc extern")
	}
	ret, exit, err = callocFn(context.Background(), &ExternContext{Memory: mem}, []Value{UIntValue(bytecode.TypeU64, 2), UIntValue(bytecode.TypeU64, 3)})
	if err != nil || exit != nil {
		t.Fatalf("calloc ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	zero, err := mem.Load(ret.Int+5, bytecode.TypeI8, 1)
	if err != nil {
		t.Fatalf("Load calloc byte: %v", err)
	}
	if zero.Int != 0 {
		t.Fatalf("calloc byte = %#v, want zero", zero)
	}

	strdupFn, ok := reg.Lookup("strdup")
	if !ok {
		t.Fatal("missing strdup extern")
	}
	src := mustAllocBytes(t, mem, "plain-allocation:src", []byte("hi\x00"), true, blockString)
	ret, exit, err = strdupFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(src)})
	if err != nil || exit != nil {
		t.Fatalf("strdup ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	got, err := mem.ReadCString(ret.Int)
	if err != nil {
		t.Fatalf("ReadCString strdup: %v", err)
	}
	if got != "hi" {
		t.Fatalf("strdup string = %q, want hi", got)
	}

	strndupFn, ok := reg.Lookup("strndup")
	if !ok {
		t.Fatal("missing strndup extern")
	}
	src = mustAllocBytes(t, mem, "plain-allocation:bounded-src", []byte("hello\x00"), true, blockString)
	ret, exit, err = strndupFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(src), UIntValue(bytecode.TypeU64, 2)})
	if err != nil || exit != nil {
		t.Fatalf("strndup ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	got, err = mem.ReadCString(ret.Int)
	if err != nil {
		t.Fatalf("ReadCString strndup: %v", err)
	}
	if got != "he" {
		t.Fatalf("strndup string = %q, want he", got)
	}
}

func TestStdlibReallocExtern(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	fn, ok := reg.Lookup("realloc")
	if !ok {
		t.Fatal("missing realloc extern")
	}

	p, err := mem.TryAlloc("realloc:old", 3, 1, false, blockGlobal)
	if err != nil {
		t.Fatalf("alloc old: %v", err)
	}
	if err := writeMemoryBytes(mem, p, []byte{'a', 'b', 'c'}); err != nil {
		t.Fatalf("write old: %v", err)
	}
	ret, exit, err := fn(context.Background(), &ExternContext{Memory: mem}, []Value{
		PtrValue(p),
		UIntValue(bytecode.TypeU64, 5),
	})
	if err != nil || exit != nil {
		t.Fatalf("realloc grow ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypePtr || ret.Int == 0 {
		t.Fatalf("realloc grow ret=%#v, want non-null pointer", ret)
	}
	block, off, err := mem.rangeAccess(ret.Int, 3, false)
	if err != nil {
		t.Fatalf("read grown block: %v", err)
	}
	if got := string(block.data[off : off+3]); got != "abc" {
		t.Fatalf("grown prefix = %q, want abc", got)
	}
	if _, _, err := mem.rangeAccess(p, 1, false); err == nil || !strings.Contains(err.Error(), "use after free") {
		t.Fatalf("old realloc block err = %v, want use after free", err)
	}

	ret, exit, err = fn(context.Background(), &ExternContext{Memory: mem}, []Value{
		PtrValue(0),
		UIntValue(bytecode.TypeU64, 4),
	})
	if err != nil || exit != nil {
		t.Fatalf("realloc null ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypePtr || ret.Int == 0 {
		t.Fatalf("realloc null ret=%#v, want allocated pointer", ret)
	}

	ret, exit, err = fn(context.Background(), &ExternContext{Memory: mem}, []Value{
		ret,
		UIntValue(bytecode.TypeU64, 0),
	})
	if err != nil || exit != nil {
		t.Fatalf("realloc zero ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypePtr || ret.Int != 0 {
		t.Fatalf("realloc zero ret=%#v, want null pointer", ret)
	}
}

func TestStdlibRandExterns(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	srandFn, ok := reg.Lookup("srand")
	if !ok {
		t.Fatal("missing srand extern")
	}
	randFn, ok := reg.Lookup("rand")
	if !ok {
		t.Fatal("missing rand extern")
	}

	defaultFirst, exit, err := randFn(context.Background(), nil, nil)
	if err != nil || exit != nil {
		t.Fatalf("rand default ret=%#v exit=%#v err=%v", defaultFirst, exit, err)
	}
	if _, exit, err := srandFn(context.Background(), nil, []Value{UIntValue(bytecode.TypeU32, 1)}); err != nil || exit != nil {
		t.Fatalf("srand default seed repeat exit=%#v err=%v", exit, err)
	}
	defaultAgain, exit, err := randFn(context.Background(), nil, nil)
	if err != nil || exit != nil {
		t.Fatalf("rand default repeat ret=%#v exit=%#v err=%v", defaultAgain, exit, err)
	}
	if defaultFirst.Type != bytecode.TypeI32 || defaultAgain.Int != defaultFirst.Int {
		t.Fatalf("rand default=%#v after srand(1)=%#v, want repeat", defaultFirst, defaultAgain)
	}

	if ret, exit, err := srandFn(context.Background(), nil, []Value{UIntValue(bytecode.TypeU32, 123)}); err != nil || exit != nil || ret.Type != bytecode.TypeVoid {
		t.Fatalf("srand ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	first, exit, err := randFn(context.Background(), nil, nil)
	if err != nil || exit != nil {
		t.Fatalf("rand first ret=%#v exit=%#v err=%v", first, exit, err)
	}
	second, exit, err := randFn(context.Background(), nil, nil)
	if err != nil || exit != nil {
		t.Fatalf("rand second ret=%#v exit=%#v err=%v", second, exit, err)
	}
	if first.Type != bytecode.TypeI32 || second.Type != bytecode.TypeI32 || first.Int > 32767 || second.Int > 32767 {
		t.Fatalf("rand values first=%#v second=%#v, want i32 values <= 32767", first, second)
	}
	if first.Int == second.Int {
		t.Fatalf("rand repeated value %#v", first)
	}

	if _, exit, err := srandFn(context.Background(), nil, []Value{UIntValue(bytecode.TypeU32, 123)}); err != nil || exit != nil {
		t.Fatalf("srand repeat exit=%#v err=%v", exit, err)
	}
	again, exit, err := randFn(context.Background(), nil, nil)
	if err != nil || exit != nil {
		t.Fatalf("rand again ret=%#v exit=%#v err=%v", again, exit, err)
	}
	if again.Int != first.Int {
		t.Fatalf("rand after reseed=%#v, want first %#v", again, first)
	}
}

func TestStdlibGetenvExtern(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	fn, ok := reg.Lookup("getenv")
	if !ok {
		t.Fatal("missing getenv extern")
	}
	name := mustAllocBytes(t, mem, "getenv:name", []byte("PATH\x00"), true, blockString)
	ret, exit, err := fn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(name)})
	if err != nil || exit != nil {
		t.Fatalf("getenv ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypePtr || ret.Int != 0 {
		t.Fatalf("getenv ret=%#v, want null pointer", ret)
	}
}

func TestStdlibGetenvExternReadsConfiguredEnvironment(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	reg.SetEnv("CVM_TEST", "configured-value")
	mem := NewMemory(bytecode.DefaultTarget())
	fn, ok := reg.Lookup("getenv")
	if !ok {
		t.Fatal("missing getenv extern")
	}
	name := mustAllocBytes(t, mem, "getenv:configured-name", []byte("CVM_TEST\x00"), true, blockString)
	ret, exit, err := fn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(name)})
	if err != nil || exit != nil {
		t.Fatalf("getenv ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypePtr || ret.Int == 0 {
		t.Fatalf("getenv ret=%#v, want non-null pointer", ret)
	}
	got, err := mem.ReadCString(ret.Int)
	if err != nil || got != "configured-value" {
		t.Fatalf("getenv value=%q err=%v, want configured-value", got, err)
	}
	ret2, exit, err := fn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(name)})
	if err != nil || exit != nil {
		t.Fatalf("getenv second ret=%#v exit=%#v err=%v", ret2, exit, err)
	}
	if ret2.Int != ret.Int {
		t.Fatalf("getenv pointer changed %#x -> %#x", ret.Int, ret2.Int)
	}

	mem2 := NewMemory(bytecode.DefaultTarget())
	name2 := mustAllocBytes(t, mem2, "getenv:configured-name-2", []byte("CVM_TEST\x00"), true, blockString)
	ret3, exit, err := fn(context.Background(), &ExternContext{Memory: mem2}, []Value{PtrValue(name2)})
	if err != nil || exit != nil {
		t.Fatalf("getenv other memory ret=%#v exit=%#v err=%v", ret3, exit, err)
	}
	if ret3.Type != bytecode.TypePtr || ret3.Int == 0 {
		t.Fatalf("getenv other memory ret=%#v, want non-null pointer", ret3)
	}
	got, err = mem2.ReadCString(ret3.Int)
	if err != nil || got != "configured-value" {
		t.Fatalf("getenv other memory value=%q err=%v, want configured-value", got, err)
	}
}

func TestStdlibSystemExtern(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	fn, ok := reg.Lookup("system")
	if !ok {
		t.Fatal("missing system extern")
	}
	ret, exit, err := fn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(0)})
	if err != nil || exit != nil {
		t.Fatalf("system null ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeI32 || signedInt(ret) != 0 {
		t.Fatalf("system null ret=%#v, want i32 0", ret)
	}
	cmd := mustAllocBytes(t, mem, "system:cmd", []byte("echo nope\x00"), true, blockString)
	ret, exit, err = fn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(cmd)})
	if err != nil || exit != nil {
		t.Fatalf("system cmd ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeI32 || signedInt(ret) != -1 {
		t.Fatalf("system cmd ret=%#v, want i32 -1", ret)
	}
}

func TestStdlibAtexitExtern(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	fn, ok := reg.Lookup("atexit")
	if !ok {
		t.Fatal("missing atexit extern")
	}
	ret, exit, err := fn(context.Background(), nil, []Value{PtrValue(0x1234)})
	if err != nil || exit != nil {
		t.Fatalf("atexit ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeI32 || signedInt(ret) != 0 {
		t.Fatalf("atexit ret=%#v, want i32 0", ret)
	}
}

func TestLocaleSetlocaleExtern(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	fn, ok := reg.Lookup("setlocale")
	if !ok {
		t.Fatal("missing setlocale extern")
	}
	categories := []int64{0, 1, 2, 3, 4, 5}
	ret, exit, err := fn(context.Background(), &ExternContext{Memory: mem}, []Value{IntValue(bytecode.TypeI32, 0), PtrValue(0)})
	if err != nil || exit != nil {
		t.Fatalf("setlocale query ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypePtr || ret.Int == 0 {
		t.Fatalf("setlocale query ret=%#v, want non-null C locale string", ret)
	}
	if got, err := mem.ReadCString(ret.Int); err != nil || got != "C" {
		t.Fatalf("setlocale query string=%q err=%v, want C", got, err)
	}
	initialPtr := ret.Int

	cLocale := mustAllocBytes(t, mem, "locale:C", []byte("C\x00"), true, blockString)
	for _, category := range categories {
		ret, exit, err = fn(context.Background(), &ExternContext{Memory: mem}, []Value{IntValue(bytecode.TypeI32, category), PtrValue(cLocale)})
		if err != nil || exit != nil {
			t.Fatalf("setlocale C category %d ret=%#v exit=%#v err=%v", category, ret, exit, err)
		}
		if ret.Type != bytecode.TypePtr || ret.Int != initialPtr {
			t.Fatalf("setlocale C category %d ret=%#v, want stable C locale pointer %#x", category, ret, initialPtr)
		}
		ret, exit, err = fn(context.Background(), &ExternContext{Memory: mem}, []Value{IntValue(bytecode.TypeI32, category), PtrValue(0)})
		if err != nil || exit != nil {
			t.Fatalf("setlocale query category %d ret=%#v exit=%#v err=%v", category, ret, exit, err)
		}
		if ret.Type != bytecode.TypePtr || ret.Int != initialPtr {
			t.Fatalf("setlocale query category %d ret=%#v, want stable C locale pointer %#x", category, ret, initialPtr)
		}
	}

	emptyLocale := mustAllocBytes(t, mem, "locale:empty", []byte("\x00"), true, blockString)
	for _, category := range categories {
		ret, exit, err = fn(context.Background(), &ExternContext{Memory: mem}, []Value{IntValue(bytecode.TypeI32, category), PtrValue(emptyLocale)})
		if err != nil || exit != nil {
			t.Fatalf("setlocale empty category %d ret=%#v exit=%#v err=%v", category, ret, exit, err)
		}
		if ret.Type != bytecode.TypePtr || ret.Int != initialPtr {
			t.Fatalf("setlocale empty category %d ret=%#v, want stable C locale pointer %#x", category, ret, initialPtr)
		}
		if got, err := mem.ReadCString(ret.Int); err != nil || got != "C" {
			t.Fatalf("setlocale empty category %d string=%q err=%v, want C", category, got, err)
		}
	}

	unsupported := mustAllocBytes(t, mem, "locale:unsupported", []byte("ja_JP.UTF-8\x00"), true, blockString)
	ret, exit, err = fn(context.Background(), &ExternContext{Memory: mem}, []Value{IntValue(bytecode.TypeI32, 0), PtrValue(unsupported)})
	if err != nil || exit != nil {
		t.Fatalf("setlocale unsupported ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypePtr || ret.Int != 0 {
		t.Fatalf("setlocale unsupported ret=%#v, want null", ret)
	}
	ret, exit, err = fn(context.Background(), &ExternContext{Memory: mem}, []Value{IntValue(bytecode.TypeI32, 0), PtrValue(0)})
	if err != nil || exit != nil {
		t.Fatalf("setlocale query after unsupported ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypePtr || ret.Int != initialPtr {
		t.Fatalf("setlocale query after unsupported ret=%#v, want stable C locale pointer %#x", ret, initialPtr)
	}
}

func TestLocaleconvExtern(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	target := bytecode.DefaultTarget()
	mem := NewMemory(target)
	fn, ok := reg.Lookup("localeconv")
	if !ok {
		t.Fatal("missing localeconv extern")
	}
	ret, exit, err := fn(context.Background(), &ExternContext{Memory: mem}, nil)
	if err != nil || exit != nil {
		t.Fatalf("localeconv ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypePtr || ret.Int == 0 {
		t.Fatalf("localeconv ret=%#v, want non-null struct pointer", ret)
	}
	addr := ret.Int
	checkStringField := func(name string, index int, want string) {
		t.Helper()
		ptr, err := mem.Load(addr+uint64(int64(index)*target.PointerSize), bytecode.TypePtr, target.PointerAlign)
		if err != nil {
			t.Fatalf("load %s pointer: %v", name, err)
		}
		got, err := mem.ReadCString(ptr.Int)
		if err != nil || got != want {
			t.Fatalf("%s=%q err=%v, want %q", name, got, err, want)
		}
	}
	checkStringField("decimal_point", 0, ".")
	checkStringField("thousands_sep", 1, "")
	checkStringField("grouping", 2, "")

	charBase := addr + uint64(10*target.PointerSize)
	for i, name := range []string{"int_frac_digits", "frac_digits", "p_cs_precedes"} {
		got, err := mem.Load(charBase+uint64(i), bytecode.TypeI8, 1)
		if err != nil {
			t.Fatalf("load %s: %v", name, err)
		}
		if got.Int != 127 {
			t.Fatalf("%s=%#v, want CHAR_MAX", name, got)
		}
	}

	ret, exit, err = fn(context.Background(), &ExternContext{Memory: mem}, nil)
	if err != nil || exit != nil {
		t.Fatalf("localeconv second ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypePtr || ret.Int != addr {
		t.Fatalf("localeconv second ret=%#v, want stable pointer %#x", ret, addr)
	}

	other := NewMemory(target)
	ret, exit, err = fn(context.Background(), &ExternContext{Memory: other}, nil)
	if err != nil || exit != nil {
		t.Fatalf("localeconv other memory ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypePtr || ret.Int == 0 {
		t.Fatalf("localeconv other memory ret=%#v, want independent non-null pointer", ret)
	}
}

func TestTimeExterns(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	timeFn, ok := reg.Lookup("time")
	if !ok {
		t.Fatal("missing time extern")
	}
	ret, exit, err := timeFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(0)})
	if err != nil || exit != nil {
		t.Fatalf("time null ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeI64 || signedInt(ret) != 0 {
		t.Fatalf("time null ret=%#v, want i64 0", ret)
	}
	slot := mustAlloc(t, mem, "time:slot", 8, 8, false, blockLocal)
	ret, exit, err = timeFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(slot)})
	if err != nil || exit != nil {
		t.Fatalf("time slot ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	stored, err := mem.Load(slot, bytecode.TypeI64, 8)
	if err != nil {
		t.Fatalf("load stored time: %v", err)
	}
	if signedInt(stored) != 0 {
		t.Fatalf("stored time=%#v, want 0", stored)
	}
	if ret, exit, err = timeFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(1)}); err == nil || exit != nil {
		t.Fatalf("time invalid ret=%#v exit=%#v err=%v, want memory error without exit", ret, exit, err)
	}

	diffFn, ok := reg.Lookup("difftime")
	if !ok {
		t.Fatal("missing difftime extern")
	}
	ret, exit, err = diffFn(context.Background(), nil, []Value{IntValue(bytecode.TypeI64, 7), IntValue(bytecode.TypeI64, 2)})
	if err != nil || exit != nil {
		t.Fatalf("difftime ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeF64 || ret.Float != 5 {
		t.Fatalf("difftime ret=%#v, want f64 5", ret)
	}
	ret, exit, err = diffFn(context.Background(), nil, []Value{IntValue(bytecode.TypeI64, 7), IntValue(bytecode.TypeI64, 7)})
	if err != nil || exit != nil {
		t.Fatalf("difftime zero ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeF64 || ret.Float != 0 {
		t.Fatalf("difftime zero ret=%#v, want f64 0", ret)
	}
	ret, exit, err = diffFn(context.Background(), nil, []Value{IntValue(bytecode.TypeI64, 2), IntValue(bytecode.TypeI64, 7)})
	if err != nil || exit != nil {
		t.Fatalf("difftime negative ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeF64 || ret.Float != -5 {
		t.Fatalf("difftime negative ret=%#v, want f64 -5", ret)
	}

	clockFn, ok := reg.Lookup("clock")
	if !ok {
		t.Fatal("missing clock extern")
	}
	ret, exit, err = clockFn(context.Background(), nil, nil)
	if err != nil || exit != nil {
		t.Fatalf("clock ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeI64 || signedInt(ret) != 0 {
		t.Fatalf("clock ret=%#v, want i64 0", ret)
	}
}

func TestErrnoExternVariable(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	addr, ok := reg.LookupVariable("errno", mem)
	if !ok {
		t.Fatal("missing errno extern variable")
	}
	again, ok := reg.LookupVariable("errno", mem)
	if !ok {
		t.Fatal("missing errno extern variable on second lookup")
	}
	if again != addr {
		t.Fatalf("errno address changed: first=%#x second=%#x", addr, again)
	}
	ret, err := mem.Load(addr, bytecode.TypeI32, 4)
	if err != nil {
		t.Fatalf("load errno: %v", err)
	}
	if signedInt(ret) != 0 {
		t.Fatalf("initial errno=%#v, want 0", ret)
	}
	if err := mem.Store(addr, bytecode.TypeI32, 4, IntValue(bytecode.TypeI32, 34)); err != nil {
		t.Fatalf("store errno: %v", err)
	}
	ret, err = mem.Load(addr, bytecode.TypeI32, 4)
	if err != nil {
		t.Fatalf("reload errno: %v", err)
	}
	if signedInt(ret) != 34 {
		t.Fatalf("updated errno=%#v, want 34", ret)
	}

	other := NewMemory(bytecode.DefaultTarget())
	otherAddr, ok := reg.LookupVariable("errno", other)
	if !ok {
		t.Fatal("missing errno extern variable for second memory")
	}
	otherAgain, ok := reg.LookupVariable("errno", other)
	if !ok {
		t.Fatal("missing errno extern variable for second memory on second lookup")
	}
	if otherAgain != otherAddr {
		t.Fatalf("second memory errno address changed: first=%#x second=%#x", otherAddr, otherAgain)
	}
	otherRet, err := other.Load(otherAddr, bytecode.TypeI32, 4)
	if err != nil {
		t.Fatalf("load second memory errno: %v", err)
	}
	if signedInt(otherRet) != 0 {
		t.Fatalf("second memory initial errno=%#v, want 0", otherRet)
	}
}

func TestStdlibDivExterns(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	tests := []struct {
		name       string
		numer      Value
		denom      Value
		fieldType  bytecode.ValueType
		fieldAlign int64
		remOffset  uint64
		wantQuot   int64
		wantRem    int64
	}{
		{name: "div", numer: IntValue(bytecode.TypeI32, -7), denom: IntValue(bytecode.TypeI32, 3), fieldType: bytecode.TypeI32, fieldAlign: 4, remOffset: 4, wantQuot: -2, wantRem: -1},
		{name: "ldiv", numer: IntValue(bytecode.TypeI64, -9), denom: IntValue(bytecode.TypeI64, 4), fieldType: bytecode.TypeI64, fieldAlign: 8, remOffset: 8, wantQuot: -2, wantRem: -1},
		{name: "lldiv", numer: IntValue(bytecode.TypeI64, 10), denom: IntValue(bytecode.TypeI64, -4), fieldType: bytecode.TypeI64, fieldAlign: 8, remOffset: 8, wantQuot: -2, wantRem: 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn, ok := reg.Lookup(tt.name)
			if !ok {
				t.Fatalf("missing %s extern", tt.name)
			}
			ret, exit, err := fn(context.Background(), &ExternContext{Memory: mem}, []Value{tt.numer, tt.denom})
			if err != nil || exit != nil {
				t.Fatalf("%s ret=%#v exit=%#v err=%v", tt.name, ret, exit, err)
			}
			if ret.Type != bytecode.TypeObjectAddr || ret.Int == 0 {
				t.Fatalf("%s ret=%#v, want object address", tt.name, ret)
			}
			quot, err := mem.Load(ret.Int, tt.fieldType, tt.fieldAlign)
			if err != nil {
				t.Fatalf("%s load quot: %v", tt.name, err)
			}
			rem, err := mem.Load(ret.Int+tt.remOffset, tt.fieldType, tt.fieldAlign)
			if err != nil {
				t.Fatalf("%s load rem: %v", tt.name, err)
			}
			if signedInt(quot) != tt.wantQuot || signedInt(rem) != tt.wantRem {
				t.Fatalf("%s result quot=%d rem=%d, want quot=%d rem=%d", tt.name, signedInt(quot), signedInt(rem), tt.wantQuot, tt.wantRem)
			}
		})
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

func TestStdioPerrorExtern(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer
	reg := DefaultExternRegistry(&out, &errOut)
	mem := NewMemory(bytecode.DefaultTarget())
	prefix := mustAllocBytes(t, mem, "perror:prefix", []byte("phase\x00"), true, blockString)
	empty := mustAllocBytes(t, mem, "perror:empty", []byte("\x00"), true, blockString)
	fn, ok := reg.Lookup("perror")
	if !ok {
		t.Fatal("missing perror extern")
	}
	ret, exit, err := fn(context.Background(), &ExternContext{Memory: mem, Stderr: &errOut}, []Value{PtrValue(prefix)})
	if err != nil || exit != nil {
		t.Fatalf("perror prefix ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeVoid {
		t.Fatalf("perror ret=%#v, want void", ret)
	}
	ret, exit, err = fn(context.Background(), &ExternContext{Memory: mem, Stderr: &errOut}, []Value{PtrValue(empty)})
	if err != nil || exit != nil {
		t.Fatalf("perror empty ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	ret, exit, err = fn(context.Background(), &ExternContext{Memory: mem, Stderr: &errOut}, []Value{PtrValue(0)})
	if err != nil || exit != nil {
		t.Fatalf("perror null ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if out.String() != "" {
		t.Fatalf("stdout output = %q, want empty", out.String())
	}
	if errOut.String() != "phase: error\nerror\nerror\n" {
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

func TestStringCharSearchExterns(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	text := mustAllocBytes(t, mem, "strchr:text", []byte{'a', 0x82, 'c', 0}, true, blockString)
	for _, name := range []string{"strchr", "__builtin_strchr"} {
		fn, ok := reg.Lookup(name)
		if !ok {
			t.Fatalf("missing %s extern", name)
		}
		ret, exit, err := fn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(text), IntValue(bytecode.TypeI32, 'c')})
		if err != nil || exit != nil {
			t.Fatalf("%s hit ret=%#v exit=%#v err=%v", name, ret, exit, err)
		}
		if ret.Type != bytecode.TypePtr || ret.Int != text+2 {
			t.Fatalf("%s hit ret=%#v, want pointer %#x", name, ret, text+2)
		}
		ret, exit, err = fn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(text), IntValue(bytecode.TypeI32, 0)})
		if err != nil || exit != nil {
			t.Fatalf("%s nul ret=%#v exit=%#v err=%v", name, ret, exit, err)
		}
		if ret.Type != bytecode.TypePtr || ret.Int != text+3 {
			t.Fatalf("%s nul ret=%#v, want pointer %#x", name, ret, text+3)
		}
		ret, exit, err = fn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(text), IntValue(bytecode.TypeI32, 0x182)})
		if err != nil || exit != nil {
			t.Fatalf("%s masked ret=%#v exit=%#v err=%v", name, ret, exit, err)
		}
		if ret.Type != bytecode.TypePtr || ret.Int != text+1 {
			t.Fatalf("%s masked ret=%#v, want pointer %#x", name, ret, text+1)
		}
		ret, exit, err = fn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(text), IntValue(bytecode.TypeI32, 'z')})
		if err != nil || exit != nil {
			t.Fatalf("%s miss ret=%#v exit=%#v err=%v", name, ret, exit, err)
		}
		if ret.Type != bytecode.TypePtr || ret.Int != 0 {
			t.Fatalf("%s miss ret=%#v, want null pointer", name, ret)
		}
	}
}

func TestStringSubstringSearchExterns(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	haystack := mustAllocBytes(t, mem, "strstr:haystack", []byte("abcdef\x00"), true, blockString)
	tests := []struct {
		name   string
		needle string
		want   uint64
	}{
		{name: "empty", needle: "", want: haystack},
		{name: "full", needle: "abcdef", want: haystack},
		{name: "suffix", needle: "ef", want: haystack + 4},
		{name: "middle", needle: "cd", want: haystack + 2},
		{name: "miss", needle: "gh", want: 0},
	}
	for _, fnName := range []string{"strstr", "__builtin_strstr"} {
		fn, ok := reg.Lookup(fnName)
		if !ok {
			t.Fatalf("missing %s extern", fnName)
		}
		for _, tt := range tests {
			t.Run(fnName+"/"+tt.name, func(t *testing.T) {
				needle := mustAllocBytes(t, mem, "strstr:needle:"+tt.name, []byte(tt.needle+"\x00"), true, blockString)
				ret, exit, err := fn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(haystack), PtrValue(needle)})
				if err != nil || exit != nil {
					t.Fatalf("%s %s ret=%#v exit=%#v err=%v", fnName, tt.name, ret, exit, err)
				}
				if ret.Type != bytecode.TypePtr || ret.Int != tt.want {
					t.Fatalf("%s %s ret=%#v, want pointer %#x", fnName, tt.name, ret, tt.want)
				}
			})
		}
	}
}

func TestStringBoundedCompareSearchExterns(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	left := mustAllocBytes(t, mem, "strncmp:left", []byte("abcdef\x00"), true, blockString)
	samePrefix := mustAllocBytes(t, mem, "strncmp:same", []byte("abcxyz\x00"), true, blockString)
	less := mustAllocBytes(t, mem, "strncmp:less", []byte("abbzzz\x00"), true, blockString)
	leftTerminated := mustAllocBytes(t, mem, "strncmp:left-terminated", []byte{'a', 0, 'z'}, true, blockString)
	rightTerminated := mustAllocBytes(t, mem, "strncmp:right-terminated", []byte{'a', 0, 'y'}, true, blockString)
	leftUnterminated := mustAllocBytes(t, mem, "strncmp:left-unterminated", []byte{'a', 'b', 'c'}, true, blockString)
	rightUnterminated := mustAllocBytes(t, mem, "strncmp:right-unterminated", []byte{'a', 'b', 'd'}, true, blockString)
	high := mustAllocBytes(t, mem, "strncmp:high", []byte{0xff, 0}, true, blockString)
	low := mustAllocBytes(t, mem, "strncmp:low", []byte{0x7f, 0}, true, blockString)
	strncmpFn, ok := reg.Lookup("strncmp")
	if !ok {
		t.Fatal("missing strncmp extern")
	}
	tests := []struct {
		name string
		a    uint64
		b    uint64
		n    uint64
		want int64
	}{
		{name: "zero", a: left, b: less, n: 0, want: 0},
		{name: "zero invalid pointers", a: 0xdeadbeef, b: 0xcafebabe, n: 0, want: 0},
		{name: "prefix", a: left, b: samePrefix, n: 3, want: 0},
		{name: "greater", a: left, b: less, n: 3, want: 1},
		{name: "terminator before count", a: leftTerminated, b: rightTerminated, n: 3, want: 0},
		{name: "unterminated equal prefix", a: leftUnterminated, b: rightUnterminated, n: 2, want: 0},
		{name: "unterminated less", a: leftUnterminated, b: rightUnterminated, n: 3, want: -1},
		{name: "unsigned byte greater", a: high, b: low, n: 1, want: 1},
	}
	for _, tt := range tests {
		ret, exit, err := strncmpFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(tt.a), PtrValue(tt.b), UIntValue(bytecode.TypeU64, tt.n)})
		if err != nil || exit != nil {
			t.Fatalf("strncmp %s ret=%#v exit=%#v err=%v", tt.name, ret, exit, err)
		}
		got := int64(int32(ret.Int))
		if ret.Type != bytecode.TypeI32 || got != tt.want {
			t.Fatalf("strncmp %s ret=%#v, want i32 %d", tt.name, ret, tt.want)
		}
	}

	memchrFn, ok := reg.Lookup("memchr")
	if !ok {
		t.Fatal("missing memchr extern")
	}
	data := mustAllocBytes(t, mem, "memchr:data", []byte{1, 2, 3, 2, 0}, true, blockString)
	ret, exit, err := memchrFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(data), IntValue(bytecode.TypeI32, 2), UIntValue(bytecode.TypeU64, 4)})
	if err != nil || exit != nil {
		t.Fatalf("memchr hit ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypePtr || ret.Int != data+1 {
		t.Fatalf("memchr hit ret=%#v, want pointer %#x", ret, data+1)
	}
	ret, exit, err = memchrFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(data), IntValue(bytecode.TypeI32, 2), UIntValue(bytecode.TypeU64, 0)})
	if err != nil || exit != nil {
		t.Fatalf("memchr zero ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypePtr || ret.Int != 0 {
		t.Fatalf("memchr zero ret=%#v, want null pointer", ret)
	}
	ret, exit, err = memchrFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(data), IntValue(bytecode.TypeI32, 0x102), UIntValue(bytecode.TypeU64, 4)})
	if err != nil || exit != nil {
		t.Fatalf("memchr masked ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypePtr || ret.Int != data+1 {
		t.Fatalf("memchr masked ret=%#v, want pointer %#x", ret, data+1)
	}
	ret, exit, err = memchrFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(data), IntValue(bytecode.TypeI32, 9), UIntValue(bytecode.TypeU64, 4)})
	if err != nil || exit != nil {
		t.Fatalf("memchr miss ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypePtr || ret.Int != 0 {
		t.Fatalf("memchr miss ret=%#v, want null pointer", ret)
	}
}

func TestWideMemoryCharSearchExtern(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	fn, ok := reg.Lookup("wmemchr")
	if !ok {
		t.Fatal("missing wmemchr extern")
	}
	mem := NewMemory(bytecode.DefaultTarget())
	data := mustAllocBytes(t, mem, "wmemchr:data", make([]byte, 16), false, blockLocal)
	for i, ch := range []int64{'A', 0x1234, 0, 'Z'} {
		if err := mem.Store(data+uint64(i*4), bytecode.TypeI32, 4, IntValue(bytecode.TypeI32, ch)); err != nil {
			t.Fatalf("store wide %d: %v", i, err)
		}
	}
	tests := []struct {
		name  string
		ptr   uint64
		ch    int64
		count uint64
		want  uint64
	}{
		{name: "hit", ptr: data, ch: 0x1234, count: 4, want: data + 4},
		{name: "nul", ptr: data, ch: 0, count: 4, want: data + 8},
		{name: "miss", ptr: data, ch: 'x', count: 4, want: 0},
		{name: "zero invalid", ptr: 0xdeadbeef, ch: 'x', count: 0, want: 0},
	}
	for _, tt := range tests {
		ret, exit, err := fn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(tt.ptr), IntValue(bytecode.TypeI32, tt.ch), UIntValue(bytecode.TypeU64, tt.count)})
		if err != nil || exit != nil {
			t.Fatalf("%s ret=%#v exit=%#v err=%v", tt.name, ret, exit, err)
		}
		if ret.Type != bytecode.TypePtr || ret.Int != tt.want {
			t.Fatalf("%s ret=%#v, want pointer %#x", tt.name, ret, tt.want)
		}
	}
}

func TestWideMemoryCompareExtern(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	fn, ok := reg.Lookup("wmemcmp")
	if !ok {
		t.Fatal("missing wmemcmp extern")
	}
	mem := NewMemory(bytecode.DefaultTarget())
	left := mustAllocBytes(t, mem, "wmemcmp:left", make([]byte, 12), false, blockLocal)
	right := mustAllocBytes(t, mem, "wmemcmp:right", make([]byte, 12), false, blockLocal)
	for i, ch := range []int64{'A', 0x1234, 'C'} {
		if err := mem.Store(left+uint64(i*4), bytecode.TypeI32, 4, IntValue(bytecode.TypeI32, ch)); err != nil {
			t.Fatalf("store left %d: %v", i, err)
		}
	}
	for i, ch := range []int64{'A', 0x1235, 'B'} {
		if err := mem.Store(right+uint64(i*4), bytecode.TypeI32, 4, IntValue(bytecode.TypeI32, ch)); err != nil {
			t.Fatalf("store right %d: %v", i, err)
		}
	}
	tests := []struct {
		name  string
		a     uint64
		b     uint64
		count uint64
		want  int64
	}{
		{name: "zero invalid", a: 0xdeadbeef, b: 0xcafebabe, count: 0, want: 0},
		{name: "equal prefix", a: left, b: right, count: 1, want: 0},
		{name: "less", a: left, b: right, count: 2, want: -1},
		{name: "greater", a: right, b: left, count: 2, want: 1},
	}
	for _, tt := range tests {
		ret, exit, err := fn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(tt.a), PtrValue(tt.b), UIntValue(bytecode.TypeU64, tt.count)})
		if err != nil || exit != nil {
			t.Fatalf("%s ret=%#v exit=%#v err=%v", tt.name, ret, exit, err)
		}
		got := int64(int32(ret.Int))
		if ret.Type != bytecode.TypeI32 || got != tt.want {
			t.Fatalf("%s ret=%#v, want i32 %d", tt.name, ret, tt.want)
		}
	}
}

func TestWideMemoryCopyExtern(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	fn, ok := reg.Lookup("wmemcpy")
	if !ok {
		t.Fatal("missing wmemcpy extern")
	}
	mem := NewMemory(bytecode.DefaultTarget())
	src := mustAllocBytes(t, mem, "wmemcpy:src", make([]byte, 12), false, blockLocal)
	dst := mustAllocBytes(t, mem, "wmemcpy:dst", make([]byte, 12), false, blockLocal)
	for i, ch := range []int64{'Q', 0x1234, 0} {
		if err := mem.Store(src+uint64(i*4), bytecode.TypeI32, 4, IntValue(bytecode.TypeI32, ch)); err != nil {
			t.Fatalf("store src %d: %v", i, err)
		}
	}
	ret, exit, err := fn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(dst), PtrValue(src), UIntValue(bytecode.TypeU64, 3)})
	if err != nil || exit != nil {
		t.Fatalf("copy ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypePtr || ret.Int != dst {
		t.Fatalf("copy ret=%#v, want dst %#x", ret, dst)
	}
	for i, want := range []uint32{'Q', 0x1234, 0} {
		got, err := loadWideChar(mem, dst+uint64(i*4))
		if err != nil {
			t.Fatalf("load dst %d: %v", i, err)
		}
		if got != want {
			t.Fatalf("dst[%d]=%#x, want %#x", i, got, want)
		}
	}
	ret, exit, err = fn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(0xdeadbeef), PtrValue(0xcafebabe), UIntValue(bytecode.TypeU64, 0)})
	if err != nil || exit != nil {
		t.Fatalf("zero ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypePtr || ret.Int != 0xdeadbeef {
		t.Fatalf("zero ret=%#v, want invalid dst preserved", ret)
	}
}

func TestWideMemoryMoveExtern(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	fn, ok := reg.Lookup("wmemmove")
	if !ok {
		t.Fatal("missing wmemmove extern")
	}
	mem := NewMemory(bytecode.DefaultTarget())
	data := mustAllocBytes(t, mem, "wmemmove:data", make([]byte, 20), false, blockLocal)
	store := func(values ...int64) {
		t.Helper()
		for i, ch := range values {
			if err := mem.Store(data+uint64(i*4), bytecode.TypeI32, 4, IntValue(bytecode.TypeI32, ch)); err != nil {
				t.Fatalf("store %d: %v", i, err)
			}
		}
	}
	check := func(values ...uint32) {
		t.Helper()
		for i, want := range values {
			got, err := loadWideChar(mem, data+uint64(i*4))
			if err != nil {
				t.Fatalf("load %d: %v", i, err)
			}
			if got != want {
				t.Fatalf("data[%d]=%#x, want %#x", i, got, want)
			}
		}
	}
	store('a', 'b', 'c', 'd', 0)
	ret, exit, err := fn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(data + 4), PtrValue(data), UIntValue(bytecode.TypeU64, 3)})
	if err != nil || exit != nil {
		t.Fatalf("forward overlap ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypePtr || ret.Int != data+4 {
		t.Fatalf("forward overlap ret=%#v, want dst %#x", ret, data+4)
	}
	check('a', 'a', 'b', 'c', 0)
	store('a', 'b', 'c', 'd', 0)
	ret, exit, err = fn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(data), PtrValue(data + 4), UIntValue(bytecode.TypeU64, 3)})
	if err != nil || exit != nil {
		t.Fatalf("backward overlap ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypePtr || ret.Int != data {
		t.Fatalf("backward overlap ret=%#v, want dst %#x", ret, data)
	}
	check('b', 'c', 'd', 'd', 0)
	ret, exit, err = fn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(0xdeadbeef), PtrValue(0xcafebabe), UIntValue(bytecode.TypeU64, 0)})
	if err != nil || exit != nil {
		t.Fatalf("zero ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypePtr || ret.Int != 0xdeadbeef {
		t.Fatalf("zero ret=%#v, want invalid dst preserved", ret)
	}
}

func TestWideMemorySetExtern(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	fn, ok := reg.Lookup("wmemset")
	if !ok {
		t.Fatal("missing wmemset extern")
	}
	mem := NewMemory(bytecode.DefaultTarget())
	data := mustAllocBytes(t, mem, "wmemset:data", make([]byte, 12), false, blockLocal)
	ret, exit, err := fn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(data), IntValue(bytecode.TypeI32, 0x1234), UIntValue(bytecode.TypeU64, 3)})
	if err != nil || exit != nil {
		t.Fatalf("set ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypePtr || ret.Int != data {
		t.Fatalf("set ret=%#v, want dst %#x", ret, data)
	}
	for i := 0; i < 3; i++ {
		got, err := loadWideChar(mem, data+uint64(i*4))
		if err != nil {
			t.Fatalf("load %d: %v", i, err)
		}
		if got != 0x1234 {
			t.Fatalf("data[%d]=%#x, want 0x1234", i, got)
		}
	}
	ret, exit, err = fn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(0xdeadbeef), IntValue(bytecode.TypeI32, 'x'), UIntValue(bytecode.TypeU64, 0)})
	if err != nil || exit != nil {
		t.Fatalf("zero ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypePtr || ret.Int != 0xdeadbeef {
		t.Fatalf("zero ret=%#v, want invalid dst preserved", ret)
	}
}

func TestWideStringLengthExtern(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	fn, ok := reg.Lookup("wcslen")
	if !ok {
		t.Fatal("missing wcslen extern")
	}
	mem := NewMemory(bytecode.DefaultTarget())
	empty := mustAllocBytes(t, mem, "wcslen:empty", make([]byte, 4), false, blockLocal)
	text := mustAllocBytes(t, mem, "wcslen:text", make([]byte, 16), false, blockLocal)
	for i, ch := range []uint32{'A', 0x1234, 'Z', 0} {
		if err := storeWideChar(mem, text+uint64(i*4), ch); err != nil {
			t.Fatalf("store text %d: %v", i, err)
		}
	}
	tests := []struct {
		name string
		addr uint64
		want uint64
	}{
		{name: "empty", addr: empty, want: 0},
		{name: "text", addr: text, want: 3},
	}
	for _, tt := range tests {
		ret, exit, err := fn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(tt.addr)})
		if err != nil || exit != nil {
			t.Fatalf("%s ret=%#v exit=%#v err=%v", tt.name, ret, exit, err)
		}
		if ret.Type != bytecode.TypeU64 || ret.Int != tt.want {
			t.Fatalf("%s ret=%#v, want u64 %d", tt.name, ret, tt.want)
		}
	}
}

func TestWideStringCompareExtern(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	fn, ok := reg.Lookup("wcscmp")
	if !ok {
		t.Fatal("missing wcscmp extern")
	}
	mem := NewMemory(bytecode.DefaultTarget())
	makeWide := func(name string, chars ...uint32) uint64 {
		t.Helper()
		addr := mustAllocBytes(t, mem, name, make([]byte, len(chars)*4), false, blockLocal)
		for i, ch := range chars {
			if err := storeWideChar(mem, addr+uint64(i*4), ch); err != nil {
				t.Fatalf("store %s[%d]: %v", name, i, err)
			}
		}
		return addr
	}
	abc := makeWide("wcscmp:abc", 'a', 'b', 'c', 0)
	abd := makeWide("wcscmp:abd", 'a', 'b', 'd', 0)
	ab := makeWide("wcscmp:ab", 'a', 'b', 0)
	high := makeWide("wcscmp:high", 0x1234, 0)
	low := makeWide("wcscmp:low", 0x1233, 0)
	empty := makeWide("wcscmp:empty", 0)
	tests := []struct {
		name string
		a    uint64
		b    uint64
		want int64
	}{
		{name: "equal", a: abc, b: abc, want: 0},
		{name: "less", a: abc, b: abd, want: -1},
		{name: "greater", a: abd, b: abc, want: 1},
		{name: "prefix greater", a: abc, b: ab, want: 1},
		{name: "prefix less", a: ab, b: abc, want: -1},
		{name: "empty", a: empty, b: empty, want: 0},
		{name: "high", a: high, b: low, want: 1},
	}
	for _, tt := range tests {
		ret, exit, err := fn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(tt.a), PtrValue(tt.b)})
		if err != nil || exit != nil {
			t.Fatalf("%s ret=%#v exit=%#v err=%v", tt.name, ret, exit, err)
		}
		got := int64(int32(ret.Int))
		if ret.Type != bytecode.TypeI32 || got != tt.want {
			t.Fatalf("%s ret=%#v, want i32 %d", tt.name, ret, tt.want)
		}
	}
}

func TestWideStringCollateExtern(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	fn, ok := reg.Lookup("wcscoll")
	if !ok {
		t.Fatal("missing wcscoll extern")
	}
	mem := NewMemory(bytecode.DefaultTarget())
	makeWide := func(name string, chars ...uint32) uint64 {
		t.Helper()
		addr := mustAllocBytes(t, mem, name, make([]byte, len(chars)*4), false, blockLocal)
		for i, ch := range chars {
			if err := storeWideChar(mem, addr+uint64(i*4), ch); err != nil {
				t.Fatalf("store %s[%d]: %v", name, i, err)
			}
		}
		return addr
	}
	abc := makeWide("wcscoll:abc", 'a', 'b', 'c', 0)
	abd := makeWide("wcscoll:abd", 'a', 'b', 'd', 0)
	high := makeWide("wcscoll:high", 0x1234, 0)
	low := makeWide("wcscoll:low", 0x1233, 0)
	tests := []struct {
		name string
		a    uint64
		b    uint64
		want int64
	}{
		{name: "equal", a: abc, b: abc, want: 0},
		{name: "less", a: abc, b: abd, want: -1},
		{name: "greater", a: abd, b: abc, want: 1},
		{name: "high", a: high, b: low, want: 1},
	}
	for _, tt := range tests {
		ret, exit, err := fn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(tt.a), PtrValue(tt.b)})
		if err != nil || exit != nil {
			t.Fatalf("%s ret=%#v exit=%#v err=%v", tt.name, ret, exit, err)
		}
		got := int64(int32(ret.Int))
		if ret.Type != bytecode.TypeI32 || got != tt.want {
			t.Fatalf("%s ret=%#v, want i32 %d", tt.name, ret, tt.want)
		}
	}
}

func TestWideStringNCompareExtern(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	fn, ok := reg.Lookup("wcsncmp")
	if !ok {
		t.Fatal("missing wcsncmp extern")
	}
	mem := NewMemory(bytecode.DefaultTarget())
	makeWide := func(name string, chars ...uint32) uint64 {
		t.Helper()
		addr := mustAllocBytes(t, mem, name, make([]byte, len(chars)*4), false, blockLocal)
		for i, ch := range chars {
			if err := storeWideChar(mem, addr+uint64(i*4), ch); err != nil {
				t.Fatalf("store %s[%d]: %v", name, i, err)
			}
		}
		return addr
	}
	abc := makeWide("wcsncmp:abc", 'a', 'b', 'c', 0)
	abd := makeWide("wcsncmp:abd", 'a', 'b', 'd', 0)
	ab := makeWide("wcsncmp:ab", 'a', 'b', 0)
	high := makeWide("wcsncmp:high", 0x1234, 0)
	low := makeWide("wcsncmp:low", 0x1233, 0)
	tests := []struct {
		name  string
		a     uint64
		b     uint64
		count uint64
		want  int64
	}{
		{name: "zero invalid", a: 0xdeadbeef, b: 0xcafebabe, count: 0, want: 0},
		{name: "equal prefix", a: abc, b: abd, count: 2, want: 0},
		{name: "less", a: abc, b: abd, count: 3, want: -1},
		{name: "greater", a: abd, b: abc, count: 3, want: 1},
		{name: "early terminator", a: ab, b: abc, count: 3, want: -1},
		{name: "high", a: high, b: low, count: 1, want: 1},
	}
	for _, tt := range tests {
		ret, exit, err := fn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(tt.a), PtrValue(tt.b), UIntValue(bytecode.TypeU64, tt.count)})
		if err != nil || exit != nil {
			t.Fatalf("%s ret=%#v exit=%#v err=%v", tt.name, ret, exit, err)
		}
		got := int64(int32(ret.Int))
		if ret.Type != bytecode.TypeI32 || got != tt.want {
			t.Fatalf("%s ret=%#v, want i32 %d", tt.name, ret, tt.want)
		}
	}
}

func TestWideStringCharSearchExtern(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	fn, ok := reg.Lookup("wcschr")
	if !ok {
		t.Fatal("missing wcschr extern")
	}
	mem := NewMemory(bytecode.DefaultTarget())
	text := mustAllocBytes(t, mem, "wcschr:text", make([]byte, 16), false, blockLocal)
	for i, ch := range []uint32{'a', 0x1234, 'c', 0} {
		if err := storeWideChar(mem, text+uint64(i*4), ch); err != nil {
			t.Fatalf("store text %d: %v", i, err)
		}
	}
	tests := []struct {
		name string
		ch   int64
		want uint64
	}{
		{name: "hit", ch: 0x1234, want: text + 4},
		{name: "nul", ch: 0, want: text + 12},
		{name: "miss", ch: 'z', want: 0},
	}
	for _, tt := range tests {
		ret, exit, err := fn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(text), IntValue(bytecode.TypeI32, tt.ch)})
		if err != nil || exit != nil {
			t.Fatalf("%s ret=%#v exit=%#v err=%v", tt.name, ret, exit, err)
		}
		if ret.Type != bytecode.TypePtr || ret.Int != tt.want {
			t.Fatalf("%s ret=%#v, want pointer %#x", tt.name, ret, tt.want)
		}
	}
}

func TestWideStringReverseCharSearchExtern(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	fn, ok := reg.Lookup("wcsrchr")
	if !ok {
		t.Fatal("missing wcsrchr extern")
	}
	mem := NewMemory(bytecode.DefaultTarget())
	text := mustAllocBytes(t, mem, "wcsrchr:text", make([]byte, 24), false, blockLocal)
	for i, ch := range []uint32{'a', 0x1234, 'b', 0x1234, 'c', 0} {
		if err := storeWideChar(mem, text+uint64(i*4), ch); err != nil {
			t.Fatalf("store text %d: %v", i, err)
		}
	}
	tests := []struct {
		name string
		ch   int64
		want uint64
	}{
		{name: "last hit", ch: 0x1234, want: text + 12},
		{name: "nul", ch: 0, want: text + 20},
		{name: "miss", ch: 'z', want: 0},
	}
	for _, tt := range tests {
		ret, exit, err := fn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(text), IntValue(bytecode.TypeI32, tt.ch)})
		if err != nil || exit != nil {
			t.Fatalf("%s ret=%#v exit=%#v err=%v", tt.name, ret, exit, err)
		}
		if ret.Type != bytecode.TypePtr || ret.Int != tt.want {
			t.Fatalf("%s ret=%#v, want pointer %#x", tt.name, ret, tt.want)
		}
	}
}

func TestWideStringSearchExtern(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	fn, ok := reg.Lookup("wcsstr")
	if !ok {
		t.Fatal("missing wcsstr extern")
	}
	mem := NewMemory(bytecode.DefaultTarget())
	makeWide := func(name string, chars ...uint32) uint64 {
		t.Helper()
		addr := mustAllocBytes(t, mem, name, make([]byte, len(chars)*4), false, blockLocal)
		for i, ch := range chars {
			if err := storeWideChar(mem, addr+uint64(i*4), ch); err != nil {
				t.Fatalf("store %s[%d]: %v", name, i, err)
			}
		}
		return addr
	}
	haystack := makeWide("wcsstr:haystack", 'a', 0x1234, 'c', 'd', 0)
	empty := makeWide("wcsstr:empty", 0)
	full := makeWide("wcsstr:full", 'a', 0x1234, 'c', 'd', 0)
	middle := makeWide("wcsstr:middle", 0x1234, 'c', 0)
	suffix := makeWide("wcsstr:suffix", 'd', 0)
	miss := makeWide("wcsstr:miss", 'x', 0)
	tests := []struct {
		name   string
		needle uint64
		want   uint64
	}{
		{name: "empty", needle: empty, want: haystack},
		{name: "full", needle: full, want: haystack},
		{name: "middle", needle: middle, want: haystack + 4},
		{name: "suffix", needle: suffix, want: haystack + 12},
		{name: "miss", needle: miss, want: 0},
	}
	for _, tt := range tests {
		ret, exit, err := fn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(haystack), PtrValue(tt.needle)})
		if err != nil || exit != nil {
			t.Fatalf("%s ret=%#v exit=%#v err=%v", tt.name, ret, exit, err)
		}
		if ret.Type != bytecode.TypePtr || ret.Int != tt.want {
			t.Fatalf("%s ret=%#v, want pointer %#x", tt.name, ret, tt.want)
		}
	}
}

func TestWideStringSetSearchExtern(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	fn, ok := reg.Lookup("wcspbrk")
	if !ok {
		t.Fatal("missing wcspbrk extern")
	}
	mem := NewMemory(bytecode.DefaultTarget())
	makeWide := func(name string, chars ...uint32) uint64 {
		t.Helper()
		addr := mustAllocBytes(t, mem, name, make([]byte, len(chars)*4), false, blockLocal)
		for i, ch := range chars {
			if err := storeWideChar(mem, addr+uint64(i*4), ch); err != nil {
				t.Fatalf("store %s[%d]: %v", name, i, err)
			}
		}
		return addr
	}
	text := makeWide("wcspbrk:text", 'a', 'b', 0x1234, 'd', 0)
	highSet := makeWide("wcspbrk:high", 0x1234, 'x', 0)
	missSet := makeWide("wcspbrk:miss", 'x', 'y', 0)
	emptySet := makeWide("wcspbrk:empty", 0)
	tests := []struct {
		name string
		set  uint64
		want uint64
	}{
		{name: "hit", set: highSet, want: text + 8},
		{name: "miss", set: missSet, want: 0},
		{name: "empty", set: emptySet, want: 0},
	}
	for _, tt := range tests {
		ret, exit, err := fn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(text), PtrValue(tt.set)})
		if err != nil || exit != nil {
			t.Fatalf("%s ret=%#v exit=%#v err=%v", tt.name, ret, exit, err)
		}
		if ret.Type != bytecode.TypePtr || ret.Int != tt.want {
			t.Fatalf("%s ret=%#v, want pointer %#x", tt.name, ret, tt.want)
		}
	}
}

func TestWideStringSpanExtern(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	fn, ok := reg.Lookup("wcsspn")
	if !ok {
		t.Fatal("missing wcsspn extern")
	}
	mem := NewMemory(bytecode.DefaultTarget())
	makeWide := func(name string, chars ...uint32) uint64 {
		t.Helper()
		addr := mustAllocBytes(t, mem, name, make([]byte, len(chars)*4), false, blockLocal)
		for i, ch := range chars {
			if err := storeWideChar(mem, addr+uint64(i*4), ch); err != nil {
				t.Fatalf("store %s[%d]: %v", name, i, err)
			}
		}
		return addr
	}
	text := makeWide("wcsspn:text", 'a', 0x1234, 'c', 0)
	fullSet := makeWide("wcsspn:full", 'a', 0x1234, 'c', 0)
	partialSet := makeWide("wcsspn:partial", 'a', 0x1234, 0)
	emptySet := makeWide("wcsspn:empty", 0)
	emptyText := makeWide("wcsspn:empty-text", 0)
	tests := []struct {
		name string
		s    uint64
		set  uint64
		want uint64
	}{
		{name: "full", s: text, set: fullSet, want: 3},
		{name: "partial", s: text, set: partialSet, want: 2},
		{name: "empty set", s: text, set: emptySet, want: 0},
		{name: "empty text", s: emptyText, set: fullSet, want: 0},
	}
	for _, tt := range tests {
		ret, exit, err := fn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(tt.s), PtrValue(tt.set)})
		if err != nil || exit != nil {
			t.Fatalf("%s ret=%#v exit=%#v err=%v", tt.name, ret, exit, err)
		}
		if ret.Type != bytecode.TypeU64 || ret.Int != tt.want {
			t.Fatalf("%s ret=%#v, want u64 %d", tt.name, ret, tt.want)
		}
	}
}

func TestWideStringComplementSpanExtern(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	fn, ok := reg.Lookup("wcscspn")
	if !ok {
		t.Fatal("missing wcscspn extern")
	}
	mem := NewMemory(bytecode.DefaultTarget())
	makeWide := func(name string, chars ...uint32) uint64 {
		t.Helper()
		addr := mustAllocBytes(t, mem, name, make([]byte, len(chars)*4), false, blockLocal)
		for i, ch := range chars {
			if err := storeWideChar(mem, addr+uint64(i*4), ch); err != nil {
				t.Fatalf("store %s[%d]: %v", name, i, err)
			}
		}
		return addr
	}
	text := makeWide("wcscspn:text", 'a', 0x1234, 'c', 0)
	rejectHigh := makeWide("wcscspn:reject-high", 0x1234, 0)
	rejectMiss := makeWide("wcscspn:reject-miss", 'x', 0)
	emptyReject := makeWide("wcscspn:empty", 0)
	emptyText := makeWide("wcscspn:empty-text", 0)
	tests := []struct {
		name   string
		s      uint64
		reject uint64
		want   uint64
	}{
		{name: "early reject", s: text, reject: rejectHigh, want: 1},
		{name: "full", s: text, reject: rejectMiss, want: 3},
		{name: "empty reject", s: text, reject: emptyReject, want: 3},
		{name: "empty text", s: emptyText, reject: rejectHigh, want: 0},
	}
	for _, tt := range tests {
		ret, exit, err := fn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(tt.s), PtrValue(tt.reject)})
		if err != nil || exit != nil {
			t.Fatalf("%s ret=%#v exit=%#v err=%v", tt.name, ret, exit, err)
		}
		if ret.Type != bytecode.TypeU64 || ret.Int != tt.want {
			t.Fatalf("%s ret=%#v, want u64 %d", tt.name, ret, tt.want)
		}
	}
}

func TestWideStringCopyExtern(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	fn, ok := reg.Lookup("wcscpy")
	if !ok {
		t.Fatal("missing wcscpy extern")
	}
	mem := NewMemory(bytecode.DefaultTarget())
	makeWide := func(name string, chars ...uint32) uint64 {
		t.Helper()
		addr := mustAllocBytes(t, mem, name, make([]byte, len(chars)*4), false, blockLocal)
		for i, ch := range chars {
			if err := storeWideChar(mem, addr+uint64(i*4), ch); err != nil {
				t.Fatalf("store %s[%d]: %v", name, i, err)
			}
		}
		return addr
	}
	src := makeWide("wcscpy:src", 'A', 0x1234, 'Z', 0)
	dst := makeWide("wcscpy:dst", '?', '?', '?', '?')
	ret, exit, err := fn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(dst), PtrValue(src)})
	if err != nil || exit != nil {
		t.Fatalf("ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypePtr || ret.Int != dst {
		t.Fatalf("ret=%#v, want pointer %#x", ret, dst)
	}
	for i, want := range []uint32{'A', 0x1234, 'Z', 0} {
		got, err := loadWideChar(mem, dst+uint64(i*4))
		if err != nil {
			t.Fatalf("load dst[%d]: %v", i, err)
		}
		if got != want {
			t.Fatalf("dst[%d]=%#x, want %#x", i, got, want)
		}
	}

	empty := makeWide("wcscpy:empty", 0)
	emptyDst := makeWide("wcscpy:empty-dst", 'x', 'y')
	ret, exit, err = fn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(emptyDst), PtrValue(empty)})
	if err != nil || exit != nil {
		t.Fatalf("empty ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypePtr || ret.Int != emptyDst {
		t.Fatalf("empty ret=%#v, want pointer %#x", ret, emptyDst)
	}
	got, err := loadWideChar(mem, emptyDst)
	if err != nil {
		t.Fatalf("load empty dst: %v", err)
	}
	if got != 0 {
		t.Fatalf("empty dst[0]=%#x, want 0", got)
	}
}

func TestWideStringNCopyExtern(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	fn, ok := reg.Lookup("wcsncpy")
	if !ok {
		t.Fatal("missing wcsncpy extern")
	}
	mem := NewMemory(bytecode.DefaultTarget())
	makeWide := func(name string, chars ...uint32) uint64 {
		t.Helper()
		addr := mustAllocBytes(t, mem, name, make([]byte, len(chars)*4), false, blockLocal)
		for i, ch := range chars {
			if err := storeWideChar(mem, addr+uint64(i*4), ch); err != nil {
				t.Fatalf("store %s[%d]: %v", name, i, err)
			}
		}
		return addr
	}
	assertWide := func(name string, addr uint64, want ...uint32) {
		t.Helper()
		for i, ch := range want {
			got, err := loadWideChar(mem, addr+uint64(i*4))
			if err != nil {
				t.Fatalf("%s load[%d]: %v", name, i, err)
			}
			if got != ch {
				t.Fatalf("%s[%d]=%#x, want %#x", name, i, got, ch)
			}
		}
	}

	src := makeWide("wcsncpy:src", 'x', 0x1234, 0)
	dst := makeWide("wcsncpy:dst", '?', '?', '?', '?', '?')
	ret, exit, err := fn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(dst), PtrValue(src), UIntValue(bytecode.TypeU64, 4)})
	if err != nil || exit != nil {
		t.Fatalf("ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypePtr || ret.Int != dst {
		t.Fatalf("ret=%#v, want pointer %#x", ret, dst)
	}
	assertWide("padded dst", dst, 'x', 0x1234, 0, 0, '?')

	truncSrc := makeWide("wcsncpy:trunc-src", 'p', 'q', 'r', 0)
	truncDst := makeWide("wcsncpy:trunc-dst", '!', '!', '!', 0)
	ret, exit, err = fn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(truncDst), PtrValue(truncSrc), UIntValue(bytecode.TypeU64, 2)})
	if err != nil || exit != nil {
		t.Fatalf("trunc ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypePtr || ret.Int != truncDst {
		t.Fatalf("trunc ret=%#v, want pointer %#x", ret, truncDst)
	}
	assertWide("truncated dst", truncDst, 'p', 'q', '!', 0)

	zeroDst := makeWide("wcsncpy:zero-dst", 0x7777)
	ret, exit, err = fn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(zeroDst), PtrValue(0), UIntValue(bytecode.TypeU64, 0)})
	if err != nil || exit != nil {
		t.Fatalf("zero ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypePtr || ret.Int != zeroDst {
		t.Fatalf("zero ret=%#v, want pointer %#x", ret, zeroDst)
	}
	assertWide("zero-count dst", zeroDst, 0x7777)
}

func TestWideStringConcatExtern(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	fn, ok := reg.Lookup("wcscat")
	if !ok {
		t.Fatal("missing wcscat extern")
	}
	mem := NewMemory(bytecode.DefaultTarget())
	makeWide := func(name string, chars ...uint32) uint64 {
		t.Helper()
		addr := mustAllocBytes(t, mem, name, make([]byte, len(chars)*4), false, blockLocal)
		for i, ch := range chars {
			if err := storeWideChar(mem, addr+uint64(i*4), ch); err != nil {
				t.Fatalf("store %s[%d]: %v", name, i, err)
			}
		}
		return addr
	}
	assertWide := func(name string, addr uint64, want ...uint32) {
		t.Helper()
		for i, ch := range want {
			got, err := loadWideChar(mem, addr+uint64(i*4))
			if err != nil {
				t.Fatalf("%s load[%d]: %v", name, i, err)
			}
			if got != ch {
				t.Fatalf("%s[%d]=%#x, want %#x", name, i, got, ch)
			}
		}
	}

	dst := makeWide("wcscat:dst", 'A', 0, '?', '?', '?')
	src := makeWide("wcscat:src", 0x1234, 'Z', 0)
	ret, exit, err := fn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(dst), PtrValue(src)})
	if err != nil || exit != nil {
		t.Fatalf("ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypePtr || ret.Int != dst {
		t.Fatalf("ret=%#v, want pointer %#x", ret, dst)
	}
	assertWide("appended dst", dst, 'A', 0x1234, 'Z', 0, '?')

	empty := makeWide("wcscat:empty", 0)
	ret, exit, err = fn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(dst), PtrValue(empty)})
	if err != nil || exit != nil {
		t.Fatalf("empty ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypePtr || ret.Int != dst {
		t.Fatalf("empty ret=%#v, want pointer %#x", ret, dst)
	}
	assertWide("empty append dst", dst, 'A', 0x1234, 'Z', 0, '?')
}

func TestWideStringNConcatExtern(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	fn, ok := reg.Lookup("wcsncat")
	if !ok {
		t.Fatal("missing wcsncat extern")
	}
	mem := NewMemory(bytecode.DefaultTarget())
	makeWide := func(name string, chars ...uint32) uint64 {
		t.Helper()
		addr := mustAllocBytes(t, mem, name, make([]byte, len(chars)*4), false, blockLocal)
		for i, ch := range chars {
			if err := storeWideChar(mem, addr+uint64(i*4), ch); err != nil {
				t.Fatalf("store %s[%d]: %v", name, i, err)
			}
		}
		return addr
	}
	assertWide := func(name string, addr uint64, want ...uint32) {
		t.Helper()
		for i, ch := range want {
			got, err := loadWideChar(mem, addr+uint64(i*4))
			if err != nil {
				t.Fatalf("%s load[%d]: %v", name, i, err)
			}
			if got != ch {
				t.Fatalf("%s[%d]=%#x, want %#x", name, i, got, ch)
			}
		}
	}

	dst := makeWide("wcsncat:dst", 'A', 0, '?', '?', '?')
	src := makeWide("wcsncat:src", 0x1234, 'Z', 'Q', 0)
	ret, exit, err := fn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(dst), PtrValue(src), UIntValue(bytecode.TypeU64, 2)})
	if err != nil || exit != nil {
		t.Fatalf("ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypePtr || ret.Int != dst {
		t.Fatalf("ret=%#v, want pointer %#x", ret, dst)
	}
	assertWide("limited append dst", dst, 'A', 0x1234, 'Z', 0, '?')

	emptyDst := makeWide("wcsncat:empty-dst", 'x', 0, '?')
	empty := makeWide("wcsncat:empty", 0)
	ret, exit, err = fn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(emptyDst), PtrValue(empty), UIntValue(bytecode.TypeU64, 3)})
	if err != nil || exit != nil {
		t.Fatalf("empty ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypePtr || ret.Int != emptyDst {
		t.Fatalf("empty ret=%#v, want pointer %#x", ret, emptyDst)
	}
	assertWide("empty append dst", emptyDst, 'x', 0, '?')

	zeroDst := makeWide("wcsncat:zero-dst", 'z', 0, '?')
	ret, exit, err = fn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(zeroDst), PtrValue(0), UIntValue(bytecode.TypeU64, 0)})
	if err != nil || exit != nil {
		t.Fatalf("zero ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypePtr || ret.Int != zeroDst {
		t.Fatalf("zero ret=%#v, want pointer %#x", ret, zeroDst)
	}
	assertWide("zero append dst", zeroDst, 'z', 0, '?')
}

func TestWideStringTransformExtern(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	fn, ok := reg.Lookup("wcsxfrm")
	if !ok {
		t.Fatal("missing wcsxfrm extern")
	}
	mem := NewMemory(bytecode.DefaultTarget())
	makeWide := func(name string, chars ...uint32) uint64 {
		t.Helper()
		addr := mustAllocBytes(t, mem, name, make([]byte, len(chars)*4), false, blockLocal)
		for i, ch := range chars {
			if err := storeWideChar(mem, addr+uint64(i*4), ch); err != nil {
				t.Fatalf("store %s[%d]: %v", name, i, err)
			}
		}
		return addr
	}
	assertWide := func(name string, addr uint64, want ...uint32) {
		t.Helper()
		for i, ch := range want {
			got, err := loadWideChar(mem, addr+uint64(i*4))
			if err != nil {
				t.Fatalf("%s load[%d]: %v", name, i, err)
			}
			if got != ch {
				t.Fatalf("%s[%d]=%#x, want %#x", name, i, got, ch)
			}
		}
	}

	src := makeWide("wcsxfrm:src", 'a', 0x1234, 'z', 0)
	dst := makeWide("wcsxfrm:dst", '?', '?', '?', '?')
	ret, exit, err := fn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(dst), PtrValue(src), UIntValue(bytecode.TypeU64, 3)})
	if err != nil || exit != nil {
		t.Fatalf("ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeU64 || ret.Int != 3 {
		t.Fatalf("ret=%#v, want size 3", ret)
	}
	assertWide("truncated dst", dst, 'a', 0x1234, 0, '?')

	exact := makeWide("wcsxfrm:exact", '?', '?', '?', '?')
	ret, exit, err = fn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(exact), PtrValue(src), UIntValue(bytecode.TypeU64, 4)})
	if err != nil || exit != nil {
		t.Fatalf("exact ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeU64 || ret.Int != 3 {
		t.Fatalf("exact ret=%#v, want size 3", ret)
	}
	assertWide("exact dst", exact, 'a', 0x1234, 'z', 0)

	ret, exit, err = fn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(0), PtrValue(src), UIntValue(bytecode.TypeU64, 0)})
	if err != nil || exit != nil {
		t.Fatalf("zero ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeU64 || ret.Int != 3 {
		t.Fatalf("zero ret=%#v, want size 3", ret)
	}
}

func TestWideStringTokenExtern(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	fn, ok := reg.Lookup("wcstok")
	if !ok {
		t.Fatal("missing wcstok extern")
	}
	mem := NewMemory(bytecode.DefaultTarget())
	makeWide := func(name string, chars ...uint32) uint64 {
		t.Helper()
		addr := mustAllocBytes(t, mem, name, make([]byte, len(chars)*4), false, blockLocal)
		for i, ch := range chars {
			if err := storeWideChar(mem, addr+uint64(i*4), ch); err != nil {
				t.Fatalf("store %s[%d]: %v", name, i, err)
			}
		}
		return addr
	}
	assertToken := func(name string, addr uint64, want ...uint32) {
		t.Helper()
		for i, ch := range append(want, 0) {
			got, err := loadWideChar(mem, addr+uint64(i*4))
			if err != nil {
				t.Fatalf("%s load[%d]: %v", name, i, err)
			}
			if got != ch {
				t.Fatalf("%s[%d]=%#x, want %#x", name, i, got, ch)
			}
		}
	}
	wideAt := func(base uint64, index int64) uint64 {
		t.Helper()
		addr, err := wideElementAddr(base, index)
		if err != nil {
			t.Fatalf("wide address %d: %v", index, err)
		}
		return addr
	}
	savePtr := func(name string) uint64 {
		t.Helper()
		return mustAlloc(t, mem, name, mem.target.PointerSize, mem.target.PointerAlign, false, blockLocal)
	}
	loadSave := func(name string, addr uint64) uint64 {
		t.Helper()
		got, err := mem.Load(addr, bytecode.TypePtr, mem.target.PointerAlign)
		if err != nil {
			t.Fatalf("%s load save: %v", name, err)
		}
		return got.Int
	}
	call := func(name string, s uint64, delim uint64, save uint64) uint64 {
		t.Helper()
		ret, exit, err := fn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(s), PtrValue(delim), PtrValue(save)})
		if err != nil || exit != nil {
			t.Fatalf("%s ret=%#v exit=%#v err=%v", name, ret, exit, err)
		}
		if ret.Type != bytecode.TypePtr {
			t.Fatalf("%s ret=%#v, want pointer", name, ret)
		}
		return ret.Int
	}

	delims := makeWide("wcstok:delims", ',', ';', 0)
	text := makeWide("wcstok:text", ',', ',', 'a', 'l', 'p', 'h', 'a', ',', 'b', 'e', 't', 'a', ';', 'g', 'a', 'm', 'm', 'a', 0)
	save := savePtr("wcstok:save")
	tok := call("first", text, delims, save)
	if tok != wideAt(text, 2) || loadSave("first", save) != wideAt(text, 8) {
		t.Fatalf("first tok=%#x save=%#x", tok, loadSave("first", save))
	}
	assertToken("first token", tok, 'a', 'l', 'p', 'h', 'a')
	tok = call("second", 0, delims, save)
	if tok != wideAt(text, 8) || loadSave("second", save) != wideAt(text, 13) {
		t.Fatalf("second tok=%#x save=%#x", tok, loadSave("second", save))
	}
	assertToken("second token", tok, 'b', 'e', 't', 'a')
	tok = call("third", 0, delims, save)
	if tok != wideAt(text, 13) || loadSave("third", save) != 0 {
		t.Fatalf("third tok=%#x save=%#x", tok, loadSave("third", save))
	}
	assertToken("third token", tok, 'g', 'a', 'm', 'm', 'a')
	if tok = call("exhausted", 0, delims, save); tok != 0 {
		t.Fatalf("exhausted tok=%#x, want null", tok)
	}

	text2 := makeWide("wcstok:changing", 'a', 'l', 'p', 'h', 'a', ',', 'b', 'e', 't', 'a', ';', 'g', 'a', 'm', 'm', 'a', '.', 'd', 'e', 'l', 't', 'a', 0)
	comma := makeWide("wcstok:comma", ',', 0)
	semicolon := makeWide("wcstok:semicolon", ';', 0)
	dot := makeWide("wcstok:dot", '.', 0)
	save2 := savePtr("wcstok:save-changing")
	if tok = call("changing first", text2, comma, save2); tok != text2 {
		t.Fatalf("changing first tok=%#x, want %#x", tok, text2)
	}
	if tok = call("changing second", 0, semicolon, save2); tok != wideAt(text2, 6) {
		t.Fatalf("changing second tok=%#x", tok)
	}
	if tok = call("changing third", 0, dot, save2); tok != wideAt(text2, 11) {
		t.Fatalf("changing third tok=%#x", tok)
	}
	if tok = call("changing fourth", 0, dot, save2); tok != wideAt(text2, 17) || loadSave("changing fourth", save2) != 0 {
		t.Fatalf("changing fourth tok=%#x save=%#x", tok, loadSave("changing fourth", save2))
	}

	emptyDelims := makeWide("wcstok:empty-delims", 0)
	whole := makeWide("wcstok:whole", 'w', 'i', 'd', 'e', 0)
	save3 := savePtr("wcstok:save-empty")
	tok = call("empty delims", whole, emptyDelims, save3)
	if tok != whole || loadSave("empty delims", save3) != 0 {
		t.Fatalf("empty delimiter tok=%#x save=%#x", tok, loadSave("empty delims", save3))
	}
	assertToken("empty delimiter token", tok, 'w', 'i', 'd', 'e')
}

func TestStringReverseAndSetSearchExterns(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	text := mustAllocBytes(t, mem, "strrchr:text", []byte("abacad\x00"), true, blockString)
	strrchrFn, ok := reg.Lookup("strrchr")
	if !ok {
		t.Fatal("missing strrchr extern")
	}
	ret, exit, err := strrchrFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(text), IntValue(bytecode.TypeI32, 'a')})
	if err != nil || exit != nil {
		t.Fatalf("strrchr hit ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypePtr || ret.Int != text+4 {
		t.Fatalf("strrchr hit ret=%#v, want pointer %#x", ret, text+4)
	}
	ret, exit, err = strrchrFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(text), IntValue(bytecode.TypeI32, 0)})
	if err != nil || exit != nil {
		t.Fatalf("strrchr nul ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypePtr || ret.Int != text+6 {
		t.Fatalf("strrchr nul ret=%#v, want pointer %#x", ret, text+6)
	}
	ret, exit, err = strrchrFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(text), IntValue(bytecode.TypeI32, 'z')})
	if err != nil || exit != nil {
		t.Fatalf("strrchr miss ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypePtr || ret.Int != 0 {
		t.Fatalf("strrchr miss ret=%#v, want null pointer", ret)
	}
	highText := mustAllocBytes(t, mem, "strrchr:high", []byte{0x82, 'a', 0x82, 0}, true, blockString)
	ret, exit, err = strrchrFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(highText), IntValue(bytecode.TypeI32, 0x182)})
	if err != nil || exit != nil {
		t.Fatalf("strrchr masked ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypePtr || ret.Int != highText+2 {
		t.Fatalf("strrchr masked ret=%#v, want pointer %#x", ret, highText+2)
	}

	strpbrkFn, ok := reg.Lookup("strpbrk")
	if !ok {
		t.Fatal("missing strpbrk extern")
	}
	first := mustAllocBytes(t, mem, "strpbrk:first", []byte("a\x00"), true, blockString)
	ret, exit, err = strpbrkFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(text), PtrValue(first)})
	if err != nil || exit != nil {
		t.Fatalf("strpbrk first ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypePtr || ret.Int != text {
		t.Fatalf("strpbrk first ret=%#v, want pointer %#x", ret, text)
	}
	accept := mustAllocBytes(t, mem, "strpbrk:accept", []byte("xyc\x00"), true, blockString)
	ret, exit, err = strpbrkFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(text), PtrValue(accept)})
	if err != nil || exit != nil {
		t.Fatalf("strpbrk hit ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypePtr || ret.Int != text+3 {
		t.Fatalf("strpbrk hit ret=%#v, want pointer %#x", ret, text+3)
	}
	none := mustAllocBytes(t, mem, "strpbrk:none", []byte("xyz\x00"), true, blockString)
	ret, exit, err = strpbrkFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(text), PtrValue(none)})
	if err != nil || exit != nil {
		t.Fatalf("strpbrk miss ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypePtr || ret.Int != 0 {
		t.Fatalf("strpbrk miss ret=%#v, want null pointer", ret)
	}
	empty := mustAllocBytes(t, mem, "strpbrk:empty", []byte{0}, true, blockString)
	ret, exit, err = strpbrkFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(text), PtrValue(empty)})
	if err != nil || exit != nil {
		t.Fatalf("strpbrk empty ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypePtr || ret.Int != 0 {
		t.Fatalf("strpbrk empty ret=%#v, want null pointer", ret)
	}
}

func TestStringSpanExterns(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	text := mustAllocBytes(t, mem, "span:text", []byte("abcde312\x00"), true, blockString)
	abc := mustAllocBytes(t, mem, "span:abc", []byte("abc\x00"), true, blockString)
	de := mustAllocBytes(t, mem, "span:de", []byte("de\x00"), true, blockString)
	a := mustAllocBytes(t, mem, "span:a", []byte("a\x00"), true, blockString)
	xyz := mustAllocBytes(t, mem, "span:xyz", []byte("xyz\x00"), true, blockString)
	empty := mustAllocBytes(t, mem, "span:empty", []byte{0}, true, blockString)

	strspnFn, ok := reg.Lookup("strspn")
	if !ok {
		t.Fatal("missing strspn extern")
	}
	ret, exit, err := strspnFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(text), PtrValue(abc)})
	if err != nil || exit != nil {
		t.Fatalf("strspn hit ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeU64 || ret.Int != 3 {
		t.Fatalf("strspn hit ret=%#v, want size 3", ret)
	}
	ret, exit, err = strspnFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(text), PtrValue(xyz)})
	if err != nil || exit != nil {
		t.Fatalf("strspn miss ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeU64 || ret.Int != 0 {
		t.Fatalf("strspn miss ret=%#v, want size 0", ret)
	}
	ret, exit, err = strspnFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(text), PtrValue(empty)})
	if err != nil || exit != nil {
		t.Fatalf("strspn empty ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeU64 || ret.Int != 0 {
		t.Fatalf("strspn empty ret=%#v, want size 0", ret)
	}

	strcspnFn, ok := reg.Lookup("strcspn")
	if !ok {
		t.Fatal("missing strcspn extern")
	}
	ret, exit, err = strcspnFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(text), PtrValue(de)})
	if err != nil || exit != nil {
		t.Fatalf("strcspn hit ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeU64 || ret.Int != 3 {
		t.Fatalf("strcspn hit ret=%#v, want size 3", ret)
	}
	ret, exit, err = strcspnFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(text), PtrValue(a)})
	if err != nil || exit != nil {
		t.Fatalf("strcspn immediate ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeU64 || ret.Int != 0 {
		t.Fatalf("strcspn immediate ret=%#v, want size 0", ret)
	}
	ret, exit, err = strcspnFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(text), PtrValue(xyz)})
	if err != nil || exit != nil {
		t.Fatalf("strcspn miss ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeU64 || ret.Int != 8 {
		t.Fatalf("strcspn miss ret=%#v, want size 8", ret)
	}
	ret, exit, err = strcspnFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(text), PtrValue(empty)})
	if err != nil || exit != nil {
		t.Fatalf("strcspn empty ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeU64 || ret.Int != 8 {
		t.Fatalf("strcspn empty ret=%#v, want size 8", ret)
	}
}

func TestStringStrtokExtern(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	fn, ok := reg.Lookup("strtok")
	if !ok {
		t.Fatal("missing strtok extern")
	}
	text := mustAllocBytes(t, mem, "strtok:text", []byte(",,alpha,beta;gamma\x00"), false, blockString)
	delims := mustAllocBytes(t, mem, "strtok:delims", []byte(",;\x00"), true, blockString)

	ret, exit, err := fn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(text), PtrValue(delims)})
	if err != nil || exit != nil {
		t.Fatalf("strtok first ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypePtr || ret.Int != text+2 {
		t.Fatalf("strtok first ret=%#v, want pointer %#x", ret, text+2)
	}
	first, err := mem.ReadCString(ret.Int)
	if err != nil || first != "alpha" {
		t.Fatalf("strtok first token=%q err=%v, want alpha", first, err)
	}

	ret, exit, err = fn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(0), PtrValue(delims)})
	if err != nil || exit != nil {
		t.Fatalf("strtok second ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypePtr || ret.Int != text+8 {
		t.Fatalf("strtok second ret=%#v, want pointer %#x", ret, text+8)
	}
	second, err := mem.ReadCString(ret.Int)
	if err != nil || second != "beta" {
		t.Fatalf("strtok second token=%q err=%v, want beta", second, err)
	}

	ret, exit, err = fn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(0), PtrValue(delims)})
	if err != nil || exit != nil {
		t.Fatalf("strtok third ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypePtr || ret.Int != text+13 {
		t.Fatalf("strtok third ret=%#v, want pointer %#x", ret, text+13)
	}
	third, err := mem.ReadCString(ret.Int)
	if err != nil || third != "gamma" {
		t.Fatalf("strtok third token=%q err=%v, want gamma", third, err)
	}

	ret, exit, err = fn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(0), PtrValue(delims)})
	if err != nil || exit != nil {
		t.Fatalf("strtok end ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypePtr || ret.Int != 0 {
		t.Fatalf("strtok end ret=%#v, want null pointer", ret)
	}

	changing := mustAllocBytes(t, mem, "strtok:changing", []byte("alpha,beta;gamma.delta\x00"), false, blockString)
	comma := mustAllocBytes(t, mem, "strtok:comma", []byte(",\x00"), true, blockString)
	semicolon := mustAllocBytes(t, mem, "strtok:semicolon", []byte(";\x00"), true, blockString)
	dot := mustAllocBytes(t, mem, "strtok:dot", []byte(".\x00"), true, blockString)

	ret, exit, err = fn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(changing), PtrValue(comma)})
	if err != nil || exit != nil {
		t.Fatalf("strtok changing first ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypePtr || ret.Int != changing {
		t.Fatalf("strtok changing first ret=%#v, want pointer %#x", ret, changing)
	}
	if tok, err := mem.ReadCString(ret.Int); err != nil || tok != "alpha" {
		t.Fatalf("strtok changing first token=%q err=%v, want alpha", tok, err)
	}
	ret, exit, err = fn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(0), PtrValue(semicolon)})
	if err != nil || exit != nil {
		t.Fatalf("strtok changing second ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypePtr || ret.Int != changing+6 {
		t.Fatalf("strtok changing second ret=%#v, want pointer %#x", ret, changing+6)
	}
	if tok, err := mem.ReadCString(ret.Int); err != nil || tok != "beta" {
		t.Fatalf("strtok changing second token=%q err=%v, want beta", tok, err)
	}
	ret, exit, err = fn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(0), PtrValue(dot)})
	if err != nil || exit != nil {
		t.Fatalf("strtok changing third ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypePtr || ret.Int != changing+11 {
		t.Fatalf("strtok changing third ret=%#v, want pointer %#x", ret, changing+11)
	}
	if tok, err := mem.ReadCString(ret.Int); err != nil || tok != "gamma" {
		t.Fatalf("strtok changing third token=%q err=%v, want gamma", tok, err)
	}

	allDelims := mustAllocBytes(t, mem, "strtok:all-delims", []byte(";;;\x00"), false, blockString)
	ret, exit, err = fn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(allDelims), PtrValue(semicolon)})
	if err != nil || exit != nil {
		t.Fatalf("strtok all-delims ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypePtr || ret.Int != 0 {
		t.Fatalf("strtok all-delims ret=%#v, want null pointer", ret)
	}
	ret, exit, err = fn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(0), PtrValue(semicolon)})
	if err != nil || exit != nil {
		t.Fatalf("strtok exhausted ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypePtr || ret.Int != 0 {
		t.Fatalf("strtok exhausted ret=%#v, want null pointer", ret)
	}

	restart := mustAllocBytes(t, mem, "strtok:restart", []byte("one,two\x00"), false, blockString)
	ret, exit, err = fn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(restart), PtrValue(comma)})
	if err != nil || exit != nil {
		t.Fatalf("strtok restart first ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypePtr || ret.Int != restart {
		t.Fatalf("strtok restart first ret=%#v, want pointer %#x", ret, restart)
	}
	if tok, err := mem.ReadCString(ret.Int); err != nil || tok != "one" {
		t.Fatalf("strtok restart first token=%q err=%v, want one", tok, err)
	}
	ret, exit, err = fn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(0), PtrValue(comma)})
	if err != nil || exit != nil {
		t.Fatalf("strtok restart second ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypePtr || ret.Int != restart+4 {
		t.Fatalf("strtok restart second ret=%#v, want pointer %#x", ret, restart+4)
	}
	if tok, err := mem.ReadCString(ret.Int); err != nil || tok != "two" {
		t.Fatalf("strtok restart second token=%q err=%v, want two", tok, err)
	}
}

func TestStringCollateAndTransformExterns(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	left := mustAllocBytes(t, mem, "strcoll:left", []byte("abc\x00"), true, blockString)
	right := mustAllocBytes(t, mem, "strcoll:right", []byte("abd\x00"), true, blockString)

	strcollFn, ok := reg.Lookup("strcoll")
	if !ok {
		t.Fatal("missing strcoll extern")
	}
	strcollTests := []struct {
		name string
		a    uint64
		b    uint64
		sign int
	}{
		{name: "equal", a: left, b: left, sign: 0},
		{name: "less", a: left, b: right, sign: -1},
		{name: "greater", a: right, b: left, sign: 1},
		{name: "prefix less", a: mustAllocBytes(t, mem, "strcoll:prefix-less", []byte("ab\x00"), true, blockString), b: left, sign: -1},
		{name: "prefix greater", a: left, b: mustAllocBytes(t, mem, "strcoll:prefix-greater", []byte("ab\x00"), true, blockString), sign: 1},
	}
	for _, tt := range strcollTests {
		t.Run("strcoll "+tt.name, func(t *testing.T) {
			ret, exit, err := strcollFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(tt.a), PtrValue(tt.b)})
			if err != nil || exit != nil {
				t.Fatalf("strcoll ret=%#v exit=%#v err=%v", ret, exit, err)
			}
			got := signedInt(ret)
			if ret.Type != bytecode.TypeI32 || (tt.sign == 0 && got != 0) || (tt.sign < 0 && got >= 0) || (tt.sign > 0 && got <= 0) {
				t.Fatalf("strcoll ret=%#v, want sign %d", ret, tt.sign)
			}
		})
	}

	strxfrmFn, ok := reg.Lookup("strxfrm")
	if !ok {
		t.Fatal("missing strxfrm extern")
	}
	src := mustAllocBytes(t, mem, "strxfrm:src", []byte("abcdef\x00"), true, blockString)
	dst := mustAllocBytes(t, mem, "strxfrm:dst", []byte("xxxx\x00"), false, blockString)
	ret, exit, err := strxfrmFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(dst), PtrValue(src), UIntValue(bytecode.TypeU64, 4)})
	if err != nil || exit != nil {
		t.Fatalf("strxfrm ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeU64 || ret.Int != 6 {
		t.Fatalf("strxfrm ret=%#v, want size 6", ret)
	}
	got, err := mem.ReadCString(dst)
	if err != nil || got != "abc" {
		t.Fatalf("strxfrm dst=%q err=%v, want abc", got, err)
	}

	exact := mustAllocBytes(t, mem, "strxfrm:exact", make([]byte, 7), false, blockString)
	ret, exit, err = strxfrmFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(exact), PtrValue(src), UIntValue(bytecode.TypeU64, 7)})
	if err != nil || exit != nil {
		t.Fatalf("strxfrm exact ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeU64 || ret.Int != 6 {
		t.Fatalf("strxfrm exact ret=%#v, want size 6", ret)
	}
	got, err = mem.ReadCString(exact)
	if err != nil || got != "abcdef" {
		t.Fatalf("strxfrm exact dst=%q err=%v, want abcdef", got, err)
	}

	ret, exit, err = strxfrmFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(0), PtrValue(src), UIntValue(bytecode.TypeU64, 0)})
	if err != nil || exit != nil {
		t.Fatalf("strxfrm zero ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeU64 || ret.Int != 6 {
		t.Fatalf("strxfrm zero ret=%#v, want size 6", ret)
	}
}

func TestStringNLengthExtern(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	fn, ok := reg.Lookup("strnlen")
	if !ok {
		t.Fatal("missing strnlen extern")
	}

	ret, exit, err := fn(context.Background(), &ExternContext{Memory: mem}, []Value{
		PtrValue(0xdeadbeef),
		UIntValue(bytecode.TypeU64, 0),
	})
	if err != nil || exit != nil {
		t.Fatalf("strnlen zero ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeU64 || ret.Int != 0 {
		t.Fatalf("strnlen zero ret=%#v, want u64 0", ret)
	}

	noNull := mustAllocBytes(t, mem, "strnlen:no-null", []byte("abcdef"), true, blockString)
	ret, exit, err = fn(context.Background(), &ExternContext{Memory: mem}, []Value{
		PtrValue(noNull),
		UIntValue(bytecode.TypeU64, 3),
	})
	if err != nil || exit != nil {
		t.Fatalf("strnlen no-null ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeU64 || ret.Int != 3 {
		t.Fatalf("strnlen no-null ret=%#v, want u64 3", ret)
	}
	ret, exit, err = fn(context.Background(), &ExternContext{Memory: mem}, []Value{
		PtrValue(noNull),
		UIntValue(bytecode.TypeU64, 6),
	})
	if err != nil || exit != nil {
		t.Fatalf("strnlen exact no-null ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeU64 || ret.Int != 6 {
		t.Fatalf("strnlen exact no-null ret=%#v, want u64 6", ret)
	}

	withNull := mustAllocBytes(t, mem, "strnlen:with-null", []byte("ab\x00cd"), true, blockString)
	ret, exit, err = fn(context.Background(), &ExternContext{Memory: mem}, []Value{
		PtrValue(withNull),
		UIntValue(bytecode.TypeU64, 5),
	})
	if err != nil || exit != nil {
		t.Fatalf("strnlen with-null ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeU64 || ret.Int != 2 {
		t.Fatalf("strnlen with-null ret=%#v, want u64 2", ret)
	}
}

func TestStringStrerrorExtern(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	fn, ok := reg.Lookup("strerror")
	if !ok {
		t.Fatal("missing strerror extern")
	}
	ret, exit, err := fn(context.Background(), &ExternContext{Memory: mem}, []Value{IntValue(bytecode.TypeI32, 2)})
	if err != nil || exit != nil {
		t.Fatalf("strerror ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypePtr || ret.Int == 0 {
		t.Fatalf("strerror ret=%#v, want non-null pointer", ret)
	}
	got, err := mem.ReadCString(ret.Int)
	if err != nil {
		t.Fatalf("ReadCString strerror: %v", err)
	}
	if got != "error" {
		t.Fatalf("strerror string = %q, want error", got)
	}
	again, exit, err := fn(context.Background(), &ExternContext{Memory: mem}, []Value{IntValue(bytecode.TypeI32, 0)})
	if err != nil || exit != nil {
		t.Fatalf("strerror again ret=%#v exit=%#v err=%v", again, exit, err)
	}
	if again.Int != ret.Int {
		t.Fatalf("strerror stable pointer = %#x, want %#x", again.Int, ret.Int)
	}
	for _, errno := range []int64{1, 2, 123} {
		next, exit, err := fn(context.Background(), &ExternContext{Memory: mem}, []Value{IntValue(bytecode.TypeI32, errno)})
		if err != nil || exit != nil {
			t.Fatalf("strerror errno %d ret=%#v exit=%#v err=%v", errno, next, exit, err)
		}
		if next.Int != ret.Int {
			t.Fatalf("strerror errno %d pointer = %#x, want stable pointer %#x", errno, next.Int, ret.Int)
		}
	}

	otherMem := NewMemory(bytecode.DefaultTarget())
	other, exit, err := fn(context.Background(), &ExternContext{Memory: otherMem}, []Value{IntValue(bytecode.TypeI32, 2)})
	if err != nil || exit != nil {
		t.Fatalf("strerror other memory ret=%#v exit=%#v err=%v", other, exit, err)
	}
	otherAgain, exit, err := fn(context.Background(), &ExternContext{Memory: otherMem}, []Value{IntValue(bytecode.TypeI32, 0)})
	if err != nil || exit != nil {
		t.Fatalf("strerror other memory again ret=%#v exit=%#v err=%v", otherAgain, exit, err)
	}
	if other.Type != bytecode.TypePtr || other.Int == 0 || otherAgain.Int != other.Int {
		t.Fatalf("strerror other memory pointers = %#v then %#v, want stable non-null", other, otherAgain)
	}
	otherString, err := otherMem.ReadCString(other.Int)
	if err != nil || otherString != "error" {
		t.Fatalf("strerror other memory string=%q err=%v, want error", otherString, err)
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
	high := mustAllocBytes(t, mem, "memcmp:high", []byte{0xff, 0}, true, blockString)
	low := mustAllocBytes(t, mem, "memcmp:low", []byte{0x7f, 0}, true, blockString)
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
	ret, exit, err := fn(context.Background(), &ExternContext{Memory: mem}, []Value{ObjectAddrValue(high), ObjectAddrValue(low), UIntValue(bytecode.TypeU64, 1)})
	if err != nil || exit != nil {
		t.Fatalf("memcmp unsigned ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeI32 || int64(int32(ret.Int)) <= 0 {
		t.Fatalf("memcmp unsigned = %#v, want positive i32", ret)
	}
}

func TestSscanfExternScansBasicIntegerStringAndChar(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	fn, ok := reg.Lookup("sscanf")
	if !ok {
		t.Fatal("missing sscanf extern")
	}

	inputAddr := mustAllocBytes(t, mem, "sscanf:input", []byte(" -12 34 hello Z\x00"), true, blockString)
	fmtAddr := mustAllocBytes(t, mem, "sscanf:fmt", []byte("%d %u %5s %c\x00"), true, blockString)
	signedAddr := mustAlloc(t, mem, "sscanf:signed", 4, 4, false, blockLocal)
	unsignedAddr := mustAlloc(t, mem, "sscanf:unsigned", 4, 4, false, blockLocal)
	wordAddr := mustAllocBytes(t, mem, "sscanf:word", []byte{0, 0, 0, 0, 0, 0}, false, blockLocal)
	charAddr := mustAllocBytes(t, mem, "sscanf:char", []byte{0}, false, blockLocal)

	ret, exit, err := fn(context.Background(), &ExternContext{Memory: mem}, []Value{
		ObjectAddrValue(inputAddr),
		ObjectAddrValue(fmtAddr),
		PtrValue(signedAddr),
		PtrValue(unsignedAddr),
		PtrValue(wordAddr),
		PtrValue(charAddr),
	})
	if err != nil || exit != nil {
		t.Fatalf("sscanf ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeI32 || ret.Int != 4 {
		t.Fatalf("sscanf ret=%#v, want i32 4", ret)
	}
	signed, err := mem.Load(signedAddr, bytecode.TypeI32, 4)
	if err != nil || int32(signed.Int) != -12 {
		t.Fatalf("signed=%#v err=%v, want -12", signed, err)
	}
	unsigned, err := mem.Load(unsignedAddr, bytecode.TypeU32, 4)
	if err != nil || uint32(unsigned.Int) != 34 {
		t.Fatalf("unsigned=%#v err=%v, want 34", unsigned, err)
	}
	word, err := mem.ReadCString(wordAddr)
	if err != nil || word != "hello" {
		t.Fatalf("word=%q err=%v, want hello", word, err)
	}
	ch, err := mem.Load(charAddr, bytecode.TypeI8, 1)
	if err != nil || byte(ch.Int) != 'Z' {
		t.Fatalf("char=%#v err=%v, want Z", ch, err)
	}
}

func TestSscanfExternScansSignedIntegerLengthModifiers(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	fn, ok := reg.Lookup("sscanf")
	if !ok {
		t.Fatal("missing sscanf extern")
	}

	inputAddr := mustAllocBytes(t, mem, "sscanf:signed-length-input", []byte("-1 -2 -3 -4 -5 -6 -7 -8\x00"), true, blockString)
	fmtAddr := mustAllocBytes(t, mem, "sscanf:signed-length-fmt", []byte("%hhd %hd %d %ld %lld %jd %zd %td\x00"), true, blockString)
	count8 := mustAlloc(t, mem, "sscanf:signed-length-i8", 1, 1, false, blockLocal)
	count16 := mustAlloc(t, mem, "sscanf:signed-length-i16", 2, 2, false, blockLocal)
	count32 := mustAlloc(t, mem, "sscanf:signed-length-i32", 4, 4, false, blockLocal)
	countLong := mustAlloc(t, mem, "sscanf:signed-length-long", 8, 8, false, blockLocal)
	countLL := mustAlloc(t, mem, "sscanf:signed-length-ll", 8, 8, false, blockLocal)
	countJ := mustAlloc(t, mem, "sscanf:signed-length-j", 8, 8, false, blockLocal)
	countZ := mustAlloc(t, mem, "sscanf:signed-length-z", 8, 8, false, blockLocal)
	countT := mustAlloc(t, mem, "sscanf:signed-length-t", 8, 8, false, blockLocal)

	ret, exit, err := fn(context.Background(), &ExternContext{Memory: mem}, []Value{
		ObjectAddrValue(inputAddr),
		ObjectAddrValue(fmtAddr),
		PtrValue(count8),
		PtrValue(count16),
		PtrValue(count32),
		PtrValue(countLong),
		PtrValue(countLL),
		PtrValue(countJ),
		PtrValue(countZ),
		PtrValue(countT),
	})
	if err != nil || exit != nil {
		t.Fatalf("sscanf ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeI32 || ret.Int != 8 {
		t.Fatalf("sscanf ret=%#v, want i32 8", ret)
	}
	check := func(name string, addr uint64, typ bytecode.ValueType, align int64, want int64) {
		t.Helper()
		got, err := mem.Load(addr, typ, align)
		if err != nil || signedInt(got) != want {
			t.Fatalf("%s=%#v err=%v, want %d", name, got, err, want)
		}
	}
	check("i8", count8, bytecode.TypeI8, 1, -1)
	check("i16", count16, bytecode.TypeI16, 2, -2)
	check("i32", count32, bytecode.TypeI32, 4, -3)
	check("long", countLong, bytecode.TypeI64, 8, -4)
	check("ll", countLL, bytecode.TypeI64, 8, -5)
	check("j", countJ, bytecode.TypeI64, 8, -6)
	check("z", countZ, bytecode.TypeI64, 8, -7)
	check("t", countT, bytecode.TypeI64, 8, -8)
}

func TestSscanfExternScansUnsignedIntegerLengthModifiers(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	fn, ok := reg.Lookup("sscanf")
	if !ok {
		t.Fatal("missing sscanf extern")
	}

	input := "255 65535 4294967295 4 5 6 7 8 ff 2345 1 1234ABCDEF 377 177777 10"
	inputAddr := mustAllocBytes(t, mem, "sscanf:unsigned-length-input", append([]byte(input), 0), true, blockString)
	fmtAddr := mustAllocBytes(t, mem, "sscanf:unsigned-length-fmt", []byte("%hhu %hu %u %lu %llu %ju %zu %tu %hhx %hx %x %llX %hho %ho %o\x00"), true, blockString)
	u8 := mustAlloc(t, mem, "sscanf:unsigned-length-u8", 1, 1, false, blockLocal)
	u16 := mustAlloc(t, mem, "sscanf:unsigned-length-u16", 2, 2, false, blockLocal)
	u32 := mustAlloc(t, mem, "sscanf:unsigned-length-u32", 4, 4, false, blockLocal)
	ul := mustAlloc(t, mem, "sscanf:unsigned-length-ul", 8, 8, false, blockLocal)
	ull := mustAlloc(t, mem, "sscanf:unsigned-length-ull", 8, 8, false, blockLocal)
	uj := mustAlloc(t, mem, "sscanf:unsigned-length-j", 8, 8, false, blockLocal)
	uz := mustAlloc(t, mem, "sscanf:unsigned-length-z", 8, 8, false, blockLocal)
	ut := mustAlloc(t, mem, "sscanf:unsigned-length-t", 8, 8, false, blockLocal)
	hhX := mustAlloc(t, mem, "sscanf:unsigned-length-hhx", 1, 1, false, blockLocal)
	hX := mustAlloc(t, mem, "sscanf:unsigned-length-hx", 2, 2, false, blockLocal)
	x := mustAlloc(t, mem, "sscanf:unsigned-length-x", 4, 4, false, blockLocal)
	llX := mustAlloc(t, mem, "sscanf:unsigned-length-llx", 8, 8, false, blockLocal)
	hhO := mustAlloc(t, mem, "sscanf:unsigned-length-hho", 1, 1, false, blockLocal)
	hO := mustAlloc(t, mem, "sscanf:unsigned-length-ho", 2, 2, false, blockLocal)
	o := mustAlloc(t, mem, "sscanf:unsigned-length-o", 4, 4, false, blockLocal)

	ret, exit, err := fn(context.Background(), &ExternContext{Memory: mem}, []Value{
		ObjectAddrValue(inputAddr),
		ObjectAddrValue(fmtAddr),
		PtrValue(u8),
		PtrValue(u16),
		PtrValue(u32),
		PtrValue(ul),
		PtrValue(ull),
		PtrValue(uj),
		PtrValue(uz),
		PtrValue(ut),
		PtrValue(hhX),
		PtrValue(hX),
		PtrValue(x),
		PtrValue(llX),
		PtrValue(hhO),
		PtrValue(hO),
		PtrValue(o),
	})
	if err != nil || exit != nil {
		t.Fatalf("sscanf ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeI32 || ret.Int != 15 {
		t.Fatalf("sscanf ret=%#v, want i32 15", ret)
	}
	check := func(name string, addr uint64, typ bytecode.ValueType, align int64, want uint64) {
		t.Helper()
		got, err := mem.Load(addr, typ, align)
		if err != nil || unsignedInt(got) != want {
			t.Fatalf("%s=%#v err=%v, want %d", name, got, err, want)
		}
	}
	check("u8", u8, bytecode.TypeU8, 1, 255)
	check("u16", u16, bytecode.TypeU16, 2, 65535)
	check("u32", u32, bytecode.TypeU32, 4, 4294967295)
	check("ul", ul, bytecode.TypeU64, 8, 4)
	check("ull", ull, bytecode.TypeU64, 8, 5)
	check("uj", uj, bytecode.TypeU64, 8, 6)
	check("uz", uz, bytecode.TypeU64, 8, 7)
	check("ut", ut, bytecode.TypeU64, 8, 8)
	check("hhx", hhX, bytecode.TypeU8, 1, 0xff)
	check("hx", hX, bytecode.TypeU16, 2, 0x2345)
	check("x", x, bytecode.TypeU32, 4, 1)
	check("llx", llX, bytecode.TypeU64, 8, 0x1234abcdef)
	check("hho", hhO, bytecode.TypeU8, 1, 0377)
	check("ho", hO, bytecode.TypeU16, 2, 0177777)
	check("o", o, bytecode.TypeU32, 4, 010)
}

func TestSscanfExternScansCountLengthModifiers(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	fn, ok := reg.Lookup("sscanf")
	if !ok {
		t.Fatal("missing sscanf extern")
	}

	inputAddr := mustAllocBytes(t, mem, "sscanf:count-length-input", []byte("abcdefgh9\x00"), true, blockString)
	fmtAddr := mustAllocBytes(t, mem, "sscanf:count-length-fmt", []byte("a%hhnb%hnc%nd%lne%llnf%jng%znh%tn%d\x00"), true, blockString)
	count8 := mustAlloc(t, mem, "sscanf:count-length-i8", 1, 1, false, blockLocal)
	count16 := mustAlloc(t, mem, "sscanf:count-length-i16", 2, 2, false, blockLocal)
	count32 := mustAlloc(t, mem, "sscanf:count-length-i32", 4, 4, false, blockLocal)
	countLong := mustAlloc(t, mem, "sscanf:count-length-long", 8, 8, false, blockLocal)
	countLL := mustAlloc(t, mem, "sscanf:count-length-ll", 8, 8, false, blockLocal)
	countJ := mustAlloc(t, mem, "sscanf:count-length-j", 8, 8, false, blockLocal)
	countZ := mustAlloc(t, mem, "sscanf:count-length-z", 8, 8, false, blockLocal)
	countT := mustAlloc(t, mem, "sscanf:count-length-t", 8, 8, false, blockLocal)
	value := mustAlloc(t, mem, "sscanf:count-length-value", 4, 4, false, blockLocal)

	ret, exit, err := fn(context.Background(), &ExternContext{Memory: mem}, []Value{
		ObjectAddrValue(inputAddr),
		ObjectAddrValue(fmtAddr),
		PtrValue(count8),
		PtrValue(count16),
		PtrValue(count32),
		PtrValue(countLong),
		PtrValue(countLL),
		PtrValue(countJ),
		PtrValue(countZ),
		PtrValue(countT),
		PtrValue(value),
	})
	if err != nil || exit != nil {
		t.Fatalf("sscanf ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeI32 || ret.Int != 1 {
		t.Fatalf("sscanf ret=%#v, want i32 1", ret)
	}
	check := func(name string, addr uint64, typ bytecode.ValueType, align int64, want int64) {
		t.Helper()
		got, err := mem.Load(addr, typ, align)
		if err != nil || signedInt(got) != want {
			t.Fatalf("%s=%#v err=%v, want %d", name, got, err, want)
		}
	}
	check("i8", count8, bytecode.TypeI8, 1, 1)
	check("i16", count16, bytecode.TypeI16, 2, 2)
	check("i32", count32, bytecode.TypeI32, 4, 3)
	check("long", countLong, bytecode.TypeI64, 8, 4)
	check("ll", countLL, bytecode.TypeI64, 8, 5)
	check("j", countJ, bytecode.TypeI64, 8, 6)
	check("z", countZ, bytecode.TypeI64, 8, 7)
	check("t", countT, bytecode.TypeI64, 8, 8)
	check("value", value, bytecode.TypeI32, 4, 9)
}

func TestSscanfExternScansCharacterLengthModifiers(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	fn, ok := reg.Lookup("sscanf")
	if !ok {
		t.Fatal("missing sscanf extern")
	}

	inputAddr := mustAllocBytes(t, mem, "sscanf:char-length-input", []byte("ABC\x00"), true, blockString)
	fmtAddr := mustAllocBytes(t, mem, "sscanf:char-length-fmt", []byte("%c%2lc\x00"), true, blockString)
	chAddr := mustAllocBytes(t, mem, "sscanf:char-length-char", []byte{0}, false, blockLocal)
	wideAddr := mustAlloc(t, mem, "sscanf:char-length-wide", 2*4, 4, false, blockLocal)

	ret, exit, err := fn(context.Background(), &ExternContext{Memory: mem}, []Value{
		ObjectAddrValue(inputAddr),
		ObjectAddrValue(fmtAddr),
		PtrValue(chAddr),
		PtrValue(wideAddr),
	})
	if err != nil || exit != nil {
		t.Fatalf("sscanf ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeI32 || ret.Int != 2 {
		t.Fatalf("sscanf ret=%#v, want i32 2", ret)
	}
	ch, err := mem.Load(chAddr, bytecode.TypeI8, 1)
	if err != nil || byte(ch.Int) != 'A' {
		t.Fatalf("char=%#v err=%v, want A", ch, err)
	}
	first, err := mem.Load(wideAddr, bytecode.TypeI32, 4)
	if err != nil || first.Int != 'B' {
		t.Fatalf("wide[0]=%#v err=%v, want B", first, err)
	}
	second, err := mem.Load(wideAddr+4, bytecode.TypeI32, 4)
	if err != nil || second.Int != 'C' {
		t.Fatalf("wide[1]=%#v err=%v, want C", second, err)
	}
}

func TestSscanfExternScansStringLengthModifiers(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	fn, ok := reg.Lookup("sscanf")
	if !ok {
		t.Fatal("missing sscanf extern")
	}

	inputAddr := mustAllocBytes(t, mem, "sscanf:string-length-input", []byte("abc ghijkl\x00"), true, blockString)
	fmtAddr := mustAllocBytes(t, mem, "sscanf:string-length-fmt", []byte("%3s %4ls\x00"), true, blockString)
	narrowAddr := mustAllocBytes(t, mem, "sscanf:string-length-narrow", []byte{0, 0, 0, 0}, false, blockLocal)
	wideAddr := mustAlloc(t, mem, "sscanf:string-length-wide", 5*4, 4, false, blockLocal)

	ret, exit, err := fn(context.Background(), &ExternContext{Memory: mem}, []Value{
		ObjectAddrValue(inputAddr),
		ObjectAddrValue(fmtAddr),
		PtrValue(narrowAddr),
		PtrValue(wideAddr),
	})
	if err != nil || exit != nil {
		t.Fatalf("sscanf ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeI32 || ret.Int != 2 {
		t.Fatalf("sscanf ret=%#v, want i32 2", ret)
	}
	narrow, err := mem.ReadCString(narrowAddr)
	if err != nil || narrow != "abc" {
		t.Fatalf("narrow=%q err=%v, want abc", narrow, err)
	}
	chars, err := readWideCString(mem, wideAddr)
	if err != nil {
		t.Fatalf("read wide string: %v", err)
	}
	if len(chars) != 4 || chars[0] != 'g' || chars[1] != 'h' || chars[2] != 'i' || chars[3] != 'j' {
		t.Fatalf("wide chars=%#v, want ghij", chars)
	}
}

func TestSscanfExternScansScansetLengthModifiers(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	fn, ok := reg.Lookup("sscanf")
	if !ok {
		t.Fatal("missing sscanf extern")
	}

	inputAddr := mustAllocBytes(t, mem, "sscanf:scanset-length-input", []byte("abc123Z\x00"), true, blockString)
	fmtAddr := mustAllocBytes(t, mem, "sscanf:scanset-length-fmt", []byte("%3l[a-z]%3[^Z]\x00"), true, blockString)
	wideAddr := mustAlloc(t, mem, "sscanf:scanset-length-wide", 4*4, 4, false, blockLocal)
	narrowAddr := mustAllocBytes(t, mem, "sscanf:scanset-length-narrow", []byte{0, 0, 0, 0}, false, blockLocal)

	ret, exit, err := fn(context.Background(), &ExternContext{Memory: mem}, []Value{
		ObjectAddrValue(inputAddr),
		ObjectAddrValue(fmtAddr),
		PtrValue(wideAddr),
		PtrValue(narrowAddr),
	})
	if err != nil || exit != nil {
		t.Fatalf("sscanf ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeI32 || ret.Int != 2 {
		t.Fatalf("sscanf ret=%#v, want i32 2", ret)
	}
	chars, err := readWideCString(mem, wideAddr)
	if err != nil {
		t.Fatalf("read wide string: %v", err)
	}
	if len(chars) != 3 || chars[0] != 'a' || chars[1] != 'b' || chars[2] != 'c' {
		t.Fatalf("wide chars=%#v, want abc", chars)
	}
	narrow, err := mem.ReadCString(narrowAddr)
	if err != nil || narrow != "123" {
		t.Fatalf("narrow=%q err=%v, want 123", narrow, err)
	}
}

func TestSscanfExternScansLiteralPercentAndWhitespace(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	fn, ok := reg.Lookup("sscanf")
	if !ok {
		t.Fatal("missing sscanf extern")
	}
	inputAddr := mustAllocBytes(t, mem, "sscanf:literal-input", []byte("  % 42\x00"), true, blockString)
	fmtAddr := mustAllocBytes(t, mem, "sscanf:literal-fmt", []byte(" %% %d\x00"), true, blockString)
	valueAddr := mustAlloc(t, mem, "sscanf:literal-value", 4, 4, false, blockLocal)
	ret, exit, err := fn(context.Background(), &ExternContext{Memory: mem}, []Value{
		ObjectAddrValue(inputAddr),
		ObjectAddrValue(fmtAddr),
		PtrValue(valueAddr),
	})
	if err != nil || exit != nil {
		t.Fatalf("sscanf ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeI32 || ret.Int != 1 {
		t.Fatalf("sscanf ret=%#v, want i32 1", ret)
	}
	value, err := mem.Load(valueAddr, bytecode.TypeI32, 4)
	if err != nil || value.Int != 42 {
		t.Fatalf("value=%#v err=%v, want 42", value, err)
	}
}

func TestScanfExternPreservesUnreadLiteralMismatch(t *testing.T) {
	reg := DefaultExternRegistryWithIO(strings.NewReader("abX9"), nil, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	fn, ok := reg.Lookup("scanf")
	if !ok {
		t.Fatal("missing scanf extern")
	}
	fmtAddr := mustAllocBytes(t, mem, "scanf:literal-mismatch-fmt", []byte("abY%d\x00"), true, blockString)
	valueAddr := mustAlloc(t, mem, "scanf:literal-mismatch-value", 4, 4, false, blockLocal)
	ret, exit, err := fn(context.Background(), &ExternContext{Memory: mem}, []Value{
		ObjectAddrValue(fmtAddr),
		PtrValue(valueAddr),
	})
	if err != nil || exit != nil {
		t.Fatalf("scanf ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeI32 || ret.Int != 0 {
		t.Fatalf("scanf ret=%#v, want i32 0", ret)
	}
	ch, ok := reg.readHostChar(reg.stdinHandle)
	if !ok || ch != 'X' {
		t.Fatalf("next stdin char=%q ok=%v, want X", ch, ok)
	}
}

func TestScanfExternScansStdinAndPreservesUnreadInput(t *testing.T) {
	reg := DefaultExternRegistryWithIO(strings.NewReader("17 tail"), nil, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	fn, ok := reg.Lookup("scanf")
	if !ok {
		t.Fatal("missing scanf extern")
	}
	fmtAddr := mustAllocBytes(t, mem, "scanf:fmt", []byte("%d\x00"), true, blockString)
	valueAddr := mustAlloc(t, mem, "scanf:value", 4, 4, false, blockLocal)

	ret, exit, err := fn(context.Background(), &ExternContext{Memory: mem}, []Value{
		ObjectAddrValue(fmtAddr),
		PtrValue(valueAddr),
	})
	if err != nil || exit != nil {
		t.Fatalf("scanf ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeI32 || ret.Int != 1 {
		t.Fatalf("scanf ret=%#v, want i32 1", ret)
	}
	value, err := mem.Load(valueAddr, bytecode.TypeI32, 4)
	if err != nil || value.Int != 17 {
		t.Fatalf("value=%#v err=%v, want 17", value, err)
	}
	ch, ok := reg.readHostChar(reg.stdinHandle)
	if !ok || ch != ' ' {
		t.Fatalf("next stdin char=%q ok=%v, want space", ch, ok)
	}
}

func TestScanfExternScansSetAndPreservesUnreadInput(t *testing.T) {
	reg := DefaultExternRegistryWithIO(strings.NewReader("abc!"), nil, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	fn, ok := reg.Lookup("scanf")
	if !ok {
		t.Fatal("missing scanf extern")
	}
	fmtAddr := mustAllocBytes(t, mem, "scanf:scanset-fmt", []byte("%[a-z]\x00"), true, blockString)
	wordAddr := mustAllocBytes(t, mem, "scanf:scanset-word", []byte{0, 0, 0, 0}, false, blockLocal)

	ret, exit, err := fn(context.Background(), &ExternContext{Memory: mem}, []Value{
		ObjectAddrValue(fmtAddr),
		PtrValue(wordAddr),
	})
	if err != nil || exit != nil {
		t.Fatalf("scanf ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeI32 || ret.Int != 1 {
		t.Fatalf("scanf ret=%#v, want i32 1", ret)
	}
	word, err := mem.ReadCString(wordAddr)
	if err != nil || word != "abc" {
		t.Fatalf("word=%q err=%v, want abc", word, err)
	}
	ch, ok := reg.readHostChar(reg.stdinHandle)
	if !ok || ch != '!' {
		t.Fatalf("next stdin char=%q ok=%v, want !", ch, ok)
	}
}

func TestFscanfExternScansConfiguredFileAndPreservesUnreadInput(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	reg.AddFile("data.txt", []byte("23 rest"))
	mem := NewMemory(bytecode.DefaultTarget())
	path := mustAllocBytes(t, mem, "fscanf:path", []byte("data.txt\x00"), true, blockString)
	mode := mustAllocBytes(t, mem, "fscanf:mode", []byte("r\x00"), true, blockString)
	fmtAddr := mustAllocBytes(t, mem, "fscanf:fmt", []byte("%d\x00"), true, blockString)
	valueAddr := mustAlloc(t, mem, "fscanf:value", 4, 4, false, blockLocal)

	fopenFn, ok := reg.Lookup("fopen")
	if !ok {
		t.Fatal("missing fopen extern")
	}
	ret, exit, err := fopenFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(path), PtrValue(mode)})
	if err != nil || exit != nil || ret.Type != bytecode.TypePtr || ret.Int == 0 {
		t.Fatalf("fopen ret=%#v exit=%#v err=%v, want file handle", ret, exit, err)
	}
	file := ret.Int
	fn, ok := reg.Lookup("fscanf")
	if !ok {
		t.Fatal("missing fscanf extern")
	}
	ret, exit, err = fn(context.Background(), &ExternContext{Memory: mem}, []Value{
		PtrValue(file),
		ObjectAddrValue(fmtAddr),
		PtrValue(valueAddr),
	})
	if err != nil || exit != nil {
		t.Fatalf("fscanf ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeI32 || ret.Int != 1 {
		t.Fatalf("fscanf ret=%#v, want i32 1", ret)
	}
	value, err := mem.Load(valueAddr, bytecode.TypeI32, 4)
	if err != nil || value.Int != 23 {
		t.Fatalf("value=%#v err=%v, want 23", value, err)
	}
	ch, ok := reg.readHostChar(file)
	if !ok || ch != ' ' {
		t.Fatalf("next file char=%q ok=%v, want space", ch, ok)
	}
}

func TestFscanfExternScansFloatAndPreservesUnreadInput(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	reg.AddFile("float.txt", []byte("1.5!"))
	mem := NewMemory(bytecode.DefaultTarget())
	path := mustAllocBytes(t, mem, "fscanf:float-path", []byte("float.txt\x00"), true, blockString)
	mode := mustAllocBytes(t, mem, "fscanf:float-mode", []byte("r\x00"), true, blockString)
	fmtAddr := mustAllocBytes(t, mem, "fscanf:float-fmt", []byte("%f\x00"), true, blockString)
	valueAddr := mustAlloc(t, mem, "fscanf:float-value", 4, 4, false, blockLocal)

	fopenFn, ok := reg.Lookup("fopen")
	if !ok {
		t.Fatal("missing fopen extern")
	}
	ret, exit, err := fopenFn(context.Background(), &ExternContext{Memory: mem}, []Value{PtrValue(path), PtrValue(mode)})
	if err != nil || exit != nil || ret.Type != bytecode.TypePtr || ret.Int == 0 {
		t.Fatalf("fopen ret=%#v exit=%#v err=%v, want file handle", ret, exit, err)
	}
	file := ret.Int
	fn, ok := reg.Lookup("fscanf")
	if !ok {
		t.Fatal("missing fscanf extern")
	}
	ret, exit, err = fn(context.Background(), &ExternContext{Memory: mem}, []Value{
		PtrValue(file),
		ObjectAddrValue(fmtAddr),
		PtrValue(valueAddr),
	})
	if err != nil || exit != nil {
		t.Fatalf("fscanf ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeI32 || ret.Int != 1 {
		t.Fatalf("fscanf ret=%#v, want i32 1", ret)
	}
	value, err := mem.Load(valueAddr, bytecode.TypeF32, 4)
	if err != nil || value.Float != 1.5 {
		t.Fatalf("value=%#v err=%v, want 1.5", value, err)
	}
	ch, ok := reg.readHostChar(file)
	if !ok || ch != '!' {
		t.Fatalf("next file char=%q ok=%v, want !", ch, ok)
	}
}

func TestSscanfExternScansRadixSuppressionAndCount(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	fn, ok := reg.Lookup("sscanf")
	if !ok {
		t.Fatal("missing sscanf extern")
	}
	inputAddr := mustAllocBytes(t, mem, "sscanf:radix-input", []byte("11 0x2a 10 rest\x00"), true, blockString)
	fmtAddr := mustAllocBytes(t, mem, "sscanf:radix-fmt", []byte("%*d %x %o%n\x00"), true, blockString)
	hexAddr := mustAlloc(t, mem, "sscanf:hex", 4, 4, false, blockLocal)
	octalAddr := mustAlloc(t, mem, "sscanf:octal", 4, 4, false, blockLocal)
	countAddr := mustAlloc(t, mem, "sscanf:count", 4, 4, false, blockLocal)

	ret, exit, err := fn(context.Background(), &ExternContext{Memory: mem}, []Value{
		ObjectAddrValue(inputAddr),
		ObjectAddrValue(fmtAddr),
		PtrValue(hexAddr),
		PtrValue(octalAddr),
		PtrValue(countAddr),
	})
	if err != nil || exit != nil {
		t.Fatalf("sscanf ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeI32 || ret.Int != 2 {
		t.Fatalf("sscanf ret=%#v, want i32 2", ret)
	}
	hex, err := mem.Load(hexAddr, bytecode.TypeU32, 4)
	if err != nil || hex.Int != 42 {
		t.Fatalf("hex=%#v err=%v, want 42", hex, err)
	}
	octal, err := mem.Load(octalAddr, bytecode.TypeU32, 4)
	if err != nil || octal.Int != 8 {
		t.Fatalf("octal=%#v err=%v, want 8", octal, err)
	}
	count, err := mem.Load(countAddr, bytecode.TypeI32, 4)
	if err != nil || count.Int != 10 {
		t.Fatalf("count=%#v err=%v, want 10", count, err)
	}
}

func TestSscanfExternScansSets(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	fn, ok := reg.Lookup("sscanf")
	if !ok {
		t.Fatal("missing sscanf extern")
	}
	inputAddr := mustAllocBytes(t, mem, "sscanf:scanset-input", []byte("abc123 xyz!\x00"), true, blockString)
	fmtAddr := mustAllocBytes(t, mem, "sscanf:scanset-fmt", []byte("%[a-z]%3[0-9] %*[^!]%c\x00"), true, blockString)
	wordAddr := mustAllocBytes(t, mem, "sscanf:scanset-word", []byte{0, 0, 0, 0}, false, blockLocal)
	digitsAddr := mustAllocBytes(t, mem, "sscanf:scanset-digits", []byte{0, 0, 0, 0}, false, blockLocal)
	chAddr := mustAllocBytes(t, mem, "sscanf:scanset-char", []byte{0}, false, blockLocal)

	ret, exit, err := fn(context.Background(), &ExternContext{Memory: mem}, []Value{
		ObjectAddrValue(inputAddr),
		ObjectAddrValue(fmtAddr),
		PtrValue(wordAddr),
		PtrValue(digitsAddr),
		PtrValue(chAddr),
	})
	if err != nil || exit != nil {
		t.Fatalf("sscanf ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeI32 || ret.Int != 3 {
		t.Fatalf("sscanf ret=%#v, want i32 3", ret)
	}
	word, err := mem.ReadCString(wordAddr)
	if err != nil || word != "abc" {
		t.Fatalf("word=%q err=%v, want abc", word, err)
	}
	digits, err := mem.ReadCString(digitsAddr)
	if err != nil || digits != "123" {
		t.Fatalf("digits=%q err=%v, want 123", digits, err)
	}
	ch, err := mem.Load(chAddr, bytecode.TypeI8, 1)
	if err != nil || byte(ch.Int) != '!' {
		t.Fatalf("char=%#v err=%v, want !", ch, err)
	}
}

func TestSscanfExternScansFloatingValues(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	fn, ok := reg.Lookup("sscanf")
	if !ok {
		t.Fatal("missing sscanf extern")
	}
	inputAddr := mustAllocBytes(t, mem, "sscanf:float-input", []byte("1.5 -2.25 0x1.8p+2 rest\x00"), true, blockString)
	fmtAddr := mustAllocBytes(t, mem, "sscanf:float-fmt", []byte("%f %lf %La\x00"), true, blockString)
	floatAddr := mustAlloc(t, mem, "sscanf:float-f32", 4, 4, false, blockLocal)
	doubleAddr := mustAlloc(t, mem, "sscanf:float-f64", 8, 8, false, blockLocal)
	longAddr := mustAlloc(t, mem, "sscanf:float-flong", 16, 8, false, blockLocal)

	ret, exit, err := fn(context.Background(), &ExternContext{Memory: mem}, []Value{
		ObjectAddrValue(inputAddr),
		ObjectAddrValue(fmtAddr),
		PtrValue(floatAddr),
		PtrValue(doubleAddr),
		PtrValue(longAddr),
	})
	if err != nil || exit != nil {
		t.Fatalf("sscanf ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeI32 || ret.Int != 3 {
		t.Fatalf("sscanf ret=%#v, want i32 3", ret)
	}
	f, err := mem.Load(floatAddr, bytecode.TypeF32, 4)
	if err != nil || f.Float != 1.5 {
		t.Fatalf("float=%#v err=%v, want 1.5", f, err)
	}
	d, err := mem.Load(doubleAddr, bytecode.TypeF64, 8)
	if err != nil || d.Float != -2.25 {
		t.Fatalf("double=%#v err=%v, want -2.25", d, err)
	}
	l, err := mem.Load(longAddr, bytecode.TypeFLong, 8)
	if err != nil || l.Float != 6 {
		t.Fatalf("long double=%#v err=%v, want 6", l, err)
	}
}

func TestSscanfExternScansPointerValues(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	target := bytecode.DefaultTarget()
	mem := NewMemory(target)
	fn, ok := reg.Lookup("sscanf")
	if !ok {
		t.Fatal("missing sscanf extern")
	}
	inputAddr := mustAllocBytes(t, mem, "sscanf:pointer-input", []byte("0x1234 0x5 tail\x00"), true, blockString)
	fmtAddr := mustAllocBytes(t, mem, "sscanf:pointer-fmt", []byte("%p %3p\x00"), true, blockString)
	ptrAddr := mustAlloc(t, mem, "sscanf:pointer", target.PointerSize, target.PointerAlign, false, blockLocal)
	narrowAddr := mustAlloc(t, mem, "sscanf:pointer-narrow", target.PointerSize, target.PointerAlign, false, blockLocal)

	ret, exit, err := fn(context.Background(), &ExternContext{Memory: mem}, []Value{
		ObjectAddrValue(inputAddr),
		ObjectAddrValue(fmtAddr),
		PtrValue(ptrAddr),
		PtrValue(narrowAddr),
	})
	if err != nil || exit != nil {
		t.Fatalf("sscanf ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeI32 || ret.Int != 2 {
		t.Fatalf("sscanf ret=%#v, want i32 2", ret)
	}
	ptr, err := mem.Load(ptrAddr, bytecode.TypePtr, target.PointerAlign)
	if err != nil || ptr.Int != 0x1234 {
		t.Fatalf("ptr=%#v err=%v, want 0x1234", ptr, err)
	}
	narrow, err := mem.Load(narrowAddr, bytecode.TypePtr, target.PointerAlign)
	if err != nil || narrow.Int != 0x5 {
		t.Fatalf("narrow ptr=%#v err=%v, want 0x5", narrow, err)
	}
}

func TestSscanfExternDistinguishesInputFailure(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	fn, ok := reg.Lookup("sscanf")
	if !ok {
		t.Fatal("missing sscanf extern")
	}
	fmtAddr := mustAllocBytes(t, mem, "sscanf:failure-fmt", []byte("%d\x00"), true, blockString)
	valueAddr := mustAlloc(t, mem, "sscanf:failure-value", 4, 4, false, blockLocal)
	tests := []struct {
		name  string
		input string
		want  int64
	}{
		{name: "empty input", input: "", want: -1},
		{name: "whitespace only", input: " \t\n", want: -1},
		{name: "matching failure", input: "word", want: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inputAddr := mustAllocBytes(t, mem, "sscanf:failure-input:"+tt.name, append([]byte(tt.input), 0), true, blockString)
			ret, exit, err := fn(context.Background(), &ExternContext{Memory: mem}, []Value{
				ObjectAddrValue(inputAddr),
				ObjectAddrValue(fmtAddr),
				PtrValue(valueAddr),
			})
			if err != nil || exit != nil {
				t.Fatalf("sscanf ret=%#v exit=%#v err=%v", ret, exit, err)
			}
			if ret.Type != bytecode.TypeI32 || int64(int32(ret.Int)) != tt.want {
				t.Fatalf("sscanf ret=%#v, want i32 %d", ret, tt.want)
			}
		})
	}
}

func TestSscanfExternReturnsAssignmentsBeforeLaterFailure(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	fn, ok := reg.Lookup("sscanf")
	if !ok {
		t.Fatal("missing sscanf extern")
	}
	inputAddr := mustAllocBytes(t, mem, "sscanf:later-failure-input", []byte("12 word\x00"), true, blockString)
	fmtAddr := mustAllocBytes(t, mem, "sscanf:later-failure-fmt", []byte("%d %d\x00"), true, blockString)
	firstAddr := mustAlloc(t, mem, "sscanf:later-failure-first", 4, 4, false, blockLocal)
	secondAddr := mustAlloc(t, mem, "sscanf:later-failure-second", 4, 4, false, blockLocal)
	ret, exit, err := fn(context.Background(), &ExternContext{Memory: mem}, []Value{
		ObjectAddrValue(inputAddr),
		ObjectAddrValue(fmtAddr),
		PtrValue(firstAddr),
		PtrValue(secondAddr),
	})
	if err != nil || exit != nil {
		t.Fatalf("sscanf ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeI32 || ret.Int != 1 {
		t.Fatalf("sscanf ret=%#v, want i32 1", ret)
	}
	first, err := mem.Load(firstAddr, bytecode.TypeI32, 4)
	if err != nil || first.Int != 12 {
		t.Fatalf("first=%#v err=%v, want 12", first, err)
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
	out.Reset()
	fn, ok = reg.Lookup("vprintf_unlocked")
	if !ok {
		t.Fatal("missing vprintf_unlocked extern")
	}
	ret, exit, callErr = fn(context.Background(), &ExternContext{Memory: mem, Stdout: &out, Stderr: &errOut}, []Value{
		ObjectAddrValue(fmtAddr),
		PtrValue(0),
	})
	if callErr != nil || exit != nil {
		t.Fatalf("vprintf_unlocked ret=%#v exit=%#v err=%v", ret, exit, callErr)
	}
	if ret.Type != bytecode.TypeI32 || ret.Int != 2 || out.String() != "v%" {
		t.Fatalf("vprintf_unlocked ret=%#v output=%q, want i32 2 and v%%", ret, out.String())
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
	errOut.Reset()
	fn, ok = reg.Lookup("vfprintf_unlocked")
	if !ok {
		t.Fatal("missing vfprintf_unlocked extern")
	}
	ret, exit, callErr = fn(context.Background(), &ExternContext{Memory: mem, Stdout: &out, Stderr: &errOut}, []Value{
		ObjectAddrValue(stderrAddr),
		ObjectAddrValue(fmtErrAddr),
		PtrValue(0),
	})
	if callErr != nil || exit != nil {
		t.Fatalf("vfprintf_unlocked ret=%#v exit=%#v err=%v", ret, exit, callErr)
	}
	if ret.Type != bytecode.TypeI32 || ret.Int != 3 || errOut.String() != "err" {
		t.Fatalf("vfprintf_unlocked ret=%#v stderr=%q, want i32 3 and err", ret, errOut.String())
	}
}

func TestVFormatExternsReadMemoryVaList(t *testing.T) {
	var out bytes.Buffer
	reg := DefaultExternRegistry(&out, nil)
	mem := NewMemory(bytecode.DefaultTarget())

	bufAddr := mustAlloc(t, mem, "buf:vformat-va-list", 32, 1, false, blockLocal)
	fmtAddr := mustAllocBytes(t, mem, "fmt:vformat-va-list", []byte("%d %s\x00"), true, blockString)
	strAddr := mustAllocBytes(t, mem, "str:vformat-va-list", []byte("ok\x00"), true, blockString)
	apAddr := mustAllocVaList(t, mem, "ap:vformat-va-list",
		IntValue(bytecode.TypeI32, 42),
		ObjectAddrValue(strAddr),
	)

	fn, ok := reg.Lookup("vsprintf")
	if !ok {
		t.Fatal("missing vsprintf extern")
	}
	ret, exit, callErr := fn(context.Background(), &ExternContext{Memory: mem}, []Value{
		ObjectAddrValue(bufAddr),
		ObjectAddrValue(fmtAddr),
		PtrValue(apAddr),
	})
	if callErr != nil || exit != nil {
		t.Fatalf("vsprintf ret=%#v exit=%#v err=%v", ret, exit, callErr)
	}
	got, err := mem.ReadCString(bufAddr)
	if err != nil {
		t.Fatalf("ReadCString: %v", err)
	}
	if ret.Type != bytecode.TypeI32 || ret.Int != 5 || got != "42 ok" {
		t.Fatalf("vsprintf ret=%#v output=%q, want i32 5 and 42 ok", ret, got)
	}

	fn, ok = reg.Lookup("vprintf")
	if !ok {
		t.Fatal("missing vprintf extern")
	}
	ret, exit, callErr = fn(context.Background(), &ExternContext{Memory: mem, Stdout: &out}, []Value{
		ObjectAddrValue(fmtAddr),
		PtrValue(apAddr),
	})
	if callErr != nil || exit != nil {
		t.Fatalf("vprintf ret=%#v exit=%#v err=%v", ret, exit, callErr)
	}
	if ret.Type != bytecode.TypeI32 || ret.Int != 5 || out.String() != "42 ok" {
		t.Fatalf("vprintf ret=%#v output=%q, want i32 5 and 42 ok", ret, out.String())
	}
}

func TestVFormatExternsReadFloatingAndCountMemoryVaList(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	mem := NewMemory(bytecode.DefaultTarget())

	bufAddr := mustAlloc(t, mem, "buf:vformat-va-list-float", 32, 1, false, blockLocal)
	countAddr := mustAlloc(t, mem, "count:vformat-va-list-float", 4, 4, false, blockLocal)
	fmtAddr := mustAllocBytes(t, mem, "fmt:vformat-va-list-float", []byte("%+.1f%n\x00"), true, blockString)
	apAddr := mustAllocVaList(t, mem, "ap:vformat-va-list-float",
		FloatValue(bytecode.TypeF64, 1.5),
		ObjectAddrValue(countAddr),
	)

	fn, ok := reg.Lookup("vsprintf")
	if !ok {
		t.Fatal("missing vsprintf extern")
	}
	ret, exit, callErr := fn(context.Background(), &ExternContext{Memory: mem}, []Value{
		ObjectAddrValue(bufAddr),
		ObjectAddrValue(fmtAddr),
		PtrValue(apAddr),
	})
	if callErr != nil || exit != nil {
		t.Fatalf("vsprintf ret=%#v exit=%#v err=%v", ret, exit, callErr)
	}
	got, err := mem.ReadCString(bufAddr)
	if err != nil {
		t.Fatalf("ReadCString: %v", err)
	}
	count, err := mem.Load(countAddr, bytecode.TypeI32, 4)
	if err != nil {
		t.Fatalf("Load count: %v", err)
	}
	if ret.Type != bytecode.TypeI32 || ret.Int != 4 || got != "+1.5" || signedInt(count) != 4 {
		t.Fatalf("vsprintf ret=%#v output=%q count=%d, want i32 4, +1.5, count 4", ret, got, signedInt(count))
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

func TestCheckedVFormatExternsReadMemoryVaList(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer
	reg := DefaultExternRegistry(&out, &errOut)
	mem := NewMemory(bytecode.DefaultTarget())

	bufAddr := mustAlloc(t, mem, "buf:vformat-va-list-chk", 32, 1, false, blockLocal)
	fmtAddr := mustAllocBytes(t, mem, "fmt:vformat-va-list-chk", []byte("%*s %u\x00"), true, blockString)
	strAddr := mustAllocBytes(t, mem, "str:vformat-va-list-chk", []byte("ok\x00"), true, blockString)
	apAddr := mustAllocVaList(t, mem, "ap:vformat-va-list-chk",
		IntValue(bytecode.TypeI32, 4),
		ObjectAddrValue(strAddr),
		UIntValue(bytecode.TypeU32, 9),
	)

	fn, ok := reg.Lookup("__builtin___vsnprintf_chk")
	if !ok {
		t.Fatal("missing __builtin___vsnprintf_chk extern")
	}
	ret, exit, callErr := fn(context.Background(), &ExternContext{Memory: mem}, []Value{
		ObjectAddrValue(bufAddr),
		UIntValue(bytecode.TypeU64, 16),
		IntValue(bytecode.TypeI32, 0),
		UIntValue(bytecode.TypeU64, 32),
		ObjectAddrValue(fmtAddr),
		PtrValue(apAddr),
	})
	if callErr != nil || exit != nil {
		t.Fatalf("__builtin___vsnprintf_chk ret=%#v exit=%#v err=%v", ret, exit, callErr)
	}
	got, err := mem.ReadCString(bufAddr)
	if err != nil {
		t.Fatalf("ReadCString: %v", err)
	}
	if ret.Type != bytecode.TypeI32 || ret.Int != 6 || got != "  ok 9" {
		t.Fatalf("__builtin___vsnprintf_chk ret=%#v output=%q, want i32 6 and two-space ok 9", ret, got)
	}

	fn, ok = reg.Lookup("__builtin___vfprintf_chk")
	if !ok {
		t.Fatal("missing __builtin___vfprintf_chk extern")
	}
	stderrAddr, ok := reg.LookupVariable("stderr", mem)
	if !ok {
		t.Fatal("missing stderr extern variable")
	}
	ret, exit, callErr = fn(context.Background(), &ExternContext{Memory: mem, Stderr: &errOut}, []Value{
		ObjectAddrValue(stderrAddr),
		IntValue(bytecode.TypeI32, 0),
		ObjectAddrValue(fmtAddr),
		PtrValue(apAddr),
	})
	if callErr != nil || exit != nil {
		t.Fatalf("__builtin___vfprintf_chk ret=%#v exit=%#v err=%v", ret, exit, callErr)
	}
	if ret.Type != bytecode.TypeI32 || ret.Int != 6 || errOut.String() != "  ok 9" {
		t.Fatalf("__builtin___vfprintf_chk ret=%#v stderr=%q, want i32 6 and two-space ok 9", ret, errOut.String())
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

func TestFormatExternsHonorSignedIntegerLengthModifiers(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	bufAddr := mustAlloc(t, mem, "buf:format-signed-length", 64, 1, false, blockLocal)
	fmtAddr := mustAllocBytes(t, mem, "fmt:format-signed-length", []byte("%hhd %hd %d %ld %lld %jd %zd %td\x00"), true, blockString)
	fn, ok := reg.Lookup("__builtin_sprintf")
	if !ok {
		t.Fatal("missing __builtin_sprintf extern")
	}
	ret, exit, callErr := fn(context.Background(), &ExternContext{Memory: mem}, []Value{
		ObjectAddrValue(bufAddr),
		ObjectAddrValue(fmtAddr),
		IntValue(bytecode.TypeI32, 255),
		IntValue(bytecode.TypeI32, 65534),
		IntValue(bytecode.TypeI32, -3),
		IntValue(bytecode.TypeI64, -4),
		IntValue(bytecode.TypeI64, -5),
		IntValue(bytecode.TypeI64, -6),
		IntValue(bytecode.TypeI64, -7),
		IntValue(bytecode.TypeI64, -8),
	})
	if callErr != nil || exit != nil {
		t.Fatalf("__builtin_sprintf ret=%#v exit=%#v err=%v", ret, exit, callErr)
	}
	got, err := mem.ReadCString(bufAddr)
	if err != nil {
		t.Fatalf("ReadCString: %v", err)
	}
	if ret.Type != bytecode.TypeI32 || ret.Int != 23 || got != "-1 -2 -3 -4 -5 -6 -7 -8" {
		t.Fatalf("__builtin_sprintf ret=%#v output=%q, want i32 23 and signed length output", ret, got)
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

func TestFormatExternsHonorUnsignedIntegerLengthModifiers(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	bufAddr := mustAlloc(t, mem, "buf:format-unsigned-length", 160, 1, false, blockLocal)
	fmtAddr := mustAllocBytes(t, mem, "fmt:format-unsigned-length", []byte("%hhu %hu %u %lu %llu %ju %zu %tu %hhx %hx %x %llX %hho %ho %o\x00"), true, blockString)
	fn, ok := reg.Lookup("__builtin_sprintf")
	if !ok {
		t.Fatal("missing __builtin_sprintf extern")
	}
	ret, exit, callErr := fn(context.Background(), &ExternContext{Memory: mem}, []Value{
		ObjectAddrValue(bufAddr),
		ObjectAddrValue(fmtAddr),
		IntValue(bytecode.TypeI32, -1),
		IntValue(bytecode.TypeI32, -1),
		IntValue(bytecode.TypeI64, -1),
		UIntValue(bytecode.TypeU64, 4),
		UIntValue(bytecode.TypeU64, 5),
		UIntValue(bytecode.TypeU64, 6),
		UIntValue(bytecode.TypeU64, 7),
		UIntValue(bytecode.TypeU64, 8),
		UIntValue(bytecode.TypeU32, 0x12ff),
		UIntValue(bytecode.TypeU32, 0x12345),
		UIntValue(bytecode.TypeU64, 0x100000001),
		UIntValue(bytecode.TypeU64, 0x1234abcdef),
		UIntValue(bytecode.TypeU32, 0777),
		UIntValue(bytecode.TypeU32, 0x1ffff),
		UIntValue(bytecode.TypeU64, 0x100000008),
	})
	if callErr != nil || exit != nil {
		t.Fatalf("__builtin_sprintf ret=%#v exit=%#v err=%v", ret, exit, callErr)
	}
	got, err := mem.ReadCString(bufAddr)
	if err != nil {
		t.Fatalf("ReadCString: %v", err)
	}
	want := "255 65535 4294967295 4 5 6 7 8 ff 2345 1 1234ABCDEF 377 177777 10"
	if ret.Type != bytecode.TypeI32 || ret.Int != uint64(len(want)) || got != want {
		t.Fatalf("__builtin_sprintf ret=%#v output=%q, want i32 %d and unsigned length output", ret, got, len(want))
	}
}

func TestFormatExternsSupportCharacterLengthModifiers(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	bufAddr := mustAlloc(t, mem, "buf:format-char-length", 16, 1, false, blockLocal)
	fmtAddr := mustAllocBytes(t, mem, "fmt:format-char-length", []byte("%c %lc\x00"), true, blockString)
	fn, ok := reg.Lookup("__builtin_sprintf")
	if !ok {
		t.Fatal("missing __builtin_sprintf extern")
	}
	ret, exit, callErr := fn(context.Background(), &ExternContext{Memory: mem}, []Value{
		ObjectAddrValue(bufAddr),
		ObjectAddrValue(fmtAddr),
		IntValue(bytecode.TypeI32, 'A'),
		IntValue(bytecode.TypeI32, 'Z'),
	})
	if callErr != nil || exit != nil {
		t.Fatalf("__builtin_sprintf ret=%#v exit=%#v err=%v", ret, exit, callErr)
	}
	got, err := mem.ReadCString(bufAddr)
	if err != nil {
		t.Fatalf("ReadCString: %v", err)
	}
	if ret.Type != bytecode.TypeI32 || ret.Int != 3 || got != "A Z" {
		t.Fatalf("__builtin_sprintf ret=%#v output=%q, want i32 3 and A Z", ret, got)
	}
}

func TestFormatExternsRejectHighWideCharacter(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	bufAddr := mustAlloc(t, mem, "buf:format-high-wide-char", 16, 1, false, blockLocal)
	fmtAddr := mustAllocBytes(t, mem, "fmt:format-high-wide-char", []byte("%lc\x00"), true, blockString)
	fn, ok := reg.Lookup("__builtin_sprintf")
	if !ok {
		t.Fatal("missing __builtin_sprintf extern")
	}
	_, exit, callErr := fn(context.Background(), &ExternContext{Memory: mem}, []Value{
		ObjectAddrValue(bufAddr),
		ObjectAddrValue(fmtAddr),
		IntValue(bytecode.TypeI32, 0x80),
	})
	if callErr == nil || exit != nil {
		t.Fatalf("__builtin_sprintf exit=%#v err=%v, want high wide character error", exit, callErr)
	}
}

func TestFormatExternsSupportWideStringLengthModifier(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	bufAddr := mustAlloc(t, mem, "buf:format-wide-string", 32, 1, false, blockLocal)
	fmtAddr := mustAllocBytes(t, mem, "fmt:format-wide-string", []byte("%s %.3ls %ls\x00"), true, blockString)
	narrow := mustAllocBytes(t, mem, "str:format-wide-string-narrow", []byte("ok\x00"), true, blockString)
	wide, err := mem.TryAlloc("str:format-wide-string-wide", 7*4, 4, false, blockString)
	if err != nil {
		t.Fatalf("alloc wide string: %v", err)
	}
	for i, ch := range []uint32{'a', 'b', 'c', 'd', 'e', 'f', 0} {
		if err := storeWideChar(mem, wide+uint64(i*4), ch); err != nil {
			t.Fatalf("store wide[%d]: %v", i, err)
		}
	}
	fn, ok := reg.Lookup("__builtin_sprintf")
	if !ok {
		t.Fatal("missing __builtin_sprintf extern")
	}
	ret, exit, callErr := fn(context.Background(), &ExternContext{Memory: mem}, []Value{
		ObjectAddrValue(bufAddr),
		ObjectAddrValue(fmtAddr),
		PtrValue(narrow),
		PtrValue(wide),
		PtrValue(wide),
	})
	if callErr != nil || exit != nil {
		t.Fatalf("__builtin_sprintf ret=%#v exit=%#v err=%v", ret, exit, callErr)
	}
	got, err := mem.ReadCString(bufAddr)
	if err != nil {
		t.Fatalf("ReadCString: %v", err)
	}
	if ret.Type != bytecode.TypeI32 || ret.Int != 13 || got != "ok abc abcdef" {
		t.Fatalf("__builtin_sprintf ret=%#v output=%q, want i32 13 and ok abc abcdef", ret, got)
	}
}

func TestFormatExternsRejectHighWideString(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	bufAddr := mustAlloc(t, mem, "buf:format-high-wide-string", 16, 1, false, blockLocal)
	fmtAddr := mustAllocBytes(t, mem, "fmt:format-high-wide-string", []byte("%ls\x00"), true, blockString)
	wide, err := mem.TryAlloc("str:format-high-wide-string", 2*4, 4, false, blockString)
	if err != nil {
		t.Fatalf("alloc wide string: %v", err)
	}
	for i, ch := range []uint32{0x80, 0} {
		if err := storeWideChar(mem, wide+uint64(i*4), ch); err != nil {
			t.Fatalf("store wide[%d]: %v", i, err)
		}
	}
	fn, ok := reg.Lookup("__builtin_sprintf")
	if !ok {
		t.Fatal("missing __builtin_sprintf extern")
	}
	_, exit, callErr := fn(context.Background(), &ExternContext{Memory: mem}, []Value{
		ObjectAddrValue(bufAddr),
		ObjectAddrValue(fmtAddr),
		PtrValue(wide),
	})
	if callErr == nil || exit != nil {
		t.Fatalf("__builtin_sprintf exit=%#v err=%v, want high wide string error", exit, callErr)
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

func TestFormatExternsSupportPointerWidth(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	bufAddr := mustAlloc(t, mem, "buf:format-pointer-width", 48, 1, false, blockLocal)
	fmtAddr := mustAllocBytes(t, mem, "fmt:format-pointer-width", []byte("%8p|%-8p|%6p\x00"), true, blockString)
	fn, ok := reg.Lookup("__builtin_sprintf")
	if !ok {
		t.Fatal("missing __builtin_sprintf extern")
	}
	ret, exit, callErr := fn(context.Background(), &ExternContext{Memory: mem}, []Value{
		ObjectAddrValue(bufAddr),
		ObjectAddrValue(fmtAddr),
		PtrValue(0x1234),
		PtrValue(0x1234),
		PtrValue(0),
	})
	if callErr != nil || exit != nil {
		t.Fatalf("__builtin_sprintf ret=%#v exit=%#v err=%v", ret, exit, callErr)
	}
	got, err := mem.ReadCString(bufAddr)
	if err != nil {
		t.Fatalf("ReadCString: %v", err)
	}
	want := "  0x1234|0x1234  |   0x0"
	if ret.Type != bytecode.TypeI32 || ret.Int != uint64(len(want)) || got != want {
		t.Fatalf("__builtin_sprintf ret=%#v output=%q, want i32 %d and pointer width output", ret, got, len(want))
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

func TestFormatExternsSupportZeroIntegerPrecision(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	bufAddr := mustAlloc(t, mem, "buf:format-zero-precision", 16, 1, false, blockLocal)
	fmtAddr := mustAllocBytes(t, mem, "fmt:format-zero-precision", []byte("%.0d|%.0u|%.0x|%.0o|%#.0o|%#.0x|%#.0X\x00"), true, blockString)
	fn, ok := reg.Lookup("__builtin_sprintf")
	if !ok {
		t.Fatal("missing __builtin_sprintf extern")
	}
	ret, exit, callErr := fn(context.Background(), &ExternContext{Memory: mem}, []Value{
		ObjectAddrValue(bufAddr),
		ObjectAddrValue(fmtAddr),
		IntValue(bytecode.TypeI32, 0),
		UIntValue(bytecode.TypeU32, 0),
		UIntValue(bytecode.TypeU32, 0),
		UIntValue(bytecode.TypeU32, 0),
		UIntValue(bytecode.TypeU32, 0),
		UIntValue(bytecode.TypeU32, 0),
		UIntValue(bytecode.TypeU32, 0),
	})
	if callErr != nil || exit != nil {
		t.Fatalf("__builtin_sprintf ret=%#v exit=%#v err=%v", ret, exit, callErr)
	}
	got, err := mem.ReadCString(bufAddr)
	if err != nil {
		t.Fatalf("ReadCString: %v", err)
	}
	if ret.Type != bytecode.TypeI32 || ret.Int != 7 || got != "||||0||" {
		t.Fatalf("__builtin_sprintf ret=%#v output=%q, want i32 7 and zero precision output", ret, got)
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

func TestFormatExternsSupportFloatAlternateForms(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	bufAddr := mustAlloc(t, mem, "buf:format-float-alternate", 48, 1, false, blockLocal)
	fmtAddr := mustAllocBytes(t, mem, "fmt:format-float-alternate", []byte("%#x|%#X|%#o|%#.0f|%#.0g|%#.0a\x00"), true, blockString)
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
		FloatValue(bytecode.TypeF64, 1),
		FloatValue(bytecode.TypeF64, 1),
		FloatValue(bytecode.TypeF64, 1),
	})
	if callErr != nil || exit != nil {
		t.Fatalf("__builtin_sprintf ret=%#v exit=%#v err=%v", ret, exit, callErr)
	}
	got, err := mem.ReadCString(bufAddr)
	if err != nil {
		t.Fatalf("ReadCString: %v", err)
	}
	want := "0xbeef|0XBEEF|0777|1.|1.|0x1.p+00"
	if ret.Type != bytecode.TypeI32 || ret.Int != uint64(len(want)) || got != want {
		t.Fatalf("__builtin_sprintf ret=%#v output=%q, want i32 %d and alternate float output", ret, got, len(want))
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

func TestFormatExternsSupportFloatWidthAndPrecision(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	bufAddr := mustAlloc(t, mem, "buf:format-float-width", 96, 1, false, blockLocal)
	fmtAddr := mustAllocBytes(t, mem, "fmt:format-float-width", []byte("%8.2f|%-8.1e|%8.1E|%.3g|%.3G|%12.2a|%-12.2A\x00"), true, blockString)
	fn, ok := reg.Lookup("__builtin_sprintf")
	if !ok {
		t.Fatal("missing __builtin_sprintf extern")
	}
	ret, exit, callErr := fn(context.Background(), &ExternContext{Memory: mem}, []Value{
		ObjectAddrValue(bufAddr),
		ObjectAddrValue(fmtAddr),
		FloatValue(bytecode.TypeF64, 1.5),
		FloatValue(bytecode.TypeF64, 12),
		FloatValue(bytecode.TypeF64, 12),
		FloatValue(bytecode.TypeF64, 12345),
		FloatValue(bytecode.TypeF64, 12345),
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
	want := "    1.50|1.2e+01 | 1.2E+01|1.23e+04|1.23E+04|  0x1.80p+00|0X1.80P+00  "
	if ret.Type != bytecode.TypeI32 || ret.Int != uint64(len(want)) || got != want {
		t.Fatalf("__builtin_sprintf ret=%#v output=%q, want i32 %d and float width output", ret, got, len(want))
	}
}

func TestFormatExternsSupportSpecialFloatValues(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	bufAddr := mustAlloc(t, mem, "buf:format-special-float", 64, 1, false, blockLocal)
	fmtAddr := mustAllocBytes(t, mem, "fmt:format-special-float", []byte("%f|%F|%e|%E|%g|%G|%a|%A|%+f|% f|%f\x00"), true, blockString)
	fn, ok := reg.Lookup("__builtin_sprintf")
	if !ok {
		t.Fatal("missing __builtin_sprintf extern")
	}
	ret, exit, callErr := fn(context.Background(), &ExternContext{Memory: mem}, []Value{
		ObjectAddrValue(bufAddr),
		ObjectAddrValue(fmtAddr),
		FloatValue(bytecode.TypeF64, math.Inf(1)),
		FloatValue(bytecode.TypeF64, math.Inf(1)),
		FloatValue(bytecode.TypeF64, math.NaN()),
		FloatValue(bytecode.TypeF64, math.NaN()),
		FloatValue(bytecode.TypeF64, math.Copysign(0, -1)),
		FloatValue(bytecode.TypeF64, math.Copysign(0, -1)),
		FloatValue(bytecode.TypeF64, math.Inf(1)),
		FloatValue(bytecode.TypeF64, math.Inf(1)),
		FloatValue(bytecode.TypeF64, math.Inf(1)),
		FloatValue(bytecode.TypeF64, math.Inf(1)),
		FloatValue(bytecode.TypeF64, math.Inf(-1)),
	})
	if callErr != nil || exit != nil {
		t.Fatalf("__builtin_sprintf ret=%#v exit=%#v err=%v", ret, exit, callErr)
	}
	got, err := mem.ReadCString(bufAddr)
	if err != nil {
		t.Fatalf("ReadCString: %v", err)
	}
	want := "inf|INF|nan|NAN|-0|-0|inf|INF|+inf| inf|-inf"
	if ret.Type != bytecode.TypeI32 || ret.Int != uint64(len(want)) || got != want {
		t.Fatalf("__builtin_sprintf ret=%#v output=%q, want i32 %d and special float output", ret, got, len(want))
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

func TestFormatExternsSupportNegativeDynamicWidthAndPrecision(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	bufAddr := mustAlloc(t, mem, "buf:format-negative-dynamic", 48, 1, false, blockLocal)
	fmtAddr := mustAllocBytes(t, mem, "fmt:format-negative-dynamic", []byte("%*d|%*s|%.*s|%*.*d|%*.*f\x00"), true, blockString)
	strAddr := mustAllocBytes(t, mem, "str:format-negative-dynamic", []byte("abcdef\x00"), true, blockString)
	shortAddr := mustAllocBytes(t, mem, "short:format-negative-dynamic", []byte("xy\x00"), true, blockString)
	fn, ok := reg.Lookup("__builtin_sprintf")
	if !ok {
		t.Fatal("missing __builtin_sprintf extern")
	}
	ret, exit, callErr := fn(context.Background(), &ExternContext{Memory: mem}, []Value{
		ObjectAddrValue(bufAddr),
		ObjectAddrValue(fmtAddr),
		IntValue(bytecode.TypeI32, -5),
		IntValue(bytecode.TypeI32, 7),
		IntValue(bytecode.TypeI32, -4),
		ObjectAddrValue(shortAddr),
		IntValue(bytecode.TypeI32, -1),
		ObjectAddrValue(strAddr),
		IntValue(bytecode.TypeI32, 6),
		IntValue(bytecode.TypeI32, -1),
		IntValue(bytecode.TypeI32, 42),
		IntValue(bytecode.TypeI32, -8),
		IntValue(bytecode.TypeI32, 2),
		FloatValue(bytecode.TypeF64, 1.5),
	})
	if callErr != nil || exit != nil {
		t.Fatalf("__builtin_sprintf ret=%#v exit=%#v err=%v", ret, exit, callErr)
	}
	got, err := mem.ReadCString(bufAddr)
	if err != nil {
		t.Fatalf("ReadCString: %v", err)
	}
	want := "7    |xy  |abcdef|    42|1.50    "
	if ret.Type != bytecode.TypeI32 || ret.Int != uint64(len(want)) || got != want {
		t.Fatalf("__builtin_sprintf ret=%#v output=%q, want i32 %d and negative dynamic output", ret, got, len(want))
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

func TestFormatExternsSupportWideSizedWriteCount(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	bufAddr := mustAlloc(t, mem, "buf:format-wide-sized-count", 16, 1, false, blockLocal)
	countLong := mustAlloc(t, mem, "count-long:format-wide-sized-count", 8, 8, false, blockLocal)
	countJ := mustAlloc(t, mem, "count-j:format-wide-sized-count", 8, 8, false, blockLocal)
	countZ := mustAlloc(t, mem, "count-z:format-wide-sized-count", 8, 8, false, blockLocal)
	countT := mustAlloc(t, mem, "count-t:format-wide-sized-count", 8, 8, false, blockLocal)
	fmtAddr := mustAllocBytes(t, mem, "fmt:format-wide-sized-count", []byte("a%lnb%jnc%znd%tne\x00"), true, blockString)
	fn, ok := reg.Lookup("__builtin_sprintf")
	if !ok {
		t.Fatal("missing __builtin_sprintf extern")
	}
	ret, exit, callErr := fn(context.Background(), &ExternContext{Memory: mem}, []Value{
		ObjectAddrValue(bufAddr),
		ObjectAddrValue(fmtAddr),
		PtrValue(countLong),
		PtrValue(countJ),
		PtrValue(countZ),
		PtrValue(countT),
	})
	if callErr != nil || exit != nil {
		t.Fatalf("__builtin_sprintf ret=%#v exit=%#v err=%v", ret, exit, callErr)
	}
	got, err := mem.ReadCString(bufAddr)
	if err != nil {
		t.Fatalf("ReadCString: %v", err)
	}
	vLong, err := mem.Load(countLong, bytecode.TypeI64, 8)
	if err != nil {
		t.Fatalf("Load countLong: %v", err)
	}
	vJ, err := mem.Load(countJ, bytecode.TypeI64, 8)
	if err != nil {
		t.Fatalf("Load countJ: %v", err)
	}
	vZ, err := mem.Load(countZ, bytecode.TypeI64, 8)
	if err != nil {
		t.Fatalf("Load countZ: %v", err)
	}
	vT, err := mem.Load(countT, bytecode.TypeI64, 8)
	if err != nil {
		t.Fatalf("Load countT: %v", err)
	}
	if ret.Type != bytecode.TypeI32 || ret.Int != 5 || got != "abcde" || vLong.Int != 1 || vJ.Int != 2 || vZ.Int != 3 || vT.Int != 4 {
		t.Fatalf("__builtin_sprintf ret=%#v output=%q counts=%#v/%#v/%#v/%#v, want abcde and 1/2/3/4", ret, got, vLong, vJ, vZ, vT)
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

func TestPlainMathRealIntExterns(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	tests := []struct {
		name string
		args []Value
		want bytecode.ValueType
		val  float64
	}{
		{name: "scalbnf", args: []Value{FloatValue(bytecode.TypeF32, 2), IntValue(bytecode.TypeI32, 3)}, want: bytecode.TypeF32, val: 16},
		{name: "scalbn", args: []Value{FloatValue(bytecode.TypeF64, 2), IntValue(bytecode.TypeI32, 3)}, want: bytecode.TypeF64, val: 16},
		{name: "scalbnl", args: []Value{FloatValue(bytecode.TypeFLong, 2), IntValue(bytecode.TypeI32, 3)}, want: bytecode.TypeFLong, val: 16},
		{name: "scalblnf", args: []Value{FloatValue(bytecode.TypeF32, 3), IntValue(bytecode.TypeI64, 2)}, want: bytecode.TypeF32, val: 12},
		{name: "scalbln", args: []Value{FloatValue(bytecode.TypeF64, 3), IntValue(bytecode.TypeI64, 2)}, want: bytecode.TypeF64, val: 12},
		{name: "scalblnl", args: []Value{FloatValue(bytecode.TypeFLong, 3), IntValue(bytecode.TypeI64, 2)}, want: bytecode.TypeFLong, val: 12},
		{name: "ldexpf", args: []Value{FloatValue(bytecode.TypeF32, 2), IntValue(bytecode.TypeI32, 3)}, want: bytecode.TypeF32, val: 16},
		{name: "ldexp", args: []Value{FloatValue(bytecode.TypeF64, 2), IntValue(bytecode.TypeI32, 3)}, want: bytecode.TypeF64, val: 16},
		{name: "ldexpl", args: []Value{FloatValue(bytecode.TypeFLong, 2), IntValue(bytecode.TypeI32, 3)}, want: bytecode.TypeFLong, val: 16},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn, ok := reg.Lookup(tt.name)
			if !ok {
				t.Fatalf("missing %s extern", tt.name)
			}
			ret, exit, err := fn(context.Background(), nil, tt.args)
			if err != nil || exit != nil {
				t.Fatalf("%s ret=%#v exit=%#v err=%v", tt.name, ret, exit, err)
			}
			if ret.Type != tt.want || ret.Float != tt.val {
				t.Fatalf("%s ret=%#v, want %s %v", tt.name, ret, tt.want, tt.val)
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
		{name: "plain ilogbf", fn: "ilogbf", arg: FloatValue(bytecode.TypeF32, 8), want: 3},
		{name: "plain ilogb", fn: "ilogb", arg: FloatValue(bytecode.TypeF64, 8), want: 3},
		{name: "plain ilogbl", fn: "ilogbl", arg: FloatValue(bytecode.TypeFLong, 8), want: 3},
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
		{name: "plain lrintf", fn: "lrintf", arg: FloatValue(bytecode.TypeF32, 3), want: 3},
		{name: "plain lrint", fn: "lrint", arg: FloatValue(bytecode.TypeF64, 3), want: 3},
		{name: "plain lrintl", fn: "lrintl", arg: FloatValue(bytecode.TypeFLong, 3), want: 3},
		{name: "plain lroundf", fn: "lroundf", arg: FloatValue(bytecode.TypeF32, 4), want: 4},
		{name: "plain lround", fn: "lround", arg: FloatValue(bytecode.TypeF64, 4), want: 4},
		{name: "plain lroundl", fn: "lroundl", arg: FloatValue(bytecode.TypeFLong, 4), want: 4},
		{name: "plain llrintf", fn: "llrintf", arg: FloatValue(bytecode.TypeF32, 5), want: 5},
		{name: "plain llrint", fn: "llrint", arg: FloatValue(bytecode.TypeF64, 5), want: 5},
		{name: "plain llrintl", fn: "llrintl", arg: FloatValue(bytecode.TypeFLong, 5), want: 5},
		{name: "plain llroundf", fn: "llroundf", arg: FloatValue(bytecode.TypeF32, 6), want: 6},
		{name: "plain llround", fn: "llround", arg: FloatValue(bytecode.TypeF64, 6), want: 6},
		{name: "plain llroundl", fn: "llroundl", arg: FloatValue(bytecode.TypeFLong, 6), want: 6},
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
	tests := []struct {
		name string
		arg  Value
		want bytecode.ValueType
	}{
		{name: "frexpf", arg: FloatValue(bytecode.TypeF32, 8), want: bytecode.TypeF32},
		{name: "frexp", arg: FloatValue(bytecode.TypeF64, 8), want: bytecode.TypeF64},
		{name: "frexpl", arg: FloatValue(bytecode.TypeFLong, 8), want: bytecode.TypeFLong},
		{name: "__cvm_tgmath_frexp", arg: FloatValue(bytecode.TypeF64, 8), want: bytecode.TypeF64},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn, ok := reg.Lookup(tt.name)
			if !ok {
				t.Fatalf("missing %s extern", tt.name)
			}
			mem := NewMemory(bytecode.DefaultTarget())
			expAddr := mustAlloc(t, mem, "exp", 4, 4, false, blockLocal)
			ret, exit, err := fn(context.Background(), &ExternContext{Memory: mem}, []Value{
				tt.arg,
				PtrValue(expAddr),
			})
			if err != nil || exit != nil {
				t.Fatalf("%s ret=%#v exit=%#v err=%v", tt.name, ret, exit, err)
			}
			if ret.Type != tt.want || ret.Float != 0.5 {
				t.Fatalf("%s ret = %#v, want %s 0.5", tt.name, ret, tt.want)
			}
			exp, err := mem.Load(expAddr, bytecode.TypeI32, 4)
			if err != nil {
				t.Fatalf("Load exponent: %v", err)
			}
			if int64(exp.Int) != 4 {
				t.Fatalf("frexp exponent = %#v, want 4", exp)
			}
		})
	}
}

func TestTgmathRemquoExtern(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	tests := []struct {
		name string
		arg  Value
		want bytecode.ValueType
	}{
		{name: "remquof", arg: FloatValue(bytecode.TypeF32, 4), want: bytecode.TypeF32},
		{name: "remquo", arg: FloatValue(bytecode.TypeF64, 4), want: bytecode.TypeF64},
		{name: "remquol", arg: FloatValue(bytecode.TypeFLong, 4), want: bytecode.TypeFLong},
		{name: "__cvm_tgmath_remquo", arg: FloatValue(bytecode.TypeF64, 4), want: bytecode.TypeF64},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn, ok := reg.Lookup(tt.name)
			if !ok {
				t.Fatalf("missing %s extern", tt.name)
			}
			mem := NewMemory(bytecode.DefaultTarget())
			quoAddr := mustAlloc(t, mem, "quo", 4, 4, false, blockLocal)
			ret, exit, err := fn(context.Background(), &ExternContext{Memory: mem}, []Value{
				tt.arg,
				tt.arg,
				PtrValue(quoAddr),
			})
			if err != nil || exit != nil {
				t.Fatalf("%s ret=%#v exit=%#v err=%v", tt.name, ret, exit, err)
			}
			if ret.Type != tt.want || ret.Float != 0 {
				t.Fatalf("%s ret = %#v, want %s 0", tt.name, ret, tt.want)
			}
			quo, err := mem.Load(quoAddr, bytecode.TypeI32, 4)
			if err != nil {
				t.Fatalf("Load quotient: %v", err)
			}
			if int64(quo.Int) != 1 {
				t.Fatalf("remquo quotient = %#v, want 1", quo)
			}
		})
	}
}

func TestComplexProjectionExterns(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	tests := []struct {
		name     string
		realType bytecode.ValueType
		realSize uint64
		want     float64
	}{
		{name: "crealf", realType: bytecode.TypeF32, realSize: 4, want: 3},
		{name: "creal", realType: bytecode.TypeF64, realSize: 8, want: 3},
		{name: "creall", realType: bytecode.TypeFLong, realSize: 16, want: 3},
		{name: "cimagf", realType: bytecode.TypeF32, realSize: 4, want: 4},
		{name: "cimag", realType: bytecode.TypeF64, realSize: 8, want: 4},
		{name: "cimagl", realType: bytecode.TypeFLong, realSize: 16, want: 4},
		{name: "cargf", realType: bytecode.TypeF32, realSize: 4, want: 0},
		{name: "carg", realType: bytecode.TypeF64, realSize: 8, want: 0},
		{name: "cargl", realType: bytecode.TypeFLong, realSize: 16, want: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn, ok := reg.Lookup(tt.name)
			if !ok {
				t.Fatalf("missing %s extern", tt.name)
			}
			mem := NewMemory(bytecode.DefaultTarget())
			z := mustAlloc(t, mem, "z", int64(tt.realSize*2), int64(tt.realSize), false, blockLocal)
			realPart := 3.0
			imagPart := 4.0
			if strings.HasPrefix(tt.name, "carg") {
				realPart = 1.0
				imagPart = 0.0
			}
			if err := mem.Store(z, tt.realType, int64(tt.realSize), FloatValue(tt.realType, realPart)); err != nil {
				t.Fatalf("store real: %v", err)
			}
			if err := mem.Store(z+tt.realSize, tt.realType, int64(tt.realSize), FloatValue(tt.realType, imagPart)); err != nil {
				t.Fatalf("store imag: %v", err)
			}
			ret, exit, err := fn(context.Background(), &ExternContext{Memory: mem}, []Value{ObjectAddrValue(z)})
			if err != nil || exit != nil {
				t.Fatalf("%s ret=%#v exit=%#v err=%v", tt.name, ret, exit, err)
			}
			if ret.Type != tt.realType || ret.Float != tt.want {
				t.Fatalf("%s ret = %#v, want %s %v", tt.name, ret, tt.realType, tt.want)
			}
		})
	}
}

func TestComplexAbsExterns(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	tests := []struct {
		name     string
		realType bytecode.ValueType
		realSize uint64
		want     float64
	}{
		{name: "cabsf", realType: bytecode.TypeF32, realSize: 4, want: 5},
		{name: "cabs", realType: bytecode.TypeF64, realSize: 8, want: 5},
		{name: "cabsl", realType: bytecode.TypeFLong, realSize: 16, want: 5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn, ok := reg.Lookup(tt.name)
			if !ok {
				t.Fatalf("missing %s extern", tt.name)
			}
			mem := NewMemory(bytecode.DefaultTarget())
			z := mustAlloc(t, mem, "z", int64(tt.realSize*2), int64(tt.realSize), false, blockLocal)
			if err := mem.Store(z, tt.realType, int64(tt.realSize), FloatValue(tt.realType, 3)); err != nil {
				t.Fatalf("store real: %v", err)
			}
			if err := mem.Store(z+tt.realSize, tt.realType, int64(tt.realSize), FloatValue(tt.realType, 4)); err != nil {
				t.Fatalf("store imag: %v", err)
			}
			ret, exit, err := fn(context.Background(), &ExternContext{Memory: mem}, []Value{ObjectAddrValue(z)})
			if err != nil || exit != nil {
				t.Fatalf("%s ret=%#v exit=%#v err=%v", tt.name, ret, exit, err)
			}
			if ret.Type != tt.realType || ret.Float != tt.want {
				t.Fatalf("%s ret = %#v, want %s %v", tt.name, ret, tt.realType, tt.want)
			}
		})
	}
}

func TestComplexUnaryExterns(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	tests := []struct {
		name     string
		realType bytecode.ValueType
		realSize uint64
		wantImag float64
	}{
		{name: "conjf", realType: bytecode.TypeF32, realSize: 4, wantImag: -4},
		{name: "conj", realType: bytecode.TypeF64, realSize: 8, wantImag: -4},
		{name: "conjl", realType: bytecode.TypeFLong, realSize: 16, wantImag: -4},
		{name: "cprojf", realType: bytecode.TypeF32, realSize: 4, wantImag: 4},
		{name: "cproj", realType: bytecode.TypeF64, realSize: 8, wantImag: 4},
		{name: "cprojl", realType: bytecode.TypeFLong, realSize: 16, wantImag: 4},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn, ok := reg.Lookup(tt.name)
			if !ok {
				t.Fatalf("missing %s extern", tt.name)
			}
			mem := NewMemory(bytecode.DefaultTarget())
			z := mustAlloc(t, mem, "z", int64(tt.realSize*2), int64(tt.realSize), false, blockLocal)
			if err := mem.Store(z, tt.realType, int64(tt.realSize), FloatValue(tt.realType, 3)); err != nil {
				t.Fatalf("store real: %v", err)
			}
			if err := mem.Store(z+tt.realSize, tt.realType, int64(tt.realSize), FloatValue(tt.realType, 4)); err != nil {
				t.Fatalf("store imag: %v", err)
			}
			ret, exit, err := fn(context.Background(), &ExternContext{Memory: mem}, []Value{ObjectAddrValue(z)})
			if err != nil || exit != nil {
				t.Fatalf("%s ret=%#v exit=%#v err=%v", tt.name, ret, exit, err)
			}
			if ret.Type != bytecode.TypeObjectAddr {
				t.Fatalf("%s ret = %#v, want object address", tt.name, ret)
			}
			realPart, err := mem.Load(ret.Int, tt.realType, int64(tt.realSize))
			if err != nil {
				t.Fatalf("load real: %v", err)
			}
			imagPart, err := mem.Load(ret.Int+tt.realSize, tt.realType, int64(tt.realSize))
			if err != nil {
				t.Fatalf("load imag: %v", err)
			}
			if realPart.Float != 3 || imagPart.Float != tt.wantImag {
				t.Fatalf("%s result = (%v, %v), want (3, %v)", tt.name, realPart.Float, imagPart.Float, tt.wantImag)
			}
		})
	}
}

func TestComplexTrigExterns(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	tests := []struct {
		name     string
		realType bytecode.ValueType
		realSize uint64
		wantReal float64
	}{
		{name: "csinf", realType: bytecode.TypeF32, realSize: 4, wantReal: 0},
		{name: "csin", realType: bytecode.TypeF64, realSize: 8, wantReal: 0},
		{name: "csinl", realType: bytecode.TypeFLong, realSize: 16, wantReal: 0},
		{name: "ccosf", realType: bytecode.TypeF32, realSize: 4, wantReal: 1},
		{name: "ccos", realType: bytecode.TypeF64, realSize: 8, wantReal: 1},
		{name: "ccosl", realType: bytecode.TypeFLong, realSize: 16, wantReal: 1},
		{name: "ctanf", realType: bytecode.TypeF32, realSize: 4, wantReal: 0},
		{name: "ctan", realType: bytecode.TypeF64, realSize: 8, wantReal: 0},
		{name: "ctanl", realType: bytecode.TypeFLong, realSize: 16, wantReal: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn, ok := reg.Lookup(tt.name)
			if !ok {
				t.Fatalf("missing %s extern", tt.name)
			}
			mem := NewMemory(bytecode.DefaultTarget())
			z := mustAlloc(t, mem, "z", int64(tt.realSize*2), int64(tt.realSize), false, blockLocal)
			if err := mem.Store(z, tt.realType, int64(tt.realSize), FloatValue(tt.realType, 0)); err != nil {
				t.Fatalf("store real: %v", err)
			}
			if err := mem.Store(z+tt.realSize, tt.realType, int64(tt.realSize), FloatValue(tt.realType, 0)); err != nil {
				t.Fatalf("store imag: %v", err)
			}
			ret, exit, err := fn(context.Background(), &ExternContext{Memory: mem}, []Value{ObjectAddrValue(z)})
			if err != nil || exit != nil {
				t.Fatalf("%s ret=%#v exit=%#v err=%v", tt.name, ret, exit, err)
			}
			if ret.Type != bytecode.TypeObjectAddr {
				t.Fatalf("%s ret = %#v, want object address", tt.name, ret)
			}
			realPart, err := mem.Load(ret.Int, tt.realType, int64(tt.realSize))
			if err != nil {
				t.Fatalf("load real: %v", err)
			}
			imagPart, err := mem.Load(ret.Int+tt.realSize, tt.realType, int64(tt.realSize))
			if err != nil {
				t.Fatalf("load imag: %v", err)
			}
			if realPart.Float != tt.wantReal || imagPart.Float != 0 {
				t.Fatalf("%s result = (%v, %v), want (%v, 0)", tt.name, realPart.Float, imagPart.Float, tt.wantReal)
			}
		})
	}
}

func TestComplexHyperbolicExterns(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	tests := []struct {
		name     string
		realType bytecode.ValueType
		realSize uint64
		wantReal float64
	}{
		{name: "csinhf", realType: bytecode.TypeF32, realSize: 4, wantReal: 0},
		{name: "csinh", realType: bytecode.TypeF64, realSize: 8, wantReal: 0},
		{name: "csinhl", realType: bytecode.TypeFLong, realSize: 16, wantReal: 0},
		{name: "ccoshf", realType: bytecode.TypeF32, realSize: 4, wantReal: 1},
		{name: "ccosh", realType: bytecode.TypeF64, realSize: 8, wantReal: 1},
		{name: "ccoshl", realType: bytecode.TypeFLong, realSize: 16, wantReal: 1},
		{name: "ctanhf", realType: bytecode.TypeF32, realSize: 4, wantReal: 0},
		{name: "ctanh", realType: bytecode.TypeF64, realSize: 8, wantReal: 0},
		{name: "ctanhl", realType: bytecode.TypeFLong, realSize: 16, wantReal: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn, ok := reg.Lookup(tt.name)
			if !ok {
				t.Fatalf("missing %s extern", tt.name)
			}
			mem := NewMemory(bytecode.DefaultTarget())
			z := mustAlloc(t, mem, "z", int64(tt.realSize*2), int64(tt.realSize), false, blockLocal)
			if err := mem.Store(z, tt.realType, int64(tt.realSize), FloatValue(tt.realType, 0)); err != nil {
				t.Fatalf("store real: %v", err)
			}
			if err := mem.Store(z+tt.realSize, tt.realType, int64(tt.realSize), FloatValue(tt.realType, 0)); err != nil {
				t.Fatalf("store imag: %v", err)
			}
			ret, exit, err := fn(context.Background(), &ExternContext{Memory: mem}, []Value{ObjectAddrValue(z)})
			if err != nil || exit != nil {
				t.Fatalf("%s ret=%#v exit=%#v err=%v", tt.name, ret, exit, err)
			}
			if ret.Type != bytecode.TypeObjectAddr {
				t.Fatalf("%s ret = %#v, want object address", tt.name, ret)
			}
			realPart, err := mem.Load(ret.Int, tt.realType, int64(tt.realSize))
			if err != nil {
				t.Fatalf("load real: %v", err)
			}
			imagPart, err := mem.Load(ret.Int+tt.realSize, tt.realType, int64(tt.realSize))
			if err != nil {
				t.Fatalf("load imag: %v", err)
			}
			if realPart.Float != tt.wantReal || imagPart.Float != 0 {
				t.Fatalf("%s result = (%v, %v), want (%v, 0)", tt.name, realPart.Float, imagPart.Float, tt.wantReal)
			}
		})
	}
}

func TestComplexInverseTrigExterns(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	tests := []struct {
		name      string
		realType  bytecode.ValueType
		realSize  uint64
		inputReal float64
	}{
		{name: "casinf", realType: bytecode.TypeF32, realSize: 4, inputReal: 0},
		{name: "casin", realType: bytecode.TypeF64, realSize: 8, inputReal: 0},
		{name: "casinl", realType: bytecode.TypeFLong, realSize: 16, inputReal: 0},
		{name: "cacosf", realType: bytecode.TypeF32, realSize: 4, inputReal: 1},
		{name: "cacos", realType: bytecode.TypeF64, realSize: 8, inputReal: 1},
		{name: "cacosl", realType: bytecode.TypeFLong, realSize: 16, inputReal: 1},
		{name: "catanf", realType: bytecode.TypeF32, realSize: 4, inputReal: 0},
		{name: "catan", realType: bytecode.TypeF64, realSize: 8, inputReal: 0},
		{name: "catanl", realType: bytecode.TypeFLong, realSize: 16, inputReal: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn, ok := reg.Lookup(tt.name)
			if !ok {
				t.Fatalf("missing %s extern", tt.name)
			}
			mem := NewMemory(bytecode.DefaultTarget())
			z := mustAlloc(t, mem, "z", int64(tt.realSize*2), int64(tt.realSize), false, blockLocal)
			if err := mem.Store(z, tt.realType, int64(tt.realSize), FloatValue(tt.realType, tt.inputReal)); err != nil {
				t.Fatalf("store real: %v", err)
			}
			if err := mem.Store(z+tt.realSize, tt.realType, int64(tt.realSize), FloatValue(tt.realType, 0)); err != nil {
				t.Fatalf("store imag: %v", err)
			}
			ret, exit, err := fn(context.Background(), &ExternContext{Memory: mem}, []Value{ObjectAddrValue(z)})
			if err != nil || exit != nil {
				t.Fatalf("%s ret=%#v exit=%#v err=%v", tt.name, ret, exit, err)
			}
			realPart, err := mem.Load(ret.Int, tt.realType, int64(tt.realSize))
			if err != nil {
				t.Fatalf("load real: %v", err)
			}
			imagPart, err := mem.Load(ret.Int+tt.realSize, tt.realType, int64(tt.realSize))
			if err != nil {
				t.Fatalf("load imag: %v", err)
			}
			if realPart.Float != 0 || imagPart.Float != 0 {
				t.Fatalf("%s result = (%v, %v), want (0, 0)", tt.name, realPart.Float, imagPart.Float)
			}
		})
	}
}

func TestComplexInverseHyperbolicExterns(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	tests := []struct {
		name      string
		realType  bytecode.ValueType
		realSize  uint64
		inputReal float64
	}{
		{name: "casinhf", realType: bytecode.TypeF32, realSize: 4, inputReal: 0},
		{name: "casinh", realType: bytecode.TypeF64, realSize: 8, inputReal: 0},
		{name: "casinhl", realType: bytecode.TypeFLong, realSize: 16, inputReal: 0},
		{name: "cacoshf", realType: bytecode.TypeF32, realSize: 4, inputReal: 1},
		{name: "cacosh", realType: bytecode.TypeF64, realSize: 8, inputReal: 1},
		{name: "cacoshl", realType: bytecode.TypeFLong, realSize: 16, inputReal: 1},
		{name: "catanhf", realType: bytecode.TypeF32, realSize: 4, inputReal: 0},
		{name: "catanh", realType: bytecode.TypeF64, realSize: 8, inputReal: 0},
		{name: "catanhl", realType: bytecode.TypeFLong, realSize: 16, inputReal: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn, ok := reg.Lookup(tt.name)
			if !ok {
				t.Fatalf("missing %s extern", tt.name)
			}
			mem := NewMemory(bytecode.DefaultTarget())
			z := mustAlloc(t, mem, "z", int64(tt.realSize*2), int64(tt.realSize), false, blockLocal)
			if err := mem.Store(z, tt.realType, int64(tt.realSize), FloatValue(tt.realType, tt.inputReal)); err != nil {
				t.Fatalf("store real: %v", err)
			}
			if err := mem.Store(z+tt.realSize, tt.realType, int64(tt.realSize), FloatValue(tt.realType, 0)); err != nil {
				t.Fatalf("store imag: %v", err)
			}
			ret, exit, err := fn(context.Background(), &ExternContext{Memory: mem}, []Value{ObjectAddrValue(z)})
			if err != nil || exit != nil {
				t.Fatalf("%s ret=%#v exit=%#v err=%v", tt.name, ret, exit, err)
			}
			realPart, err := mem.Load(ret.Int, tt.realType, int64(tt.realSize))
			if err != nil {
				t.Fatalf("load real: %v", err)
			}
			imagPart, err := mem.Load(ret.Int+tt.realSize, tt.realType, int64(tt.realSize))
			if err != nil {
				t.Fatalf("load imag: %v", err)
			}
			if realPart.Float != 0 || imagPart.Float != 0 {
				t.Fatalf("%s result = (%v, %v), want (0, 0)", tt.name, realPart.Float, imagPart.Float)
			}
		})
	}
}

func TestComplexExpLogSqrtExterns(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	tests := []struct {
		name      string
		realType  bytecode.ValueType
		realSize  uint64
		inputReal float64
		wantReal  float64
	}{
		{name: "cexpf", realType: bytecode.TypeF32, realSize: 4, inputReal: 0, wantReal: 1},
		{name: "cexp", realType: bytecode.TypeF64, realSize: 8, inputReal: 0, wantReal: 1},
		{name: "cexpl", realType: bytecode.TypeFLong, realSize: 16, inputReal: 0, wantReal: 1},
		{name: "clogf", realType: bytecode.TypeF32, realSize: 4, inputReal: 1, wantReal: 0},
		{name: "clog", realType: bytecode.TypeF64, realSize: 8, inputReal: 1, wantReal: 0},
		{name: "clogl", realType: bytecode.TypeFLong, realSize: 16, inputReal: 1, wantReal: 0},
		{name: "csqrtf", realType: bytecode.TypeF32, realSize: 4, inputReal: 0, wantReal: 0},
		{name: "csqrt", realType: bytecode.TypeF64, realSize: 8, inputReal: 0, wantReal: 0},
		{name: "csqrtl", realType: bytecode.TypeFLong, realSize: 16, inputReal: 0, wantReal: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn, ok := reg.Lookup(tt.name)
			if !ok {
				t.Fatalf("missing %s extern", tt.name)
			}
			mem := NewMemory(bytecode.DefaultTarget())
			z := mustAlloc(t, mem, "z", int64(tt.realSize*2), int64(tt.realSize), false, blockLocal)
			if err := mem.Store(z, tt.realType, int64(tt.realSize), FloatValue(tt.realType, tt.inputReal)); err != nil {
				t.Fatalf("store real: %v", err)
			}
			if err := mem.Store(z+tt.realSize, tt.realType, int64(tt.realSize), FloatValue(tt.realType, 0)); err != nil {
				t.Fatalf("store imag: %v", err)
			}
			ret, exit, err := fn(context.Background(), &ExternContext{Memory: mem}, []Value{ObjectAddrValue(z)})
			if err != nil || exit != nil {
				t.Fatalf("%s ret=%#v exit=%#v err=%v", tt.name, ret, exit, err)
			}
			realPart, err := mem.Load(ret.Int, tt.realType, int64(tt.realSize))
			if err != nil {
				t.Fatalf("load real: %v", err)
			}
			imagPart, err := mem.Load(ret.Int+tt.realSize, tt.realType, int64(tt.realSize))
			if err != nil {
				t.Fatalf("load imag: %v", err)
			}
			if realPart.Float != tt.wantReal || imagPart.Float != 0 {
				t.Fatalf("%s result = (%v, %v), want (%v, 0)", tt.name, realPart.Float, imagPart.Float, tt.wantReal)
			}
		})
	}
}

func TestComplexPowExterns(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	tests := []struct {
		name     string
		realType bytecode.ValueType
		realSize uint64
	}{
		{name: "cpowf", realType: bytecode.TypeF32, realSize: 4},
		{name: "cpow", realType: bytecode.TypeF64, realSize: 8},
		{name: "cpowl", realType: bytecode.TypeFLong, realSize: 16},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn, ok := reg.Lookup(tt.name)
			if !ok {
				t.Fatalf("missing %s extern", tt.name)
			}
			mem := NewMemory(bytecode.DefaultTarget())
			left := mustAlloc(t, mem, "left", int64(tt.realSize*2), int64(tt.realSize), false, blockLocal)
			right := mustAlloc(t, mem, "right", int64(tt.realSize*2), int64(tt.realSize), false, blockLocal)
			if err := mem.Store(left, tt.realType, int64(tt.realSize), FloatValue(tt.realType, 1)); err != nil {
				t.Fatalf("store left real: %v", err)
			}
			if err := mem.Store(left+tt.realSize, tt.realType, int64(tt.realSize), FloatValue(tt.realType, 0)); err != nil {
				t.Fatalf("store left imag: %v", err)
			}
			if err := mem.Store(right, tt.realType, int64(tt.realSize), FloatValue(tt.realType, 2)); err != nil {
				t.Fatalf("store right real: %v", err)
			}
			if err := mem.Store(right+tt.realSize, tt.realType, int64(tt.realSize), FloatValue(tt.realType, 0)); err != nil {
				t.Fatalf("store right imag: %v", err)
			}
			ret, exit, err := fn(context.Background(), &ExternContext{Memory: mem}, []Value{ObjectAddrValue(left), ObjectAddrValue(right)})
			if err != nil || exit != nil {
				t.Fatalf("%s ret=%#v exit=%#v err=%v", tt.name, ret, exit, err)
			}
			realPart, err := mem.Load(ret.Int, tt.realType, int64(tt.realSize))
			if err != nil {
				t.Fatalf("load real: %v", err)
			}
			imagPart, err := mem.Load(ret.Int+tt.realSize, tt.realType, int64(tt.realSize))
			if err != nil {
				t.Fatalf("load imag: %v", err)
			}
			if realPart.Float != 1 || imagPart.Float != 0 {
				t.Fatalf("%s result = (%v, %v), want (1, 0)", tt.name, realPart.Float, imagPart.Float)
			}
		})
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
