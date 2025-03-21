package linters

import (
	"github.com/golangci/plugin-module-register/register"
	"github.com/manuelarte/gofuncor/pkg/analyzer"
	"golang.org/x/tools/go/analysis"
)

//nolint:gochecknoinits // linter
func init() {
	register.Plugin("gofuncor", New)
}

func New(_ any) (register.LinterPlugin, error) {
	return GoFuncor{}, nil
}

var _ register.LinterPlugin = new(GoFuncor)

type GoFuncor struct{}

func (g GoFuncor) BuildAnalyzers() ([]*analysis.Analyzer, error) {
	return []*analysis.Analyzer{analyzer.NewAnalyzer()}, nil
}

func (g GoFuncor) GetLoadMode() string {
	return register.LoadModeSyntax
}
