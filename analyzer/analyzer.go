package analyzer

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"

	"github.com/manuelarte/funcorder/internal"
)

const (
	ConstructorCheckName  = "constructor"
	StructMethodCheckName = "struct-method"
	AlphabeticalCheckName = "alphabetical"
)

func NewAnalyzer() *analysis.Analyzer {
	f := funcorder{}

	a := &analysis.Analyzer{
		Name:     "funcorder",
		Doc:      "checks the order of functions, methods, and constructors",
		URL:      "https://github.com/manuelarte/funcorder",
		Run:      f.run,
		Requires: []*analysis.Analyzer{inspect.Analyzer},
	}

	a.Flags.BoolVar(&f.constructorCheck, ConstructorCheckName, true,
		"Checks that constructors are placed after the structure declaration.")
	a.Flags.BoolVar(&f.structMethodCheck, StructMethodCheckName, true,
		"Checks if the exported methods of a structure are placed before the unexported ones.")
	a.Flags.BoolVar(&f.alphabeticalCheck, AlphabeticalCheckName, false,
		"Checks if the constructors and/or structure methods are sorted alphabetically.")
	a.Flags.BoolVar(&f.fix, "fix", false,
		"Automatically fix code layout issues by reordering functions, methods, and constructors.")

	return a
}

type funcorder struct {
	constructorCheck  bool
	structMethodCheck bool
	alphabeticalCheck bool
	fix               bool
}

// buildEnabledCheckers builds the enabled checkers feature set.
func (f *funcorder) buildEnabledCheckers() internal.Feature {
	var enabledCheckers internal.Feature

	if f.constructorCheck {
		enabledCheckers.Enable(internal.ConstructorCheck)
	}

	if f.structMethodCheck {
		enabledCheckers.Enable(internal.StructMethodCheck)
	}

	if f.alphabeticalCheck {
		enabledCheckers.Enable(internal.AlphabeticalCheck)
	}

	return enabledCheckers
}

func (f *funcorder) run(pass *analysis.Pass) (any, error) {
	insp, found := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !found {
		//nolint:nilnil // impossible case.
		return nil, nil
	}

	enabledCheckers := f.buildEnabledCheckers()
	fp := internal.NewFileProcessor(enabledCheckers)

	nodeFilter := []ast.Node{
		(*ast.File)(nil),
		(*ast.FuncDecl)(nil),
		(*ast.TypeSpec)(nil),
	}

	var filesToFix []*ast.File

	var currentFile *ast.File

	insp.Preorder(nodeFilter, func(n ast.Node) {
		filesToFix, currentFile = f.handleNode(n, pass, fp, f.fix, filesToFix, currentFile, enabledCheckers)
	})

	// Analyze the last file (if not in fix mode)
	if !f.fix && currentFile != nil {
		fp.Analyze(pass)
	}

	// Fix files after all declarations have been collected
	if f.fix {
		// Check the last file if we're in fix mode
		if currentFile != nil {
			// Recollect declarations for the last file to check if it needs fixing
			fp.ResetStructs()
			f.recollectDeclarations(currentFile, fp)

			if internal.NeedsFixing(currentFile, fp, enabledCheckers) {
				filesToFix = append(filesToFix, currentFile)
			}
		}

		f.fixFiles(pass, filesToFix, fp, enabledCheckers)
	}

	//nolint:nilnil //any, error
	return nil, nil
}

// handleNode processes a single AST node.
func (f *funcorder) handleNode(
	n ast.Node,
	pass *analysis.Pass,
	fp *internal.FileProcessor,
	fixMode bool,
	filesToFix []*ast.File,
	currentFile *ast.File,
	enabledCheckers internal.Feature,
) ([]*ast.File, *ast.File) {
	switch node := n.(type) {
	case *ast.File:
		// Analyze previous file before moving to next one
		if currentFile != nil {
			if !fixMode {
				fp.Analyze(pass)
			} else if internal.NeedsFixing(currentFile, fp, enabledCheckers) {
				// Check if the previous file needs fixing (before resetting)
				filesToFix = append(filesToFix, currentFile)
			}

			fp.ResetStructs()
		}

		currentFile = node

	case *ast.FuncDecl:
		fp.AddFuncDecl(node)

	case *ast.TypeSpec:
		fp.AddTypeSpec(node)
	}

	return filesToFix, currentFile
}

// fixFiles fixes all collected files.
func (f *funcorder) fixFiles(
	pass *analysis.Pass,
	filesToFix []*ast.File,
	fp *internal.FileProcessor,
	enabledCheckers internal.Feature,
) {
	for _, file := range filesToFix {
		fp.ResetStructs()
		f.recollectDeclarations(file, fp)
		internal.FixFile(pass, file, fp, enabledCheckers)
	}
}

// recollectDeclarations recollects declarations from a file.
func (f *funcorder) recollectDeclarations(file *ast.File, fp *internal.FileProcessor) {
	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			fp.AddFuncDecl(d)
		case *ast.GenDecl:
			for _, spec := range d.Specs {
				if typeSpec, ok := spec.(*ast.TypeSpec); ok {
					fp.AddTypeSpec(typeSpec)
				}
			}
		}
	}
}
