package main

import (
	"fmt"
	"github.com/stong1994/go-recover-checker/entity"
	"go.uber.org/multierr"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Config struct {
	IgnoreComment   string // ignore no-recover warning, ex: // no-recover-warning
	RecoverMaxDepth int    // recover function should not too far from the method definition, -1 means no limit
}

func main() {
	go hello()
	fset := token.NewFileSet()
	files, err := findFiles([]string{"./"})
	if err != nil {
		return
	}
	fm := entity.NewFuncSet()
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
			if fnc.Body == nil || len(fnc.Body.List) == 0 { // body is empty, ignore
				continue
			}
			for _, stmt := range fnc.Body.List {
				switch stmt.(type) {
				case *ast.GoStmt:
					goStmt := stmt.(*ast.GoStmt)
					fmt.Println("hit go func", fnc.Name.String())
					ident := goStmt.Call.Fun.(*ast.Ident)
					fm.AddFunc(ident)
				case *ast.AssignStmt:
					aStmt := stmt.(*ast.AssignStmt)
					if len(aStmt.Rhs) == 0 {
						continue
					}
					for _, expr := range aStmt.Rhs {
						switch expr.(type) {
						case *ast.CallExpr:
							c := expr.(*ast.CallExpr)
							switch c.Fun.(type) {
							case *ast.SelectorExpr: // fset := token.NewFileSet()
								s := c.Fun.(*ast.SelectorExpr)
								fm.AddFunc(s.Sel) // 递归or迭代
							case *ast.Ident:
								fm.AddFunc(c.Fun.(*ast.Ident)) // files, err := findFiles([]string{"./"})
							}
						}
					}
				case *ast.IfStmt:
					ifStmt := stmt.(*ast.IfStmt)
					if init := ifStmt.Init; init != nil {
						a := init.(*ast.AssignStmt)
						for _, expr := range a.Rhs {
							switch expr.(type) {
							case *ast.CallExpr:
								cExpr := expr.(*ast.CallExpr)
								if cExpr.Fun != nil {
									switch cExpr.Fun.(type) {
									case *ast.SelectorExpr:
										fm.AddFunc(cExpr.Fun.(*ast.SelectorExpr).Sel)
									}
								}
							}
						}
					}
					// todo handle body and cond
				}
			}
		}
	}

	if err = multierr.Combine(errors...); err != nil {
		fmt.Print(err)
		return
	}
	for _, fnc := range fm.FindNeedRecoveredFunc() {
		fmt.Println("need recover", fnc)
	}
}

func findFiles(patterns []string) (_ []string, err error) {
	files := make(map[string]struct{})
	for _, pat := range patterns {
		fs, findErr := findGoFiles(pat)
		if findErr != nil {
			err = multierr.Append(err, fmt.Errorf("enumerating Go files in %q: %v", pat, err))
			continue
		}

		for _, f := range fs {
			files[f] = struct{}{}
		}
	}

	sortedFiles := make([]string, 0, len(files))
	for f := range files {
		sortedFiles = append(sortedFiles, f)
	}
	sort.Strings(sortedFiles)

	return sortedFiles, err
}

func findGoFiles(path string) ([]string, error) {
	path, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	var files []string
	err = filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		mode := info.Mode()
		switch {
		case mode.IsRegular():
			if strings.HasSuffix(path, ".go") {
				files = append(files, path)
			}

		case mode.IsDir():
			base := filepath.Base(path)
			switch {
			case len(base) == 0,
				base[0] == '.',
				base[0] == '_',
				base == "vendor":
				return filepath.SkipDir
			}
		}

		return nil
	})

	return files, err
}

// hello
// This is a function
func hello() {}
