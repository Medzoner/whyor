package gen

import (
	"cmp"
	"errors"
	"fmt"
	"go/types"
	"maps"
	"slices"
	"strings"
)

// resolver turns the dependency graph of one injector into straight-line code.
type resolver struct {
	f           *file
	res         results
	bindings    map[typeID]*binding
	done        map[typeID]string // type identity -> variable holding its value
	stack       []types.Type      // types being resolved, to report cycles
	lines       strings.Builder
	cleanups    []cleanupCall // resources to release, in acquisition order
	autos       []types.Type
	closers     map[typeID]bool
	fns         map[providerKey]string // provider instantiation -> variable
	n           int
	initContext string
	names       map[string]bool
	errName     string
}

type cleanupCall struct {
	name    string
	context bool
	err     bool
}

func (r *resolver) resolve(t types.Type) (string, error) {
	k := r.f.types.key(t)
	if v, ok := r.done[k]; ok {
		return v, nil
	}
	for _, s := range r.stack {
		if types.Identical(s, t) {
			return "", r.failure(fmt.Errorf("dependency cycle: %s", r.path(t)), t)
		}
	}
	b, ok := r.lookup(t)
	if !ok {
		if candidates := r.autoCandidates(t); len(candidates) > 1 {
			names := make([]string, len(candidates))
			for i, candidate := range candidates {
				names[i] = r.f.typ(candidate)
			}
			slices.Sort(names)
			return "", r.failure(fmt.Errorf("ambiguous AutoBind for %s: %s; use an explicit whyor.Bind",
				r.f.typ(t), strings.Join(names, ", ")), t)
		}
		msg := "no provider for " + r.f.typ(t)
		if len(r.stack) > 0 {
			msg += " (needed by " + r.f.typ(r.stack[len(r.stack)-1]) + ")"
		}
		if h := r.hint(t); h != "" {
			msg += "\n\thint: " + h
		}
		return "", r.failure(errors.New(msg), t)
	}
	r.stack = append(r.stack, t)
	defer func() { r.stack = r.stack[:len(r.stack)-1] }()

	v, err := r.emit(b)
	if err != nil {
		return "", r.failure(err, nil)
	}
	r.done[k] = v
	return v, nil
}

func (r *resolver) lookup(t types.Type) (*binding, bool) {
	if b, ok := r.bindings[r.f.types.key(t)]; ok {
		return b, true
	}
	return r.autoBind(t)
}

// hint suggests how to provide t from what is already registered.
func (r *resolver) hint(t types.Type) string {
	bindings := slices.Collect(maps.Values(r.bindings))
	// Sort only for stable diagnostics, never for deciding type identity.
	slices.SortFunc(bindings, func(a, b *binding) int {
		return cmp.Compare(types.TypeString(a.out, nil), types.TypeString(b.out, nil))
	})
	for _, b := range bindings {
		o := b.out
		if iface, ok := t.Underlying().(*types.Interface); ok && !types.IsInterface(o) && types.Implements(o, iface) {
			return fmt.Sprintf("%s implements %s: add whyor.Bind[%s, %s]() or whyor.AutoBind[%s]()",
				r.f.typ(o), r.f.typ(t), r.f.typ(t), r.f.typ(o), r.f.typ(o))
		}
		if p, ok := types.Unalias(o).(*types.Pointer); ok && types.Identical(p.Elem(), t) {
			return fmt.Sprintf("a provider for %s exists, but not for %s", r.f.typ(o), r.f.typ(t))
		}
		if p, ok := types.Unalias(t).(*types.Pointer); ok && types.Identical(p.Elem(), o) {
			return fmt.Sprintf("a provider for %s exists, but not for %s", r.f.typ(o), r.f.typ(t))
		}
	}
	return ""
}

// autoBind finds the single AutoBind type implementing interface t.
func (r *resolver) autoBind(t types.Type) (*binding, bool) {
	found := r.autoCandidates(t)
	if len(found) != 1 {
		return nil, false
	}
	return &binding{out: t, target: found[0]}, true
}

func (r *resolver) autoCandidates(t types.Type) []types.Type {
	iface, ok := t.Underlying().(*types.Interface)
	if !ok {
		return nil
	}
	var found []types.Type
	seen := map[typeID]bool{}
	for _, a := range r.autos {
		id := r.f.types.key(a)
		if !seen[id] && types.Implements(a, iface) {
			found = append(found, a)
			seen[id] = true
		}
	}
	return found
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
	case b.field != "":
		src, err := r.resolve(b.ins[0])
		if err != nil {
			return "", err
		}
		v := r.next()
		fmt.Fprintf(&r.lines, "\t%s := %s.%s\n", v, src, b.field)
		return v, nil
	case b.fields != nil:
		parts := make([]string, len(b.ins))
		for i, in := range b.ins {
			a, err := r.resolve(in)
			if err != nil {
				return "", err
			}
			parts[i] = b.fields[i] + ": " + a
		}
		lit, typ := "", b.out
		if b.ptr {
			lit, typ = "&", b.out.(*types.Pointer).Elem()
		}
		v := r.next()
		fmt.Fprintf(&r.lines, "\t%s := %s%s{%s}\n", v, lit, r.f.typ(typ), strings.Join(parts, ", "))
		return v, nil
	case b.fn == nil:
		v := r.next()
		fmt.Fprintf(&r.lines, "\tvar %s %s = %s\n", v, r.f.typ(b.out), b.expr)
		return v, nil
	}

	key := r.f.types.provider(b.fn, b.args)
	if v, ok := r.fns[key]; ok {
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
	if b.res.lifecycle && !r.res.lifecycle {
		return "", fmt.Errorf("provider %s returns a context-aware cleanup; the injector must return whyor.Cleanup or func(context.Context) error", r.f.callName(b))
	}

	closer := false
	if !b.res.cleanup && r.closers[r.f.types.key(b.out)] {
		if !r.res.cleanup {
			return "", fmt.Errorf("%s has a Closer but the injector returns no cleanup", r.f.typ(b.out))
		}
		closer = true
	}

	v := r.next()
	lhs := []string{v}
	cleanupName := ""
	if b.res.cleanup {
		cleanupName = r.localName(fmt.Sprintf("cleanup%d", r.n))
		lhs = append(lhs, cleanupName)
	}
	if b.res.err {
		lhs = append(lhs, r.errName)
	}
	callee := r.f.callName(b)
	fmt.Fprintf(&r.lines, "\t%s := %s(%s)\n", strings.Join(lhs, ", "), callee, strings.Join(args, ", "))
	if b.res.err {
		fmt.Fprintf(&r.lines, "\tif %s != nil {\n%s\t\treturn %s\n\t}\n", r.errName, r.cleanupCalls("\t\t", r.failureContext()), r.failReturn())
	}
	if b.res.cleanup {
		r.cleanups = append(r.cleanups, cleanupCall{name: cleanupName, context: b.res.lifecycle, err: b.res.lifecycle})
	}
	if closer {
		stmt, _ := closeStmt(b.out, v)
		r.cleanups = append(r.cleanups, cleanupCall{name: v + ".Close", err: strings.HasPrefix(stmt, "_ = ")})
	}
	r.fns[key] = v
	return v, nil
}

func (r *resolver) next() string {
	for {
		r.n++
		name := fmt.Sprintf("v%d", r.n)
		if !r.nameTaken(name) {
			r.names[name] = true
			return name
		}
	}
}

// cleanupCalls calls the cleanups acquired so far, in reverse order.
func (r *resolver) cleanupCalls(indent, ctx string) string {
	var b strings.Builder
	for _, c := range slices.Backward(r.cleanups) {
		args := ""
		if c.context {
			args = ctx
		}
		call := fmt.Sprintf("%s(%s)", c.name, args)
		if c.err {
			if r.res.lifecycle {
				call = fmt.Sprintf("%s = %s.Join(%s, %s)", r.errName, r.standardPackage("errors"), r.errName, call)
			} else {
				call = "_ = " + call
			}
		}
		fmt.Fprintf(&b, "%s%s\n", indent, call)
	}
	return b.String()
}

func (r *resolver) failReturn() string {
	ret := r.f.zero(r.res.out)
	if r.res.cleanup {
		ret += ", nil"
	}
	return fmt.Sprintf("%s, %s", ret, r.errName)
}

func (r *resolver) cleanupFunc() string {
	if r.res.lifecycle {
		ctx := r.localName("_cleanupCtx")
		return fmt.Sprintf("func(%s %s.Context) error {\n\t\tvar %s error\n%s\t\treturn %s\n\t}",
			ctx, r.standardPackage("context"), r.errName, r.cleanupCalls("\t\t", ctx), r.errName)
	}
	if len(r.cleanups) == 0 {
		return "func() {}"
	}
	return "func() {\n" + r.cleanupCalls("\t\t", "") + "\t}"
}

// Failed initialization must not pass a canceled acquisition context to cleanup.
// Preserve values with WithoutCancel when an injector context is available.
func (r *resolver) failureContext() string {
	if !r.res.lifecycle {
		return ""
	}
	for _, c := range r.cleanups {
		if c.context {
			pkg := r.standardPackage("context")
			if r.initContext != "" {
				return fmt.Sprintf("%s.WithoutCancel(%s)", pkg, r.initContext)
			}
			return fmt.Sprintf("%s.Background()", pkg)
		}
	}
	return ""
}

func (r *resolver) standardPackage(path string) string {
	if pkg, ok := r.f.pkg.Imports[path]; ok {
		return r.f.qual(pkg.Types)
	}
	return r.f.qual(types.NewPackage(path, path))
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
		return "", fmt.Errorf("%s has no Close method", types.TypeString(t, nil))
	}
	sig := fn.Type().(*types.Signature)
	switch {
	case sig.Params().Len() != 0:
	case sig.Results().Len() == 0:
		return v + ".Close()", nil
	case sig.Results().Len() == 1 && isError(sig.Results().At(0).Type()):
		return "_ = " + v + ".Close()", nil
	}
	return "", fmt.Errorf("%s: Close must be func() or func() error", types.TypeString(t, nil))
}
