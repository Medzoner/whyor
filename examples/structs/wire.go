//go:build whyor

package structs

import "github.com/Medzoner/whyor"

func InitServer() *Server {
	panic(whyor.Build(
		NewConfig,
		whyor.Struct[Deps]("Cfg"),
		whyor.FieldsOf[*Config]("Addr", "Port"),
		NewServer,
	))
}
