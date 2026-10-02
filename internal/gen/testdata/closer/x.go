//go:build whyor

package x

import "github.com/Medzoner/whyor"

type A struct{}

func (*A) Close() error { return nil }
func NewA() *A          { return nil }

func Init() *A { panic(whyor.Build(NewA, whyor.Closer[*A]())) }
