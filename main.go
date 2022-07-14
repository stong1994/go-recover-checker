package main

import (
	"fmt"
	"go.uber.org/multierr"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
)

func main() {
	fset := token.NewFileSet()
	files, err := findFiles([]string{"./data/"})
	if err != nil {
		return
	}
	p := NewChecker()
	var errors []error
	for _, filename := range files {
		f, err := parser.ParseFile(fset, filename, nil, parser.AllErrors|parser.ParseComments|parser.Trace)
		if err != nil {
			log.Printf("%s: failed: %v", filename, err)
			errors = append(errors, fmt.Errorf("could not parse %q: %v", filename, err))
			continue
		}
		// 找到go关键词
		for _, decl := range f.Decls {
			fnc, ok := decl.(*ast.FuncDecl)
			if !ok {
				continue
			}
			p.parseFunc(fnc)
		}
	}

	if err = multierr.Combine(errors...); err != nil {
		fmt.Print(err)
		return
	}
	fmt.Println("need recover list")
	for _, fnc := range p.needRecoverList {
		fmt.Println("\t", fset.PositionFor(fnc.Pos(), false))
	}
}
