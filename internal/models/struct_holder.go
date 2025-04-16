package models

import (
	"cmp"
	"go/ast"
	"slices"

	"golang.org/x/tools/go/analysis"

	"github.com/manuelarte/funcorder/internal/diag"
	"github.com/manuelarte/funcorder/internal/features"
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

// Analyze applies the linter to the struct holder.
func (sh *StructHolder) Analyze() []analysis.Diagnostic {
	// TODO maybe sort constructors and then report also, like NewXXX before MustXXX

	slices.SortFunc(sh.StructMethods, func(a, b *ast.FuncDecl) int {
		return cmp.Compare(a.Pos(), b.Pos())
	})

	var reports []analysis.Diagnostic

	if sh.Features.IsEnabled(features.ConstructorCheck) {
		reports = append(reports, sh.analyzeConstructor()...)
	}

	if sh.Features.IsEnabled(features.StructMethodCheck) {
		reports = append(reports, sh.analyzeStructMethod()...)
	}

	// TODO also check that the methods are declared after the struct
	return reports
}

func (sh *StructHolder) analyzeConstructor() []analysis.Diagnostic {
	var reports []analysis.Diagnostic
	structPos := sh.Struct.Pos()

	for i, c := range sh.Constructors {
		if c.Pos() < structPos {
			reports = append(reports, diag.NewConstructorNotAfterStructType(sh.Struct, c))
		}

		if len(sh.StructMethods) > 0 && c.Pos() > sh.StructMethods[0].Pos() {
			reports = append(reports, diag.NewConstructorNotBeforeStructMethod(sh.Struct, c, sh.StructMethods[0]))
		}

		if sh.Features.IsEnabled(features.AlphabeticalCheck) && i < len(sh.Constructors)-1 {
			if sh.Constructors[i].Name.Name > sh.Constructors[i+1].Name.Name {
				reports = append(reports,
					diag.NewAdjacentConstructorsNotSortedAlphabetically(sh.Struct,
						sh.Constructors[i], sh.Constructors[i+1]))
			}
		}
	}
	return reports
}

func (sh *StructHolder) analyzeStructMethod() []analysis.Diagnostic {
	var reports []analysis.Diagnostic
	var lastExportedMethod *ast.FuncDecl

	for _, m := range sh.StructMethods {
		if !m.Name.IsExported() {
			continue
		}

		if lastExportedMethod == nil {
			lastExportedMethod = m
		}

		if lastExportedMethod.Pos() < m.Pos() {
			lastExportedMethod = m
		}
	}

	if lastExportedMethod != nil {
		for _, m := range sh.StructMethods {
			if m.Name.IsExported() || m.Pos() >= lastExportedMethod.Pos() {
				continue
			}

			reports = append(reports, diag.NewNonExportedMethodBeforeExportedForStruct(sh.Struct, m, lastExportedMethod))
		}
	}

	if sh.Features.IsEnabled(features.AlphabeticalCheck) {
		ef := filterMethods(sh.StructMethods, true)
		reports = append(reports, isSorted(sh.Struct, ef)...)
		nef := filterMethods(sh.StructMethods, false)
		reports = append(reports, isSorted(sh.Struct, nef)...)
	}

	return reports
}

func filterMethods(fs []*ast.FuncDecl, isExported bool) []*ast.FuncDecl {
	var ff []*ast.FuncDecl
	for _, f := range fs {
		if f.Name.IsExported() == isExported {
			ff = append(ff, f)
		}
	}
	return ff
}

func isSorted(s *ast.TypeSpec, ff []*ast.FuncDecl) []analysis.Diagnostic {
	var reports []analysis.Diagnostic
	for i := range ff {
		if i < len(ff)-1 {
			if ff[i].Name.Name > ff[i+1].Name.Name {
				reports = append(reports, diag.NewAdjacentStructMethodsNotSortedAlphabetically(s, ff[i], ff[i+1]))
			}
		}
	}
	return reports
}
