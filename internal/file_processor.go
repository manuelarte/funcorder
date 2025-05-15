package internal

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

// FileProcessor Holder to store all the functions that are potential to be constructors and all the structs.
type FileProcessor struct {
	structs  map[string]*StructHolder
	features Feature
}

// NewFileProcessor creates a new file processor.
func NewFileProcessor(checkers Feature) *FileProcessor {
	return &FileProcessor{
		structs:  make(map[string]*StructHolder),
		features: checkers,
	}
}

// Analyze check whether the order of the methods in the constructor is correct.
func (fp *FileProcessor) Analyze(pass *analysis.Pass) {
	for _, sh := range fp.structs {
		// filter out structs that are not declared inside that file
		if sh.Struct != nil {
			sh.Analyze(pass)
		}
	}
}

func (fp *FileProcessor) ResetStructs() {
	fp.structs = make(map[string]*StructHolder)
}

func (fp *FileProcessor) AddFuncDecl(n *ast.FuncDecl) {
	if sc := NewStructConstructor(n); sc != nil {
		sh := fp.getOrCreate(sc.StructReturn.Name)
		sh.Constructors = append(sh.Constructors, sc.Constructor)

		return
	}

	if st := funcIsMethod(n); st != nil {
		sh := fp.getOrCreate(st.Name)
		sh.StructMethods = append(sh.StructMethods, n)
	}
}

func (fp *FileProcessor) AddTypeSpec(n *ast.TypeSpec) {
	sh := fp.getOrCreate(n.Name.Name)
	sh.Struct = n
}

func (fp *FileProcessor) getOrCreate(structName string) *StructHolder {
	if holder, ok := fp.structs[structName]; ok {
		return holder
	}

	created := &StructHolder{
		Features: fp.features,
	}
	fp.structs[structName] = created

	return created
}

func funcIsMethod(n *ast.FuncDecl) *ast.Ident {
	if n.Recv == nil {
		return nil
	}

	if len(n.Recv.List) != 1 {
		return nil
	}

	return getIdent(n.Recv.List[0].Type)
}

func getIdent(expr ast.Expr) *ast.Ident {
	switch exp := expr.(type) {
	case *ast.StarExpr:
		return getIdent(exp.X)

	case *ast.Ident:
		return exp

	default:
		return nil
	}
}
