package main

import (
	"golang.org/x/tools/go/analysis/singlechecker"

	"github.com/manuelarte/funcorder/analyzer"
)

func main() {
	// singlechecker.Main automatically adds a -fix flag to analyzers.
	// The analyzer reads the flag value in run() from pass.Analyzer.Flags.
	// For golangci-lint usage, the flag is also handled automatically.
	a := analyzer.NewAnalyzer()
	singlechecker.Main(a)
}
