//go:build whyor

package lifecycle

import (
	"context"
	"github.com/Medzoner/whyor"
)

func InitApp(ctx context.Context, events *Events) (*App, whyor.Cleanup, error) {
	panic(whyor.Build(NewLegacy, NewCache, NewSocket, NewApp, whyor.Closer[*Socket]()))
}

func InitDetached(events *Events) (*App, func(context.Context) error, error) {
	panic(whyor.Build(NewDetached, NewFailed))
}

func InitLegacy(events *Events) (*Legacy, whyor.Cleanup) {
	panic(whyor.Build(NewLegacy))
}

func InitEmpty() (*App, whyor.Cleanup) {
	panic(whyor.Build(whyor.Value[*App](&App{})))
}
