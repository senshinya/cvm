package runtime

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"shinya.click/cvm/bytecode"
)

func TestDefaultExternRegistryHasExitAndAbort(t *testing.T) {
	reg := DefaultExternRegistry(nil, nil)
	for _, name := range []string{"exit", "abort", "puts", "fputs"} {
		if _, ok := reg.Lookup(name); !ok {
			t.Fatalf("missing extern %s", name)
		}
	}
}

func TestPutsWritesCString(t *testing.T) {
	var out bytes.Buffer
	reg := DefaultExternRegistry(&out, nil)
	mem := NewMemory(bytecode.DefaultTarget())
	addr := mem.AllocBytes("string:0", []byte("hello\x00"), true, blockString)
	fn, _ := reg.Lookup("puts")
	ret, exit, err := fn(context.Background(), &ExternContext{Memory: mem}, []Value{ObjectAddrValue(addr)})
	if err != nil || exit != nil {
		t.Fatalf("puts ret=%#v exit=%#v err=%v", ret, exit, err)
	}
	if out.String() != "hello\n" {
		t.Fatalf("puts output = %q", out.String())
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
