package preprocessor

import (
	"strconv"
	"strings"
)

type preprocessor struct {
	opts         Options
	sm           *SourceManager
	macros       *MacroTable
	output       []PPToken
	conds        []conditionalGroup
	includeDepth int
	includeTrace []IncludeTraceEntry
}

type conditionalGroup struct {
	parentActive bool
	branchTaken  bool
	active       bool
	seenElse     bool
}

func newPreprocessor(name, source string, opts Options) *preprocessor {
	opts = normalizeOptions(opts)
	pp := &preprocessor{
		opts:   opts,
		sm:     NewSourceManager(),
		macros: NewMacroTable(opts.Target),
	}
	pp.applyMacroActions(opts.MacroActions)
	return pp
}

func (pp *preprocessor) applyMacroActions(actions []MacroAction) {
	for _, action := range actions {
		switch action.Kind {
		case MacroDefine:
			value := action.Value
			if value == "" {
				value = "1"
			}
			pp.macros.DefineObject(action.Name, pp.scanReplacement(value))
		case MacroUndef:
			pp.macros.Undef(action.Name)
		}
	}
}

func (pp *preprocessor) scanReplacement(value string) []PPToken {
	sm := NewSourceManager()
	fileID := sm.AddFile("<command-line>", value)
	tokens, err := scanFile(sm, fileID, value, pp.opts)
	if err != nil {
		return []PPToken{{Kind: PPIdentifier, Lexeme: value}}
	}
	return dropNewlines(tokens)
}

func (pp *preprocessor) process(tokens []PPToken) ([]PPToken, error) {
	pp.output = nil
	return pp.processLines(tokens)
}

func (pp *preprocessor) processLines(tokens []PPToken) ([]PPToken, error) {
	condDepth := len(pp.conds)
	for _, line := range splitLogicalLines(tokens) {
		if len(line) == 0 {
			continue
		}
		if isDirectiveLine(line) {
			if err := pp.handleDirective(line); err != nil {
				return nil, err
			}
			continue
		}
		if pp.isActive() {
			pp.output = append(pp.output, line...)
		}
	}
	if len(pp.conds) != condDepth {
		return nil, ppError(tokens[len(tokens)-1].Location, "unterminated conditional inclusion")
	}
	return pp.output, nil
}

func splitLogicalLines(tokens []PPToken) [][]PPToken {
	var lines [][]PPToken
	var current []PPToken
	for _, tok := range tokens {
		current = append(current, tok)
		if tok.Kind == PPNewline {
			lines = append(lines, current)
			current = nil
		}
	}
	if len(current) > 0 {
		lines = append(lines, current)
	}
	return lines
}

func isDirectiveLine(line []PPToken) bool {
	for _, tok := range line {
		if tok.Kind == PPPadding || tok.Kind == PPNewline {
			continue
		}
		return tok.StartOfLine && tok.Kind == PPPunctuator && isHash(tok)
	}
	return false
}

func (pp *preprocessor) handleDirective(line []PPToken) error {
	body := directiveBody(line)
	if len(body) == 0 {
		return nil
	}
	name := body[0]
	if name.Kind == PPNumber {
		if pp.isActive() {
			return pp.handleLine(body, line)
		}
		return nil
	}
	if name.Kind != PPIdentifier {
		return ppError(name.Location, "invalid preprocessing directive")
	}
	switch name.Lexeme {
	case "if":
		parent := pp.isActive()
		value, err := pp.evalIfExpression(body[1:])
		if err != nil {
			return err
		}
		active := parent && value != 0
		pp.conds = append(pp.conds, conditionalGroup{parentActive: parent, branchTaken: active, active: active})
	case "ifdef":
		if len(body) < 2 {
			return ppError(name.Location, "missing macro name in #ifdef")
		}
		_, ok := pp.macros.Lookup(body[1].Lexeme)
		parent := pp.isActive()
		active := parent && ok
		pp.conds = append(pp.conds, conditionalGroup{parentActive: parent, branchTaken: active, active: active})
	case "ifndef":
		if len(body) < 2 {
			return ppError(name.Location, "missing macro name in #ifndef")
		}
		_, ok := pp.macros.Lookup(body[1].Lexeme)
		parent := pp.isActive()
		active := parent && !ok
		pp.conds = append(pp.conds, conditionalGroup{parentActive: parent, branchTaken: active, active: active})
	case "elif":
		if len(pp.conds) == 0 {
			return ppError(name.Location, "unexpected #elif")
		}
		top := &pp.conds[len(pp.conds)-1]
		if top.seenElse {
			return ppError(name.Location, "#elif after #else")
		}
		value, err := pp.evalIfExpression(body[1:])
		if err != nil {
			return err
		}
		top.active = top.parentActive && !top.branchTaken && value != 0
		top.branchTaken = top.branchTaken || top.active
	case "else":
		if len(pp.conds) == 0 {
			return ppError(name.Location, "unexpected #else")
		}
		top := &pp.conds[len(pp.conds)-1]
		if top.seenElse {
			return ppError(name.Location, "duplicate #else")
		}
		top.active = top.parentActive && !top.branchTaken
		top.branchTaken = true
		top.seenElse = true
	case "endif":
		if len(pp.conds) == 0 {
			return ppError(name.Location, "unexpected #endif")
		}
		pp.conds = pp.conds[:len(pp.conds)-1]
	case "define":
		if !pp.isActive() {
			return nil
		}
		return pp.handleDefine(body[1:])
	case "include":
		if pp.isActive() {
			return pp.handleInclude(body[1:], name)
		}
	case "undef":
		if pp.isActive() && len(body) >= 2 {
			pp.macros.Undef(body[1].Lexeme)
		}
	case "line":
		if pp.isActive() {
			return pp.handleLine(body[1:], line)
		}
	case "pragma":
		return nil
	case "error":
		if pp.isActive() {
			return ppError(name.Location, "#error %s", joinLexemes(body[1:]))
		}
	default:
		if pp.isActive() {
			return ppError(name.Location, "unknown preprocessing directive #%s", name.Lexeme)
		}
	}
	return nil
}

func directiveBody(line []PPToken) []PPToken {
	var out []PPToken
	sawSharp := false
	for _, tok := range line {
		if tok.Kind == PPNewline || tok.Kind == PPPadding {
			continue
		}
		if !sawSharp {
			if tok.Kind == PPPunctuator && isHash(tok) {
				sawSharp = true
			}
			continue
		}
		out = append(out, tok)
	}
	return out
}

func (pp *preprocessor) handleDefine(tokens []PPToken) error {
	if len(tokens) == 0 || tokens[0].Kind != PPIdentifier {
		if len(tokens) == 0 {
			return nil
		}
		return ppError(tokens[0].Location, "macro name must be an identifier")
	}
	name := tokens[0].Lexeme
	if len(tokens) >= 2 && tokens[1].Kind == PPPunctuator && tokens[1].Lexeme == "(" && !tokens[1].LeadingSpace {
		params, variadic, end, err := parseMacroParams(tokens)
		if err != nil {
			return err
		}
		replacement := dropNewlines(tokens[end+1:])
		return pp.macros.Define(&Macro{
			Name:        name,
			Kind:        MacroFunction,
			Params:      params,
			Variadic:    variadic,
			Replacement: replacement,
			Definition:  tokens[0].Location,
		})
	}
	replacement := dropNewlines(tokens[1:])
	return pp.macros.Define(&Macro{Name: name, Kind: MacroObject, Replacement: replacement, Definition: tokens[0].Location})
}

func parseMacroParams(tokens []PPToken) ([]string, bool, int, error) {
	var params []string
	variadic := false
	i := 2
	if i < len(tokens) && tokens[i].Kind == PPPunctuator && tokens[i].Lexeme == ")" {
		return params, false, i, nil
	}
	for i < len(tokens) {
		tok := tokens[i]
		switch {
		case tok.Kind == PPIdentifier:
			params = append(params, tok.Lexeme)
			i++
		case tok.Kind == PPPunctuator && tok.Lexeme == "...":
			params = append(params, "__VA_ARGS__")
			variadic = true
			i++
		default:
			return nil, false, i, ppError(tok.Location, "invalid macro parameter list")
		}
		if i >= len(tokens) {
			return nil, false, i, ppError(tok.Location, "unterminated macro parameter list")
		}
		if tokens[i].Kind == PPPunctuator && tokens[i].Lexeme == ")" {
			return params, variadic, i, nil
		}
		if tokens[i].Kind != PPPunctuator || tokens[i].Lexeme != "," {
			return nil, false, i, ppError(tokens[i].Location, "expected comma in macro parameter list")
		}
		i++
		if variadic {
			return nil, false, i, ppError(tokens[i-1].Location, "variadic parameter must be last")
		}
	}
	return nil, false, len(tokens), ppError(tokens[len(tokens)-1].Location, "unterminated macro parameter list")
}

func (pp *preprocessor) handleLine(tokens []PPToken, line []PPToken) error {
	if len(tokens) == 0 || tokens[0].Kind != PPNumber {
		if len(tokens) == 0 {
			return nil
		}
		return ppError(tokens[0].Location, "invalid #line")
	}
	lineNo, err := strconv.Atoi(strings.TrimRight(tokens[0].Lexeme, "uUlL"))
	if err != nil {
		return ppError(tokens[0].Location, "invalid #line number")
	}
	file := ""
	if len(tokens) >= 2 && tokens[1].Kind == PPString {
		file = strings.Trim(tokens[1].Lexeme, "\"")
	}
	fileID, offset := pp.lineEndOffset(line)
	pp.sm.SetPresumedLine(fileID, offset, file, lineNo)
	return nil
}

func (pp *preprocessor) lineEndOffset(line []PPToken) (int, int) {
	for i := len(line) - 1; i >= 0; i-- {
		tok := line[i]
		if tok.Location.LocationID <= 0 {
			continue
		}
		loc := pp.sm.locations[tok.Location.LocationID]
		if tok.Kind == PPNewline {
			return loc.fileID, loc.offset + 1
		}
		offset := loc.offset + len(tok.Lexeme)
		if loc.fileID > 0 && loc.fileID < len(pp.sm.files) {
			content := pp.sm.files[loc.fileID].content
			for offset < len(content) && content[offset] != '\n' {
				offset++
			}
			if offset < len(content) {
				offset++
			}
		}
		return loc.fileID, offset
	}
	return 0, 0
}

func (pp *preprocessor) isActive() bool {
	for _, cond := range pp.conds {
		if !cond.active {
			return false
		}
	}
	return true
}

func dropNewlines(tokens []PPToken) []PPToken {
	out := make([]PPToken, 0, len(tokens))
	for _, tok := range tokens {
		if tok.Kind != PPNewline {
			out = append(out, tok)
		}
	}
	return out
}

func joinLexemes(tokens []PPToken) string {
	parts := make([]string, 0, len(tokens))
	for _, tok := range tokens {
		parts = append(parts, tok.Lexeme)
	}
	return strings.Join(parts, " ")
}
