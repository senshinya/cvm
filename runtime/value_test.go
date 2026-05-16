package runtime

import (
	"math"
	"testing"

	"shinya.click/cvm/bytecode"
)

func TestValueAsExitCode(t *testing.T) {
	tests := []struct {
		name string
		v    Value
		want int
	}{
		{"i32", IntValue(bytecode.TypeI32, 42), 42},
		{"u8", IntValue(bytecode.TypeU8, 255), 255},
		{"bool", IntValue(bytecode.TypeBool, 1), 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.v.ExitCode()
			if err != nil {
				t.Fatalf("ExitCode returned error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("ExitCode = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestValueRejectsPointerExitCode(t *testing.T) {
	_, err := PtrValue(0x1000).ExitCode()
	if err == nil {
		t.Fatal("ExitCode accepted pointer")
	}
}

func TestValueAsExitCodeAcceptsNegativeSignedInt(t *testing.T) {
	got, err := IntValue(bytecode.TypeI32, -1).ExitCode()
	if err != nil {
		t.Fatalf("ExitCode returned error: %v", err)
	}
	if got != -1 {
		t.Fatalf("ExitCode = %d, want -1", got)
	}
}

func TestValueAsExitCodeSignExtendsRawSignedValues(t *testing.T) {
	tests := []struct {
		name string
		v    Value
		want int
	}{
		{"i8", Value{Type: bytecode.TypeI8, Int: 0xff}, -1},
		{"i32", Value{Type: bytecode.TypeI32, Int: 0xffffffff}, -1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.v.ExitCode()
			if err != nil {
				t.Fatalf("ExitCode returned error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("ExitCode = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestValueRejectsUnsignedExitCodeOverflow(t *testing.T) {
	_, err := UIntValue(bytecode.TypeU64, uint64(math.MaxInt)+1).ExitCode()
	if err == nil {
		t.Fatal("ExitCode accepted overflowing unsigned value")
	}
}

func TestValueRejectsFloatExitCode(t *testing.T) {
	_, err := FloatValue(bytecode.TypeF64, 1).ExitCode()
	if err == nil {
		t.Fatal("ExitCode accepted float")
	}
}

func TestValueIsZero(t *testing.T) {
	tests := []struct {
		name string
		v    Value
		want bool
	}{
		{"zero int", IntValue(bytecode.TypeI32, 0), true},
		{"nonzero int", IntValue(bytecode.TypeI32, 1), false},
		{"zero float", FloatValue(bytecode.TypeF64, 0), true},
		{"nonzero float", FloatValue(bytecode.TypeF64, 1), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.v.IsZero(); got != tt.want {
				t.Fatalf("IsZero = %v, want %v", got, tt.want)
			}
		})
	}
}
