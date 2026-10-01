package basic

import "fmt"

type Config struct{ DSN string }

type Store interface{ Get(id int) string }

type PG struct{ cfg *Config }

func (p *PG) Get(id int) string { return fmt.Sprintf("%s/%d", p.cfg.DSN, id) }

type App struct{ Store Store }

func NewConfig(dsn string) *Config { return &Config{DSN: dsn} }

func NewPG(cfg *Config) (*PG, func(), error) {
	return &PG{cfg: cfg}, func() { fmt.Println("pg closed") }, nil
}

func NewApp(s Store) *App { return &App{Store: s} }
