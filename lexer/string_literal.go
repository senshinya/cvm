package lexer

import (
	"fmt"
	"shinya.click/cvm/common"
	"shinya.click/cvm/lexer/util"
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

var stringLiteralConditionTable = conditionTable{
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
	"escape_suffix": util.IsSimpleEscapeSuffix,
	"oct": func(b byte) bool {
		octs := map[byte]struct{}{
			'0': {}, '1': {}, '2': {}, '3': {}, '4': {}, '5': {}, '6': {}, '7': {},
		}
		_, ok := octs[b]
		return ok
	},
	"hex": func(b byte) bool {
		hexes := map[byte]struct{}{
			'0': {}, '1': {}, '2': {}, '3': {}, '4': {}, '5': {}, '6': {}, '7': {}, '8': {}, '9': {},
			'a': {}, 'b': {}, 'c': {}, 'd': {}, 'e': {}, 'f': {},
			'A': {}, 'B': {}, 'C': {}, 'D': {}, 'E': {}, 'F': {},
		}
		_, ok := hexes[b]
		return ok
	},
}

type stringLiteralStore struct {
	result       string
	currentBytes string
}

func stringLiteralTransferInterceptor(before, next state, char byte, store interface{}) {
	cs := store.(*stringLiteralStore)
	fmt.Printf("before: %s, next: %s, char %c\n", before, next, char)
	if len(cs.currentBytes) != 0 &&
		((next == "C") ||
			(next == "D") ||
			(before.in([]state{"B", "FB", "GB", "HB", "IB"}) && next == "B")) {
		// a character has been read!
		// check if out of range
		b := util.CheckAndUnquoteCharacterInString(cs.currentBytes)
		cs.currentBytes = ""
		cs.result += string(b)
	}

	if char == '"' && (before == "A" || next == "B") {
		return
	}
	cs.currentBytes += string(char)
}

func newStringLiteralScanner() *Scanner {
	return NewScannerBuilder().
		StateTable(stringLiteralStateTable).
		ConditionTable(stringLiteralConditionTable).
		TokenConstructor(func(s string, l int, store interface{}) common.Token {
			cs := store.(*stringLiteralStore)
			return common.NewToken(common.STRING, s, cs.result, l)
		}).
		StartState("A").
		EndState([]state{"C"}).
		transferInterceptor(stringLiteralTransferInterceptor).
		store(&stringLiteralStore{}).
		Build()
}
