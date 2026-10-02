package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInit(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "app.go"), []byte("package shop\n"), 0o644)

	path, err := initFile(dir)
	if err != nil {
		t.Fatal(err)
	}
	src, _ := os.ReadFile(path)
	if !strings.Contains(string(src), "package shop\n") || !strings.HasPrefix(string(src), "//go:build whyor") {
		t.Fatalf("unexpected skeleton:\n%s", src)
	}
	if _, err := initFile(dir); err == nil {
		t.Fatal("second init must not overwrite")
	}
}

func TestInitUsesDirName(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "my-app")
	path, err := initFile(dir)
	if err != nil {
		t.Fatal(err)
	}
	if src, _ := os.ReadFile(path); !strings.Contains(string(src), "package my_app\n") {
		t.Fatalf("unexpected skeleton:\n%s", src)
	}
}

func TestUnknownCommand(t *testing.T) {
	if run([]string{"nope"}) == nil || run(nil) == nil {
		t.Fatal("want errors")
	}
}
