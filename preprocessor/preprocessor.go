package preprocessor

import (
	"os"

	"shinya.click/cvm/entity"
)

type Result struct {
	Tokens  []entity.Token
	Sources *SourceManager
}

func PreprocessSource(name, source string, opts Options) (*Result, error) {
	opts = normalizeOptions(opts)
	sm := NewSourceManager()
	fileID := sm.AddFile(name, source)
	ppTokens, err := scanFile(sm, fileID, source, opts)
	if err != nil {
		return nil, err
	}
	tokens, err := convertToParserTokens(ppTokens, sm)
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
