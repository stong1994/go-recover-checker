package main

import (
	"go/ast"
	"go/parser"
	"go/token"
)

type MethodMap struct {
	methods map[*ast.Object]map[string]*ast.FuncDecl
}

func NewMethodMap() *MethodMap {
	return &MethodMap{methods: make(map[*ast.Object]map[string]*ast.FuncDecl, 0)}
}

func (mm *MethodMap) AddMethod(obj *ast.Object, name string, fnc *ast.FuncDecl) {
	if mm.methods[obj] == nil {
		mm.methods[obj] = map[string]*ast.FuncDecl{name: fnc}
	} else {
		mm.methods[obj][name] = fnc
	}
}

func (mm *MethodMap) GetMethod(se *ast.SelectorExpr) (*ast.FuncDecl, bool) {
	if se.X == nil {
		return nil, false
	}
	ident, ok := se.X.(*ast.Ident)
	if !ok {
		return nil, false
	}
	if ident.Obj == nil {
		return nil, false
	}
	field, ok := ident.Obj.Decl.(*ast.Field)
	if !ok || field.Type == nil {
		return nil, false
	}
	ident, ok = field.Type.(*ast.Ident)
	if !ok || ident.Obj == nil {
		return nil, false
	}
	method, ok := mm.methods[ident.Obj][se.Sel.Name]
	return method, ok
}

func CollectFunc(filenameList []string) (*MethodMap, error) {
	m := NewMethodMap()
	fs := token.NewFileSet()
	files, err := findFiles(filenameList)
	if err != nil {
		return nil, err
	}
	for _, filename := range files {
		f, err := parser.ParseFile(fs, filename, nil, parser.AllErrors|parser.ParseComments)
		if err != nil {
			return nil, err
		}
		ast.Inspect(f, func(node ast.Node) bool {
			n, ok := node.(*ast.FuncDecl)
			if !ok || n.Recv == nil || len(n.Recv.List) == 0 {
				return true
			}
			t, ok := n.Recv.List[0].Type.(*ast.Ident)
			if !ok || t.Obj == nil {
				return true
			}
			m.AddMethod(t.Obj, n.Name.Name, n)
			return true
		})
	}
	return m, nil
}
