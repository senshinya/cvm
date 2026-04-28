package preprocessor

import "shinya.click/cvm/entity"

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
	pp := newPreprocessor(name, source, opts)
	pp.sm = sm
	processed, err := pp.process(ppTokens)
	if err != nil {
		return nil, err
	}
	expanded, err := pp.expand(processed)
	if err != nil {
		return nil, err
	}
	tokens, err := convertToParserTokens(expanded, sm)
	if err != nil {
		return nil, err
	}
	return &Result{Tokens: tokens, Sources: sm}, nil
}

func PreprocessFile(path string, opts Options) (*Result, error) {
	opts = normalizeOptions(opts)
	fs := opts.FileSystem
	if fs == nil {
		fs = osFileSystem{}
	}
	content, err := fs.ReadFile(path)
	if err != nil {
		return nil, err
	}
	opts.FileSystem = fs
	return PreprocessSource(path, string(content), opts)
}
