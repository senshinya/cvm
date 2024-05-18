package main

import (
	"bufio"
	"io"
	"log"
	"os"
	"strings"
)

var terminals = map[string]struct{}{
	"IDENTIFIER":        {},
	"STRING":            {},
	"CHARACTER":         {},
	"INTEGER_CONSTANT":  {},
	"FLOATING_CONSTANT": {},
	"AUTO":              {},
	"BREAK":             {},
	"CASE":              {},
	"CHAR":              {},
	"CONST":             {},
	"CONTINUE":          {},
	"DEFAULT":           {},
	"DO":                {},
	"DOUBLE":            {},
	"ELSE":              {},
	"ENUM":              {},
	"EXTERN":            {},
	"FLOAT":             {},
	"FOR":               {},
	"GOTO":              {},
	"IF":                {},
	"INLINE":            {},
	"INT":               {},
	"LONG":              {},
	"REGISTER":          {},
	"RESTRICT":          {},
	"RETURN":            {},
	"SHORT":             {},
	"SIGNED":            {},
	"SIZEOF":            {},
	"STATIC":            {},
	"STRUCT":            {},
	"SWITCH":            {},
	"TYPEDEF":           {},
	"UNION":             {},
	"UNSIGNED":          {},
	"VOID":              {},
	"VOLATILE":          {},
	"WHILE":             {},
	"BOOL":              {},
	"COMPLEX":           {},
	"IMAGINARY":         {},
	"LEFT_BRACKETS":     {},
	"RIGHT_BRACKETS":    {},
	"LEFT_PARENTHESES":  {},
	"RIGHT_PARENTHESES": {},
	"LEFT_BRACES":       {},
	"RIGHT_BRACES":      {},
	"PERIOD":            {},
	"ARROW":             {},
	"PLUS_PLUS":         {},
	"MINUS_MINUS":       {},
	"AND":               {},
	"ASTERISK":          {},
	"PLUS":              {},
	"MINUS":             {},
	"TILDE":             {},
	"EXCLAMATION":       {},
	"SLASH":             {},
	"PERCENT":           {},
	"LEFT_SHIFT":        {},
	"RIGHT_SHIFT":       {},
	"LESS":              {},
	"GREATER":           {},
	"LESS_EQUAL":        {},
	"GREATER_EQUAL":     {},
	"EQUAL_EQUAL":       {},
	"NOT_EQUAL":         {},
	"XOR":               {},
	"OR":                {},
	"AND_AND":           {},
	"OR_OR":             {},
	"QUESTION":          {},
	"COLON":             {},
	"SEMICOLON":         {},
	"VARIADIC":          {},
	"EQUAL":             {},
	"MULTIPLY_EQUAL":    {},
	"DIVIDE_EQUAL":      {},
	"MOD_EQUAL":         {},
	"PLUS_EQUAL":        {},
	"MINUS_EQUAL":       {},
	"LEFT_SHIFT_EQUAL":  {},
	"RIGHT_SHIFT_EQUAL": {},
	"AND_EQUAL":         {},
	"XOR_EQUAL":         {},
	"OR_EQUAL":          {},
	"COMMA":             {},
	"SHARP":             {},
	"SHARP_SHARP":       {},
}

func main() {
	lines := readLines()
	nonTerminalDefs := splitNonTerminalDefs(lines)
	checkDefs(nonTerminalDefs)
}

func readLines() []string {
	file, err := os.OpenFile("./parser/phrase_structure_grammar.txt", os.O_RDONLY, 0666)
	if err != nil {
		log.Panicf("open file failed: %v", err.Error())
	}

	defer file.Close()

	reader := bufio.NewReader(file)
	var results []string
	for {
		line, _, err := reader.ReadLine()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Panicf("read line failed: %v", err.Error())
		}
		if len(line) == 0 {
			continue
		}
		results = append(results, string(line))
	}
	return results
}

func splitNonTerminalDefs(lines []string) map[string][][]string {
	nonTerminals := map[string][][]string{}
	for _, line := range lines {
		splits := strings.Split(line, ":=")
		if len(splits) != 2 {
			log.Panicf("Invalid line format: %s", line)
		}
		def := strings.TrimSpace(splits[0])
		parts := strings.Split(strings.TrimSpace(splits[1]), " ")
		nonTerminals[def] = append(nonTerminals[def], parts)
	}
	return nonTerminals
}

func checkDefs(nonTerminalDefs map[string][][]string) {
	for _, defs := range nonTerminalDefs {
		for _, parts := range defs {
			for _, part := range parts {
				if _, ok := terminals[part]; ok {
					continue
				}
				if _, ok := nonTerminalDefs[part]; !ok {
					log.Panicf("Invalid nonTerminal: %s", part)
				}
			}
		}
	}
}
