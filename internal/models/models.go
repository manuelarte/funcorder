package models

import "go/ast"

type ExportedMethods []*ast.FuncDecl
type UnexportedMethods []*ast.FuncDecl
