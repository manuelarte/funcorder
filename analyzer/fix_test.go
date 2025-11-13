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

// This is a standalone comment
// It has newlines on both sides

type MyStruct struct {
	Name string
}

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
		diff := computeDiff(expectedTrimmed, got)
		t.Errorf("Fix failed for %s\n\nDescription: %s\n\nDiff:\n%s\n\nFull Got:\n%s\n\nFull Expected:\n%s",
			desc, description, diff, got, expectedTrimmed)
	}

	// Verify comments are preserved (only if the input had comments)
	if strings.Contains(input, "Comment") && !strings.Contains(got, "Comment") {
		t.Errorf("Comments were lost in fix for %s", desc)
	}
}

// computeDiff computes a unified diff between expected and got.
func computeDiff(expected, got string) string {
	expectedLines := strings.Split(expected, "\n")
	gotLines := strings.Split(got, "\n")

	var diff strings.Builder
	diff.WriteString("--- expected\n")
	diff.WriteString("+++ got\n")
	diff.WriteString("@@ -1,0 +1,0 @@\n") // Placeholder, will be updated if we find differences

	// Simple unified diff algorithm
	i, j := 0, 0
	hasChanges := false

	for i < len(expectedLines) || j < len(gotLines) {
		if i < len(expectedLines) && j < len(gotLines) && expectedLines[i] == gotLines[j] {
			// Lines match, output context
			diff.WriteString(" ")
			diff.WriteString(expectedLines[i])
			diff.WriteString("\n")
			i++
			j++
		} else {
			hasChanges = true
			// Find the next matching line
			nextMatchI := i
			nextMatchJ := j
			foundMatch := false

			// Look ahead for matching lines
			for lookAhead := 1; lookAhead <= 3 && (i+lookAhead < len(expectedLines) || j+lookAhead < len(gotLines)); lookAhead++ {
				if i+lookAhead < len(expectedLines) && j < len(gotLines) && expectedLines[i+lookAhead] == gotLines[j] {
					nextMatchI = i + lookAhead
					nextMatchJ = j
					foundMatch = true
					break
				}
				if i < len(expectedLines) && j+lookAhead < len(gotLines) && expectedLines[i] == gotLines[j+lookAhead] {
					nextMatchI = i
					nextMatchJ = j + lookAhead
					foundMatch = true
					break
				}
			}

			// Output deletions
			for k := i; k < nextMatchI; k++ {
				diff.WriteString("-")
				diff.WriteString(expectedLines[k])
				diff.WriteString("\n")
			}

			// Output additions
			for k := j; k < nextMatchJ; k++ {
				diff.WriteString("+")
				diff.WriteString(gotLines[k])
				diff.WriteString("\n")
			}

			if foundMatch {
				i = nextMatchI
				j = nextMatchJ
			} else {
				// No match found, advance both
				if i < len(expectedLines) {
					diff.WriteString("-")
					diff.WriteString(expectedLines[i])
					diff.WriteString("\n")
					i++
				}
				if j < len(gotLines) {
					diff.WriteString("+")
					diff.WriteString(gotLines[j])
					diff.WriteString("\n")
					j++
				}
			}
		}
	}

	if !hasChanges {
		return "No differences found (but strings don't match - check whitespace)"
	}

	return diff.String()
}
