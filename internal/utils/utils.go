package utils

import (
	"go/ast"
	"strings"
)

func FuncNameCanBeConstructor(n *ast.FuncDecl) bool {
	isExported := n.Name.IsExported()
	startsWithNew := strings.HasPrefix(strings.ToLower(n.Name.Name), "new") && len(n.Name.Name) > 3
	hasAtLeastOneOutput := n.Type.Results != nil && len(n.Type.Results.List) > 0
	return isExported && startsWithNew && hasAtLeastOneOutput
}
