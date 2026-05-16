package main

import (
	"fmt"
	"os"
)

func main() {
	dumpIR := false
	dumpBytecode := false
	emitBytecode := ""
	args := os.Args[1:]
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
				return
			}
			emitBytecode = args[i]
		default:
			files = append(files, arg)
		}
	}
	if len(files) != 1 {
		fmt.Println("Usage: cvm [--dump-ir|--dump-bytecode|--emit-bytecode out.cvmbc] [file]")
		return
	}
	c := &Compiler{DumpIR: dumpIR, DumpBytecode: dumpBytecode, EmitBytecode: emitBytecode}
	if err := c.RunFile(files[0]); err != nil {
		c.handleError(err)
	}
}
