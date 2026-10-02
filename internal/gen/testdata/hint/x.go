//go:build whyor

package x

import "github.com/Medzoner/whyor"

type I interface{ M() }
type A struct{}

func (*A) M()  {}
func NewA() *A { return nil }

type S struct{ I I }

func NewS(I) *S { return nil }

func Init() *S { panic(whyor.Build(NewA, NewS)) }
