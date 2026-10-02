//go:build whyor

package x

import "github.com/Medzoner/whyor"

type A int

func Init() A { panic(whyor.Build(whyor.Struct[A]())) }
