//go:build whyor

package generics

import (
	"github.com/Medzoner/whyor"
	"github.com/Medzoner/whyor/examples/generics/providers"
)

var Repositories = whyor.Set(providers.NewRepository[User], providers.NewRepository[Order])

func InitApp(user User, order Order, events *providers.Events) (*App, func(), error) {
	panic(whyor.Build(
		Repositories,
		providers.NewLabel[User],
		whyor.Many[providers.Label](providers.NewLabel[User], providers.NewLabel[Order], providers.NewLabel[Person]),
		providers.NewPair[User, Order],
		NewApp,
	))
}
