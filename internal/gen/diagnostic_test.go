package gen

import (
	"errors"
	"strings"
	"testing"
)

func TestMissingDependencyPath(t *testing.T) {
	_, err := Run("../..", []string{"./internal/gen/testdata/deepmissing"}, false)
	var diagnostic *dependencyError
	if !errors.As(err, &diagnostic) {
		t.Fatalf("expected dependency diagnostic, got %v", err)
	}
	if len(diagnostic.path) != 5 {
		t.Fatalf("incomplete path: %v", diagnostic.path)
	}
	for i, want := range []string{"*Server [NewServer]", "*App [NewApp]", "Store [Bind]", "*Postgres [NewPostgres]", "*DB"} {
		if !strings.HasPrefix(diagnostic.path[i], want) {
			t.Fatalf("step %d: got %q, want prefix %q", i, diagnostic.path[i], want)
		}
	}
	if !strings.Contains(err.Error(), "x.go:") || !strings.Contains(err.Error(), "dependency path:") {
		t.Fatalf("missing position or path in message: %v", err)
	}
}

func TestAmbiguousAutoBind(t *testing.T) {
	_, err := Run("../..", []string{"./internal/gen/testdata/ambiguous"}, false)
	if err == nil || !strings.Contains(err.Error(), "ambiguous AutoBind for Store: *First, *Second") {
		t.Fatalf("expected explicit ambiguity, got %v", err)
	}
	if strings.Contains(err.Error(), "no provider") {
		t.Fatalf("ambiguity must not be reported as a missing provider: %v", err)
	}
}
