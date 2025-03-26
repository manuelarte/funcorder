package analyzer_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/manuelarte/funcorder/pkg/analyzer"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAll(t *testing.T) {
	path, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get wd: %s", err)
	}

	testdata := filepath.Join(filepath.Dir(filepath.Dir(path)), "testdata")
	analysistest.Run(t, testdata, analyzer.NewAnalyzer(), "simple")
}
