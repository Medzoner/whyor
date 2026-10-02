package gen

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/constant"
	"go/printer"
	"go/token"
	"go/types"
	"reflect"

	"golang.org/x/tools/go/packages"
)

// binding says how to obtain a value of type out: by calling a provider
// function, by aliasing a concrete type (target), or from a literal (expr).
type binding struct {
	out    types.Type
	fn     *types.Func
	ins    []types.Type
	res    results
	target types.Type
	expr   string
	fields []string   // Struct: fields of the struct literal, in order (ins hold their types)
	ptr    bool       // Struct: build &T{} instead of T{}
	field  string     // FieldsOf: field read from ins[0]
	many   bool       // a []T built from elems
	elems  []*binding // elements of a Many
}

func (b *binding) describe() string {
	switch {
	case b.fn != nil:
		return b.fn.Name()
	case b.target != nil:
		return "Bind"
	case b.many:
		return "Many"
	case b.field != "":
		return "FieldsOf"
	case b.fields != nil:
		return "Struct"
	}
	return "Value"
}

// results describes the shape `T, [func()], [error]` of providers and injectors.
type results struct {
	out     types.Type
	cleanup bool
	err     bool
}

func parseResults(sig *types.Signature) (r results, err error) {
	rs := sig.Results()
	if rs.Len() == 0 || rs.Len() > 3 {
		return r, fmt.Errorf("results must be T, [func()], [error]")
	}
	r.out = rs.At(0).Type()
	for i := 1; i < rs.Len(); i++ {
		t := rs.At(i).Type()
		switch {
		case isError(t) && i == rs.Len()-1:
			r.err = true
		case isCleanup(t) && i == 1:
			r.cleanup = true
		default:
			return r, fmt.Errorf("results must be T, [func()], [error]")
		}
	}
	return r, nil
}

func isError(t types.Type) bool { return types.Identical(t, types.Universe.Lookup("error").Type()) }

func isCleanup(t types.Type) bool {
	s, ok := t.(*types.Signature)
	return ok && s.Params().Len() == 0 && s.Results().Len() == 0
}

// typeKey identifies a type, seeing through aliases (type X = Y).
func typeKey(t types.Type) string { return types.TypeString(unalias(t), nil) }

func unalias(t types.Type) types.Type {
	switch t := types.Unalias(t).(type) {
	case *types.Pointer:
		return types.NewPointer(unalias(t.Elem()))
	case *types.Slice:
		return types.NewSlice(unalias(t.Elem()))
	case *types.Array:
		return types.NewArray(unalias(t.Elem()), t.Len())
	case *types.Map:
		return types.NewMap(unalias(t.Key()), unalias(t.Elem()))
	case *types.Chan:
		return types.NewChan(t.Dir(), unalias(t.Elem()))
	default:
		return t
	}
}

// collector expands the arguments of Build into bindings.
type collector struct {
	f        *file
	bindings map[string]*binding
	visiting map[token.Pos]bool
	many     *[]*binding     // set while collecting the elements of a Many
	autos    []types.Type    // AutoBind types
	closers  map[string]bool // type keys registered with Closer
}

func (c *collector) add(b *binding) error {
	if c.many != nil {
		*c.many = append(*c.many, b)
		return nil
	}
	k := typeKey(b.out)
	if old, ok := c.bindings[k]; ok && (old.fn == nil || old.fn != b.fn) {
		return fmt.Errorf("multiple providers for %s (%s, %s)", c.f.typ(b.out), old.describe(), b.describe())
	}
	c.bindings[k] = b
	return nil
}

func (c *collector) collect(x ast.Expr, p *packages.Package) error {
	x = ast.Unparen(x)
	at := p.Fset.Position(x.Pos())
	switch v := x.(type) {
	case *ast.CallExpr:
		name, targs := api(p.TypesInfo, v.Fun)
		switch name {
		case "Set":
			for _, a := range v.Args {
				if err := c.collect(a, p); err != nil {
					return err
				}
			}
			return nil
		case "Many":
			if len(targs) != 1 || c.many != nil {
				break
			}
			var elems []*binding
			sub := *c
			sub.many = &elems
			for _, a := range v.Args {
				if err := sub.collect(a, p); err != nil {
					return err
				}
			}
			for _, e := range elems {
				if !types.AssignableTo(e.out, targs[0]) {
					return fmt.Errorf("%s: Many: %s is not assignable to %s", at, c.f.typ(e.out), c.f.typ(targs[0]))
				}
			}
			return c.add(&binding{out: types.NewSlice(targs[0]), many: true, elems: elems})
		case "AutoBind", "Closer":
			if len(targs) != 1 || c.many != nil {
				break
			}
			if name == "AutoBind" {
				c.autos = append(c.autos, targs[0])
			} else if _, err := closeStmt(targs[0], "x"); err != nil {
				return fmt.Errorf("%s: Closer: %w", at, err)
			} else {
				c.closers[typeKey(targs[0])] = true
			}
			return nil
		case "Bind":
			if len(targs) != 2 || c.many != nil {
				break
			}
			if iface, ok := targs[0].Underlying().(*types.Interface); !ok || !types.Implements(targs[1], iface) {
				return fmt.Errorf("%s: Bind: %s does not implement %s", at, c.f.typ(targs[1]), c.f.typ(targs[0]))
			}
			return c.add(&binding{out: targs[0], target: targs[1]})
		case "Struct":
			if len(targs) != 1 || c.many != nil {
				break
			}
			return c.addStruct(targs[0], at)
		case "FieldsOf":
			if len(targs) != 1 || c.many != nil {
				break
			}
			return c.addFieldsOf(targs[0], v, p, at)
		case "Value":
			if len(targs) != 1 || len(v.Args) != 1 {
				break
			}
			s, err := c.f.valueExpr(v.Args[0], p)
			if err != nil {
				return err
			}
			return c.add(&binding{out: targs[0], expr: s})
		}
	case *ast.Ident, *ast.SelectorExpr:
		id, ok := v.(*ast.Ident)
		if !ok {
			id = v.(*ast.SelectorExpr).Sel
		}
		switch o := p.TypesInfo.Uses[id].(type) {
		case *types.Func:
			return c.addFunc(o, at)
		case *types.Var:
			d, ok := c.f.decls[o.Pos()]
			if !ok || c.visiting[o.Pos()] {
				return fmt.Errorf("%s: cannot expand set %s", at, o.Name())
			}
			c.visiting[o.Pos()] = true
			defer delete(c.visiting, o.Pos())
			return c.collect(d.expr, d.pkg)
		}
	}
	return fmt.Errorf("%s: unsupported provider expression", at)
}

func (c *collector) addFunc(fn *types.Func, at token.Position) error {
	sig := fn.Type().(*types.Signature)
	if sig.Recv() != nil || sig.TypeParams().Len() > 0 || sig.Variadic() {
		return fmt.Errorf("%s: provider %s must be a plain, non-generic, non-variadic function", at, fn.Name())
	}
	if fn.Pkg() != c.f.pkg.Types && !fn.Exported() {
		return fmt.Errorf("%s: provider %s is not exported", at, fn.FullName())
	}
	res, err := parseResults(sig)
	if err != nil {
		return fmt.Errorf("%s: provider %s: %w", at, fn.Name(), err)
	}
	b := &binding{out: res.out, fn: fn, res: res}
	for p := range sig.Params().Variables() {
		b.ins = append(b.ins, p.Type())
	}
	return c.add(b)
}

// valueExpr prints a Value argument, reserving the imports it relies on.
func (f *file) valueExpr(x ast.Expr, p *packages.Package) (string, error) {
	var err error
	ast.Inspect(x, func(n ast.Node) bool {
		switch v := n.(type) {
		case *ast.SelectorExpr:
			if id, ok := v.X.(*ast.Ident); ok {
				if pn, ok := p.TypesInfo.Uses[id].(*types.PkgName); ok {
					err = f.reserve(id.Name, pn.Imported().Path())
					return false
				}
			}
		case *ast.Ident:
			if o := p.TypesInfo.Uses[v]; o != nil && o.Pkg() != nil && o.Pkg() != f.pkg.Types && o.Parent() == o.Pkg().Scope() {
				err = fmt.Errorf("%s: Value expression uses unqualified %s from another package", p.Fset.Position(v.Pos()), v.Name)
			}
		}
		return err == nil
	})
	if err != nil {
		return "", err
	}
	var b bytes.Buffer
	err = printer.Fprint(&b, p.Fset, x)
	return b.String(), err
}

// addStruct registers T and *T, both built from the exported fields of struct T.
func (c *collector) addStruct(t types.Type, at token.Position) error {
	st, ok := t.Underlying().(*types.Struct)
	if !ok {
		return fmt.Errorf("%s: Struct: %s is not a struct", at, c.f.typ(t))
	}
	b := &binding{out: t, fields: []string{}}
	for i := range st.NumFields() {
		f := st.Field(i)
		if !f.Exported() || reflect.StructTag(st.Tag(i)).Get("whyor") == "-" {
			continue
		}
		b.fields = append(b.fields, f.Name())
		b.ins = append(b.ins, f.Type())
	}
	if len(b.fields) == 0 {
		return fmt.Errorf("%s: Struct: %s has no injectable field", at, c.f.typ(t))
	}
	ptr := *b
	ptr.out, ptr.ptr = types.NewPointer(t), true
	if err := c.add(b); err != nil {
		return err
	}
	return c.add(&ptr)
}

// addFieldsOf registers each named field of t as a dependency read from a t value.
func (c *collector) addFieldsOf(t types.Type, call *ast.CallExpr, p *packages.Package, at token.Position) error {
	if ptr, ok := t.Underlying().(*types.Pointer); ok {
		if _, ok := ptr.Elem().Underlying().(*types.Struct); !ok {
			return fmt.Errorf("%s: FieldsOf: %s is not a struct", at, c.f.typ(t))
		}
	} else if _, ok := t.Underlying().(*types.Struct); !ok {
		return fmt.Errorf("%s: FieldsOf: %s is not a struct", at, c.f.typ(t))
	}
	for _, a := range call.Args {
		tv := p.TypesInfo.Types[a]
		if tv.Value == nil || tv.Value.Kind() != constant.String {
			return fmt.Errorf("%s: FieldsOf: field names must be string constants", at)
		}
		name := constant.StringVal(tv.Value)
		f, _, _ := types.LookupFieldOrMethod(t, true, c.f.pkg.Types, name)
		fv, ok := f.(*types.Var)
		if !ok || !fv.IsField() {
			return fmt.Errorf("%s: FieldsOf: %s has no field %s", at, c.f.typ(t), name)
		}
		if err := c.add(&binding{out: fv.Type(), ins: []types.Type{t}, field: name}); err != nil {
			return err
		}
	}
	return nil
}
