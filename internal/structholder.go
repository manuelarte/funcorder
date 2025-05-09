package internal

import (
	"cmp"
	"go/ast"
	"slices"
	"strings"

	"golang.org/x/tools/go/analysis"
)

type (
	ExportedMethods   []*ast.FuncDecl
	UnexportedMethods []*ast.FuncDecl
)

// StructHolder contains all the information around a Go struct.
type StructHolder struct {
	// The features to be analyzed
	Features Feature

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
func (sh *StructHolder) Analyze(pass *analysis.Pass) ([]analysis.Diagnostic, error) {
	// TODO maybe sort constructors and then report also, like NewXXX before MustXXX
	slices.SortFunc(sh.StructMethods, func(a, b *ast.FuncDecl) int {
		return cmp.Compare(a.Pos(), b.Pos())
	})

	var reports []analysis.Diagnostic

	if sh.Features.IsEnabled(ConstructorCheck) {
		newReports, err := sh.analyzeConstructor(pass)
		if err != nil {
			return nil, err
		}

		reports = append(reports, newReports...)
	}

	if sh.Features.IsEnabled(StructMethodCheck) {
		newReports, err := sh.analyzeStructMethod(pass)
		if err != nil {
			return nil, err
		}

		reports = append(reports, newReports...)
	}

	// TODO also check that the methods are declared after the struct
	return reports, nil
}

func (sh *StructHolder) analyzeConstructor(pass *analysis.Pass) ([]analysis.Diagnostic, error) {
	var reports []analysis.Diagnostic

	for i, constructor := range sh.Constructors {
		if constructor.Pos() < sh.Struct.Pos() {
			reports = append(reports, NewConstructorNotAfterStructType(sh.Struct, constructor))
		}

		if len(sh.StructMethods) > 0 && constructor.Pos() > sh.StructMethods[0].Pos() {
			reports = append(reports, NewConstructorNotBeforeStructMethod(sh.Struct, constructor, sh.StructMethods[0]))
		}

		if sh.Features.IsEnabled(AlphabeticalCheck) &&
			i < len(sh.Constructors)-1 && sh.Constructors[i].Name.Name > sh.Constructors[i+1].Name.Name {
			reports = append(reports,
				NewAdjacentConstructorsNotSortedAlphabetically(sh.Struct, sh.Constructors[i], sh.Constructors[i+1]),
			)
		}

		// propose fix
		if len(reports) > 0 {
			suggestedFixes, err := sh.suggestConstructorFix(pass)
			if err != nil {
				return nil, err
			}

			reports[0].SuggestedFixes = suggestedFixes
		}
	}

	return reports, nil
}

func (sh *StructHolder) analyzeStructMethod(pass *analysis.Pass) ([]analysis.Diagnostic, error) {
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

	var reports []analysis.Diagnostic

	if lastExportedMethod != nil {
		for _, m := range sh.StructMethods {
			if m.Name.IsExported() || m.Pos() >= lastExportedMethod.Pos() {
				continue
			}

			reports = append(reports, NewUnexportedMethodBeforeExportedForStruct(sh.Struct, m, lastExportedMethod))
		}
	}

	if sh.Features.IsEnabled(AlphabeticalCheck) {
		exported, unexported := SplitExportedUnexported(sh.StructMethods)
		reports = slices.Concat(reports,
			sortDiagnostics(sh.Struct, exported),
			sortDiagnostics(sh.Struct, unexported),
		)
	}

	if len(reports) > 0 {
		suggestedFixes, err := sh.suggestMethodFix(pass)
		if err != nil {
			return nil, err
		}

		reports[0].SuggestedFixes = suggestedFixes
	}

	return reports, nil
}

func (sh *StructHolder) suggestConstructorFix(pass *analysis.Pass) ([]analysis.SuggestedFix, error) {
	sortedConstructors := sh.copyAndSortConstructors()
	removingConstructorsTextEdit := make([]analysis.TextEdit, len(sh.Constructors))
	addingConstructorsTextEdit := make([]analysis.TextEdit, len(sh.Constructors))

	for i, constructor := range sortedConstructors {
		removingConstructorsTextEdit[i] = analysis.TextEdit{
			Pos:     GetStartingPos(constructor),
			End:     constructor.End(),
			NewText: make([]byte, 0),
		}

		constructorBytes, err := NodeToBytes(pass, constructor)
		if err != nil {
			return nil, err
		}

		addingConstructorsTextEdit[i] = analysis.TextEdit{
			Pos:     sh.Struct.End(),
			NewText: slices.Concat([]byte("\n\n"), constructorBytes),
		}
	}

	suggestedFixes := []analysis.SuggestedFix{
		{
			Message:   "Removing current constructors and adding them after struct declaration",
			TextEdits: slices.Concat(removingConstructorsTextEdit, addingConstructorsTextEdit),
		},
	}

	return suggestedFixes, nil
}

func (sh *StructHolder) copyAndSortConstructors() []*ast.FuncDecl {
	sortedConstructors := make([]*ast.FuncDecl, len(sh.Constructors))
	copy(sortedConstructors, sh.Constructors)

	if sh.Features.IsEnabled(AlphabeticalCheck) {
		slices.SortFunc(sortedConstructors, alphabeticalSort)
	}

	return sortedConstructors
}

func (sh *StructHolder) suggestMethodFix(pass *analysis.Pass) ([]analysis.SuggestedFix, error) {
	sortedExported, sortedUnexported := sh.copyAndSortMethods()

	removingMethodsTextEdit := make([]analysis.TextEdit, len(sh.StructMethods))
	addingMethodsTextEdit := make([]analysis.TextEdit, len(sh.StructMethods))

	for i, method := range sortedExported {
		removingMethodsTextEdit[i] = analysis.TextEdit{
			Pos:     GetStartingPos(method),
			End:     method.End(),
			NewText: make([]byte, 0),
		}

		methodBytes, err := NodeToBytes(pass, method)
		if err != nil {
			return nil, err
		}

		addingMethodsTextEdit[i] = analysis.TextEdit{
			Pos:     GetStartingPos(sh.StructMethods[0]),
			NewText: slices.Concat(methodBytes, []byte("\n\n")),
		}
	}

	for i, method := range sortedUnexported {
		removingMethodsTextEdit[i+len(sortedExported)] = analysis.TextEdit{
			Pos:     GetStartingPos(method),
			End:     method.End(),
			NewText: make([]byte, 0),
		}

		methodBytes, err := NodeToBytes(pass, method)
		if err != nil {
			return nil, err
		}

		addingMethodsTextEdit[i+len(sortedExported)] = analysis.TextEdit{
			Pos:     GetStartingPos(sh.StructMethods[0]),
			NewText: slices.Concat(methodBytes, []byte("\n\n")),
		}
	}

	suggestedFixes := []analysis.SuggestedFix{
		{
			Message:   "Removing current methods and adding them sorted",
			TextEdits: slices.Concat(removingMethodsTextEdit, addingMethodsTextEdit),
		},
	}

	return suggestedFixes, nil
}

func (sh *StructHolder) copyAndSortMethods() (ExportedMethods, UnexportedMethods) {
	exported, unexported := SplitExportedUnexported(sh.StructMethods)
	sortedExported := make([]*ast.FuncDecl, len(exported))
	sortedUnexported := make([]*ast.FuncDecl, len(unexported))

	copy(sortedExported, exported)
	copy(sortedUnexported, unexported)

	if sh.Features.IsEnabled(AlphabeticalCheck) {
		slices.SortFunc(sortedExported, alphabeticalSort)
		slices.SortFunc(sortedUnexported, alphabeticalSort)
	}

	return sortedExported, sortedUnexported
}

func sortDiagnostics(typeSpec *ast.TypeSpec, funcDecls []*ast.FuncDecl) []analysis.Diagnostic {
	var reports []analysis.Diagnostic

	for i := range funcDecls {
		if i >= len(funcDecls)-1 {
			continue
		}

		if funcDecls[i].Name.Name > funcDecls[i+1].Name.Name {
			reports = append(reports,
				NewAdjacentStructMethodsNotSortedAlphabetically(typeSpec, funcDecls[i], funcDecls[i+1]))
		}
	}

	return reports
}

func alphabeticalSort(a, b *ast.FuncDecl) int {
	return strings.Compare(a.Name.Name, b.Name.Name)
}
