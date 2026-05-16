package lexer

import (
	"shinya.click/cvm/entity"
)

var stringLiteralStateTable = stateTable{
	"A":  []Edge{{"double_quotation", "B"}},
	"B":  []Edge{{"double_quotation", "C"}, {"string_ascii", "B"}, {"back_slash", "D"}},
	"D":  []Edge{{"escape_suffix", "B"}, {"x", "E"}, {"oct", "GB"}},
	"E":  []Edge{{"hex", "FB"}},
	"GB": []Edge{{"oct", "HB"}, {"double_quotation", "C"}, {"string_ascii", "B"}, {"back_slash", "D"}},
	"FB": []Edge{{"hex", "FB"}, {"double_quotation", "C"}, {"string_ascii", "B"}, {"back_slash", "D"}},
	"HB": []Edge{{"oct", "IB"}, {"double_quotation", "C"}, {"string_ascii", "B"}, {"back_slash", "D"}},
	"IB": []Edge{{"double_quotation", "C"}, {"string_ascii", "B"}, {"back_slash", "D"}},
	"C":  []Edge{},
}

var wideStringLiteralStateTable = stateTable{
	"A":  []Edge{{"wide_prefix", "W"}},
	"W":  []Edge{{"double_quotation", "B"}},
	"B":  []Edge{{"double_quotation", "C"}, {"string_ascii", "B"}, {"back_slash", "D"}},
	"D":  []Edge{{"escape_suffix", "B"}, {"x", "E"}, {"oct", "GB"}},
	"E":  []Edge{{"hex", "FB"}},
	"GB": []Edge{{"oct", "HB"}, {"double_quotation", "C"}, {"string_ascii", "B"}, {"back_slash", "D"}},
	"FB": []Edge{{"hex", "FB"}, {"double_quotation", "C"}, {"string_ascii", "B"}, {"back_slash", "D"}},
	"HB": []Edge{{"oct", "IB"}, {"double_quotation", "C"}, {"string_ascii", "B"}, {"back_slash", "D"}},
	"IB": []Edge{{"double_quotation", "C"}, {"string_ascii", "B"}, {"back_slash", "D"}},
	"C":  []Edge{},
}

var stringLiteralConditionTable = conditionTable{
	"wide_prefix": func(b byte) bool {
		return b == 'L'
	},
	"double_quotation": func(b byte) bool {
		return b == '"'
	},
	"string_ascii": func(b byte) bool {
		no_schar := map[byte]struct{}{
			'"': {}, '\\': {}, '\n': {},
		}
		_, ok := no_schar[b]
		return !ok
	},
	"back_slash": func(b byte) bool {
		return b == '\\'
	},
	"x": func(b byte) bool {
		return b == 'x'
	},
	"escape_suffix": IsSimpleEscapeSuffix,
	"oct":           IsOctDigit,
	"hex":           IsHexDigit,
}

type stringLiteralStore struct {
	result       string
	currentBytes string
}

func stringLiteralTransferInterceptor(before, next state, char byte, store interface{}, l, _, ec int) error {
	cs := store.(*stringLiteralStore)
	//fmt.Printf("before: %s, next: %s, char %c\n", before, next, char)
	if len(cs.currentBytes) != 0 &&
		((next == "C") ||
			(next == "D") ||
			(before.in([]state{"B", "FB", "GB", "HB", "IB"}) && next == "B")) {
		// a character has been read!
		// check if out of range
		b, err := CheckAndUnquoteCharacterInString(cs.currentBytes)
		if err != nil {
			return InvalidCharacter(l, ec-len(cs.currentBytes)+1)
		}
		cs.currentBytes = ""
		cs.result += string(b)
	}

	if char == 'L' && before == "A" && next == "W" {
		return nil
	}
	if char == '"' && (before == "A" || before == "W" || next == "B") {
		return nil
	}
	cs.currentBytes += string(char)
	return nil
}

var stringLiteralScanner *Scanner
var wideStringLiteralScanner *Scanner

func init() {
	stringLiteralScanner = newStringLiteralScanner()
	wideStringLiteralScanner = newWideStringLiteralScanner()
}

func StringLiteralScanner() *Scanner {
	stringLiteralScanner.Store(&stringLiteralStore{})
	return stringLiteralScanner
}

func WideStringLiteralScanner() *Scanner {
	wideStringLiteralScanner.Store(&stringLiteralStore{})
	return wideStringLiteralScanner
}

func newStringLiteralScanner() *Scanner {
	return NewScannerBuilder("String Literal").
		StateTable(stringLiteralStateTable).
		ConditionTable(stringLiteralConditionTable).
		TokenConstructor(func(s string, l, sc, ec int, _ state, store interface{}) (entity.Token, error) {
			cs := store.(*stringLiteralStore)
			return entity.NewToken(entity.STRING, s, cs.result, l, sc, ec), nil
		}).
		StartState("A").
		EndState([]state{"C"}).
		transferInterceptor(stringLiteralTransferInterceptor).
		Build()
}

func newWideStringLiteralScanner() *Scanner {
	return NewScannerBuilder("Wide String Literal").
		StateTable(wideStringLiteralStateTable).
		ConditionTable(stringLiteralConditionTable).
		TokenConstructor(func(s string, l, sc, ec int, _ state, store interface{}) (entity.Token, error) {
			cs := store.(*stringLiteralStore)
			return entity.NewToken(entity.STRING, s, cs.result, l, sc, ec), nil
		}).
		StartState("A").
		EndState([]state{"C"}).
		transferInterceptor(stringLiteralTransferInterceptor).
		Build()
}
