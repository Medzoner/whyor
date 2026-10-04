package gen

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestOutputPreservesUnownedFiles(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, genFile)
	old := []byte("package manual\n")
	if err := os.WriteFile(path, old, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := (outputFile{path: path, source: []byte(generatedHeader + "\npackage generated\n")}).write(); err == nil {
		t.Fatal("must refuse a file without the ownership header")
	}
	if data, err := os.ReadFile(path); err != nil || !bytes.Equal(data, old) {
		t.Fatal("manual file was changed", err)
	}
}

func TestOutputRejectsSymlinks(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("creating symlinks requires extra privileges")
	}
	dir := t.TempDir()
	target := filepath.Join(dir, "target.go")
	if err := os.WriteFile(target, []byte(generatedHeader+"\npackage x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, genFile)
	if err := os.Symlink(target, path); err != nil {
		t.Fatal(err)
	}
	if _, err := readOutput(path); err == nil {
		t.Fatal("must reject symlinks even when target has a generated header")
	}
}

func TestOutputReplacement(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, genFile)
	for _, source := range []string{generatedHeader + "\npackage x\n", generatedHeader + "\npackage x\nvar V = 1\n"} {
		if err := (outputFile{path: path, source: []byte(source)}).write(); err != nil {
			t.Fatal(err)
		}
		if data, err := os.ReadFile(path); err != nil || string(data) != source {
			t.Fatal("incorrect replacement", err)
		}
	}
	files, err := filepath.Glob(filepath.Join(dir, ".whyor-*.tmp"))
	if err != nil || len(files) != 0 {
		t.Fatalf("temporary files left behind: %v %v", files, err)
	}
}

func TestGenerationValidatesAllPackagesBeforeWriting(t *testing.T) {
	dir, err := os.MkdirTemp("testdata", "validation-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })
	for _, name := range []string{"a", "b"} {
		if err := os.Mkdir(filepath.Join(dir, name), 0o755); err != nil {
			t.Fatal(err)
		}
		provider := "package x\ntype A struct{}\nfunc New() *A { return nil }\n"
		if name == "b" {
			provider = "package x\ntype A struct{}\ntype B struct{}\nfunc New(*B) *A { return nil }\n"
		}
		if err := os.WriteFile(filepath.Join(dir, name, "app.go"), []byte(provider), 0o644); err != nil {
			t.Fatal(err)
		}
		wire := "//go:build whyor\n\npackage x\nimport \"github.com/Medzoner/whyor\"\nfunc Init() *A { panic(whyor.Build(New)) }\n"
		if err := os.WriteFile(filepath.Join(dir, name, "wire.go"), []byte(wire), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	base := "./internal/gen/testdata/" + filepath.Base(dir)
	if _, err := Run("../..", []string{base + "/a", base + "/b"}, true); err == nil {
		t.Fatal("missing dependency must fail generation")
	}
	for _, name := range []string{"a", "b"} {
		if _, err := os.Stat(filepath.Join(dir, name, genFile)); !os.IsNotExist(err) {
			t.Fatalf("output written before all graphs were validated: %s %v", name, err)
		}
	}
}
