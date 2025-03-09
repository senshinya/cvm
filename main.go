package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: cvm [file]")
		return
	}
	(&Compiler{}).RunFile(os.Args[1])
}
