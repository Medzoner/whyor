package gen

import (
	"fmt"
	"go/types"
	"strings"
)

// child is a node of a dependency tree: a type, and the binding providing it
// when that is already known (elements of a Many).
type child struct {
	t types.Type
	b *binding
}

// writeTree prints the dependency tree of an injector returning out.
// given holds the injector parameters, keyed by type.
func (r *resolver) writeTree(w *strings.Builder, name string, out types.Type, given map[string]string) {
	fmt.Fprintf(w, "%s\n", name)
	seen := map[string]bool{}
	r.node(w, child{t: out}, "", true, given, seen)
	w.WriteString("\n")
}

func (r *resolver) node(w *strings.Builder, c child, prefix string, last bool, given map[string]string, seen map[string]bool) {
	branch, next := "├── ", "│   "
	if last {
		branch, next = "└── ", "    "
	}
	k := typeKey(c.t)
	b := c.b
	if b == nil {
		if p, ok := given[k]; ok {
			fmt.Fprintf(w, "%s%s%s [parameter %s]\n", prefix, branch, r.f.typ(c.t), p)
			return
		}
		b, _ = r.lookup(c.t)
	}
	label := fmt.Sprintf("%s [%s]", r.f.typ(c.t), r.provider(b))
	if seen[k] && c.b == nil {
		fmt.Fprintf(w, "%s%s%s (*)\n", prefix, branch, label)
		return
	}
	seen[k] = true
	fmt.Fprintf(w, "%s%s%s\n", prefix, branch, label)

	kids := children(b)
	for i, kid := range kids {
		r.node(w, kid, prefix+next, i == len(kids)-1, given, seen)
	}
}

func children(b *binding) []child {
	var out []child
	switch {
	case b.target != nil:
		out = append(out, child{t: b.target})
	case b.many:
		for _, e := range b.elems {
			out = append(out, child{t: e.out, b: e})
		}
	default:
		for _, in := range b.ins {
			out = append(out, child{t: in})
		}
	}
	return out
}

// provider names what builds a binding, qualifying functions of other packages.
func (r *resolver) provider(b *binding) string {
	switch {
	case b.fn != nil && b.fn.Pkg() != r.f.pkg.Types:
		return b.fn.Pkg().Name() + "." + b.fn.Name()
	case b.field != "":
		return "field " + b.field
	case b.target != nil:
		return "Bind"
	}
	return b.describe()
}
