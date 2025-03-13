package analyzer

import (
	"flag"
	"fmt"
	"go/ast"
	"reflect"

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

	var lastFuncName string

	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			fmt.Printf("%+v\n", reflect.TypeOf(n))
			continueChild, _ := np.Process(n)
			if fn, ok := n.(*ast.FuncDecl); ok {
				funcName := fn.Name.Name

				// check for ordering rule
				if lastFuncName != "" && funcName < lastFuncName {
					pass.Reportf(fn.Pos(), "functions %q appears before %q but should be ordered differently", funcName, lastFuncName)
				}

				lastFuncName = funcName
				return false
			}
			return continueChild
		})
	}
	//nolint:nilnil //interface{}, error
	return nil, nil
}
