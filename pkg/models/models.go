package models

import (
	"github.com/manuelarte/gofuncor/internal/utils"
	"go/ast"
)

// StructHolder contains all the information around a Go struct.
type StructHolder struct {
	Struct *ast.TypeSpec
	// A Struct constructor is considered if starts with `New...` and the 1st output parameter is a struct
	Constructors []*ast.FuncDecl
}

func (sh *StructHolder) AddConstructor(fn *ast.FuncDecl) {
	sh.Constructors = append(sh.Constructors, fn)
}

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
