package main

import (
	"fmt"
	"io"
	"os"
	"shinya.click/cvm/util"
)

func main() {
	if len(os.Args) != 1 {
		fmt.Println("Usage: cvm [file]")
		return
	}
	runFile(os.Args[0])
}

func runFile(filePath string) {
	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	bytes, err := io.ReadAll(file)
	if err != nil {
		panic(err)
	}
	run(string(bytes))
	if util.HadError {
		os.Exit(65)
	}
}

func run(source string) {

}
