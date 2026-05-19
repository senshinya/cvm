package preprocessor

import "fmt"

func builtinHeader(name string, target TargetInfo) (string, bool) {
	switch name {
	case "stdbool.h":
		return "#ifndef __CVM_STDBOOL_H\n#define __CVM_STDBOOL_H\n#define bool _Bool\n#define true 1\n#define false 0\n#define __bool_true_false_are_defined 1\n#endif\n", true
	case "stddef.h":
		return fmt.Sprintf("#ifndef __CVM_STDDEF_H\n#define __CVM_STDDEF_H\n#define __SIZE_TYPE__ %s\n#define __PTRDIFF_TYPE__ %s\n#ifndef __CVM_SIZE_T\n#define __CVM_SIZE_T\ntypedef __SIZE_TYPE__ size_t;\n#endif\ntypedef __PTRDIFF_TYPE__ ptrdiff_t;\n#define NULL ((void *)0)\n#endif\n", target.SizeType, target.PtrdiffType), true
	case "stdarg.h":
		return "#ifndef __CVM_STDARG_H\n#define __CVM_STDARG_H\ntypedef __builtin_va_list va_list;\n#define va_start(ap, last) __builtin_va_start(ap, last)\n#define va_end(ap) __builtin_va_end(ap)\n#define va_copy(dst, src) __builtin_va_copy(dst, src)\n#define va_arg(ap, type) (*(type *)__builtin_va_arg(ap))\n#endif\n", true
	case "stdint.h":
		return stdintHeader(target), true
	case "inttypes.h":
		return "#ifndef __CVM_INTTYPES_H\n#define __CVM_INTTYPES_H\n#include <stdint.h>\n#endif\n", true
	case "iso646.h":
		return "#ifndef __CVM_ISO646_H\n#define __CVM_ISO646_H\n#define and &&\n#define and_eq &=\n#define bitand &\n#define bitor |\n#define compl ~\n#define not !\n#define not_eq !=\n#define or ||\n#define or_eq |=\n#define xor ^\n#define xor_eq ^=\n#endif\n", true
	case "math.h":
		return mathHeader(), true
	case "complex.h":
		return complexHeader(), true
	case "fenv.h":
		return "#ifndef __CVM_FENV_H\n#define __CVM_FENV_H\n#define FE_ALL_EXCEPT 0\nint feclearexcept(int);\nint fetestexcept(int);\n#endif\n", true
	case "errno.h":
		return "#ifndef __CVM_ERRNO_H\n#define __CVM_ERRNO_H\n#define EDOM 33\n#define ERANGE 34\n#define EILSEQ 84\nextern int errno;\n#endif\n", true
	case "assert.h":
		return "#ifndef __CVM_ASSERT_H\n#define __CVM_ASSERT_H\n#ifdef NDEBUG\n#define assert(expr) ((void)0)\n#else\nvoid abort(void);\n#define assert(expr) ((expr) ? (void)0 : abort())\n#endif\n#endif\n", true
	case "tgmath.h":
		return tgmathHeader(), true
	case "chk.h":
		return builtinChkHeader(), true
	case "builtins-config.h":
		return "#ifndef __CVM_BUILTINS_CONFIG_H\n#define __CVM_BUILTINS_CONFIG_H\n#define HAVE_C99_RUNTIME 1\n#endif\n", true
	case "sys/types.h":
		return "#ifndef __CVM_SYS_TYPES_H\n#define __CVM_SYS_TYPES_H\n#endif\n", true
	case "stdio.h":
		return stdioHeader(), true
	case "stdlib.h":
		return stdlibHeader(), true
	case "ctype.h":
		return ctypeHeader(), true
	case "locale.h":
		return localeHeader(), true
	case "wchar.h":
		return wcharHeader(), true
	case "wctype.h":
		return wctypeHeader(), true
	case "time.h":
		return timeHeader(), true
	case "string.h":
		return stringHeader(), true
	case "strings.h":
		return stringsHeader(), true
	case "signal.h":
		return "#ifndef __CVM_SIGNAL_H\n#define __CVM_SIGNAL_H\ntypedef int sig_atomic_t;\n#define SIG_ATOMIC_MIN (-2147483647-1)\n#define SIG_ATOMIC_MAX 2147483647\n#endif\n", true
	case "limits.h":
		return "#ifndef __CVM_LIMITS_H\n#define __CVM_LIMITS_H\n#define CHAR_BIT 8\n#define SCHAR_MIN (-128)\n#define SCHAR_MAX 127\n#define UCHAR_MAX 255\n#define SHRT_MIN (-32768)\n#define SHRT_MAX 32767\n#define USHRT_MAX 65535\n#define INT_MIN (-2147483647-1)\n#define INT_MAX 2147483647\n#define UINT_MAX 4294967295U\n#define LONG_MIN (-9223372036854775807L-1L)\n#define LONG_MAX 9223372036854775807L\n#define ULONG_MAX 18446744073709551615UL\n#define LLONG_MIN (-9223372036854775807LL-1LL)\n#define LLONG_MAX 9223372036854775807LL\n#define ULLONG_MAX 18446744073709551615ULL\n#endif\n", true
	case "float.h":
		return "#ifndef __CVM_FLOAT_H\n#define __CVM_FLOAT_H\n#define FLT_ROUNDS 1\n#define FLT_RADIX 2\n#define FLT_MANT_DIG 24\n#define FLT_DIG 6\n#define FLT_MIN_EXP (-125)\n#define FLT_MIN_10_EXP (-37)\n#define FLT_MAX_EXP 128\n#define FLT_MAX_10_EXP 38\n#define FLT_MAX 3.40282346638528859812e+38F\n#define FLT_EPSILON 1.1920928955078125e-7F\n#define FLT_MIN 1.17549435082228750797e-38F\n#define DBL_MANT_DIG 53\n#define DBL_DIG 15\n#define DBL_MIN_EXP (-1021)\n#define DBL_MIN_10_EXP (-307)\n#define DBL_MAX_EXP 1024\n#define DBL_MAX_10_EXP 308\n#define DBL_MAX 1.79769313486231570815e+308\n#define DBL_EPSILON 2.22044604925031308085e-16\n#define DBL_MIN 2.22507385850720138309e-308\n#define LDBL_MANT_DIG 53\n#define LDBL_DIG 15\n#define LDBL_MIN_EXP (-1021)\n#define LDBL_MIN_10_EXP (-307)\n#define LDBL_MAX_EXP 1024\n#define LDBL_MAX_10_EXP 308\n#define LDBL_MAX 1.79769313486231570815e+308L\n#define LDBL_EPSILON 2.22044604925031308085e-16L\n#define LDBL_MIN 2.22507385850720138309e-308L\n#define FLT_EVAL_METHOD 0\n#define DECIMAL_DIG 17\n#endif\n", true
	default:
		return "", false
	}
}

func complexHeader() string {
	return `#ifndef __CVM_COMPLEX_H
#define __CVM_COMPLEX_H
#ifndef complex
#define complex _Complex
#endif
float cabsf(float complex);
double cabs(double complex);
long double cabsl(long double complex);
float crealf(float complex);
double creal(double complex);
long double creall(long double complex);
float cimagf(float complex);
double cimag(double complex);
long double cimagl(long double complex);
float cargf(float complex);
double carg(double complex);
long double cargl(long double complex);
float complex conjf(float complex);
double complex conj(double complex);
long double complex conjl(long double complex);
float complex cprojf(float complex);
double complex cproj(double complex);
long double complex cprojl(long double complex);
float complex csinf(float complex);
double complex csin(double complex);
long double complex csinl(long double complex);
float complex ccosf(float complex);
double complex ccos(double complex);
long double complex ccosl(long double complex);
float complex ctanf(float complex);
double complex ctan(double complex);
long double complex ctanl(long double complex);
float complex csinhf(float complex);
double complex csinh(double complex);
long double complex csinhl(long double complex);
float complex ccoshf(float complex);
double complex ccosh(double complex);
long double complex ccoshl(long double complex);
float complex ctanhf(float complex);
double complex ctanh(double complex);
long double complex ctanhl(long double complex);
float complex casinf(float complex);
double complex casin(double complex);
long double complex casinl(long double complex);
float complex cacosf(float complex);
double complex cacos(double complex);
long double complex cacosl(long double complex);
float complex catanf(float complex);
double complex catan(double complex);
long double complex catanl(long double complex);
float complex casinhf(float complex);
double complex casinh(double complex);
long double complex casinhl(long double complex);
float complex cacoshf(float complex);
double complex cacosh(double complex);
long double complex cacoshl(long double complex);
float complex catanhf(float complex);
double complex catanh(double complex);
long double complex catanhl(long double complex);
float complex cexpf(float complex);
double complex cexp(double complex);
long double complex cexpl(long double complex);
float complex clogf(float complex);
double complex clog(double complex);
long double complex clogl(long double complex);
float complex csqrtf(float complex);
double complex csqrt(double complex);
long double complex csqrtl(long double complex);
float complex cpowf(float complex, float complex);
double complex cpow(double complex, double complex);
long double complex cpowl(long double complex, long double complex);
#endif
`
}

func stdioHeader() string {
	return `#ifndef __CVM_STDIO_H
#define __CVM_STDIO_H
#ifndef __CVM_SIZE_T
#define __CVM_SIZE_T
typedef __SIZE_TYPE__ size_t;
#endif
typedef struct __cvm_FILE FILE;
typedef long fpos_t;
#define EOF (-1)
#define _IOFBF 0
#define _IOLBF 1
#define _IONBF 2
#define BUFSIZ 8192
#define L_tmpnam 20
#define TMP_MAX 25
#define SEEK_SET 0
#define SEEK_CUR 1
#define SEEK_END 2
extern FILE *stdin;
extern FILE *stdout;
extern FILE *stderr;
int remove(const char *);
int rename(const char *, const char *);
FILE *fopen(const char * restrict, const char * restrict);
FILE *freopen(const char * restrict, const char * restrict, FILE * restrict);
FILE *tmpfile(void);
char *tmpnam(char *);
int fseek(FILE *, long, int);
long ftell(FILE *);
void rewind(FILE *);
int fgetpos(FILE * restrict, fpos_t * restrict);
int fsetpos(FILE *, const fpos_t *);
int fputs(const char * restrict, FILE * restrict);
int fputs_unlocked(const char * restrict, FILE * restrict);
int fputc(int, FILE *);
int fputc_unlocked(int, FILE *);
int putc(int, FILE *);
int putc_unlocked(int, FILE *);
int puts(const char *);
int puts_unlocked(const char *);
int putchar(int);
int putchar_unlocked(int);
int getchar(void);
int getchar_unlocked(void);
int fflush(FILE *);
int fflush_unlocked(FILE *);
int fclose(FILE *);
int fileno(FILE *);
int fileno_unlocked(FILE *);
void setbuf(FILE * restrict, char * restrict);
int setvbuf(FILE * restrict, char * restrict, int, size_t);
void flockfile(FILE *);
int ftrylockfile(FILE *);
void funlockfile(FILE *);
void perror(const char *);
int ferror(FILE *);
int ferror_unlocked(FILE *);
void clearerr(FILE *);
void clearerr_unlocked(FILE *);
int feof(FILE *);
int feof_unlocked(FILE *);
size_t fwrite(const void *, size_t, size_t, FILE *);
size_t fwrite_unlocked(const void *, size_t, size_t, FILE *);
size_t fread(void *, size_t, size_t, FILE *);
size_t fread_unlocked(void *, size_t, size_t, FILE *);
int fgetc(FILE *);
int fgetc_unlocked(FILE *);
int getc(FILE *);
int getc_unlocked(FILE *);
int ungetc(int, FILE *);
char *fgets(char * restrict, int, FILE * restrict);
char *fgets_unlocked(char * restrict, int, FILE * restrict);
int scanf(const char *, ...);
int fscanf(FILE *, const char *, ...);
int printf(const char *, ...);
int printf_unlocked(const char *, ...);
int fprintf(FILE *, const char *, ...);
int fprintf_unlocked(FILE *, const char *, ...);
int sprintf(char *, const char *, ...);
int snprintf(char *, size_t, const char *, ...);
int sscanf(const char *, const char *, ...);
int vprintf(const char *, void *);
int vprintf_unlocked(const char *, void *);
int vfprintf(FILE *, const char *, void *);
int vfprintf_unlocked(FILE *, const char *, void *);
int vsprintf(char *, const char *, void *);
int vsnprintf(char *, size_t, const char *, void *);
#endif
`
}

func stdlibHeader() string {
	return `#ifndef __CVM_STDLIB_H
#define __CVM_STDLIB_H
#ifndef __CVM_SIZE_T
#define __CVM_SIZE_T
typedef __SIZE_TYPE__ size_t;
#endif
#ifndef __CVM_WCHAR_T
#define __CVM_WCHAR_T
typedef __WCHAR_TYPE__ wchar_t;
#endif
typedef struct { int quot; int rem; } div_t;
typedef struct { long quot; long rem; } ldiv_t;
typedef struct { long long quot; long long rem; } lldiv_t;
#define RAND_MAX 32767
int abs(int);
long labs(long);
long long llabs(long long);
div_t div(int, int);
ldiv_t ldiv(long, long);
lldiv_t lldiv(long long, long long);
int atoi(const char *);
long atol(const char *);
long long atoll(const char *);
double atof(const char *);
long strtol(const char *, char **, int);
unsigned long strtoul(const char *, char **, int);
long long strtoll(const char *, char **, int);
unsigned long long strtoull(const char *, char **, int);
double strtod(const char *, char **);
float strtof(const char *, char **);
long double strtold(const char *, char **);
int mblen(const char *, size_t);
int mbtowc(wchar_t *, const char *, size_t);
int wctomb(char *, wchar_t);
size_t mbstowcs(wchar_t *, const char *, size_t);
size_t wcstombs(char *, const wchar_t *, size_t);
void *malloc(size_t);
void *calloc(size_t, size_t);
void *realloc(void *, size_t);
void free(void *);
char *strdup(const char *);
int rand(void);
void srand(unsigned int);
char *getenv(const char *);
int system(const char *);
int atexit(void (*)(void));
void exit(int);
void _Exit(int);
void abort(void);
#endif
`
}

func ctypeHeader() string {
	return `#ifndef __CVM_CTYPE_H
#define __CVM_CTYPE_H
int isdigit(int);
int isalpha(int);
int isalnum(int);
int isspace(int);
int islower(int);
int isupper(int);
int isxdigit(int);
int isprint(int);
int isblank(int);
int iscntrl(int);
int isgraph(int);
int ispunct(int);
int tolower(int);
int toupper(int);
#endif
`
}

func localeHeader() string {
	return `#ifndef __CVM_LOCALE_H
#define __CVM_LOCALE_H
#define LC_ALL 0
#define LC_COLLATE 1
#define LC_CTYPE 2
#define LC_MONETARY 3
#define LC_NUMERIC 4
#define LC_TIME 5
struct lconv {
  char *decimal_point;
  char *thousands_sep;
  char *grouping;
  char *int_curr_symbol;
  char *currency_symbol;
  char *mon_decimal_point;
  char *mon_thousands_sep;
  char *mon_grouping;
  char *positive_sign;
  char *negative_sign;
  char int_frac_digits;
  char frac_digits;
  char p_cs_precedes;
  char p_sep_by_space;
  char n_cs_precedes;
  char n_sep_by_space;
  char p_sign_posn;
  char n_sign_posn;
  char int_p_cs_precedes;
  char int_p_sep_by_space;
  char int_n_cs_precedes;
  char int_n_sep_by_space;
  char int_p_sign_posn;
  char int_n_sign_posn;
};
char *setlocale(int, const char *);
struct lconv *localeconv(void);
#endif
`
}

func wcharHeader() string {
	return `#ifndef __CVM_WCHAR_H
#define __CVM_WCHAR_H
#include <stddef.h>
#ifndef __CVM_WCHAR_T
#define __CVM_WCHAR_T
typedef __WCHAR_TYPE__ wchar_t;
#endif
#ifndef __CVM_WINT_T
#define __CVM_WINT_T
typedef int wint_t;
#endif
#ifndef __CVM_MBSTATE_T
#define __CVM_MBSTATE_T
typedef struct { unsigned int __count; unsigned int __value; } mbstate_t;
#endif
#define WEOF (-1)
size_t wcslen(const wchar_t *);
int wcscmp(const wchar_t *, const wchar_t *);
int wcsncmp(const wchar_t *, const wchar_t *, size_t);
wchar_t *wmemchr(const wchar_t *, wchar_t, size_t);
int wmemcmp(const wchar_t *, const wchar_t *, size_t);
wchar_t *wmemcpy(wchar_t * restrict, const wchar_t * restrict, size_t);
wchar_t *wmemmove(wchar_t *, const wchar_t *, size_t);
wchar_t *wmemset(wchar_t *, wchar_t, size_t);
size_t mbrlen(const char *, size_t, mbstate_t *);
size_t mbrtowc(wchar_t *, const char *, size_t, mbstate_t *);
size_t wcrtomb(char *, wchar_t, mbstate_t *);
size_t mbsrtowcs(wchar_t *, const char **, size_t, mbstate_t *);
size_t wcsrtombs(char *, const wchar_t **, size_t, mbstate_t *);
#endif
`
}

func wctypeHeader() string {
	return `#ifndef __CVM_WCTYPE_H
#define __CVM_WCTYPE_H
#include <wchar.h>
#ifndef __CVM_WCTYPE_T
#define __CVM_WCTYPE_T
typedef unsigned long wctype_t;
#endif
#ifndef __CVM_WCTRANS_T
#define __CVM_WCTRANS_T
typedef unsigned long wctrans_t;
#endif
int iswalnum(wint_t);
int iswalpha(wint_t);
int iswblank(wint_t);
int iswcntrl(wint_t);
int iswdigit(wint_t);
int iswgraph(wint_t);
int iswlower(wint_t);
int iswprint(wint_t);
int iswpunct(wint_t);
int iswspace(wint_t);
int iswupper(wint_t);
int iswxdigit(wint_t);
wint_t towlower(wint_t);
wint_t towupper(wint_t);
wctype_t wctype(const char *);
int iswctype(wint_t, wctype_t);
wctrans_t wctrans(const char *);
wint_t towctrans(wint_t, wctrans_t);
#endif
`
}

func timeHeader() string {
	return `#ifndef __CVM_TIME_H
#define __CVM_TIME_H
typedef long clock_t;
typedef long time_t;
#define CLOCKS_PER_SEC 1000000L
clock_t clock(void);
double difftime(time_t, time_t);
time_t time(time_t *);
#endif
`
}

func stringHeader() string {
	return `#ifndef __CVM_STRING_H
#define __CVM_STRING_H
#ifndef __CVM_SIZE_T
#define __CVM_SIZE_T
typedef __SIZE_TYPE__ size_t;
#endif
int strcmp(const char *, const char *);
int memcmp(const void *, const void *, size_t);
int strncmp(const char *, const char *, size_t);
int strcoll(const char *, const char *);
size_t strlen(const char *);
size_t strnlen(const char *, size_t);
char *strerror(int);
char *strchr(const char *, int);
char *strrchr(const char *, int);
char *strstr(const char *, const char *);
char *strpbrk(const char *, const char *);
size_t strspn(const char *, const char *);
size_t strcspn(const char *, const char *);
void *memchr(const void *, int, size_t);
char *strtok(char *, const char *);
size_t strxfrm(char * restrict, const char * restrict, size_t);
char *strndup(const char *, size_t);
void *memcpy(void *, const void *, size_t);
void *memmove(void *, const void *, size_t);
void *mempcpy(void *, const void *, size_t);
void *memccpy(void *, const void *, int, size_t);
void *memset(void *, int, size_t);
void bzero(void *, size_t);
char *strcpy(char *, const char *);
char *stpcpy(char *, const char *);
char *strcat(char *, const char *);
char *strncpy(char *, const char *, size_t);
char *stpncpy(char *, const char *, size_t);
char *strncat(char *, const char *, size_t);
#endif
`
}

func stringsHeader() string {
	return `#ifndef __CVM_STRINGS_H
#define __CVM_STRINGS_H
#ifndef __CVM_SIZE_T
#define __CVM_SIZE_T
typedef __SIZE_TYPE__ size_t;
#endif
int bcmp(const void *, const void *, size_t);
void bcopy(const void *, void *, size_t);
void bzero(void *, size_t);
#endif
`
}

func builtinChkHeader() string {
	return `#ifndef __CVM_BUILTIN_CHK_H
#define __CVM_BUILTIN_CHK_H
#ifndef __CVM_SIZE_T
#define __CVM_SIZE_T
typedef __SIZE_TYPE__ size_t;
#endif
void *memcpy(void *, const void *, size_t);
void *mempcpy(void *, const void *, size_t);
void *memmove(void *, const void *, size_t);
void *memset(void *, int, size_t);
char *strcpy(char *, const char *);
char *stpcpy(char *, const char *);
char *strncpy(char *, const char *, size_t);
char *strcat(char *, const char *);
char *strncat(char *, const char *, size_t);
int sprintf(char *, const char *, ...);
int snprintf(char *, size_t, const char *, ...);
int vsprintf(char *, const char *, void *);
int vsnprintf(char *, size_t, const char *, void *);
#endif
`
}

func stdintHeader(target TargetInfo) string {
	return fmt.Sprintf(`#ifndef __CVM_STDINT_H
#define __CVM_STDINT_H
typedef signed char int8_t;
typedef short int16_t;
typedef int int32_t;
typedef long int64_t;
typedef unsigned char uint8_t;
typedef unsigned short uint16_t;
typedef unsigned int uint32_t;
typedef unsigned long uint64_t;
typedef signed char int_least8_t;
typedef short int_least16_t;
typedef int int_least32_t;
typedef long int_least64_t;
typedef unsigned char uint_least8_t;
typedef unsigned short uint_least16_t;
typedef unsigned int uint_least32_t;
typedef unsigned long uint_least64_t;
typedef signed char int_fast8_t;
typedef long int_fast16_t;
typedef long int_fast32_t;
typedef long int_fast64_t;
typedef unsigned char uint_fast8_t;
typedef unsigned long uint_fast16_t;
typedef unsigned long uint_fast32_t;
typedef unsigned long uint_fast64_t;
typedef %s intptr_t;
typedef %s uintptr_t;
typedef %s intmax_t;
typedef %s uintmax_t;
#define INT8_MAX 127
#define INT8_MIN (-127 - 1)
#define UINT8_MAX 255
#define INT16_MAX 32767
#define INT16_MIN (-32767 - 1)
#define UINT16_MAX 65535
#define INT32_MAX 2147483647
#define INT32_MIN (-2147483647 - 1)
#define UINT32_MAX 4294967295U
#define INT64_MAX 9223372036854775807L
#define INT64_MIN (-9223372036854775807L - 1L)
#define UINT64_MAX 18446744073709551615UL
#define INT_LEAST8_MIN INT8_MIN
#define INT_LEAST8_MAX INT8_MAX
#define UINT_LEAST8_MAX UINT8_MAX
#define INT_LEAST16_MIN INT16_MIN
#define INT_LEAST16_MAX INT16_MAX
#define UINT_LEAST16_MAX UINT16_MAX
#define INT_LEAST32_MIN INT32_MIN
#define INT_LEAST32_MAX INT32_MAX
#define UINT_LEAST32_MAX UINT32_MAX
#define INT_LEAST64_MIN INT64_MIN
#define INT_LEAST64_MAX INT64_MAX
#define UINT_LEAST64_MAX UINT64_MAX
#define INT_FAST8_MIN INT8_MIN
#define INT_FAST8_MAX INT8_MAX
#define UINT_FAST8_MAX UINT8_MAX
#define INT_FAST16_MIN INT64_MIN
#define INT_FAST16_MAX INT64_MAX
#define UINT_FAST16_MAX UINT64_MAX
#define INT_FAST32_MIN INT64_MIN
#define INT_FAST32_MAX INT64_MAX
#define UINT_FAST32_MAX UINT64_MAX
#define INT_FAST64_MIN INT64_MIN
#define INT_FAST64_MAX INT64_MAX
#define UINT_FAST64_MAX UINT64_MAX
#define INTPTR_MIN (-9223372036854775807L - 1L)
#define INTPTR_MAX 9223372036854775807L
#define UINTPTR_MAX 18446744073709551615UL
#define INTMAX_MIN (-9223372036854775807L - 1L)
#define INTMAX_MAX 9223372036854775807L
#define UINTMAX_MAX 18446744073709551615UL
#define PTRDIFF_MIN (-9223372036854775807L - 1L)
#define PTRDIFF_MAX 9223372036854775807L
#define SIG_ATOMIC_MIN (-2147483647 - 1)
#define SIG_ATOMIC_MAX 2147483647
#define SIZE_MAX 18446744073709551615UL
#define WCHAR_MIN (-2147483647 - 1)
#define WCHAR_MAX 2147483647
#define WINT_MIN 0U
#define WINT_MAX 4294967295U
#define INT8_C(c) c
#define INT16_C(c) c
#define INT32_C(c) c
#define INT64_C(c) c ## L
#define UINT8_C(c) c ## U
#define UINT16_C(c) c ## U
#define UINT32_C(c) c ## U
#define UINT64_C(c) c ## UL
#define INTMAX_C(c) c ## L
#define UINTMAX_C(c) c ## UL
#endif
`, target.PtrdiffType, target.SizeType, target.IntmaxType, target.UIntmaxType)
}

func mathHeader() string {
	return `#ifndef __CVM_MATH_H
#define __CVM_MATH_H
#define INFINITY 1e+100000000f
#define NAN (0.0/0.0)
#define HUGE_VAL 1e+100000000
#define HUGE_VALF 1e+100000000f
#define HUGE_VALL 1e+100000000L
#define FP_NAN 0
#define FP_INFINITE 1
#define FP_NORMAL 2
#define FP_SUBNORMAL 3
#define FP_ZERO 4
#ifndef __CVM_TGMATH_INCLUDE
float nanf(const char *);
double nan(const char *);
long double nanl(const char *);
float fabsf(float);
double fabs(double);
long double fabsl(long double);
float sqrtf(float);
double sqrt(double);
long double sqrtl(long double);
float sinf(float);
double sin(double);
long double sinl(long double);
float cosf(float);
double cos(double);
long double cosl(long double);
float tanf(float);
double tan(double);
long double tanl(long double);
float sinhf(float);
double sinh(double);
long double sinhl(long double);
float coshf(float);
double cosh(double);
long double coshl(long double);
float tanhf(float);
double tanh(double);
long double tanhl(long double);
float asinf(float);
double asin(double);
long double asinl(long double);
float acosf(float);
double acos(double);
long double acosl(long double);
float atanf(float);
double atan(double);
long double atanl(long double);
float asinhf(float);
double asinh(double);
long double asinhl(long double);
float acoshf(float);
double acosh(double);
long double acoshl(long double);
float atanhf(float);
double atanh(double);
long double atanhl(long double);
float cbrtf(float);
double cbrt(double);
long double cbrtl(long double);
float erff(float);
double erf(double);
long double erfl(long double);
float erfcf(float);
double erfc(double);
long double erfcl(long double);
float tgammaf(float);
double tgamma(double);
long double tgammal(long double);
float lgammaf(float);
double lgamma(double);
long double lgammal(long double);
float expf(float);
double exp(double);
long double expl(long double);
float exp2f(float);
double exp2(double);
long double exp2l(long double);
float expm1f(float);
double expm1(double);
long double expm1l(long double);
float logf(float);
double log(double);
long double logl(long double);
float log10f(float);
double log10(double);
long double log10l(long double);
float log1pf(float);
double log1p(double);
long double log1pl(long double);
float log2f(float);
double log2(double);
long double log2l(long double);
float ceilf(float);
double ceil(double);
long double ceill(long double);
float floorf(float);
double floor(double);
long double floorl(long double);
float truncf(float);
double trunc(double);
long double truncl(long double);
float roundf(float);
double round(double);
long double roundl(long double);
float nearbyintf(float);
double nearbyint(double);
long double nearbyintl(long double);
float rintf(float);
double rint(double);
long double rintl(long double);
float logbf(float);
double logb(double);
long double logbl(long double);
int ilogbf(float);
int ilogb(double);
int ilogbl(long double);
long lrintf(float);
long lrint(double);
long lrintl(long double);
long lroundf(float);
long lround(double);
long lroundl(long double);
long long llrintf(float);
long long llrint(double);
long long llrintl(long double);
long long llroundf(float);
long long llround(double);
long long llroundl(long double);
float scalbnf(float, int);
double scalbn(double, int);
long double scalbnl(long double, int);
float scalblnf(float, long);
double scalbln(double, long);
long double scalblnl(long double, long);
float ldexpf(float, int);
double ldexp(double, int);
long double ldexpl(long double, int);
float frexpf(float, int *);
double frexp(double, int *);
long double frexpl(long double, int *);
float modff(float, float *);
double modf(double, double *);
long double modfl(long double, long double *);
float remquof(float, float, int *);
double remquo(double, double, int *);
long double remquol(long double, long double, int *);
float powf(float, float);
double pow(double, double);
long double powl(long double, long double);
float atan2f(float, float);
double atan2(double, double);
long double atan2l(long double, long double);
float hypotf(float, float);
double hypot(double, double);
long double hypotl(long double, long double);
float fdimf(float, float);
double fdim(double, double);
long double fdiml(long double, long double);
float fmaxf(float, float);
double fmax(double, double);
long double fmaxl(long double, long double);
float fminf(float, float);
double fmin(double, double);
long double fminl(long double, long double);
float fmodf(float, float);
double fmod(double, double);
long double fmodl(long double, long double);
float remainderf(float, float);
double remainder(double, double);
long double remainderl(long double, long double);
float copysignf(float, float);
double copysign(double, double);
long double copysignl(long double, long double);
float fmaf(float, float, float);
double fma(double, double, double);
long double fmal(long double, long double, long double);
float nextafterf(float, float);
double nextafter(double, double);
long double nextafterl(long double, long double);
float nexttowardf(float, long double);
double nexttoward(double, long double);
long double nexttowardl(long double, long double);
#endif
int __cvm_fpclassifyf(float);
int __cvm_fpclassify(double);
int __cvm_fpclassifyl(long double);
int __cvm_isfinitef(float);
int __cvm_isfinite(double);
int __cvm_isfinitel(long double);
int __cvm_isinff(float);
int __cvm_isinf(double);
int __cvm_isinfl(long double);
int __cvm_isnanf(float);
int __cvm_isnan(double);
int __cvm_isnanl(long double);
int __cvm_isnormalf(float);
int __cvm_isnormal(double);
int __cvm_isnormall(long double);
int __cvm_signbitf(float);
int __cvm_signbit(double);
int __cvm_signbitl(long double);
int __cvm_isunordered(double, double);
#define __cvm_math_select1(x, f, d, l) ((sizeof(x) == sizeof(float)) ? f(x) : ((sizeof(x) == sizeof(long double)) ? l(x) : d(x)))
#define fpclassify(x) __cvm_math_select1((x), __cvm_fpclassifyf, __cvm_fpclassify, __cvm_fpclassifyl)
#define isfinite(x) __cvm_math_select1((x), __cvm_isfinitef, __cvm_isfinite, __cvm_isfinitel)
#define isinf(x) __cvm_math_select1((x), __cvm_isinff, __cvm_isinf, __cvm_isinfl)
#define isnan(x) __cvm_math_select1((x), __cvm_isnanf, __cvm_isnan, __cvm_isnanl)
#define isnormal(x) __cvm_math_select1((x), __cvm_isnormalf, __cvm_isnormal, __cvm_isnormall)
#define signbit(x) __cvm_math_select1((x), __cvm_signbitf, __cvm_signbit, __cvm_signbitl)
#define isgreater(x, y) ((x) > (y))
#define isgreaterequal(x, y) ((x) >= (y))
#define isless(x, y) ((x) < (y))
#define islessequal(x, y) ((x) <= (y))
#define islessgreater(x, y) (!__cvm_isunordered((x), (y)) && ((x) != (y)))
#define isunordered(x, y) __cvm_isunordered((x), (y))
#endif
`
}

func tgmathHeader() string {
	return `#ifndef __CVM_TGMATH_H
#define __CVM_TGMATH_H
#define __CVM_TGMATH_INCLUDE 1
#include <math.h>
#undef __CVM_TGMATH_INCLUDE
#ifndef complex
#define complex _Complex
#endif
#define acos(x) __cvm_tgmath_acos(x)
#define asin(x) __cvm_tgmath_asin(x)
#define atan(x) __cvm_tgmath_atan(x)
#define acosh(x) __cvm_tgmath_acosh(x)
#define asinh(x) __cvm_tgmath_asinh(x)
#define atanh(x) __cvm_tgmath_atanh(x)
#define cos(x) __cvm_tgmath_cos(x)
#define sin(x) __cvm_tgmath_sin((x))
#define tan(x) __cvm_tgmath_tan(x)
#define cosh(x) __cvm_tgmath_cosh(x)
#define sinh(x) __cvm_tgmath_sinh(x)
#define tanh(x) __cvm_tgmath_tanh(x)
#define exp(x) __cvm_tgmath_exp((x))
#define log(x) __cvm_tgmath_log(x)
#define pow(x, y) __cvm_tgmath_pow((x), (y))
#define sqrt(x) __cvm_tgmath_sqrt(x)
#define fabs(x) __cvm_tgmath_fabs(x)
#define atan2(x, y) __cvm_tgmath_atan2((x), (y))
#define cbrt(x) __cvm_tgmath_cbrt(x)
#define ceil(x) __cvm_tgmath_ceil(x)
#define copysign(x, y) __cvm_tgmath_copysign((x), (y))
#define erf(x) __cvm_tgmath_erf(x)
#define erfc(x) __cvm_tgmath_erfc(x)
#define exp2(x) __cvm_tgmath_exp2(x)
#define expm1(x) __cvm_tgmath_expm1(x)
#define fdim(x, y) __cvm_tgmath_fdim((x), (y))
#define floor(x) __cvm_tgmath_floor(x)
#define fma(x, y, z) __cvm_tgmath_fma((x), (y), (z))
#define fmax(x, y) __cvm_tgmath_fmax((x), (y))
#define fmin(x, y) __cvm_tgmath_fmin((x), (y))
#define fmod(x, y) __cvm_tgmath_fmod((x), (y))
#define frexp(x, y) __cvm_tgmath_frexp((x), (y))
#define modf(x, y) __cvm_tgmath_modf((x), (y))
#define hypot(x, y) __cvm_tgmath_hypot((x), (y))
#define ilogb(x) __cvm_tgmath_ilogb(x)
#define ldexp(x, y) __cvm_tgmath_ldexp((x), (y))
#define lgamma(x) __cvm_tgmath_lgamma(x)
#define llrint(x) __cvm_tgmath_llrint(x)
#define llround(x) __cvm_tgmath_llround(x)
#define log10(x) __cvm_tgmath_log10(x)
#define log1p(x) __cvm_tgmath_log1p(x)
#define log2(x) __cvm_tgmath_log2(x)
#define logb(x) __cvm_tgmath_logb(x)
#define lrint(x) __cvm_tgmath_lrint(x)
#define lround(x) __cvm_tgmath_lround(x)
#define nearbyint(x) __cvm_tgmath_nearbyint(x)
#define nextafter(x, y) __cvm_tgmath_nextafter((x), (y))
#define nexttoward(x, y) __cvm_tgmath_nexttoward((x), (y))
#define remainder(x, y) __cvm_tgmath_remainder((x), (y))
#define remquo(x, y, z) __cvm_tgmath_remquo((x), (y), (z))
#define rint(x) __cvm_tgmath_rint(x)
#define round(x) __cvm_tgmath_round(x)
#define scalbn(x, y) __cvm_tgmath_scalbn((x), (y))
#define scalbln(x, y) __cvm_tgmath_scalbln((x), (y))
#define tgamma(x) __cvm_tgmath_tgamma(x)
#define trunc(x) __cvm_tgmath_trunc(x)
#define carg(x) __cvm_tgmath_carg(x)
#define cimag(x) __cvm_tgmath_cimag(x)
#define conj(x) __cvm_tgmath_conj(x)
#define cproj(x) __cvm_tgmath_cproj(x)
#define creal(x) __cvm_tgmath_creal(x)
double __cvm_tgmath_sin();
double __cvm_tgmath_exp();
double __cvm_tgmath_pow();
double __cvm_tgmath_sqrt();
double __cvm_tgmath_fabs();
double __cvm_tgmath_cos();
double __cvm_tgmath_tan();
double __cvm_tgmath_log();
double __cvm_tgmath_sinh();
double __cvm_tgmath_cosh();
double __cvm_tgmath_tanh();
double __cvm_tgmath_asin();
double __cvm_tgmath_acos();
double __cvm_tgmath_atan();
double __cvm_tgmath_asinh();
double __cvm_tgmath_acosh();
double __cvm_tgmath_atanh();
double __cvm_tgmath_atan2();
double __cvm_tgmath_hypot();
double __cvm_tgmath_cbrt();
double __cvm_tgmath_ceil();
double __cvm_tgmath_floor();
double __cvm_tgmath_trunc();
double __cvm_tgmath_round();
double __cvm_tgmath_exp2();
double __cvm_tgmath_expm1();
double __cvm_tgmath_log10();
double __cvm_tgmath_log1p();
double __cvm_tgmath_log2();
double __cvm_tgmath_fdim();
double __cvm_tgmath_fmax();
double __cvm_tgmath_fmin();
double __cvm_tgmath_fmod();
double __cvm_tgmath_remainder();
double __cvm_tgmath_copysign();
double __cvm_tgmath_fma();
double __cvm_tgmath_nextafter();
double __cvm_tgmath_nexttoward();
double __cvm_tgmath_erf();
double __cvm_tgmath_erfc();
double __cvm_tgmath_tgamma();
double __cvm_tgmath_lgamma();
double __cvm_tgmath_nearbyint();
double __cvm_tgmath_rint();
double __cvm_tgmath_logb();
double __cvm_tgmath_scalbn();
double __cvm_tgmath_scalbln();
double __cvm_tgmath_ldexp();
int __cvm_tgmath_ilogb();
double __cvm_tgmath_frexp();
double __cvm_tgmath_modf();
double __cvm_tgmath_remquo();
double __cvm_tgmath_carg();
double __cvm_tgmath_cimag();
double __cvm_tgmath_creal();
double __cvm_tgmath_conj();
double __cvm_tgmath_cproj();
long __cvm_tgmath_lrint();
long __cvm_tgmath_lround();
long long __cvm_tgmath_llrint();
long long __cvm_tgmath_llround();
#endif
`
}
