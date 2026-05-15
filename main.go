package main

import (
	"fmt"
	"os"
)

func main() {
	dumpIR := false
	dumpBytecode := false
	args := os.Args[1:]
	files := make([]string, 0, 1)
	for _, arg := range args {
		switch arg {
		case "--dump-ir":
			dumpIR = true
		case "--dump-bytecode":
			dumpBytecode = true
		default:
			files = append(files, arg)
		}
	}
	if len(files) != 1 {
		fmt.Println("Usage: cvm [--dump-ir|--dump-bytecode] [file]")
		return
	}
	c := &Compiler{DumpIR: dumpIR, DumpBytecode: dumpBytecode}
	if err := c.RunFile(files[0]); err != nil {
		c.handleError(err)
	}
}
