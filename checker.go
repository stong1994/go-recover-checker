package main

import (
	"fmt"
	"go.uber.org/multierr"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

type Config struct {
	IgnoreComment   string // ignore no-recover warning, ex: // no-recover-warning
	RecoverMaxDepth int    // recover function should not too far from the method definition, -1 means no limit
}

type Checker struct {
	fset            *token.FileSet
	needRecoverList []*ast.GoStmt
	Config
}

func NewChecker(fset *token.FileSet) *Checker {
	c := &Checker{}
	if fset == nil {
		c.fset = token.NewFileSet()
	} else {
		c.fset = fset
	}
	return c
}

func (c *Checker) GetNeedRecoverList() []*ast.GoStmt {
	return c.needRecoverList
}

func (c *Checker) GetFileSet() *token.FileSet {
	return c.fset
}

func (c *Checker) DisplayNeedRecoverList() {
	fmt.Println("need recover list:")
	for i, fnc := range c.needRecoverList {
		fmt.Printf("\t%d: %s\n", i, c.fset.PositionFor(fnc.Pos(), false))
	}
}

func (c *Checker) ParseFiles(filenameList []string) error {
	var errors []error
	files, err := findFiles(filenameList)
	if err != nil {
		return err
	}
	for _, file := range files {
		if err = c.ParseFile(file, nil); err != nil {
			errors = append(errors, err)
		}
	}
	if err = multierr.Combine(errors...); err != nil {
		return err
	}
	return nil
}

func (c *Checker) ParseFile(filename string, src interface{}) error {
	if strings.Contains(filename, "global\\service") {
		fmt.Println(filename)
	}
	f, err := parser.ParseFile(c.fset, filename, src, parser.AllErrors|parser.ParseComments)
	if err != nil {
		return err
	}
	// 找到go关键词
	for _, decl := range f.Decls {
		fnc, ok := decl.(*ast.FuncDecl)
		if !ok || c.fncHasIgnoreComment(fnc) {
			continue
		}
		c.ParseFunc(fnc)
	}
	return nil
}

func (c *Checker) ParseFunc(fnc *ast.FuncDecl) (containRecover bool) {
	if fnc.Body == nil || len(fnc.Body.List) == 0 { // body is empty, ignore
		return
	}
	for _, stmt := range fnc.Body.List {
		c.handleStmt(stmt)
	}
	return
}

func (c *Checker) parseFunc(fnc *ast.FuncDecl) (containRecover bool) {
	if fnc.Body == nil || len(fnc.Body.List) == 0 { // body is empty, ignore
		return
	}
	for _, stmt := range fnc.Body.List {
		if c.handleStmt(stmt) {
			return true
		}
	}
	return
}

func (c *Checker) handleStmt(stmt ast.Stmt) bool {
	switch stmt.(type) {
	case *ast.GoStmt:
		if c.isNeedRecover(stmt.(*ast.GoStmt)) {
			c.needRecoverList = append(c.needRecoverList, stmt.(*ast.GoStmt))
		}

	case *ast.DeferStmt:
		sf := stmt.(*ast.DeferStmt).Call.Fun
		if c.handleExpr(sf) {
			return true
		}
	case *ast.AssignStmt:
		for _, r := range stmt.(*ast.AssignStmt).Rhs {
			if c.handleExpr(r) {
				return true
			}
		}
	case *ast.ExprStmt:
		switch stmt.(*ast.ExprStmt).X.(type) {
		case *ast.CallExpr:
			c.handleExpr(stmt.(*ast.ExprStmt).X.(*ast.CallExpr).Fun)
		}
	}
	return false
}

func (c *Checker) handleExpr(r ast.Expr) (hasRecover bool) {
	switch r.(type) {
	case *ast.CallExpr:
		if f, ok := r.(*ast.CallExpr).Fun.(*ast.Ident); ok {
			if identIsRecover(f) {
				return true
			}
			if f.Obj != nil {
				if fd, ok := f.Obj.Decl.(*ast.FuncDecl); ok {
					containRecover := c.parseFunc(fd)
					if containRecover {
						return true
					}
				}
			}
		}
	case *ast.Ident:
		if c.handleIdent(r.(*ast.Ident)) {
			return true
		}
	case *ast.FuncLit:
		if c.funcLitHasRecover(r.(*ast.FuncLit)) {
			return true
		}
	}
	return false
}

// go函数中，recover的使用方式可以归类为：
// 1. 直接调用
// go func(){
//     defer recover()
// }
// 2. 在函数中调用（理论上，调用recover的函数可以被多次嵌套）
// go func(){
//     defer func() {
//          _ = recover()
//     }
// }
func (c *Checker) isNeedRecover(stmt *ast.GoStmt) (needRecover bool) {
	fnc := stmt.Call.Fun
	switch fnc.(type) {
	case *ast.Ident:
		if c.handleIdent(fnc.(*ast.Ident)) {
			return false
		}
	case *ast.FuncLit:
		if c.funcLitHasRecover(fnc.(*ast.FuncLit)) {
			return false
		}
	}
	return true
}

func (c *Checker) handleIdent(ident *ast.Ident) (hasRecover bool) {
	if identIsRecover(ident) {
		return true
	}
	if obj := ident.Obj; obj != nil {
		if fd, ok := obj.Decl.(*ast.FuncDecl); ok {
			if c.fncHasIgnoreComment(fd) {
				return true
			}
			if hasRecover := c.fncHasRecover(fd); hasRecover {
				return true
			}
		}
	}
	return false
}

func (c *Checker) fncHasRecover(fnc *ast.FuncDecl) bool {
	for _, stmt := range fnc.Body.List {
		if c.handleStmt(stmt) {
			return true
		}
	}
	return false
}

func (c *Checker) fncHasIgnoreComment(fnc *ast.FuncDecl) bool {
	if fnc.Doc == nil {
		return false
	}
	for _, doc := range fnc.Doc.List {
		if strings.Contains(doc.Text, c.IgnoreComment) {
			return true
		}
	}
	return false
}

func (c *Checker) funcLitHasRecover(fl *ast.FuncLit) bool {
	for _, s := range fl.Body.List {
		if c.handleStmt(s) {
			return true
		}
	}
	return false
}

func identIsRecover(ident *ast.Ident) bool {
	return ident.Name == "recover"
}
