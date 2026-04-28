package preprocessor

import "fmt"

func builtinHeader(name string, target TargetInfo) (string, bool) {
	switch name {
	case "stdbool.h":
		return "#ifndef __CVM_STDBOOL_H\n#define __CVM_STDBOOL_H\n#define bool _Bool\n#define true 1\n#define false 0\n#define __bool_true_false_are_defined 1\n#endif\n", true
	case "stddef.h":
		return fmt.Sprintf("#ifndef __CVM_STDDEF_H\n#define __CVM_STDDEF_H\n#define __SIZE_TYPE__ %s\n#define __PTRDIFF_TYPE__ %s\ntypedef __SIZE_TYPE__ size_t;\ntypedef __PTRDIFF_TYPE__ ptrdiff_t;\n#define NULL ((void *)0)\n#endif\n", target.SizeType, target.PtrdiffType), true
	case "stdint.h":
		return fmt.Sprintf("#ifndef __CVM_STDINT_H\n#define __CVM_STDINT_H\ntypedef signed char int8_t;\ntypedef short int16_t;\ntypedef int int32_t;\ntypedef long int64_t;\ntypedef unsigned char uint8_t;\ntypedef unsigned short uint16_t;\ntypedef unsigned int uint32_t;\ntypedef unsigned long uint64_t;\ntypedef %s intptr_t;\ntypedef %s uintptr_t;\ntypedef %s intmax_t;\ntypedef %s uintmax_t;\n#define INT8_MAX 127\n#define INT8_MIN (-128)\n#define UINT8_MAX 255\n#define INT16_MAX 32767\n#define INT16_MIN (-32768)\n#define UINT16_MAX 65535\n#define INT32_MAX 2147483647\n#define INT32_MIN (-2147483647-1)\n#define UINT32_MAX 4294967295U\n#define INT64_MAX 9223372036854775807L\n#define INT64_MIN (-9223372036854775807L-1L)\n#define UINT64_MAX 18446744073709551615UL\n#define INTPTR_MAX 9223372036854775807L\n#define UINTPTR_MAX 18446744073709551615UL\n#define INTMAX_MAX 9223372036854775807L\n#define UINTMAX_MAX 18446744073709551615UL\n#define SIZE_MAX 18446744073709551615UL\n#endif\n", target.PtrdiffType, target.SizeType, target.IntmaxType, target.UIntmaxType), true
	case "iso646.h":
		return "#ifndef __CVM_ISO646_H\n#define __CVM_ISO646_H\n#define and &&\n#define and_eq &=\n#define bitand &\n#define bitor |\n#define compl ~\n#define not !\n#define not_eq !=\n#define or ||\n#define or_eq |=\n#define xor ^\n#define xor_eq ^=\n#endif\n", true
	case "math.h":
		return "#ifndef __CVM_MATH_H\n#define __CVM_MATH_H\n#define INFINITY 1e+100000000f\n#endif\n", true
	case "stdio.h":
		return "#ifndef __CVM_STDIO_H\n#define __CVM_STDIO_H\ntypedef struct __cvm_FILE FILE;\nextern FILE *stdin;\nextern FILE *stdout;\nextern FILE *stderr;\nint fputs(const char * restrict, FILE * restrict);\nint fputs_unlocked(const char * restrict, FILE * restrict);\n#endif\n", true
	case "signal.h":
		return "#ifndef __CVM_SIGNAL_H\n#define __CVM_SIGNAL_H\ntypedef int sig_atomic_t;\n#define SIG_ATOMIC_MIN (-2147483647-1)\n#define SIG_ATOMIC_MAX 2147483647\n#endif\n", true
	case "limits.h":
		return "#ifndef __CVM_LIMITS_H\n#define __CVM_LIMITS_H\n#define CHAR_BIT 8\n#define SCHAR_MIN (-128)\n#define SCHAR_MAX 127\n#define UCHAR_MAX 255\n#define SHRT_MIN (-32768)\n#define SHRT_MAX 32767\n#define USHRT_MAX 65535\n#define INT_MIN (-2147483647-1)\n#define INT_MAX 2147483647\n#define UINT_MAX 4294967295U\n#define LONG_MIN (-9223372036854775807L-1L)\n#define LONG_MAX 9223372036854775807L\n#define ULONG_MAX 18446744073709551615UL\n#define LLONG_MIN (-9223372036854775807LL-1LL)\n#define LLONG_MAX 9223372036854775807LL\n#define ULLONG_MAX 18446744073709551615ULL\n#endif\n", true
	case "float.h":
		return "#ifndef __CVM_FLOAT_H\n#define __CVM_FLOAT_H\n#define FLT_ROUNDS 1\n#define FLT_RADIX 2\n#define FLT_MANT_DIG 24\n#define FLT_DIG 6\n#define FLT_MIN_EXP (-125)\n#define FLT_MIN_10_EXP (-37)\n#define FLT_MAX_EXP 128\n#define FLT_MAX_10_EXP 38\n#define FLT_MAX 3.40282346638528859812e+38F\n#define FLT_EPSILON 1.1920928955078125e-7F\n#define FLT_MIN 1.17549435082228750797e-38F\n#define DBL_MANT_DIG 53\n#define DBL_DIG 15\n#define DBL_MIN_EXP (-1021)\n#define DBL_MIN_10_EXP (-307)\n#define DBL_MAX_EXP 1024\n#define DBL_MAX_10_EXP 308\n#define DBL_MAX 1.79769313486231570815e+308\n#define DBL_EPSILON 2.22044604925031308085e-16\n#define DBL_MIN 2.22507385850720138309e-308\n#define LDBL_MANT_DIG 64\n#define LDBL_DIG 18\n#define LDBL_MIN_EXP (-16381)\n#define LDBL_MIN_10_EXP (-4931)\n#define LDBL_MAX_EXP 16384\n#define LDBL_MAX_10_EXP 4932\n#define LDBL_MAX 1.18973149535723176502e+4932L\n#define LDBL_EPSILON 1.08420217248550443401e-19L\n#define LDBL_MIN 3.36210314311209350626e-4932L\n#define FLT_EVAL_METHOD 0\n#define DECIMAL_DIG 21\n#endif\n", true
	default:
		return "", false
	}
}
