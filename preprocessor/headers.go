package preprocessor

import "fmt"

func builtinHeader(name string, target TargetInfo) (string, bool) {
	switch name {
	case "stdbool.h":
		return "#ifndef __CVM_STDBOOL_H\n#define __CVM_STDBOOL_H\n#define bool _Bool\n#define true 1\n#define false 0\n#define __bool_true_false_are_defined 1\n#endif\n", true
	case "stddef.h":
		return fmt.Sprintf("#ifndef __CVM_STDDEF_H\n#define __CVM_STDDEF_H\n#define __SIZE_TYPE__ %s\n#define __PTRDIFF_TYPE__ %s\ntypedef __SIZE_TYPE__ size_t;\ntypedef __PTRDIFF_TYPE__ ptrdiff_t;\n#define NULL ((void *)0)\n#endif\n", target.SizeType, target.PtrdiffType), true
	case "stdarg.h":
		return "#ifndef __CVM_STDARG_H\n#define __CVM_STDARG_H\ntypedef __builtin_va_list va_list;\n#define va_start(ap, last) __builtin_va_start(ap, last)\n#define va_end(ap) __builtin_va_end(ap)\n#define va_arg(ap, type) ((type)0)\n#endif\n", true
	case "stdint.h":
		return stdintHeader(target), true
	case "inttypes.h":
		return "#ifndef __CVM_INTTYPES_H\n#define __CVM_INTTYPES_H\n#include <stdint.h>\n#endif\n", true
	case "iso646.h":
		return "#ifndef __CVM_ISO646_H\n#define __CVM_ISO646_H\n#define and &&\n#define and_eq &=\n#define bitand &\n#define bitor |\n#define compl ~\n#define not !\n#define not_eq !=\n#define or ||\n#define or_eq |=\n#define xor ^\n#define xor_eq ^=\n#endif\n", true
	case "math.h":
		return mathHeader(), true
	case "fenv.h":
		return "#ifndef __CVM_FENV_H\n#define __CVM_FENV_H\n#define FE_ALL_EXCEPT 0\nint feclearexcept(int);\nint fetestexcept(int);\n#endif\n", true
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
	case "string.h":
		return stringHeader(), true
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

func stdioHeader() string {
	return `#ifndef __CVM_STDIO_H
#define __CVM_STDIO_H
typedef __SIZE_TYPE__ size_t;
typedef struct __cvm_FILE FILE;
#define EOF (-1)
#define _IOFBF 0
#define _IOLBF 1
#define _IONBF 2
#define BUFSIZ 8192
extern FILE *stdin;
extern FILE *stdout;
extern FILE *stderr;
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
int printf(const char *, ...);
int printf_unlocked(const char *, ...);
int fprintf(FILE *, const char *, ...);
int fprintf_unlocked(FILE *, const char *, ...);
int sprintf(char *, const char *, ...);
int snprintf(char *, size_t, const char *, ...);
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
typedef __SIZE_TYPE__ size_t;
int abs(int);
long labs(long);
long long llabs(long long);
int atoi(const char *);
long atol(const char *);
long long atoll(const char *);
void *malloc(size_t);
void *calloc(size_t, size_t);
void free(void *);
char *strdup(const char *);
void exit(int);
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

func stringHeader() string {
	return `#ifndef __CVM_STRING_H
#define __CVM_STRING_H
typedef __SIZE_TYPE__ size_t;
int strcmp(const char *, const char *);
int memcmp(const void *, const void *, size_t);
size_t strlen(const char *);
char *strchr(const char *, int);
char *strstr(const char *, const char *);
void *memcpy(void *, const void *, size_t);
void *memmove(void *, const void *, size_t);
void *mempcpy(void *, const void *, size_t);
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

func builtinChkHeader() string {
	return `#ifndef __CVM_BUILTIN_CHK_H
#define __CVM_BUILTIN_CHK_H
typedef __SIZE_TYPE__ size_t;
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
#define signbit(x) __cvm_signbit((double)(x))
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
#include <math.h>
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
