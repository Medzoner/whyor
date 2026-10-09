package main

import (
	"context"
	"errors"
	"fmt"
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

func TestWatchReturnsCallbackError(t *testing.T) {
	cause := errors.New("output failed")
	err := watch(context.Background(), t.TempDir(), time.Millisecond, func() error { return cause })
	if !errors.Is(err, cause) || !strings.Contains(err.Error(), "initial watch callback") {
		t.Fatalf("callback error was lost: %v", err)
	}
}

func TestWatchReturnsCallbackErrorAfterChange(t *testing.T) {
	cause := errors.New("later output failed")
	dir := t.TempDir()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	calls := 0
	err := watch(ctx, dir, time.Millisecond, func() error {
		calls++
		if calls == 1 {
			if err := os.WriteFile(filepath.Join(dir, "new.go"), []byte("package x\n"), 0o644); err != nil {
				return fmt.Errorf("create changed watch fixture: %w", err)
			}
			return nil
		}
		return cause
	})
	if calls != 2 || !errors.Is(err, cause) {
		t.Fatalf("later callback error was lost: calls=%d error=%v", calls, err)
	}
}

func TestWatchFailsBeforeGenerationOnSnapshotError(t *testing.T) {
	called := false
	err := watch(context.Background(), filepath.Join(t.TempDir(), "missing"), time.Millisecond, func() error { called = true; return nil })
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
	err := watch(ctx, root, time.Millisecond, func() error {
		if err := os.Remove(root); err != nil {
			t.Fatal(err)
		}
		return nil
	})
	if !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("later snapshot error was lost: %v", err)
	}
}
