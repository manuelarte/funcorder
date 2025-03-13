package analyzer

import (
	"github.com/manuelarte/gofuncor/pkg/models"
	"go/ast"
	"strings"
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

func (p *NodeProcessor) Process(n ast.Node) (bool, error) {
	switch castedN := n.(type) {
	case *ast.File:
		p.newFileNode(castedN)
		return true, nil
	case *ast.FuncDecl:
		p.newFuncDecl(castedN)
		return false, nil
	}
	return true, nil
}

func (p *NodeProcessor) newFileNode(n *ast.File) {
	// clear all the structs, maybe
	p.fileName = n.Name.String()
	//p.packageName = n.Package
	p.structs = make(map[string]*models.StructHolder)
}

func (p *NodeProcessor) newFuncDecl(n *ast.FuncDecl) {
	if funcNameCanBeConstructor(n) {
		potentialStructField := n.Type.Results.List[0].Type
		returnTypeIsStruct(potentialStructField)
		//if p.structs[potentialStructField.] == nil {}
	}
}

func funcNameCanBeConstructor(n *ast.FuncDecl) bool {
	isExported := n.Name.IsExported()
	startsWithNew := strings.HasPrefix("new", strings.ToLower(n.Name.Name))
	hasAtLeastOneOutput := n.Type.Results != nil && len(n.Type.Results.List) > 0
	return isExported && startsWithNew && hasAtLeastOneOutput
}

func returnTypeIsStruct(expr ast.Expr) bool {
	if pointerExpr, isPointerExpr := expr.(*ast.StarExpr); isPointerExpr {
		return returnTypeIsStruct(pointerExpr.X)
	}
	if structExpr, isStructExpr := expr.(*ast.Ident); isStructExpr {
		println(structExpr.Name)

	}
	return true
}
