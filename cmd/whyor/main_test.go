package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestInit(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "app.go"), []byte("package shop\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	path, err := initFile(dir)
	if err != nil {
		t.Fatal(err)
	}
	src, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
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
	src, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(src), "package my_app\n") {
		t.Fatalf("unexpected skeleton:\n%s", src)
	}
}

func TestUnknownCommand(t *testing.T) {
	if run([]string{"nope"}) == nil || run(nil) == nil {
		t.Fatal("want errors")
	}
}

func TestWatchRunsOnChange(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "a.go")
	if err := os.WriteFile(file, []byte("package a\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	runs := make(chan struct{}, 10)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- watch(ctx, dir, 10*time.Millisecond, func() error { runs <- struct{}{}; return nil }) }()
	t.Cleanup(func() {
		cancel()
		select {
		case err := <-done:
			if err != nil {
				t.Errorf("watch failed: %v", err)
			}
		case <-time.After(2 * time.Second):
			t.Error("watch did not stop after cancellation")
		}
	})

	wait := func(what string) {
		t.Helper()
		select {
		case <-runs:
		case <-time.After(2 * time.Second):
			t.Fatalf("no run: %s", what)
		}
	}
	wait("initial")

	if err := os.WriteFile(file, []byte("package a\n\nvar X = 1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	wait("after edit")

	if err := os.WriteFile(filepath.Join(dir, "whyor_gen.go"), []byte("package a\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	select {
	case <-runs:
		t.Fatal("generated file must not retrigger")
	case <-time.After(100 * time.Millisecond):
	}
}
