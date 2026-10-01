//go:build whyor

package features

import "github.com/medzoner/whyor"

func InitServer(log *[]string) (*Server, func()) {
	panic(whyor.Build(
		whyor.Many[Handler](NewUsers, NewOrders),
		whyor.AutoBind[*Users](), // satisfies the Handler needed by NewServer
		whyor.Closer[*Cache](),
		NewUsers, NewCache, NewRouter, NewServer,
	))
}
