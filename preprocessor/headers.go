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
	case "limits.h":
		return "#ifndef __CVM_LIMITS_H\n#define __CVM_LIMITS_H\n#define CHAR_BIT 8\n#define SCHAR_MIN (-128)\n#define SCHAR_MAX 127\n#define UCHAR_MAX 255\n#define SHRT_MIN (-32768)\n#define SHRT_MAX 32767\n#define USHRT_MAX 65535\n#define INT_MIN (-2147483647-1)\n#define INT_MAX 2147483647\n#define UINT_MAX 4294967295U\n#define LONG_MIN (-9223372036854775807L-1L)\n#define LONG_MAX 9223372036854775807L\n#define ULONG_MAX 18446744073709551615UL\n#endif\n", true
	case "float.h":
		return "#ifndef __CVM_FLOAT_H\n#define __CVM_FLOAT_H\n#define FLT_RADIX 2\n#define FLT_MANT_DIG 24\n#define DBL_MANT_DIG 53\n#define LDBL_MANT_DIG 64\n#define FLT_MAX 3.40282346638528859812e+38F\n#define DBL_MAX 1.79769313486231570815e+308\n#define LDBL_MAX 1.18973149535723176502e+4932L\n#endif\n", true
	default:
		return "", false
	}
}
