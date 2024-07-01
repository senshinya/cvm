package main

import (
	"fmt"
	"io"
	"os"
	"shinya.click/cvm/common"
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
		panic(common.NewInitError(common.ErrOpenFile, err.Error()))
	}
	defer file.Close()
	bytes, err := io.ReadAll(file)
	if err != nil {
		panic(common.NewInitError(common.ErrReadFile, err.Error()))
	}
	run(string(bytes))
}

func run(source string) {

}
