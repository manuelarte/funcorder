package fileprocessor

import (
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/analysis"

	"github.com/manuelarte/funcorder/internal/astutils"
	"github.com/manuelarte/funcorder/internal/features"
	"github.com/manuelarte/funcorder/internal/models"
	"github.com/manuelarte/funcorder/internal/structholder"
)

// FileProcessor Holder to store all the functions that are potential to be constructors and all the structs.
type FileProcessor struct {
	fset     *token.FileSet
	structs  map[string]*structholder.StructHolder
	features features.Feature
}

// NewFileProcessor creates a new file processor.
func NewFileProcessor(fset *token.FileSet, checkers features.Feature) *FileProcessor {
	return &FileProcessor{
		fset:     fset,
		structs:  make(map[string]*structholder.StructHolder),
		features: checkers,
	}
}

// Analyze check whether the order of the methods in the constructor is correct.
func (fp *FileProcessor) Analyze() ([]analysis.Diagnostic, error) {
	var reports []analysis.Diagnostic

	for _, sh := range fp.structs {
		// filter out structs that are not declared inside that file
		if sh.Struct != nil {
			newReports, err := sh.Analyze()
			if err != nil {
				return nil, err
			}
			reports = append(reports, newReports...)
		}
	}

	return reports, nil
}

func (fp *FileProcessor) NewFileNode(_ *ast.File) {
	fp.structs = make(map[string]*structholder.StructHolder)
}

func (fp *FileProcessor) NewFuncDecl(n *ast.FuncDecl) {
	if sc, ok := models.NewStructConstructor(n); ok {
		fp.addConstructor(sc)
		return
	}

	if st, ok := astutils.FuncIsMethod(n); ok {
		fp.addMethod(st.Name, n)
	}
}

func (fp *FileProcessor) NewTypeSpec(n *ast.TypeSpec) {
	sh := fp.getOrCreate(n.Name.Name)
	sh.Struct = n
}

func (fp *FileProcessor) addConstructor(sc models.StructConstructor) {
	sh := fp.getOrCreate(sc.GetStructReturn().Name)
	sh.AddConstructor(sc.GetConstructor())
}

func (fp *FileProcessor) addMethod(st string, n *ast.FuncDecl) {
	sh := fp.getOrCreate(st)
	sh.AddMethod(n)
}

func (fp *FileProcessor) getOrCreate(structName string) *structholder.StructHolder {
	if holder, ok := fp.structs[structName]; ok {
		return holder
	}

	created := &structholder.StructHolder{
		Fset:     fp.fset,
		Features: fp.features,
	}
	fp.structs[structName] = created

	return created
}
