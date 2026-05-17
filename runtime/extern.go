package runtime

import (
	"context"
	"fmt"
	"io"
	"math"
	"math/cmplx"
	"os"

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
	r.Register("abort", func(ctx context.Context, ec *ExternContext, args []Value) (Value, *ExitStatus, error) {
		return Value{}, nil, &TrapError{Reason: "abort"}
	})
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
	registerTgmathRealBinaryExterns(r, "__cvm_tgmath_pow", math.Pow)
	registerTgmathRealBinaryExterns(r, "__cvm_tgmath_atan2", math.Atan2)
	registerTgmathRealBinaryExterns(r, "__cvm_tgmath_hypot", math.Hypot)
	registerTgmathRealBinaryExterns(r, "__cvm_tgmath_fdim", math.Dim)
	registerTgmathRealBinaryExterns(r, "__cvm_tgmath_fmax", math.Max)
	registerTgmathRealBinaryExterns(r, "__cvm_tgmath_fmin", math.Min)
	registerTgmathRealBinaryExterns(r, "__cvm_tgmath_fmod", math.Mod)
	registerTgmathRealBinaryExterns(r, "__cvm_tgmath_remainder", math.Remainder)
	registerTgmathRealBinaryExterns(r, "__cvm_tgmath_copysign", math.Copysign)
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

func registerTgmathRealExterns(r *ExternRegistry, base string, fn func(float64) float64) {
	r.Register(base+"f", mathUnaryFloatExtern(base+"f", bytecode.TypeF32, fn))
	r.Register(base, mathUnaryFloatExtern(base, bytecode.TypeF64, fn))
	r.Register(base+"l", mathUnaryFloatExtern(base+"l", bytecode.TypeFLong, fn))
}

func registerTgmathRealBinaryExterns(r *ExternRegistry, base string, fn func(float64, float64) float64) {
	r.Register(base+"f", mathBinaryFloatExtern(base+"f", bytecode.TypeF32, fn))
	r.Register(base, mathBinaryFloatExtern(base, bytecode.TypeF64, fn))
	r.Register(base+"l", mathBinaryFloatExtern(base+"l", bytecode.TypeFLong, fn))
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
