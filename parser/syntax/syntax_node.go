package syntax

type Program struct {
	Units []*TranslationUnit
}

type TranslationUnitType uint8

const (
	UnitTypeDeclaration TranslationUnitType = 1
	UnitTypeTypeDef     TranslationUnitType = 2
	UnitTypeFunctionDef TranslationUnitType = 3
)

type TranslationUnit interface {
	GetUnitType() TranslationUnitType
}

type Declaration struct {
	Specifiers  Specifiers
	Type        TypeDeclaration // may not be the final type, need to parse to the Declarator parser to determine
	Declarators []Declarator
}

func (d *Declaration) GetUnitType() TranslationUnitType {
	return UnitTypeDeclaration
}

type Specifiers struct {
	// storage class specifier
	Extern   bool
	Static   bool
	Auto     bool
	Register bool

	// function specifier
	Inline bool
}

type Declarator struct {
	Identifier string

	Type Type
}

type TypeDef struct {
}

func (d *TypeDef) GetUnitType() TranslationUnitType {
	return UnitTypeTypeDef
}

type FunctionDef struct {
}

func (d *FunctionDef) GetUnitType() TranslationUnitType {
	return UnitTypeFunctionDef
}
