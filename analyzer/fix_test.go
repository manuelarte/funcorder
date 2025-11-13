package analyzer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/tools/go/analysis"
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
		{
			desc: "imports preserved",
			input: `package fix

import (
	"fmt"
	"time"
)

func NewMyStruct() *MyStruct {
	return &MyStruct{Name: fmt.Sprintf("John-%d", time.Now().Unix())}
}

type MyStruct struct {
	Name string
}

func (m MyStruct) GetName() string {
	return m.Name
}
`,
			expected: `package fix

import (
	"fmt"
	"time"
)

type MyStruct struct {
	Name string
}

func NewMyStruct() *MyStruct {
	return &MyStruct{Name: fmt.Sprintf("John-%d", time.Now().Unix())}
}

func (m MyStruct) GetName() string {
	return m.Name
}
`,
			options: map[string]string{
				"fix": "true",
			},
			description: "Imports should be preserved in their correct position after package declaration",
		},
	}

	for _, test := range testCasesWithExpected {
		t.Run(test.desc, func(t *testing.T) {
			testFile := setupTestFile(t, test.input)
			a := setupAnalyzer(test.options)

			// Use analysistest to run the analyzer which will handle file fixing
			analysistest.Run(t, filepath.Dir(testFile), a)

			verifyFixedFile(t, testFile, test.expected, test.desc, test.description, test.input)
		})
	}
}

// setupTestFile creates a temporary test file with the given input.
func setupTestFile(t *testing.T, input string) string {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")

	err := os.WriteFile(testFile, []byte(input), 0o644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	return testFile
}

// setupAnalyzer creates and configures the analyzer with the given options.
func setupAnalyzer(options map[string]string) *analysis.Analyzer {
	a := NewAnalyzer()

	for k, v := range options {
		err := a.Flags.Set(k, v)
		if err != nil {
			// This shouldn't happen in tests, but handle it gracefully
			panic(err)
		}
	}

	return a
}

// verifyFixedFile verifies that the fixed file matches the expected output.
func verifyFixedFile(t *testing.T, testFile, expected, desc, description, input string) {
	fixedContent, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read fixed file: %v", err)
	}

	// Normalize whitespace for comparison
	got := strings.TrimSpace(string(fixedContent))
	expectedTrimmed := strings.TrimSpace(expected)

	if got != expectedTrimmed {
		t.Errorf("Fix failed for %s\n\nGot:\n%s\n\nExpected:\n%s\n\nDescription: %s",
			desc, got, expectedTrimmed, description)
	}

	// Verify comments are preserved (only if the input had comments)
	if strings.Contains(input, "Comment") && !strings.Contains(got, "Comment") {
		t.Errorf("Comments were lost in fix for %s", desc)
	}
}
