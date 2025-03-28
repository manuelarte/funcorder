package models

import (
	"go/ast"
	"sort"

	"github.com/manuelarte/funcorder/internal/features"

	"github.com/manuelarte/funcorder/internal/errors"
)

// StructHolder contains all the information around a Go struct.
type StructHolder struct {
	// The features to be analyzed
	Features features.Feature
	// The struct declaration
	Struct *ast.TypeSpec
	// A Struct constructor is considered if starts with `New...` and the 1st output parameter is a struct
	Constructors []*ast.FuncDecl
	// Struct methods
	StructMethods []*ast.FuncDecl
}

func (sh *StructHolder) AddConstructor(fn *ast.FuncDecl) {
	sh.Constructors = append(sh.Constructors, fn)
}

func (sh *StructHolder) AddMethod(fn *ast.FuncDecl) {
	sh.StructMethods = append(sh.StructMethods, fn)
}

//nolint:gocognit,nestif // refactor later
func (sh *StructHolder) Analyze() []errors.LinterError {
	// TODO maybe sort constructors and then report also, like NewXXX before MustXXX
	var errs []errors.LinterError
	sort.Slice(sh.StructMethods, func(i, j int) bool {
		return sh.StructMethods[i].Pos() < sh.StructMethods[j].Pos()
	})
	structPos := sh.Struct.Pos()
	if sh.Features.IsEnabled(features.ConstructorCheck) {
		for _, c := range sh.Constructors {
			if c.Pos() < structPos {
				errs = append(errs, errors.ConstructorNotAfterStructTypeError{
					Struct:      sh.Struct,
					Constructor: c,
				})
			}
			if len(sh.StructMethods) > 0 && c.Pos() > sh.StructMethods[0].Pos() {
				errs = append(errs, errors.ConstructorNotBeforeStructMethodsError{
					Struct:      sh.Struct,
					Constructor: c,
					Method:      sh.StructMethods[0],
				})
			}
		}
	}

	if sh.Features.IsEnabled(features.StructMethodCheck) {
		var lastExportedMethod *ast.FuncDecl
		for _, m := range sh.StructMethods {
			if m.Name.IsExported() {
				if lastExportedMethod == nil {
					lastExportedMethod = m
				}
				if lastExportedMethod.Pos() < m.Pos() {
					lastExportedMethod = m
				}
			}
		}

		if lastExportedMethod != nil {
			for _, m := range sh.StructMethods {
				if !m.Name.IsExported() && m.Pos() < lastExportedMethod.Pos() {
					errs = append(errs, errors.PrivateMethodBeforePublicForStructTypeError{
						Struct:        sh.Struct,
						PrivateMethod: m,
						PublicMethod:  lastExportedMethod,
					})
				}
			}
		}
	}

	// TODO also check that the methods are declared after the struct
	return errs
}
