package analyzer

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"

	"github.com/manuelarte/funcorder/internal/features"
	"github.com/manuelarte/funcorder/internal/fileprocessor"
)

func NewAnalyzer() *analysis.Analyzer {
	f := funcorder{}
	a := &analysis.Analyzer{
		Name:     "funcorder",
		Doc:      "checks the order of functions, methods, and constructors",
		Run:      f.run,
		Requires: []*analysis.Analyzer{inspect.Analyzer},
	}
	a.Flags.BoolVar(&f.constructorCheck, "constructor-check", true,
		"enable/disable feature to check constructors are placed after struct declaration")
	a.Flags.BoolVar(&f.structMethodCheck, "struct-method-check", true,
		"enable/disable feature to check whether the exported struct's methods "+
			"are placed before the non-exported")

	return a
}

type funcorder struct {
	constructorCheck  bool
	structMethodCheck bool
}

func (f *funcorder) run(pass *analysis.Pass) (any, error) {
	var enabledCheckers features.Feature
	if f.constructorCheck {
		enabledCheckers |= features.ConstructorCheck
	}
	if f.structMethodCheck {
		enabledCheckers |= features.StructMethodCheck
	}

	fp := fileprocessor.NewFileProcessor(enabledCheckers)

	insp, found := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !found {
		//nolint:nilnil // impossible case.
		return nil, nil
	}

	nodeFilter := []ast.Node{
		(*ast.File)(nil),
		(*ast.FuncDecl)(nil),
		(*ast.TypeSpec)(nil),
	}

	insp.Preorder(nodeFilter, func(n ast.Node) {
		if _, ok := n.(*ast.File); ok {
			for _, report := range fp.Analyze() {
				pass.Report(report)
			}
		}

		fp.Process(n)
	})

	for _, report := range fp.Analyze() {
		pass.Report(report)
	}

	//nolint:nilnil //any, error
	return nil, nil
}
