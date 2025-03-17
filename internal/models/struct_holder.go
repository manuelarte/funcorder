package models

import (
	"go/ast"

	"github.com/manuelarte/gofuncor/internal/errors"
)

// StructHolder contains all the information around a Go struct.
type StructHolder struct {
	// The struct declaration
	Struct *ast.TypeSpec
	// A Struct constructor is considered if starts with `New...` and the 1st output parameter is a struct
	Constructors []*ast.FuncDecl
	// TODO struct methods, they should be after the constructors
	StructMethods []*ast.FuncDecl
}

func (sh *StructHolder) AddConstructor(fn *ast.FuncDecl) {
	sh.Constructors = append(sh.Constructors, fn)
}

func (sh *StructHolder) Analyze() []errors.LinterError {
	// TODO maybe sort constructors and then report also, like NewXXX before MustXXX
	var errs []errors.LinterError
	structPos := sh.Struct.Pos()
	for _, c := range sh.Constructors {
		if c.Pos() < structPos {
			errs = append(errs, errors.ConstructorNotAfterStructTypeError{
				Struct:      sh.Struct,
				Constructor: c,
			})
		}
	}
	return errs
}
