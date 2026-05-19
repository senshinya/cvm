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
	for _, name := range []string{"FILE", "fpos_t", "size_t", "stdin", "stdout", "stderr", "remove", "rename", "fopen", "freopen", "tmpfile", "tmpnam", "fseek", "ftell", "rewind", "fgetpos", "fsetpos", "fputs", "fputs_unlocked", "fputc", "fputc_unlocked", "putc", "putc_unlocked", "puts", "puts_unlocked", "putchar", "putchar_unlocked", "getchar", "getchar_unlocked", "fflush", "fflush_unlocked", "fclose", "fileno", "fileno_unlocked", "ferror", "ferror_unlocked", "clearerr", "clearerr_unlocked", "feof", "feof_unlocked", "fwrite", "fwrite_unlocked", "fread", "fread_unlocked", "fgetc", "fgetc_unlocked", "getc", "getc_unlocked", "ungetc", "fgets", "fgets_unlocked", "setbuf", "setvbuf", "flockfile", "ftrylockfile", "funlockfile", "perror", "printf", "fprintf", "sprintf", "snprintf", "vprintf", "vprintf_unlocked", "vfprintf", "vfprintf_unlocked", "vsprintf", "vsnprintf"} {
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
	for _, name := range []string{"size_t", "wchar_t", "div_t", "ldiv_t", "lldiv_t", "abs", "labs", "llabs", "div", "ldiv", "lldiv", "atoi", "atol", "atoll", "atof", "strtol", "strtoul", "strtoll", "strtoull", "strtod", "strtof", "strtold", "mblen", "mbtowc", "wctomb", "mbstowcs", "wcstombs", "malloc", "calloc", "realloc", "free", "strdup", "rand", "srand", "getenv", "system", "atexit", "exit", "_Exit", "abort"} {
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
struct lconv *lc = localeconv();
`, Options{})
	if err != nil {
		t.Fatalf("PreprocessSource failed: %v", err)
	}
	for _, name := range []string{"lconv", "decimal_point", "thousands_sep", "grouping", "setlocale", "localeconv"} {
		if !hasIdentifier(res.Tokens, name) {
			t.Fatalf("locale identifier %q missing: %#v", name, res.Tokens)
		}
	}
	for _, value := range []string{"0", "1", "2", "3", "4", "5"} {
		if !hasLexeme(res.Tokens, value) {
			t.Fatalf("locale category macro value %q missing: %#v", value, res.Tokens)
		}
	}
}

func TestBuiltinWideHeadersDeclareRuntimeSurface(t *testing.T) {
	res, err := PreprocessSource("main.c", `
#include <wchar.h>
#include <wctype.h>
wint_t wc = WEOF;
mbstate_t state = {0};
wctype_t cls = 0;
wctrans_t trans = 0;
FILE *wf = 0;
int checks[] = {
  iswalnum(L'A'), iswalpha(L'A'), iswblank(L' '), iswcntrl(L'\n'),
  iswdigit(L'7'), iswgraph(L'!'), iswlower(L'a'), iswprint(L' '),
  iswpunct(L'!'), iswspace(L'\t'), iswupper(L'Z'), iswxdigit(L'f'),
  towlower(L'A'), towupper(L'a'), iswctype(L'A', cls), towctrans(L'A', trans)
};
size_t len = mbrlen("A", 1, &state);
size_t converted = mbrtowc(0, "A", 1, &state);
size_t emitted = wcrtomb(0, L'A', &state);
const char *srcp = "A";
size_t wide_count = mbsrtowcs(0, &srcp, 0, &state);
const wchar_t *wsrcp = L"A";
size_t narrow_count = wcsrtombs(0, &wsrcp, 0, &state);
wctype_t named_cls = wctype("alpha");
wctrans_t named_trans = wctrans("tolower");
wint_t wrote = fputwc(L'A', wf);
wint_t put = putwc(L'B', wf);
wint_t put_stdout = putwchar(L'C');
wint_t read = fgetwc(wf);
wint_t got = getwc(wf);
wint_t got_stdin = getwchar();
wint_t pushed = ungetwc(L'D', wf);
int wide = fwide(wf, 1);
int wrote_string = fputws(L"wide", wf);
wchar_t *read_string = fgetws(0, 0, wf);
wchar_t wbuf[16];
int wide_print = wprintf(L"%d", 1);
int file_wide_print = fwprintf(wf, L"%d", 2);
int string_wide_print = swprintf(wbuf, 16, L"%d", 3);
int wide_vprint = vwprintf(L"%d", 0);
int file_wide_vprint = vfwprintf(wf, L"%d", 0);
int string_wide_vprint = vswprintf(wbuf, 16, L"%d", 0);
`, Options{})
	if err != nil {
		t.Fatalf("PreprocessSource failed: %v", err)
	}
	for _, name := range []string{"wchar_t", "wint_t", "mbstate_t", "FILE", "wctype_t", "wctrans_t", "fwide", "fputwc", "putwc", "putwchar", "fgetwc", "getwc", "getwchar", "ungetwc", "fputws", "fgetws", "wprintf", "fwprintf", "swprintf", "vwprintf", "vfwprintf", "vswprintf", "mbrlen", "mbrtowc", "wcrtomb", "mbsrtowcs", "wcsrtombs", "iswalnum", "iswalpha", "iswblank", "iswcntrl", "iswdigit", "iswgraph", "iswlower", "iswprint", "iswpunct", "iswspace", "iswupper", "iswxdigit", "towlower", "towupper", "wctype", "iswctype", "wctrans", "towctrans"} {
		if !hasIdentifier(res.Tokens, name) {
			t.Fatalf("wide header identifier %q missing: %#v", name, res.Tokens)
		}
	}
	if !hasLexeme(res.Tokens, "-") || !hasLexeme(res.Tokens, "1") {
		t.Fatalf("WEOF did not expand to -1: %#v", res.Tokens)
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
	for _, name := range []string{"fabs", "fabsf", "fabsl", "sqrt", "sqrtf", "sqrtl", "sin", "sinf", "sinl", "cos", "cosf", "cosl", "tan", "tanf", "tanl", "sinh", "sinhf", "sinhl", "cosh", "coshf", "coshl", "tanh", "tanhf", "tanhl", "asin", "asinf", "asinl", "acos", "acosf", "acosl", "atan", "atanf", "atanl", "asinh", "asinhf", "asinhl", "acosh", "acoshf", "acoshl", "atanh", "atanhf", "atanhl", "cbrt", "cbrtf", "cbrtl", "erf", "erff", "erfl", "erfc", "erfcf", "erfcl", "tgamma", "tgammaf", "tgammal", "lgamma", "lgammaf", "lgammal", "exp", "expf", "expl", "exp2", "exp2f", "exp2l", "expm1", "expm1f", "expm1l", "log", "logf", "logl", "log10", "log10f", "log10l", "log1p", "log1pf", "log1pl", "log2", "log2f", "log2l", "ceil", "ceilf", "ceill", "floor", "floorf", "floorl", "trunc", "truncf", "truncl", "round", "roundf", "roundl", "nearbyint", "nearbyintf", "nearbyintl", "rint", "rintf", "rintl", "logb", "logbf", "logbl", "ilogb", "ilogbf", "ilogbl", "lrint", "lrintf", "lrintl", "lround", "lroundf", "lroundl", "llrint", "llrintf", "llrintl", "llround", "llroundf", "llroundl", "scalbn", "scalbnf", "scalbnl", "scalbln", "scalblnf", "scalblnl", "ldexp", "ldexpf", "ldexpl", "frexp", "frexpf", "frexpl", "modf", "modff", "modfl", "remquo", "remquof", "remquol", "pow", "powf", "powl", "atan2", "atan2f", "atan2l", "hypot", "hypotf", "hypotl", "fdim", "fdimf", "fdiml", "fmax", "fmaxf", "fmaxl", "fmin", "fminf", "fminl", "fmod", "fmodf", "fmodl", "remainder", "remainderf", "remainderl", "copysign", "copysignf", "copysignl", "fma", "fmaf", "fmal", "nextafter", "nextafterf", "nextafterl", "nexttoward", "nexttowardf", "nexttowardl", "nan", "nanf", "nanl"} {
		if !hasIdentifier(res.Tokens, name) {
			t.Fatalf("math identifier %q missing: %#v", name, res.Tokens)
		}
	}
}

func TestBuiltinComplexHeaderDeclaresProjectionSurface(t *testing.T) {
	res, err := PreprocessSource("main.c", `
#include <complex.h>
`, Options{})
	if err != nil {
		t.Fatalf("PreprocessSource failed: %v", err)
	}
	for _, name := range []string{"cabsf", "cabs", "cabsl", "crealf", "creal", "creall", "cimagf", "cimag", "cimagl", "cargf", "carg", "cargl", "conjf", "conj", "conjl", "cprojf", "cproj", "cprojl", "csinf", "csin", "csinl", "ccosf", "ccos", "ccosl", "ctanf", "ctan", "ctanl", "csinhf", "csinh", "csinhl", "ccoshf", "ccosh", "ccoshl", "ctanhf", "ctanh", "ctanhl", "casinf", "casin", "casinl", "cacosf", "cacos", "cacosl", "catanf", "catan", "catanl", "casinhf", "casinh", "casinhl", "cacoshf", "cacosh", "cacoshl", "catanhf", "catanh", "catanhl", "cexpf", "cexp", "cexpl", "clogf", "clog", "clogl", "csqrtf", "csqrt", "csqrtl", "cpowf", "cpow", "cpowl"} {
		if !hasIdentifier(res.Tokens, name) {
			t.Fatalf("complex identifier %q missing: %#v", name, res.Tokens)
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
	for _, name := range []string{"size_t", "strcmp", "memcmp", "strncmp", "strcoll", "strlen", "strnlen", "strerror", "strchr", "strrchr", "strstr", "strpbrk", "strspn", "strcspn", "memchr", "strtok", "strxfrm", "strndup", "memcpy", "memmove", "mempcpy", "memccpy", "memset", "bzero", "strcpy", "stpcpy", "strcat", "strncpy", "stpncpy", "strncat"} {
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
