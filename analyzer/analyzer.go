package analyzer

import (
	"flag"
	"go/ast"

	"golang.org/x/tools/go/analysis"

	"github.com/manuelarte/funcorder/internal/fileprocessor"
)

//nolint:gochecknoglobals // global variable
var flagSet flag.FlagSet

func NewAnalyzer() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name:  "funcorder",
		Doc:   "checks function order",
		Run:   run,
		Flags: flagSet,
	}
}

func run(pass *analysis.Pass) (any, error) {
	fp := fileprocessor.NewFileProcessor()

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
	//nolint:nilnil //interface{}, error
	return nil, nil
}
