package edge

import (
	"errors"
	"time"

	alog "github.com/medzoner/whyor/examples/edge/a/log"
	blog "github.com/medzoner/whyor/examples/edge/b/log"
)

// Out is returned by value so a failing injector needs a zero value.
type Out struct {
	Timeout time.Duration
	Name    string
	A       *alog.T
	B       *blog.T
}

type Res struct{ n int }

type R1 Res
type R2 Res
type R3 Res

var Events []string

func rec(s string) func() { return func() { Events = append(Events, s) } }

var Fail bool

func NewR1() (*R1, func())    { return &R1{}, rec("r1") }
func NewR2(*R1) (*R2, func()) { return &R2{}, rec("r2") }
func NewR3(*R2) (*R3, func(), error) {
	if Fail {
		return nil, nil, errors.New("boom")
	}
	return &R3{}, rec("r3"), nil
}

func NewOut(d time.Duration, name string, a *alog.T, b *blog.T, _ *R3) Out {
	return Out{d, name, a, b}
}
