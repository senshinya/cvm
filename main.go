package main

import (
	"fmt"
	"os"
)

func main() {
	dumpIR := false
	args := os.Args[1:]
	if len(args) > 0 && args[0] == "--dump-ir" {
		dumpIR = true
		args = args[1:]
	}
	if len(args) != 1 {
		fmt.Println("Usage: cvm [--dump-ir] [file]")
		return
	}
	(&Compiler{DumpIR: dumpIR}).RunFile(args[0])
}
