package utils

import (
	"go/ast"
	"strings"
)

func FuncCanBeConstructor(n *ast.FuncDecl) bool {
	expectedConstructorPrefix := "new"
	if !n.Name.IsExported() {
		return false
	}
	if n.Recv != nil {
		return false
	}
	if !strings.HasPrefix(strings.ToLower(n.Name.Name), expectedConstructorPrefix) ||
		len(n.Name.Name) <= len(expectedConstructorPrefix) {
		return false
	}
	if n.Type.Results == nil || len(n.Type.Results.List) == 0 {
		return false
	}
	return true
}

func FuncIsMethod(n *ast.FuncDecl) (*ast.Ident, bool) {
	if n.Recv == nil {
		return nil, false
	}
	if len(n.Recv.List) != 1 {
		return nil, false
	}
	if recv, ok := GetIdent(n.Recv.List[0].Type); ok {
		return recv, true
	}

	return nil, false
}

func GetIdent(expr ast.Expr) (*ast.Ident, bool) {
	if pointerExpr, isPointerExpr := expr.(*ast.StarExpr); isPointerExpr {
		return GetIdent(pointerExpr.X)
	}
	if structExpr, isStructExpr := expr.(*ast.Ident); isStructExpr {
		return structExpr, true
	}
	return nil, false
}
