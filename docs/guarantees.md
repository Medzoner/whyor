# Generation contracts and v1.0 readiness

This document describes the verified contracts of whyor. It is not a claim
that the pre-1.0 API is frozen or every valid Go program is supported.

## Dependency identity

- Types are compared using Go identity, including nested aliases. Different
  defined types do not become identical because they share an underlying type.
- Generic providers require explicit arguments. Calls are cached by function
  and argument identity, including arguments absent from the result signature.
- Injector parameters supply their types directly and take precedence over
  registered providers of those types. `unused` can report those unused providers.
- `Value[T](expr)` uses a typed variable declaration, preserving the assignment
  semantics of `T`, including untyped constants, nil and interface values.

## Declarations and names

- An injector contains only `panic(whyor.Build(...))`, using the builtin panic.
  Unsupported uses of Build are rejected instead of silently ignored.
- The injector source must be excluded from the default build; use the `whyor`
  build tag. Helpers and application types belong in ordinary Go files.
- Generated locals and import aliases avoid injector parameter names and
  package-level declarations. Source import aliases in Value expressions are
  rewritten to match the generated file, without changing the original AST
  after expression rendering.
- Injector parameters which shadow predeclared Go identifiers (such as
  `error`, `nil` or `new`) are currently rejected explicitly. Rename them.
- Duplicate explicit Struct field names are rejected before generation.

## File ownership and writes

- whyor writes `whyor_gen.go`, with its generated ownership header.
- Existing files without that header are not overwritten. Symbolic links and
  other non-regular output files are rejected.
- All loaded graphs and output ownership checks are validated before any write.
  A generation error in a later package does not leave earlier outputs updated.
- Each output is written to a temporary file in the same directory, then renamed.
  Atomic replacement depends on operating-system filesystem guarantees.
- This is **not a multi-file filesystem transaction**. A write failure can occur
  after another file was replaced. Concurrent writers/external file changes are
  not coordinated; output checks do not provide a security sandbox.
- Removing an injector does not automatically remove its former output. Remove
  obsolete generated files explicitly and run compilation/tests alongside check.

## Lifecycle

- Initialization is synchronous. Reachable function providers run once per
  instantiation, per injector. There is no cross-injector singleton container.
- Cleanup calls run in reverse acquisition order. In modern mode, errors are
  joined; legacy mode retains its original behavior.
- A failing provider owns its partial resources. Only cleanups from previously
  successful providers are used for rollback.
- Successful providers must return non-nil cleanup functions. Cleanup is not
  automatically idempotent, panic-safe or concurrency-safe.
- Context cancellation is cooperative; rollback does not inherit acquisition
  cancellation/deadlines or receive an implicit timeout. See the README for details.

## Evidence

`make check` runs vet, the test suite, freshness checks on the versioned examples
and unused-provider checks. Tests include generated packages compiled and run by
the Go toolchain, type identity, random graphs, lifecycle paths and output safety.
The suite also runs under the race detector during local validation.

## Remaining work before claiming v1.0

- Confirm supported Go versions and cross-platform behavior in a broader test
  matrix, beyond cross-compiling release binaries.
- Define behavior for custom build tags and platform-specific injector variants.
- Complete validation of source-level references in imported sets and helpers
  declared only in tagged files.
- Decide how obsolete outputs should be detected and removed safely.
- Complete CLI argument validation and reproducible generation-order guarantees.
- Decide which diagnostics and graph formats are compatibility commitments;
  generated local names and human-readable error text are not stable interfaces.
- Review package-level shadowing of predeclared identifiers and unsupported
  Go expressions beyond the tested cases.

These are readiness criteria, not implemented features. Releases remain 0.x
until the compatibility policy and supporting tests justify a v1.0.
