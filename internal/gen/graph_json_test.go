package gen

import (
	"encoding/json"
	"strings"
	"testing"
)

func readGraphs(t *testing.T, pattern string) graphDocument {
	t.Helper()
	text, err := Show("../..", []string{pattern}, "json")
	if err != nil {
		t.Fatal(err)
	}
	var document graphDocument
	if err := json.Unmarshal([]byte(text), &document); err != nil {
		t.Fatal(err)
	}
	if document.Version != 1 {
		t.Fatalf("unexpected schema version: %d", document.Version)
	}
	for _, g := range document.Graphs {
		if g.Package == "" || g.Injector == "" || g.Position == "" {
			t.Fatalf("missing injector metadata: %+v", g)
		}
		ids := map[string]bool{}
		for _, n := range g.Nodes {
			if ids[n.ID] || n.Type == "" || (n.Provider == "" && n.Parameter == "") {
				t.Fatalf("invalid node: %+v", n)
			}
			ids[n.ID] = true
		}
		if !ids[g.Root] {
			t.Fatal("root is not a node")
		}
		for _, e := range g.Edges {
			if !ids[e.From] || !ids[e.To] {
				t.Fatalf("invalid edge: %+v", e)
			}
		}
	}
	return document
}

func TestGraphJSONMultiplePackages(t *testing.T) {
	document := readGraphs(t, "./examples/...")
	if len(document.Graphs) < 8 {
		t.Fatalf("missing injectors: %d", len(document.Graphs))
	}
	for _, g := range document.Graphs {
		if g.Injector == "InitApp" {
			if len(g.Nodes) != 5 || len(g.Edges) != 4 {
				t.Fatalf("unexpected basic graph: %+v", g)
			}
			last := g.Nodes[4]
			if last.Parameter != "dsn" || last.Provider != "" {
				t.Fatalf("parameter represented as a provider: %+v", last)
			}
		}
	}
}

func TestGraphJSONManySharedProvider(t *testing.T) {
	document := readGraphs(t, "./internal/gen/testdata/graphmany")
	g := document.Graphs[0]
	providers := map[string]int{}
	for _, n := range g.Nodes {
		providers[n.Provider]++
	}
	if providers["NewA"] != 1 || providers["NewB"] != 1 || len(g.Nodes) != 4 || len(g.Edges) != 4 {
		t.Fatalf("same-type Many providers must be distinct but shared calls merged: %+v", g)
	}
}

func TestGraphJSONEmptyAndInvalid(t *testing.T) {
	text, err := Show("../..", []string{"."}, "json")
	if err != nil || !strings.Contains(text, `"graphs": []`) {
		t.Fatalf("expected an empty array: %s %v", text, err)
	}
	text, err = Show("../..", []string{"./internal/gen/testdata/deepmissing"}, "json")
	if err == nil || text != "" {
		t.Fatalf("invalid graphs must not produce partial JSON: %s %v", text, err)
	}
}
