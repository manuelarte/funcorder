package analyzer

import (
	"go/ast"

	"github.com/manuelarte/gofuncor/pkg/errors"
	"github.com/manuelarte/gofuncor/pkg/models"
)

// NodeProcessor Holder to store all the functions that are potential to be constructors and all the structs.
type NodeProcessor struct {
	fileName string
	structs  map[string]*models.StructHolder
}

func NewNodeProcessor() *NodeProcessor {
	return &NodeProcessor{
		structs: make(map[string]*models.StructHolder),
	}
}

func (np *NodeProcessor) Process(n ast.Node) bool {
	switch castedN := n.(type) {
	case *ast.File:
		np.newFileNode(castedN)
		return true
	case *ast.FuncDecl:
		np.newFuncDecl(castedN)
		return false
	case *ast.TypeSpec:
		np.newTypeSpec(castedN)
	}

	return true
}

func (np *NodeProcessor) addConstructor(sc models.StructConstructor) {
	structReturn, _ := sc.GetStructReturn()
	sh := np.getOrCreate(structReturn.Name)
	sh.AddConstructor(sc.Func)
}

func (np *NodeProcessor) newFileNode(n *ast.File) {
	np.fileName = n.Name.String()
	np.structs = make(map[string]*models.StructHolder)
}

func (np *NodeProcessor) newFuncDecl(n *ast.FuncDecl) {
	if sc, isSC := models.NewStructConstructor(n); isSC {
		np.addConstructor(sc)
	}
}

func (np *NodeProcessor) newTypeSpec(n *ast.TypeSpec) {
	sh := np.getOrCreate(n.Name.Name)
	sh.Struct = n
}

func (np *NodeProcessor) getOrCreate(structName string) *models.StructHolder {
	if holder, ok := np.structs[structName]; ok {
		return holder
	}
	created := &models.StructHolder{}
	np.structs[structName] = created
	return created
}

func (np *NodeProcessor) Analyze() []errors.LinterError {
	var errs []errors.LinterError
	for _, sh := range np.structs {
		errs = append(errs, sh.Analyze()...)
	}
	return errs
}
