package errors

import (
	"fmt"
	"go/ast"
	"go/token"
)

type LinterError interface {
	error
	GetPos() token.Pos
	isLinterError()
}

var _ LinterError = new(ConstructorNotAfterStructTypeError)

type ConstructorNotAfterStructTypeError struct {
	Struct      *ast.TypeSpec
	Constructor *ast.FuncDecl
}

func (c ConstructorNotAfterStructTypeError) GetPos() token.Pos {
	return c.Constructor.Pos()
}

func (c ConstructorNotAfterStructTypeError) Error() string {
	return fmt.Sprintf("function %q for struct %q should be placed after the struct declaration",
		c.Constructor.Name, c.Struct.Name)
}

func (c ConstructorNotAfterStructTypeError) isLinterError() {}
