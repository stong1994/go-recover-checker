package entity

import (
	"fmt"
	"go/ast"
)

type Func struct {
	ident           *ast.Ident // same function different position is different Func
	goChildren      []*Func    // go function list that has called in Func
	recoverList     []bool     // corresponding to goChildren (that means len(recoverList) == len(goChildren)), mark each go func has been recovered in Func
	isSelfRecovered bool       // ex: func hello() { recover()}
	childrenMap     map[*ast.Ident]struct{}
}

func NewFunc(ident *ast.Ident) *Func {
	return &Func{
		ident:       ident,
		childrenMap: make(map[*ast.Ident]struct{}),
	}
}

func (f Func) String() string {
	return f.ident.String()
}

func (f Func) HasSelfRecovered() bool {
	return f.isSelfRecovered
}

func (f Func) HasChild(ident *ast.Ident) bool {
	_, ok := f.childrenMap[ident]
	return ok
}

func (f Func) AddChild(ident *ast.Ident) {
	if f.HasChild(ident) {
		return
	}
	child := NewFunc(ident)
	f.childrenMap[ident] = struct{}{}
	f.goChildren = append(f.goChildren, child)
	f.recoverList = append(f.recoverList, f.HasOuterRecovered(ident))
	return
}

func (f Func) ChildFunc(i int) (child *Func, hasOuterRecovered bool) {
	if i >= len(f.goChildren) {
		panic(fmt.Sprintf("invalid index, len: %d, index: %d", len(f.goChildren), i))
	}
	if len(f.goChildren) != len(f.recoverList) {
		panic(fmt.Sprintf("recoverList not match goChildren, %d:%d", len(f.goChildren), len(f.recoverList)))
	}
	return f.goChildren[i], f.recoverList[i]
}

func (f Func) Iterator(fnc func(child *Func, hasOuterRecovered bool)) {
	if len(f.goChildren) != len(f.recoverList) {
		panic(fmt.Sprintf("recoverList not match goChildren, %d:%d", len(f.goChildren), len(f.recoverList)))
	}
	for i := range f.recoverList {
		fnc(f.ChildFunc(i))
	}
}

func (f Func) HasOuterRecovered(ident *ast.Ident) bool {
	return false // todo
}

type FuncSet struct {
	m     map[*ast.Ident]*Func
	roots []*Func
}

func NewFuncSet() FuncSet {
	return FuncSet{
		m: make(map[*ast.Ident]*Func),
	}
}

func (fs FuncSet) HasFunc(ident *ast.Ident) bool {
	_, ok := fs.m[ident]
	return ok
}

func (fs FuncSet) Iterator(f func(ident *Func)) {
	for _, v := range fs.m {
		f(v)
	}
}

func (fs FuncSet) AddFunc(ident *ast.Ident) {
	if _, ok := fs.m[ident]; ok {
		return
	}
	fs.m[ident] = &Func{
		ident: ident,
	}
}

func (fs FuncSet) AddFuncChild(parent, child *ast.Ident) {
	if _, ok := fs.m[child]; ok {
		p := fs.m[parent]
		if !p.HasChild(child) {
			p.AddChild(child)
		}
		return
	}
	fs.roots = append(fs.roots, NewFunc(child))
}

func (fs FuncSet) FindNeedRecoveredFunc() (rst []*Func) {
	var dfs func(fnc *Func, hasOuterRecovered bool)
	dfs = func(fnc *Func, hasOuterRecovered bool) {
		if !hasOuterRecovered && !fnc.HasSelfRecovered() {
			rst = append(rst, fnc)
		}
		fnc.Iterator(func(child *Func, hasOuterRecovered bool) {
			dfs(child, hasOuterRecovered)
		})
	}
	for _, v := range fs.roots {
		dfs(v, false)
	}
	return
}
