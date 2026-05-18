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
	for _, name := range []string{"FILE", "size_t", "stdin", "stdout", "stderr", "fputs", "fputs_unlocked", "fputc", "fputc_unlocked", "putc", "putc_unlocked", "puts", "puts_unlocked", "putchar", "putchar_unlocked", "getchar", "getchar_unlocked", "fflush", "fflush_unlocked", "fclose", "fileno", "fileno_unlocked", "ferror", "ferror_unlocked", "clearerr", "clearerr_unlocked", "feof", "feof_unlocked", "fwrite", "fwrite_unlocked", "fread", "fread_unlocked", "fgetc", "fgetc_unlocked", "getc", "getc_unlocked", "ungetc", "fgets", "fgets_unlocked", "setbuf", "setvbuf", "flockfile", "ftrylockfile", "funlockfile", "perror", "printf", "fprintf", "sprintf", "snprintf", "vprintf", "vprintf_unlocked", "vfprintf", "vfprintf_unlocked", "vsprintf", "vsnprintf"} {
		if !hasIdentifier(res.Tokens, name) {
			t.Fatalf("stdio identifier %q missing: %#v", name, res.Tokens)
		}
	}
}

func TestBuiltinStdioHeaderDefinesBufferingMacros(t *testing.T) {
	res, err := PreprocessSource("main.c", `
#include <stdio.h>
int values[] = { _IOFBF, _IOLBF, _IONBF, BUFSIZ };
`, Options{})
	if err != nil {
		t.Fatalf("PreprocessSource failed: %v", err)
	}
	for _, value := range []string{"0", "1", "2", "8192"} {
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
	for _, name := range []string{"abs", "labs", "llabs", "atoi", "atol", "atoll", "atof", "strtol", "strtoul", "strtoll", "strtoull", "strtod", "strtof", "strtold", "malloc", "calloc", "free", "strdup", "exit", "abort"} {
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

func TestBuiltinStringHeaderDeclaresReadOnlySurface(t *testing.T) {
	res, err := PreprocessSource("main.c", `
#include <string.h>
`, Options{})
	if err != nil {
		t.Fatalf("PreprocessSource failed: %v", err)
	}
	for _, name := range []string{"size_t", "strcmp", "memcmp", "strncmp", "strlen", "strnlen", "strchr", "strrchr", "strstr", "strpbrk", "strspn", "strcspn", "memchr", "memcpy", "memmove", "mempcpy", "memset", "bzero", "strcpy", "stpcpy", "strcat", "strncpy", "stpncpy", "strncat"} {
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
