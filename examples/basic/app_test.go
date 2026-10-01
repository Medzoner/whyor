package basic

import "testing"

func TestInitApp(t *testing.T) {
	app, cleanup, err := InitApp("db")
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	if got := app.Store.Get(1); got != "db/1" {
		t.Fatalf("got %q", got)
	}
}
