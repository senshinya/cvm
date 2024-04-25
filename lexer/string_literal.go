package lexer

import (
	"shinya.click/cvm/common"
)

var stringLiteralStateTable = stateTable{
	"A":  map[condition]state{"double_quotation": "B"},
	"B":  map[condition]state{"double_quotation": "C", "string_ascii": "B", "back_slash": "D"},
	"D":  map[condition]state{"escape_suffix": "B", "x": "E", "oct": "GB"},
	"E":  map[condition]state{"hex": "FB"},
	"GB": map[condition]state{"double_quotation": "C", "string_ascii": "B", "back_slash": "D", "oct": "HB"},
	"FB": map[condition]state{"double_quotation": "C", "string_ascii": "B", "back_slash": "D", "hex": "FB"},
	"HB": map[condition]state{"double_quotation": "C", "string_ascii": "B", "back_slash": "D", "oct": "B"},
	"C":  map[condition]state{},
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
	"escape_suffix": func(b byte) bool {
		suffixMap := map[byte]struct{}{
			'\'': {}, '"': {}, '?': {}, '\\': {}, 'a': {}, 'b': {}, 'f': {}, 'n': {}, 'r': {}, 't': {}, 'v': {},
		}
		_, ok := suffixMap[b]
		return ok
	},
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

func newStringLiteralScanner() *Scanner {
	return NewScannerBuilder().
		StateTable(stringLiteralStateTable).
		ConditionTable(stringLiteralConditionTable).
		TokenConstructor(func(s string, l int) common.Token {
			// TODO check escape range
			return common.NewToken(common.STRING, s, nil, l)
		}).
		StartState("A").
		EndState([]state{"C"}).
		Build()
}
