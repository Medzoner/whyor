package lifecycle

import (
	"context"
	"errors"
	"slices"
	"testing"
)

type contextKey struct{}

func TestCleanupErrorsAndContext(t *testing.T) {
	var events Events
	_, cleanup, err := InitApp(context.Background(), &events)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.WithValue(context.Background(), contextKey{}, "shutdown"))
	cancel()
	err = cleanup(ctx)
	if !errors.Is(err, ErrCache) || !errors.Is(err, ErrSocket) {
		t.Fatalf("cleanup errors were lost: %v", err)
	}
	if !slices.Equal(events.Order, []string{"socket", "cache", "legacy"}) {
		t.Fatalf("cleanup stopped early or ran out of order: %v", events.Order)
	}
	if len(events.Contexts) != 1 || events.Contexts[0] != ctx || events.Contexts[0].Err() != context.Canceled {
		t.Fatal("shutdown context must be passed through unchanged")
	}
}

func TestRollbackPreservesErrorsAndContextValues(t *testing.T) {
	events := Events{Fail: true}
	ctx, cancel := context.WithCancel(context.WithValue(context.Background(), contextKey{}, "acquisition"))
	cancel()
	app, cleanup, err := InitApp(ctx, &events)
	if app != nil || cleanup != nil || !errors.Is(err, ErrInit) || !errors.Is(err, ErrCache) || !errors.Is(err, ErrSocket) {
		t.Fatalf("rollback lost an error: %v", err)
	}
	if !slices.Equal(events.Order, []string{"socket", "cache", "legacy"}) || len(events.Contexts) != 1 {
		t.Fatalf("wrong rollback: %+v", events)
	}
	rollback := events.Contexts[0]
	if rollback.Err() != nil || rollback.Value(contextKey{}) != "acquisition" {
		t.Fatal("rollback must preserve values without acquisition cancellation")
	}
}

func TestRollbackWithoutAcquisitionContext(t *testing.T) {
	var events Events
	app, cleanup, err := InitDetached(&events)
	if app != nil || cleanup != nil || !errors.Is(err, ErrInit) || !errors.Is(err, ErrCache) || len(events.Contexts) != 1 {
		t.Fatalf("wrong rollback: err=%v events=%+v", err, events)
	}
	if events.Contexts[0].Err() != nil {
		t.Fatal("rollback without an injector context must use Background")
	}
}

func TestLegacyAndEmptyCleanups(t *testing.T) {
	var events Events
	_, cleanup := InitLegacy(&events)
	if err := cleanup(context.Background()); err != nil || !slices.Equal(events.Order, []string{"legacy"}) {
		t.Fatalf("legacy cleanup must be adapted: %v %+v", err, events)
	}
	_, cleanup = InitEmpty()
	if err := cleanup(context.Background()); err != nil {
		t.Fatal(err)
	}
}
