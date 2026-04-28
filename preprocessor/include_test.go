package preprocessor

import (
	"fmt"
	"testing"
)

type mapFS map[string]string

func (m mapFS) ReadFile(path string) ([]byte, error) {
	if s, ok := m[path]; ok {
		return []byte(s), nil
	}
	return nil, fmt.Errorf("missing %s", path)
}

func TestQuotedIncludeAndMacroExpandedInclude(t *testing.T) {
	fs := mapFS{
		"/work/main.c": "#define H \"defs.h\"\n#include H\nT x;\n",
		"/work/defs.h": "typedef int T;\n",
	}
	res, err := PreprocessFile("/work/main.c", Options{FileSystem: fs})
	if err != nil {
		t.Fatalf("PreprocessFile failed: %v", err)
	}
	if !hasIdentifier(res.Tokens, "T") {
		t.Fatalf("included typedef token missing: %#v", res.Tokens)
	}
}

func TestIncludedSourceIsExpandedOnceWithOuterStream(t *testing.T) {
	fs := mapFS{
		"/work/main.c": "#define X 1\n#include \"defs.h\"\n",
		"/work/defs.h": "int y = X;\n",
	}
	res, err := PreprocessFile("/work/main.c", Options{FileSystem: fs})
	if err != nil {
		t.Fatalf("PreprocessFile failed: %v", err)
	}
	count := 0
	for _, tok := range res.Tokens {
		if tok.Lexeme == "1" {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("included tokens should be expanded exactly once, saw %d ones: %#v", count, res.Tokens)
	}
}

func TestIncludeStackTraceForHeaderToken(t *testing.T) {
	fs := mapFS{
		"/work/main.c": "#include \"defs.h\"\n",
		"/work/defs.h": "int from_header;\n",
	}
	res, err := PreprocessFile("/work/main.c", Options{FileSystem: fs})
	if err != nil {
		t.Fatalf("PreprocessFile failed: %v", err)
	}
	for _, tok := range res.Tokens {
		if tok.Lexeme == "from_header" {
			trace := res.Sources.IncludeTrace(tok.SourceStart)
			if len(trace) != 1 || trace[0].File != "/work/main.c" {
				t.Fatalf("include trace = %#v, want one entry for /work/main.c", trace)
			}
			return
		}
	}
	t.Fatalf("from_header token missing: %#v", res.Tokens)
}
