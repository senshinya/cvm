package parser

import (
	"github.com/thoas/go-funk"
	"shinya.click/cvm/common"
	"shinya.click/cvm/parser/entity"
	"shinya.click/cvm/parser/glr"
)

func parseDeclaration(root *glr.RawAstNode) (*entity.Declaration, error) {
	if err := root.AssertNonTerminal(glr.Declaration); err != nil {
		panic(err)
	}

	res := &entity.Declaration{SourceRange: root.GetSourceRange()}

	// parse specifiers
	var err error
	res.Specifiers, res.MidType, err = parseDeclarationSpecifiers(root.Children[0])
	if err != nil {
		return nil, err
	}

	switch {
	case root.ReducedBy(glr.Declaration, 1):
		// declaration := declaration_specifiers SEMICOLON
		return res, nil
	case root.ReducedBy(glr.Declaration, 2):
		// declaration := declaration_specifiers init_declarator_list SEMICOLON
		res.InitDeclarators, err = parseInitDeclarators(root.Children[1], res.MidType)
		if err != nil {
			return nil, err
		}
		return res, nil
	default:
		panic("unreachable")
	}
}

func parseDeclarationSpecifiers(specifiersNode *glr.RawAstNode) (specifiers entity.Specifiers, midType entity.Type, err error) {
	if err = specifiersNode.AssertNonTerminal(glr.DeclarationSpecifiers); err != nil {
		panic(err)
	}

	specifierNodes := flattenDeclarationSpecifier(specifiersNode)

	// parse storage class specifier
	storagesSpecifiers := funk.Filter(specifierNodes, func(specifier *glr.RawAstNode) bool {
		return specifier.Typ == glr.StorageClassSpecifier
	}).([]*glr.RawAstNode)
	for _, storagesSpecifier := range storagesSpecifiers {
		parseStorageClassSpecifier(storagesSpecifier, &specifiers)
	}

	// parse type specifier and qualifiers
	midType, err = parseTypeSpecifiersAndQualifiers(
		funk.Filter(specifierNodes, func(specifier *glr.RawAstNode) bool {
			return specifier.Typ == glr.TypeSpecifier
		}).([]*glr.RawAstNode),
		funk.Filter(specifierNodes, func(specifier *glr.RawAstNode) bool {
			return specifier.Typ == glr.TypeQualifier
		}).([]*glr.RawAstNode),
		specifiersNode,
	)
	if err != nil {
		return
	}
	specifiers.SourceRange = specifiersNode.GetSourceRange()
	midType.SourceRange = specifiersNode.GetSourceRange()

	// parse function specifier
	functionSpecifiers := funk.Filter(specifierNodes, func(specifier *glr.RawAstNode) bool {
		return specifier.Typ == glr.FunctionSpecifier
	}).([]*glr.RawAstNode)
	if len(functionSpecifiers) != 0 {
		specifiers.Inline = true
	}

	return
}

func parseStorageClassSpecifier(storageSpecifier *glr.RawAstNode, spe *entity.Specifiers) {
	n := storageSpecifier.Children[0]
	switch n.Typ {
	case common.TYPEDEF:
		spe.TypeDef = true
	case common.EXTERN:
		spe.Extern = true
	case common.STATIC:
		spe.Static = true
	case common.AUTO:
		spe.Auto = true
	case common.REGISTER:
		spe.Register = true
	}
}

func parseInitDeclarators(declarators *glr.RawAstNode, midType entity.Type) ([]entity.InitDeclarator, error) {
	if err := declarators.AssertNonTerminal(glr.InitDeclaratorList); err != nil {
		panic(err)
	}

	var res []entity.InitDeclarator
	initDeclarators := flattenInitDeclarators(declarators)
	for _, initDeclaratorNode := range initDeclarators {
		initDeclarator := entity.InitDeclarator{SourceRange: initDeclaratorNode.GetSourceRange()}

		var err error
		// parse declarator
		initDeclarator.Declarator, err = parseDeclarator(initDeclaratorNode.Children[0], midType)
		if err != nil {
			return nil, err
		}

		switch {
		case initDeclaratorNode.ReducedBy(glr.InitDeclarator, 1):
			// init_declarator := declarator
		case initDeclaratorNode.ReducedBy(glr.InitDeclarator, 2):
			// init_declarator := declarator EQUAL initializer
			initDeclarator.Initializer, err = ParseInitializer(initDeclaratorNode.Children[2])
			if err != nil {
				return nil, err
			}
		default:
			panic("unreachable")
		}

		res = append(res, initDeclarator)
	}
	return res, nil
}

func parseDeclarator(root *glr.RawAstNode, midType entity.Type) (res *entity.Declarator, err error) {
	if err = root.AssertNonTerminal(glr.Declarator); err != nil {
		panic(err)
	}

	res = &entity.Declarator{SourceRange: root.GetSourceRange()}

	// 1. find the most inner direct_declarator node that contains only IDENTIFIER
	currentNode := root
	for {
		gotcha := false
		switch {
		case currentNode.ReducedBy(glr.Declarator, 1):
			// declarator := direct_declarator
			currentNode = currentNode.Children[0]
		case currentNode.ReducedBy(glr.Declarator, 2):
			// declarator := pointer direct_declarator
			currentNode = currentNode.Children[1]
		case currentNode.ReducedBy(glr.DirectDeclarator, 1):
			// direct_declarator := IDENTIFIER
			gotcha = true
		case currentNode.ReducedBy(glr.DirectDeclarator, 2):
			// direct_declarator := LEFT_PARENTHESES declarator RIGHT_PARENTHESES
			currentNode = currentNode.Children[1]
		case currentNode.ReducedBy(glr.DirectDeclarator, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14):
			// direct_declarator := direct_declarator ...
			currentNode = currentNode.Children[0]
		default:
			panic("unreachable")
		}
		if gotcha {
			break
		}
	}
	res.Identifier = currentNode.Children[0].Terminal

	currentType := &res.Type
	for {
		// need to parse the most out node
		if currentNode == root.Parent {
			break
		}
		switch {
		case currentNode.ReducedBy(glr.Declarator, 1):
			// declarator := direct_declarator
			currentNode = currentNode.Parent
		case currentNode.ReducedBy(glr.Declarator, 2):
			// declarator := pointer direct_declarator
			currentType = parsePointer(currentNode.Children[0], currentType).PointerInnerType
			currentNode = currentNode.Parent
		case currentNode.ReducedBy(glr.DirectDeclarator, 1):
			// direct_declarator := IDENTIFIER, do nothing
			currentNode = currentNode.Parent
		case currentNode.ReducedBy(glr.DirectDeclarator, 2):
			// direct_declarator := LEFT_PARENTHESES declarator RIGHT_PARENTHESES
			currentNode = currentNode.Parent
		case currentNode.ReducedBy(glr.DirectDeclarator, 3, 4, 5, 6, 7, 8, 9, 10, 11):
			currentType.MetaType = entity.MetaTypeArray
			currentType.ArrayMetaInfo, err = parseArrayMetaInfo(currentNode)
			if err != nil {
				return res, err
			}
			currentType = currentType.ArrayMetaInfo.InnerType
			currentNode = currentNode.Parent
		case currentNode.ReducedBy(glr.DirectDeclarator, 12, 13, 14):
			currentType.MetaType = entity.MetaTypeFunction
			currentType.FunctionMetaInfo, err = parseFunctionMetaInfo(currentNode)
			if err != nil {
				return res, err
			}
			currentType = currentType.FunctionMetaInfo.ReturnType
			currentNode = currentNode.Parent
		default:
			panic("unreachable")
		}
	}
	*currentType = midType
	res.Type.SourceRange = root.GetSourceRange()
	return res, nil
}

func parsePointer(rootPointer *glr.RawAstNode, currentType *entity.Type) *entity.Type {
	if err := rootPointer.AssertNonTerminal(glr.Pointer); err != nil {
		panic(err)
	}
	// find the most inner pointer
	currentPointer := rootPointer
	for {
		gotcha := false
		switch {
		case currentPointer.ReducedBy(glr.Pointer, 1, 2):
			// pointer := ASTERISK
			// pointer := ASTERISK type_qualifier_list
			gotcha = true
		case currentPointer.ReducedBy(glr.Pointer, 3):
			// pointer := ASTERISK pointer
			currentPointer = currentPointer.Children[1]
		case currentPointer.ReducedBy(glr.Pointer, 4):
			// pointer := ASTERISK type_qualifier_list pointer
			currentPointer = currentPointer.Children[2]
		default:
			panic("unreachable")
		}
		if gotcha {
			break
		}
	}

	for {
		if currentPointer == rootPointer {
			break
		}
		currentType.MetaType = entity.MetaTypePointer
		currentType.PointerInnerType = &entity.Type{}
		switch {
		case currentPointer.ReducedBy(glr.Pointer, 1, 3):
			// pointer := ASTERISK
			// pointer := ASTERISK pointer
			currentType = currentType.PointerInnerType
			currentPointer = currentPointer.Parent
		case currentPointer.ReducedBy(glr.Pointer, 2, 4):
			// pointer := ASTERISK type_qualifier_list
			// pointer := ASTERISK type_qualifier_list pointer
			typeQualifiers := flattenTypeQualifierList(currentPointer.Children[1])
			parseTypeQualifiers(typeQualifiers, &currentType.TypeQualifiers)
			currentType = currentType.PointerInnerType
			currentPointer = currentPointer.Parent
		default:
			panic("unreachable")
		}
	}
	currentType.MetaType = entity.MetaTypePointer
	currentType.PointerInnerType = &entity.Type{}
	if len(rootPointer.Children) == 2 {
		// parse the root qualifiers
		typeQualifiers := flattenTypeQualifierList(rootPointer.Children[1])
		parseTypeQualifiers(typeQualifiers, &currentType.TypeQualifiers)
	}
	return currentType
}

func ParseDeclarationList(root *glr.RawAstNode) ([]*entity.Declaration, error) {
	if err := root.AssertNonTerminal(glr.DeclarationList); err != nil {
		panic(err)
	}

	declarationNodes := flattenDeclarationList(root)

	var res []*entity.Declaration
	for _, node := range declarationNodes {
		declaration, err := parseDeclaration(node)
		if err != nil {
			return nil, err
		}
		res = append(res, declaration)
	}

	return res, nil
}
