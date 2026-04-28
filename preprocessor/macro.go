package preprocessor

import (
	"slices"
	"strings"

	"shinya.click/cvm/entity"
)

type MacroKind int

const (
	MacroObject MacroKind = iota
	MacroFunction
)

type Macro struct {
	Name        string
	Kind        MacroKind
	Params      []string
	Variadic    bool
	Replacement []PPToken
	Definition  entity.SourcePos
	Disabled    bool
}

type MacroTable struct {
	entries map[string]*Macro
}

func NewMacroTable(target TargetInfo) *MacroTable {
	m := &MacroTable{entries: map[string]*Macro{}}
	m.DefineObject("__STDC__", []PPToken{{Kind: PPNumber, Lexeme: "1"}})
	m.DefineObject("__STDC_VERSION__", []PPToken{{Kind: PPNumber, Lexeme: "199901L"}})
	if target.Hosted {
		m.DefineObject("__STDC_HOSTED__", []PPToken{{Kind: PPNumber, Lexeme: "1"}})
	} else {
		m.DefineObject("__STDC_HOSTED__", []PPToken{{Kind: PPNumber, Lexeme: "0"}})
	}
	m.DefineObject("__SIZE_TYPE__", typeSpellingTokens(target.SizeType))
	m.DefineObject("__PTRDIFF_TYPE__", typeSpellingTokens(target.PtrdiffType))
	m.DefineObject("__WCHAR_TYPE__", typeSpellingTokens(target.WCharType))
	m.DefineObject("__CHAR_BIT__", []PPToken{{Kind: PPNumber, Lexeme: "8"}})
	return m
}

func typeSpellingTokens(spelling string) []PPToken {
	parts := strings.Fields(spelling)
	out := make([]PPToken, 0, len(parts))
	for i, part := range parts {
		out = append(out, PPToken{Kind: PPIdentifier, Lexeme: part, LeadingSpace: i > 0})
	}
	return out
}

func (m *MacroTable) DefineObject(name string, replacement []PPToken) {
	m.entries[name] = &Macro{Name: name, Kind: MacroObject, Replacement: replacement}
}

func (m *MacroTable) DefineFunction(name string, params []string, variadic bool, replacement []PPToken, pos entity.SourcePos) {
	m.entries[name] = &Macro{Name: name, Kind: MacroFunction, Params: params, Variadic: variadic, Replacement: replacement, Definition: pos}
}

func (m *MacroTable) Undef(name string) {
	delete(m.entries, name)
}

func (m *MacroTable) Lookup(name string) (*Macro, bool) {
	got, ok := m.entries[name]
	return got, ok
}

func (m *MacroTable) Define(macro *Macro) error {
	if old, ok := m.entries[macro.Name]; ok && !sameMacro(old, macro) {
		return ppError(macro.Definition, "conflicting macro redefinition for %s", macro.Name)
	}
	m.entries[macro.Name] = macro
	return nil
}

func sameMacro(a, b *Macro) bool {
	if a.Kind != b.Kind || a.Name != b.Name || a.Variadic != b.Variadic || !slices.Equal(a.Params, b.Params) {
		return false
	}
	if len(a.Replacement) != len(b.Replacement) {
		return false
	}
	for i := range a.Replacement {
		if a.Replacement[i].Kind != b.Replacement[i].Kind || a.Replacement[i].Lexeme != b.Replacement[i].Lexeme {
			return false
		}
	}
	return true
}
