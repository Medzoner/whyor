package lifecycle

import (
	"context"
	"errors"

	"github.com/Medzoner/whyor"
)

var (
	ErrInit   = errors.New("initialization failed")
	ErrCache  = errors.New("cache close failed")
	ErrSocket = errors.New("socket close failed")
)

type Events struct {
	Order    []string
	Contexts []context.Context
	Fail     bool
}

type Legacy struct{}
type Cache struct{}
type Socket struct{ events *Events }
type App struct{}

func NewLegacy(events *Events) (*Legacy, func()) {
	return &Legacy{}, func() { events.Order = append(events.Order, "legacy") }
}

func NewCache(ctx context.Context, legacy *Legacy, events *Events) (*Cache, whyor.Cleanup) {
	return &Cache{}, func(ctx context.Context) error {
		events.Order = append(events.Order, "cache")
		events.Contexts = append(events.Contexts, ctx)
		return ErrCache
	}
}

func NewSocket(cache *Cache, events *Events) *Socket { return &Socket{events: events} }

func (s *Socket) Close() error {
	s.events.Order = append(s.events.Order, "socket")
	return ErrSocket
}

func NewApp(legacy *Legacy, cache *Cache, socket *Socket, events *Events) (*App, error) {
	if events.Fail {
		return nil, ErrInit
	}
	return &App{}, nil
}

// A context-aware cleanup does not require a context parameter at acquisition.
func NewDetached(events *Events) (*Cache, func(context.Context) error) {
	return &Cache{}, func(ctx context.Context) error {
		events.Order = append(events.Order, "detached")
		events.Contexts = append(events.Contexts, ctx)
		return ErrCache
	}
}

func NewFailed(cache *Cache) (*App, error) { return nil, ErrInit }
