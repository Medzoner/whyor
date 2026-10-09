package main

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestSnapshotMissingRoot(t *testing.T) {
	root := filepath.Join(t.TempDir(), "missing")
	files, err := snapshot(root)
	if files != nil || !errors.Is(err, fs.ErrNotExist) || !strings.Contains(err.Error(), "snapshot watch directory") {
		t.Fatalf("missing root error was lost: %v %v", files, err)
	}
	var pathErr *os.PathError
	if !errors.As(err, &pathErr) {
		t.Fatalf("PathError was lost: %v", err)
	}
}

func TestWatchFailsBeforeGenerationOnSnapshotError(t *testing.T) {
	called := false
	err := watch(context.Background(), filepath.Join(t.TempDir(), "missing"), time.Millisecond, func() { called = true })
	if called || !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("watch used a failed snapshot: called=%v error=%v", called, err)
	}
}

func TestWatchPropagatesLaterFilesystemFailure(t *testing.T) {
	root := filepath.Join(t.TempDir(), "watch")
	if err := os.Mkdir(root, 0o755); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	err := watch(ctx, root, time.Millisecond, func() {
		if err := os.Remove(root); err != nil {
			t.Fatal(err)
		}
	})
	if !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("later snapshot error was lost: %v", err)
	}
}
