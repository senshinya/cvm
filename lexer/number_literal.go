package lexer

import (
	"fmt"
	"shinya.click/cvm/common"
	"shinya.click/cvm/lexer/util"
	"sync"
)

var numberLiteralStateTable = stateTable{
	"A": []Edge{{"nozero_digit", "B"}, {"0", "N"}, {".", "W"}},
	"B": []Edge{{"digit", "B"}, {"L/l", "F"}, {"U/u", "C"}, {".", "I"}, {"E/e", "J"}},
	"C": []Edge{{"L/l", "D"}},
	"D": []Edge{{"L/l", "E"}},
	"E": []Edge{},
	"F": []Edge{{"L/l", "G"}, {"U/u", "H"}},
	"G": []Edge{{"U/u", "H"}},
	"H": []Edge{},
	"I": []Edge{{"digit", "I"}, {"E/e", "J"}, {"F/L/f/l", "M"}},
	"J": []Edge{{"digit", "L"}, {"+/-", "K"}},
	"K": []Edge{{"digit", "L"}},
	"L": []Edge{{"digit", "L"}, {"F/L/f/l", "M"}},
	"M": []Edge{},
	"N": []Edge{{"oct_digit", "O"}, {"nooct_digit", "P"}, {".", "I"}, {"E/e", "J"}, {"X/x", "Q"}},
	"O": []Edge{{"oct_digit", "O"}, {"L/l", "F"}, {"U/u", "C"}, {".", "I"}, {"nooct_digit", "P"}, {"E/e", "J"}},
	"P": []Edge{{"digit", "P"}, {".", "I"}, {"E/e", "J"}},
	"Q": []Edge{{"hex_digit", "R"}, {".", "X"}},
	"R": []Edge{{"hex_digit", "R"}, {"L/l", "F"}, {"U/u", "C"}, {".", "S"}, {"P/p", "T"}},
	"S": []Edge{{"hex_digit", "S"}, {"P/p", "T"}},
	"T": []Edge{{"digit", "V"}, {"+/-", "U"}},
	"U": []Edge{{"digit", "V"}},
	"V": []Edge{{"digit", "V"}, {"F/L/f/l", "M"}},
	"W": []Edge{{"digit", "I"}},
	"X": []Edge{{"hex_digit", "S"}},
}

var numberLiteralConditionTable = conditionTable{
	"nozero_digit": func(b byte) bool {
		nozeros := map[byte]struct{}{
			'1': {}, '2': {}, '3': {}, '4': {}, '5': {}, '6': {}, '7': {}, '8': {}, '9': {},
		}
		_, ok := nozeros[b]
		return ok
	},
	"0":         func(b byte) bool { return b == '0' },
	"digit":     util.IsDigit,
	"L/l":       func(b byte) bool { return b == 'L' || b == 'l' },
	"U/u":       func(b byte) bool { return b == 'U' || b == 'u' },
	".":         func(b byte) bool { return b == '.' },
	"E/e":       func(b byte) bool { return b == 'E' || b == 'e' },
	"+/-":       func(b byte) bool { return b == '+' || b == '-' },
	"F/L/f/l":   func(b byte) bool { return b == 'F' || b == 'L' || b == 'f' || b == 'l' },
	"oct_digit": util.IsOctDigit,
	"nooct_digit": func(b byte) bool {
		return util.IsDigit(b) && !util.IsOctDigit(b)
	},
	"X/x":       func(b byte) bool { return b == 'X' || b == 'x' },
	"P/p":       func(b byte) bool { return b == 'P' || b == 'p' },
	"hex_digit": util.IsHexDigit,
}

var integerEndStates = map[state]struct{}{
	"B": {}, "C": {}, "D": {}, "E": {}, "F": {}, "G": {}, "H": {},
	"N": {}, "O": {}, "R": {},
}

var floatEndStates = map[state]struct{}{
	"I": {}, "L": {}, "M": {}, "V": {},
}

func numberLiteralConstructor(s string, l int, endState state, _ interface{}) common.Token {
	if _, ok := integerEndStates[endState]; ok {
		return common.NewToken(common.INTEGER_CONSTANT, s, nil, l)
	}
	if _, ok := floatEndStates[endState]; ok {
		return common.NewToken(common.FLOATING_CONSTANT, s, nil, l)
	}
	panic(fmt.Sprintf("Unknown number literal type: %s", s))
}

var numberLiteralEndStates = []state{
	"B", "C", "D", "E", "F", "G", "H", "I", "L", "M", "N", "O", "R", "V",
}

var (
	numberLiteralScanner     *Scanner
	numberLiteralScannerOnce sync.Once
)

func NumberLiteralScanner() *Scanner {
	numberLiteralScannerOnce.Do(func() {
		numberLiteralScanner = newNumberLiteral()
	})
	return numberLiteralScanner
}

func newNumberLiteral() *Scanner {
	return NewScannerBuilder().
		StateTable(numberLiteralStateTable).
		ConditionTable(numberLiteralConditionTable).
		TokenConstructor(numberLiteralConstructor).
		StartState("A").
		EndState(numberLiteralEndStates).
		Build()
}
