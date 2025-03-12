package analyzer

import "go/ast"

type Processor struct {
	packageName string
	fileName    string
	structs     map[string]nodeDataHolder
}

type nodeDataHolder struct {
	Pos int
}

func NewNodeProcessor() *Processor {
	return &Processor{
		structs: make(map[string]nodeDataHolder),
	}
}

func (p *Processor) Process(n ast.Node) error {
	// think of a switch
	if fileNode, ok := n.(*ast.File); ok {
		p.newFileNode(fileNode)
		return nil
	}
	return nil
}

func (p *Processor) newFileNode(n *ast.File) {
	// clear all the structs, maybe
	p.fileName = n.Name.String()
	//p.packageName = n.Package
	p.structs = make(map[string]nodeDataHolder)
}
