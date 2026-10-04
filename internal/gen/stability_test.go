package gen

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// Compile generated code, rather than only accepting its formatted text.
func TestGeneratedEdgeCasesCompile(t *testing.T) {
	for _, tc := range []struct {
		name, imports, providers, injector, test string
	}{
		{
			name: "typed values",
			providers: `type Label string
type Result struct { N int64; Value any; Label Label; Callback func() }
func New(n int64, value any, label Label, callback func()) Result { return Result{n, value, label, callback} }`,
			injector: `func Init() Result { panic(whyor.Build(New, whyor.Value[int64](7), whyor.Value[any](nil), whyor.Value[Label]("x"), whyor.Value[func()](nil))) }`,
			test:     `func TestInit(t *testing.T) { r := Init(); if r.N != 7 || r.Value != nil || r.Label != "x" || r.Callback != nil { t.Fatal(r) } }`,
		},
		{
			name: "local and package names",
			providers: `var v2 = "package value"
type Result struct{ Text string }
func NewText(v1 string) string { return v1 }
func New(n int, b bool, text string) (Result, func(), error) { return Result{text}, func(){}, nil }`,
			injector: `func Init(v1 string, err int, cleanup1 bool) (Result, func(), error) { panic(whyor.Build(New)) }`,
			test:     `func TestInit(t *testing.T) { r, cleanup, err := Init("input", 1, true); if err != nil || r.Text != "input" { t.Fatal(r, err) }; cleanup() }`,
		},
		{
			name: "import aliases and lifecycle names",
			imports: `"context"
"time"`,
			providers: `import("context"; "time")
type Result struct{ Duration time.Duration }
type Events struct{ Closed bool }
func New(d time.Duration, events *Events) (Result, func(context.Context) error, error) {
 return Result{d}, func(context.Context) error { events.Closed = true; return nil }, nil
}`,
			injector: `func Init(context context.Context, errors *Events) (Result, whyor.Cleanup, error) {
 panic(whyor.Build(New, whyor.Value[time.Duration](time.Second)))
}`,
			test: `import "context"
func TestInit(t *testing.T) { var e Events; r, cleanup, err := Init(context.Background(), &e); if err != nil || r.Duration != 1000000000 { t.Fatal(r, err) }; if err := cleanup(context.Background()); err != nil || !e.Closed { t.Fatal(err, e) } }`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir, err := os.MkdirTemp("testdata", "stability-")
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { os.RemoveAll(dir) })
			write := func(name, content string) {
				t.Helper()
				if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			write("providers.go", "package x\n"+tc.providers)
			write("wire.go", "//go:build whyor\n\npackage x\nimport(\"github.com/Medzoner/whyor\"\n"+tc.imports+")\n"+tc.injector)
			write("app_test.go", "package x\nimport \"testing\"\n"+tc.test)
			pattern := "./internal/gen/testdata/" + filepath.Base(dir)
			if _, err := Run("../..", []string{pattern}, true); err != nil {
				t.Fatal(err)
			}
			cmd := exec.Command("go", "test", pattern)
			cmd.Dir = "../.."
			if out, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("generated package fails: %v\n%s", err, out)
			}
		})
	}
}
