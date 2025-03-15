package common

import "fmt"

const debug = true

func DebugPrintf(format string, args ...any) {
	if !debug {
		return
	}
	fmt.Printf(format, args...)
}

func DebugPrint(str ...any) {
	if !debug {
		return
	}
	fmt.Print(str...)
}

func DebugPrintln(str ...any) {
	if !debug {
		return
	}
	fmt.Println(str...)
}
