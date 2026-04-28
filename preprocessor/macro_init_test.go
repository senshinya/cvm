package preprocessor

import "testing"

func TestOptionsDefinesAndUndefines(t *testing.T) {
	pp := newPreprocessor("main.c", "", Options{
		MacroActions: []MacroAction{
			{Kind: MacroDefine, Name: "A", Value: "1"},
			{Kind: MacroDefine, Name: "B"},
			{Kind: MacroUndef, Name: "B"},
			{Kind: MacroUndef, Name: "C"},
			{Kind: MacroDefine, Name: "C", Value: "2"},
		},
	})
	if _, ok := pp.macros.Lookup("A"); !ok {
		t.Fatalf("A macro was not defined from options")
	}
	if _, ok := pp.macros.Lookup("B"); ok {
		t.Fatalf("B macro should be undefined by options")
	}
	if _, ok := pp.macros.Lookup("C"); !ok {
		t.Fatalf("C macro should be defined because actions preserve -U/-D order")
	}
}

func TestPredefinedTargetMacros(t *testing.T) {
	pp := newPreprocessor("main.c", "", Options{})
	for _, name := range []string{"__STDC__", "__STDC_VERSION__", "__STDC_HOSTED__", "__SIZE_TYPE__", "__PTRDIFF_TYPE__", "__WCHAR_TYPE__", "__CHAR_BIT__"} {
		if _, ok := pp.macros.Lookup(name); !ok {
			t.Fatalf("predefined macro %s missing", name)
		}
	}
}
