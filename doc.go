// Package whyor provides declarations for compile-time dependency injection in Go.
// It is an independent, pre-1.0 alternative inspired by Google Wire, not an
// official continuation or a drop-in replacement.
//
// # How it works
//
// Constructors are ordinary Go functions. Their parameters describe dependencies
// and their results describe the values they provide. The whyor CLI analyzes
// injector declarations and writes whyor_gen.go with regular constructor calls.
// There is no runtime dependency-injection container or reflection-based lookup.
//
// Injector declarations belong in files guarded by the whyor build tag:
//
//	//go:build whyor
//
//	package app
//
//	import "github.com/Medzoner/whyor"
//
//	var Providers = whyor.Set(NewConfig, NewStore, whyor.Bind[Store, *MemoryStore]())
//
//	func InitApp(name string) *App {
//		panic(whyor.Build(Providers, NewApp))
//	}
//
// The constructors and application types in this illustration are defined in
// ordinary, untagged files. The declaration's panic is a generation marker,
// not an application startup strategy. Generated files use the !whyor tag.
//
// # CLI and installation
//
// Install the CLI separately from the declaration package:
//
//	go install github.com/Medzoner/whyor/cmd/whyor@latest
//	go get github.com/Medzoner/whyor@latest
//	whyor gen ./...
//	go test ./...
//	whyor check ./...
//
// Commit generated files so normal builds need not run the generator. Current
// releases require Go 1.27.1 or newer; the generator must be built with a Go
// version at least as recent as the target application's version.
//
// The CLI also supports dependency graph exports (tree, JSON, DOT and Mermaid),
// unused-provider checks, a declaration skeleton and a polling watch mode.
//
// # Declarations
//
// [Set] groups reusable providers. [Bind] explicitly selects an implementation
// for an interface, while [AutoBind] opts into unambiguous automatic binding.
// [Value] provides a typed expression; [Struct] constructs structs and [FieldsOf]
// exposes configuration fields. [Many] provides a slice of implementations.
// Generic provider functions require explicit instantiation, such as NewStore[User].
//
// # Cleanup
//
// Providers can return T, (T, error), (T, func()), (T, func(), error), or the
// corresponding forms using [Cleanup]. Acquired cleanups execute in reverse
// order. Modern Cleanup callbacks receive the caller's shutdown context and
// aggregate errors. [Closer] can register an existing Close method.
//
// Rollback strips acquisition cancellation and deadlines; no implicit timeout,
// idempotency or concurrency safety is added. Cleanup implementations must bound
// blocking operations themselves. A failing provider owns its partial resources.
//
// # Examples and migration
//
// Start with the [runnable demo], [complete examples] and [Wire migration guide].
// The [generation contracts] document records verified guarantees and deliberate
// limitations. API signatures and behavior may evolve before v1.0.
//
// [runnable demo]: https://github.com/Medzoner/whyor/blob/main/docs/demo/README.md
// [complete examples]: https://github.com/Medzoner/whyor/tree/main/examples
// [Wire migration guide]: https://github.com/Medzoner/whyor/blob/main/docs/migrating-from-wire.md
// [generation contracts]: https://github.com/Medzoner/whyor/blob/main/docs/guarantees.md
package whyor
