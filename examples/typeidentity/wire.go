//go:build whyor

package typeidentity

import "github.com/Medzoner/whyor"

func InitResult(p struct{ Value Number }) Result {
	panic(whyor.Build(NewTransform, NewOptions, NewReader, NewBox, NewResult))
}

func InitBox() (*Box[int], func()) {
	panic(whyor.Build(NewBox, whyor.Closer[*Box[int]]()))
}

func InitBoxes() []*Box[int] {
	panic(whyor.Build(whyor.Many[*Box[Number]](NewBox)))
}
