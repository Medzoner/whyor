package providers

import (
	"errors"
	"reflect"
)

type Events struct {
	Closed []string
	Fail   bool
}

type Repository[T any] struct{ Value T }

func NewRepository[T any](value T, events *Events) (*Repository[T], func(), error) {
	name := reflect.TypeFor[T]().Name()
	if events.Fail && name == "Order" {
		return nil, nil, errors.New("order initialization failed")
	}
	return &Repository[T]{Value: value}, func() { events.Closed = append(events.Closed, name) }, nil
}

// T deliberately occurs only in the implementation, not in the signature.
type Label struct{ Name string }

func NewLabel[T any]() Label { return Label{Name: reflect.TypeFor[T]().Name()} }

type Pair[A, B any] struct {
	First  A
	Second B
}

func NewPair[A, B any](a A, b B) Pair[A, B] { return Pair[A, B]{First: a, Second: b} }
