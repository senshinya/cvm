package syntax

type TypeDef struct {
}

func (d *TypeDef) GetUnitType() TranslationUnitType {
	return UnitTypeTypeDef
}
