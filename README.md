# whyor

Compile-time dependency injection for Go, in the spirit of
[Google Wire](https://github.com/google/wire) (archived in 2025), with a
generics-based API. It writes plain Go: no reflection, no runtime container.

## Install

```
go install github.com/Medzoner/whyor/cmd/whyor@latest
go get github.com/Medzoner/whyor@latest
```

The binary must be built with a Go version at least as recent as the one your
project uses.

## Example

Providers are ordinary constructors, returning `T`, `(T, func())`,
`(T, error)` or `(T, func(), error)` (the `func()` is a cleanup):

```go
// app.go
func NewConfig(dsn string) *Config                { ... }
func NewPG(cfg *Config) (*PG, func(), error)      { ... }
func NewApp(s Store) *App                         { ... }
```

Declare the injector in a file guarded by the `whyor` build tag:

```go
// wire.go
//go:build whyor

package app

import "github.com/Medzoner/whyor"

var DBSet = whyor.Set(NewConfig, NewPG, whyor.Bind[Store, *PG]())

func InitApp(dsn string) (*App, func(), error) {
	panic(whyor.Build(DBSet, NewApp))
}
```

Run `whyor gen ./...`. It writes `whyor_gen.go` (guarded by `!whyor`), plain
Go with cleanups called in reverse order, also on error:

```go
func InitApp(dsn string) (*App, func(), error) {
	v1 := NewConfig(dsn)
	v2, cleanup2, err := NewPG(v1)
	if err != nil {
		return nil, nil, err
	}
	v3 := NewApp(v2)
	return v3, func() {
		cleanup2()
	}, nil
}
```

Commit `whyor_gen.go`: builds and CI then need no generator.

## Commands

| Command | Does |
|---|---|
| `whyor gen [-w] [packages]` | write `whyor_gen.go` (`-w`: regenerate on change) |
| `whyor check [packages]` | exit 1 if a generated file is stale |
| `whyor show [-f tree\|mermaid\|dot] [packages]` | print each injector's dependencies |
| `whyor unused [packages]` | list providers no injector calls (exit 1) |
| `whyor init [dir]` | create a `wire.go` skeleton |

With `go generate`, from a file without the `whyor` tag:

```go
//go:generate whyor gen .
```

## API

| | |
|---|---|
| `Build(...)`, `Set(...)` | declare an injector / group providers |
| `Bind[I, T]()` | provide interface `I` with concrete `T` |
| `Value[T](v)` | provide a literal |
| `Struct[T](fields...)` | provide `T` and `*T` by filling fields: all exported ones by default (skip with `whyor:"-"`), or only the named ones |
| `FieldsOf[T]("A", "B")` | provide fields of a struct as dependencies |
| `Many[T](providers...)` | provide a `[]T` |
| `AutoBind[T]()` | use `T` for any needed interface it implements, if unambiguous |
| `Closer[T]()` | call `T.Close()` in the injector's cleanup |

Each provider runs once per injector. Type aliases (`type X = Y`) are the same
type. Errors carry a position and, when possible, a hint:

```
wire.go:17:24: Init: no provider for Store (needed by *App)
	hint: *PG implements Store: add whyor.Bind[Store, *PG]() or whyor.AutoBind[*PG]()
```

## Compared to Wire

| | Wire | whyor |
|---|---|---|
| API | `new(T)` and `interface{}` arguments | generics, checked by the compiler |
| Several providers into a slice | no (wire#207) | `Many` |
| Bind to all implemented interfaces | no (wire#242) | `AutoBind` |
| `Close()` as cleanup | no (wire#193) | `Closer` |
| Type aliases | lost (wire#415) | handled |
| Dependency graph | `wire show` | `whyor show`, also mermaid and dot |
| Unused providers | no | `whyor unused` |
| Watch mode | no | `whyor gen -w` |
| Maintained | archived Aug 2025 | yes |

Migrating: see [docs/migrating-from-wire.md](docs/migrating-from-wire.md).
More examples live in [examples/](examples).

## Develop

```
make install   # install the binary
make check     # vet, tests, whyor check, whyor unused
```

## License

MIT
