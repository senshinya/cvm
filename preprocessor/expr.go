package preprocessor

import (
	"math"
	"strconv"
	"strings"
)

const unsignedIfSentinel int64 = math.MinInt64

// C99 preprocessing expressions are specified in intmax_t/uintmax_t terms.
// This first evaluator uses int64 for the initial C99 gate; Task 8 must either
// confirm GCC cases do not require unsigned wraparound yet or add uintmax_t
// evaluation before importing UINT64_MAX/SIZE_MAX #if cases.
func (pp *preprocessor) evalIfExpression(tokens []PPToken) (int64, error) {
	expanded := pp.expandIfTokens(dropNewlines(tokens))
	parser := ifExprParser{tokens: expanded}
	return parser.parseOr()
}

func (pp *preprocessor) expandIfTokens(tokens []PPToken) []PPToken {
	var out []PPToken
	for i := 0; i < len(tokens); i++ {
		tok := tokens[i]
		if tok.Kind == PPIdentifier && tok.Lexeme == "defined" {
			if i+1 < len(tokens) && tokens[i+1].Lexeme == "(" && i+2 < len(tokens) {
				name := tokens[i+2].Lexeme
				_, ok := pp.macros.Lookup(name)
				out = append(out, boolNumber(ok, tok))
				for i < len(tokens) && tokens[i].Lexeme != ")" {
					i++
				}
				continue
			}
			if i+1 < len(tokens) {
				_, ok := pp.macros.Lookup(tokens[i+1].Lexeme)
				out = append(out, boolNumber(ok, tok))
				i++
				continue
			}
		}
		if tok.Kind == PPIdentifier {
			if macro, ok := pp.macros.Lookup(tok.Lexeme); ok && macro.Kind == MacroObject {
				out = append(out, macro.Replacement...)
				continue
			}
			out = append(out, PPToken{Kind: PPNumber, Lexeme: "0", Location: tok.Location})
			continue
		}
		out = append(out, tok)
	}
	return out
}

func boolNumber(ok bool, loc PPToken) PPToken {
	if ok {
		return PPToken{Kind: PPNumber, Lexeme: "1", Location: loc.Location}
	}
	return PPToken{Kind: PPNumber, Lexeme: "0", Location: loc.Location}
}

type ifExprParser struct {
	tokens []PPToken
	pos    int
}

func (p *ifExprParser) parseOr() (int64, error) {
	left, err := p.parseAnd()
	if err != nil {
		return 0, err
	}
	for p.match("||") {
		right, err := p.parseAnd()
		if err != nil {
			return 0, err
		}
		if left != 0 || right != 0 {
			left = 1
		} else {
			left = 0
		}
	}
	return left, nil
}

func (p *ifExprParser) parseAnd() (int64, error) {
	left, err := p.parseBitOr()
	if err != nil {
		return 0, err
	}
	for p.match("&&") {
		right, err := p.parseBitOr()
		if err != nil {
			return 0, err
		}
		if left != 0 && right != 0 {
			left = 1
		} else {
			left = 0
		}
	}
	return left, nil
}

func (p *ifExprParser) parseBitOr() (int64, error) {
	left, err := p.parseBitXor()
	if err != nil {
		return 0, err
	}
	for p.match("|") {
		right, err := p.parseBitXor()
		if err != nil {
			return 0, err
		}
		left |= right
	}
	return left, nil
}

func (p *ifExprParser) parseBitXor() (int64, error) {
	left, err := p.parseBitAnd()
	if err != nil {
		return 0, err
	}
	for p.match("^") {
		right, err := p.parseBitAnd()
		if err != nil {
			return 0, err
		}
		left ^= right
	}
	return left, nil
}

func (p *ifExprParser) parseBitAnd() (int64, error) {
	left, err := p.parseEquality()
	if err != nil {
		return 0, err
	}
	for p.match("&") {
		right, err := p.parseEquality()
		if err != nil {
			return 0, err
		}
		left &= right
	}
	return left, nil
}

func (p *ifExprParser) parseEquality() (int64, error) {
	left, err := p.parseRel()
	if err != nil {
		return 0, err
	}
	for {
		switch {
		case p.match("=="):
			right, err := p.parseRel()
			if err != nil {
				return 0, err
			}
			left = truth(compareIfValues(left, right) == 0)
		case p.match("!="):
			right, err := p.parseRel()
			if err != nil {
				return 0, err
			}
			left = truth(compareIfValues(left, right) != 0)
		default:
			return left, nil
		}
	}
}

func (p *ifExprParser) parseRel() (int64, error) {
	left, err := p.parseShift()
	if err != nil {
		return 0, err
	}
	for {
		switch {
		case p.match("<"):
			right, err := p.parseShift()
			if err != nil {
				return 0, err
			}
			left = truth(compareIfValues(left, right) < 0)
		case p.match("<="):
			right, err := p.parseShift()
			if err != nil {
				return 0, err
			}
			left = truth(compareIfValues(left, right) <= 0)
		case p.match(">"):
			right, err := p.parseShift()
			if err != nil {
				return 0, err
			}
			left = truth(compareIfValues(left, right) > 0)
		case p.match(">="):
			right, err := p.parseShift()
			if err != nil {
				return 0, err
			}
			left = truth(compareIfValues(left, right) >= 0)
		default:
			return left, nil
		}
	}
}

func (p *ifExprParser) parseShift() (int64, error) {
	left, err := p.parseAdd()
	if err != nil {
		return 0, err
	}
	for {
		switch {
		case p.match("<<"):
			right, err := p.parseAdd()
			if err != nil {
				return 0, err
			}
			left <<= right
		case p.match(">>"):
			right, err := p.parseAdd()
			if err != nil {
				return 0, err
			}
			left >>= right
		default:
			return left, nil
		}
	}
}

func (p *ifExprParser) parseAdd() (int64, error) {
	left, err := p.parseMul()
	if err != nil {
		return 0, err
	}
	for {
		switch {
		case p.match("+"):
			right, err := p.parseMul()
			if err != nil {
				return 0, err
			}
			left += right
		case p.match("-"):
			right, err := p.parseMul()
			if err != nil {
				return 0, err
			}
			left -= right
		default:
			return left, nil
		}
	}
}

func (p *ifExprParser) parseMul() (int64, error) {
	left, err := p.parseUnary()
	if err != nil {
		return 0, err
	}
	for {
		switch {
		case p.match("*"):
			right, err := p.parseUnary()
			if err != nil {
				return 0, err
			}
			left *= right
		case p.match("/"):
			right, err := p.parseUnary()
			if err != nil {
				return 0, err
			}
			left /= right
		case p.match("%"):
			right, err := p.parseUnary()
			if err != nil {
				return 0, err
			}
			left %= right
		default:
			return left, nil
		}
	}
}

func (p *ifExprParser) parseUnary() (int64, error) {
	switch {
	case p.match("+"):
		return p.parseUnary()
	case p.match("-"):
		v, err := p.parseUnary()
		return -v, err
	case p.match("!"):
		v, err := p.parseUnary()
		return truth(v == 0), err
	case p.match("~"):
		v, err := p.parseUnary()
		return ^v, err
	default:
		return p.parsePrimary()
	}
}

func (p *ifExprParser) parsePrimary() (int64, error) {
	if p.match("(") {
		v, err := p.parseOr()
		if err != nil {
			return 0, err
		}
		p.match(")")
		return v, nil
	}
	if p.pos >= len(p.tokens) {
		return 0, nil
	}
	tok := p.tokens[p.pos]
	p.pos++
	if tok.Kind != PPNumber {
		return 0, nil
	}
	return parseIfNumber(tok.Lexeme)
}

func (p *ifExprParser) match(lexeme string) bool {
	if p.pos >= len(p.tokens) || p.tokens[p.pos].Lexeme != lexeme {
		return false
	}
	p.pos++
	return true
}

func parseIfNumber(s string) (int64, error) {
	clean := strings.TrimRight(s, "uUlL")
	if clean == "" {
		return 0, nil
	}
	v, err := strconv.ParseInt(clean, 0, 64)
	if err == nil {
		return v, nil
	}
	u, uerr := strconv.ParseUint(clean, 0, 64)
	if uerr != nil {
		return 0, err
	}
	if u > math.MaxInt64 {
		return unsignedIfSentinel, nil
	}
	return int64(u), nil
}

func compareIfValues(left, right int64) int {
	leftHuge := left == unsignedIfSentinel
	rightHuge := right == unsignedIfSentinel
	switch {
	case leftHuge && rightHuge:
		return 0
	case leftHuge:
		return 1
	case rightHuge:
		return -1
	case left < right:
		return -1
	case left > right:
		return 1
	default:
		return 0
	}
}

func truth(ok bool) int64 {
	if ok {
		return 1
	}
	return 0
}
