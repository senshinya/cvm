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
	for _, name := range []string{"FILE", "size_t", "stdin", "stdout", "stderr", "fputs", "fputs_unlocked", "fputc", "fputc_unlocked", "puts", "putchar", "fflush", "printf", "fprintf", "sprintf", "snprintf", "vprintf", "vfprintf", "vsprintf", "vsnprintf"} {
		if !hasIdentifier(res.Tokens, name) {
			t.Fatalf("stdio identifier %q missing: %#v", name, res.Tokens)
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
