package main

import (
	"errors"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"io/fs"
)

type Program struct {
	runtimeImporter types.Importer
	fs              map[string]string
	ast             map[string]*ast.File
	pkgs            map[string]*types.Package
	fset            *token.FileSet
	info            *types.Info
}

func NewProgram(fs map[string]string) *Program {
	return &Program{
		runtimeImporter: importer.Default(),
		fs:              fs,
		ast:             make(map[string]*ast.File),
		pkgs:            make(map[string]*types.Package),
		fset:            token.NewFileSet(),
		info: &types.Info{
			Types:      map[ast.Expr]types.TypeAndValue{},
			Defs:       map[*ast.Ident]types.Object{},
			Uses:       map[*ast.Ident]types.Object{},
			Implicits:  map[ast.Node]types.Object{},
			Selections: map[*ast.SelectorExpr]*types.Selection{},
			Scopes:     map[ast.Node]*types.Scope{},
			InitOrder:  []*types.Initializer{},
		},
	}
}

func (p *Program) LoadFile(path string) (pkg *types.Package, f *ast.File, err error) {
	if pkg, ok := p.pkgs[path]; ok {
		return pkg, p.ast[path], nil
	}
	if src, ok := p.fs[path]; ok {
		f, err = parser.ParseFile(p.fset, path, src, parser.AllErrors)
	} else {
		f, err = parser.ParseFile(p.fset, path, nil, parser.AllErrors)
	}
	if err != nil {
		return nil, nil, err
	}
	conf := types.Config{Importer: p}
	pkg, err = conf.Check(path, p.fset, []*ast.File{f}, p.info)
	if err != nil {
		return nil, nil, err
	}
	p.ast[path] = f
	p.pkgs[path] = pkg
	return pkg, f, nil
}

func (p *Program) LoadPackage(path string) (pkgs map[string]*ast.Package, err error) {
	//if pkg, ok := p.pkgs[path]; ok {
	//	return pkg, p.ast[path], nil
	//}
	pkgs, err = parser.ParseDir(p.fset, path, nil, parser.AllErrors)

	//if err != nil {
	return
	//}
	//conf := types.Config{Importer: p}
	//pkg, err = conf.Check(path, p.fset, []*ast.File{f}, p.info)
	//if err != nil {
	//	return nil, nil, err
	//}
	//p.ast[path] = f
	//p.pkgs[path] = pkg
	//return pkg, f, nil
}

func (p *Program) Import(path string) (*types.Package, error) {
	if pkg, ok := p.pkgs[path]; ok {
		return pkg, nil
	}
	pkg, _, err := p.LoadFile(path) // todo 循环导包检查
	if errors.Is(err, fs.ErrNotExist) {
		return p.runtimeImporter.Import(path)
	}
	return pkg, err
}

//func (p *Program) GetFunc(start, end token.Pos) (ast.FuncDecl, error) {
//
//}
