package main

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"path/filepath"
	"strings"
	"time"
)

type stamp struct {
	mod  time.Time
	size int64
}

// snapshot lists the hand-written Go files under root. Generated files are
// left out so that writing them does not retrigger the watcher.
func snapshot(root string) (map[string]stamp, error) {
	out := map[string]stamp{}
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		switch {
		case err != nil:
			// Files disappearing during a poll are normal deletions, not partial
			// filesystem failures. A missing root is still an error.
			if path != root && errors.Is(err, fs.ErrNotExist) {
				return nil
			}
			return fmt.Errorf("inspect watch entry %s: %w", path, err)
		case d.IsDir():
			if n := d.Name(); path != root && (n == ".git" || n == "vendor" || n == "testdata" || strings.HasPrefix(n, ".")) {
				return fs.SkipDir
			}
		case strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "_gen.go"):
			info, err := d.Info()
			if errors.Is(err, fs.ErrNotExist) {
				return nil
			}
			if err != nil {
				return fmt.Errorf("read watch entry metadata %s: %w", path, err)
			}
			out[path] = stamp{info.ModTime(), info.Size()}
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("snapshot watch directory %s: %w", root, err)
	}
	return out, nil
}

// watch calls fn once, then again each time a Go file under root changes,
// until ctx is done or filesystem inspection fails. It polls, so it needs no extra dependency.
func watch(ctx context.Context, root string, every time.Duration, fn func()) error {
	last, err := snapshot(root)
	if err != nil {
		return err
	}
	fn()
	tick := time.NewTicker(every)
	defer tick.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-tick.C:
			cur, err := snapshot(root)
			if err != nil {
				return err
			}
			if !maps.Equal(cur, last) {
				last = cur
				fn()
			}
		}
	}
}
