package lexer

import (
	"shinya.click/cvm/common"
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
	"escape_suffix": common.IsSimpleEscapeSuffix,
	"oct":           common.IsOctDigit,
	"hex":           common.IsHexDigit,
}

type characterLiteralStore struct {
	last         byte
	currentBytes string
}

func characterLiteralTransferInterceptor(before, next state, char byte, store interface{}, l, _, ec int) error {
	cs := store.(*characterLiteralStore)
	if (next == "N") ||
		(next == "I") ||
		(before.in([]state{"H", "EH", "FH", "GH", "JH"}) && next == "H") {
		// a character has been read!
		// check if out of range
		b, err := common.CheckAndUnquoteCharacterLiteral(cs.currentBytes)
		if err != nil {
			return common.NewLexerError(common.ErrInvalidCharacter, l, ec-len(cs.currentBytes)+1, ec, err.Error())
		}
		cs.currentBytes = ""
		cs.last = b
	}

	if char == '\'' && (before == "A" || next == "N") {
		return nil
	}
	cs.currentBytes += string(char)
	return nil
}

var characterLiteralScanner *Scanner

func init() {
	characterLiteralScanner = newCharacterLiteralScanner()
}

func CharacterLiteralScanner() *Scanner {
	characterLiteralScanner.Store(&characterLiteralStore{})
	return characterLiteralScanner
}

func newCharacterLiteralScanner() *Scanner {
	return NewScannerBuilder("Character Literal").
		StateTable(characterLiteralStateTable).
		ConditionTable(characterLiteralConditionTable).
		TokenConstructor(func(s string, l, sc, ec int, _ state, store interface{}) (common.Token, error) {
			cs := store.(*characterLiteralStore)
			return common.NewToken(common.CHARACTER, s, cs.last, l, sc, ec), nil
		}).
		StartState("A").
		EndState([]state{"N"}).
		transferInterceptor(characterLiteralTransferInterceptor).
		Build()
}
