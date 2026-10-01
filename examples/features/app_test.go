package features

import "testing"

func TestInitServer(t *testing.T) {
	var log []string
	s, cleanup := InitServer(&log)
	if len(s.Router.Handlers) != 2 || s.H.Name() != "users" {
		t.Fatalf("unexpected server: %+v", s)
	}
	cleanup()
	if len(log) != 1 || log[0] != "cache closed" {
		t.Fatalf("log=%v", log)
	}
}
