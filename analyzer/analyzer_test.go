package analyzer_test

import (
	"testing"

	"github.com/manuelarte/funcorder/analyzer"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAll(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), analyzer.NewAnalyzer(), "simple")
}
