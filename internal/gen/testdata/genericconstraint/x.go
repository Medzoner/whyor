//go:build whyor

package x

import "github.com/Medzoner/whyor"

func New[T ~int]() T { return 0 }
func Init() string   { panic(whyor.Build(New[string])) }
