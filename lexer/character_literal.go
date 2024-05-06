package lexer

import (
	"shinya.click/cvm/common"
	"shinya.click/cvm/lexer/util"
	"sync"
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
	"oct":           util.IsOctDigit,
	"hex":           util.IsHexDigit,
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

var (
	characterLiteralScanner     *Scanner
	characterLiteralScannerOnce sync.Once
)

func CharacterLiteralScanner() *Scanner {
	characterLiteralScannerOnce.Do(func() {
		characterLiteralScanner = newCharacterLiteralScanner()
	})
	characterLiteralScanner.Store(&characterLiteralStore{})
	return characterLiteralScanner
}

func newCharacterLiteralScanner() *Scanner {
	return NewScannerBuilder().
		StateTable(characterLiteralStateTable).
		ConditionTable(characterLiteralConditionTable).
		TokenConstructor(func(s string, l int, _ state, store interface{}) common.Token {
			cs := store.(*characterLiteralStore)
			return common.NewToken(common.CHARACTER, s, cs.last, l)
		}).
		StartState("A").
		EndState([]state{"N"}).
		transferInterceptor(characterLiteralTransferInterceptor).
		Build()
}
