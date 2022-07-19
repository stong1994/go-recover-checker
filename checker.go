package main

import (
	"fmt"
	"go.uber.org/multierr"
	"go/ast"
	"go/parser"
	"go/token"
	"sort"
	"strings"
)

type Config struct {
	IgnoreComment   string // ignore no-recover warning, ex: // no-recover-warning
	RecoverMaxDepth int    // recover function should not too far from the method definition, -1 means no limit
}

type Checker struct {
	fset            *token.FileSet
	needRecoverList []*ast.GoStmt
	collectedFunc   *MethodMap
	Config
}

func NewChecker(fset *token.FileSet) *Checker {
	c := &Checker{}
	if fset == nil {
		c.fset = token.NewFileSet()
	} else {
		c.fset = fset
	}
	c.collectedFunc = NewMethodMap()
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
	sort.Slice(c.needRecoverList, func(i, j int) bool {
		return c.needRecoverList[i].Pos() < c.needRecoverList[j].Pos()
	})
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

	var astFile []*ast.File
	for _, file := range files {
		if f, err := c.ParseFile(file, nil); err != nil {
			errors = append(errors, err)
		} else {
			astFile = append(astFile, f)
		}
	}
	for _, f := range astFile {
		for _, decl := range f.Decls {
			fnc, ok := decl.(*ast.FuncDecl)
			if !ok || c.fncHasIgnoreComment(fnc) {
				continue
			}
			c.ParseFunc(fnc)
		}
	}
	if err = multierr.Combine(errors...); err != nil {
		return err
	}
	return nil
}

func (c *Checker) ParseFile(filename string, src interface{}) (*ast.File, error) {
	f, err := parser.ParseFile(c.fset, filename, src, parser.AllErrors|parser.ParseComments)
	if err != nil {
		return nil, err
	}
	if strings.Contains(filename, "timer.go") {
		fmt.Println(filename)
	}
	c.collectedFunc.CollectMethodInFile(f)

	return f, nil
}

func (c *Checker) ParseFunc(fnc *ast.FuncDecl) {
	if fnc.Body == nil || len(fnc.Body.List) == 0 { // body is empty, ignore
		return
	}
	for _, stmt := range fnc.Body.List {
		c.handleStmt(stmt)
	}
	return
}

type rst struct {
	hasRecover, hasDeferRecover bool
	noNeedRecover               bool
}

func (r rst) needRecover() bool {
	return !(r.noNeedRecover || r.hasDeferRecover)
}

var statelessRst = rst{}

func (c *Checker) parseFunc(fnc *ast.FuncDecl) (r rst) {
	if fnc.Body == nil || len(fnc.Body.List) == 0 { // body is empty, ignore
		return
	}
	if c.fncHasIgnoreComment(fnc) {
		r.noNeedRecover = true
		return
	}
	for _, stmt := range fnc.Body.List {
		c.r2r(&r, c.handleStmt(stmt))
	}
	return
}

func (c *Checker) r2r(dst *rst, src rst) {
	if src.noNeedRecover {
		dst.noNeedRecover = true
	}
	if src.hasDeferRecover {
		dst.hasDeferRecover = true
	}
	if src.hasRecover {
		dst.hasRecover = true
	}
}

// 如果是*ast.GoStmt，则进入一个新的判断周期，这个周期中必须包含*ast.DeferStmt和recover()，
// 同时dfs过程中，如果碰到其他*ast.GoStmt，则开启另一个周期，其结果不应影响之前的*ast.GoStmt
func (c *Checker) handleStmt(stmt ast.Stmt) (r rst) {
	switch stmt.(type) {
	case *ast.GoStmt:
		if c.isNeedRecover(stmt.(*ast.GoStmt)) {
			c.needRecoverList = append(c.needRecoverList, stmt.(*ast.GoStmt))
		}
		return statelessRst
	case *ast.DeferStmt:
		r2 := c.handleExpr(stmt.(*ast.DeferStmt).Call.Fun)
		if r2.hasRecover {
			r.hasDeferRecover = true
		}
		r.hasRecover = r2.hasRecover
		r.noNeedRecover = r2.noNeedRecover
	case *ast.DeclStmt:
	// stmt.(*ast.DeclStmt).Decl
	case *ast.LabeledStmt:
		c.r2r(&r, c.handleStmt(stmt.(*ast.LabeledStmt).Stmt))
		c.r2r(&r, c.handleIdent(stmt.(*ast.LabeledStmt).Label))
	case *ast.AssignStmt:
		for _, rh := range stmt.(*ast.AssignStmt).Rhs {
			c.r2r(&r, c.handleExpr(rh))
		}
	case *ast.ExprStmt:
		c.r2r(&r, c.handleExpr(stmt.(*ast.ExprStmt).X))
	case *ast.IfStmt:
		c.r2r(&r, c.handleStmt(stmt.(*ast.IfStmt).Init))
		c.r2r(&r, c.handleExpr(stmt.(*ast.IfStmt).Cond))
		c.r2r(&r, c.handleStmt(stmt.(*ast.IfStmt).Else))
		c.r2r(&r, c.handleStmt(stmt.(*ast.IfStmt).Body))
	case *ast.BlockStmt:
		for _, v := range stmt.(*ast.BlockStmt).List {
			c.r2r(&r, c.handleStmt(v))
		}
	case *ast.ForStmt:
		c.r2r(&r, c.handleStmt(stmt.(*ast.ForStmt).Init))
		c.r2r(&r, c.handleExpr(stmt.(*ast.ForStmt).Cond))
		c.r2r(&r, c.handleStmt(stmt.(*ast.ForStmt).Post))
		c.r2r(&r, c.handleStmt(stmt.(*ast.ForStmt).Body))
	case *ast.ReturnStmt:
		for _, expr := range stmt.(*ast.ReturnStmt).Results {
			c.r2r(&r, c.handleExpr(expr))
		}

	}
	return
}

func (c *Checker) handleExpr(r ast.Expr) (rs rst) {
	switch r.(type) {
	case *ast.CallExpr:
		return c.handleExpr(r.(*ast.CallExpr).Fun)
	case *ast.Ident:
		c.r2r(&rs, c.handleIdent(r.(*ast.Ident)))
	case *ast.FuncLit:
		for _, s := range r.(*ast.FuncLit).Body.List {
			c.r2r(&rs, c.handleStmt(s))
		}
	case *ast.SelectorExpr:
		c.r2r(&rs, c.handleExpr(r.(*ast.SelectorExpr).X))
		c.r2r(&rs, c.handleIdent(r.(*ast.SelectorExpr).Sel))
		method, ok := c.collectedFunc.GetMethod(r.(*ast.SelectorExpr))
		if ok {
			if method.Name.Name == "ExecTimingSyncTask" {
				fmt.Println("hi")
			}
			c.r2r(&rs, c.parseFunc(method))
		}
	}
	return
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
	r := c.handleExpr(fnc)
	return r.needRecover()
}

func (c *Checker) handleIdent(ident *ast.Ident) (r rst) {
	//if _, ok := c.collectedFunc[ident]; ok {
	//	fmt.Println("abc ", c.fset.PositionFor(ident.Pos(), false))
	//}

	if identIsRecover(ident) {
		r.hasRecover = true
	}
	if obj := ident.Obj; obj != nil {
		if fd, ok := obj.Decl.(*ast.FuncDecl); ok {
			c.r2r(&r, c.parseFunc(fd))
		}
	}
	return
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

	return false
}

func identIsRecover(ident *ast.Ident) bool {
	return ident.Name == "recover"
}
