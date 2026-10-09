package main

import (
	"errors"
	"fmt"
	"go/parser"
	"go/token"
	"io"
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
		return "", fmt.Errorf("create package directory %s: %w", dir, err)
	}
	name, err := packageName(dir)
	if err != nil {
		return "", err
	}
	path := filepath.Join(dir, "wire.go")
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return "", fmt.Errorf("create skeleton %s: %w", path, err)
	}
	if err := writeSkeleton(f, name); err != nil {
		removeErr := os.Remove(path)
		if errors.Is(removeErr, os.ErrNotExist) {
			removeErr = nil
		}
		if removeErr != nil {
			removeErr = fmt.Errorf("remove incomplete skeleton %s: %w", path, removeErr)
		}
		return "", fmt.Errorf("create %s: %w", path, errors.Join(err, removeErr))
	}
	return path, nil
}

func writeSkeleton(w io.WriteCloser, name string) error {
	_, writeErr := fmt.Fprintf(w, skeleton, name)
	if writeErr != nil {
		writeErr = fmt.Errorf("write skeleton: %w", writeErr)
	}
	closeErr := w.Close()
	if closeErr != nil {
		closeErr = fmt.Errorf("close skeleton: %w", closeErr)
	}
	return errors.Join(writeErr, closeErr)
}

// packageName is the package of the Go files already in dir, else the directory name.
func packageName(dir string) (string, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return "", fmt.Errorf("read package directory %s: %w", dir, err)
	}
	for _, entry := range files {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") {
			continue
		}
		name := filepath.Join(dir, entry.Name())
		if strings.HasSuffix(name, "_test.go") || strings.HasSuffix(name, "_gen.go") {
			continue
		}
		f, err := parser.ParseFile(token.NewFileSet(), name, nil, parser.PackageClauseOnly)
		if err != nil {
			return "", fmt.Errorf("read package clause in %s: %w", name, err)
		}
		return f.Name.Name, nil
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		return "", fmt.Errorf("resolve package directory %s: %w", dir, err)
	}
	name := strings.NewReplacer("-", "_", ".", "_").Replace(filepath.Base(abs))
	if name == "_" || !token.IsIdentifier(name) || token.IsKeyword(name) {
		return "", fmt.Errorf("directory name %q is not a valid Go package name", name)
	}
	return name, nil
}
