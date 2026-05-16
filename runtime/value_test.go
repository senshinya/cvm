package runtime

import (
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
