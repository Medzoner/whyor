//go:build whyor

package x

import "github.com/Medzoner/whyor"

type DB struct{}
type Store interface{ Read() }
type Postgres struct{}

func (*Postgres) Read() {}

type App struct{}
type Server struct{}

func NewPostgres(*DB) *Postgres { return nil }
func NewApp(Store) *App         { return nil }
func NewServer(*App) *Server    { return nil }

func Init() *Server {
	panic(whyor.Build(NewServer, NewApp, NewPostgres, whyor.Bind[Store, *Postgres]()))
}
