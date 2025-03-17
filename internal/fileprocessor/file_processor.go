package fileprocessor

import (
	"go/ast"

	"github.com/manuelarte/gofuncor/internal/errors"

	"github.com/manuelarte/gofuncor/internal/models"
)

// FileProcessor Holder to store all the functions that are potential to be constructors and all the structs.
type FileProcessor struct {
	fileName string
	structs  map[string]*models.StructHolder
}

func NewFileProcessor() *FileProcessor {
	return &FileProcessor{
		structs: make(map[string]*models.StructHolder),
	}
}

// Process process the ast node. It keeps track of the structs and their "constructors" and methods.
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
	structReturn, _ := sc.GetStructReturn()
	sh := fp.getOrCreate(structReturn.Name)
	sh.AddConstructor(sc.GetConstructor())
}

func (fp *FileProcessor) newFileNode(n *ast.File) {
	fp.fileName = n.Name.String()
	fp.structs = make(map[string]*models.StructHolder)
}

func (fp *FileProcessor) newFuncDecl(n *ast.FuncDecl) {
	if sc, isSC := models.NewStructConstructor(n); isSC {
		fp.addConstructor(sc)
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
