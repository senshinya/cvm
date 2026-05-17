package runtime

import (
	"context"
	"fmt"
	"io"
	"math"
	"math/cmplx"
	"os"
	"strings"

	"shinya.click/cvm/bytecode"
)

type ExternFunc func(context.Context, *ExternContext, []Value) (Value, *ExitStatus, error)

type ExternContext struct {
	Memory *Memory
	Stdout io.Writer
	Stderr io.Writer
}

type ExternRegistry struct {
	funcs       map[string]ExternFunc
	stdout      io.Writer
	stderr      io.Writer
	hostWriters map[uint64]io.Writer
}

func NewExternRegistry(stdout, stderr io.Writer) *ExternRegistry {
	if stdout == nil {
		stdout = os.Stdout
	}
	if stderr == nil {
		stderr = os.Stderr
	}
	return &ExternRegistry{
		funcs:       make(map[string]ExternFunc),
		stdout:      stdout,
		stderr:      stderr,
		hostWriters: make(map[uint64]io.Writer),
	}
}

func DefaultExternRegistry(stdout, stderr io.Writer) *ExternRegistry {
	r := NewExternRegistry(stdout, stderr)
	r.Register("exit", func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 1 {
			return Value{}, nil, fmt.Errorf("exit expects 1 argument")
		}
		code, err := args[0].ExitCode()
		if err != nil {
			return Value{}, nil, err
		}
		return Value{}, &ExitStatus{Code: code}, nil
	})
	r.Register("abort", abortExtern())
	r.Register("__builtin_abort", abortExtern())
	r.Register("puts", func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 1 {
			return Value{}, nil, fmt.Errorf("puts expects 1 argument")
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("puts requires memory")
		}
		s, err := ec.Memory.ReadCString(args[0].Int)
		if err != nil {
			return Value{}, nil, err
		}
		if _, err := fmt.Fprintln(r.externStdout(ec), s); err != nil {
			return Value{}, nil, err
		}
		return IntValue(bytecode.TypeI32, int64(len(s)+1)), nil, nil
	})
	r.Register("fputs", func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 2 {
			return Value{}, nil, fmt.Errorf("fputs expects 2 arguments")
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("fputs requires memory")
		}
		s, err := ec.Memory.ReadCString(args[0].Int)
		if err != nil {
			return Value{}, nil, err
		}
		w, ok := r.lookupHostWriter(args[1].Int)
		if !ok {
			return Value{}, nil, fmt.Errorf("unknown stream handle %#x", args[1].Int)
		}
		if _, err := fmt.Fprint(w, s); err != nil {
			return Value{}, nil, err
		}
		return IntValue(bytecode.TypeI32, int64(len(s))), nil, nil
	})
	r.Register("strcmp", func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 2 {
			return Value{}, nil, fmt.Errorf("strcmp expects 2 arguments")
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("strcmp requires memory")
		}
		left, err := ec.Memory.ReadCString(args[0].Int)
		if err != nil {
			return Value{}, nil, err
		}
		right, err := ec.Memory.ReadCString(args[1].Int)
		if err != nil {
			return Value{}, nil, err
		}
		return IntValue(bytecode.TypeI32, int64(strcmpResult(left, right))), nil, nil
	})
	r.Register("memcmp", func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 3 {
			return Value{}, nil, fmt.Errorf("memcmp expects 3 arguments")
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("memcmp requires memory")
		}
		n := unsignedInt(args[2])
		if n == 0 {
			return IntValue(bytecode.TypeI32, 0), nil, nil
		}
		if n > uint64(maxInt()) {
			return Value{}, nil, fmt.Errorf("memcmp size %d exceeds int range", n)
		}
		leftBlock, leftOff, err := ec.Memory.rangeAccess(args[0].Int, int64(n), false)
		if err != nil {
			return Value{}, nil, err
		}
		rightBlock, rightOff, err := ec.Memory.rangeAccess(args[1].Int, int64(n), false)
		if err != nil {
			return Value{}, nil, err
		}
		left := leftBlock.data[leftOff : leftOff+int(n)]
		right := rightBlock.data[rightOff : rightOff+int(n)]
		return IntValue(bytecode.TypeI32, int64(memcmpResult(left, right))), nil, nil
	})
	registerAllocationExterns(r)
	registerMemoryExterns(r)
	r.Register("feclearexcept", func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 1 {
			return Value{}, nil, fmt.Errorf("feclearexcept expects 1 argument")
		}
		return IntValue(bytecode.TypeI32, 0), nil, nil
	})
	r.Register("fetestexcept", func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 1 {
			return Value{}, nil, fmt.Errorf("fetestexcept expects 1 argument")
		}
		return IntValue(bytecode.TypeI32, 0), nil, nil
	})
	registerMathExterns(r)
	return r
}

func registerAllocationExterns(r *ExternRegistry) {
	r.Register("__builtin_malloc", mallocExtern("__builtin_malloc"))
	r.Register("__builtin_calloc", callocExtern("__builtin_calloc"))
	r.Register("__builtin_strdup", strdupExtern("__builtin_strdup"))
	r.Register("__builtin_object_size", objectSizeExtern("__builtin_object_size"))
	r.Register("__builtin_dynamic_object_size", objectSizeExtern("__builtin_dynamic_object_size"))
}

func registerMemoryExterns(r *ExternRegistry) {
	for _, name := range []string{"__builtin_memcpy", "memcpy", "__builtin_memmove", "memmove"} {
		r.Register(name, memoryCopyExtern(name, false))
	}
	for _, name := range []string{"__builtin_mempcpy", "mempcpy"} {
		r.Register(name, memoryCopyExtern(name, true))
	}
	for _, name := range []string{"__builtin_memset", "memset"} {
		r.Register(name, memorySetExtern(name))
	}
	r.Register("__builtin_bzero", memoryBzeroExtern("__builtin_bzero"))
	for _, name := range []string{"__builtin___memcpy_chk", "__builtin___memmove_chk"} {
		r.Register(name, memoryCheckedCopyExtern(name, false))
	}
	r.Register("__builtin___mempcpy_chk", memoryCheckedCopyExtern("__builtin___mempcpy_chk", true))
	r.Register("__builtin___memset_chk", memoryCheckedSetExtern("__builtin___memset_chk"))
	for _, name := range []string{"__builtin_strlen", "strlen"} {
		r.Register(name, stringLengthExtern(name))
	}
	for _, name := range []string{"__builtin_strchr", "strchr"} {
		r.Register(name, stringCharSearchExtern(name))
	}
	for _, name := range []string{"__builtin_strstr", "strstr"} {
		r.Register(name, stringSearchExtern(name))
	}
	for _, name := range []string{"__builtin_strcpy", "strcpy"} {
		r.Register(name, stringCopyExtern(name, false))
	}
	for _, name := range []string{"__builtin_stpcpy", "stpcpy"} {
		r.Register(name, stringCopyExtern(name, true))
	}
	for _, name := range []string{"__builtin_strcat", "strcat"} {
		r.Register(name, stringConcatExtern(name))
	}
	for _, name := range []string{"__builtin_strncpy", "strncpy"} {
		r.Register(name, stringNCopyExtern(name, false))
	}
	for _, name := range []string{"__builtin_stpncpy", "stpncpy"} {
		r.Register(name, stringNCopyExtern(name, true))
	}
	for _, name := range []string{"__builtin_strncat", "strncat"} {
		r.Register(name, stringNConcatExtern(name))
	}
	for _, name := range []string{"__builtin___strcpy_chk", "__builtin___stpcpy_chk"} {
		r.Register(name, stringCheckedCopyExtern(name, name == "__builtin___stpcpy_chk"))
	}
	r.Register("__builtin___strcat_chk", stringCheckedConcatExtern("__builtin___strcat_chk"))
	for _, name := range []string{"__builtin___strncpy_chk", "__builtin___stpncpy_chk"} {
		r.Register(name, stringCheckedNCopyExtern(name, name == "__builtin___stpncpy_chk"))
	}
	r.Register("__builtin___strncat_chk", stringCheckedNConcatExtern("__builtin___strncat_chk"))
}

func registerMathExterns(r *ExternRegistry) {
	for _, suffix := range []string{"f", "", "l"} {
		r.Register("__cvm_fpclassify"+suffix, mathUnaryExtern("__cvm_fpclassify"+suffix, func(v Value) int64 {
			return cvmFPClassify(v)
		}))
		r.Register("__cvm_isfinite"+suffix, mathUnaryExtern("__cvm_isfinite"+suffix, func(v Value) int64 {
			class := cvmFPClassify(v)
			return boolInt(class != fpClassNaN && class != fpClassInfinite)
		}))
		r.Register("__cvm_isinf"+suffix, mathUnaryExtern("__cvm_isinf"+suffix, func(v Value) int64 {
			return boolInt(cvmFPClassify(v) == fpClassInfinite)
		}))
		r.Register("__cvm_isnan"+suffix, mathUnaryExtern("__cvm_isnan"+suffix, func(v Value) int64 {
			return boolInt(cvmFPClassify(v) == fpClassNaN)
		}))
		r.Register("__cvm_isnormal"+suffix, mathUnaryExtern("__cvm_isnormal"+suffix, func(v Value) int64 {
			return boolInt(cvmFPClassify(v) == fpClassNormal)
		}))
		r.Register("__cvm_signbit"+suffix, mathUnaryExtern("__cvm_signbit"+suffix, func(v Value) int64 {
			return boolInt(math.Signbit(cvmFloat(v)))
		}))
	}
	registerTgmathRealExterns(r, "__cvm_tgmath_sin", math.Sin)
	registerTgmathRealExterns(r, "__cvm_tgmath_exp", math.Exp)
	registerTgmathRealExterns(r, "__cvm_tgmath_sqrt", math.Sqrt)
	registerTgmathRealExterns(r, "__cvm_tgmath_fabs", math.Abs)
	registerTgmathRealExterns(r, "__cvm_tgmath_cos", math.Cos)
	registerTgmathRealExterns(r, "__cvm_tgmath_tan", math.Tan)
	registerTgmathRealExterns(r, "__cvm_tgmath_log", math.Log)
	registerTgmathRealExterns(r, "__cvm_tgmath_sinh", math.Sinh)
	registerTgmathRealExterns(r, "__cvm_tgmath_cosh", math.Cosh)
	registerTgmathRealExterns(r, "__cvm_tgmath_tanh", math.Tanh)
	registerTgmathRealExterns(r, "__cvm_tgmath_asin", math.Asin)
	registerTgmathRealExterns(r, "__cvm_tgmath_acos", math.Acos)
	registerTgmathRealExterns(r, "__cvm_tgmath_atan", math.Atan)
	registerTgmathRealExterns(r, "__cvm_tgmath_asinh", math.Asinh)
	registerTgmathRealExterns(r, "__cvm_tgmath_acosh", math.Acosh)
	registerTgmathRealExterns(r, "__cvm_tgmath_atanh", math.Atanh)
	registerTgmathRealExterns(r, "__cvm_tgmath_cbrt", math.Cbrt)
	registerTgmathRealExterns(r, "__cvm_tgmath_ceil", math.Ceil)
	registerTgmathRealExterns(r, "__cvm_tgmath_floor", math.Floor)
	registerTgmathRealExterns(r, "__cvm_tgmath_trunc", math.Trunc)
	registerTgmathRealExterns(r, "__cvm_tgmath_round", math.Round)
	registerTgmathRealExterns(r, "__cvm_tgmath_exp2", math.Exp2)
	registerTgmathRealExterns(r, "__cvm_tgmath_expm1", math.Expm1)
	registerTgmathRealExterns(r, "__cvm_tgmath_log10", math.Log10)
	registerTgmathRealExterns(r, "__cvm_tgmath_log1p", math.Log1p)
	registerTgmathRealExterns(r, "__cvm_tgmath_log2", math.Log2)
	registerTgmathRealExterns(r, "__cvm_tgmath_erf", math.Erf)
	registerTgmathRealExterns(r, "__cvm_tgmath_erfc", math.Erfc)
	registerTgmathRealExterns(r, "__cvm_tgmath_tgamma", math.Gamma)
	registerTgmathRealExterns(r, "__cvm_tgmath_lgamma", func(v float64) float64 {
		r, _ := math.Lgamma(v)
		return r
	})
	registerTgmathRealExterns(r, "__cvm_tgmath_nearbyint", math.RoundToEven)
	registerTgmathRealExterns(r, "__cvm_tgmath_rint", math.RoundToEven)
	registerTgmathRealExterns(r, "__cvm_tgmath_logb", math.Logb)
	registerTgmathIntExterns(r, "__cvm_tgmath_ilogb", math.Ilogb)
	registerTgmathLongExterns(r, "__cvm_tgmath_lrint", math.RoundToEven)
	registerTgmathLongExterns(r, "__cvm_tgmath_lround", math.Round)
	registerTgmathLongExterns(r, "__cvm_tgmath_llrint", math.RoundToEven)
	registerTgmathLongExterns(r, "__cvm_tgmath_llround", math.Round)
	registerTgmathRealBinaryExterns(r, "__cvm_tgmath_pow", math.Pow)
	r.Register("__builtin_pow", mathBinaryFloatExtern("__builtin_pow", bytecode.TypeF64, math.Pow))
	r.Register("__builtin_huge_valf", mathConstantFloatExtern("__builtin_huge_valf", bytecode.TypeF32, math.Inf(1)))
	r.Register("__builtin_huge_val", mathConstantFloatExtern("__builtin_huge_val", bytecode.TypeF64, math.Inf(1)))
	r.Register("__builtin_huge_vall", mathConstantFloatExtern("__builtin_huge_vall", bytecode.TypeFLong, math.Inf(1)))
	r.Register("__builtin_nan", mathNanExtern("__builtin_nan"))
	r.Register("nan", mathNanExtern("nan"))
	registerTgmathRealBinaryExterns(r, "__cvm_tgmath_atan2", math.Atan2)
	registerTgmathRealBinaryExterns(r, "__cvm_tgmath_hypot", math.Hypot)
	registerTgmathRealBinaryExterns(r, "__cvm_tgmath_fdim", math.Dim)
	registerTgmathRealBinaryExterns(r, "__cvm_tgmath_fmax", math.Max)
	registerTgmathRealBinaryExterns(r, "__cvm_tgmath_fmin", math.Min)
	registerTgmathRealBinaryExterns(r, "__cvm_tgmath_fmod", math.Mod)
	registerTgmathRealBinaryExterns(r, "__cvm_tgmath_remainder", math.Remainder)
	registerTgmathRealBinaryExterns(r, "__cvm_tgmath_copysign", math.Copysign)
	registerTgmathRealBinaryExterns(r, "__cvm_tgmath_nextafter", math.Nextafter)
	registerTgmathRealBinaryExterns(r, "__cvm_tgmath_nexttoward", math.Nextafter)
	registerTgmathRealIntBinaryExterns(r, "__cvm_tgmath_scalbn", math.Ldexp)
	registerTgmathRealIntBinaryExterns(r, "__cvm_tgmath_scalbln", math.Ldexp)
	registerTgmathRealIntBinaryExterns(r, "__cvm_tgmath_ldexp", math.Ldexp)
	registerTgmathFrexpExterns(r, "__cvm_tgmath_frexp")
	registerTgmathRemquoExterns(r, "__cvm_tgmath_remquo")
	registerTgmathComplexRealExterns(r, "__cvm_tgmath_carg", func(z complex128) float64 { return math.Atan2(imag(z), real(z)) })
	registerTgmathComplexRealExterns(r, "__cvm_tgmath_cimag", func(z complex128) float64 { return imag(z) })
	registerTgmathComplexRealExterns(r, "__cvm_tgmath_creal", func(z complex128) float64 { return real(z) })
	registerTgmathRealTernaryExterns(r, "__cvm_tgmath_fma", math.FMA)
	registerTgmathComplexExterns(r, "__cvm_tgmath_conj", cmplx.Conj)
	registerTgmathComplexExterns(r, "__cvm_tgmath_cproj", cvmComplexProject)
	registerTgmathComplexExterns(r, "__cvm_tgmath_csin", cmplx.Sin)
	registerTgmathComplexExterns(r, "__cvm_tgmath_cexp", cmplx.Exp)
	registerTgmathComplexExterns(r, "__cvm_tgmath_csqrt", cmplx.Sqrt)
	registerTgmathComplexExterns(r, "__cvm_tgmath_ccos", cmplx.Cos)
	registerTgmathComplexExterns(r, "__cvm_tgmath_ctan", cmplx.Tan)
	registerTgmathComplexExterns(r, "__cvm_tgmath_clog", cmplx.Log)
	registerTgmathComplexExterns(r, "__cvm_tgmath_csinh", cmplx.Sinh)
	registerTgmathComplexExterns(r, "__cvm_tgmath_ccosh", cmplx.Cosh)
	registerTgmathComplexExterns(r, "__cvm_tgmath_ctanh", cmplx.Tanh)
	registerTgmathComplexExterns(r, "__cvm_tgmath_casin", cmplx.Asin)
	registerTgmathComplexExterns(r, "__cvm_tgmath_cacos", cmplx.Acos)
	registerTgmathComplexExterns(r, "__cvm_tgmath_catan", cmplx.Atan)
	registerTgmathComplexExterns(r, "__cvm_tgmath_casinh", cmplx.Asinh)
	registerTgmathComplexExterns(r, "__cvm_tgmath_cacosh", cmplx.Acosh)
	registerTgmathComplexExterns(r, "__cvm_tgmath_catanh", cmplx.Atanh)
	registerTgmathComplexBinaryExterns(r, "__cvm_tgmath_cpow", cmplx.Pow)
	r.Register("__builtin_cabsf", complexAbsExtern("__builtin_cabsf", bytecode.TypeF32, 4))
	r.Register("__builtin_cabs", complexAbsExtern("__builtin_cabs", bytecode.TypeF64, 8))
	r.Register("__builtin_cabsl", complexAbsExtern("__builtin_cabsl", bytecode.TypeFLong, 16))
	r.Register("__cvm_isunordered", func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 2 {
			return Value{}, nil, fmt.Errorf("__cvm_isunordered expects 2 arguments")
		}
		return IntValue(bytecode.TypeI32, boolInt(math.IsNaN(cvmFloat(args[0])) || math.IsNaN(cvmFloat(args[1])))), nil, nil
	})
}

func abortExtern() ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		return Value{}, nil, &TrapError{Reason: "abort"}
	}
}

func mallocExtern(name string) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 1 {
			return Value{}, nil, fmt.Errorf("%s expects 1 argument", name)
		}
		if !isIntegerLike(args[0].Type) {
			return Value{}, nil, fmt.Errorf("%s expects size argument", name)
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("%s requires memory", name)
		}
		size, err := memorySizeArg(name, args[0])
		if err != nil {
			return Value{}, nil, err
		}
		addr, err := ec.Memory.TryAlloc("extern:"+name, nonzeroAllocSize(size), ec.Memory.target.PointerAlign, false, blockGlobal)
		if err != nil {
			return Value{}, nil, err
		}
		return PtrValue(addr), nil, nil
	}
}

func callocExtern(name string) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 2 {
			return Value{}, nil, fmt.Errorf("%s expects 2 arguments", name)
		}
		if !isIntegerLike(args[0].Type) || !isIntegerLike(args[1].Type) {
			return Value{}, nil, fmt.Errorf("%s expects count and size arguments", name)
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("%s requires memory", name)
		}
		count := unsignedInt(args[0])
		size := unsignedInt(args[1])
		if count != 0 && size > uint64(maxInt())/count {
			return Value{}, nil, fmt.Errorf("%s allocation size overflows", name)
		}
		total := count * size
		if total > uint64(maxInt()) {
			return Value{}, nil, fmt.Errorf("%s allocation size %d exceeds int range", name, total)
		}
		addr, err := ec.Memory.TryAlloc("extern:"+name, nonzeroAllocSize(int64(total)), ec.Memory.target.PointerAlign, false, blockGlobal)
		if err != nil {
			return Value{}, nil, err
		}
		return PtrValue(addr), nil, nil
	}
}

func strdupExtern(name string) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 1 {
			return Value{}, nil, fmt.Errorf("%s expects 1 argument", name)
		}
		if !isPointerType(args[0].Type) {
			return Value{}, nil, fmt.Errorf("%s expects string argument", name)
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("%s requires memory", name)
		}
		src, err := ec.Memory.ReadCString(args[0].Int)
		if err != nil {
			return Value{}, nil, err
		}
		data := append([]byte(src), 0)
		addr, err := ec.Memory.TryAlloc("extern:"+name, int64(len(data)), 1, false, blockGlobal)
		if err != nil {
			return Value{}, nil, err
		}
		if err := writeMemoryBytes(ec.Memory, addr, data); err != nil {
			return Value{}, nil, err
		}
		return PtrValue(addr), nil, nil
	}
}

func nonzeroAllocSize(size int64) int64 {
	if size == 0 {
		return 1
	}
	return size
}

func objectSizeExtern(name string) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 2 {
			return Value{}, nil, fmt.Errorf("%s expects 2 arguments", name)
		}
		if !isPointerType(args[0].Type) || !isIntegerLike(args[1].Type) {
			return Value{}, nil, fmt.Errorf("%s expects pointer and type arguments", name)
		}
		switch signedInt(args[1]) & 3 {
		case 2, 3:
			return UIntValue(bytecode.TypeU64, 0), nil, nil
		default:
			return UIntValue(bytecode.TypeU64, ^uint64(0)), nil, nil
		}
	}
}

func memoryCopyExtern(name string, returnEnd bool) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 3 {
			return Value{}, nil, fmt.Errorf("%s expects 3 arguments", name)
		}
		if !isPointerType(args[0].Type) || !isPointerType(args[1].Type) || !isIntegerLike(args[2].Type) {
			return Value{}, nil, fmt.Errorf("%s expects pointer, pointer, and size arguments", name)
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("%s requires memory", name)
		}
		size, err := memorySizeArg(name, args[2])
		if err != nil {
			return Value{}, nil, err
		}
		if err := ec.Memory.Copy(args[0].Int, args[1].Int, size); err != nil {
			return Value{}, nil, err
		}
		addr := args[0].Int
		if returnEnd {
			addr += uint64(size)
		}
		return PtrValue(addr), nil, nil
	}
}

func memorySetExtern(name string) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 3 {
			return Value{}, nil, fmt.Errorf("%s expects 3 arguments", name)
		}
		if !isPointerType(args[0].Type) || !isIntegerLike(args[1].Type) || !isIntegerLike(args[2].Type) {
			return Value{}, nil, fmt.Errorf("%s expects pointer, integer, and size arguments", name)
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("%s requires memory", name)
		}
		size, err := memorySizeArg(name, args[2])
		if err != nil {
			return Value{}, nil, err
		}
		if err := ec.Memory.Set(args[0].Int, byte(unsignedInt(args[1])), size); err != nil {
			return Value{}, nil, err
		}
		return PtrValue(args[0].Int), nil, nil
	}
}

func memoryBzeroExtern(name string) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 2 {
			return Value{}, nil, fmt.Errorf("%s expects 2 arguments", name)
		}
		if !isPointerType(args[0].Type) || !isIntegerLike(args[1].Type) {
			return Value{}, nil, fmt.Errorf("%s expects pointer and size arguments", name)
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("%s requires memory", name)
		}
		size, err := memorySizeArg(name, args[1])
		if err != nil {
			return Value{}, nil, err
		}
		if err := ec.Memory.Set(args[0].Int, 0, size); err != nil {
			return Value{}, nil, err
		}
		return Value{}, nil, nil
	}
}

func memoryCheckedCopyExtern(name string, returnEnd bool) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 4 {
			return Value{}, nil, fmt.Errorf("%s expects 4 arguments", name)
		}
		if !isPointerType(args[0].Type) || !isPointerType(args[1].Type) || !isIntegerLike(args[2].Type) || !isIntegerLike(args[3].Type) {
			return Value{}, nil, fmt.Errorf("%s expects pointer, pointer, size, and object-size arguments", name)
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("%s requires memory", name)
		}
		size, err := memorySizeArg(name, args[2])
		if err != nil {
			return Value{}, nil, err
		}
		if err := checkObjectSize(name, uint64(size), args[3]); err != nil {
			return Value{}, nil, err
		}
		if err := ec.Memory.Copy(args[0].Int, args[1].Int, size); err != nil {
			return Value{}, nil, err
		}
		addr := args[0].Int
		if returnEnd {
			addr += uint64(size)
		}
		return PtrValue(addr), nil, nil
	}
}

func memoryCheckedSetExtern(name string) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 4 {
			return Value{}, nil, fmt.Errorf("%s expects 4 arguments", name)
		}
		if !isPointerType(args[0].Type) || !isIntegerLike(args[1].Type) || !isIntegerLike(args[2].Type) || !isIntegerLike(args[3].Type) {
			return Value{}, nil, fmt.Errorf("%s expects pointer, integer, size, and object-size arguments", name)
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("%s requires memory", name)
		}
		size, err := memorySizeArg(name, args[2])
		if err != nil {
			return Value{}, nil, err
		}
		if err := checkObjectSize(name, uint64(size), args[3]); err != nil {
			return Value{}, nil, err
		}
		if err := ec.Memory.Set(args[0].Int, byte(unsignedInt(args[1])), size); err != nil {
			return Value{}, nil, err
		}
		return PtrValue(args[0].Int), nil, nil
	}
}

func memorySizeArg(name string, arg Value) (int64, error) {
	size := unsignedInt(arg)
	if size > uint64(maxInt()) {
		return 0, fmt.Errorf("%s size %d exceeds int range", name, size)
	}
	return int64(size), nil
}

func checkObjectSize(name string, size uint64, objectSize Value) error {
	limit := unsignedInt(objectSize)
	if limit != ^uint64(0) && size > limit {
		return &TrapError{Reason: fmt.Sprintf("%s object size %d is smaller than operation size %d", name, limit, size)}
	}
	return nil
}

func stringLengthExtern(name string) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 1 {
			return Value{}, nil, fmt.Errorf("%s expects 1 argument", name)
		}
		if !isPointerType(args[0].Type) {
			return Value{}, nil, fmt.Errorf("%s expects string argument", name)
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("%s requires memory", name)
		}
		s, err := ec.Memory.ReadCString(args[0].Int)
		if err != nil {
			return Value{}, nil, err
		}
		return UIntValue(bytecode.TypeU64, uint64(len(s))), nil, nil
	}
}

func stringCharSearchExtern(name string) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 2 {
			return Value{}, nil, fmt.Errorf("%s expects 2 arguments", name)
		}
		if !isPointerType(args[0].Type) || !isIntegerLike(args[1].Type) {
			return Value{}, nil, fmt.Errorf("%s expects string and integer arguments", name)
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("%s requires memory", name)
		}
		s, err := ec.Memory.ReadCString(args[0].Int)
		if err != nil {
			return Value{}, nil, err
		}
		ch := byte(unsignedInt(args[1]))
		if ch == 0 {
			return PtrValue(args[0].Int + uint64(len(s))), nil, nil
		}
		idx := strings.IndexByte(s, ch)
		if idx < 0 {
			return PtrValue(0), nil, nil
		}
		return PtrValue(args[0].Int + uint64(idx)), nil, nil
	}
}

func stringSearchExtern(name string) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 2 {
			return Value{}, nil, fmt.Errorf("%s expects 2 arguments", name)
		}
		if !isPointerType(args[0].Type) || !isPointerType(args[1].Type) {
			return Value{}, nil, fmt.Errorf("%s expects string arguments", name)
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("%s requires memory", name)
		}
		haystack, err := ec.Memory.ReadCString(args[0].Int)
		if err != nil {
			return Value{}, nil, err
		}
		needle, err := ec.Memory.ReadCString(args[1].Int)
		if err != nil {
			return Value{}, nil, err
		}
		idx := strings.Index(haystack, needle)
		if idx < 0 {
			return PtrValue(0), nil, nil
		}
		return PtrValue(args[0].Int + uint64(idx)), nil, nil
	}
}

func stringCopyExtern(name string, returnEnd bool) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 2 {
			return Value{}, nil, fmt.Errorf("%s expects 2 arguments", name)
		}
		if !isPointerType(args[0].Type) || !isPointerType(args[1].Type) {
			return Value{}, nil, fmt.Errorf("%s expects string arguments", name)
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("%s requires memory", name)
		}
		src, err := ec.Memory.ReadCString(args[1].Int)
		if err != nil {
			return Value{}, nil, err
		}
		data := append([]byte(src), 0)
		if err := writeMemoryBytes(ec.Memory, args[0].Int, data); err != nil {
			return Value{}, nil, err
		}
		addr := args[0].Int
		if returnEnd {
			addr, err = addSignedOffset(args[0].Int, int64(len(src)))
			if err != nil {
				return Value{}, nil, err
			}
		}
		return PtrValue(addr), nil, nil
	}
}

func stringConcatExtern(name string) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 2 {
			return Value{}, nil, fmt.Errorf("%s expects 2 arguments", name)
		}
		if !isPointerType(args[0].Type) || !isPointerType(args[1].Type) {
			return Value{}, nil, fmt.Errorf("%s expects string arguments", name)
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("%s requires memory", name)
		}
		dst, err := ec.Memory.ReadCString(args[0].Int)
		if err != nil {
			return Value{}, nil, err
		}
		src, err := ec.Memory.ReadCString(args[1].Int)
		if err != nil {
			return Value{}, nil, err
		}
		start, err := addSignedOffset(args[0].Int, int64(len(dst)))
		if err != nil {
			return Value{}, nil, err
		}
		if err := writeMemoryBytes(ec.Memory, start, append([]byte(src), 0)); err != nil {
			return Value{}, nil, err
		}
		return PtrValue(args[0].Int), nil, nil
	}
}

func stringNCopyExtern(name string, returnEnd bool) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 3 {
			return Value{}, nil, fmt.Errorf("%s expects 3 arguments", name)
		}
		if !isPointerType(args[0].Type) || !isPointerType(args[1].Type) || !isIntegerLike(args[2].Type) {
			return Value{}, nil, fmt.Errorf("%s expects string, string, and size arguments", name)
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("%s requires memory", name)
		}
		size, err := memorySizeArg(name, args[2])
		if err != nil {
			return Value{}, nil, err
		}
		src, err := ec.Memory.ReadCString(args[1].Int)
		if err != nil {
			return Value{}, nil, err
		}
		data := make([]byte, int(size))
		copied := copy(data, []byte(src))
		if err := writeMemoryBytes(ec.Memory, args[0].Int, data); err != nil {
			return Value{}, nil, err
		}
		addr := args[0].Int
		if returnEnd {
			offset := size
			if copied < int(size) {
				offset = int64(copied)
			}
			addr, err = addSignedOffset(args[0].Int, offset)
			if err != nil {
				return Value{}, nil, err
			}
		}
		return PtrValue(addr), nil, nil
	}
}

func stringNConcatExtern(name string) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 3 {
			return Value{}, nil, fmt.Errorf("%s expects 3 arguments", name)
		}
		if !isPointerType(args[0].Type) || !isPointerType(args[1].Type) || !isIntegerLike(args[2].Type) {
			return Value{}, nil, fmt.Errorf("%s expects string, string, and size arguments", name)
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("%s requires memory", name)
		}
		size, err := memorySizeArg(name, args[2])
		if err != nil {
			return Value{}, nil, err
		}
		dst, err := ec.Memory.ReadCString(args[0].Int)
		if err != nil {
			return Value{}, nil, err
		}
		src, err := ec.Memory.ReadCString(args[1].Int)
		if err != nil {
			return Value{}, nil, err
		}
		count := len(src)
		if int64(count) > size {
			count = int(size)
		}
		start, err := addSignedOffset(args[0].Int, int64(len(dst)))
		if err != nil {
			return Value{}, nil, err
		}
		data := append([]byte(src[:count]), 0)
		if err := writeMemoryBytes(ec.Memory, start, data); err != nil {
			return Value{}, nil, err
		}
		return PtrValue(args[0].Int), nil, nil
	}
}

func stringCheckedCopyExtern(name string, returnEnd bool) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 3 {
			return Value{}, nil, fmt.Errorf("%s expects 3 arguments", name)
		}
		if !isPointerType(args[0].Type) || !isPointerType(args[1].Type) || !isIntegerLike(args[2].Type) {
			return Value{}, nil, fmt.Errorf("%s expects string, string, and object-size arguments", name)
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("%s requires memory", name)
		}
		src, err := ec.Memory.ReadCString(args[1].Int)
		if err != nil {
			return Value{}, nil, err
		}
		if err := checkObjectSize(name, uint64(len(src)+1), args[2]); err != nil {
			return Value{}, nil, err
		}
		if err := writeMemoryBytes(ec.Memory, args[0].Int, append([]byte(src), 0)); err != nil {
			return Value{}, nil, err
		}
		addr := args[0].Int
		if returnEnd {
			addr, err = addSignedOffset(args[0].Int, int64(len(src)))
			if err != nil {
				return Value{}, nil, err
			}
		}
		return PtrValue(addr), nil, nil
	}
}

func stringCheckedConcatExtern(name string) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 3 {
			return Value{}, nil, fmt.Errorf("%s expects 3 arguments", name)
		}
		if !isPointerType(args[0].Type) || !isPointerType(args[1].Type) || !isIntegerLike(args[2].Type) {
			return Value{}, nil, fmt.Errorf("%s expects string, string, and object-size arguments", name)
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("%s requires memory", name)
		}
		dst, err := ec.Memory.ReadCString(args[0].Int)
		if err != nil {
			return Value{}, nil, err
		}
		src, err := ec.Memory.ReadCString(args[1].Int)
		if err != nil {
			return Value{}, nil, err
		}
		if err := checkObjectSize(name, uint64(len(dst)+len(src)+1), args[2]); err != nil {
			return Value{}, nil, err
		}
		start, err := addSignedOffset(args[0].Int, int64(len(dst)))
		if err != nil {
			return Value{}, nil, err
		}
		if err := writeMemoryBytes(ec.Memory, start, append([]byte(src), 0)); err != nil {
			return Value{}, nil, err
		}
		return PtrValue(args[0].Int), nil, nil
	}
}

func stringCheckedNCopyExtern(name string, returnEnd bool) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 4 {
			return Value{}, nil, fmt.Errorf("%s expects 4 arguments", name)
		}
		if !isPointerType(args[0].Type) || !isPointerType(args[1].Type) || !isIntegerLike(args[2].Type) || !isIntegerLike(args[3].Type) {
			return Value{}, nil, fmt.Errorf("%s expects string, string, size, and object-size arguments", name)
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("%s requires memory", name)
		}
		size, err := memorySizeArg(name, args[2])
		if err != nil {
			return Value{}, nil, err
		}
		if err := checkObjectSize(name, uint64(size), args[3]); err != nil {
			return Value{}, nil, err
		}
		src, err := ec.Memory.ReadCString(args[1].Int)
		if err != nil {
			return Value{}, nil, err
		}
		data := make([]byte, int(size))
		copied := copy(data, []byte(src))
		if err := writeMemoryBytes(ec.Memory, args[0].Int, data); err != nil {
			return Value{}, nil, err
		}
		addr := args[0].Int
		if returnEnd {
			offset := size
			if copied < int(size) {
				offset = int64(copied)
			}
			addr, err = addSignedOffset(args[0].Int, offset)
			if err != nil {
				return Value{}, nil, err
			}
		}
		return PtrValue(addr), nil, nil
	}
}

func stringCheckedNConcatExtern(name string) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 4 {
			return Value{}, nil, fmt.Errorf("%s expects 4 arguments", name)
		}
		if !isPointerType(args[0].Type) || !isPointerType(args[1].Type) || !isIntegerLike(args[2].Type) || !isIntegerLike(args[3].Type) {
			return Value{}, nil, fmt.Errorf("%s expects string, string, size, and object-size arguments", name)
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("%s requires memory", name)
		}
		size, err := memorySizeArg(name, args[2])
		if err != nil {
			return Value{}, nil, err
		}
		dst, err := ec.Memory.ReadCString(args[0].Int)
		if err != nil {
			return Value{}, nil, err
		}
		src, err := ec.Memory.ReadCString(args[1].Int)
		if err != nil {
			return Value{}, nil, err
		}
		count := len(src)
		if int64(count) > size {
			count = int(size)
		}
		if err := checkObjectSize(name, uint64(len(dst)+count+1), args[3]); err != nil {
			return Value{}, nil, err
		}
		start, err := addSignedOffset(args[0].Int, int64(len(dst)))
		if err != nil {
			return Value{}, nil, err
		}
		if err := writeMemoryBytes(ec.Memory, start, append([]byte(src[:count]), 0)); err != nil {
			return Value{}, nil, err
		}
		return PtrValue(args[0].Int), nil, nil
	}
}

func writeMemoryBytes(mem *Memory, addr uint64, data []byte) error {
	if len(data) == 0 {
		return nil
	}
	block, off, err := mem.rangeAccess(addr, int64(len(data)), true)
	if err != nil {
		return err
	}
	copy(block.data[off:off+len(data)], data)
	return nil
}

func complexAbsExtern(name string, realType bytecode.ValueType, realSize uint64) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 1 {
			return Value{}, nil, fmt.Errorf("%s expects 1 argument", name)
		}
		z, err := loadComplexArg(name, ec, args[0], realType, realSize)
		if err != nil {
			return Value{}, nil, err
		}
		return FloatValue(realType, math.Hypot(real(z), imag(z))), nil, nil
	}
}

func registerTgmathComplexExterns(r *ExternRegistry, base string, fn func(complex128) complex128) {
	r.Register(base+"f", complexUnaryExtern(base+"f", bytecode.TypeF32, 4, fn))
	r.Register(base, complexUnaryExtern(base, bytecode.TypeF64, 8, fn))
	r.Register(base+"l", complexUnaryExtern(base+"l", bytecode.TypeFLong, 16, fn))
}

func registerTgmathComplexBinaryExterns(r *ExternRegistry, base string, fn func(complex128, complex128) complex128) {
	r.Register(base+"f", complexBinaryExtern(base+"f", bytecode.TypeF32, 4, fn))
	r.Register(base, complexBinaryExtern(base, bytecode.TypeF64, 8, fn))
	r.Register(base+"l", complexBinaryExtern(base+"l", bytecode.TypeFLong, 16, fn))
}

func registerTgmathComplexRealExterns(r *ExternRegistry, base string, fn func(complex128) float64) {
	r.Register(base+"f", complexRealExtern(base+"f", bytecode.TypeF32, 4, fn))
	r.Register(base, complexRealExtern(base, bytecode.TypeF64, 8, fn))
	r.Register(base+"l", complexRealExtern(base+"l", bytecode.TypeFLong, 16, fn))
}

func complexUnaryExtern(name string, realType bytecode.ValueType, realSize uint64, fn func(complex128) complex128) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 1 {
			return Value{}, nil, fmt.Errorf("%s expects 1 argument", name)
		}
		z, err := loadComplexArg(name, ec, args[0], realType, realSize)
		if err != nil {
			return Value{}, nil, err
		}
		addr, err := storeComplexResult(name, ec, realType, realSize, fn(z))
		if err != nil {
			return Value{}, nil, err
		}
		return ObjectAddrValue(addr), nil, nil
	}
}

func complexRealExtern(name string, realType bytecode.ValueType, realSize uint64, fn func(complex128) float64) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 1 {
			return Value{}, nil, fmt.Errorf("%s expects 1 argument", name)
		}
		z, err := loadComplexArg(name, ec, args[0], realType, realSize)
		if err != nil {
			return Value{}, nil, err
		}
		return FloatValue(realType, fn(z)), nil, nil
	}
}

func complexBinaryExtern(name string, realType bytecode.ValueType, realSize uint64, fn func(complex128, complex128) complex128) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 2 {
			return Value{}, nil, fmt.Errorf("%s expects 2 arguments", name)
		}
		left, err := loadComplexArg(name, ec, args[0], realType, realSize)
		if err != nil {
			return Value{}, nil, err
		}
		right, err := loadComplexOrRealArg(name, ec, args[1], realType, realSize)
		if err != nil {
			return Value{}, nil, err
		}
		addr, err := storeComplexResult(name, ec, realType, realSize, fn(left, right))
		if err != nil {
			return Value{}, nil, err
		}
		return ObjectAddrValue(addr), nil, nil
	}
}

func loadComplexArg(name string, ec *ExternContext, arg Value, realType bytecode.ValueType, realSize uint64) (complex128, error) {
	if arg.Type != bytecode.TypeObjectAddr {
		return 0, fmt.Errorf("%s expects complex object address", name)
	}
	if ec == nil || ec.Memory == nil {
		return 0, fmt.Errorf("%s requires memory", name)
	}
	if arg.Int > math.MaxUint64-realSize {
		return 0, fmt.Errorf("%s complex imaginary component address overflows", name)
	}
	realPart, err := ec.Memory.Load(arg.Int, realType, 1)
	if err != nil {
		return 0, err
	}
	imagPart, err := ec.Memory.Load(arg.Int+realSize, realType, 1)
	if err != nil {
		return 0, err
	}
	return complex(cvmFloat(realPart), cvmFloat(imagPart)), nil
}

func loadComplexOrRealArg(name string, ec *ExternContext, arg Value, realType bytecode.ValueType, realSize uint64) (complex128, error) {
	if arg.Type == bytecode.TypeObjectAddr {
		return loadComplexArg(name, ec, arg, realType, realSize)
	}
	if !isFloatType(arg.Type) {
		return 0, fmt.Errorf("%s expects complex or floating argument", name)
	}
	return complex(cvmFloat(arg), 0), nil
}

func storeComplexResult(name string, ec *ExternContext, realType bytecode.ValueType, realSize uint64, z complex128) (uint64, error) {
	if ec == nil || ec.Memory == nil {
		return 0, fmt.Errorf("%s requires memory", name)
	}
	if realSize > math.MaxInt64/2 {
		return 0, fmt.Errorf("%s complex result size overflows", name)
	}
	addr, err := ec.Memory.TryAlloc("extern:"+name+":result", int64(realSize*2), int64(realSize), false, blockGlobal)
	if err != nil {
		return 0, err
	}
	if err := ec.Memory.Store(addr, realType, int64(realSize), FloatValue(realType, real(z))); err != nil {
		return 0, err
	}
	if err := ec.Memory.Store(addr+realSize, realType, int64(realSize), FloatValue(realType, imag(z))); err != nil {
		return 0, err
	}
	return addr, nil
}

func cvmComplexProject(z complex128) complex128 {
	if math.IsInf(real(z), 0) || math.IsInf(imag(z), 0) {
		return complex(math.Inf(1), math.Copysign(0, imag(z)))
	}
	return z
}

func registerTgmathRealExterns(r *ExternRegistry, base string, fn func(float64) float64) {
	r.Register(base+"f", mathUnaryFloatExtern(base+"f", bytecode.TypeF32, fn))
	r.Register(base, mathUnaryFloatExtern(base, bytecode.TypeF64, fn))
	r.Register(base+"l", mathUnaryFloatExtern(base+"l", bytecode.TypeFLong, fn))
}

func registerTgmathIntExterns(r *ExternRegistry, base string, fn func(float64) int) {
	r.Register(base+"f", mathUnaryIntExtern(base+"f", fn))
	r.Register(base, mathUnaryIntExtern(base, fn))
	r.Register(base+"l", mathUnaryIntExtern(base+"l", fn))
}

func registerTgmathLongExterns(r *ExternRegistry, base string, fn func(float64) float64) {
	r.Register(base+"f", mathUnaryLongExtern(base+"f", fn))
	r.Register(base, mathUnaryLongExtern(base, fn))
	r.Register(base+"l", mathUnaryLongExtern(base+"l", fn))
}

func registerTgmathRealBinaryExterns(r *ExternRegistry, base string, fn func(float64, float64) float64) {
	r.Register(base+"f", mathBinaryFloatExtern(base+"f", bytecode.TypeF32, fn))
	r.Register(base, mathBinaryFloatExtern(base, bytecode.TypeF64, fn))
	r.Register(base+"l", mathBinaryFloatExtern(base+"l", bytecode.TypeFLong, fn))
}

func registerTgmathRealIntBinaryExterns(r *ExternRegistry, base string, fn func(float64, int) float64) {
	r.Register(base+"f", mathFloatIntExtern(base+"f", bytecode.TypeF32, fn))
	r.Register(base, mathFloatIntExtern(base, bytecode.TypeF64, fn))
	r.Register(base+"l", mathFloatIntExtern(base+"l", bytecode.TypeFLong, fn))
}

func registerTgmathFrexpExterns(r *ExternRegistry, base string) {
	r.Register(base+"f", mathFrexpExtern(base+"f", bytecode.TypeF32))
	r.Register(base, mathFrexpExtern(base, bytecode.TypeF64))
	r.Register(base+"l", mathFrexpExtern(base+"l", bytecode.TypeFLong))
}

func registerTgmathRemquoExterns(r *ExternRegistry, base string) {
	r.Register(base+"f", mathRemquoExtern(base+"f", bytecode.TypeF32))
	r.Register(base, mathRemquoExtern(base, bytecode.TypeF64))
	r.Register(base+"l", mathRemquoExtern(base+"l", bytecode.TypeFLong))
}

func registerTgmathRealTernaryExterns(r *ExternRegistry, base string, fn func(float64, float64, float64) float64) {
	r.Register(base+"f", mathTernaryFloatExtern(base+"f", bytecode.TypeF32, fn))
	r.Register(base, mathTernaryFloatExtern(base, bytecode.TypeF64, fn))
	r.Register(base+"l", mathTernaryFloatExtern(base+"l", bytecode.TypeFLong, fn))
}

func mathUnaryFloatExtern(name string, ret bytecode.ValueType, fn func(float64) float64) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 1 {
			return Value{}, nil, fmt.Errorf("%s expects 1 argument", name)
		}
		if !isFloatType(args[0].Type) {
			return Value{}, nil, fmt.Errorf("%s expects floating argument", name)
		}
		return floatResult(ret, fn(cvmFloat(args[0]))), nil, nil
	}
}

func mathUnaryIntExtern(name string, fn func(float64) int) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 1 {
			return Value{}, nil, fmt.Errorf("%s expects 1 argument", name)
		}
		if !isFloatType(args[0].Type) {
			return Value{}, nil, fmt.Errorf("%s expects floating argument", name)
		}
		return IntValue(bytecode.TypeI32, int64(fn(cvmFloat(args[0])))), nil, nil
	}
}

func mathUnaryLongExtern(name string, fn func(float64) float64) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 1 {
			return Value{}, nil, fmt.Errorf("%s expects 1 argument", name)
		}
		if !isFloatType(args[0].Type) {
			return Value{}, nil, fmt.Errorf("%s expects floating argument", name)
		}
		return IntValue(bytecode.TypeI64, int64(fn(cvmFloat(args[0])))), nil, nil
	}
}

func mathConstantFloatExtern(name string, ret bytecode.ValueType, value float64) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 0 {
			return Value{}, nil, fmt.Errorf("%s expects 0 arguments", name)
		}
		return floatResult(ret, value), nil, nil
	}
}

func mathNanExtern(name string) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 1 {
			return Value{}, nil, fmt.Errorf("%s expects 1 argument", name)
		}
		if !isPointerType(args[0].Type) {
			return Value{}, nil, fmt.Errorf("%s expects string argument", name)
		}
		return FloatValue(bytecode.TypeF64, math.NaN()), nil, nil
	}
}

func mathBinaryFloatExtern(name string, ret bytecode.ValueType, fn func(float64, float64) float64) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 2 {
			return Value{}, nil, fmt.Errorf("%s expects 2 arguments", name)
		}
		if !isFloatType(args[0].Type) || !isFloatType(args[1].Type) {
			return Value{}, nil, fmt.Errorf("%s expects floating arguments", name)
		}
		return floatResult(ret, fn(cvmFloat(args[0]), cvmFloat(args[1]))), nil, nil
	}
}

func mathFloatIntExtern(name string, ret bytecode.ValueType, fn func(float64, int) float64) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 2 {
			return Value{}, nil, fmt.Errorf("%s expects 2 arguments", name)
		}
		if !isFloatType(args[0].Type) || !isIntegerLike(args[1].Type) {
			return Value{}, nil, fmt.Errorf("%s expects floating and integer arguments", name)
		}
		exp := signedInt(args[1])
		if int64(int(exp)) != exp {
			return Value{}, nil, fmt.Errorf("%s exponent %d exceeds int range", name, exp)
		}
		return floatResult(ret, fn(cvmFloat(args[0]), int(exp))), nil, nil
	}
}

func mathFrexpExtern(name string, ret bytecode.ValueType) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 2 {
			return Value{}, nil, fmt.Errorf("%s expects 2 arguments", name)
		}
		if !isFloatType(args[0].Type) || !isPointerType(args[1].Type) {
			return Value{}, nil, fmt.Errorf("%s expects floating and pointer arguments", name)
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("%s requires memory", name)
		}
		frac, exp := math.Frexp(cvmFloat(args[0]))
		if err := ec.Memory.Store(args[1].Int, bytecode.TypeI32, 4, IntValue(bytecode.TypeI32, int64(exp))); err != nil {
			return Value{}, nil, err
		}
		return floatResult(ret, frac), nil, nil
	}
}

func mathRemquoExtern(name string, ret bytecode.ValueType) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 3 {
			return Value{}, nil, fmt.Errorf("%s expects 3 arguments", name)
		}
		if !isFloatType(args[0].Type) || !isFloatType(args[1].Type) || !isPointerType(args[2].Type) {
			return Value{}, nil, fmt.Errorf("%s expects floating, floating, and pointer arguments", name)
		}
		if ec == nil || ec.Memory == nil {
			return Value{}, nil, fmt.Errorf("%s requires memory", name)
		}
		x, y := cvmFloat(args[0]), cvmFloat(args[1])
		quo := int64(math.RoundToEven(x / y))
		if err := ec.Memory.Store(args[2].Int, bytecode.TypeI32, 4, IntValue(bytecode.TypeI32, quo)); err != nil {
			return Value{}, nil, err
		}
		return floatResult(ret, math.Remainder(x, y)), nil, nil
	}
}

func mathTernaryFloatExtern(name string, ret bytecode.ValueType, fn func(float64, float64, float64) float64) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 3 {
			return Value{}, nil, fmt.Errorf("%s expects 3 arguments", name)
		}
		if !isFloatType(args[0].Type) || !isFloatType(args[1].Type) || !isFloatType(args[2].Type) {
			return Value{}, nil, fmt.Errorf("%s expects floating arguments", name)
		}
		return floatResult(ret, fn(cvmFloat(args[0]), cvmFloat(args[1]), cvmFloat(args[2]))), nil, nil
	}
}

func mathUnaryExtern(name string, pred func(Value) int64) ExternFunc {
	return func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		if len(args) != 1 {
			return Value{}, nil, fmt.Errorf("%s expects 1 argument", name)
		}
		if !isFloatType(args[0].Type) {
			return Value{}, nil, fmt.Errorf("%s expects floating argument", name)
		}
		return IntValue(bytecode.TypeI32, pred(args[0])), nil, nil
	}
}

const (
	fpClassNaN       = 0
	fpClassInfinite  = 1
	fpClassNormal    = 2
	fpClassSubnormal = 3
	fpClassZero      = 4

	minNormalFloat32 = 1.1754943508222875e-38
	minNormalFloat64 = 2.2250738585072014e-308
)

func cvmFPClassify(v Value) int64 {
	f := cvmFloat(v)
	switch {
	case math.IsNaN(f):
		return fpClassNaN
	case math.IsInf(f, 0):
		return fpClassInfinite
	case f == 0:
		return fpClassZero
	}
	minNormal := minNormalFloat64
	if v.Type == bytecode.TypeF32 {
		minNormal = minNormalFloat32
	}
	if math.Abs(f) < minNormal {
		return fpClassSubnormal
	}
	return fpClassNormal
}

func cvmFloat(v Value) float64 {
	if v.Type == bytecode.TypeF32 {
		return float64(float32(v.Float))
	}
	return v.Float
}

func strcmpResult(left, right string) int {
	for i := 0; i < len(left) && i < len(right); i++ {
		if left[i] < right[i] {
			return -1
		}
		if left[i] > right[i] {
			return 1
		}
	}
	if len(left) < len(right) {
		return -1
	}
	if len(left) > len(right) {
		return 1
	}
	return 0
}

func memcmpResult(left, right []byte) int {
	for i := 0; i < len(left) && i < len(right); i++ {
		if left[i] < right[i] {
			return -1
		}
		if left[i] > right[i] {
			return 1
		}
	}
	if len(left) < len(right) {
		return -1
	}
	if len(left) > len(right) {
		return 1
	}
	return 0
}

func (r *ExternRegistry) Register(name string, fn ExternFunc) {
	if fn == nil {
		delete(r.funcs, name)
		return
	}
	r.funcs[name] = fn
}

func (r *ExternRegistry) Lookup(name string) (ExternFunc, bool) {
	fn, ok := r.funcs[name]
	return fn, ok && fn != nil
}

func (r *ExternRegistry) LookupVariable(name string, mem *Memory) (uint64, bool) {
	addr, ok, err := r.LookupVariableAddr(name, mem)
	if err != nil {
		return 0, false
	}
	return addr, ok
}

func (r *ExternRegistry) LookupVariableAddr(name string, mem *Memory) (uint64, bool, error) {
	switch name {
	case "stdout":
		addr, err := r.allocHostWriter(name, mem, r.stdout)
		return addr, true, err
	case "stderr":
		addr, err := r.allocHostWriter(name, mem, r.stderr)
		return addr, true, err
	default:
		return 0, false, nil
	}
}

func (r *ExternRegistry) context(mem *Memory) *ExternContext {
	return &ExternContext{Memory: mem, Stdout: r.stdout, Stderr: r.stderr}
}

func (r *ExternRegistry) externStdout(ec *ExternContext) io.Writer {
	if ec != nil && ec.Stdout != nil {
		return ec.Stdout
	}
	return r.stdout
}

func (r *ExternRegistry) lookupHostWriter(addr uint64) (io.Writer, bool) {
	w, ok := r.hostWriters[addr]
	return w, ok
}

func (r *ExternRegistry) allocHostWriter(name string, mem *Memory, w io.Writer) (uint64, error) {
	if mem == nil {
		return 0, fmt.Errorf("memory is nil")
	}
	addr, err := mem.TryAlloc("extern:"+name, mem.target.PointerSize, mem.target.PointerAlign, false, blockHostHandle)
	if err != nil {
		return 0, err
	}
	if err := mem.WritePointer(addr, addr); err != nil {
		return 0, err
	}
	if b, _, err := mem.rangeAccess(addr, mem.target.PointerSize, false); err == nil {
		b.readonly = true
	}
	r.hostWriters[addr] = w
	return addr, nil
}
