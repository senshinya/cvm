package runtime

import (
	"bytes"
	"context"
	"testing"

	"shinya.click/cvm/bytecode"
)

func runModule(t *testing.T, mod *bytecode.Module) (ExitStatus, error) {
	t.Helper()
	var buf bytes.Buffer
	if err := bytecode.EncodeModule(&buf, mod); err != nil {
		t.Fatalf("EncodeModule: %v", err)
	}
	p, err := Load(bytes.NewReader(buf.Bytes()), LoadOptions{Externs: DefaultExternRegistry(nil, nil)})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	return Run(context.Background(), p, RunOptions{})
}

func TestRunReturnsMainConstant(t *testing.T) {
	st, err := runModule(t, testMainModule(bytecode.I32Const(5), bytecode.Return(bytecode.TypeI32)))
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if st.Code != 5 {
		t.Fatalf("exit code = %d, want 5", st.Code)
	}
}

func TestRunLocalStoreLoad(t *testing.T) {
	mod := testMainModule(
		bytecode.I32Const(9),
		bytecode.StoreLocal(bytecode.TypeI32, 0),
		bytecode.LoadLocal(bytecode.TypeI32, 0),
		bytecode.Return(bytecode.TypeI32),
	)
	mod.Functions[0].Locals = []bytecode.LocalSlot{{ID: 0, Name: "x", Type: bytecode.TypeI32}}
	st, err := runModule(t, mod)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if st.Code != 9 {
		t.Fatalf("exit code = %d, want 9", st.Code)
	}
}
