package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

func TestParser_handleStmt(t *testing.T) {
	contentDirectRecover := `
package data
import "fmt"
func helloNotRecover() {
	go recover()
	fmt.Println("hello")
}
`
	contentRecoverNoRecv := `
package data
import "fmt"
func helloNotRecover() {
	go func() {
		recover()
	}()
	fmt.Println("hello")
}
`
	contentRecoverIdent := `
package data
import "fmt"
func helloNotRecover() {
	go recoverWorld()
	fmt.Println("hello")
}
func recoverWorld() {
	defer recover()
	fmt.Println("world")
}
`
	contentRecoverStmt := `
package data
import "fmt"
func helloNotRecover() {
	go recoverWorld()
	fmt.Println("hello")
}
func recoverWorld() {
	defer func() {
		_ = recover()
	}()
	fmt.Println("world")
}
`

	contentNoRecover := `
package data
import "fmt"
func helloNotRecover() {
	go recoverWorld()
	fmt.Println("hello")
}
func recoverWorld() {
	defer func() {
	}()
	fmt.Println("world")
}
`
	contentNoDeferWhenRecover := `
package data
import "fmt"
func helloNotRecover() {
	go recoverWorld()
	fmt.Println("hello")
}
func recoverWorld() {
	recover()
	fmt.Println("world")
}
`

	contentRecoverIn2LayerFunc := `
package data
import "fmt"
func helloNotRecover() {
	go recoverWorld()
	fmt.Println("hello")
}
func recoverWorld() {
	defer func() {
		func() {
			recover()
		}()
	}()
	fmt.Println("world")
}`

	contentRecoverIn3LayerFunc := `
package data
import "fmt"
func helloNotRecover() {
	go recoverWorld()
	fmt.Println("hello")
}
func recoverWorld() {
	defer func() {
		func() {
			func() {
				recover()
			}()
		}()
	}()
	fmt.Println("world")
}`

	contentIgnoreComment := `
package data
import "fmt"
func helloNotRecover() {
	go recoverWorld()
	fmt.Println("hello")
}

// no-recover-warning
func recoverWorld() {
	fmt.Println("world")
}`

	tests := []struct {
		name        string
		fileContent string
		check       func(p *Checker, content string) bool
	}{
		{
			name:        "direct recover",
			fileContent: contentDirectRecover,
			check: func(p *Checker, content string) bool {
				helloNotRecover := getTestFile(content).Decls[1].(*ast.FuncDecl)
				needRecover := p.isNeedRecover(helloNotRecover.Body.List[0].(*ast.GoStmt))
				return needRecover == false
			},
		}, {
			name:        "recover no receiver",
			fileContent: contentRecoverNoRecv,
			check: func(p *Checker, content string) bool {
				helloNotRecover := getTestFile(content).Decls[1].(*ast.FuncDecl)
				needRecover := p.isNeedRecover(helloNotRecover.Body.List[0].(*ast.GoStmt))
				return needRecover == false
			},
		}, {
			name:        "recover func",
			fileContent: contentRecoverIdent,
			check: func(p *Checker, content string) bool {
				helloNotRecover := getTestFile(content).Decls[1].(*ast.FuncDecl)
				needRecover := p.isNeedRecover(helloNotRecover.Body.List[0].(*ast.GoStmt))
				return needRecover == false
			},
		}, {
			name:        "recover stmt",
			fileContent: contentRecoverStmt,
			check: func(p *Checker, content string) bool {
				helloNotRecover := getTestFile(content).Decls[1].(*ast.FuncDecl)
				needRecover := p.isNeedRecover(helloNotRecover.Body.List[0].(*ast.GoStmt))
				return needRecover == false
			},
		}, {
			name:        "no recover",
			fileContent: contentNoRecover,
			check: func(p *Checker, content string) bool {
				helloNotRecover := getTestFile(content).Decls[1].(*ast.FuncDecl)
				needRecover := p.isNeedRecover(helloNotRecover.Body.List[0].(*ast.GoStmt))
				return needRecover == true
			},
		}, {
			name:        "no recover and defer",
			fileContent: contentNoDeferWhenRecover,
			check: func(p *Checker, content string) bool {
				helloNotRecover := getTestFile(content).Decls[1].(*ast.FuncDecl)
				needRecover := p.isNeedRecover(helloNotRecover.Body.List[0].(*ast.GoStmt))
				return needRecover == true
			},
		}, {
			name:        "recover in 2 layer func",
			fileContent: contentRecoverIn2LayerFunc,
			check: func(p *Checker, content string) bool {
				helloNotRecover := getTestFile(content).Decls[1].(*ast.FuncDecl)
				needRecover := p.isNeedRecover(helloNotRecover.Body.List[0].(*ast.GoStmt))
				return needRecover == true
			},
		}, {
			name:        "recover in 3 layer func",
			fileContent: contentRecoverIn3LayerFunc,
			check: func(p *Checker, content string) bool {
				helloNotRecover := getTestFile(content).Decls[1].(*ast.FuncDecl)
				needRecover := p.isNeedRecover(helloNotRecover.Body.List[0].(*ast.GoStmt))
				return needRecover == true
			},
		}, {
			name:        "ignore recover comment",
			fileContent: contentIgnoreComment,
			check: func(p *Checker, content string) bool {
				helloNotRecover := getTestFile(content).Decls[1].(*ast.FuncDecl)
				needRecover := p.isNeedRecover(helloNotRecover.Body.List[0].(*ast.GoStmt))
				return needRecover == false
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.check(NewChecker(), tt.fileContent) {
				t.Errorf("not checked")
			}
		})
	}

}

func getTestFile(content string) *ast.File {
	fs := token.NewFileSet()
	file, err := parser.ParseFile(fs, "go-test.go", content, parser.ParseComments|parser.Trace|parser.AllErrors)
	if err != nil {
		panic(err)
	}
	return file
}
