package main

import (
	"github.com/manuelarte/gofuncor/pkg/analyzer"
	"golang.org/x/tools/go/analysis"
)

//nolint:unparam // linter plugin
func New(_ any) ([]*analysis.Analyzer, error) {
	return []*analysis.Analyzer{analyzer.NewAnalyzer()}, nil
}
