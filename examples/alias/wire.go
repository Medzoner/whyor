//go:build whyor

package alias

import "github.com/Medzoner/whyor"

func InitRepo() *Repo { panic(whyor.Build(NewDB, NewRepo)) }
