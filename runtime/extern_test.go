package runtime

import (
	"bytes"
	"context"
	"math"
	"strings"
	"testing"

	"shinya.click/cvm/bytecode"
)

func TestDefaultExternRegistryHasExitAndAbort(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	for _, name := range []string{"exit", "abort", "puts", "fputs", "strcmp", "memcmp", "feclearexcept", "fetestexcept"} {
		if _, ok := reg.Lookup(name); !ok {
			t.Fatalf("missing extern %s", name)
		}
	}
}

func TestPutsWritesCString(t *testing.T) {
	var out bytes.Buffer
	reg := DefaultExternRegistry(&out, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	addr := mustAllocBytes(t, mem, "string:0", []byte("hello\x00"), true, blockString)
	fn, _ := reg.Lookup("puts")
	ret, exit, err := fn(context.Background(), &ExternContext{Memory: mem}, []Value{ObjectAddrValue(addr)})
	if err != nil || exit != nil {
		t.Fatalf("puts ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if out.String() != "hello\n" {
		t.Fatalf("puts output = %q", out.String())
	}
}

func TestFputsWritesCStringToStderrHostHandle(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer
	reg := DefaultExternRegistry(&out, &errOut)
	mem := NewMemory(bytecode.DefaultTarget())
	addr := mustAllocBytes(t, mem, "string:0", []byte("hello\x00"), true, blockString)
	stderr, ok := reg.LookupVariable("stderr", mem)
	if !ok {
		t.Fatal("missing stderr extern variable")
	}
	fn, _ := reg.Lookup("fputs")
	ret, exit, err := fn(context.Background(), &ExternContext{Memory: mem}, []Value{ObjectAddrValue(addr), PtrValue(stderr)})
	if err != nil || exit != nil {
		t.Fatalf("fputs ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if out.String() != "" {
		t.Fatalf("stdout output = %q, want empty", out.String())
	}
	if errOut.String() != "hello" {
		t.Fatalf("stderr output = %q", errOut.String())
	}
}

func TestFputsWritesCStringToLoadedStderrHostHandle(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer
	reg := DefaultExternRegistry(&out, &errOut)
	target := bytecode.DefaultTarget()
	mem := NewMemory(target)
	addr := mustAllocBytes(t, mem, "string:0", []byte("hello\x00"), true, blockString)
	stderrAddr, ok := reg.LookupVariable("stderr", mem)
	if !ok {
		t.Fatal("missing stderr extern variable")
	}
	loaded, err := mem.Load(stderrAddr, bytecode.TypePtr, target.PointerAlign)
	if err != nil {
		t.Fatalf("Load(stderr): %v", err)
	}
	fn, _ := reg.Lookup("fputs")
	ret, exit, err := fn(context.Background(), &ExternContext{Memory: mem}, []Value{ObjectAddrValue(addr), loaded})
	if err != nil || exit != nil {
		t.Fatalf("fputs ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if out.String() != "" {
		t.Fatalf("stdout output = %q, want empty", out.String())
	}
	if errOut.String() != "hello" {
		t.Fatalf("stderr output = %q", errOut.String())
	}
}

func TestFputsUnknownStreamHandleReturnsError(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer
	reg := DefaultExternRegistry(&out, &errOut)
	mem := NewMemory(bytecode.DefaultTarget())
	addr := mustAllocBytes(t, mem, "string:0", []byte("hello\x00"), true, blockString)
	fn, _ := reg.Lookup("fputs")
	_, _, err := fn(context.Background(), &ExternContext{Memory: mem}, []Value{ObjectAddrValue(addr), PtrValue(0xdeadbeef)})
	if err == nil || !strings.Contains(err.Error(), "unknown stream handle") {
		t.Fatalf("fputs err = %v, want unknown stream handle", err)
	}
	if out.String() != "" || errOut.String() != "" {
		t.Fatalf("stdout=%q stderr=%q, want no output", out.String(), errOut.String())
	}
}

func TestStrcmpComparesCStrings(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	fn, ok := reg.Lookup("strcmp")
	if !ok {
		t.Fatal("missing strcmp extern")
	}
	mem := NewMemory(bytecode.DefaultTarget())
	tests := []struct {
		name string
		a    string
		b    string
		want int64
	}{
		{name: "equal", a: "abc", b: "abc", want: 0},
		{name: "left less", a: "abc", b: "abd", want: -1},
		{name: "left greater", a: "abd", b: "abc", want: 1},
		{name: "prefix less", a: "ab", b: "abc", want: -1},
		{name: "prefix greater", a: "abc", b: "ab", want: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := mustAllocBytes(t, mem, "strcmp:a", []byte(tt.a+"\x00"), true, blockString)
			b := mustAllocBytes(t, mem, "strcmp:b", []byte(tt.b+"\x00"), true, blockString)
			ret, exit, err := fn(context.Background(), &ExternContext{Memory: mem}, []Value{ObjectAddrValue(a), ObjectAddrValue(b)})
			if err != nil || exit != nil {
				t.Fatalf("strcmp ret=%#v exit=%#v err=%v", ret, exit, err)
			}
			if ret.Type != bytecode.TypeI32 || int64(int32(ret.Int)) != tt.want {
				t.Fatalf("strcmp(%q, %q) = %#v, want i32 %d", tt.a, tt.b, ret, tt.want)
			}
		})
	}
}

func TestStrcmpRequiresMemory(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	fn, ok := reg.Lookup("strcmp")
	if !ok {
		t.Fatal("missing strcmp extern")
	}
	_, _, err := fn(context.Background(), nil, []Value{PtrValue(0), PtrValue(0)})
	if err == nil || !strings.Contains(err.Error(), "strcmp requires memory") {
		t.Fatalf("strcmp err = %v, want memory error", err)
	}
}

func TestMemcmpComparesBytes(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	fn, ok := reg.Lookup("memcmp")
	if !ok {
		t.Fatal("missing memcmp extern")
	}
	mem := NewMemory(bytecode.DefaultTarget())
	left := mustAllocBytes(t, mem, "memcmp:left", []byte{0, 1, 2, 3}, true, blockString)
	same := mustAllocBytes(t, mem, "memcmp:same", []byte{0, 1, 2, 9}, true, blockString)
	diff := mustAllocBytes(t, mem, "memcmp:diff", []byte{0, 1, 4, 3}, true, blockString)
	less := mustAllocBytes(t, mem, "memcmp:less", []byte{0, 1, 0, 3}, true, blockString)
	tests := []struct {
		name  string
		right uint64
		n     int64
		want  int64
	}{
		{name: "zero length", right: diff, n: 0, want: 0},
		{name: "equal prefix", right: same, n: 3, want: 0},
		{name: "left less", right: diff, n: 4, want: -1},
		{name: "left greater", right: less, n: 4, want: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ret, exit, err := fn(context.Background(), &ExternContext{Memory: mem}, []Value{ObjectAddrValue(left), ObjectAddrValue(tt.right), UIntValue(bytecode.TypeU64, uint64(tt.n))})
			if err != nil || exit != nil {
				t.Fatalf("memcmp ret=%#v exit=%#v err=%v", ret, exit, err)
			}
			if ret.Type != bytecode.TypeI32 || int64(int32(ret.Int)) != tt.want {
				t.Fatalf("memcmp = %#v, want i32 %d", ret, tt.want)
			}
		})
	}
}

func TestFenvExternsAreNoOps(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	for _, name := range []string{"feclearexcept", "fetestexcept"} {
		t.Run(name, func(t *testing.T) {
			fn, ok := reg.Lookup(name)
			if !ok {
				t.Fatalf("missing %s extern", name)
			}
			ret, exit, err := fn(context.Background(), nil, []Value{IntValue(bytecode.TypeI32, 0)})
			if err != nil || exit != nil {
				t.Fatalf("%s ret=%#v exit=%#v err=%v", name, ret, exit, err)
			}
			if ret.Type != bytecode.TypeI32 || ret.Int != 0 {
				t.Fatalf("%s ret = %#v, want i32 0", name, ret)
			}
		})
	}
}

func TestMathClassificationExterns(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	tests := []struct {
		name string
		fn   string
		arg  Value
		want int64
	}{
		{name: "float subnormal", fn: "__cvm_fpclassifyf", arg: FloatValue(bytecode.TypeF32, float64(math.SmallestNonzeroFloat32)), want: 3},
		{name: "double subnormal", fn: "__cvm_fpclassify", arg: FloatValue(bytecode.TypeF64, math.SmallestNonzeroFloat64*2), want: 3},
		{name: "double normal", fn: "__cvm_fpclassify", arg: FloatValue(bytecode.TypeF64, minNormalFloat64), want: 2},
		{name: "double zero", fn: "__cvm_fpclassify", arg: FloatValue(bytecode.TypeF64, 0), want: 4},
		{name: "double nan", fn: "__cvm_isnan", arg: FloatValue(bytecode.TypeF64, math.NaN()), want: 1},
		{name: "double inf", fn: "__cvm_isinf", arg: FloatValue(bytecode.TypeF64, math.Inf(1)), want: 1},
		{name: "negative zero sign", fn: "__cvm_signbit", arg: FloatValue(bytecode.TypeF64, math.Copysign(0, -1)), want: 1},
		{name: "long double normal", fn: "__cvm_isnormall", arg: FloatValue(bytecode.TypeFLong, 1), want: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn, ok := reg.Lookup(tt.fn)
			if !ok {
				t.Fatalf("missing %s extern", tt.fn)
			}
			ret, exit, err := fn(context.Background(), nil, []Value{tt.arg})
			if err != nil || exit != nil {
				t.Fatalf("%s ret=%#v exit=%#v err=%v", tt.fn, ret, exit, err)
			}
			if ret.Type != bytecode.TypeI32 || int64(int32(ret.Int)) != tt.want {
				t.Fatalf("%s ret = %#v, want i32 %d", tt.fn, ret, tt.want)
			}
		})
	}
}

func TestMathUnorderedExtern(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	fn, ok := reg.Lookup("__cvm_isunordered")
	if !ok {
		t.Fatal("missing __cvm_isunordered extern")
	}
	ret, exit, err := fn(context.Background(), nil, []Value{
		FloatValue(bytecode.TypeF64, math.NaN()),
		FloatValue(bytecode.TypeF64, 1),
	})
	if err != nil || exit != nil {
		t.Fatalf("__cvm_isunordered ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if ret.Type != bytecode.TypeI32 || ret.Int != 1 {
		t.Fatalf("__cvm_isunordered ret = %#v, want i32 1", ret)
	}
}

func TestTgmathFloatExterns(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	tests := []struct {
		name string
		fn   string
		args []Value
		want float64
	}{
		{name: "sin", fn: "__cvm_tgmath_sin", args: []Value{FloatValue(bytecode.TypeF64, 0)}, want: 0},
		{name: "exp", fn: "__cvm_tgmath_exp", args: []Value{FloatValue(bytecode.TypeF64, 0)}, want: 1},
		{name: "sqrt", fn: "__cvm_tgmath_sqrt", args: []Value{FloatValue(bytecode.TypeF64, 9)}, want: 3},
		{name: "cos", fn: "__cvm_tgmath_cos", args: []Value{FloatValue(bytecode.TypeF64, 0)}, want: 1},
		{name: "tan", fn: "__cvm_tgmath_tan", args: []Value{FloatValue(bytecode.TypeF64, 0)}, want: 0},
		{name: "log", fn: "__cvm_tgmath_log", args: []Value{FloatValue(bytecode.TypeF64, 1)}, want: 0},
		{name: "sinh", fn: "__cvm_tgmath_sinh", args: []Value{FloatValue(bytecode.TypeF64, 0)}, want: 0},
		{name: "cosh", fn: "__cvm_tgmath_cosh", args: []Value{FloatValue(bytecode.TypeF64, 0)}, want: 1},
		{name: "tanh", fn: "__cvm_tgmath_tanh", args: []Value{FloatValue(bytecode.TypeF64, 0)}, want: 0},
		{name: "asin", fn: "__cvm_tgmath_asin", args: []Value{FloatValue(bytecode.TypeF64, 0)}, want: 0},
		{name: "acos", fn: "__cvm_tgmath_acos", args: []Value{FloatValue(bytecode.TypeF64, 1)}, want: 0},
		{name: "atan", fn: "__cvm_tgmath_atan", args: []Value{FloatValue(bytecode.TypeF64, 0)}, want: 0},
		{name: "asinh", fn: "__cvm_tgmath_asinh", args: []Value{FloatValue(bytecode.TypeF64, 0)}, want: 0},
		{name: "acosh", fn: "__cvm_tgmath_acosh", args: []Value{FloatValue(bytecode.TypeF64, 1)}, want: 0},
		{name: "atanh", fn: "__cvm_tgmath_atanh", args: []Value{FloatValue(bytecode.TypeF64, 0)}, want: 0},
		{name: "cbrt", fn: "__cvm_tgmath_cbrt", args: []Value{FloatValue(bytecode.TypeF64, 8)}, want: 2},
		{name: "ceil", fn: "__cvm_tgmath_ceil", args: []Value{FloatValue(bytecode.TypeF64, 1.25)}, want: 2},
		{name: "floor", fn: "__cvm_tgmath_floor", args: []Value{FloatValue(bytecode.TypeF64, 1.75)}, want: 1},
		{name: "trunc", fn: "__cvm_tgmath_trunc", args: []Value{FloatValue(bytecode.TypeF64, -1.75)}, want: -1},
		{name: "round", fn: "__cvm_tgmath_round", args: []Value{FloatValue(bytecode.TypeF64, -1.5)}, want: -2},
		{name: "exp2", fn: "__cvm_tgmath_exp2", args: []Value{FloatValue(bytecode.TypeF64, 3)}, want: 8},
		{name: "expm1", fn: "__cvm_tgmath_expm1", args: []Value{FloatValue(bytecode.TypeF64, 0)}, want: 0},
		{name: "log10", fn: "__cvm_tgmath_log10", args: []Value{FloatValue(bytecode.TypeF64, 1)}, want: 0},
		{name: "log1p", fn: "__cvm_tgmath_log1p", args: []Value{FloatValue(bytecode.TypeF64, 0)}, want: 0},
		{name: "log2", fn: "__cvm_tgmath_log2", args: []Value{FloatValue(bytecode.TypeF64, 8)}, want: 3},
		{name: "atan2", fn: "__cvm_tgmath_atan2", args: []Value{FloatValue(bytecode.TypeF64, 0), FloatValue(bytecode.TypeF64, 1)}, want: 0},
		{name: "hypot", fn: "__cvm_tgmath_hypot", args: []Value{FloatValue(bytecode.TypeF64, 3), FloatValue(bytecode.TypeF64, 4)}, want: 5},
		{name: "fdim", fn: "__cvm_tgmath_fdim", args: []Value{FloatValue(bytecode.TypeF64, 5), FloatValue(bytecode.TypeF64, 3)}, want: 2},
		{name: "fmax", fn: "__cvm_tgmath_fmax", args: []Value{FloatValue(bytecode.TypeF64, 2), FloatValue(bytecode.TypeF64, 3)}, want: 3},
		{name: "fmin", fn: "__cvm_tgmath_fmin", args: []Value{FloatValue(bytecode.TypeF64, 2), FloatValue(bytecode.TypeF64, 3)}, want: 2},
		{name: "fmod", fn: "__cvm_tgmath_fmod", args: []Value{FloatValue(bytecode.TypeF64, 5), FloatValue(bytecode.TypeF64, 2)}, want: 1},
		{name: "remainder", fn: "__cvm_tgmath_remainder", args: []Value{FloatValue(bytecode.TypeF64, 4), FloatValue(bytecode.TypeF64, 2)}, want: 0},
		{name: "copysign", fn: "__cvm_tgmath_copysign", args: []Value{FloatValue(bytecode.TypeF64, 2), FloatValue(bytecode.TypeF64, -1)}, want: -2},
		{name: "nextafter", fn: "__cvm_tgmath_nextafter", args: []Value{FloatValue(bytecode.TypeF64, 1), FloatValue(bytecode.TypeF64, 1)}, want: 1},
		{name: "fma", fn: "__cvm_tgmath_fma", args: []Value{FloatValue(bytecode.TypeF64, 2), FloatValue(bytecode.TypeF64, 3), FloatValue(bytecode.TypeF64, 4)}, want: 10},
		{name: "pow", fn: "__cvm_tgmath_pow", args: []Value{FloatValue(bytecode.TypeF64, 2), FloatValue(bytecode.TypeF64, 3)}, want: 8},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn, ok := reg.Lookup(tt.fn)
			if !ok {
				t.Fatalf("missing %s extern", tt.fn)
			}
			ret, exit, err := fn(context.Background(), nil, tt.args)
			if err != nil || exit != nil {
				t.Fatalf("%s ret=%#v exit=%#v err=%v", tt.fn, ret, exit, err)
			}
			if ret.Type != bytecode.TypeF64 || ret.Float != tt.want {
				t.Fatalf("%s ret = %#v, want f64 %v", tt.fn, ret, tt.want)
			}
		})
	}
}

func TestAbortReturnsTrap(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	fn, _ := reg.Lookup("abort")
	_, _, err := fn(context.Background(), &ExternContext{Memory: NewMemory(bytecode.DefaultTarget())}, nil)
	if err == nil || !strings.Contains(err.Error(), "abort") {
		t.Fatalf("abort err = %v, want abort trap", err)
	}
}
