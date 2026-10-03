//go:build whyor

package x

import "github.com/Medzoner/whyor"

type Number = int
type Callback = func(Number) Number

func NewAlias() Callback      { return nil }
func NewPlain() func(int) int { return nil }
func Init() Callback          { panic(whyor.Build(NewAlias, NewPlain)) }
