//go:build whyor

package alias

import "github.com/medzoner/whyor"

func InitRepo() *Repo { panic(whyor.Build(NewDB, NewRepo)) }
