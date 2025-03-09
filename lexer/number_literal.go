package lexer

import (
	"shinya.click/cvm/entity"
	"strconv"
	"strings"
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
	"digit":     IsDigit,
	"L/l":       func(b byte) bool { return b == 'L' || b == 'l' },
	"U/u":       func(b byte) bool { return b == 'U' || b == 'u' },
	".":         func(b byte) bool { return b == '.' },
	"E/e":       func(b byte) bool { return b == 'E' || b == 'e' },
	"+/-":       func(b byte) bool { return b == '+' || b == '-' },
	"F/L/f/l":   func(b byte) bool { return b == 'F' || b == 'L' || b == 'f' || b == 'l' },
	"oct_digit": IsOctDigit,
	"nooct_digit": func(b byte) bool {
		return IsDigit(b) && !IsOctDigit(b)
	},
	"X/x":       func(b byte) bool { return b == 'X' || b == 'x' },
	"P/p":       func(b byte) bool { return b == 'P' || b == 'p' },
	"hex_digit": IsHexDigit,
}

var integerEndStates = map[state]struct{}{
	"B": {}, "C": {}, "D": {}, "E": {}, "F": {}, "G": {}, "H": {},
	"N": {}, "O": {}, "R": {},
}

var floatEndStates = map[state]struct{}{
	"I": {}, "L": {}, "M": {}, "V": {},
}

func numberLiteralConstructor(s string, l, sc, ec int, endState state, _ interface{}) (entity.Token, error) {
	if _, ok := integerEndStates[endState]; ok {
		return constructIntegerToken(s, l, sc, ec)
	}
	if _, ok := floatEndStates[endState]; ok {
		return constructFloatToken(s, l, sc, ec)
	}
	return emptyToken, UnidentifiedToken(l, sc)
}

func constructIntegerToken(rawStr string, l, sc, ec int) (entity.Token, error) {
	s := rawStr

	// check integer suffix
	long, unsigned := false, false
	if strings.Contains(s, "ll") || strings.Contains(s, "LL") {
		long = true
	}
	if strings.Contains(s, "u") || strings.Contains(s, "U") {
		unsigned = true
	}
	s = strings.ReplaceAll(s, "l", "")
	s = strings.ReplaceAll(s, "L", "")
	s = strings.ReplaceAll(s, "u", "")
	s = strings.ReplaceAll(s, "U", "")

	var (
		raw uint64
		err error
	)
	switch {
	case strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X"):
		// hex
		s = s[2:]
		raw, err = strconv.ParseUint(s, 16, 64)
	case strings.HasPrefix(s, "0") && len(s) > 1:
		// oct
		s = s[1:]
		raw, err = strconv.ParseUint(s, 8, 64)
	default:
		// dec
		raw, err = strconv.ParseUint(s, 10, 64)
	}
	if err != nil {
		return emptyToken, UnidentifiedToken(l, sc)
	}
	if long && unsigned {
		return entity.NewToken(entity.INTEGER_CONSTANT, rawStr, raw, l, sc, ec), nil
	}
	if !long && unsigned {
		return entity.NewToken(entity.INTEGER_CONSTANT, rawStr, uint32(raw), l, sc, ec), nil
	}
	if long && !unsigned {
		return entity.NewToken(entity.INTEGER_CONSTANT, rawStr, int64(raw), l, sc, ec), nil
	}
	// !long && !unsigned
	return entity.NewToken(entity.INTEGER_CONSTANT, rawStr, int32(raw), l, sc, ec), nil
}

func constructFloatToken(rawStr string, l, sc, ec int) (entity.Token, error) {
	s := rawStr

	// check float suffix
	float := false
	if strings.HasSuffix(s, "f") || strings.HasSuffix(s, "F") {
		float = true
		s = s[:len(s)-1]
	}
	if strings.Contains(s, "l") || strings.Contains(s, "L") {
		float = true
		s = s[:len(s)-1]
	}

	raw, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return emptyToken, UnidentifiedToken(l, sc)
	}
	if float {
		return entity.NewToken(entity.FLOATING_CONSTANT, rawStr, float32(raw), l, sc, ec), nil
	}
	return entity.NewToken(entity.FLOATING_CONSTANT, rawStr, raw, l, sc, ec), nil
}

var numberLiteralEndStates = []state{
	"B", "C", "D", "E", "F", "G", "H", "I", "L", "M", "N", "O", "R", "V",
}

var numberLiteralScanner *Scanner

func init() {
	numberLiteralScanner = newNumberLiteral()
}

func NumberLiteralScanner() *Scanner {
	return numberLiteralScanner
}

func newNumberLiteral() *Scanner {
	return NewScannerBuilder("Number Literal").
		StateTable(numberLiteralStateTable).
		ConditionTable(numberLiteralConditionTable).
		TokenConstructor(numberLiteralConstructor).
		StartState("A").
		EndState(numberLiteralEndStates).
		Build()
}
