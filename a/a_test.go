package data

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

func TestA(t *testing.T) {
	fs := token.NewFileSet()
	f, err := parser.ParseFile(fs, "a.go", nil, parser.AllErrors|parser.ParseComments)
	if err != nil {
		panic(err)
	}
	for _, decl := range f.Decls {
		ast.Inspect(decl, func(node ast.Node) bool {
			switch n := node.(type) {
			case *ast.CallExpr:
				fmt.Println(n)                                             // prints every func call expression
				fmt.Println("\t", fs.PositionFor(n.Pos(), false).String()) // prints every func call expression
			}
			return true
		})
	}

}
