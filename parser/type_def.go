package parser

import (
	"errors"
	"shinya.click/cvm/parser/entity"
	"shinya.click/cvm/parser/glr"
)

func parseTypeDef(root *entity.AstNode, specifiers entity.Specifiers, midType entity.Type) (entity.TranslationUnit, error) {
	var err error
	if err = checkTypeDefSpecifiers(specifiers); err != nil {
		return nil, err
	}

	res := &entity.TypeDef{MidType: midType}
	if len(glr.Productions[root.ProdIndex].Right) == 2 {
		// reduced by declaration := declaration_specifiers SEMICOLON
		// this production can only declare struct, union or enum
		// otherwise "declaration does not declare anything" occurs
		// we treat it as error
		if midType.MetaType != entity.MetaTypeStruct &&
			midType.MetaType != entity.MetaTypeUnion &&
			midType.MetaType != entity.MetaTypeEnum {
			return nil, errors.New("declaration does not declare anything")
		}
		return res, nil
	}

	// parse init declarators
	res.Declarators, err = parseInitDeclarators(root.Children[1], res.MidType)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func checkTypeDefSpecifiers(specifiers entity.Specifiers) error {
	// When TypeDef, no other specifier is allowed
	if specifiers.Extern || specifiers.Static || specifiers.Auto ||
		specifiers.Register || specifiers.Inline {
		return errors.New("specifiers are not allowed in type def")
	}
	return nil
}
