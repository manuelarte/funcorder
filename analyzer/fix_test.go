package analyzer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestFix(t *testing.T) {
	// Test cases with specific expected outputs
	// These cover edge cases not present in testdata/src:
	// - Comment preservation
	// - Other top-level declarations (const, var)
	// - Standalone comments with newlines
	testCasesWithExpected := []struct {
		desc        string
		input       string
		expected    string
		options     map[string]string
		description string
	}{
		// Keep only edge cases not covered by testdata:
		// 1. Comments preservation (not in testdata)
		// 2. Other top-level declarations (const, var) - not in testdata
		// 3. Standalone comments with newlines - not in testdata
		{
			desc: "comments preservation",
			input: `package fix

// Comment before constructor
func NewMyStruct() *MyStruct {
	return &MyStruct{Name: "John"}
}

// Comment before struct
type MyStruct struct {
	Name string
}

func (m MyStruct) GetName() string {
	return m.Name
}
`,
			expected: `package fix

// Comment before struct
type MyStruct struct {
	Name string
}

// Comment before constructor
func NewMyStruct() *MyStruct {
	return &MyStruct{Name: "John"}
}

func (m MyStruct) GetName() string {
	return m.Name
}
`,
			options: map[string]string{
				"fix": "true",
			},
			description: "Comments should be preserved when reordering",
		},
		{
			desc: "other top level declarations preserved",
			input: `package fix

const ConstValue = 42

func NewMyStruct() *MyStruct {
	return &MyStruct{}
}

type MyStruct struct {
	Name string
}

var GlobalVar = "test"
`,
			expected: `package fix

const ConstValue = 42

type MyStruct struct {
	Name string
}

func NewMyStruct() *MyStruct {
	return &MyStruct{}
}

var GlobalVar = "test"
`,
			options: map[string]string{
				"fix": "true",
			},
			description: "Other top-level declarations (const, var) should be preserved",
		},
		{
			desc: "standalone comment with newlines",
			input: `package fix

// This is a standalone comment
// It has newlines on both sides

func NewMyStruct() *MyStruct {
	return &MyStruct{}
}

type MyStruct struct {
	Name string
}

// Another standalone comment block
// Not attached to any symbol

func (m MyStruct) GetName() string {
	return m.Name
}
`,
			expected: `package fix

type MyStruct struct {
	Name string
}

// This is a standalone comment
// It has newlines on both sides

func NewMyStruct() *MyStruct {
	return &MyStruct{}
}

// Another standalone comment block
// Not attached to any symbol

func (m MyStruct) GetName() string {
	return m.Name
}
`,
			options: map[string]string{
				"fix": "true",
			},
			description: "Standalone comments with newlines on both sides should be preserved",
		},
	}

	for _, test := range testCasesWithExpected {
		t.Run(test.desc, func(t *testing.T) {
			// Create temporary directory
			tmpDir := t.TempDir()
			testFile := filepath.Join(tmpDir, "test.go")

			// Write input file
			if err := os.WriteFile(testFile, []byte(test.input), 0o644); err != nil {
				t.Fatalf("Failed to write test file: %v", err)
			}

			// Create analyzer with fix enabled
			a := NewAnalyzer()
			for k, v := range test.options {
				if err := a.Flags.Set(k, v); err != nil {
					t.Fatalf("Failed to set flag %s=%s: %v", k, v, err)
				}
			}

			// Use analysistest to run the analyzer which will handle file fixing
			// We'll check the file contents after
			analysistest.Run(t, tmpDir, a)

			// Read the fixed file
			fixedContent, err := os.ReadFile(testFile)
			if err != nil {
				t.Fatalf("Failed to read fixed file: %v", err)
			}

			// Normalize whitespace for comparison
			got := strings.TrimSpace(string(fixedContent))
			expected := strings.TrimSpace(test.expected)

			if got != expected {
				t.Errorf("Fix failed for %s\n\nGot:\n%s\n\nExpected:\n%s\n\nDescription: %s",
					test.desc, got, expected, test.description)
			}

			// Verify comments are preserved (only if the input had comments)
			if strings.Contains(test.input, "Comment") && !strings.Contains(got, "Comment") {
				t.Errorf("Comments were lost in fix for %s", test.desc)
			}
		})
	}
}
