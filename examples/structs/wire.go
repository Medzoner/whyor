//go:build whyor

package structs

import "github.com/Medzoner/whyor"

func InitServer() *Server {
	panic(whyor.Build(
		NewConfig, NewLogger,
		whyor.Struct[Deps](),
		whyor.FieldsOf[*Config]("Addr", "Port"),
		NewServer,
	))
}
