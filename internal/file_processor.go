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
func (fp *FileProcessor) Analyze(pass *analysis.Pass) error {
	for _, sh := range fp.structs {
		// filter out structs that are not declared inside that file
		if sh.Struct != nil {
			err := sh.Analyze(pass)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (fp *FileProcessor) NewFileNode(_ *ast.File) {
	fp.structs = make(map[string]*StructHolder)
}

func (fp *FileProcessor) NewFuncDecl(n *ast.FuncDecl) {
	if sc, ok := NewStructConstructor(n); ok {
		fp.addConstructor(sc)
		return
	}

	if st, ok := FuncIsMethod(n); ok {
		fp.addMethod(st.Name, n)
	}
}

func (fp *FileProcessor) NewTypeSpec(n *ast.TypeSpec) {
	sh := fp.getOrCreate(n.Name.Name)
	sh.Struct = n
}

func (fp *FileProcessor) addConstructor(sc StructConstructor) {
	sh := fp.getOrCreate(sc.GetStructReturn().Name)
	sh.AddConstructor(sc.GetConstructor())
}

func (fp *FileProcessor) addMethod(st string, n *ast.FuncDecl) {
	sh := fp.getOrCreate(st)
	sh.AddMethod(n)
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
