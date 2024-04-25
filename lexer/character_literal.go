package lexer

import (
	"shinya.click/cvm/common"
	"shinya.click/cvm/lexer/util"
)

var characterLiteralStateTable = stateTable{
	"A":  []Edge{{"single_quotation", "B"}},
	"B":  []Edge{{"char_ascii", "H"}, {"back_slash", "C"}},
	"C":  []Edge{{"x", "D"}, {"escape_suffix", "H"}, {"oct", "FH"}},
	"D":  []Edge{{"hex", "EH"}},
	"EH": []Edge{{"hex", "EH"}, {"single_quotation", "N"}, {"char_ascii", "H"}, {"back_slash", "I"}},
	"H":  []Edge{{"single_quotation", "N"}, {"char_ascii", "H"}, {"back_slash", "I"}},
	"FH": []Edge{{"oct", "GH"}, {"single_quotation", "N"}, {"char_ascii", "H"}, {"back_slash", "I"}},
	"GH": []Edge{{"oct", "JH"}, {"single_quotation", "N"}, {"char_ascii", "H"}, {"back_slash", "I"}},
	"JH": []Edge{{"single_quotation", "N"}, {"char_ascii", "H"}, {"back_slash", "I"}},
	"I":  []Edge{{"x", "D"}, {"escape_suffix", "H"}, {"oct", "FH"}},
	"N":  []Edge{},
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

type characterLiteralStore struct {
	last         byte
	currentBytes string
}

func characterLiteralTransferInterceptor(before, next state, char byte, store interface{}) {
	cs := store.(*characterLiteralStore)
	if (next == "N") ||
		(next == "I") ||
		(before.in([]state{"H", "EH", "FH", "GH", "JH"}) && next == "H") {
		// a character has been read!
		// check if out of range
		b := util.CheckAndUnquoteCharacterLiteral(cs.currentBytes)
		cs.currentBytes = ""
		cs.last = b
	}

	if char == '\'' && (before == "A" || next == "N") {
		return
	}
	cs.currentBytes += string(char)
}

func newCharacterLiteralScanner() *Scanner {
	return NewScannerBuilder().
		StateTable(characterLiteralStateTable).
		ConditionTable(characterLiteralConditionTable).
		TokenConstructor(func(s string, l int, store interface{}) common.Token {
			cs := store.(*characterLiteralStore)
			return common.NewToken(common.CHARACTER, s, cs.last, l)
		}).
		StartState("A").
		EndState([]state{"N"}).
		transferInterceptor(characterLiteralTransferInterceptor).
		store(&characterLiteralStore{}).
		Build()
}
