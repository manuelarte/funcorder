package diag

import (
	"fmt"
	"go/ast"
	"go/token"
	"slices"

	"golang.org/x/tools/go/analysis"

	"github.com/manuelarte/funcorder/internal/astutils"
)

func NewConstructorNotAfterStructType(fset *token.FileSet,
	structSpec *ast.TypeSpec, constructor *ast.FuncDecl) (analysis.Diagnostic, error) {
	suggestedFixConstructorByte, err := astutils.NodeToByteArray(fset, constructor)
	if err != nil {
		return analysis.Diagnostic{}, err
	}
	removingPos := constructor.Pos()
	if constructor.Doc != nil {
		removingPos = constructor.Doc.Pos()
	}
	return analysis.Diagnostic{
		Pos: constructor.Pos(),
		Message: fmt.Sprintf("constructor %q for struct %q should be placed after the struct declaration",
			constructor.Name, structSpec.Name),
		URL: "https://github.com/manuelarte/funcorder?tab=readme-ov-file#check-constructors-functions-are-placed-after-struct-declaration", //nolint: lll // url
		SuggestedFixes: []analysis.SuggestedFix{
			{
				Message: fmt.Sprintf("The constructor %q should be placed after the struct declaration", constructor.Name),
				TextEdits: []analysis.TextEdit{
					{
						Pos:     removingPos,
						End:     constructor.End(),
						NewText: make([]byte, 0),
					},
					{
						Pos: structSpec.Type.End(),
						End: token.NoPos,
						NewText: slices.Concat([]byte("\n"),
							[]byte("\n"),
							suggestedFixConstructorByte),
					},
				},
			},
		},
	}, nil
}

func NewConstructorNotBeforeStructMethod(
	structSpec *ast.TypeSpec,
	constructor *ast.FuncDecl,
	method *ast.FuncDecl,
) analysis.Diagnostic {
	return analysis.Diagnostic{
		Pos: constructor.Pos(),
		URL: "https://github.com/manuelarte/funcorder?tab=readme-ov-file#check-constructors-functions-are-placed-after-struct-declaration", //nolint: lll // url
		Message: fmt.Sprintf("constructor %q for struct %q should be placed before struct method %q",
			constructor.Name, structSpec.Name, method.Name),
	}
}

func NewAdjacentConstructorsNotSortedAlphabetically(
	structSpec *ast.TypeSpec,
	constructorNotSorted *ast.FuncDecl,
	otherConstructorNotSorted *ast.FuncDecl,
) analysis.Diagnostic {
	return analysis.Diagnostic{
		Pos: otherConstructorNotSorted.Pos(),
		URL: "https://github.com/manuelarte/funcorder?tab=readme-ov-file#check-constructorsmethods-are-sorted-alphabetically",
		Message: fmt.Sprintf("constructor %q for struct %q should be placed before constructor %q",
			otherConstructorNotSorted.Name, structSpec.Name, constructorNotSorted.Name),
	}
}

func NewNonExportedMethodBeforeExportedForStruct(
	structSpec *ast.TypeSpec,
	privateMethod *ast.FuncDecl,
	publicMethod *ast.FuncDecl,
) analysis.Diagnostic {
	return analysis.Diagnostic{
		Pos: privateMethod.Pos(),
		URL: "https://github.com/manuelarte/funcorder?tab=readme-ov-file#check-exported-methods-are-placed-before-non-exported-methods", //nolint: lll // url
		Message: fmt.Sprintf("unexported method %q for struct %q should be placed after the exported method %q",
			privateMethod.Name, structSpec.Name, publicMethod.Name),
	}
}

func NewAdjacentStructMethodsNotSortedAlphabetically(
	structSpec *ast.TypeSpec,
	method *ast.FuncDecl,
	otherMethod *ast.FuncDecl,
) analysis.Diagnostic {
	return analysis.Diagnostic{
		Pos: otherMethod.Pos(),
		URL: "https://github.com/manuelarte/funcorder?tab=readme-ov-file#check-constructorsmethods-are-sorted-alphabetically",
		Message: fmt.Sprintf("method %q for struct %q should be placed before method %q",
			otherMethod.Name, structSpec.Name, method.Name),
	}
}
