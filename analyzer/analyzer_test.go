package analyzer

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	testCases := []struct {
		desc     string
		patterns string
		options  map[string]string
	}{
		{
			desc:     "all",
			patterns: "simple",
		},
		{
			desc:     "constructor check only",
			patterns: "constructor-check",
			options: map[string]string{
				ConstructorCheckName:  "true",
				StructMethodCheckName: "false",
			},
		},
		{
			desc:     "method check only",
			patterns: "struct-method-check",
			options: map[string]string{
				ConstructorCheckName:  "false",
				StructMethodCheckName: "true",
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			a := NewAnalyzer()

			for k, v := range test.options {
				if err := a.Flags.Set(k, v); err != nil {
					t.Fatal(err)
				}
			}

			analysistest.Run(t, analysistest.TestData(), a, test.patterns)
		})
	}
}
