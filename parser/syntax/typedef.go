package syntax

type TypeDef struct {
	MidType     Type
	Declarators []Declarator
}

func (d *TypeDef) GetUnitType() TranslationUnitType {
	return UnitTypeTypeDef
}
