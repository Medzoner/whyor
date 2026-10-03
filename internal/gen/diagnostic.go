package gen

import (
	"errors"
	"fmt"
	"go/types"
	"strings"
)

// dependencyError captures the resolution path before the resolver unwinds.
type dependencyError struct {
	cause error
	path  []string
}

func (e *dependencyError) Error() string {
	return fmt.Sprintf("%s\n\tdependency path:\n\t  %s", e.cause, strings.Join(e.path, "\n\t  → "))
}

func (e *dependencyError) Unwrap() error { return e.cause }

func (r *resolver) failure(err error, requested types.Type) error {
	var existing *dependencyError
	if errors.As(err, &existing) {
		return err
	}
	path := append([]types.Type(nil), r.stack...)
	if requested != nil {
		path = append(path, requested)
	}
	steps := make([]string, 0, len(path))
	for _, t := range path {
		step := r.f.typ(t)
		if b, ok := r.lookup(t); ok {
			step = fmt.Sprintf("%s [%s]", step, r.provider(b))
			if b.position.IsValid() {
				step = fmt.Sprintf("%s at %s", step, b.position)
			}
		}
		steps = append(steps, step)
	}
	return &dependencyError{cause: err, path: steps}
}
