package gen

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"testing"
)

func TestTypeIdentity(t *testing.T) {
	const src = `package identity
type Alias = int
type Defined int
type Box[T any] struct{ Value T }
type A = func(Alias) (struct{ Value Alias }, error)
type B = func(int) (struct{ Value int }, error)
type C = struct{ Apply func(Alias) Alias; Values []Box[Alias] }
type D = struct{ Apply func(int) int; Values []Box[int] }
type E = interface{ Read(Alias); Write(string) }
type F = interface{ Write(string); Read(int) }
type G = Box[Alias]
type H = Box[int]
type I = Box[string]
type J = map[Alias][]chan *Box[Alias]
type K = map[int][]chan *Box[int]
type L = [2]Alias
type M = [2]int
type N = [3]int
type O = struct{ Value int ` + "`json:\"a\"`" + ` }
type P = struct{ Value int ` + "`json:\"b\"`" + ` }
type Q = <-chan Alias
type R = <-chan int
type S = chan int
type Cleanup = func()
`
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "identity.go", src, 0)
	if err != nil {
		t.Fatal(err)
	}
	pkg, err := new(types.Config).Check("identity", fset, []*ast.File{f}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !isCleanup(pkg.Scope().Lookup("Cleanup").Type()) {
		t.Fatal("a function alias must be accepted as a cleanup result")
	}
	for _, tc := range []struct {
		name, a, b string
		same       bool
	}{
		{"alias", "Alias", "int", true},
		{"defined type", "Defined", "int", false},
		{"nested signature", "A", "B", true},
		{"nested struct and generics", "C", "D", true},
		{"interface method order", "E", "F", true},
		{"generic alias argument", "G", "H", true},
		{"different generic arguments", "H", "I", false},
		{"nested map slice channel pointer", "J", "K", true},
		{"array alias", "L", "M", true},
		{"array length", "M", "N", false},
		{"struct tags", "O", "P", false},
		{"channel alias", "Q", "R", true},
		{"channel direction", "R", "S", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			lookup := func(name string) types.Type {
				if obj := pkg.Scope().Lookup(name); obj != nil {
					return obj.Type()
				}
				return types.Universe.Lookup(name).Type()
			}
			a, b := lookup(tc.a), lookup(tc.b)
			if got := types.Identical(a, b); got != tc.same {
				t.Fatalf("invalid test: types.Identical = %v, want %v", got, tc.same)
			}
			var index typeIndex
			ka, kb := index.key(a), index.key(b)
			if (ka == kb) != tc.same || index.key(a) != ka || index.key(b) != kb {
				t.Fatalf("unstable or incorrect identities: %d, %d", ka, kb)
			}
		})
	}
}

func TestIdenticalPrintedNamesDoNotImplyIdenticalTypes(t *testing.T) {
	// Unexported fields belong to their declaring package even though printing
	// the anonymous struct does not include that package's path.
	a := types.NewPackage("example/a", "shared")
	b := types.NewPackage("example/b", "shared")
	makeStruct := func(pkg *types.Package) types.Type {
		field := types.NewVar(token.NoPos, pkg, "hidden", types.Typ[types.Int])
		return types.NewStruct([]*types.Var{field}, nil)
	}
	x, y := makeStruct(a), makeStruct(b)
	if types.TypeString(x, nil) != types.TypeString(y, nil) {
		t.Fatal("test requires a textual collision")
	}
	var index typeIndex
	if index.key(x) == index.key(y) {
		t.Fatal("different packages' private fields must remain different types")
	}
	if index.key(x) != index.key(makeStruct(a)) {
		t.Fatal("separately constructed identical types must share identity")
	}
}
