package lexer

import (
	"shinya.click/cvm/common"
)

var characterLiteralStateTable = stateTable{
	"A":  map[condition]state{"single_quotation": "B"},
	"B":  map[condition]state{"char_ascii": "H", "back_slash": "C"},
	"C":  map[condition]state{"x": "D", "escape_suffix": "H", "oct": "FH"},
	"D":  map[condition]state{"hex": "EH"},
	"EH": map[condition]state{"single_quotation": "N", "char_ascii": "H", "back_slash": "I", "hex": "EH"},
	"H":  map[condition]state{"single_quotation": "N", "char_ascii": "H", "back_slash": "I"},
	"FH": map[condition]state{"single_quotation": "N", "char_ascii": "H", "back_slash": "I", "oct": "GH"},
	"GH": map[condition]state{"single_quotation": "N", "char_ascii": "H", "back_slash": "I", "oct": "H"},
	"I":  map[condition]state{"x": "J", "escape_suffix": "H", "oct": "LH"},
	"J":  map[condition]state{"hex": "EH"},
	"LH": map[condition]state{"single_quotation": "N", "char_ascii": "H", "back_slash": "I", "oct": "GH"},
	"N":  map[condition]state{},
}

var characterLiteralConditionTable = conditionTable{
	"single_quotation": func(b byte) bool {
		return b == '\''
	},
	"char_ascii": func(b byte) bool {
		no_cchar := map[byte]struct{}{
			'\'': {}, '\\': {}, '\n': {},
		}
		_, ok := no_cchar[b]
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

func newCharacterLiteralScanner() *Scanner {
	return newScanner(characterLiteralStateTable,
		characterLiteralConditionTable,
		func(s string, l int) common.Token {
			return common.NewToken(common.CHARACTER, s, nil, l)
		},
		"A",
		[]state{"N"})
}
