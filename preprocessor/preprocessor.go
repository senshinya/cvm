package preprocessor

import (
	"os"

	"shinya.click/cvm/entity"
	"shinya.click/cvm/lexer"
)

type Result struct {
	Tokens  []entity.Token
	Sources *SourceManager
}

func PreprocessSource(name, source string, opts Options) (*Result, error) {
	opts = normalizeOptions(opts)
	_ = opts
	sm := NewSourceManager()
	sm.AddFile(name, source)
	tokens, err := lexer.NewLexer(source).ScanTokens()
	if err != nil {
		return nil, err
	}
	return &Result{Tokens: tokens, Sources: sm}, nil
}

func PreprocessFile(path string, opts Options) (*Result, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return PreprocessSource(path, string(content), opts)
}
