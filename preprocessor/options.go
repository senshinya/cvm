package preprocessor

type Standard int

const (
	StandardC99 Standard = iota
)

type TargetInfo struct {
	SizeType    string
	PtrdiffType string
	IntmaxType  string
	UIntmaxType string
	WCharType   string
	CharSigned  bool
	Hosted      bool
}

type MacroActionKind int

const (
	MacroDefine MacroActionKind = iota
	MacroUndef
)

type MacroAction struct {
	Kind  MacroActionKind
	Name  string
	Value string
}

func DefaultTarget() TargetInfo {
	return TargetInfo{
		SizeType:    "unsigned long",
		PtrdiffType: "long",
		IntmaxType:  "long",
		UIntmaxType: "unsigned long",
		WCharType:   "int",
		CharSigned:  true,
		Hosted:      true,
	}
}

type FileSystem interface {
	ReadFile(path string) ([]byte, error)
}

type Options struct {
	IncludePaths []string
	MacroActions []MacroAction
	Std          Standard
	Target       TargetInfo
	FileSystem   FileSystem
}

func normalizeOptions(opts Options) Options {
	if opts.Target.SizeType == "" {
		opts.Target = DefaultTarget()
	}
	if opts.Std != StandardC99 {
		opts.Std = StandardC99
	}
	return opts
}
