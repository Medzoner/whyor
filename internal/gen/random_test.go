package gen

import (
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// shapes are the result lists a provider may have, with a matching return statement.
var shapes = []struct{ ret, body string }{
	{"*T%d", "nil"},
	{"(*T%d, error)", "nil, nil"},
	{"(*T%d, func())", "nil, func() {}"},
	{"(*T%d, func(), error)", "nil, func() {}, nil"},
}

// TestRandomGraphs generates random provider graphs and checks that the output
// compiles and calls every reachable provider exactly once, and no other.
func TestRandomGraphs(t *testing.T) {
	for seed := int64(1); seed <= 25; seed++ {
		t.Run(fmt.Sprint(seed), func(t *testing.T) { checkRandomGraph(t, rand.New(rand.NewSource(seed))) })
	}
}

func checkRandomGraph(t *testing.T, rnd *rand.Rand) {
	n := 3 + rnd.Intn(10)
	deps := make([][]int, n)
	var providers strings.Builder
	for i := range n {
		var params []string
		for j := range i {
			if rnd.Float64() < 0.4 || (i == n-1 && j == i-1) {
				deps[i] = append(deps[i], j)
				params = append(params, fmt.Sprintf("*T%d", j))
			}
		}
		shape := shapes[rnd.Intn(len(shapes))]
		fmt.Fprintf(&providers, "type T%d struct{}\n\nfunc NewT%d(%s) %s { return %s }\n\n",
			i, i, strings.Join(params, ", "), fmt.Sprintf(shape.ret, i), shape.body)
	}

	order := rnd.Perm(n)
	names := make([]string, n)
	for i, o := range order {
		names[i] = fmt.Sprintf("NewT%d", o)
	}
	wire := fmt.Sprintf("//go:build whyor\n\npackage x\n\nimport \"github.com/Medzoner/whyor\"\n\n"+
		"func Init() (*T%d, func(), error) { panic(whyor.Build(%s)) }\n", n-1, strings.Join(names, ", "))

	dir, err := os.MkdirTemp("testdata", "rand")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(dir); err != nil {
			t.Errorf("remove fixture: %v", err)
		}
	})
	write := func(name, src string) {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(src), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("providers.go", "package x\n\n"+providers.String())
	write("wire.go", wire)

	pkg := "./internal/gen/testdata/" + filepath.Base(dir)
	if _, err := Run("../..", []string{pkg}, true); err != nil {
		t.Fatal(err)
	}
	out, err := os.ReadFile(filepath.Join(dir, genFile))
	if err != nil {
		t.Fatal(err)
	}

	reach := map[int]bool{}
	var walk func(int)
	walk = func(i int) {
		if reach[i] {
			return
		}
		reach[i] = true
		for _, d := range deps[i] {
			walk(d)
		}
	}
	walk(n - 1)
	for i := range n {
		want := 0
		if reach[i] {
			want = 1
		}
		if got := strings.Count(string(out), fmt.Sprintf("NewT%d(", i)); got != want {
			t.Errorf("NewT%d called %d times, want %d\n%s", i, got, want, out)
		}
	}

	cmd := exec.Command("go", "vet", pkg)
	cmd.Dir = "../.."
	if b, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("generated code does not compile: %v\n%s\n%s", err, b, out)
	}
}
