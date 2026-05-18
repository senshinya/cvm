package main

import (
	"context"
	"fmt"
	"os"

	cvmruntime "shinya.click/cvm/runtime"
)

func main() {
	os.Exit(runMain(os.Args[1:]))
}

func runMain(args []string) int {
	if len(args) > 0 && args[0] == "run" {
		return runBytecode(args[1:])
	}
	return runCompileMode(args)
}

func runBytecode(args []string) int {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: cvm run file.cvmbc [args...]")
		return 2
	}
	f, err := os.Open(args[0])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	defer f.Close()
	progArgs := append([]string(nil), args...)
	prog, err := cvmruntime.Load(f, cvmruntime.LoadOptions{Args: progArgs})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	st, err := cvmruntime.Run(context.Background(), prog, cvmruntime.RunOptions{})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return st.Code
}

func runCompileMode(args []string) int {
	dumpIR := false
	dumpBytecode := false
	emitBytecode := ""
	files := make([]string, 0, 1)
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--dump-ir":
			dumpIR = true
		case "--dump-bytecode":
			dumpBytecode = true
		case "--emit-bytecode":
			i++
			if i >= len(args) {
				fmt.Println("Usage: cvm [--dump-ir|--dump-bytecode|--emit-bytecode out.cvmbc] [file]")
				return 2
			}
			emitBytecode = args[i]
		default:
			files = append(files, arg)
		}
	}
	if len(files) != 1 {
		fmt.Println("Usage: cvm [--dump-ir|--dump-bytecode|--emit-bytecode out.cvmbc] [file]")
		return 2
	}
	c := &Compiler{DumpIR: dumpIR, DumpBytecode: dumpBytecode, EmitBytecode: emitBytecode}
	if err := c.RunFile(files[0]); err != nil {
		c.handleError(err)
		return 1
	}
	return 0
}
