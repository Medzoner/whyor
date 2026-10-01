// Package gen implements the whyor code generator.
package gen

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"sort"

	"golang.org/x/tools/go/packages"
)

const (
	whyorPath = "github.com/medzoner/whyor"
	genFile   = "whyor_gen.go"
)

// Run loads the packages matching patterns (with the "whyor" build tag) and
// generates their injectors. It returns the files that are new or changed;
// they are written to disk only if write is true.
func Run(dir string, patterns []string, write bool) ([]string, error) {
	pkgs, err := packages.Load(&packages.Config{
		Mode: packages.NeedName | packages.NeedSyntax | packages.NeedTypes |
			packages.NeedTypesInfo | packages.NeedImports | packages.NeedDeps,
		Dir:        dir,
		BuildFlags: []string{"-tags=whyor"},
	}, patterns...)
	if err != nil {
		return nil, err
	}

	decls := map[token.Pos]varDecl{}
	var loadErr error
	packages.Visit(pkgs, nil, func(p *packages.Package) {
		if loadErr == nil && len(p.Errors) > 0 {
			loadErr = fmt.Errorf("%s: %v", p.PkgPath, p.Errors[0])
		}
		for _, f := range p.Syntax {
			indexVars(p, f, decls)
		}
	})
	if loadErr != nil {
		return nil, loadErr
	}

	var changed []string
	for _, pkg := range pkgs {
		src, err := generate(pkg, decls)
		if err != nil || src == nil {
			if err != nil {
				return nil, err
			}
			continue
		}
		out := filepath.Join(filepath.Dir(pkg.Fset.Position(pkg.Syntax[0].Pos()).Filename), genFile)
		if old, _ := os.ReadFile(out); bytes.Equal(old, src) {
			continue
		}
		changed = append(changed, out)
		if write {
			if err := os.WriteFile(out, src, 0o644); err != nil {
				return nil, err
			}
		}
	}
	sort.Strings(changed)
	return changed, nil
}

// varDecl is a package-level `var x = expr`, used to expand sets.
type varDecl struct {
	expr ast.Expr
	pkg  *packages.Package
}

func indexVars(p *packages.Package, f *ast.File, out map[token.Pos]varDecl) {
	for _, d := range f.Decls {
		gd, ok := d.(*ast.GenDecl)
		if !ok || gd.Tok != token.VAR {
			continue
		}
		for _, s := range gd.Specs {
			vs := s.(*ast.ValueSpec)
			if len(vs.Values) != len(vs.Names) {
				continue
			}
			for i, n := range vs.Names {
				out[n.Pos()] = varDecl{vs.Values[i], p}
			}
		}
	}
}

// generate returns the formatted generated file of pkg, or nil if it has no injector.
func generate(pkg *packages.Package, decls map[token.Pos]varDecl) ([]byte, error) {
	g := newFile(pkg, decls)
	for _, f := range pkg.Syntax {
		for _, d := range f.Decls {
			if fd, ok := d.(*ast.FuncDecl); ok {
				if call := buildCall(pkg.TypesInfo, fd); call != nil {
					if err := g.injector(fd, call); err != nil {
						return nil, fmt.Errorf("%s.%s: %w", pkg.PkgPath, fd.Name.Name, err)
					}
				}
			}
		}
	}
	if g.body.Len() == 0 {
		return nil, nil
	}
	src, err := format.Source(g.bytes())
	if err != nil {
		return nil, fmt.Errorf("%s: formatting generated code: %w", pkg.PkgPath, err)
	}
	return src, nil
}

// buildCall returns the whyor.Build call if fd's body is `panic(whyor.Build(...))`.
func buildCall(info *types.Info, fd *ast.FuncDecl) *ast.CallExpr {
	if fd.Body == nil || len(fd.Body.List) != 1 {
		return nil
	}
	es, _ := fd.Body.List[0].(*ast.ExprStmt)
	if es == nil {
		return nil
	}
	p, _ := es.X.(*ast.CallExpr)
	if p == nil || len(p.Args) != 1 {
		return nil
	}
	if id, _ := p.Fun.(*ast.Ident); id == nil || id.Name != "panic" {
		return nil
	}
	c, _ := p.Args[0].(*ast.CallExpr)
	if c == nil {
		return nil
	}
	if name, _ := api(info, c.Fun); name != "Build" {
		return nil
	}
	return c
}

// api reports the whyor function (and its type arguments) that e refers to.
func api(info *types.Info, e ast.Expr) (string, []types.Type) {
	for {
		switch x := e.(type) {
		case *ast.ParenExpr:
			e = x.X
			continue
		case *ast.IndexExpr:
			e = x.X
			continue
		case *ast.IndexListExpr:
			e = x.X
			continue
		}
		break
	}
	id, _ := e.(*ast.Ident)
	if sel, ok := e.(*ast.SelectorExpr); ok {
		id = sel.Sel
	}
	if id == nil {
		return "", nil
	}
	fn, _ := info.Uses[id].(*types.Func)
	if fn == nil || fn.Pkg() == nil || fn.Pkg().Path() != whyorPath {
		return "", nil
	}
	var targs []types.Type
	if inst, ok := info.Instances[id]; ok {
		for t := range inst.TypeArgs.Types() {
			targs = append(targs, t)
		}
	}
	return fn.Name(), targs
}
