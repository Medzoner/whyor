//go:build whyor

package x

import "github.com/Medzoner/whyor"

type A struct{}
type S struct{}

func NewA() A    { return A{} }
func NewS(*A) *S { return nil }

func Init() *S { panic(whyor.Build(NewA, NewS)) }
