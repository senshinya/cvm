package lexer

import (
	"shinya.click/cvm/common"
	"shinya.click/cvm/lexer/util"
)

var identifierStateTable = stateTable{
	"A": map[condition]state{"letter_": "B"},
	"B": map[condition]state{"letter_": "B", "digit": "B"},
}

var identifierConditionTable = conditionTable{
	"letter_": util.IsLetter_,
	"digit": func(b byte) bool {
		return b >= '0' && b <= '9'
	},
}

func identifierConstructor(s string, l int) common.Token {
	// TODO add keyword constructor
	return common.NewToken(common.IDENTIFIER, s, nil, l)
}

func newIdentifierScanner() *Scanner {
	return newScanner(identifierStateTable,
		identifierConditionTable,
		identifierConstructor,
		"A",
		[]state{"B"})
}
