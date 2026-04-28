package preprocessor

import (
	"strconv"
	"strings"

	"shinya.click/cvm/entity"
)

type macroArg struct {
	raw      []PPToken
	expanded []PPToken
}

func (pp *preprocessor) expand(tokens []PPToken) ([]PPToken, error) {
	p := NewTokenPreprocessor(pp, tokens)
	var out []PPToken
	for {
		tok, err := p.Lex()
		if err != nil {
			return nil, err
		}
		if tok.Kind == PPEOF {
			break
		}
		if tok.Kind == PPNewline || tok.Kind == PPPadding {
			continue
		}
		if tok.Kind == PPIdentifier && tok.Lexeme == "_Pragma" {
			if consumed, err := p.consumePragma(); err != nil || consumed {
				if err != nil {
					return nil, err
				}
				continue
			}
		}
		out = append(out, tok)
	}
	return out, nil
}

func (p *Preprocessor) expandIdentifier(tok PPToken) (PPToken, bool, error) {
	if tok.DisableExpand {
		return tok, false, nil
	}
	if tok.Lexeme == "__FILE__" {
		loc := p.pp.sm.DisplayLocation(tok.Location)
		return PPToken{Kind: PPString, Lexeme: strconv.Quote(loc.File), Location: tok.Location}, false, nil
	}
	if tok.Lexeme == "__LINE__" {
		loc := p.pp.sm.DisplayLocation(tok.Location)
		return PPToken{Kind: PPNumber, Lexeme: strconv.Itoa(loc.Line), Location: tok.Location}, false, nil
	}
	macro, ok := p.pp.macros.Lookup(tok.Lexeme)
	if !ok {
		return tok, false, nil
	}
	if macro.Disabled {
		// 蓝漆规则：宏在禁用期间再次遇到同名 token，该 token 永久标记为不可展开。
		tok.DisableExpand = true
		return tok, false, nil
	}
	if macro.Kind == MacroFunction {
		consumed, ok, err := p.readFunctionInvocation()
		if err != nil {
			return tok, false, err
		}
		if !ok {
			p.unreadTokens(consumed)
			return tok, false, nil
		}
		args, err := p.pp.buildMacroArgs(macro, consumed[1:len(consumed)-1])
		if err != nil {
			return tok, false, err
		}
		replacement, err := p.pp.substitute(macro, args, tok.Location)
		if err != nil {
			return tok, false, err
		}
		p.pushMacro(macro, replacement)
		return PPToken{}, true, nil
	}
	p.pushMacro(macro, cloneTokens(macro.Replacement))
	return PPToken{}, true, nil
}

func (p *Preprocessor) pushMacro(macro *Macro, replacement []PPToken) {
	// 展开替换列表期间禁用当前宏，防止递归展开；遇到的同名 token 会被蓝漆标记。
	macro.Disabled = true
	p.push(&MacroTokenSource{tokens: replacement, macro: macro})
}

func (p *Preprocessor) readFunctionInvocation() ([]PPToken, bool, error) {
	var consumed []PPToken
	for {
		tok, err := p.readRaw()
		if err != nil {
			return nil, false, err
		}
		if tok.Kind == PPEOF {
			return consumed, false, nil
		}
		consumed = append(consumed, tok)
		if tok.Kind == PPPadding || tok.Kind == PPNewline {
			continue
		}
		if tok.Kind != PPPunctuator || tok.Lexeme != "(" {
			return consumed, false, nil
		}
		break
	}
	depth := 1
	for depth > 0 {
		tok, err := p.readRaw()
		if err != nil {
			return nil, false, err
		}
		if tok.Kind == PPEOF {
			return nil, false, ppError(consumed[0].Location, "unterminated macro invocation")
		}
		consumed = append(consumed, tok)
		if tok.Kind == PPPunctuator {
			switch tok.Lexeme {
			case "(":
				depth++
			case ")":
				depth--
			}
		}
	}
	return consumed, true, nil
}

func (pp *preprocessor) buildMacroArgs(m *Macro, tokens []PPToken) ([]macroArg, error) {
	raw := splitMacroArgs(tokens)
	if len(raw) == 1 && len(raw[0]) == 0 && len(m.Params) == 0 {
		raw = nil
	}
	if m.Variadic {
		fixed := len(m.Params) - 1
		if len(raw) < fixed {
			return nil, ppError(m.Definition, "too few arguments for macro %s", m.Name)
		}
		if len(raw) > len(m.Params) {
			var va []PPToken
			for i := fixed; i < len(raw); i++ {
				if i > fixed {
					va = append(va, PPToken{Kind: PPPunctuator, Lexeme: ",", Location: m.Definition, LeadingSpace: true})
				}
				va = append(va, raw[i]...)
			}
			raw = append(raw[:fixed], va)
		}
	} else if len(raw) != len(m.Params) {
		return nil, ppError(m.Definition, "wrong number of arguments for macro %s", m.Name)
	}
	args := make([]macroArg, len(raw))
	for i := range raw {
		expanded, err := pp.expand(raw[i])
		if err != nil {
			return nil, err
		}
		args[i] = macroArg{raw: trimArgumentPadding(raw[i]), expanded: expanded}
	}
	return args, nil
}

func splitMacroArgs(tokens []PPToken) [][]PPToken {
	var args [][]PPToken
	var current []PPToken
	depth := 0
	for _, tok := range tokens {
		if tok.Kind == PPPunctuator {
			switch tok.Lexeme {
			case "(":
				depth++
			case ")":
				if depth > 0 {
					depth--
				}
			case ",":
				if depth == 0 {
					args = append(args, trimArgumentPadding(current))
					current = nil
					continue
				}
			}
		}
		current = append(current, tok)
	}
	args = append(args, trimArgumentPadding(current))
	return args
}

func trimArgumentPadding(tokens []PPToken) []PPToken {
	start, end := 0, len(tokens)
	for start < end && (tokens[start].Kind == PPNewline || tokens[start].Kind == PPPadding) {
		start++
	}
	for end > start && (tokens[end-1].Kind == PPNewline || tokens[end-1].Kind == PPPadding) {
		end--
	}
	return cloneTokens(tokens[start:end])
}

func (pp *preprocessor) substitute(m *Macro, args []macroArg, use entity.SourcePos) ([]PPToken, error) {
	paramIndex := map[string]int{}
	for i, name := range m.Params {
		paramIndex[name] = i
	}
	rawParam := rawParamSet(m.Replacement, paramIndex)
	var out []PPToken
	for i := 0; i < len(m.Replacement); i++ {
		tok := m.Replacement[i]
		if tok.Kind == PPPunctuator && tok.Lexeme == "#" && i+1 < len(m.Replacement) {
			next := m.Replacement[i+1]
			if idx, ok := paramIndex[next.Lexeme]; ok {
				out = append(out, PPToken{Kind: PPString, Lexeme: stringifyArg(args[idx].raw), Location: use})
				i++
				continue
			}
		}
		if tok.Kind == PPPunctuator && tok.Lexeme == "##" {
			if len(out) == 0 || i+1 >= len(m.Replacement) {
				continue
			}
			left := out[len(out)-1]
			out = out[:len(out)-1]
			rightTokens := pp.substitutionFor(m.Replacement[i+1], args, paramIndex, rawParam, use)
			if len(rightTokens) == 0 {
				out = append(out, left)
				i++
				continue
			}
			pasted, err := pp.pasteTokens(left, rightTokens[0], use)
			if err != nil {
				return nil, err
			}
			out = append(out, pasted)
			out = append(out, rightTokens[1:]...)
			i++
			continue
		}
		out = append(out, pp.substitutionFor(tok, args, paramIndex, rawParam, use)...)
	}
	return out, nil
}

func rawParamSet(replacement []PPToken, params map[string]int) map[string]bool {
	raw := map[string]bool{}
	for i, tok := range replacement {
		if tok.Kind != PPIdentifier {
			continue
		}
		if _, ok := params[tok.Lexeme]; !ok {
			continue
		}
		if i > 0 && replacement[i-1].Kind == PPPunctuator && (replacement[i-1].Lexeme == "#" || replacement[i-1].Lexeme == "##") {
			raw[tok.Lexeme] = true
		}
		if i+1 < len(replacement) && replacement[i+1].Kind == PPPunctuator && replacement[i+1].Lexeme == "##" {
			raw[tok.Lexeme] = true
		}
	}
	return raw
}

func (pp *preprocessor) substitutionFor(tok PPToken, args []macroArg, params map[string]int, rawParam map[string]bool, use entity.SourcePos) []PPToken {
	if tok.Kind != PPIdentifier {
		tok.Location = use
		return []PPToken{tok}
	}
	idx, ok := params[tok.Lexeme]
	if !ok || idx >= len(args) {
		tok.Location = use
		return []PPToken{tok}
	}
	if rawParam[tok.Lexeme] {
		return withLocation(args[idx].raw, use)
	}
	return withLocation(args[idx].expanded, use)
}

func (pp *preprocessor) pasteTokens(left, right PPToken, use entity.SourcePos) (PPToken, error) {
	spelling := left.Lexeme + right.Lexeme
	sm := NewSourceManager()
	fileID := sm.AddFile("<paste>", spelling)
	tokens, err := scanFile(sm, fileID, spelling, pp.opts)
	if err != nil {
		return PPToken{}, err
	}
	tokens = dropNewlines(tokens)
	if len(tokens) != 1 {
		return PPToken{}, ppError(use, "token paste produced invalid token %q", spelling)
	}
	tok := tokens[0]
	tok.Location = use
	return tok, nil
}

func stringifyArg(tokens []PPToken) string {
	var b strings.Builder
	b.WriteByte('"')
	for i, tok := range tokens {
		if i > 0 && (tok.LeadingSpace || needsStringifySpace(tokens[i-1], tok)) {
			b.WriteByte(' ')
		}
		b.WriteString(escapeStringify(tok.Lexeme))
	}
	b.WriteByte('"')
	return b.String()
}

func needsStringifySpace(prev, next PPToken) bool {
	return prev.Kind == PPIdentifier && next.Kind == PPIdentifier
}

func escapeStringify(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r == '\\' || r == '"' {
			b.WriteByte('\\')
		}
		b.WriteRune(r)
	}
	return b.String()
}

func (p *Preprocessor) consumePragma() (bool, error) {
	consumed, ok, err := p.readFunctionInvocation()
	if err != nil || !ok {
		if !ok {
			p.unreadTokens(consumed)
		}
		return false, err
	}
	real := make([]PPToken, 0, len(consumed))
	for _, tok := range consumed {
		if tok.Kind != PPNewline && tok.Kind != PPPadding {
			real = append(real, tok)
		}
	}
	if len(real) == 3 && real[0].Lexeme == "(" && real[1].Kind == PPString && real[2].Lexeme == ")" {
		return true, nil
	}
	p.unreadTokens(consumed)
	return false, nil
}

func cloneTokens(tokens []PPToken) []PPToken {
	if len(tokens) == 0 {
		return nil
	}
	out := make([]PPToken, len(tokens))
	copy(out, tokens)
	return out
}

func withLocation(tokens []PPToken, loc entity.SourcePos) []PPToken {
	out := cloneTokens(tokens)
	for i := range out {
		out[i].Location = loc
	}
	return out
}
