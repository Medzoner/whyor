package gen

import (
	"fmt"
	"go/types"
)

// Graph JSON is versioned separately from the Go API. IDs are local to each
// injector; edges point from a consumer to a dependency. Empty graphs is [].
type graphDocument struct {
	Version int             `json:"version"`
	Graphs  []injectorGraph `json:"graphs"`
}

type injectorGraph struct {
	Package  string      `json:"package"`
	Injector string      `json:"injector"`
	Position string      `json:"position"`
	Root     string      `json:"root"`
	Nodes    []graphNode `json:"nodes"`
	Edges    []graphEdge `json:"edges"`
}

type graphNode struct {
	ID        string `json:"id"`
	Type      string `json:"type"`
	Provider  string `json:"provider,omitempty"`
	Parameter string `json:"parameter,omitempty"`
	Position  string `json:"position,omitempty"`
}

type graphEdge struct {
	From string `json:"from"`
	To   string `json:"to"`
}

func (r *resolver) jsonGraph(name, position string, out types.Type, given map[typeID]string) injectorGraph {
	g := &graph{ids: map[nodeKey]int{}}
	g.walk(r, child{t: out}, given)
	edges := make([]graphEdge, len(g.edges))
	for i, e := range g.edges {
		edges[i] = graphEdge{From: fmt.Sprintf("n%d", e[0]), To: fmt.Sprintf("n%d", e[1])}
	}
	return injectorGraph{
		Package: r.f.pkg.PkgPath, Injector: name, Position: position,
		Root: "n0", Nodes: g.nodes, Edges: edges,
	}
}
