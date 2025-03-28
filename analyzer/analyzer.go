package analyzer

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"

	"github.com/manuelarte/funcorder/internal/features"
	"github.com/manuelarte/funcorder/internal/fileprocessor"
)

func NewAnalyzer() *analysis.Analyzer {
	f := funcorder{}
	a := &analysis.Analyzer{
		Name: "funcorder",
		Doc:  "checks the order of functions, methods, and constructors",
		Run:  f.run,
	}
	a.Flags.BoolVar(&f.constructorCheck, "constructor-check", true,
		"enable/disable feature to check constructors are placed after struct declaration")
	a.Flags.BoolVar(&f.structMethodsCheck, "struct-methods-check", true,
		"enable/disable feature to check whether the exported struct's methods "+
			"are placed before the non-exported")

	return a
}

type funcorder struct {
	constructorCheck   bool
	structMethodsCheck bool
}

func (f *funcorder) run(pass *analysis.Pass) (any, error) {
	var enabledCheckers features.Feature
	if f.constructorCheck {
		enabledCheckers |= features.ConstructorCheck
	}
	if f.structMethodsCheck {
		enabledCheckers |= features.StructMethodsCheck
	}
	fp := fileprocessor.NewFileProcessor(enabledCheckers)
	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			if _, ok := n.(*ast.File); ok {
				errs := fp.Analyze()
				for _, err := range errs {
					pass.Report(analysis.Diagnostic{Pos: err.GetPos(), Message: err.Error()})
				}
			}
			continueChild := fp.Process(n)
			return continueChild
		})
	}
	errs := fp.Analyze()
	for _, err := range errs {
		pass.Report(analysis.Diagnostic{Pos: err.GetPos(), Message: err.Error()})
	}
	//nolint:nilnil //any, error
	return nil, nil
}
