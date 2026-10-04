//go:build whyor

package x

import "github.com/Medzoner/whyor"

type A struct{ N int }

func Init(n int) *A { panic(whyor.Build(whyor.Struct[A]("N", "N"))) }
