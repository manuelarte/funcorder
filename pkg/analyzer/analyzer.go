package analyzer

import (
	"flag"
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

//nolint:gochecknoglobals // global variable
var flagSet flag.FlagSet

func NewAnalyzer() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name:  "gofuncor",
		Doc:   "checks function order",
		Run:   run,
		Flags: flagSet,
	}
}

func run(pass *analysis.Pass) (interface{}, error) {
	np := NewNodeProcessor()

	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			continueChild := np.Process(n)
			if _, ok := n.(*ast.File); ok {
				errs := np.Analyze()
				for _, err := range errs {
					pass.Report(analysis.Diagnostic{Pos: err.GetPos(), Message: err.Error()})
				}
			}
			return continueChild
		})
	}
	errs := np.Analyze()
	for _, err := range errs {
		pass.Report(analysis.Diagnostic{Pos: err.GetPos(), Message: err.Error()})
	}
	//nolint:nilnil //interface{}, error
	return nil, nil
}
