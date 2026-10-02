# whyor

Compile-time dependency injection for Go. Successor to Google Wire, with a
generics-based API. Generates plain Go; nothing runs at runtime.

```go
//go:build whyor

var DBSet = whyor.Set(NewConfig, NewPG, whyor.Bind[Store, *PG]())

func InitApp(dsn string) (*App, func(), error) {
	panic(whyor.Build(DBSet, NewApp))
}
```

```
go run github.com/Medzoner/whyor/cmd/whyor gen ./...    # writes whyor_gen.go
go run github.com/Medzoner/whyor/cmd/whyor check ./...  # exit 1 if stale
go run github.com/Medzoner/whyor/cmd/whyor init ./internal/app  # wire.go skeleton
```

Regenerate on every change (polling, no extra dependency):

```
go run github.com/Medzoner/whyor/cmd/whyor gen -w ./...
```

Or with `go generate`, from a file without the `whyor` tag:

```go
//go:generate go run github.com/Medzoner/whyor/cmd/whyor gen .
```

API:
- `Build`, `Set`: declare an injector / group providers.
- `Bind[I, T]()`: provide interface `I` with concrete `T`.
- `Value[T](v)`: provide a literal.
- `Struct[T]()`: provide `T` and `*T` by filling exported fields (skip with `whyor:"-"`).
- `FieldsOf[T]("A", "B")`: provide fields of a struct as dependencies.
- `Many[T](providers...)`: provide a `[]T`.
- `AutoBind[T]()`: use `T` for any needed interface it implements (must be unambiguous).
- `Closer[T]()`: call `T.Close()` in the injector's cleanup.

 Providers are functions
returning `T`, `(T, func())`, `(T, error)` or `(T, func(), error)`.
See `examples/basic`.

Coming from Wire? See [docs/migrating-from-wire.md](docs/migrating-from-wire.md).
