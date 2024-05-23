package syntax

type Declaration struct {
	Specifiers  Specifiers
	MidType     Type
	Declarators []Declarator
}

func (d *Declaration) GetUnitType() TranslationUnitType {
	return UnitTypeDeclaration
}

type TypeQualifiers struct {
	Const    bool
	Restrict bool
	Volatile bool
}

type Declarator struct {
	Identifier string

	Type Type
}
