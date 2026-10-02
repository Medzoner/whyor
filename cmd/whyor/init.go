package main

import (
	"errors"
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

const skeleton = `//go:build whyor

package %s

import "github.com/Medzoner/whyor"

// Providers groups the constructors of this package.
var Providers = whyor.Set(
// NewConfig,
// whyor.Bind[Store, *Postgres](),
)

// InitApp is replaced by generated code. Uncomment and adapt:
//
// func InitApp() (*App, func(), error) {
// 	panic(whyor.Build(Providers, NewApp))
// }
`

// initFile writes dir/wire.go and returns its path. It never overwrites.
func initFile(dir string) (string, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	path := filepath.Join(dir, "wire.go")
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if errors.Is(err, os.ErrExist) {
		return "", fmt.Errorf("%s already exists", path)
	}
	if err != nil {
		return "", err
	}
	defer f.Close()
	_, err = fmt.Fprintf(f, skeleton, packageName(dir))
	return path, err
}

// packageName is the package of the Go files already in dir, else the directory name.
func packageName(dir string) string {
	files, _ := filepath.Glob(filepath.Join(dir, "*.go"))
	for _, name := range files {
		if strings.HasSuffix(name, "_test.go") || strings.HasSuffix(name, "_gen.go") {
			continue
		}
		if f, err := parser.ParseFile(token.NewFileSet(), name, nil, parser.PackageClauseOnly); err == nil {
			return f.Name.Name
		}
	}
	abs, _ := filepath.Abs(dir)
	return strings.NewReplacer("-", "_", ".", "_").Replace(filepath.Base(abs))
}
