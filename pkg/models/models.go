package models

import "go/ast"

// StructHolder contains all the information around a Go struct.
type StructHolder struct {
	// A Struct constructor is considered if starts with `New...` and the 1st output parameter is a struct
	Constructors []*ast.FuncDecl
}
