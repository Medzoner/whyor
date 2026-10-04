package gen

import (
	"os"
	"strings"
	"testing"
)

func TestExampleUpToDate(t *testing.T) {
	files, err := Run("../..", []string{"./examples/..."}, false)
	if err != nil || len(files) != 0 {
		t.Fatalf("files=%v err=%v", files, err)
	}
}

func TestAliasIdentityAcrossGenerationAndGraphs(t *testing.T) {
	const pkg = "./examples/typeidentity"
	if changed, err := Run("../..", []string{pkg}, false); err != nil || len(changed) != 0 {
		t.Fatalf("generated example is stale: %v %v", changed, err)
	}
	src, err := os.ReadFile("../../examples/typeidentity/whyor_gen.go")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(string(src), "NewTransform()") != 1 {
		t.Fatal("alias-equivalent requests must share a single provider call")
	}
	for _, format := range []string{"tree", "dot", "mermaid"} {
		graph, err := Show("../..", []string{pkg}, format)
		if err != nil {
			t.Fatal(err)
		}
		if format == "tree" {
			if !strings.Contains(graph, "[NewTransform] (*)") {
				t.Fatal("tree must recognize the repeated alias-equivalent dependency")
			}
		} else if strings.Count(graph, "[NewTransform]") != 1 {
			t.Fatalf("%s must have one node for the shared function type", format)
		}
	}
}

func TestErrors(t *testing.T) {
	for name, want := range map[string]string{
		"missing":           "no provider for",
		"cycle":             "cycle",
		"err":               "returns an error",
		"closer":            "injector returns no cleanup",
		"many":              "not assignable",
		"hint":              "hint: *A implements I: add whyor.Bind[I, *A]()",
		"hintptr":           "a provider for A exists, but not for *A",
		"notstruct":         "is not a struct",
		"structfield":       "has no exported field Nope",
		"dup":               "(NewA, NewA2)",
		"identityduplicate": "multiple providers for",
		"identitycycle":     "dependency cycle",
		"identityparams":    "multiple parameters of type",
		"identitygeneric":   "no provider for *Box[string]",
		"lifecyclerejected": "injector must return whyor.Cleanup",
		"duplicatedfield":   "listed more than once",
		"malformedbuild":    "Build must be the sole",
		"untaggedinjector":  "injector file must be excluded",
		"predeclaredparam":  "shadows a predeclared",
	} {
		t.Run(name, func(t *testing.T) {
			_, err := Run("../..", []string{"./internal/gen/testdata/" + name}, false)
			if err == nil || !strings.Contains(err.Error(), want) {
				t.Fatalf("want %q, got %v", want, err)
			}
		})
	}
}

func TestShow(t *testing.T) {
	got, err := Show("../..", []string{"./examples/basic"}, "tree")
	if err != nil {
		t.Fatal(err)
	}
	want := `InitApp
└── *App [NewApp]
    └── Store [Bind]
        └── *PG [NewPG]
            └── *Config [NewConfig]
                └── string [parameter dsn]

`
	if got != want {
		t.Fatalf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestUnused(t *testing.T) {
	got, err := Unused("../..", []string{"./internal/gen/testdata/unusedprov"})
	if err != nil || len(got) != 1 || !strings.HasSuffix(got[0], "provider NewB is never used") {
		t.Fatalf("got=%v err=%v", got, err)
	}
	if got, err := Unused("../..", []string{"./examples/..."}); err != nil || len(got) != 0 {
		t.Fatalf("examples must have no unused provider: %v %v", got, err)
	}
}

func TestShowFormats(t *testing.T) {
	for format, want := range map[string]string{
		"mermaid": "graph TD",
		"dot":     `digraph "InitApp"`,
	} {
		got, err := Show("../..", []string{"./examples/basic"}, format)
		if err != nil || !strings.Contains(got, want) || !strings.Contains(got, "n3 -> n4") && !strings.Contains(got, "n3 --> n4") {
			t.Fatalf("%s: got %q err=%v", format, got, err)
		}
	}
	if _, err := Show("../..", []string{"./examples/basic"}, "svg"); err == nil {
		t.Fatal("unknown format must fail")
	}
}
