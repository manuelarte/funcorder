package main

import (
	"github.com/golangci/plugin-module-register/register"
	"github.com/manuelarte/gofuncor/pkg/analyzer"
	"golang.org/x/tools/go/analysis"
)

func init() {
	register.Plugin("gofuncor", New)
}

//nolint:unparam // linter
func New(settings any) (register.LinterPlugin, error) {
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
