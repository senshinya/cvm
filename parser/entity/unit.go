package entity

type TranslationUnitType uint8

const (
	UnitTypeDeclaration TranslationUnitType = 1
	UnitTypeFunctionDef TranslationUnitType = 2
)

type TranslationUnit interface {
	GetUnitType() TranslationUnitType
}
