package preprocessor

import (
	"path/filepath"
	"strconv"
	"strings"

	"shinya.click/cvm/entity"
)

const maxIncludeDepth = 64

type IncludeResolver struct {
	opts Options
}

func newIncludeResolver(opts Options) IncludeResolver {
	opts = normalizeOptions(opts)
	if opts.FileSystem == nil {
		opts.FileSystem = osFileSystem{}
	}
	return IncludeResolver{opts: opts}
}

func (r IncludeResolver) resolveQuoted(currentFile, name string) (string, string, error) {
	candidates := []string{name}
	if currentFile != "" && !filepath.IsAbs(name) {
		candidates = append([]string{filepath.Join(filepath.Dir(currentFile), name)}, candidates...)
	}
	for _, dir := range r.opts.IncludePaths {
		candidates = append(candidates, filepath.Join(dir, name))
	}
	for _, path := range dedupeStrings(candidates) {
		content, err := r.opts.FileSystem.ReadFile(path)
		if err == nil {
			return path, string(content), nil
		}
	}
	if content, ok := builtinHeader(name, r.opts.Target); ok {
		return "<" + name + ">", content, nil
	}
	if base := filepath.Base(name); base != name {
		if content, ok := builtinHeader(base, r.opts.Target); ok {
			return "<" + base + ">", content, nil
		}
	}
	return "", "", ppError(entity.SourcePos{}, "include file not found: %s", name)
}

func (r IncludeResolver) resolveAngled(name string) (string, string, error) {
	if content, ok := builtinHeader(name, r.opts.Target); ok {
		return "<" + name + ">", content, nil
	}
	for _, dir := range r.opts.IncludePaths {
		path := filepath.Join(dir, name)
		content, err := r.opts.FileSystem.ReadFile(path)
		if err == nil {
			return path, string(content), nil
		}
	}
	return "", "", ppError(entity.SourcePos{}, "include file not found: %s", name)
}

func dedupeStrings(values []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, value := range values {
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
	}
	return out
}

func (pp *preprocessor) handleInclude(tokens []PPToken, directive PPToken) error {
	if pp.includeDepth >= maxIncludeDepth {
		return ppError(directive.Location, "include depth exceeds %d", maxIncludeDepth)
	}
	operand := dropNewlines(tokens)
	header, angled, err := pp.parseIncludeOperand(operand, directive.Location)
	if err != nil {
		// include 操作数可以由宏展开得到，例如 #define H "x.h" 后的 #include H。
		expanded, expandErr := pp.expand(operand)
		if expandErr != nil {
			return expandErr
		}
		header, angled, err = pp.parseIncludeOperand(expanded, directive.Location)
		if err != nil {
			return err
		}
	}
	resolver := newIncludeResolver(pp.opts)
	current := pp.fileForLocation(directive.Location)
	var resolved, content string
	if angled {
		resolved, content, err = resolver.resolveAngled(header)
	} else {
		resolved, content, err = resolver.resolveQuoted(current, header)
	}
	if err != nil {
		return ppError(directive.Location, err.Error())
	}
	fileID := pp.sm.AddFile(resolved, content)
	trace := append([]IncludeTraceEntry(nil), pp.includeTrace...)
	trace = append(trace, includeEntry(pp.sm.DisplayLocation(directive.Location)))
	pp.sm.SetIncludeTrace(fileID, trace)
	pp.includeDepth++
	pp.includeTrace = trace
	tokens, err = scanFile(pp.sm, fileID, content, pp.opts)
	if err == nil {
		_, err = pp.processLines(tokens)
	}
	pp.includeDepth--
	if len(trace) > 0 {
		pp.includeTrace = trace[:len(trace)-1]
	} else {
		pp.includeTrace = nil
	}
	return err
}

func (pp *preprocessor) parseIncludeOperand(tokens []PPToken, fallback entity.SourcePos) (string, bool, error) {
	tokens = dropNewlines(tokens)
	if len(tokens) == 1 && tokens[0].Kind == PPString {
		name, err := strconv.Unquote(tokens[0].Lexeme)
		if err != nil {
			return "", false, ppError(tokens[0].Location, "invalid include string")
		}
		return name, false, nil
	}
	if len(tokens) >= 3 && tokens[0].Kind == PPPunctuator && tokens[0].Lexeme == "<" {
		var b strings.Builder
		for i := 1; i < len(tokens); i++ {
			if tokens[i].Kind == PPPunctuator && tokens[i].Lexeme == ">" {
				return b.String(), true, nil
			}
			b.WriteString(tokens[i].Lexeme)
		}
	}
	if len(tokens) > 0 {
		return "", false, ppError(tokens[0].Location, "invalid include operand")
	}
	return "", false, ppError(fallback, "missing include operand")
}

func (pp *preprocessor) fileForLocation(pos entity.SourcePos) string {
	if pos.LocationID <= 0 || pos.LocationID >= len(pp.sm.locations) {
		return ""
	}
	fileID := pp.sm.locations[pos.LocationID].fileID
	if fileID <= 0 || fileID >= len(pp.sm.files) {
		return ""
	}
	return pp.sm.files[fileID].name
}

func includeEntry(loc DisplayLocation) IncludeTraceEntry {
	return IncludeTraceEntry{File: loc.File, Line: loc.Line, Column: loc.Column}
}
