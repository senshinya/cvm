package entity

type TranslationUnitType uint8

const (
	UnitTypeDeclaration TranslationUnitType = 1
	UnitTypeFunctionDef TranslationUnitType = 2
)

type TranslationUnit interface {
	GetUnitType() TranslationUnitType
}

type Specifiers struct {
	// storage class specifier
	Extern   bool
	Static   bool
	Auto     bool
	Register bool

	// typedef
	TypeDef bool

	// function specifier
	Inline bool
}
