package structs

import "testing"

func TestInitServer(t *testing.T) {
	if got := InitServer().Addr; got != "localhost:8" {
		t.Fatalf("got %q", got)
	}
}
