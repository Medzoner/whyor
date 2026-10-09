package basic

import "testing"

func TestInitApp(t *testing.T) {
	app, cleanup, err := InitApp("db")
	if err != nil {
		t.Fatal(err)
	}
	if got := app.Store.Get(1); got != "db/1" {
		t.Fatalf("got %q", got)
	}
	cleanup()
	if !app.Store.(*PG).closed {
		t.Fatal("cleanup did not close the example store")
	}
}
