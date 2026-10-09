package main

import (
	"errors"
	"io"
	"path/filepath"
	"strings"
	"testing"
)

type failedOutput struct{ err error }

func (w failedOutput) Write([]byte) (int, error) { return 0, w.err }

func TestOutputErrorsPreserveCause(t *testing.T) {
	cause := errors.New("closed output")
	for _, tc := range []struct {
		name    string
		run     func() error
		context string
	}{
		{"lines", func() error { return writeLines(failedOutput{cause}, []string{"path"}) }, "write command output"},
		{"diagnostic", func() error { return reportError(failedOutput{cause}, errors.New("generation failed")) }, "write command diagnostic"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.run()
			if !errors.Is(err, cause) || !strings.Contains(err.Error(), tc.context) {
				t.Fatalf("output error was lost: %v", err)
			}
		})
	}
}

func TestInitOutputFailureIsReturned(t *testing.T) {
	cause := errors.New("stdout failed")
	err := runWithWriters([]string{"init", filepath.Join(t.TempDir(), "app")}, failedOutput{cause}, io.Discard)
	if !errors.Is(err, cause) {
		t.Fatalf("init output failure was lost: %v", err)
	}
}

func TestShowOutputFailureIsReturned(t *testing.T) {
	cause := errors.New("stdout failed")
	err := runWithWriters([]string{"show", "../../examples/basic"}, failedOutput{cause}, io.Discard)
	if !errors.Is(err, cause) || !strings.Contains(err.Error(), "write dependency graph") {
		t.Fatalf("graph output failure was lost: %v", err)
	}
}

func TestWatchStartupOutputFailureIsReturned(t *testing.T) {
	cause := errors.New("stderr failed")
	err := runWatch([]string{"."}, io.Discard, failedOutput{cause})
	if !errors.Is(err, cause) || !strings.Contains(err.Error(), "write watch startup diagnostic") {
		t.Fatalf("startup failure was lost: %v", err)
	}
}

func TestDiagnosticFailureKeepsOriginalError(t *testing.T) {
	commandErr := errors.New("package loading failed")
	outputErr := errors.New("stderr failed")
	err := reportError(failedOutput{outputErr}, commandErr)
	if !errors.Is(err, commandErr) || !errors.Is(err, outputErr) {
		t.Fatalf("original or reporting error lost: %v", err)
	}
}
