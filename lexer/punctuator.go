package lexer

import (
	"shinya.click/cvm/common"
	"shinya.click/cvm/lexer/util"
)

var punctuatorStateTable = stateTable{
	"A": []Edge{{"-", "C"}, {"+", "D"}, {"<", "E"}, {">", "H"}, {"&", "J"}, {"|", "K"}, {"AL", "L"}, {"#", "M"}, {".", "N"}},
	"B": []Edge{},
	"C": []Edge{{"CG", "G"}},
	"D": []Edge{{"DG", "G"}},
	"E": []Edge{{"<", "L"}, {"=", "G"}},
	"G": []Edge{},
	"H": []Edge{{">", "L"}, {"=", "G"}},
	"J": []Edge{{"JG", "G"}},
	"K": []Edge{{"KG", "G"}},
	"L": []Edge{{"=", "G"}},
	"M": []Edge{{"#", "G"}},
	"N": []Edge{{".", "O"}},
	"O": []Edge{{".", "G"}},
}

var punctuatorConditionTable = conditionTable{
	"-": func(b byte) bool { return b == '-' },
	"+": func(b byte) bool { return b == '+' },
	">": func(b byte) bool { return b == '>' },
	"<": func(b byte) bool { return b == '<' },
	"=": func(b byte) bool { return b == '=' },
	"&": func(b byte) bool { return b == '&' },
	"|": func(b byte) bool { return b == '|' },
	"#": func(b byte) bool { return b == '#' },
	".": func(b byte) bool { return b == '.' },
	"AB": func(b byte) bool {
		ABs := map[byte]struct{}{
			'[': {}, ']': {}, '(': {}, ')': {}, '{': {}, '}': {},
			'~': {}, '?': {}, ':': {}, ';': {}, ',': {},
		}
		_, ok := ABs[b]
		return ok
	},
	"CG": func(b byte) bool {
		CGs := map[byte]struct{}{
			'>': {}, '=': {}, '-': {},
		}
		_, ok := CGs[b]
		return ok
	},
	"DG": func(b byte) bool {
		DGs := map[byte]struct{}{
			'=': {}, '+': {},
		}
		_, ok := DGs[b]
		return ok
	},
	"JG": func(b byte) bool {
		JGs := map[byte]struct{}{
			'&': {}, '=': {},
		}
		_, ok := JGs[b]
		return ok
	},
	"KG": func(b byte) bool {
		KGs := map[byte]struct{}{
			'|': {}, '=': {},
		}
		_, ok := KGs[b]
		return ok
	},
	"AL": func(b byte) bool {
		ALs := map[byte]struct{}{
			'*': {}, '/': {}, '%': {}, '^': {}, '=': {}, '!': {},
		}
		_, ok := ALs[b]
		return ok
	},
}

var punctuatorTokenMap = map[string]common.TokenType{
	"[":   common.LEFT_BRACKETS,
	"]":   common.RIGHT_BRACKETS,
	"(":   common.LEFT_PARENTHESES,
	")":   common.RIGHT_PARENTHESES,
	"{":   common.LEFT_BRACES,
	"}":   common.RIGHT_BRACES,
	".":   common.PERIOD,
	"->":  common.ARROW,
	"++":  common.PLUS_PLUS,
	"--":  common.MINUS_MINUS,
	"&":   common.AND,
	"*":   common.ASTERISK,
	"+":   common.PLUS,
	"-":   common.MINUS,
	"~":   common.TILDE,
	"!":   common.EXCLAMATION,
	"/":   common.SLASH,
	"%":   common.PERCENT,
	"<<":  common.LEFT_SHIFT,
	">>":  common.RIGHT_SHIFT,
	"<":   common.LESS,
	">":   common.GREATER,
	"<=":  common.LESS_EQUAL,
	">=":  common.GREATER_EQUAL,
	"==":  common.EQUAL_EQUAL,
	"!=":  common.NOT_EQUAL,
	"^":   common.XOR,
	"|":   common.OR,
	"&&":  common.AND_AND,
	"||":  common.OR_OR,
	"?":   common.QUESTION,
	":":   common.COLON,
	";":   common.SEMICOLON,
	"...": common.VARIADIC,
	"=":   common.EQUAL,
	"*=":  common.MULTIPLY_EQUAL,
	"/=":  common.DIVIDE_EQUAL,
	"%=":  common.MOD_EQUAL,
	"+=":  common.PLUS_EQUAL,
	"-=":  common.MINUS_EQUAL,
	"<<=": common.LEFT_SHIFT_EQUAL,
	">>=": common.RIGHT_SHIFT_EQUAL,
	"&=":  common.AND_EQUAL,
	"^=":  common.XOR_EQUAL,
	"|=":  common.OR_EQUAL,
	",":   common.COMMA,
	"#":   common.SHARP,
	"##":  common.SHARP_SHARP,
}

func isPunctuatorPrefix(b byte) bool {
	prefix := map[byte]struct{}{
		'[': {}, ']': {}, '(': {}, ')': {}, '{': {}, '}': {}, '.': {}, '-': {},
		'+': {}, '&': {}, '*': {}, '~': {}, '!': {}, '/': {}, '%': {}, '<': {},
		'>': {}, '=': {}, '^': {}, '|': {}, '?': {}, ':': {}, ';': {}, '#': {},
	}
	_, ok := prefix[b]
	return ok
}

func punctuatorConstructor(s string, l, sc, ec int, _ state, _ interface{}) (common.Token, error) {
	tokenType, ok := punctuatorTokenMap[s]
	if !ok {
		return emptyToken, util.NewLexerError(util.ErrUnidentifiableToken, l, sc, ec, "Unknown Punctuator: %s", s)
	}
	return common.NewToken(tokenType, s, nil, l, sc, ec), nil
}

var punctuatorScanner *Scanner

func init() {
	punctuatorScanner = newPunctuatorScanner()
}

func PunctuatorScanner() *Scanner {
	return punctuatorScanner
}

func newPunctuatorScanner() *Scanner {
	return NewScannerBuilder("Punctuator").
		StateTable(punctuatorStateTable).
		ConditionTable(punctuatorConditionTable).
		TokenConstructor(punctuatorConstructor).
		StartState("A").
		EndState([]state{"B", "C", "D", "E", "G", "H", "J", "K", "L", "M", "N"}).
		Build()
}
