# Changelog

Notable changes are recorded here starting with v0.8.0. Earlier releases are
available in the [GitHub release history](https://github.com/Medzoner/whyor/releases).
whyor is pre-1.0; review release notes before upgrading.

## [Unreleased]

## [v0.9.0]

### Added

- Explicit generic provider instantiations, including imported constructors
  and multiple type arguments: `NewRepository[User]`, `pkg.NewPair[A, B]`.
- Generation resolves the instantiated signature and emits explicit generic
  calls; Go validates the type constraints.
- Function instantiations have separate identities even when the type arguments
  do not appear in their signatures. Alias-equivalent instantiations share one
  call per injector.
- Generic providers work inside `Set` and `Many`, with error propagation and
  reverse-order cleanup.
- Graphs and `unused` distinguish instantiations rather than only function names.
- Executable generic-provider examples and regression tests for conflicts,
  constraints, shared calls and failure-path cleanup.

### Upgrade notes

- Requires Go **1.27.1+**; existing declaration signatures are unchanged.
- All generic type arguments must be explicit. There is no automatic inference
  or discovery of generic instantiations.
- Different instantiations supplying the same result type conflict outside
  `Many`, just like different non-generic providers.
- Injectors remain non-generic; method and variadic providers remain unsupported.
- JSON graph schema remains at version `1`; generic provider names include
  their type arguments.

```sh
go install github.com/Medzoner/whyor/cmd/whyor@v0.9.0
go get github.com/Medzoner/whyor@v0.9.0
whyor gen ./...
go test ./...
```

## [v0.8.0]

### Added

- `whyor show -f json`: a versioned document containing all matching injectors,
  with packages, source positions, nodes, providers, parameters and edges.
- Dependency diagnostics include the path from the injector result to the
  missing dependency, with provider and binding positions.
- Explicit diagnostics for ambiguous `AutoBind` candidates and both source
  positions for conflicting providers.
- Regression tests for nested aliases, equivalent interfaces, instantiated
  generic types, duplicate providers, cycles and graph identity.
- Release artifacts include `SHA256SUMS`; release notes come from this changelog.

### Fixed

- Dependency identity now uses Go's `types.Identical` semantics through
  `typeutil.Map`, rather than printed type names and partial alias normalization.
  This handles aliases nested in signatures, structs and generic arguments
  while preserving the distinction between different defined types.
- Cleanup result aliases such as `type Cleanup = func()` are recognized.
- DOT, Mermaid and JSON graphs distinguish different providers of the same
  type inside `Many`, while sharing nodes for calls to the same function.
- Graph provider names use the generated import aliases to distinguish packages
  that share a name.
- Repeated equivalent `AutoBind` registrations do not create false ambiguity.

### Documentation

- Expanded README with badges, a runnable quick start, API examples, limitations
  and an SVG explaining generation and runtime constructor calls.
- IDE settings are no longer tracked; `.idea/` is ignored.

### Upgrade notes

- Requires Go **1.27.1+**, unchanged from v0.7.1.
- Public declaration signatures are unchanged. Error messages are richer;
  integrations should not depend on their exact text.
- JSON schema version is `1`. Node IDs are local to each injector; source paths
  may be absolute. JSON export does not include machine-readable error output.
- Generic **types** are supported; explicitly instantiated generic **provider
  functions** remain unsupported in this release.

```sh
go install github.com/Medzoner/whyor/cmd/whyor@v0.8.0
go get github.com/Medzoner/whyor@v0.8.0
whyor gen ./...
go test ./...
```

[Unreleased]: https://github.com/Medzoner/whyor/compare/v0.9.0...HEAD
[v0.9.0]: https://github.com/Medzoner/whyor/compare/v0.8.0...v0.9.0
[v0.8.0]: https://github.com/Medzoner/whyor/compare/v0.7.1...v0.8.0
