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

func (f *funcorder) run(pass *analysis.Pass) (any, error) {
	insp, found := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !found {
		//nolint:nilnil // impossible case.
		return nil, nil
	}

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

	fp := internal.NewFileProcessor(enabledCheckers)

	nodeFilter := []ast.Node{
		(*ast.File)(nil),
		(*ast.FuncDecl)(nil),
		(*ast.TypeSpec)(nil),
	}

	// Collect all files for fixing
	var filesToFix []*ast.File

	insp.Preorder(nodeFilter, func(n ast.Node) {
		switch node := n.(type) {
		case *ast.File:
			// Collect file for later fixing
			if f.fix {
				filesToFix = append(filesToFix, node)
			}
			// Analyze the previous file's data (if not fixing)
			// Note: Preorder visits File before its children, so when we hit a new File,
			// we've already processed all children of the previous File
			if !f.fix {
				fp.Analyze(pass)
			}
			fp.ResetStructs()

		case *ast.FuncDecl:
			fp.AddFuncDecl(node)

		case *ast.TypeSpec:
			fp.AddTypeSpec(node)
		}
	})

	// Analyze the last file (if not in fix mode)
	// This is needed because we only analyze when encountering the next file
	if !f.fix {
		fp.Analyze(pass)
	}

	// Fix files after all declarations have been collected
	if f.fix {
		for _, file := range filesToFix {
			// Reset and recollect for this file
			fp.ResetStructs()
			// Re-collect declarations for this file
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
			internal.FixFile(pass, file, fp, enabledCheckers)
		}
	}

	//nolint:nilnil //any, error
	return nil, nil
}
