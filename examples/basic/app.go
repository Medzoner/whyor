package basic

import "fmt"

type Config struct{ DSN string }

type Store interface{ Get(id int) string }

type PG struct {
	cfg    *Config
	closed bool
}

func (p *PG) Get(id int) string { return fmt.Sprintf("%s/%d", p.cfg.DSN, id) }

type App struct{ Store Store }

func NewConfig(dsn string) *Config { return &Config{DSN: dsn} }

func NewPG(cfg *Config) (*PG, func(), error) {
	pg := &PG{cfg: cfg}
	return pg, func() { pg.closed = true }, nil
}

func NewApp(s Store) *App { return &App{Store: s} }
