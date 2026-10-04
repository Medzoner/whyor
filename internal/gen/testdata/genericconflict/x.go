//go:build whyor

package x

import "github.com/Medzoner/whyor"

type Result struct{}

func New[T any]() *Result { return nil }
func Init() *Result       { panic(whyor.Build(New[int], New[string])) }
