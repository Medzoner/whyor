//go:build whyor

package basic

import "github.com/medzoner/whyor"

var DBSet = whyor.Set(NewConfig, NewPG, whyor.Bind[Store, *PG]())

func InitApp(dsn string) (*App, func(), error) {
	panic(whyor.Build(DBSet, NewApp))
}
