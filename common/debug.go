package common

import "fmt"

const debug = false

func DebugPrintf(format string, args ...interface{}) {
	if !debug {
		return
	}
	fmt.Printf(format, args...)
}

func DebugPrint(str ...string) {
	if !debug {
		return
	}
	fmt.Print(str)
}

func DebugPrintln(str ...string) {
	if !debug {
		return
	}
	fmt.Println(str)
}
