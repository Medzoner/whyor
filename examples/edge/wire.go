//go:build whyor

package edge

import (
	"time"

	"github.com/Medzoner/whyor"
	alog "github.com/Medzoner/whyor/examples/edge/a/log"
	blog "github.com/Medzoner/whyor/examples/edge/b/log"
)

func InitOut() (Out, func(), error) {
	panic(whyor.Build(
		whyor.Value[time.Duration](5*time.Second),
		whyor.Value[string]("edge"),
		alog.New, blog.New,
		NewR1, NewR2, NewR3, NewOut,
	))
}
