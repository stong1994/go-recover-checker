package main

import (
	"go/ast"
	"strings"
)

type Config struct {
	IgnoreComment   string // ignore no-recover warning, ex: // no-recover-warning
	RecoverMaxDepth int    // recover function should not too far from the method definition, -1 means no limit
}

type Checker struct {
	needRecoverList []*ast.GoStmt
	Config
}

func NewChecker() *Checker {
	return &Checker{}
}

func (p *Checker) parseFunc(fnc *ast.FuncDecl) {
	if fnc.Body == nil || len(fnc.Body.List) == 0 { // body is empty, ignore
		return
	}
	for _, stmt := range fnc.Body.List {
		if g, ok := stmt.(*ast.GoStmt); ok {
			if p.isNeedRecover(g) {
				p.needRecoverList = append(p.needRecoverList, g)
			}
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
func (p *Checker) isNeedRecover(stmt *ast.GoStmt) (needRecover bool) {
	fnc := stmt.Call.Fun
	switch fnc.(type) {
	case *ast.Ident:
		f := fnc.(*ast.Ident)
		if identIsRecover(f) {
			return false
		}
		if obj := f.Obj; obj != nil {
			if fd, ok := obj.Decl.(*ast.FuncDecl); ok {
				if hasRecover := p.fncHasRecover(fd); hasRecover {
					return false
				}
			}
		}
	case *ast.FuncLit:
		if funcLitHasRecover(fnc.(*ast.FuncLit)) {
			return false
		}
	}
	return true
}

func (p *Checker) fncHasRecover(fnc *ast.FuncDecl) bool {
	if p.fncHasIgnoreComment(fnc) {
		return true
	}
	for _, stmt := range fnc.Body.List {
		switch stmt.(type) {
		case *ast.DeferStmt:
			sf := stmt.(*ast.DeferStmt).Call.Fun
			switch sf.(type) {
			case *ast.Ident:
				if identIsRecover(sf.(*ast.Ident)) {
					return true
				}
			case *ast.FuncLit:
				if funcLitHasRecover(sf.(*ast.FuncLit)) {
					return true
				}
			}
		}
	}
	return false
}

func (p *Checker) fncHasIgnoreComment(fnc *ast.FuncDecl) bool {
	if fnc.Doc == nil {
		return false
	}
	for _, doc := range fnc.Doc.List {
		if strings.Contains(doc.Text, p.IgnoreComment) {
			return true
		}
	}
	return false
}

func funcLitHasRecover(fl *ast.FuncLit) bool {
	for _, s := range fl.Body.List {
		switch s.(type) {
		case *ast.AssignStmt:
			for _, r := range s.(*ast.AssignStmt).Rhs {
				if c, ok := r.(*ast.CallExpr); ok {
					if f, ok := c.Fun.(*ast.Ident); ok {
						if identIsRecover(f) {
							return true
						}
					}
				}
			}
		case *ast.ExprStmt:
			if ce, ok := s.(*ast.ExprStmt).X.(*ast.CallExpr); ok {
				switch ce.Fun.(type) {
				case *ast.FuncLit:
					if funcLitHasRecover(ce.Fun.(*ast.FuncLit)) {
						return true
					}
				case *ast.Ident:
					if identIsRecover(ce.Fun.(*ast.Ident)) {
						return true
					}
				}
			}
		}
	}
	return false
}

func identIsRecover(ident *ast.Ident) bool {
	return ident.Name == "recover"
}
