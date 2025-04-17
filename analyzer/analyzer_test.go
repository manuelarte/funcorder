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
			desc:     "default",
			patterns: "simple",
		},
		{
			desc:     "all options",
			patterns: "simple-alphabetical",
			options: map[string]string{
				AlphabeticalCheckName: "true",
			},
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
			desc:     "constructor check and alphabetical",
			patterns: "constructor-check-alphabetical",
			options: map[string]string{
				ConstructorCheckName:  "true",
				StructMethodCheckName: "false",
				AlphabeticalCheckName: "true",
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
		{
			desc:     "alphabetical method check",
			patterns: "struct-method-check-alphabetical",
			options: map[string]string{
				ConstructorCheckName:  "false",
				StructMethodCheckName: "true",
				AlphabeticalCheckName: "true",
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
