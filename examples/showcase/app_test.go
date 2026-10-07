package main

import "testing"

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
