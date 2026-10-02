package main

import (
	"context"
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
func snapshot(root string) map[string]stamp {
	out := map[string]stamp{}
	filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		switch {
		case err != nil:
			return nil
		case d.IsDir():
			if n := d.Name(); path != root && (n == ".git" || n == "vendor" || n == "testdata" || strings.HasPrefix(n, ".")) {
				return fs.SkipDir
			}
		case strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "_gen.go"):
			if info, err := d.Info(); err == nil {
				out[path] = stamp{info.ModTime(), info.Size()}
			}
		}
		return nil
	})
	return out
}

// watch calls fn once, then again each time a Go file under root changes,
// until ctx is done. It polls, so it needs no extra dependency.
func watch(ctx context.Context, root string, every time.Duration, fn func()) {
	last := snapshot(root)
	fn()
	tick := time.NewTicker(every)
	defer tick.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
			if cur := snapshot(root); !maps.Equal(cur, last) {
				last = cur
				fn()
			}
		}
	}
}
