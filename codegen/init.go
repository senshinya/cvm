package codegen

import "shinya.click/cvm/bytecode"

func (g *generator) internString(value string) int {
	if id, ok := g.stringMap[value]; ok {
		return id
	}
	id := len(g.mod.Strings)
	bytes := append([]byte(value), 0)
	g.mod.Strings = append(g.mod.Strings, bytecode.StringConst{ID: id, Value: value, Bytes: bytes})
	g.stringMap[value] = id
	return id
}
