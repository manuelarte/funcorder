package main

import (
	"golang.org/x/tools/go/analysis/singlechecker"

	"github.com/manuelarte/funcorder/analyzer"
)

func main() {
	singlechecker.Main(analyzer.NewAnalyzer())
}
