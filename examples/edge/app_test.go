package edge

import (
	"slices"
	"testing"
	"time"
)

func TestInitOut(t *testing.T) {
	Events, Fail = nil, false
	out, cleanup, err := InitOut()
	if err != nil || out.Timeout != 5*time.Second || out.Name != "edge" || out.A == nil || out.B == nil {
		t.Fatalf("out=%+v err=%v", out, err)
	}
	cleanup()
	if !slices.Equal(Events, []string{"r3", "r2", "r1"}) {
		t.Fatalf("cleanup order %v", Events)
	}
}

func TestInitOutFailureCleansUp(t *testing.T) {
	Events, Fail = nil, true
	out, cleanup, err := InitOut()
	if err == nil || cleanup != nil || out != (Out{}) {
		t.Fatalf("out=%+v cleanup=%v err=%v", out, cleanup != nil, err)
	}
	if !slices.Equal(Events, []string{"r2", "r1"}) {
		t.Fatalf("events %v", Events)
	}
}
