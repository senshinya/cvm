package preprocessor

type Preprocessor struct {
	pp     *preprocessor
	stack  []TokenSource
	unread []PPToken
}

type FileTokenSource struct {
	tokens []PPToken
	index  int
}

type MacroTokenSource struct {
	tokens []PPToken
	index  int
	macro  *Macro
}

func NewTokenPreprocessor(pp *preprocessor, tokens []PPToken) *Preprocessor {
	p := &Preprocessor{pp: pp}
	p.push(&FileTokenSource{tokens: append(tokens, PPToken{Kind: PPEOF})})
	return p
}

func (p *Preprocessor) Lex() (PPToken, error) {
	for {
		tok, err := p.readRaw()
		if err != nil {
			return tok, err
		}
		if tok.Kind != PPIdentifier {
			return tok, nil
		}
		expanded, ok, err := p.expandIdentifier(tok)
		if err != nil || !ok {
			return expanded, err
		}
	}
}

func (p *Preprocessor) push(src TokenSource) {
	p.stack = append(p.stack, src)
}

func (p *Preprocessor) pop() {
	if len(p.stack) == 0 {
		return
	}
	p.stack = p.stack[:len(p.stack)-1]
}

func (p *Preprocessor) unreadToken(tok PPToken) {
	if tok.Kind != PPEOF {
		p.unread = append(p.unread, tok)
	}
}

func (p *Preprocessor) unreadTokens(tokens []PPToken) {
	for i := len(tokens) - 1; i >= 0; i-- {
		p.unreadToken(tokens[i])
	}
}

func (p *Preprocessor) readRaw() (PPToken, error) {
	if n := len(p.unread); n > 0 {
		tok := p.unread[n-1]
		p.unread = p.unread[:n-1]
		return tok, nil
	}
	for len(p.stack) > 0 {
		top := p.stack[len(p.stack)-1]
		tok, err := top.Lex()
		if err != nil {
			return tok, err
		}
		if tok.Kind != PPEOF {
			return tok, nil
		}
		if src, ok := top.(*MacroTokenSource); ok {
			// 宏替换列表读完后才重新启用该宏；这样最后一个替换 token 在被检查时仍处于禁用状态。
			src.macro.Disabled = false
		}
		p.pop()
	}
	return PPToken{Kind: PPEOF}, nil
}

func (s *FileTokenSource) Lex() (PPToken, error) {
	if s.index >= len(s.tokens) {
		return PPToken{Kind: PPEOF}, nil
	}
	tok := s.tokens[s.index]
	s.index++
	return tok, nil
}

func (s *MacroTokenSource) Lex() (PPToken, error) {
	if s.index >= len(s.tokens) {
		return PPToken{Kind: PPEOF}, nil
	}
	tok := s.tokens[s.index]
	s.index++
	return tok, nil
}
