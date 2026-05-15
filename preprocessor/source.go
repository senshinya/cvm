package preprocessor

import (
	"sort"
	"strings"

	"shinya.click/cvm/entity"
)

type DisplayLocation struct {
	File   string
	Line   int
	Column int
	Text   string
}

type IncludeTraceEntry struct {
	File   string
	Line   int
	Column int
}

type ExpansionTrace struct {
	Spelling   entity.SourcePos
	Expansion  entity.SourcePos
	Definition entity.SourcePos
}

type SourceManager struct {
	files      []sourceFile
	locations  []sourceLocation
	expansions []ExpansionTrace
	includes   map[int][]IncludeTraceEntry
}

type sourceFile struct {
	name      string
	content   string
	lineStart []int
	presumed  []presumedLine
}

type sourceLocation struct {
	fileID int
	offset int
}

type presumedLine struct {
	offset int
	file   string
	line   int
}

func NewSourceManager() *SourceManager {
	return &SourceManager{
		files:      []sourceFile{{name: "<invalid>", lineStart: []int{0}}},
		locations:  []sourceLocation{{}},
		expansions: []ExpansionTrace{{}},
		includes:   map[int][]IncludeTraceEntry{},
	}
}

func (sm *SourceManager) AddFile(name, content string) int {
	starts := []int{0}
	for i, b := range content {
		if b == '\n' {
			starts = append(starts, i+1)
		}
	}
	sm.files = append(sm.files, sourceFile{name: name, content: content, lineStart: starts})
	return len(sm.files) - 1
}

func (sm *SourceManager) Location(fileID, offset int) entity.SourcePos {
	sm.locations = append(sm.locations, sourceLocation{fileID: fileID, offset: offset})
	line, column := 0, 0
	if fileID > 0 && fileID < len(sm.files) {
		starts := sm.files[fileID].lineStart
		idx := sort.Search(len(starts), func(i int) bool { return starts[i] > offset }) - 1
		if idx >= 0 {
			line = idx + 1
			column = offset - starts[idx] + 1
		}
	}
	return entity.SourcePos{LocationID: len(sm.locations) - 1, Line: line, Column: column}
}

func (sm *SourceManager) SetPresumedLine(fileID, offset int, file string, line int) {
	f := &sm.files[fileID]
	f.presumed = append(f.presumed, presumedLine{offset: offset, file: file, line: line})
	sort.Slice(f.presumed, func(i, j int) bool { return f.presumed[i].offset < f.presumed[j].offset })
}

func (sm *SourceManager) AddExpansion(spelling, expansion, definition entity.SourcePos) entity.SourcePos {
	sm.expansions = append(sm.expansions, ExpansionTrace{Spelling: spelling, Expansion: expansion, Definition: definition})
	idx := len(sm.expansions) - 1
	return entity.SourcePos{LocationID: -idx}
}

func (sm *SourceManager) SetIncludeTrace(fileID int, trace []IncludeTraceEntry) {
	sm.includes[fileID] = append([]IncludeTraceEntry(nil), trace...)
}

func (sm *SourceManager) IncludeTrace(pos entity.SourcePos) []IncludeTraceEntry {
	if pos.LocationID <= 0 || pos.LocationID >= len(sm.locations) {
		return nil
	}
	loc := sm.locations[pos.LocationID]
	return append([]IncludeTraceEntry(nil), sm.includes[loc.fileID]...)
}

func (sm *SourceManager) ExpansionTrace(pos entity.SourcePos) ExpansionTrace {
	if pos.LocationID >= 0 {
		return ExpansionTrace{}
	}
	idx := -pos.LocationID
	if idx <= 0 || idx >= len(sm.expansions) {
		return ExpansionTrace{}
	}
	return sm.expansions[idx]
}

func (sm *SourceManager) DisplayLocation(pos entity.SourcePos) DisplayLocation {
	if pos.LocationID == 0 {
		return DisplayLocation{Line: pos.Line, Column: pos.Column}
	}
	if pos.LocationID < 0 {
		trace := sm.ExpansionTrace(pos)
		if trace.Expansion.LocationID != 0 {
			return sm.DisplayLocation(trace.Expansion)
		}
		return DisplayLocation{}
	}
	if pos.LocationID >= len(sm.locations) {
		return DisplayLocation{}
	}
	loc := sm.locations[pos.LocationID]
	if loc.fileID <= 0 || loc.fileID >= len(sm.files) {
		return DisplayLocation{}
	}
	f := sm.files[loc.fileID]
	line := sort.Search(len(f.lineStart), func(i int) bool { return f.lineStart[i] > loc.offset })
	if line == 0 {
		line = 1
	}
	lineStart := f.lineStart[line-1]
	display := DisplayLocation{File: f.name, Line: line, Column: loc.offset - lineStart + 1}
	if len(f.presumed) > 0 {
		idx := sort.Search(len(f.presumed), func(i int) bool { return f.presumed[i].offset > loc.offset }) - 1
		if idx >= 0 {
			p := f.presumed[idx]
			baseLine := sort.Search(len(f.lineStart), func(i int) bool { return f.lineStart[i] > p.offset })
			if baseLine == 0 {
				baseLine = 1
			}
			display.File = p.file
			display.Line = p.line + (line - baseLine)
		}
	}
	lineEnd := strings.IndexByte(f.content[lineStart:], '\n')
	if lineEnd < 0 {
		display.Text = f.content[lineStart:]
	} else {
		display.Text = f.content[lineStart : lineStart+lineEnd]
	}
	return display
}
