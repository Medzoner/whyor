// Package whyor is a compile-time dependency injection toolkit.
//
// Injectors are declared in files guarded by the "whyor" build tag:
//
//	//go:build whyor
//
//	func InitApp(path string) (*App, func(), error) {
//		panic(whyor.Build(NewConfig, NewDB, whyor.Bind[Store, *PG]()))
//	}
//
// `whyor gen` then writes a whyor_gen.go file containing plain Go code.
// Nothing in this package does any work at runtime.
package whyor

import "context"

// Cleanup releases resources using a caller-supplied shutdown context.
// Generated cleanup calls run in reverse acquisition order and aggregate errors.
type Cleanup func(context.Context) error

// Option is a provider, a Set, a Bind or a Value accepted by Build and Set.
// Plain provider functions are accepted too.
type Option any

// Set groups providers and options so they can be reused.
func Set(opts ...Option) Option { return opts }

// Build declares an injector body. It must be the argument of panic in an
// injector function. It is replaced by generated code.
func Build(opts ...Option) any { return nil }

// Bind tells the generator that interface I is satisfied by concrete type T.
func Bind[I, T any]() Option { return nil }

// Struct provides both T and *T by filling fields of struct T with
// dependencies. With no argument (or "*"), every exported field not tagged
// `whyor:"-"` is filled; otherwise only the named exported fields are.
func Struct[T any](fields ...string) Option { return nil }

// FieldsOf provides the named fields of T (a struct or pointer to struct)
// as dependencies of their own types.
func FieldsOf[T any](fields ...string) Option { return nil }

// Many provides a []T made of the given providers, each returning a value
// assignable to T. Its elements are not provided as T on their own.
func Many[T any](opts ...Option) Option { return nil }

// AutoBind makes T usable wherever an interface it implements is needed,
// as long as exactly one registered AutoBind type implements that interface.
func AutoBind[T any]() Option { return nil }

// Closer registers T's Close method as its cleanup: the injector must return
// func() or Cleanup. Close errors are preserved with Cleanup, discarded with func().
// It applies to providers of T that have no cleanup of their own.
func Closer[T any]() Option { return nil }

// Value provides the given expression as a dependency of type T.
func Value[T any](v T) Option { return nil }
