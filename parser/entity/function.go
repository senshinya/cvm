package entity

type FunctionDef struct {
}

func (d *FunctionDef) GetUnitType() TranslationUnitType {
	return UnitTypeFunctionDef
}
