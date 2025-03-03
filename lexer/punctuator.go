package lexer

import (
	"shinya.click/cvm/common"
	"shinya.click/cvm/entity"
)

var punctuatorStateTable = stateTable{
	"A": []Edge{{"AB", "B"}, {"-", "C"}, {"+", "D"}, {"<", "E"}, {">", "H"}, {"&", "J"}, {"|", "K"}, {"AL", "L"}, {"#", "M"}, {".", "N"}},
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

var punctuatorTokenMap = map[string]entity.TokenType{
	"[":   entity.LEFT_BRACKETS,
	"]":   entity.RIGHT_BRACKETS,
	"(":   entity.LEFT_PARENTHESES,
	")":   entity.RIGHT_PARENTHESES,
	"{":   entity.LEFT_BRACES,
	"}":   entity.RIGHT_BRACES,
	".":   entity.PERIOD,
	"->":  entity.ARROW,
	"++":  entity.PLUS_PLUS,
	"--":  entity.MINUS_MINUS,
	"&":   entity.AND,
	"*":   entity.ASTERISK,
	"+":   entity.PLUS,
	"-":   entity.MINUS,
	"~":   entity.TILDE,
	"!":   entity.EXCLAMATION,
	"/":   entity.SLASH,
	"%":   entity.PERCENT,
	"<<":  entity.LEFT_SHIFT,
	">>":  entity.RIGHT_SHIFT,
	"<":   entity.LESS,
	">":   entity.GREATER,
	"<=":  entity.LESS_EQUAL,
	">=":  entity.GREATER_EQUAL,
	"==":  entity.EQUAL_EQUAL,
	"!=":  entity.NOT_EQUAL,
	"^":   entity.XOR,
	"|":   entity.OR,
	"&&":  entity.AND_AND,
	"||":  entity.OR_OR,
	"?":   entity.QUESTION,
	":":   entity.COLON,
	";":   entity.SEMICOLON,
	"...": entity.VARIADIC,
	"=":   entity.EQUAL,
	"*=":  entity.MULTIPLY_EQUAL,
	"/=":  entity.DIVIDE_EQUAL,
	"%=":  entity.MOD_EQUAL,
	"+=":  entity.PLUS_EQUAL,
	"-=":  entity.MINUS_EQUAL,
	"<<=": entity.LEFT_SHIFT_EQUAL,
	">>=": entity.RIGHT_SHIFT_EQUAL,
	"&=":  entity.AND_EQUAL,
	"^=":  entity.XOR_EQUAL,
	"|=":  entity.OR_EQUAL,
	",":   entity.COMMA,
	"#":   entity.SHARP,
	"##":  entity.SHARP_SHARP,
}

func isPunctuatorPrefix(b byte) bool {
	prefix := map[byte]struct{}{
		'[': {}, ']': {}, '(': {}, ')': {}, '{': {}, '}': {}, '.': {}, '-': {},
		'+': {}, '&': {}, '*': {}, '~': {}, '!': {}, '/': {}, '%': {}, '<': {},
		'>': {}, '=': {}, '^': {}, '|': {}, '?': {}, ':': {}, ';': {}, '#': {},
		',': {},
	}
	_, ok := prefix[b]
	return ok
}

func punctuatorConstructor(s string, l, sc, ec int, _ state, _ interface{}) (entity.Token, error) {
	tokenType, ok := punctuatorTokenMap[s]
	if !ok {
		return emptyToken, common.NewLexerError(common.ErrUnidentifiableToken, l, sc, ec, "Unknown Punctuator: %s", s)
	}
	return entity.NewToken(tokenType, s, nil, l, sc, ec), nil
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
