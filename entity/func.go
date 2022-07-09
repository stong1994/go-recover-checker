package entity

import "go/ast"

type Func struct {
	AstFunc         *ast.Ident
	HasInnerRecover bool // ex: func hello() { recover()}
	HasOuterRecover bool // ex: go hello()
}

func (f Func) String() string {
	return f.AstFunc.String()
}

func (f Func) NeedRecover() bool {
	return !(f.HasInnerRecover || f.HasOuterRecover)
}

type FuncMap struct {
	m map[*ast.Ident]*Func
}

func NewFuncMap() FuncMap {
	return FuncMap{
		m: make(map[*ast.Ident]*Func),
	}
}

func (fm FuncMap) HasFunc(ident *ast.Ident) bool {
	_, ok := fm.m[ident]
	return ok
}

func (fm FuncMap) Iterator(f func(ident *Func)) {
	for _, v := range fm.m {
		f(v)
	}
}

func (fm FuncMap) InitFunc(ident *ast.Ident) {
	if _, ok := fm.m[ident]; ok {
		return
	}
	fm.m[ident] = &Func{
		AstFunc: ident,
	}
}

func (fm FuncMap) SetFuncHasOuterRecover(ident *ast.Ident) {
	fnc, ok := fm.m[ident]
	if ok {
		fnc.HasOuterRecover = true
	} else {
		fm.m[ident] = &Func{
			AstFunc:         ident,
			HasOuterRecover: true,
		}
	}
}
