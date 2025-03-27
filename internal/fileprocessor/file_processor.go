package fileprocessor

import (
	"go/ast"

	"github.com/manuelarte/funcorder/internal/astutils"

	"github.com/manuelarte/funcorder/internal/errors"

	"github.com/manuelarte/funcorder/internal/models"
)

// FileProcessor Holder to store all the functions that are potential to be constructors and all the structs.
type FileProcessor struct {
	structs map[string]*models.StructHolder
}

// NewFileProcessor creates a new file processor.
func NewFileProcessor() *FileProcessor {
	return &FileProcessor{
		structs: make(map[string]*models.StructHolder),
	}
}

// Process the ast node. It keeps track of the structs and their "constructors" and methods.
func (fp *FileProcessor) Process(n ast.Node) bool {
	switch castedN := n.(type) {
	case *ast.File:
		fp.newFileNode(castedN)
		return true
	case *ast.FuncDecl:
		fp.newFuncDecl(castedN)
		return false
	case *ast.TypeSpec:
		fp.newTypeSpec(castedN)
		return false
	}

	return true
}

// Analyze check whether the order of the methods in the constructor is correct.
func (fp *FileProcessor) Analyze() []errors.LinterError {
	var errs []errors.LinterError
	for _, sh := range fp.structs {
		// filter out structs that are not declared inside that file
		if sh.Struct != nil {
			errs = append(errs, sh.Analyze()...)
		}
	}
	return errs
}

func (fp *FileProcessor) addConstructor(sc models.StructConstructor) {
	structReturn := sc.GetStructReturn()
	sh := fp.getOrCreate(structReturn.Name)
	sh.AddConstructor(sc.GetConstructor())
}

func (fp *FileProcessor) addMethod(st string, n *ast.FuncDecl) {
	sh := fp.getOrCreate(st)
	sh.AddMethod(n)
}

func (fp *FileProcessor) newFileNode(_ *ast.File) {
	fp.structs = make(map[string]*models.StructHolder)
}

func (fp *FileProcessor) newFuncDecl(n *ast.FuncDecl) {
	if sc, isSC := models.NewStructConstructor(n); isSC {
		fp.addConstructor(sc)
	} else if st, isMethod := astutils.FuncIsMethod(n); isMethod {
		fp.addMethod(st.Name, n)
	}
}

func (fp *FileProcessor) newTypeSpec(n *ast.TypeSpec) {
	sh := fp.getOrCreate(n.Name.Name)
	sh.Struct = n
}

func (fp *FileProcessor) getOrCreate(structName string) *models.StructHolder {
	if holder, ok := fp.structs[structName]; ok {
		return holder
	}
	created := &models.StructHolder{}
	fp.structs[structName] = created
	return created
}
