package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

var (
	contentDirectRecover = `
package data
import "fmt"
func helloNotRecover() {
	go recover()
	fmt.Println("hello")
}
`
	contentRecoverNoRecv = `
package data
import "fmt"
func helloNotRecover() {
	go func() {
		recover()
	}()
	fmt.Println("hello")
}
`
	contentRecoverIdent = `
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
	contentRecoverStmt = `
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

	contentNoRecover = `
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
	contentNoDeferWhenRecover = `
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

	contentRecoverIn2LayerFunc = `
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

	contentRecoverIn3LayerFunc = `
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

	contentIgnoreComment = `
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

	contentCallThirdPartyRecover = `
package data
import "fmt"
func helloNotRecover() {
	go func(){
		_ = Do(func(){
			recoverWorld()
		})
	}()
	fmt.Println("hello")
}

func recoverWorld() {
	fmt.Println("world")
}
func Do(f func()) error {
	var err error
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()
	f()
	return err
}
`
	contentRecoverInFor = `
package data
import "fmt"
func helloNotRecover() {
	go func(){
		for i := 0; i < 10; i++ {
			_ = Do(func(){
				recoverWorld()
			})
		}
	}()
	fmt.Println("hello")
}

func recoverWorld() {
	fmt.Println("world")
}
func Do(f func()) error {
	var err error
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()
	f()
	return err
}
`

	contentRecoverInForMul = `
package data
import "fmt"

type Service struct {}

func (s Service) hi() {
	go s.ExecTask()
}

func (s Service) ExecTask() {
	for {
		if err := Do(func() {
			recoverWorld()
		}); err != nil {}
	}
}

func recoverWorld() {
	fmt.Println("world")
}
func Do(f func()) error {
	var err error
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()
	f()
	return err
}
`
)

func TestParser_handleStmt(t *testing.T) {
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
				return needRecover == true
			},
		}, {
			name:        "recover no receiver",
			fileContent: contentRecoverNoRecv,
			check: func(p *Checker, content string) bool {
				helloNotRecover := getTestFile(content).Decls[1].(*ast.FuncDecl)
				needRecover := p.isNeedRecover(helloNotRecover.Body.List[0].(*ast.GoStmt))
				return needRecover == true
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
				return needRecover == false
			},
		}, {
			name:        "recover in 3 layer func",
			fileContent: contentRecoverIn3LayerFunc,
			check: func(p *Checker, content string) bool {
				helloNotRecover := getTestFile(content).Decls[1].(*ast.FuncDecl)
				needRecover := p.isNeedRecover(helloNotRecover.Body.List[0].(*ast.GoStmt))
				return needRecover == false
			},
		}, {
			name:        "ignore recover comment",
			fileContent: contentIgnoreComment,
			check: func(p *Checker, content string) bool {
				helloNotRecover := getTestFile(content).Decls[1].(*ast.FuncDecl)
				needRecover := p.isNeedRecover(helloNotRecover.Body.List[0].(*ast.GoStmt))
				return needRecover == false
			},
		}, {
			name:        "call third party recover",
			fileContent: contentCallThirdPartyRecover,
			check: func(p *Checker, content string) bool {
				helloNotRecover := getTestFile(content).Decls[1].(*ast.FuncDecl)
				needRecover := p.isNeedRecover(helloNotRecover.Body.List[0].(*ast.GoStmt))
				return needRecover == false
			},
		}, {
			name:        "call recover in for stmt",
			fileContent: contentRecoverInFor,
			check: func(p *Checker, content string) bool {
				helloNotRecover := getTestFile(content).Decls[1].(*ast.FuncDecl)
				needRecover := p.isNeedRecover(helloNotRecover.Body.List[0].(*ast.GoStmt))
				return needRecover == false
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.check(NewChecker(nil), tt.fileContent) {
				t.Errorf("not checked:%s", tt.name)
			}
		})
	}

}

func getTestFile(content string) *ast.File {
	fs := token.NewFileSet()
	file, err := parser.ParseFile(fs, "go-test.go", content, parser.ParseComments|parser.AllErrors)
	if err != nil {
		panic(err)
	}
	return file
}

func TestParser_handleMultiFile(t *testing.T) {
	tests := []struct {
		name         string
		fileContents []string
		check        func(p *Checker, contents ...string) bool
	}{
		{
			name:         "contentRecoverInForMul1,2",
			fileContents: []string{contentRecoverInForMul},
			check: func(p *Checker, contents ...string) bool {
				needRecover := false
				handleTestFiles(func(fs *token.FileSet, f *ast.File) {
					if len(f.Decls) != 1 {
						return
					}
					goStmt := f.Decls[0].(*ast.FuncDecl).Body.List[2].(*ast.GoStmt)
					_ = goStmt
				}, contents...)
				return needRecover == false
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.check(NewChecker(nil), tt.fileContents...) {
				t.Errorf("not checked:%s", tt.name)
			}
		})
	}
}

func handleTestFiles(fn func(fs *token.FileSet, file *ast.File), contents ...string) {
	fs := token.NewFileSet()
	var files []*ast.File
	for i, content := range contents {
		file, err := parser.ParseFile(fs, fmt.Sprintf("go-test_%d.go", i), content, parser.ParseComments|parser.AllErrors)
		if err != nil {
			panic(err)
		}
		files = append(files, file)
	}
	for _, file := range files {
		fn(fs, file)
	}
}

func TestParser_ParseFile(t *testing.T) {
	tests := []struct {
		name        string
		fileContent string
		check       func(p *Checker, content string) bool
	}{
		{
			name:        "call third party recover func",
			fileContent: contentCallThirdPartyRecover,
			check: func(p *Checker, content string) bool {
				err := p.ParseFile("", content)
				assert.NoError(t, err)
				return err == nil && len(p.GetNeedRecoverList()) == 0
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.check(NewChecker(nil), tt.fileContent) {
				t.Errorf("not checked")
			}
		})
	}
}
