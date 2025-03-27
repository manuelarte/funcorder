package analyzer_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/manuelarte/funcorder/analyzer"
)

func TestAll(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), analyzer.NewAnalyzer(), "simple")
}

func TestConstructorCheckOnly(t *testing.T) {
	a := analyzer.NewAnalyzer()
	if err := a.Flags.Set("constructors_check", "true"); err != nil {
		t.Fatal(err)
	}
	if err := a.Flags.Set("struct_methods_check", "false"); err != nil {
		t.Fatal(err)
	}
	analysistest.Run(t, analysistest.TestData(), a, "constructor_check")
}

func TestStructMethodsCheckOnly(t *testing.T) {
	a := analyzer.NewAnalyzer()
	if err := a.Flags.Set("constructors_check", "false"); err != nil {
		t.Fatal(err)
	}
	if err := a.Flags.Set("struct_methods_check", "true"); err != nil {
		t.Fatal(err)
	}
	analysistest.Run(t, analysistest.TestData(), a, "struct_methods_check")
}
