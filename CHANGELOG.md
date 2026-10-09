# Changelog

Notable changes are recorded here starting with v0.8.0. Earlier releases are
available in the [GitHub release history](https://github.com/Medzoner/whyor/releases).
whyor is pre-1.0; review release notes before upgrading.

## [Unreleased]

### Error handling

- Check init skeleton writes and closure errors together; reject unreadable package
  clauses and invalid fallback package names instead of silently ignoring failures.
- Preserve temporary-output write, close and cleanup errors with errors.Join.
- Check fixture writes, reads and cleanup errors in tests.
- Add errcheck with blank-assignment checks to local, CI and release validation.
- Wrap filesystem, package-loading and rendering failures with operation/path
  context and %w, preserving errors.Is/errors.As through joined cleanup failures.
- Watch mode exits with a contextualized error when filesystem inspection fails,
  rather than treating a partial snapshot as a valid change. Normal deletions
  during polling remain supported; generation errors still keep watch running.

### Documentation

- Runnable showcase application and a five-screen terminal demo with a recording
  storyboard and an SVG preview of verified output.
- Short Wire migration diff and a factual author-maintained application case study.
- Contribution guide and roadmap separating verified features from v1.0 readiness work.
- Expanded package documentation, bug/proposal issue forms and a PR template.
- Editable social-preview artwork and a repository metadata setup guide. GitHub
  About/topics/image settings require a separate administrative action.

## [v0.11.0]

### Fixed

- `Value[T]` preserves its declared type, including typed constants, nil,
  interface values and named types, rather than relying on `:=` inference.
- Generated locals and import aliases avoid collisions with injector parameters
  and package declarations. Value expression imports follow the generated aliases.
- Unsupported Build declarations and injector files included in default builds
  are rejected instead of silently producing incorrect wiring.
- Duplicate Struct fields and injector parameters shadowing predeclared Go
  identifiers are rejected with explicit diagnostics.

### Output safety

- Validate all matching graphs before writing any output files.
- Refuse to overwrite manually owned files or follow output symbolic links.
- Replace each output through a same-directory temporary file and rename.
  Filesystem I/O failures are not a multi-file transaction.

### Documentation and tests

- Compile-and-run regression tests for typed values and generated-name collisions.
- Output preservation, symlink rejection and batch-validation tests.
- Documented generation contracts and explicit remaining work before a v1.0.

### Upgrade notes

- Requires Go **1.27.1+**; declaration signatures and graph schema are unchanged.
- Regenerate outputs: generated variable declarations and aliases can change.
- Rename injector parameters which shadow predeclared identifiers. Ensure each
  injector has a valid build tag and each existing output has the whyor header.
- This remains a pre-1.0 stabilization release, not a claim of API freeze.

## [v0.10.0]

### Added

- `whyor.Cleanup`, a `func(context.Context) error` type for explicit shutdown.
- Providers and injectors may return context-aware cleanup functions. Legacy
  `func()` provider cleanups can be mixed into a modern injector.
- Modern cleanup calls execute in reverse order and aggregate errors with
  `errors.Join`. `Closer` errors are preserved in modern injectors.
- Failed initialization aggregates its error with errors from acquired cleanups.
  Rollback preserves injector context values through `context.WithoutCancel`,
  or uses `context.Background()` when no context parameter exists.
- Lifecycle examples and tests covering shutdown context propagation, canceled
  acquisition contexts, mixed cleanups, rollback and empty cleanup functions.

### Upgrade notes

- Existing `func()` cleanup declarations retain their behavior. Modern provider
  cleanups require an injector returning `whyor.Cleanup` or the equivalent
  function signature; they cannot silently lose errors in a legacy injector.
- Shutdown context is supplied by the caller and passed unchanged. Rollback
  strips acquisition cancellation and deadlines; no timeout is added. Cleanup
  implementations must bound blocking operations themselves.
- Cleanups are cooperative, synchronous and not automatically idempotent.
- Requires Go **1.27.1+**. JSON graph schema stays at version `1`.

```sh
go install github.com/Medzoner/whyor/cmd/whyor@v0.10.0
go get github.com/Medzoner/whyor@v0.10.0
whyor gen ./...
go test ./...
```

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

[Unreleased]: https://github.com/Medzoner/whyor/compare/v0.11.0...HEAD
[v0.11.0]: https://github.com/Medzoner/whyor/compare/v0.10.0...v0.11.0
[v0.10.0]: https://github.com/Medzoner/whyor/compare/v0.9.0...v0.10.0
[v0.9.0]: https://github.com/Medzoner/whyor/compare/v0.8.0...v0.9.0
[v0.8.0]: https://github.com/Medzoner/whyor/compare/v0.7.1...v0.8.0
