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
	"slices"
	"strings"

	"golang.org/x/tools/go/packages"
)

// binding says how to obtain a value of type out: by calling a provider
// function, by aliasing a concrete type (target), or from a literal (expr).
type binding struct {
	position token.Position
	out      types.Type
	fn       *types.Func
	args     []types.Type
	ins      []types.Type
	res      results
	target   types.Type
	expr     string
	fields   []string   // Struct: fields of the struct literal, in order (ins hold their types)
	ptr      bool       // Struct: build &T{} instead of T{}
	field    string     // FieldsOf: field read from ins[0]
	many     bool       // a []T built from elems
	elems    []*binding // elements of a Many
}

func (b *binding) describe() string {
	switch {
	case b.fn != nil:
		if len(b.args) == 0 {
			return b.fn.Name()
		}
		args := make([]string, len(b.args))
		for i, t := range b.args {
			args[i] = types.TypeString(t, nil)
		}
		return fmt.Sprintf("%s[%s]", b.fn.Name(), strings.Join(args, ", "))
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
	out         types.Type
	cleanup     bool
	err         bool
	cleanupType types.Type
	lifecycle   bool
}

func parseResults(sig *types.Signature) (r results, err error) {
	const shape = "results must be T, [func() or func(context.Context) error], [error]"
	rs := sig.Results()
	if rs.Len() == 0 || rs.Len() > 3 {
		return r, fmt.Errorf("%s", shape)
	}
	r.out = rs.At(0).Type()
	for i := 1; i < rs.Len(); i++ {
		t := rs.At(i).Type()
		switch {
		case isError(t) && i == rs.Len()-1:
			r.err = true
		case (isCleanup(t) || isLifecycleCleanup(t)) && i == 1:
			r.cleanup = true
			r.cleanupType = t
			r.lifecycle = isLifecycleCleanup(t)
		default:
			return r, fmt.Errorf("%s", shape)
		}
	}
	return r, nil
}

func isError(t types.Type) bool { return types.Identical(t, types.Universe.Lookup("error").Type()) }

func isCleanup(t types.Type) bool {
	s, ok := t.Underlying().(*types.Signature)
	return ok && s.Params().Len() == 0 && s.Results().Len() == 0
}

func isLifecycleCleanup(t types.Type) bool {
	s, ok := t.Underlying().(*types.Signature)
	return ok && !s.Variadic() && s.Params().Len() == 1 && isContext(s.Params().At(0).Type()) &&
		s.Results().Len() == 1 && isError(s.Results().At(0).Type())
}

func isContext(t types.Type) bool {
	n, ok := types.Unalias(t).(*types.Named)
	return ok && n.Obj().Pkg() != nil && n.Obj().Pkg().Path() == "context" && n.Obj().Name() == "Context"
}

// collector expands the arguments of Build into bindings.
type collector struct {
	f        *file
	bindings map[typeID]*binding
	visiting map[token.Pos]bool
	many     *[]*binding     // set while collecting the elements of a Many
	autos    []types.Type    // AutoBind types
	closers  map[typeID]bool // type identities registered with Closer
}

func (c *collector) add(b *binding) error {
	if c.many != nil {
		*c.many = append(*c.many, b)
		return nil
	}
	k := c.f.types.key(b.out)
	if old, ok := c.bindings[k]; ok && (old.fn == nil || b.fn == nil ||
		c.f.types.provider(old.fn, old.args) != c.f.types.provider(b.fn, b.args)) {
		return fmt.Errorf("multiple providers for %s (%s, %s)\n\tfirst: %s\n\tsecond: %s",
			c.f.typ(b.out), old.describe(), b.describe(), old.position, b.position)
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
			return c.add(&binding{position: at, out: types.NewSlice(targs[0]), many: true, elems: elems})
		case "AutoBind", "Closer":
			if len(targs) != 1 || c.many != nil {
				break
			}
			if name == "AutoBind" {
				c.autos = append(c.autos, targs[0])
			} else if _, err := closeStmt(targs[0], "x"); err != nil {
				return fmt.Errorf("%s: Closer: %w", at, err)
			} else {
				c.closers[c.f.types.key(targs[0])] = true
			}
			return nil
		case "Bind":
			if len(targs) != 2 || c.many != nil {
				break
			}
			if iface, ok := targs[0].Underlying().(*types.Interface); !ok || !types.Implements(targs[1], iface) {
				return fmt.Errorf("%s: Bind: %s does not implement %s", at, c.f.typ(targs[1]), c.f.typ(targs[0]))
			}
			return c.add(&binding{position: at, out: targs[0], target: targs[1]})
		case "Struct":
			if len(targs) != 1 || c.many != nil {
				break
			}
			return c.addStruct(targs[0], v, p, at)
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
			return c.add(&binding{position: at, out: targs[0], expr: s})
		}
	case *ast.Ident, *ast.SelectorExpr, *ast.IndexExpr, *ast.IndexListExpr:
		id := functionIdent(v)
		if id == nil {
			break
		}
		switch o := p.TypesInfo.Uses[id].(type) {
		case *types.Func:
			return c.addFunc(o, p.TypesInfo, id, at)
		case *types.Var:
			switch v.(type) {
			case *ast.IndexExpr, *ast.IndexListExpr:
				return fmt.Errorf("%s: indexed values are not provider declarations", at)
			}
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

func (c *collector) addFunc(fn *types.Func, info *types.Info, id *ast.Ident, at token.Position) error {
	sig := fn.Type().(*types.Signature)
	var args []types.Type
	if sig.TypeParams().Len() > 0 {
		instance, ok := info.Instances[id]
		if !ok || instance.TypeArgs.Len() != sig.TypeParams().Len() {
			return fmt.Errorf("%s: provider %s requires explicit type arguments", at, fn.Name())
		}
		args = slices.Collect(instance.TypeArgs.Types())
		sig = instance.Type.(*types.Signature)
	}
	if sig.Recv() != nil || sig.Variadic() {
		return fmt.Errorf("%s: provider %s must be a plain, non-variadic function", at, fn.Name())
	}
	if fn.Pkg() != c.f.pkg.Types && !fn.Exported() {
		return fmt.Errorf("%s: provider %s is not exported", at, fn.FullName())
	}
	res, err := parseResults(sig)
	if err != nil {
		return fmt.Errorf("%s: provider %s: %w", at, fn.Name(), err)
	}
	b := &binding{position: c.f.pkg.Fset.Position(fn.Pos()), out: res.out, fn: fn, args: args, res: res}
	for p := range sig.Params().Variables() {
		b.ins = append(b.ins, p.Type())
	}
	return c.add(b)
}

// valueExpr prints a Value argument, reserving the imports it relies on.
func (f *file) valueExpr(x ast.Expr, p *packages.Package) (string, error) {
	var err error
	type rename struct {
		id   *ast.Ident
		name string
	}
	var renamed []rename
	defer func() {
		for _, r := range renamed {
			r.id.Name = r.name
		}
	}()
	ast.Inspect(x, func(n ast.Node) bool {
		switch v := n.(type) {
		case *ast.SelectorExpr:
			if id, ok := v.X.(*ast.Ident); ok {
				if pn, ok := p.TypesInfo.Uses[id].(*types.PkgName); ok {
					renamed = append(renamed, rename{id, id.Name})
					id.Name = f.qual(pn.Imported())
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

// addStruct registers T and *T, both built from fields of struct T: the
// exported untagged ones, or the named ones ("*" means the default set).
func (c *collector) addStruct(t types.Type, call *ast.CallExpr, p *packages.Package, at token.Position) error {
	st, ok := t.Underlying().(*types.Struct)
	if !ok {
		return fmt.Errorf("%s: Struct: %s is not a struct", at, c.f.typ(t))
	}
	names, err := stringArgs(call, p, at, "Struct")
	if err != nil {
		return err
	}
	if len(names) == 0 || slices.Equal(names, []string{"*"}) {
		names = nil
		for i := range st.NumFields() {
			if f := st.Field(i); f.Exported() && reflect.StructTag(st.Tag(i)).Get("whyor") != "-" {
				names = append(names, f.Name())
			}
		}
	}
	b := &binding{position: at, out: t, fields: []string{}}
	seen := map[string]bool{}
	for _, name := range names {
		if seen[name] {
			return fmt.Errorf("%s: Struct: field %s is listed more than once", at, name)
		}
		seen[name] = true
		i := slices.IndexFunc(structFields(st), func(f *types.Var) bool { return f.Name() == name })
		if i < 0 || !st.Field(i).Exported() {
			return fmt.Errorf("%s: Struct: %s has no exported field %s", at, c.f.typ(t), name)
		}
		b.fields = append(b.fields, name)
		b.ins = append(b.ins, st.Field(i).Type())
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

func structFields(st *types.Struct) []*types.Var {
	out := make([]*types.Var, st.NumFields())
	for i := range out {
		out[i] = st.Field(i)
	}
	return out
}

// stringArgs returns the constant string arguments of call.
func stringArgs(call *ast.CallExpr, p *packages.Package, at token.Position, api string) ([]string, error) {
	var out []string
	for _, a := range call.Args {
		tv := p.TypesInfo.Types[a]
		if tv.Value == nil || tv.Value.Kind() != constant.String {
			return nil, fmt.Errorf("%s: %s: field names must be string constants", at, api)
		}
		out = append(out, constant.StringVal(tv.Value))
	}
	return out, nil
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
	names, err := stringArgs(call, p, at, "FieldsOf")
	if err != nil {
		return err
	}
	for _, name := range names {
		f, _, _ := types.LookupFieldOrMethod(t, true, c.f.pkg.Types, name)
		fv, ok := f.(*types.Var)
		if !ok || !fv.IsField() {
			return fmt.Errorf("%s: FieldsOf: %s has no field %s", at, c.f.typ(t), name)
		}
		if err := c.add(&binding{position: at, out: fv.Type(), ins: []types.Type{t}, field: name}); err != nil {
			return err
		}
	}
	return nil
}
