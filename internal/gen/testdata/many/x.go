//go:build whyor

package x

import "github.com/Medzoner/whyor"

type A struct{}

func NewA() *A { return nil }

func Init() []fmtStringer { panic(whyor.Build(whyor.Many[fmtStringer](NewA))) }

type fmtStringer interface{ String() string }
