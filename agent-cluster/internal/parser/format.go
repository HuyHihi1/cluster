package parser

import (
	"go/ast"
	"go/format"
	"go/token"
	"io"
)

func formatNodeImpl(w io.Writer, fset *token.FileSet, n ast.Node) error {
	return format.Node(w, fset, n)
}

