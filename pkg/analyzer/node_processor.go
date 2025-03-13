package analyzer

import (
	"github.com/manuelarte/gofuncor/pkg/models"
	"go/ast"
)

// NodeProcessor Holder to store all the functions that are potential to be constructors and all the structs
type NodeProcessor struct {
	packageName string
	fileName    string
	structs     map[string]*models.StructHolder
}

func NewNodeProcessor() *NodeProcessor {
	return &NodeProcessor{
		structs: make(map[string]*models.StructHolder),
	}
}

func (np *NodeProcessor) Process(n ast.Node) (bool, error) {
	switch castedN := n.(type) {
	case *ast.File:
		np.newFileNode(castedN)
		return true, nil
	case *ast.FuncDecl:
		np.newFuncDecl(castedN)
		return false, nil
	case *ast.TypeSpec:
		np.newTypeSpec(castedN)
	}

	return true, nil
}

func (np *NodeProcessor) addConstructor(sc models.StructConstructor) {
	structReturn, _ := sc.GetStructReturn()
	sh := np.getOrCreate(structReturn.Name)
	sh.AddConstructor(sc.Func)
}

func (np *NodeProcessor) newFileNode(n *ast.File) {
	// clear all the structs, maybe
	np.fileName = n.Name.String()
	//p.packageName = n.Package
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
