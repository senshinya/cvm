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
	m.DefineObject("__WINT_TYPE__", typeSpellingTokens("unsigned int"))
	m.DefineObject("__SIG_ATOMIC_TYPE__", typeSpellingTokens("int"))
	m.DefineObject("__INT8_TYPE__", typeSpellingTokens("signed char"))
	m.DefineObject("__INT16_TYPE__", typeSpellingTokens("short"))
	m.DefineObject("__INT32_TYPE__", typeSpellingTokens("int"))
	m.DefineObject("__INT64_TYPE__", typeSpellingTokens("long"))
	m.DefineObject("__UINT8_TYPE__", typeSpellingTokens("unsigned char"))
	m.DefineObject("__UINT16_TYPE__", typeSpellingTokens("unsigned short"))
	m.DefineObject("__UINT32_TYPE__", typeSpellingTokens("unsigned int"))
	m.DefineObject("__UINT64_TYPE__", typeSpellingTokens("unsigned long"))
	m.DefineObject("__INT_LEAST8_TYPE__", typeSpellingTokens("signed char"))
	m.DefineObject("__INT_LEAST16_TYPE__", typeSpellingTokens("short"))
	m.DefineObject("__INT_LEAST32_TYPE__", typeSpellingTokens("int"))
	m.DefineObject("__INT_LEAST64_TYPE__", typeSpellingTokens("long"))
	m.DefineObject("__UINT_LEAST8_TYPE__", typeSpellingTokens("unsigned char"))
	m.DefineObject("__UINT_LEAST16_TYPE__", typeSpellingTokens("unsigned short"))
	m.DefineObject("__UINT_LEAST32_TYPE__", typeSpellingTokens("unsigned int"))
	m.DefineObject("__UINT_LEAST64_TYPE__", typeSpellingTokens("unsigned long"))
	m.DefineObject("__INT_FAST8_TYPE__", typeSpellingTokens("signed char"))
	m.DefineObject("__INT_FAST16_TYPE__", typeSpellingTokens("long"))
	m.DefineObject("__INT_FAST32_TYPE__", typeSpellingTokens("long"))
	m.DefineObject("__INT_FAST64_TYPE__", typeSpellingTokens("long"))
	m.DefineObject("__UINT_FAST8_TYPE__", typeSpellingTokens("unsigned char"))
	m.DefineObject("__UINT_FAST16_TYPE__", typeSpellingTokens("unsigned long"))
	m.DefineObject("__UINT_FAST32_TYPE__", typeSpellingTokens("unsigned long"))
	m.DefineObject("__UINT_FAST64_TYPE__", typeSpellingTokens("unsigned long"))
	m.DefineObject("__INTPTR_TYPE__", typeSpellingTokens(target.PtrdiffType))
	m.DefineObject("__UINTPTR_TYPE__", typeSpellingTokens(target.SizeType))
	m.DefineObject("__INTMAX_TYPE__", typeSpellingTokens(target.IntmaxType))
	m.DefineObject("__UINTMAX_TYPE__", typeSpellingTokens(target.UIntmaxType))
	m.DefineObject("__builtin_va_list", []PPToken{
		{Kind: PPIdentifier, Lexeme: "void"},
		{Kind: PPPunctuator, Lexeme: "*", LeadingSpace: true},
	})
	m.DefineObject("__CHAR_BIT__", []PPToken{{Kind: PPNumber, Lexeme: "8"}})
	m.DefineObject("__SCHAR_MAX__", []PPToken{{Kind: PPNumber, Lexeme: "127"}})
	m.DefineObject("__SHRT_MAX__", []PPToken{{Kind: PPNumber, Lexeme: "32767"}})
	m.DefineObject("__INT_MAX__", []PPToken{{Kind: PPNumber, Lexeme: "2147483647"}})
	m.DefineObject("__LONG_MAX__", []PPToken{{Kind: PPNumber, Lexeme: "9223372036854775807L"}})
	m.DefineObject("__LONG_LONG_MAX__", []PPToken{{Kind: PPNumber, Lexeme: "9223372036854775807LL"}})
	m.DefineObject("__INT8_MAX__", []PPToken{{Kind: PPNumber, Lexeme: "127"}})
	m.DefineObject("__INT16_MAX__", []PPToken{{Kind: PPNumber, Lexeme: "32767"}})
	m.DefineObject("__INT32_MAX__", []PPToken{{Kind: PPNumber, Lexeme: "2147483647"}})
	m.DefineObject("__INT64_MAX__", []PPToken{{Kind: PPNumber, Lexeme: "9223372036854775807L"}})
	m.DefineObject("__UINT8_MAX__", []PPToken{{Kind: PPNumber, Lexeme: "255"}})
	m.DefineObject("__UINT16_MAX__", []PPToken{{Kind: PPNumber, Lexeme: "65535"}})
	m.DefineObject("__UINT32_MAX__", []PPToken{{Kind: PPNumber, Lexeme: "4294967295U"}})
	m.DefineObject("__UINT64_MAX__", []PPToken{{Kind: PPNumber, Lexeme: "18446744073709551615UL"}})
	m.DefineObject("__INT_LEAST8_MAX__", []PPToken{{Kind: PPNumber, Lexeme: "127"}})
	m.DefineObject("__INT_LEAST16_MAX__", []PPToken{{Kind: PPNumber, Lexeme: "32767"}})
	m.DefineObject("__INT_LEAST32_MAX__", []PPToken{{Kind: PPNumber, Lexeme: "2147483647"}})
	m.DefineObject("__INT_LEAST64_MAX__", []PPToken{{Kind: PPNumber, Lexeme: "9223372036854775807L"}})
	m.DefineObject("__UINT_LEAST8_MAX__", []PPToken{{Kind: PPNumber, Lexeme: "255"}})
	m.DefineObject("__UINT_LEAST16_MAX__", []PPToken{{Kind: PPNumber, Lexeme: "65535"}})
	m.DefineObject("__UINT_LEAST32_MAX__", []PPToken{{Kind: PPNumber, Lexeme: "4294967295U"}})
	m.DefineObject("__UINT_LEAST64_MAX__", []PPToken{{Kind: PPNumber, Lexeme: "18446744073709551615UL"}})
	m.DefineObject("__INT_FAST8_MAX__", []PPToken{{Kind: PPNumber, Lexeme: "127"}})
	m.DefineObject("__INT_FAST16_MAX__", []PPToken{{Kind: PPNumber, Lexeme: "9223372036854775807L"}})
	m.DefineObject("__INT_FAST32_MAX__", []PPToken{{Kind: PPNumber, Lexeme: "9223372036854775807L"}})
	m.DefineObject("__INT_FAST64_MAX__", []PPToken{{Kind: PPNumber, Lexeme: "9223372036854775807L"}})
	m.DefineObject("__UINT_FAST8_MAX__", []PPToken{{Kind: PPNumber, Lexeme: "255"}})
	m.DefineObject("__UINT_FAST16_MAX__", []PPToken{{Kind: PPNumber, Lexeme: "18446744073709551615UL"}})
	m.DefineObject("__UINT_FAST32_MAX__", []PPToken{{Kind: PPNumber, Lexeme: "18446744073709551615UL"}})
	m.DefineObject("__UINT_FAST64_MAX__", []PPToken{{Kind: PPNumber, Lexeme: "18446744073709551615UL"}})
	m.DefineObject("__INTPTR_MAX__", []PPToken{{Kind: PPNumber, Lexeme: "9223372036854775807L"}})
	m.DefineObject("__UINTPTR_MAX__", []PPToken{{Kind: PPNumber, Lexeme: "18446744073709551615UL"}})
	m.DefineObject("__INTMAX_MAX__", []PPToken{{Kind: PPNumber, Lexeme: "9223372036854775807L"}})
	m.DefineObject("__UINTMAX_MAX__", []PPToken{{Kind: PPNumber, Lexeme: "18446744073709551615UL"}})
	m.DefineObject("__PTRDIFF_MAX__", []PPToken{{Kind: PPNumber, Lexeme: "9223372036854775807L"}})
	m.DefineObject("__SIZE_MAX__", []PPToken{{Kind: PPNumber, Lexeme: "18446744073709551615UL"}})
	m.DefineObject("__SIG_ATOMIC_MIN__", signedMinTokens("2147483647", ""))
	m.DefineObject("__SIG_ATOMIC_MAX__", []PPToken{{Kind: PPNumber, Lexeme: "2147483647"}})
	m.DefineObject("__WCHAR_MIN__", signedMinTokens("2147483647", ""))
	m.DefineObject("__WCHAR_MAX__", []PPToken{{Kind: PPNumber, Lexeme: "2147483647"}})
	m.DefineObject("__WINT_MIN__", []PPToken{{Kind: PPNumber, Lexeme: "0U"}})
	m.DefineObject("__WINT_MAX__", []PPToken{{Kind: PPNumber, Lexeme: "4294967295U"}})
	m.DefineObject("__DBL_MIN__", []PPToken{{Kind: PPNumber, Lexeme: "2.22507385850720138309e-308"}})
	m.DefineObject("__DBL_MAX__", []PPToken{{Kind: PPNumber, Lexeme: "1.79769313486231570815e+308"}})
	m.DefineObject("__FLT_MIN__", []PPToken{{Kind: PPNumber, Lexeme: "1.17549435082228750797e-38F"}})
	m.DefineObject("__FLT_MAX__", []PPToken{{Kind: PPNumber, Lexeme: "3.40282346638528859812e+38F"}})
	m.DefineObject("__LDBL_MIN__", []PPToken{{Kind: PPNumber, Lexeme: "3.36210314311209350626e-4932L"}})
	m.DefineObject("__LDBL_MAX__", []PPToken{{Kind: PPNumber, Lexeme: "1.18973149535723176502e+4932L"}})
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

func signedMinTokens(max, suffix string) []PPToken {
	return []PPToken{
		{Kind: PPPunctuator, Lexeme: "("},
		{Kind: PPPunctuator, Lexeme: "-"},
		{Kind: PPNumber, Lexeme: max + suffix},
		{Kind: PPPunctuator, Lexeme: "-"},
		{Kind: PPNumber, Lexeme: "1" + suffix},
		{Kind: PPPunctuator, Lexeme: ")"},
	}
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
