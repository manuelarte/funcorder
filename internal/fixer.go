package internal

import (
	"cmp"
	"fmt"
	"go/ast"
	"go/token"
	"slices"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"golang.org/x/tools/go/analysis"
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

	// Use dst to reorder declarations while preserving comments
	writeFixedFileWithDst(pass, file, decls, orderedDecls)
}

// declaration represents a top-level declaration with its associated comments.
type declaration struct {
	node               ast.Decl
	docComments        []*ast.CommentGroup // DOC comments (attached via Doc field)
	standaloneComments []*ast.CommentGroup // Standalone comments (in file.Comments)
	pos                token.Pos
}

// collectDeclarations collects all top-level declarations from the file.
func collectDeclarations(file *ast.File, _ *FileProcessor) []declaration {
	decls := make([]declaration, 0, len(file.Decls))

	// Separate DOC comments (attached via Doc field) from standalone comments (in file.Comments)
	for i, d := range file.Decls {
		// Get DOC comments (attached to nodes via Doc field)
		docComments := getDocComments(d)

		// Get standalone comments (in file.Comments, not attached to nodes)
		prevEnd := getPreviousEnd(file, i)
		nextStart, hasNext := getNextStart(file, i)
		standaloneComments := collectStandaloneComments(file.Comments, d, prevEnd, nextStart, hasNext)

		// If node.Doc is nil but there are comments immediately before this declaration,
		// treat them as DOC comments (they should be attached to the node)
		if len(docComments) == 0 {
			// Find comments that are immediately before this declaration (DOC comments)
			for _, cg := range file.Comments {
				// Skip if this is already a standalone comment
				if isCommentInList(cg, standaloneComments) {
					continue
				}

				// If comment is between prevEnd and this declaration, it's a DOC comment
				if cg.Pos() > prevEnd && cg.End() <= d.Pos() {
					docComments = append(docComments, cg)
				}
			}
		}

		decls = append(decls, declaration{
			node:               d,
			docComments:        docComments,
			standaloneComments: standaloneComments,
			pos:                d.Pos(),
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

// isCommentInList checks if a comment group is in the given list.
func isCommentInList(cg *ast.CommentGroup, list []*ast.CommentGroup) bool {
	for _, sc := range list {
		if sc == cg {
			return true
		}
	}

	return false
}

// getPreviousEnd gets the end position of the previous declaration.
func getPreviousEnd(file *ast.File, i int) token.Pos {
	if i > 0 {
		return file.Decls[i-1].End()
	}

	// For the first declaration, return the package position itself
	// This allows us to capture comments that appear right after the package line
	return file.Package
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

// removeStandaloneFromStart removes standalone comment strings from Start decorations.
func removeStandaloneFromStart(startDecs, standaloneStrs []string) []string {
	// Build a set of standalone strings for fast lookup
	standaloneSet := make(map[string]bool)

	for _, s := range standaloneStrs {
		standaloneSet[s] = true
	}

	// Filter out standalone comments from Start decorations
	var result []string

	for _, dec := range startDecs {
		if !standaloneSet[dec] {
			result = append(result, dec)
		}
	}

	return result
}

// buildDeclMap builds a map from ast.Decl to dst.Decl for reordering.
func buildDeclMap(file *ast.File, dstFile *dst.File) map[ast.Decl]dst.Decl {
	declMap := make(map[ast.Decl]dst.Decl)

	for i, astDecl := range file.Decls {
		if i < len(dstFile.Decls) {
			declMap[astDecl] = dstFile.Decls[i]
		}
	}

	return declMap
}

// buildStandaloneCommentsMap builds a map of standalone comments from declarations.
func buildStandaloneCommentsMap(orderedDecls []declaration) map[ast.Decl][]string {
	declToStandalone := make(map[ast.Decl][]string)

	for _, decl := range orderedDecls {
		if len(decl.standaloneComments) > 0 {
			// Convert ast comment groups to decoration strings
			var standaloneStrs []string

			for _, cg := range decl.standaloneComments {
				for _, comment := range cg.List {
					standaloneStrs = append(standaloneStrs, comment.Text)
				}

				standaloneStrs = append(standaloneStrs, "\n") // Blank line after standalone comment
			}

			declToStandalone[decl.node] = standaloneStrs
		}
	}

	return declToStandalone
}

// buildOriginalPosMap builds a map of original positions from declarations.
func buildOriginalPosMap(originalDecls []declaration) map[ast.Decl]int {
	originalPosMap := make(map[ast.Decl]int)

	for origIdx, decl := range originalDecls {
		originalPosMap[decl.node] = origIdx
	}

	return originalPosMap
}

// collectFirstDeclStandaloneComments collects standalone comments from declarations
// that moved from position 0 to a later position.
func collectFirstDeclStandaloneComments(
	orderedDecls []declaration,
	originalPosMap map[ast.Decl]int,
	declToStandalone map[ast.Decl][]string,
) []string {
	var firstDeclStandaloneComments []string

	for newIdx, orderedDecl := range orderedDecls {
		origIdx := originalPosMap[orderedDecl.node]
		// If this declaration moved from position 0 to a later position,
		// its standalone comments should go to the new first declaration
		if origIdx == 0 && newIdx > 0 {
			if standaloneStrs, hasStandalone := declToStandalone[orderedDecl.node]; hasStandalone {
				firstDeclStandaloneComments = append(firstDeclStandaloneComments, standaloneStrs...)
			}
		}
	}

	return firstDeclStandaloneComments
}

// applyStandaloneCommentsToDecl applies standalone comments to a declaration based on position changes.
func applyStandaloneCommentsToDecl(
	dstDecl dst.Decl,
	standaloneStrs []string,
	origIdx, newIdx int,
) {
	// Remove standalone comments from this declaration's Start
	switch d := dstDecl.(type) {
	case *dst.FuncDecl:
		d.Decs.Start = removeStandaloneFromStart(d.Decs.Start, standaloneStrs)
	case *dst.GenDecl:
		d.Decs.Start = removeStandaloneFromStart(d.Decs.Start, standaloneStrs)
	}

	// Handle standalone comments based on position changes
	// If this declaration moved from position 0 to a later position,
	// its standalone comments were already collected in the pre-pass
	switch {
	case origIdx == 0 && newIdx > 0:
		// Don't re-add, they're already in firstDeclStandaloneComments
	case origIdx > 0 && newIdx == 0, origIdx > 0 && newIdx > 0:
		// This declaration moved TO position 0 from a later position,
		// or didn't cross the first position boundary
		// Keep its standalone comments (don't remove them)
		switch d := dstDecl.(type) {
		case *dst.FuncDecl:
			d.Decs.Start = append(standaloneStrs, d.Decs.Start...)
		case *dst.GenDecl:
			d.Decs.Start = append(standaloneStrs, d.Decs.Start...)
		}
	}
}

// reorderDstDeclarations reorders dst declarations based on the ordered list.
func reorderDstDeclarations(
	orderedDecls []declaration,
	declMap map[ast.Decl]dst.Decl,
	originalPosMap map[ast.Decl]int,
	declToStandalone map[ast.Decl][]string,
	firstDeclStandaloneComments []string,
) []dst.Decl {
	newDstDecls := make([]dst.Decl, 0, len(orderedDecls))

	for newIdx, orderedDecl := range orderedDecls {
		if dstDecl, ok := declMap[orderedDecl.node]; ok {
			origIdx := originalPosMap[orderedDecl.node]

			// Check if this declaration has standalone comments
			if standaloneStrs, hasStandalone := declToStandalone[orderedDecl.node]; hasStandalone {
				applyStandaloneCommentsToDecl(dstDecl, standaloneStrs, origIdx, newIdx)
			}

			// Add standalone comments from declarations that moved away from position 0
			if newIdx == 0 && len(firstDeclStandaloneComments) > 0 {
				switch d := dstDecl.(type) {
				case *dst.FuncDecl:
					d.Decs.Start = append(firstDeclStandaloneComments, d.Decs.Start...)
				case *dst.GenDecl:
					d.Decs.Start = append(firstDeclStandaloneComments, d.Decs.Start...)
				}

				firstDeclStandaloneComments = nil // Clear so we don't add again
			}

			newDstDecls = append(newDstDecls, dstDecl)
		}
	}

	return newDstDecls
}

// writeFixedFileWithDst writes the fixed file using dst to preserve comments.
func writeFixedFileWithDst(
	pass *analysis.Pass,
	file *ast.File,
	originalDecls []declaration,
	orderedDecls []declaration,
) {
	// Get the file path
	fset := pass.Fset
	filePath := fset.File(file.Pos()).Name()

	if filePath == "" {
		return
	}

	// Convert ast.File to dst.File with decorator
	dec := decorator.NewDecorator(fset)

	dstFile, err := dec.DecorateFile(file)
	if err != nil {
		pass.Report(analysis.Diagnostic{
			Pos:     file.Pos(),
			Message: fmt.Sprintf("failed to decorate file: %v", err),
		})

		return
	}

	// Build maps and collect standalone comments
	declMap := buildDeclMap(file, dstFile)
	declToStandalone := buildStandaloneCommentsMap(orderedDecls)
	originalPosMap := buildOriginalPosMap(originalDecls)
	firstDeclStandaloneComments := collectFirstDeclStandaloneComments(orderedDecls, originalPosMap, declToStandalone)

	// Reorder dst declarations based on our ordered list
	newDstDecls := reorderDstDeclarations(
		orderedDecls,
		declMap,
		originalPosMap,
		declToStandalone,
		firstDeclStandaloneComments,
	)

	// Update the dst file with reordered declarations
	dstFile.Decls = newDstDecls

	// Format and write the file
	writeErr := writeFormattedFile(pass, file, dstFile, filePath)
	if writeErr != nil {
		return
	}
}
