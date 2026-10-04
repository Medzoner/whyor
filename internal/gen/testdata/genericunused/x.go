//go:build whyor

package x

import "github.com/Medzoner/whyor"

type Box[T any] struct{ Value T }

func New[T any]() *Box[T] { return nil }
func Init() *Box[int]     { panic(whyor.Build(New[int], New[string])) }
