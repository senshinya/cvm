package util

import "fmt"

var HadError = false

func Error(line int, message string) {
	report(line, "", message)
	HadError = true
}

func report(line int, where string, message string) {
	fmt.Printf("[line %d ]: Error%s: %s\n", line, where, message)
}
