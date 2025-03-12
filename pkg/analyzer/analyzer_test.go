package analyzer_test

import (
	"github.com/manuelarte/gofuncor/pkg/analyzer"
	"golang.org/x/tools/go/analysis/analysistest"
	"os"
	"path/filepath"
	"testing"
)

func TestAll(t *testing.T) {
	path, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get wd: %s", err)
	}

	testdata := filepath.Join(filepath.Dir(filepath.Dir(path)), "testdata")
	analysistest.Run(t, testdata, analyzer.NewAnalyzer(), "simple")
}
