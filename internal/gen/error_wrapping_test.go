package gen

import (
	"errors"
	"strings"
	"testing"

	"golang.org/x/tools/go/packages"
)

func TestPackageLoadDiagnosticsPreserveCause(t *testing.T) {
	_, err := Run("../..", []string{"./internal/gen/testdata/genericconstraint"}, false)
	var packageErr packages.Error
	if !errors.As(err, &packageErr) || !strings.Contains(err.Error(), "type-check package") {
		t.Fatalf("missing package diagnostic cause/context: %v", err)
	}
	if packageErr.Msg == "" {
		t.Fatal("empty package diagnostic")
	}
}
