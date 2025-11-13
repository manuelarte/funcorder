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

const (
	// packageDeclEstimate is a rough estimate for the end position of a package declaration.
	packageDeclEstimate = 20
	// fileMode is the file permission mode for writing fixed files.
	fileMode = 0o600
)

// NeedsFixing checks if a file needs fixing by comparing current order with desired order.
func NeedsFixing(file *ast.File, fp *FileProcessor, features Feature) bool {
	// Collect all declarations with their positions
	// Note: fp already has all struct information from AddFuncDecl/AddTypeSpec calls
	decls := collectDeclarations(file, fp)

	// Determine the correct order
	orderedDecls := determineOrder(decls, fp, features)

	// Check if reordering is needed
	return needsReordering(decls, orderedDecls)
}

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
func collectDeclarations(file *ast.File, _ *FileProcessor) []declaration {
	decls := make([]declaration, 0, len(file.Decls))

	// Find the comment group immediately preceding each declaration
	// Comments are already associated via Doc field, but we also need standalone comments
	for i, d := range file.Decls {
		comments := getDocComments(d)
		prevEnd := getPreviousEnd(file, i)
		nextStart, hasNext := getNextStart(file, i)

		standaloneComments := collectStandaloneComments(file.Comments, d, prevEnd, nextStart, hasNext)
		comments = append(comments, standaloneComments...)

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

// getDocComments extracts doc comments from a declaration.
func getDocComments(d ast.Decl) []*ast.CommentGroup {
	var comments []*ast.CommentGroup

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

	return comments
}

// getPreviousEnd gets the end position of the previous declaration.
func getPreviousEnd(file *ast.File, i int) token.Pos {
	if i > 0 {
		return file.Decls[i-1].End()
	}

	// Find the end of the package declaration - approximate position
	// Package declaration ends after "package" keyword + package name
	return file.Package + token.Pos(packageDeclEstimate)
}

// getNextStart gets the start position of the next declaration.
func getNextStart(file *ast.File, i int) (token.Pos, bool) {
	if i < len(file.Decls)-1 {
		return file.Decls[i+1].Pos(), true
	}

	return 0, false
}

// isDocComment checks if a comment group is a doc comment for a declaration.
func isDocComment(cg *ast.CommentGroup, d ast.Decl) bool {
	switch node := d.(type) {
	case *ast.GenDecl:
		return node.Doc == cg
	case *ast.FuncDecl:
		return node.Doc == cg
	}

	return false
}

// collectStandaloneComments collects standalone comments associated with a declaration.
func collectStandaloneComments(
	fileComments []*ast.CommentGroup,
	d ast.Decl,
	prevEnd, nextStart token.Pos,
	hasNext bool,
) []*ast.CommentGroup {
	var comments []*ast.CommentGroup

	for _, cg := range fileComments {
		// Skip if this comment is already a Doc comment (it's attached to the node)
		if isDocComment(cg, d) {
			continue
		}

		// Include comments that are between previous declaration and current one
		// AND closer to this declaration than the next one (if next exists)
		// This ensures standalone comments move with the declaration they're associated with
		if cg.Pos() > prevEnd && cg.End() <= d.Pos() {
			if shouldIncludeComment(cg, d, nextStart, hasNext) {
				comments = append(comments, cg)
			}
		}
	}

	return comments
}

// shouldIncludeComment determines if a comment should be included with a declaration.
func shouldIncludeComment(cg *ast.CommentGroup, d ast.Decl, nextStart token.Pos, hasNext bool) bool {
	if !hasNext {
		return true
	}

	// Check if this comment is closer to current declaration than next one
	distToCurrent := d.Pos() - cg.End()
	distToNext := nextStart - cg.End()

	// Comment is closer to next declaration, don't include it here
	return distToNext < 0 || distToNext >= distToCurrent
}

// declarationCategories holds categorized declarations.
type declarationCategories struct {
	structMap            map[string]declaration
	structOrder          []string
	constructorsByStruct map[string][]declaration
	methodsByStruct      map[string][]declaration
	otherDecls           []declaration
}

// determineOrder determines the correct order of declarations based on enabled features.
func determineOrder(decls []declaration, fp *FileProcessor, features Feature) []declaration {
	categories := categorizeDeclarations(decls)

	ordered := make([]declaration, 0, len(decls))
	added := make(map[token.Pos]bool)

	firstStructPos := getFirstStructPos(categories.structMap, categories.structOrder)
	newDecls := addDeclarationsBeforeFirstStruct(categories.otherDecls, firstStructPos, added)
	ordered = append(ordered, newDecls...)

	// Process each struct in order
	for _, structName := range categories.structOrder {
		structDecls := addStructWithRelatedDeclarations(
			structName,
			categories.structMap,
			categories.constructorsByStruct,
			categories.methodsByStruct,
			fp,
			features,
			added,
		)
		ordered = append(ordered, structDecls...)
	}

	remainingDecls := addRemainingDeclarations(categories.otherDecls, added)
	ordered = append(ordered, remainingDecls...)

	return ordered
}

// categorizeDeclarations separates declarations into categories.
func categorizeDeclarations(decls []declaration) declarationCategories {
	categories := declarationCategories{
		structMap:            make(map[string]declaration),
		structOrder:          make([]string, 0),
		constructorsByStruct: make(map[string][]declaration),
		methodsByStruct:      make(map[string][]declaration),
		otherDecls:           make([]declaration, 0),
	}

	for _, d := range decls {
		switch node := d.node.(type) {
		case *ast.GenDecl:
			if node.Tok == token.TYPE {
				categorizeTypeDecl(node, d, &categories)
			} else {
				categories.otherDecls = append(categories.otherDecls, d)
			}

		case *ast.FuncDecl:
			categorizeFuncDecl(node, d, &categories)
		}
	}

	return categories
}

// categorizeTypeDecl categorizes a type declaration.
func categorizeTypeDecl(node *ast.GenDecl, d declaration, categories *declarationCategories) {
	for _, spec := range node.Specs {
		if typeSpec, ok := spec.(*ast.TypeSpec); ok {
			if _, isStruct := typeSpec.Type.(*ast.StructType); isStruct {
				structName := typeSpec.Name.Name
				categories.structMap[structName] = d
				categories.structOrder = append(categories.structOrder, structName)
			}
		}
	}
}

// categorizeFuncDecl categorizes a function declaration.
func categorizeFuncDecl(node *ast.FuncDecl, d declaration, categories *declarationCategories) {
	// Check if it's a constructor
	if sc := NewStructConstructor(node); sc != nil {
		structName := sc.StructReturn.Name
		categories.constructorsByStruct[structName] = append(categories.constructorsByStruct[structName], d)

		return
	}

	// Check if it's a method
	if structName := funcIsMethod(node); structName != nil {
		categories.methodsByStruct[structName.Name] = append(categories.methodsByStruct[structName.Name], d)

		return
	}

	// Other functions
	categories.otherDecls = append(categories.otherDecls, d)
}

// getFirstStructPos gets the position of the first struct.
func getFirstStructPos(structMap map[string]declaration, structOrder []string) token.Pos {
	if len(structOrder) > 0 {
		return structMap[structOrder[0]].pos
	}

	return token.Pos(0)
}

// addDeclarationsBeforeFirstStruct adds declarations that come before the first struct.
func addDeclarationsBeforeFirstStruct(
	otherDecls []declaration,
	firstStructPos token.Pos,
	added map[token.Pos]bool,
) []declaration {
	slices.SortFunc(otherDecls, func(a, b declaration) int {
		return cmp.Compare(a.pos, b.pos)
	})

	result := make([]declaration, 0, len(otherDecls))

	for _, d := range otherDecls {
		if d.pos < firstStructPos || firstStructPos == 0 {
			result = append(result, d)
			added[d.pos] = true
		}
	}

	return result
}

// addStructWithRelatedDeclarations adds a struct with its constructors and methods.
func addStructWithRelatedDeclarations(
	structName string,
	structMap map[string]declaration,
	constructorsByStruct map[string][]declaration,
	methodsByStruct map[string][]declaration,
	fp *FileProcessor,
	features Feature,
	added map[token.Pos]bool,
) []declaration {
	structDecl := structMap[structName]
	sh, exists := fp.structs[structName]

	var result []declaration

	if !exists || sh.Struct == nil {
		// Struct not in our map, add it as-is
		result = append(result, structDecl)
		added[structDecl.pos] = true

		return result
	}

	// Add struct declaration
	result = append(result, structDecl)
	added[structDecl.pos] = true

	// Add constructors
	structConstructors := constructorsByStruct[structName]
	if len(structConstructors) > 0 {
		newDecls := addSortedConstructors(structConstructors, features, added)
		result = append(result, newDecls...)
	}

	// Add methods
	structMethods := methodsByStruct[structName]
	if len(structMethods) > 0 {
		newDecls := addSortedMethods(structMethods, features, added)
		result = append(result, newDecls...)
	}

	return result
}

// addSortedConstructors sorts and adds constructors to the ordered list.
func addSortedConstructors(
	constructors []declaration,
	features Feature,
	added map[token.Pos]bool,
) []declaration {
	// Sort constructors if alphabetical check is enabled
	if features.IsEnabled(AlphabeticalCheck) {
		sortDeclarationsByName(constructors)
	} else {
		// Keep original order
		slices.SortFunc(constructors, func(a, b declaration) int {
			return cmp.Compare(a.pos, b.pos)
		})
	}

	result := make([]declaration, 0, len(constructors))

	for _, c := range constructors {
		result = append(result, c)
		added[c.pos] = true
	}

	return result
}

// addRemainingDeclarations adds remaining declarations that weren't added yet.
func addRemainingDeclarations(
	otherDecls []declaration,
	added map[token.Pos]bool,
) []declaration {
	// Sort them by original position to preserve relative order
	slices.SortFunc(otherDecls, func(a, b declaration) int {
		return cmp.Compare(a.pos, b.pos)
	})

	result := make([]declaration, 0, len(otherDecls))

	for _, d := range otherDecls {
		if !added[d.pos] {
			result = append(result, d)
			added[d.pos] = true
		}
	}

	return result
}

// splitExportedUnexportedDecls splits declarations into exported and unexported methods.
func splitExportedUnexportedDecls(decls []declaration) ([]declaration, []declaration) {
	var exported, unexported []declaration

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

// addSortedMethods sorts and adds methods to the ordered list.
func addSortedMethods(structMethods []declaration, features Feature, added map[token.Pos]bool) []declaration {
	// Split into exported and unexported
	exported, unexported := splitExportedUnexportedDecls(structMethods)

	// Sort each group if alphabetical check is enabled
	if features.IsEnabled(AlphabeticalCheck) {
		sortDeclarationsByName(exported)
		sortDeclarationsByName(unexported)
	} else {
		// Keep original order within groups
		slices.SortFunc(exported, func(a, b declaration) int {
			return cmp.Compare(a.pos, b.pos)
		})
		slices.SortFunc(unexported, func(a, b declaration) int {
			return cmp.Compare(a.pos, b.pos)
		})
	}

	result := make([]declaration, 0, len(structMethods))

	// Add exported methods first, then unexported
	for _, m := range exported {
		result = append(result, m)
		added[m.pos] = true
	}

	for _, m := range unexported {
		result = append(result, m)
		added[m.pos] = true
	}

	return result
}

// sortDeclarationsByName sorts declarations by their function name.
func sortDeclarationsByName(decls []declaration) {
	slices.SortFunc(decls, func(a, b declaration) int {
		aFuncDecl, okA := a.node.(*ast.FuncDecl)
		bFuncDecl, okB := b.node.(*ast.FuncDecl)

		if !okA || !okB {
			return 0
		}

		aName := aFuncDecl.Name.Name
		bName := bFuncDecl.Name.Name

		return cmp.Compare(aName, bName)
	})
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
func reorderDeclarations(file *ast.File, ordered []declaration) { //nolint:gocognit
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

	err := format.Node(&buf, fset, file)
	if err != nil {
		pass.Report(analysis.Diagnostic{
			Pos:     file.Pos(),
			Message: fmt.Sprintf("failed to format fixed file: %v", err),
		})

		return
	}

	// Write to file
	err = os.WriteFile(filePath, buf.Bytes(), fileMode)
	if err != nil {
		pass.Report(analysis.Diagnostic{
			Pos:     file.Pos(),
			Message: fmt.Sprintf("failed to write fixed file: %v", err),
		})

		return
	}
}
