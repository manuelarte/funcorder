package utils

import (
	"go/ast"
	"strings"
)

func FuncNameCanBeConstructor(n *ast.FuncDecl) bool {
	expectedConstructorPrefix := "new"
	isExported := n.Name.IsExported()
	startsWithNew := strings.HasPrefix(strings.ToLower(n.Name.Name), expectedConstructorPrefix) &&
		len(n.Name.Name) > len(expectedConstructorPrefix)
	hasAtLeastOneOutput := n.Type.Results != nil && len(n.Type.Results.List) > 0
	return isExported && startsWithNew && hasAtLeastOneOutput
}
