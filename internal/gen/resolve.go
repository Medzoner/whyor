package gen

import (
	"fmt"
	"go/types"
	"slices"
	"strings"
)

// resolver turns the dependency graph of one injector into straight-line code.
type resolver struct {
	f        *file
	res      results
	bindings map[string]*binding
	done     map[string]string // type key -> variable holding its value
	stack    []types.Type      // types being resolved, to report cycles
	lines    strings.Builder
	cleanups []string // statements releasing what was acquired, in acquisition order
	autos    []types.Type
	closers  map[string]bool
	fns      map[*types.Func]string // provider -> variable, so each runs once
	n        int
}

func (r *resolver) resolve(t types.Type) (string, error) {
	k := typeKey(t)
	if v, ok := r.done[k]; ok {
		return v, nil
	}
	for _, s := range r.stack {
		if typeKey(s) == k {
			return "", fmt.Errorf("dependency cycle: %s", r.path(t))
		}
	}
	b, ok := r.bindings[k]
	if !ok {
		b, ok = r.autoBind(t)
	}
	if !ok {
		if len(r.stack) > 0 {
			return "", fmt.Errorf("no provider for %s (needed by %s)", r.f.typ(t), r.f.typ(r.stack[len(r.stack)-1]))
		}
		return "", fmt.Errorf("no provider for %s", r.f.typ(t))
	}
	r.stack = append(r.stack, t)
	defer func() { r.stack = r.stack[:len(r.stack)-1] }()

	v, err := r.emit(b)
	if err == nil {
		r.done[k] = v
	}
	return v, err
}

// autoBind finds the single AutoBind type implementing interface t.
func (r *resolver) autoBind(t types.Type) (*binding, bool) {
	iface, ok := t.Underlying().(*types.Interface)
	if !ok {
		return nil, false
	}
	var found []types.Type
	for _, a := range r.autos {
		if types.Implements(a, iface) {
			found = append(found, a)
		}
	}
	if len(found) != 1 {
		return nil, false
	}
	return &binding{out: t, target: found[0]}, true
}

func (r *resolver) emit(b *binding) (string, error) {
	switch {
	case b.target != nil:
		return r.resolve(b.target)
	case b.many:
		vs := make([]string, len(b.elems))
		for i, e := range b.elems {
			v, err := r.emit(e)
			if err != nil {
				return "", err
			}
			vs[i] = v
		}
		v := r.next()
		fmt.Fprintf(&r.lines, "\t%s := %s{%s}\n", v, r.f.typ(b.out), strings.Join(vs, ", "))
		return v, nil
	case b.fn == nil:
		v := r.next()
		fmt.Fprintf(&r.lines, "\t%s := %s\n", v, b.expr)
		return v, nil
	}

	if v, ok := r.fns[b.fn]; ok {
		return v, nil
	}
	args := make([]string, len(b.ins))
	for i, in := range b.ins {
		a, err := r.resolve(in)
		if err != nil {
			return "", err
		}
		args[i] = a
	}
	if b.res.err && !r.res.err {
		return "", fmt.Errorf("provider %s returns an error but the injector does not", b.fn.Name())
	}
	if b.res.cleanup && !r.res.cleanup {
		return "", fmt.Errorf("provider %s returns a cleanup but the injector does not", b.fn.Name())
	}

	closer := false
	if !b.res.cleanup && r.closers[typeKey(b.out)] {
		if !r.res.cleanup {
			return "", fmt.Errorf("%s has a Closer but the injector returns no cleanup", r.f.typ(b.out))
		}
		closer = true
	}

	v := r.next()
	lhs := []string{v}
	if b.res.cleanup {
		lhs = append(lhs, "cleanup"+v[1:])
	}
	if b.res.err {
		lhs = append(lhs, "err")
	}
	callee := b.fn.Name()
	if b.fn.Pkg() != r.f.pkg.Types {
		callee = r.f.qual(b.fn.Pkg()) + "." + callee
	}
	fmt.Fprintf(&r.lines, "\t%s := %s(%s)\n", strings.Join(lhs, ", "), callee, strings.Join(args, ", "))
	if b.res.err {
		fmt.Fprintf(&r.lines, "\tif err != nil {\n%s\t\treturn %s\n\t}\n", r.cleanupCalls("\t\t"), r.failReturn())
	}
	if b.res.cleanup {
		r.cleanups = append(r.cleanups, "cleanup"+v[1:]+"()")
	}
	if closer {
		stmt, _ := closeStmt(b.out, v)
		r.cleanups = append(r.cleanups, stmt)
	}
	r.fns[b.fn] = v
	return v, nil
}

func (r *resolver) next() string {
	r.n++
	return fmt.Sprintf("v%d", r.n)
}

// cleanupCalls calls the cleanups acquired so far, in reverse order.
func (r *resolver) cleanupCalls(indent string) string {
	var b strings.Builder
	for _, c := range slices.Backward(r.cleanups) {
		b.WriteString(indent + c + "\n")
	}
	return b.String()
}

func (r *resolver) failReturn() string {
	ret := r.f.zero(r.res.out)
	if r.res.cleanup {
		ret += ", nil"
	}
	return ret + ", err"
}

func (r *resolver) cleanupFunc() string {
	if len(r.cleanups) == 0 {
		return "func() {}"
	}
	return "func() {\n" + r.cleanupCalls("\t\t") + "\t}"
}

func (r *resolver) path(t types.Type) string {
	var names []string
	for _, s := range r.stack {
		names = append(names, r.f.typ(s))
	}
	return strings.Join(append(names, r.f.typ(t)), " -> ")
}

// closeStmt returns the statement calling v.Close(), if t has a usable Close method.
func closeStmt(t types.Type, v string) (string, error) {
	obj, _, _ := types.LookupFieldOrMethod(t, true, nil, "Close")
	fn, _ := obj.(*types.Func)
	if fn == nil {
		return "", fmt.Errorf("%s has no Close method", typeKey(t))
	}
	sig := fn.Type().(*types.Signature)
	switch {
	case sig.Params().Len() != 0:
	case sig.Results().Len() == 0:
		return v + ".Close()", nil
	case sig.Results().Len() == 1 && isError(sig.Results().At(0).Type()):
		return "_ = " + v + ".Close()", nil
	}
	return "", fmt.Errorf("%s: Close must be func() or func() error", typeKey(t))
}
