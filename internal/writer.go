package internal

import (
	"bytes"
	"fmt"
	"go/ast"
	"os"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"golang.org/x/tools/go/analysis"
)

const (
	// fileMode is the file permission mode for writing fixed files.
	fileMode = 0o600
)

// writeFormattedFile formats and writes the fixed file to disk.
func writeFormattedFile(pass *analysis.Pass, file *ast.File, dstFile *dst.File, filePath string) error {
	// Format the file using the restorer
	// dst preserves all decorations (comments) perfectly
	restorer := decorator.NewRestorer()

	var buf bytes.Buffer

	err := restorer.Fprint(&buf, dstFile)
	if err != nil {
		pass.Report(analysis.Diagnostic{
			Pos:     file.Pos(),
			Message: fmt.Sprintf("failed to format fixed file: %v", err),
		})

		return err
	}

	output := buf.Bytes()

	// Write to file
	err = os.WriteFile(filePath, output, fileMode)
	if err != nil {
		pass.Report(analysis.Diagnostic{
			Pos:     file.Pos(),
			Message: fmt.Sprintf("failed to write fixed file: %v", err),
		})

		return err
	}

	return nil
}
