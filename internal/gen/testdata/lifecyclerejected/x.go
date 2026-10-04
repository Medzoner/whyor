//go:build whyor

package x

import (
	"github.com/Medzoner/whyor"
)

type A struct{}

func NewA() (*A, whyor.Cleanup) { return nil, nil }
func Init() (*A, func())        { panic(whyor.Build(NewA)) }
