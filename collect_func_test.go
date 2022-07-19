package main

import (
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"testing"
)

func TestCollectFunc(t *testing.T) {
	//CollectFunc([]string{"./a/a.go"})
	fst := token.NewFileSet()
	f, err := parser.ParseFile(fst, "./a/a.go", nil, parser.AllErrors)
	if err != nil {
		panic(err)
	}
	conf := types.Config{
		Importer: importer.Default(),
	}
	pkg, err := conf.Check("./a/a.go", fst, []*ast.File{f}, nil)
	if err != nil {
		panic(err)
	}
	_ = pkg
}
