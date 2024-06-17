package entity

type Declaration struct {
	Specifiers  Specifiers
	MidType     Type
	Declarators []Declarator
}

func (d *Declaration) GetUnitType() TranslationUnitType {
	return UnitTypeDeclaration
}

type Declarator struct {
	Identifier string

	Type Type
}
