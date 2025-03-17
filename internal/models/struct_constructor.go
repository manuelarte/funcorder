package models

import (
	"go/ast"

	"github.com/manuelarte/gofuncor/internal/utils"
)

type StructConstructor struct {
	constructor *ast.FuncDecl
}

func NewStructConstructor(funcDec *ast.FuncDecl) (StructConstructor, bool) {
	if utils.FuncCanBeConstructor(funcDec) {
		return StructConstructor{
			constructor: funcDec,
		}, true
	}
	return StructConstructor{}, false
}

// GetStructReturn Return the struct linked to this "constructor".
func (sc StructConstructor) GetStructReturn() (*ast.Ident, bool) {
	expr := sc.constructor.Type.Results.List[0].Type
	return sc.returnType(expr)
}

func (sc StructConstructor) GetConstructor() *ast.FuncDecl {
	return sc.constructor
}

func (sc StructConstructor) returnType(expr ast.Expr) (*ast.Ident, bool) {
	return utils.GetIdent(expr)
}
