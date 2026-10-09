package main

import "fmt"

type Config struct{ Name string }
type Store interface{ Message() string }
type MemoryStore struct {
	cfg    *Config
	closed bool
}
type App struct{ Store Store }

func NewConfig(name string) *Config { return &Config{Name: name} }

func NewStore(cfg *Config) (*MemoryStore, func(), error) {
	// A demonstrative cleanup: no database or external service is required.
	store := &MemoryStore{cfg: cfg}
	return store, func() { store.closed = true }, nil
}

func (s *MemoryStore) Message() string { return fmt.Sprintf("Hello, %s", s.cfg.Name) }
func NewApp(store Store) *App          { return &App{Store: store} }
