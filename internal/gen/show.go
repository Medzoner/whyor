package gen

import (
	"fmt"
	"go/types"
	"strings"
)

// child is a node of a dependency graph: a type, and the binding providing it
// when that is already known (elements of a Many).
type child struct {
	t types.Type
	b *binding
}

// writeGraph prints the dependency graph of an injector returning out.
// given holds the injector parameters, keyed by type.
func (r *resolver) writeGraph(w *strings.Builder, name string, out types.Type, given map[typeID]string, format string) {
	if format == "tree" {
		fmt.Fprintf(w, "%s\n", name)
		r.node(w, child{t: out}, "", true, given, map[typeID]bool{})
		w.WriteString("\n")
		return
	}
	g := &graph{ids: map[nodeKey]int{}}
	g.walk(r, child{t: out}, given)
	g.write(w, name, format)
}

func (r *resolver) node(w *strings.Builder, c child, prefix string, last bool, given map[typeID]string, seen map[typeID]bool) {
	branch, next := "├── ", "│   "
	if last {
		branch, next = "└── ", "    "
	}
	k := r.f.types.key(c.t)
	b := c.b
	if b == nil {
		if p, ok := given[k]; ok {
			fmt.Fprintf(w, "%s%s%s [parameter %s]\n", prefix, branch, r.f.typ(c.t), p)
			return
		}
		b, _ = r.lookup(c.t)
	}
	label := r.label(c, b, given)
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

// label is "type [provider]" for a node.
func (r *resolver) label(c child, b *binding, given map[typeID]string) string {
	if p, ok := given[r.f.types.key(c.t)]; ok && c.b == nil {
		return fmt.Sprintf("%s [parameter %s]", r.f.typ(c.t), p)
	}
	return fmt.Sprintf("%s [%s]", r.f.typ(c.t), r.provider(b))
}

// nodeKey separates providers returning the same type inside Many while
// merging calls to the same function across Many and standalone dependencies.
type nodeKey struct {
	typeID   typeID
	provider providerKey
	value    *binding
}

// graph is a dependency graph for mermaid, dot and JSON.
type graph struct {
	ids    map[nodeKey]int
	labels []string // by node id index
	edges  [][2]int
	nodes  []graphNode
}

func (g *graph) walk(r *resolver, c child, given map[typeID]string) int {
	k := r.f.types.key(c.t)
	b := c.b
	if b == nil {
		if _, isParam := given[k]; !isParam {
			b, _ = r.lookup(c.t)
		}
	}
	key := nodeKey{typeID: k}
	if b != nil {
		if b.fn != nil {
			key.provider = r.f.types.provider(b.fn, b.args)
		}
	}
	if c.b != nil && b.fn == nil {
		key.value = b
	}
	if id, ok := g.ids[key]; ok {
		return id
	}
	n := len(g.labels)
	g.ids[key] = n
	g.labels = append(g.labels, r.label(c, b, given))
	node := graphNode{ID: fmt.Sprintf("n%d", n), Type: r.f.typ(c.t)}
	if parameter, ok := given[k]; ok && c.b == nil {
		node.Parameter = parameter
	} else {
		node.Provider = r.provider(b)
		if b.position.IsValid() {
			node.Position = b.position.String()
		}
	}
	g.nodes = append(g.nodes, node)
	if _, isParam := given[k]; isParam && c.b == nil {
		return n
	}
	for _, kid := range children(b) {
		g.edges = append(g.edges, [2]int{n, g.walk(r, kid, given)})
	}
	return n
}

func (g *graph) write(w *strings.Builder, name, format string) {
	esc := func(s string) string { return strings.ReplaceAll(s, `"`, `'`) }
	if format == "mermaid" {
		fmt.Fprintf(w, "%%%% %s\ngraph TD\n", name)
		for i, l := range g.labels {
			fmt.Fprintf(w, "  n%d[\"%s\"]\n", i, esc(l))
		}
		for _, e := range g.edges {
			fmt.Fprintf(w, "  n%d --> n%d\n", e[0], e[1])
		}
		w.WriteString("\n")
		return
	}
	fmt.Fprintf(w, "digraph %q {\n", name)
	for i, l := range g.labels {
		fmt.Fprintf(w, "  n%d [label=%q];\n", i, l)
	}
	for _, e := range g.edges {
		fmt.Fprintf(w, "  n%d -> n%d;\n", e[0], e[1])
	}
	w.WriteString("}\n\n")
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
	case b.fn != nil:
		return r.f.callName(b)
	case b.field != "":
		return "field " + b.field
	case b.target != nil:
		return "Bind"
	}
	return b.describe()
}
