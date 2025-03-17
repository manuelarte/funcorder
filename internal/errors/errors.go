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

type ConstructorNotBeforeStructMethodsError struct {
	Struct      *ast.TypeSpec
	Constructor *ast.FuncDecl
	Method      *ast.FuncDecl
}

func (c ConstructorNotBeforeStructMethodsError) GetPos() token.Pos {
	return c.Constructor.Pos()
}

func (c ConstructorNotBeforeStructMethodsError) Error() string {
	return fmt.Sprintf("constructor %q for struct %q should be placed before struct method %q",
		c.Constructor.Name, c.Struct.Name, c.Method.Name)
}

func (c ConstructorNotBeforeStructMethodsError) isLinterError() {}

var _ LinterError = new(PrivateMethodBeforePublicForStructTypeError)

type PrivateMethodBeforePublicForStructTypeError struct {
	Struct        *ast.TypeSpec
	PrivateMethod *ast.FuncDecl
	PublicMethod  *ast.FuncDecl
}

func (c PrivateMethodBeforePublicForStructTypeError) GetPos() token.Pos {
	return c.PrivateMethod.Pos()
}

func (c PrivateMethodBeforePublicForStructTypeError) Error() string {
	return fmt.Sprintf("unexported method %q for struct %q should be placed after the exported method %q",
		c.PrivateMethod.Name, c.Struct.Name, c.PublicMethod.Name)
}

func (c PrivateMethodBeforePublicForStructTypeError) isLinterError() {}
