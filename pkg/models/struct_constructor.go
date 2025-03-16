package models

import (
	"go/ast"

	"github.com/manuelarte/gofuncor/internal/utils"
)

type StructConstructor struct {
	Func *ast.FuncDecl
}

func NewStructConstructor(funcDec *ast.FuncDecl) (StructConstructor, bool) {
	if utils.FuncNameCanBeConstructor(funcDec) {
		return StructConstructor{
			Func: funcDec,
		}, true
	}
	return StructConstructor{}, false
}

func (sc StructConstructor) GetStructReturn() (*ast.Ident, bool) {
	expr := sc.Func.Type.Results.List[0].Type
	return sc.returnType(expr)
}

func (sc StructConstructor) returnType(expr ast.Expr) (*ast.Ident, bool) {
	if pointerExpr, isPointerExpr := expr.(*ast.StarExpr); isPointerExpr {
		return sc.returnType(pointerExpr.X)
	}
	if structExpr, isStructExpr := expr.(*ast.Ident); isStructExpr {
		return structExpr, true
	}
	return nil, false
}
