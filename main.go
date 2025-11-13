package main

import (
	"golang.org/x/tools/go/analysis/singlechecker"

	"github.com/manuelarte/funcorder/analyzer"
)

func main() {
	// Note: The analyzer defines a -fix flag for golangci-lint integration.
	// singlechecker.Main also provides a -fix flag, which may cause a conflict.
	// However, for golangci-lint usage, singlechecker is not used, so the flag works correctly.
	// If running as a standalone tool, singlechecker will handle the flag conflict.
	singlechecker.Main(analyzer.NewAnalyzer())
}
