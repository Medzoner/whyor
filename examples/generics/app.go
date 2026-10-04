package generics

import "github.com/Medzoner/whyor/examples/generics/providers"

type User struct{ Name string }
type Person = User
type Order struct{ ID int }

type App struct {
	Users  *providers.Repository[User]
	Orders *providers.Repository[Order]
	Labels []providers.Label
	Label  providers.Label
	Pair   providers.Pair[User, Order]
}

func NewApp(users *providers.Repository[User], orders *providers.Repository[Order], labels []providers.Label,
	label providers.Label, pair providers.Pair[User, Order]) *App {
	return &App{Users: users, Orders: orders, Labels: labels, Label: label, Pair: pair}
}
