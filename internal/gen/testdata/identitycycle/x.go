//go:build whyor

package x

import "github.com/Medzoner/whyor"

type Number = int
type Options = struct{ Value Number }

func NewOptions(struct{ Value int }) Options { return Options{} }
func Init() Options                          { panic(whyor.Build(NewOptions)) }
