# Migrating from Google Wire

whyor keeps Wire's model: injectors are declared in a file guarded by a build
tag, and a tool writes plain Go. The API is new (generics, no `interface{}`
arguments), so the migration is mechanical but not a rename.

## Files and commands

| Wire | whyor |
|---|---|
| `//go:build wireinject` | `//go:build whyor` |
| `wire_gen.go` | `whyor_gen.go` (guarded by `//go:build !whyor`) |
| `wire` / `wire gen` | `whyor gen` |
| `wire check` | `whyor check` |
| `wire diff` | `whyor check` (exits 1 if stale) |
| `wire show` | `whyor show` (dependency tree of each injector) |
| import `github.com/google/wire` | import `github.com/Medzoner/whyor` |

Extras: `whyor gen -w` regenerates on change, `whyor init` creates `wire.go`.

## API

| Wire | whyor |
|---|---|
| `panic(wire.Build(...))` | `panic(whyor.Build(...))` |
| `wire.NewSet(...)` | `whyor.Set(...)` |
| `wire.Bind(new(I), new(*T))` | `whyor.Bind[I, *T]()` |
| `wire.Value(v)` | `whyor.Value[T](v)` |
| `wire.InterfaceValue(new(I), v)` | `whyor.Value[I](v)` |
| `wire.Struct(new(T), "*")` | `whyor.Struct[T]()` (or `Struct[T]("*")`) |
| `wire.Struct(new(T), "A", "B")` | `whyor.Struct[T]("A", "B")` |
| `wire.FieldsOf(new(T), "A", "B")` | `whyor.FieldsOf[T]("A", "B")` |
| `wire.Struct` field tag `wire:"-"` | field tag `whyor:"-"` |
| `wire.Build` in a func returning `(T, func(), error)` | same shapes: `T`, `(T, func())`, `(T, error)`, `(T, func(), error)` |

Notice what changes in practice:

```go
// Wire
wire.Bind(new(Store), new(*PG))
wire.Struct(new(Deps), "*")
wire.FieldsOf(new(*Config), "Addr", "Port")

// whyor
whyor.Bind[Store, *PG]()
whyor.Struct[Deps]()
whyor.FieldsOf[*Config]("Addr", "Port")
```

`Struct[T]()` provides both `T` and `*T`, as Wire does, but builds them
independently: asking for both in one injector gives two instances.

## What whyor adds

These were long-standing open requests on Wire.

| Need | whyor |
|---|---|
| Collect several providers into a slice (wire#207) | `whyor.Many[T](a, b, c)` provides a `[]T` |
| Bind a type to every interface it implements (wire#242) | `whyor.AutoBind[T]()`, if exactly one registered type implements the interface |
| Use `Close()` as cleanup (wire#193) | `whyor.Closer[T]()` |
| Aliases (`type X = Y`) treated as the same type (wire#415) | automatic |

## Behaviour differences

- A provider runs once per injector, however many times its result is used.
- `Many` elements are not provided as their own type; add the provider outside
  the `Many` too if something else needs it.
- `Bind`, `AutoBind`, `Closer` and nested `Many` are rejected inside `Many`.
- Providers must be plain, non-variadic functions; methods are unsupported.
  Generic constructors require explicit instantiation, such as `NewRepository[User]`.
  Injectors themselves must remain non-generic.
- Unexported providers from another package are rejected.
- Errors carry a position and often a hint, for example
  `hint: *PG implements Store: add whyor.Bind[Store, *PG]()`.

## Checklist

1. Replace the import and the build tag (`wireinject` → `whyor`).
2. Rewrite `NewSet`, `Bind`, `Value`, `Struct`, `FieldsOf` as in the table.
3. Rename field tags `wire:"-"` to `whyor:"-"`.
4. Delete the old `wire_gen.go`, run `whyor gen ./...`, then your tests.
5. Add `whyor check ./...` to CI so a stale `whyor_gen.go` fails the build.
