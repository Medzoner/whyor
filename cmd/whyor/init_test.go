package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type failingSkeletonFile struct {
	writeErr error
	closeErr error
	closed   int
}

func (f *failingSkeletonFile) Write(p []byte) (int, error) {
	if f.writeErr != nil {
		return 0, f.writeErr
	}
	return len(p), nil
}

func (f *failingSkeletonFile) Close() error {
	f.closed++
	return f.closeErr
}

func TestWriteSkeletonPreservesWriteAndCloseErrors(t *testing.T) {
	writeErr := errors.New("write failed")
	closeErr := errors.New("close failed")
	for _, tc := range []struct {
		name               string
		writeErr, closeErr error
	}{
		{"success", nil, nil}, {"write", writeErr, nil}, {"close", nil, closeErr}, {"both", writeErr, closeErr},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := &failingSkeletonFile{writeErr: tc.writeErr, closeErr: tc.closeErr}
			err := writeSkeleton(f, "app")
			if f.closed != 1 {
				t.Fatalf("closed %d times", f.closed)
			}
			if tc.writeErr != nil && !errors.Is(err, tc.writeErr) {
				t.Fatalf("write error lost: %v", err)
			}
			if tc.writeErr != nil && !strings.Contains(err.Error(), "write skeleton") {
				t.Fatalf("missing write context: %v", err)
			}
			if tc.closeErr != nil && !errors.Is(err, tc.closeErr) {
				t.Fatalf("close error lost: %v", err)
			}
			if tc.closeErr != nil && !strings.Contains(err.Error(), "close skeleton") {
				t.Fatalf("missing close context: %v", err)
			}
			if tc.writeErr == nil && tc.closeErr == nil && err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestInitExistingFilePreservesCause(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "app")
	if _, err := initFile(dir); err != nil {
		t.Fatal(err)
	}
	_, err := initFile(dir)
	if !errors.Is(err, os.ErrExist) || !strings.Contains(err.Error(), "create skeleton") {
		t.Fatalf("missing existing-file cause/context: %v", err)
	}
	var pathErr *os.PathError
	if !errors.As(err, &pathErr) {
		t.Fatalf("PathError was lost: %v", err)
	}
}

func TestInitDirectoryCreationPreservesCause(t *testing.T) {
	file := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(file, []byte("not a directory"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := initFile(filepath.Join(file, "app"))
	var pathErr *os.PathError
	if !errors.As(err, &pathErr) || !strings.Contains(err.Error(), "create package directory") {
		t.Fatalf("missing directory cause/context: %v", err)
	}
}

func TestInitRejectsUnreadablePackageClause(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "app.go"), []byte("package\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := initFile(dir); err == nil || !strings.Contains(err.Error(), "read package clause") {
		t.Fatalf("expected package error: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "wire.go")); !os.IsNotExist(err) {
		t.Fatalf("skeleton created after validation failure: %v", err)
	}
}

func TestInitDirectoryContainingGlobCharacters(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "[shop]")
	if err := os.Mkdir(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "app.go"), []byte("package shop\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	path, err := initFile(dir)
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "package shop") {
		t.Fatalf("wrong package: %s", data)
	}
}

func TestInitRejectsInvalidDirectoryPackageName(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "123")
	if _, err := initFile(dir); err == nil || !strings.Contains(err.Error(), "valid Go package name") {
		t.Fatalf("expected invalid-name error: %v", err)
	}
}
