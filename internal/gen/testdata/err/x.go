//go:build whyor

package x

import "github.com/medzoner/whyor"

type A struct{}
type B struct{}

func NewA() (*A, error) { return nil, nil }
func Init() *A          { panic(whyor.Build(NewA)) }
