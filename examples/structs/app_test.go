package structs

import "testing"

func TestInitServer(t *testing.T) {
	server := InitServer()
	if got := server.Addr; got != "localhost:8" {
		t.Fatalf("got %q", got)
	}
}

// NewServer only receives the exported, non-skipped Deps fields; the unexported
// hidden field stays at its zero value and is never wired by whyor.
func TestUnexportedFieldIsNotWired(t *testing.T) {
	var deps Deps
	if deps.hidden != 0 {
		t.Fatalf("hidden = %d, want zero value", deps.hidden)
	}
}
