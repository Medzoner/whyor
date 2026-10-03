//go:build whyor

package x

import "github.com/Medzoner/whyor"

type A struct{}
type Root struct{}

func NewA() *A               { return nil }
func NewB() *A               { return nil }
func NewRoot([]*A, *A) *Root { return nil }

func Init() *Root {
	panic(whyor.Build(NewA, NewRoot, whyor.Many[*A](NewA, NewB)))
}
