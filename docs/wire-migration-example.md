# A small Wire → whyor migration

Use [examples/showcase](../examples/showcase) for the runnable result. The Wire
declaration below is a migration illustration, not an extra module dependency.
The before/after uses the same constructors and injector result shape.

## Before: Wire

```go
//go:build wireinject

package main

import "github.com/google/wire"

var Providers = wire.NewSet(
	NewConfig,
	NewStore,
	wire.Bind(new(Store), new(*MemoryStore)),
)

func InitApp(name string) (*App, func(), error) {
	panic(wire.Build(Providers, NewApp))
}
```

## The declaration diff

```diff
-//go:build wireinject
+//go:build whyor

-import "github.com/google/wire"
+import "github.com/Medzoner/whyor"

-var Providers = wire.NewSet(
+var Providers = whyor.Set(
     NewConfig,
     NewStore,
-    wire.Bind(new(Store), new(*MemoryStore)),
+    whyor.Bind[Store, *MemoryStore](),
 )

 func InitApp(name string) (*App, func(), error) {
-    panic(wire.Build(Providers, NewApp))
+    panic(whyor.Build(Providers, NewApp))
 }
```

`NewConfig`, `NewStore`, `NewApp` and the caller do not need to know about either
DI tool. Keep their implementations in ordinary, untagged files.

## Migration commands

In an existing application module, after editing the declaration:

```sh
go install github.com/Medzoner/whyor/cmd/whyor@v0.11.0
go get github.com/Medzoner/whyor@v0.11.0
# Remove the obsolete Wire-generated file, not your constructor source files.
# Example only: adjust this path to your application's composition package.
rm path/to/composition/wire_gen.go
whyor gen ./...
go mod tidy
go test ./...
whyor check ./...
```

Do not leave both generated injectors in the normal build. If Wire used to copy
helpers from a tagged file, move those helpers to an untagged file first: whyor
does not copy their definitions into its output.

## What this example proves—and does not

- It illustrates the declaration changes needed for a set and an interface bind.
- The corresponding whyor example is executable and tested in this repository.
- It is not a claim that every Wire feature or existing project migrates unchanged.
- It does not establish performance or production reliability.

For the broader API mapping and behavior differences, read
[the migration guide](migrating-from-wire.md).
