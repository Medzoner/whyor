//go:build whyor

package x

import "github.com/Medzoner/whyor"

type Box[T any] struct{ Value T }

func NewIntBox() *Box[int] { return nil }
func Init() *Box[string]   { panic(whyor.Build(NewIntBox)) }
