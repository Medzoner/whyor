package gen

import (
	"strings"
	"testing"
)

func TestExampleUpToDate(t *testing.T) {
	files, err := Run("../..", []string{"./examples/..."}, false)
	if err != nil || len(files) != 0 {
		t.Fatalf("files=%v err=%v", files, err)
	}
}

func TestErrors(t *testing.T) {
	for name, want := range map[string]string{
		"missing": "no provider for",
		"cycle":   "cycle",
		"dup":     "multiple providers",
		"err":     "returns an error",
		"closer":  "injector returns no cleanup",
		"many":    "not assignable",
	} {
		t.Run(name, func(t *testing.T) {
			_, err := Run("../..", []string{"./internal/gen/testdata/" + name}, false)
			if err == nil || !strings.Contains(err.Error(), want) {
				t.Fatalf("want %q, got %v", want, err)
			}
		})
	}
}
