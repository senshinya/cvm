package preprocessor

import (
	"testing"

	"shinya.click/cvm/entity"
)

func TestBuiltinStandardHeaders(t *testing.T) {
	res, err := PreprocessSource("main.c", `
#include <stdbool.h>
#include <stddef.h>
#include <stdint.h>
#include <stdint.h>
bool b = true;
size_t n;
intptr_t p;
intmax_t im;
uintmax_t um;
unsigned long sm = SIZE_MAX;
`, Options{})
	if err != nil {
		t.Fatalf("PreprocessSource failed: %v", err)
	}
	if !hasToken(res.Tokens, entity.BOOL) {
		t.Fatalf("stdbool bool did not become _Bool: %#v", res.Tokens)
	}
	if !hasIdentifier(res.Tokens, "size_t") || !hasIdentifier(res.Tokens, "intptr_t") || !hasIdentifier(res.Tokens, "intmax_t") || !hasIdentifier(res.Tokens, "uintmax_t") {
		t.Fatalf("standard typedef names missing: %#v", res.Tokens)
	}
	if !hasLexeme(res.Tokens, "18446744073709551615UL") {
		t.Fatalf("SIZE_MAX did not expand to target constant: %#v", res.Tokens)
	}
}

func TestBuiltinStdioHeaderDeclaresFormattingSurface(t *testing.T) {
	res, err := PreprocessSource("main.c", `
#include <stdio.h>
`, Options{})
	if err != nil {
		t.Fatalf("PreprocessSource failed: %v", err)
	}
	for _, name := range []string{"FILE", "size_t", "stdin", "stdout", "stderr", "remove", "rename", "fopen", "freopen", "tmpfile", "tmpnam", "fseek", "ftell", "rewind", "fputs", "fputs_unlocked", "fputc", "fputc_unlocked", "putc", "putc_unlocked", "puts", "puts_unlocked", "putchar", "putchar_unlocked", "getchar", "getchar_unlocked", "fflush", "fflush_unlocked", "fclose", "fileno", "fileno_unlocked", "ferror", "ferror_unlocked", "clearerr", "clearerr_unlocked", "feof", "feof_unlocked", "fwrite", "fwrite_unlocked", "fread", "fread_unlocked", "fgetc", "fgetc_unlocked", "getc", "getc_unlocked", "ungetc", "fgets", "fgets_unlocked", "setbuf", "setvbuf", "flockfile", "ftrylockfile", "funlockfile", "perror", "printf", "fprintf", "sprintf", "snprintf", "vprintf", "vprintf_unlocked", "vfprintf", "vfprintf_unlocked", "vsprintf", "vsnprintf"} {
		if !hasIdentifier(res.Tokens, name) {
			t.Fatalf("stdio identifier %q missing: %#v", name, res.Tokens)
		}
	}
}

func TestBuiltinStdioHeaderDefinesBufferingMacros(t *testing.T) {
	res, err := PreprocessSource("main.c", `
#include <stdio.h>
int values[] = { _IOFBF, _IOLBF, _IONBF, BUFSIZ, L_tmpnam, TMP_MAX, SEEK_SET, SEEK_CUR, SEEK_END };
`, Options{})
	if err != nil {
		t.Fatalf("PreprocessSource failed: %v", err)
	}
	for _, value := range []string{"0", "1", "2", "8192", "20", "25"} {
		if !hasLexeme(res.Tokens, value) {
			t.Fatalf("stdio buffering macro value %q missing: %#v", value, res.Tokens)
		}
	}
}

func TestBuiltinStdioHeaderDefinesEOF(t *testing.T) {
	res, err := PreprocessSource("main.c", `
#include <stdio.h>
int eof_value = EOF;
`, Options{})
	if err != nil {
		t.Fatalf("PreprocessSource failed: %v", err)
	}
	if !hasLexeme(res.Tokens, "-") || !hasLexeme(res.Tokens, "1") {
		t.Fatalf("EOF did not expand to -1: %#v", res.Tokens)
	}
}

func TestBuiltinStdlibHeaderDeclaresRuntimeSurface(t *testing.T) {
	res, err := PreprocessSource("main.c", `
#include <stdlib.h>
`, Options{})
	if err != nil {
		t.Fatalf("PreprocessSource failed: %v", err)
	}
	for _, name := range []string{"size_t", "div_t", "ldiv_t", "lldiv_t", "abs", "labs", "llabs", "div", "ldiv", "lldiv", "atoi", "atol", "atoll", "atof", "strtol", "strtoul", "strtoll", "strtoull", "strtod", "strtof", "strtold", "malloc", "calloc", "realloc", "free", "strdup", "rand", "srand", "getenv", "system", "atexit", "exit", "_Exit", "abort"} {
		if !hasIdentifier(res.Tokens, name) {
			t.Fatalf("stdlib identifier %q missing: %#v", name, res.Tokens)
		}
	}
}

func TestBuiltinCtypeHeaderDeclaresRuntimeSurface(t *testing.T) {
	res, err := PreprocessSource("main.c", `
#include <ctype.h>
`, Options{})
	if err != nil {
		t.Fatalf("PreprocessSource failed: %v", err)
	}
	for _, name := range []string{"isdigit", "isalpha", "isalnum", "isspace", "islower", "isupper", "isxdigit", "isprint", "isblank", "iscntrl", "isgraph", "ispunct", "tolower", "toupper"} {
		if !hasIdentifier(res.Tokens, name) {
			t.Fatalf("ctype identifier %q missing: %#v", name, res.Tokens)
		}
	}
}

func TestBuiltinLocaleHeaderDeclaresRuntimeSurface(t *testing.T) {
	res, err := PreprocessSource("main.c", `
#include <locale.h>
int categories[] = { LC_ALL, LC_COLLATE, LC_CTYPE, LC_MONETARY, LC_NUMERIC, LC_TIME };
`, Options{})
	if err != nil {
		t.Fatalf("PreprocessSource failed: %v", err)
	}
	if !hasIdentifier(res.Tokens, "setlocale") {
		t.Fatalf("locale identifier %q missing: %#v", "setlocale", res.Tokens)
	}
	for _, value := range []string{"0", "1", "2", "3", "4", "5"} {
		if !hasLexeme(res.Tokens, value) {
			t.Fatalf("locale category macro value %q missing: %#v", value, res.Tokens)
		}
	}
}

func TestBuiltinTimeHeaderDeclaresRuntimeSurface(t *testing.T) {
	res, err := PreprocessSource("main.c", `
#include <time.h>
clock_t c;
time_t t;
long ticks = CLOCKS_PER_SEC;
`, Options{})
	if err != nil {
		t.Fatalf("PreprocessSource failed: %v", err)
	}
	for _, name := range []string{"clock_t", "time_t", "clock", "difftime", "time"} {
		if !hasIdentifier(res.Tokens, name) {
			t.Fatalf("time identifier %q missing: %#v", name, res.Tokens)
		}
	}
	if !hasLexeme(res.Tokens, "1000000L") {
		t.Fatalf("CLOCKS_PER_SEC did not expand: %#v", res.Tokens)
	}
}

func TestBuiltinMathHeaderDeclaresRuntimeSurface(t *testing.T) {
	res, err := PreprocessSource("main.c", `
#include <math.h>
`, Options{})
	if err != nil {
		t.Fatalf("PreprocessSource failed: %v", err)
	}
	for _, name := range []string{"fabs", "fabsf", "fabsl", "sqrt", "sqrtf", "sqrtl"} {
		if !hasIdentifier(res.Tokens, name) {
			t.Fatalf("math identifier %q missing: %#v", name, res.Tokens)
		}
	}
}

func TestBuiltinErrnoHeaderDeclaresRuntimeSurface(t *testing.T) {
	res, err := PreprocessSource("main.c", `
#include <errno.h>
int *p = &errno;
int values[] = { EDOM, ERANGE, EILSEQ };
`, Options{})
	if err != nil {
		t.Fatalf("PreprocessSource failed: %v", err)
	}
	if !hasIdentifier(res.Tokens, "errno") {
		t.Fatalf("errno identifier missing: %#v", res.Tokens)
	}
	for _, value := range []string{"33", "34", "84"} {
		if !hasLexeme(res.Tokens, value) {
			t.Fatalf("errno macro value %q missing: %#v", value, res.Tokens)
		}
	}
}

func TestBuiltinAssertHeaderDeclaresRuntimeSurface(t *testing.T) {
	res, err := PreprocessSource("main.c", `
#include <assert.h>
void f(void) { assert(1); }
`, Options{})
	if err != nil {
		t.Fatalf("PreprocessSource failed: %v", err)
	}
	if !hasIdentifier(res.Tokens, "abort") {
		t.Fatalf("assert did not expand to abort path: %#v", res.Tokens)
	}

	res, err = PreprocessSource("main.c", `
#include <assert.h>
void f(void) { assert(0); }
`, Options{MacroActions: []MacroAction{{Kind: MacroDefine, Name: "NDEBUG", Value: "1"}}})
	if err != nil {
		t.Fatalf("PreprocessSource with NDEBUG failed: %v", err)
	}
	if hasIdentifier(res.Tokens, "abort") {
		t.Fatalf("NDEBUG assert expanded to abort path: %#v", res.Tokens)
	}
}

func TestBuiltinStringHeaderDeclaresReadOnlySurface(t *testing.T) {
	res, err := PreprocessSource("main.c", `
#include <string.h>
`, Options{})
	if err != nil {
		t.Fatalf("PreprocessSource failed: %v", err)
	}
	for _, name := range []string{"size_t", "strcmp", "memcmp", "strncmp", "strcoll", "strlen", "strnlen", "strerror", "strchr", "strrchr", "strstr", "strpbrk", "strspn", "strcspn", "memchr", "strtok", "strxfrm", "memcpy", "memmove", "mempcpy", "memset", "bzero", "strcpy", "stpcpy", "strcat", "strncpy", "stpncpy", "strncat"} {
		if !hasIdentifier(res.Tokens, name) {
			t.Fatalf("string identifier %q missing: %#v", name, res.Tokens)
		}
	}
}

func TestBuiltinStringsHeaderDeclaresRuntimeSurface(t *testing.T) {
	res, err := PreprocessSource("main.c", `
#include <strings.h>
`, Options{})
	if err != nil {
		t.Fatalf("PreprocessSource failed: %v", err)
	}
	for _, name := range []string{"size_t", "bcmp", "bcopy", "bzero"} {
		if !hasIdentifier(res.Tokens, name) {
			t.Fatalf("strings identifier %q missing: %#v", name, res.Tokens)
		}
	}
}

func hasToken(tokens []entity.Token, typ entity.TokenType) bool {
	for _, tok := range tokens {
		if tok.Typ == typ {
			return true
		}
	}
	return false
}
