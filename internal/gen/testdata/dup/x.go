//go:build whyor

package x

import "github.com/medzoner/whyor"

type A struct{}
type B struct{}

func NewA() *A  { return nil }
func NewA2() *A { return nil }
func Init() *A  { panic(whyor.Build(NewA, NewA2)) }
