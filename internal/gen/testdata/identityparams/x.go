//go:build whyor

package x

import "github.com/Medzoner/whyor"

type Number = int
type Callback = func(Number) Number

func Init(a Callback, b func(int) int) Callback { panic(whyor.Build()) }
