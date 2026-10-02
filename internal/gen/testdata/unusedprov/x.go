//go:build whyor

package x

import "github.com/Medzoner/whyor"

type A struct{}
type B struct{}

func NewA() *A { return nil }
func NewB() *B { return nil }

func Init() *A { panic(whyor.Build(NewA, NewB)) }
