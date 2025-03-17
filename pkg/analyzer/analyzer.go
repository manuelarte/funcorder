package analyzer

import (
	"flag"
	"github.com/manuelarte/gofuncor/internal/fileprocessor"
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
	//nolint:nilnil //interface{}, error
	return nil, nil
}
