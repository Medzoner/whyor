//go:build whyor

package main

import "github.com/Medzoner/whyor"

var Providers = whyor.Set(
	NewConfig,
	NewStore,
	whyor.Bind[Store, *MemoryStore](),
)

func InitApp(name string) (*App, func(), error) {
	panic(whyor.Build(Providers, NewApp))
}
