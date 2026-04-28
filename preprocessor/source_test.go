package preprocessor

import "testing"

func TestSourceManagerPhysicalAndPresumedLocations(t *testing.T) {
	sm := NewSourceManager()
	fileID := sm.AddFile("main.c", "int a;\n#line 42 \"virt.c\"\nint b;\n")
	locA := sm.Location(fileID, 0)
	gotA := sm.DisplayLocation(locA)
	if gotA.File != "main.c" || gotA.Line != 1 || gotA.Column != 1 {
		t.Fatalf("physical location = %#v, want main.c:1:1", gotA)
	}

	lineDirectiveOffset := len("int a;\n")
	sm.SetPresumedLine(fileID, lineDirectiveOffset+len("#line 42 \"virt.c\"\n"), "virt.c", 42)
	locB := sm.Location(fileID, len("int a;\n#line 42 \"virt.c\"\n"))
	gotB := sm.DisplayLocation(locB)
	if gotB.File != "virt.c" || gotB.Line != 42 || gotB.Column != 1 {
		t.Fatalf("presumed location = %#v, want virt.c:42:1", gotB)
	}
}

func TestSourceManagerExpansionLocations(t *testing.T) {
	sm := NewSourceManager()
	fileID := sm.AddFile("main.c", "#define A 1\nA\n")
	defLoc := sm.Location(fileID, len("#define "))
	useLoc := sm.Location(fileID, len("#define A 1\n"))
	spellingLoc := sm.Location(fileID, len("#define A "))
	expLoc := sm.AddExpansion(spellingLoc, useLoc, defLoc)

	got := sm.ExpansionTrace(expLoc)
	if got.Spelling.LocationID != spellingLoc.LocationID {
		t.Fatalf("spelling loc = %#v, want %#v", got.Spelling, spellingLoc)
	}
	if got.Expansion.LocationID != useLoc.LocationID {
		t.Fatalf("expansion loc = %#v, want %#v", got.Expansion, useLoc)
	}
	if got.Definition.LocationID != defLoc.LocationID {
		t.Fatalf("definition loc = %#v, want %#v", got.Definition, defLoc)
	}
}
