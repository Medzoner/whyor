package main

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func TestInitApp(t *testing.T) {
	app, cleanup, err := InitApp("Go")
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	if got := app.Store.Message(); got != "Hello, Go" {
		t.Fatalf("got %q", got)
	}
}

type brokenOutput struct{ err error }

func (w brokenOutput) Write([]byte) (int, error) { return 0, w.err }

func TestApplicationOutputIsChecked(t *testing.T) {
	var output bytes.Buffer
	if err := runApplication(&output); err != nil {
		t.Fatal(err)
	}
	if output.String() != "Hello, Go\ncleanup: memory store\n" {
		t.Fatalf("unexpected output: %q", output.String())
	}
	cause := errors.New("output unavailable")
	err := runApplication(brokenOutput{cause})
	if !errors.Is(err, cause) || !strings.Contains(err.Error(), "print example greeting") || !strings.Contains(err.Error(), "print example cleanup") {
		t.Fatalf("output errors were lost: %v", err)
	}
}
