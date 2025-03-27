package analyzer

import (
	"flag"
	"go/ast"

	"github.com/manuelarte/funcorder/internal/features"

	"golang.org/x/tools/go/analysis"

	"github.com/manuelarte/funcorder/internal/fileprocessor"
)

//nolint:gochecknoglobals // global variable
var FlagSet flag.FlagSet

var (
	//nolint:gochecknoglobals // global variable
	constructorsCheck = FlagSet.Bool("constructors_check", true,
		"enable/disable constructors after struct check")
	//nolint:gochecknoglobals // global variable
	structMethodsCheck = FlagSet.Bool("struct_methods_check", true,
		"enable/disable exported struct's methods before non-exported")
)

func NewAnalyzer() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name:  "funcorder",
		Doc:   "checks the order of functions, methods, and constructors",
		Run:   run,
		Flags: FlagSet,
	}
}

func run(pass *analysis.Pass) (any, error) {
	var enabledCheckers features.Feature
	if *constructorsCheck {
		enabledCheckers |= features.ConstructorCheck
	}
	if *structMethodsCheck {
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
