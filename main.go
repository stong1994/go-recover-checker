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
	fset := token.NewFileSet()
	files, err := findFiles([]string{"./data/"})
	if err != nil {
		return
	}
	p := NewParser()
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
	for _, fnc := range p.funcSet.FindNeedRecoveredFunc() {
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

type Parser struct {
	funcSet *entity.FuncSet
	fileSet token.FileSet
	astFile ast.File
}

func NewParser() *Parser {
	fs := entity.NewFuncSet()
	p := &Parser{funcSet: fs}

	return p
}

func (p *Parser) parseFunc(fnc *ast.FuncDecl) {
	if fnc.Body == nil || len(fnc.Body.List) == 0 { // body is empty, ignore
		return
	}
	for _, stmt := range fnc.Body.List {
		rst := p.handleStmt(stmt)
		if rst.needRecover {
			fmt.Println("need recover", fnc.Name.Name, stmt.Pos(), stmt.End())
		}
	}
}

type rstHandleIdent struct {
	isRecover bool
}

type rstHandleStmt struct {
	isRecover   bool
	needRecover bool
}

func (p *Parser) handleStmt(stmt ast.Stmt) (rst rstHandleStmt) {
	switch stmt.(type) {
	case *ast.GoStmt:
		r := p.handleExpr(stmt.(*ast.GoStmt).Call.Fun)
		if !r.isRecover {
			rst.needRecover = true
		}
		for _, expr := range stmt.(*ast.GoStmt).Call.Args {
			p.handleExpr(expr)
		}
	case *ast.DeferStmt:
		st := stmt.(*ast.DeferStmt)
		if st == nil {
			return
		}
		p.fillRstExpr2Stmt(&rst, p.handleExpr(st.Call.Fun))
		for _, expr := range st.Call.Args {
			p.fillRstExpr2Stmt(&rst, p.handleExpr(expr))
		}
	case *ast.AssignStmt:
		aStmt := stmt.(*ast.AssignStmt)
		for _, expr := range aStmt.Lhs {
			p.fillRstExpr2Stmt(&rst, p.handleExpr(expr))
		}
		for _, expr := range aStmt.Rhs {
			p.fillRstExpr2Stmt(&rst, p.handleExpr(expr))
		}
	case *ast.IfStmt:
		ifStmt := stmt.(*ast.IfStmt)
		p.handleStmt(ifStmt.Init)
		p.handleStmt(ifStmt.Else)
		p.fillRstExpr2Stmt(&rst, p.handleExpr(ifStmt.Cond))
		p.fillRstBlock2Stmt(&rst, p.handleBody(ifStmt.Body))
	case *ast.ExprStmt:
		p.fillRstExpr2Stmt(&rst, p.handleExpr(stmt.(*ast.ExprStmt).X))
	}
	return
}

type rstHandleExpr struct {
	isRecover bool
}

func (p *Parser) handleExpr(expr ast.Expr) (rst rstHandleExpr) {
	switch expr.(type) {
	case *ast.CallExpr:
		cExpr := expr.(*ast.CallExpr)
		p.fillRstExpr2Expr(&rst, p.handleExpr(cExpr.Fun))
		for _, expr := range cExpr.Args {
			p.fillRstExpr2Expr(&rst, p.handleExpr(expr))
		}
	case *ast.SelectorExpr:
		p.fillRstExpr2Expr(&rst, p.handleExpr(expr.(*ast.SelectorExpr).X))
	case *ast.FuncLit:
		body := expr.(*ast.FuncLit).Body
		if body == nil {
			return
		}
		for _, stmt := range body.List {
			p.handleStmt(stmt)
		}
	case *ast.Ident:
		obj := expr.(*ast.Ident).Obj
		if obj == nil {
			return
		}
		if expr.(*ast.Ident).Name == "recover" {
			rst.isRecover = true
		}
		if fnc, ok := obj.Decl.(*ast.FuncDecl); ok {
			p.parseFunc(fnc)
		}
	}
	return
}

func (p *Parser) fillRstIdent2Expr(dst *rstHandleExpr, src rstHandleIdent) {
	if src.isRecover {
		dst.isRecover = true
	}
}

func (p *Parser) fillRstExpr2Expr(dst *rstHandleExpr, src rstHandleExpr) {
	if src.isRecover {
		dst.isRecover = true
	}
}

func (p *Parser) fillRstExpr2Stmt(dst *rstHandleStmt, src rstHandleExpr) {
	if src.isRecover {
		dst.isRecover = true
	}
}

type rstHandleBlock struct {
	isRecover bool
}

func (p *Parser) handleBody(body *ast.BlockStmt) (rst rstHandleBlock) {
	if body == nil {
		return
	}
	for _, stmt := range body.List {
		p.fillRstStmt2Block(&rst, p.handleStmt(stmt))
	}
	return
}

func (p *Parser) fillRstStmt2Block(dst *rstHandleBlock, src rstHandleStmt) {
	if src.isRecover {
		dst.isRecover = true
	}
}

func (p *Parser) fillRstBlock2Stmt(dst *rstHandleStmt, src rstHandleBlock) {
	if src.isRecover {
		dst.isRecover = true
	}
}
