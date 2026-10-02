//go:build whyor

package x

import "github.com/Medzoner/whyor"

type A struct{}
type B struct{}

func NewA(*B) *A { return nil }
func Init() *A   { panic(whyor.Build(NewA)) }
