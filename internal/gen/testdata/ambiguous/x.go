//go:build whyor

package x

import "github.com/Medzoner/whyor"

type Store interface{ Read() }
type First struct{}
type Second struct{}

func (*First) Read()     {}
func (*Second) Read()    {}
func NewFirst() *First   { return nil }
func NewSecond() *Second { return nil }

func Init() Store {
	panic(whyor.Build(NewFirst, NewSecond, whyor.AutoBind[*First](), whyor.AutoBind[*Second]()))
}
