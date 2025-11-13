package internal

import (
	"bytes"
	"cmp"
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"os"
	"slices"

	"golang.org/x/tools/go/analysis"
)

// FixFile fixes code layout issues in a file by reordering declarations.
func FixFile(pass *analysis.Pass, file *ast.File, fp *FileProcessor, features Feature) {
	// Collect all declarations with their positions
	// Note: fp already has all struct information from AddFuncDecl/AddTypeSpec calls
	decls := collectDeclarations(file, fp)

	// Determine the correct order
	orderedDecls := determineOrder(decls, fp, features)

	// Check if reordering is needed
	if !needsReordering(decls, orderedDecls) {
		return
	}

	// Reorder declarations in the AST
	reorderDeclarations(file, orderedDecls)

	// Write the fixed code back to the file
	writeFixedFile(pass, file)
}

// declaration represents a top-level declaration with its associated comments.
type declaration struct {
	node     ast.Decl
	comments []*ast.CommentGroup
	pos      token.Pos
}

// collectDeclarations collects all top-level declarations from the file.
func collectDeclarations(file *ast.File, fp *FileProcessor) []declaration {
	decls := make([]declaration, 0, len(file.Decls))

	// Find the comment group immediately preceding each declaration
	// Comments are already associated via Doc field, but we also need standalone comments
	for i, d := range file.Decls {
		var comments []*ast.CommentGroup

		// First, include existing Doc comments (they're already attached to nodes)
		switch node := d.(type) {
		case *ast.GenDecl:
			if node.Doc != nil {
				comments = append(comments, node.Doc)
			}
		case *ast.FuncDecl:
			if node.Doc != nil {
				comments = append(comments, node.Doc)
			}
		}

		// Find standalone comments between this declaration and the previous one
		var prevEnd token.Pos
		if i > 0 {
			prevEnd = file.Decls[i-1].End()
		} else {
			// Find the end of the package declaration - approximate position
			// Package declaration ends after "package" keyword + package name
			prevEnd = file.Package + token.Pos(20) // rough estimate
		}

		// Collect standalone comments between prevEnd and current declaration
		// We want to associate comments that are closer to this declaration than the next one
		var nextStart token.Pos
		hasNext := false
		if i < len(file.Decls)-1 {
			nextStart = file.Decls[i+1].Pos()
			hasNext = true
		}

		for _, cg := range file.Comments {
			// Skip if this comment is already a Doc comment (it's attached to the node)
			isDoc := false
			switch node := d.(type) {
			case *ast.GenDecl:
				if node.Doc == cg {
					isDoc = true
				}
			case *ast.FuncDecl:
				if node.Doc == cg {
					isDoc = true
				}
			}
			if isDoc {
				continue
			}

			// Include comments that are between previous declaration and current one
			// AND closer to this declaration than the next one (if next exists)
			// This ensures standalone comments move with the declaration they're associated with
			if cg.Pos() > prevEnd && cg.End() <= d.Pos() {
				shouldInclude := true
				if hasNext {
					// Check if this comment is closer to current declaration than next one
					distToCurrent := d.Pos() - cg.End()
					distToNext := nextStart - cg.End()
					if distToNext >= 0 && distToNext < distToCurrent {
						// Comment is closer to next declaration, don't include it here
						shouldInclude = false
					}
				}
				if shouldInclude {
					comments = append(comments, cg)
				}
			}
		}

		decls = append(decls, declaration{
			node:     d,
			comments: comments,
			pos:      d.Pos(),
		})
	}

	// Sort by original position
	slices.SortFunc(decls, func(a, b declaration) int {
		return cmp.Compare(a.pos, b.pos)
	})

	return decls
}

// determineOrder determines the correct order of declarations based on enabled features.
func determineOrder(decls []declaration, fp *FileProcessor, features Feature) []declaration {
	// Separate declarations into categories
	var otherDecls []declaration

	// Map struct names to their declarations
	structMap := make(map[string]declaration)
	structOrder := make([]string, 0)

	// Map struct names to their constructors and methods
	constructorsByStruct := make(map[string][]declaration)
	methodsByStruct := make(map[string][]declaration)

	for _, d := range decls {
		switch node := d.node.(type) {
		case *ast.GenDecl:
			if node.Tok == token.TYPE {
				for _, spec := range node.Specs {
					if typeSpec, ok := spec.(*ast.TypeSpec); ok {
						if _, ok := typeSpec.Type.(*ast.StructType); ok {
							structName := typeSpec.Name.Name
							structMap[structName] = d
							structOrder = append(structOrder, structName)
						}
					}
				}
			} else {
				otherDecls = append(otherDecls, d)
			}

		case *ast.FuncDecl:
			// Check if it's a constructor
			if sc := NewStructConstructor(node); sc != nil {
				structName := sc.StructReturn.Name
				constructorsByStruct[structName] = append(constructorsByStruct[structName], d)
				continue
			}

			// Check if it's a method
			if structName := funcIsMethod(node); structName != nil {
				methodsByStruct[structName.Name] = append(methodsByStruct[structName.Name], d)
				continue
			}

			// Other functions
			otherDecls = append(otherDecls, d)
		}
	}

	// Build ordered list
	// Process structs in their original order, grouping each with its constructors and methods
	ordered := make([]declaration, 0, len(decls))
	added := make(map[token.Pos]bool)

	// Find the first struct position
	firstStructPos := token.Pos(0)
	if len(structOrder) > 0 {
		firstStructPos = structMap[structOrder[0]].pos
	}

	// Add other declarations that come before the first struct
	slices.SortFunc(otherDecls, func(a, b declaration) int {
		return cmp.Compare(a.pos, b.pos)
	})
	for _, d := range otherDecls {
		if d.pos < firstStructPos || firstStructPos == 0 {
			ordered = append(ordered, d)
			added[d.pos] = true
		}
	}

	// Process each struct in order
	for _, structName := range structOrder {
		structDecl := structMap[structName]
		sh, exists := fp.structs[structName]
		if !exists || sh.Struct == nil {
			// Struct not in our map, add it as-is
			ordered = append(ordered, structDecl)
			added[structDecl.pos] = true
			continue
		}

		// Add struct declaration
		ordered = append(ordered, structDecl)
		added[structDecl.pos] = true

		// Get constructors for this struct
		structConstructors := constructorsByStruct[structName]
		if len(structConstructors) > 0 {
			// Sort constructors if alphabetical check is enabled
			if features.IsEnabled(AlphabeticalCheck) {
				slices.SortFunc(structConstructors, func(a, b declaration) int {
					aName := a.node.(*ast.FuncDecl).Name.Name
					bName := b.node.(*ast.FuncDecl).Name.Name
					return cmp.Compare(aName, bName)
				})
			} else {
				// Keep original order
				slices.SortFunc(structConstructors, func(a, b declaration) int {
					return cmp.Compare(a.pos, b.pos)
				})
			}
			for _, c := range structConstructors {
				ordered = append(ordered, c)
				added[c.pos] = true
			}
		}

		// Get methods for this struct
		structMethods := methodsByStruct[structName]
		if len(structMethods) > 0 {
			// Split into exported and unexported
			exported, unexported := splitExportedUnexportedDecls(structMethods)

			// Sort each group if alphabetical check is enabled
			if features.IsEnabled(AlphabeticalCheck) {
				slices.SortFunc(exported, func(a, b declaration) int {
					aName := a.node.(*ast.FuncDecl).Name.Name
					bName := b.node.(*ast.FuncDecl).Name.Name
					return cmp.Compare(aName, bName)
				})
				slices.SortFunc(unexported, func(a, b declaration) int {
					aName := a.node.(*ast.FuncDecl).Name.Name
					bName := b.node.(*ast.FuncDecl).Name.Name
					return cmp.Compare(aName, bName)
				})
			} else {
				// Keep original order within groups
				slices.SortFunc(exported, func(a, b declaration) int {
					return cmp.Compare(a.pos, b.pos)
				})
				slices.SortFunc(unexported, func(a, b declaration) int {
					return cmp.Compare(a.pos, b.pos)
				})
			}

			// Add exported methods first, then unexported
			for _, m := range exported {
				ordered = append(ordered, m)
				added[m.pos] = true
			}
			for _, m := range unexported {
				ordered = append(ordered, m)
				added[m.pos] = true
			}
		}
	}

	// Add other declarations (const, var, non-struct types, etc.)
	// Sort them by original position to preserve relative order
	slices.SortFunc(otherDecls, func(a, b declaration) int {
		return cmp.Compare(a.pos, b.pos)
	})
	for _, d := range otherDecls {
		if !added[d.pos] {
			ordered = append(ordered, d)
			added[d.pos] = true
		}
	}

	return ordered
}

// splitExportedUnexportedDecls splits declarations into exported and unexported methods.
func splitExportedUnexportedDecls(decls []declaration) (exported, unexported []declaration) {
	for _, d := range decls {
		if funcDecl, ok := d.node.(*ast.FuncDecl); ok {
			if funcDecl.Name.IsExported() {
				exported = append(exported, d)
			} else {
				unexported = append(unexported, d)
			}
		}
	}
	return exported, unexported
}

// needsReordering checks if reordering is needed.
func needsReordering(original, ordered []declaration) bool {
	if len(original) != len(ordered) {
		return true
	}

	for i := range original {
		if original[i].pos != ordered[i].pos {
			return true
		}
	}

	return false
}

// reorderDeclarations reorders declarations in the AST file.
func reorderDeclarations(file *ast.File, ordered []declaration) {
	// Create new declarations list
	newDecls := make([]ast.Decl, 0, len(ordered))

	// Add declarations in order, ensuring comments are attached as Doc comments
	// This ensures comments move with their declarations
	for _, d := range ordered {
		// Ensure comments are attached as Doc comments
		// If there's already a Doc comment, keep it; otherwise use the first comment from our collection
		switch node := d.node.(type) {
		case *ast.GenDecl:
			if node.Doc == nil && len(d.comments) > 0 {
				// Use the first comment (could be standalone or existing Doc)
				node.Doc = d.comments[0]
			}
		case *ast.FuncDecl:
			if node.Doc == nil && len(d.comments) > 0 {
				// Use the first comment (could be standalone or existing Doc)
				node.Doc = d.comments[0]
			}
		}

		newDecls = append(newDecls, d.node)
	}

	file.Decls = newDecls

	// Rebuild comments list - only include comments that are NOT Doc comments
	// go/format automatically places Doc comments before their declarations
	// We only need to keep standalone comments that aren't attached to declarations
	docComments := make(map[*ast.CommentGroup]bool)
	for _, d := range ordered {
		switch node := d.node.(type) {
		case *ast.GenDecl:
			if node.Doc != nil {
				docComments[node.Doc] = true
			}
		case *ast.FuncDecl:
			if node.Doc != nil {
				docComments[node.Doc] = true
			}
		}
	}

	// Only keep comments that aren't Doc comments
	newComments := make([]*ast.CommentGroup, 0)
	for _, cg := range file.Comments {
		if !docComments[cg] {
			newComments = append(newComments, cg)
		}
	}

	file.Comments = newComments
}

// writeFixedFile writes the fixed AST back to the file.
func writeFixedFile(pass *analysis.Pass, file *ast.File) {
	// Get the file path
	fset := pass.Fset
	filePath := fset.File(file.Pos()).Name()

	if filePath == "" {
		return
	}

	// Format the file
	var buf bytes.Buffer
	if err := format.Node(&buf, fset, file); err != nil {
		pass.Report(analysis.Diagnostic{
			Pos:     file.Pos(),
			Message: fmt.Sprintf("failed to format fixed file: %v", err),
		})
		return
	}

	// Write to file
	if err := os.WriteFile(filePath, buf.Bytes(), 0o644); err != nil {
		pass.Report(analysis.Diagnostic{
			Pos:     file.Pos(),
			Message: fmt.Sprintf("failed to write fixed file: %v", err),
		})
		return
	}
}
