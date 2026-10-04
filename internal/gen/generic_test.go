package gen

import (
	"os"
	"strings"
	"testing"
)

func TestGenericInstantiationConflict(t *testing.T) {
	_, err := Run("../..", []string{"./internal/gen/testdata/genericconflict"}, false)
	if err == nil || !strings.Contains(err.Error(), "(New[int], New[string])") {
		t.Fatalf("different arguments with the same result must conflict: %v", err)
	}
}

func TestGenericInstantiationConstraints(t *testing.T) {
	_, err := Run("../..", []string{"./internal/gen/testdata/genericconstraint"}, false)
	if err == nil || !strings.Contains(err.Error(), "does not satisfy") {
		t.Fatalf("Go must validate provider constraints: %v", err)
	}
}

func TestUnusedGenericInstantiation(t *testing.T) {
	list, err := Unused("../..", []string{"./internal/gen/testdata/genericunused"})
	if err != nil || len(list) != 1 || !strings.HasSuffix(list[0], "provider New[string] is never used") {
		t.Fatalf("using New[int] must not mark New[string] as used: %v %v", list, err)
	}
}

func TestGenericCallsAndGraph(t *testing.T) {
	src, err := os.ReadFile("../../examples/generics/whyor_gen.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, call := range []string{"providers.NewRepository[User](", "providers.NewRepository[Order](",
		"providers.NewLabel[User](", "providers.NewLabel[Order](", "providers.NewPair[User, Order]("} {
		if strings.Count(string(src), call) != 1 {
			t.Fatalf("expected one call of %s:\n%s", call, src)
		}
	}
	if strings.Contains(string(src), "NewLabel[Person](") {
		t.Fatal("alias-equivalent instantiations must share the original call")
	}
	document := readGraphs(t, "./examples/generics")
	providers := map[string]int{}
	for _, node := range document.Graphs[0].Nodes {
		providers[node.Provider]++
	}
	if providers["providers.NewLabel[User]"] != 1 || providers["providers.NewLabel[Order]"] != 1 {
		t.Fatalf("graph must distinguish instantiations: %v", providers)
	}
}
