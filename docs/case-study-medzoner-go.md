# Migration notes: using whyor in the author's Go application

**Scope:** an author-maintained application migrated from Wire, with compilation
and automated tests. This is not an independent customer testimonial, a
benchmark or a production certification.

## The starting point

The application had composition roots for its HTTP server, mock-backed test
server and database migrations. Constructors were grouped into provider sets,
interfaces used explicit bindings, and configuration fields were exposed as
dependencies. The objective was to keep those constructors and caller contracts,
while replacing the generator.

The initial migration used whyor's 0.7.x declaration API. The same basic
`Set`, `Bind`, `FieldsOf` and `Build` forms remain available in the current release.

## What changed

1. Changed the injector tag from `wireinject` to `whyor`.
2. Replaced `NewSet` with `Set`, pointer markers in Bind with type arguments,
   and FieldsOf's `new(T)` marker with a generic type argument.
3. Replaced Wire's generated file with `whyor_gen.go`.
4. Moved local helper constructors into untagged Go files. Wire had copied those
   helpers into its generated output; whyor calls them without copying definitions.
5. Updated the local generation command and verified generated-file freshness.

The server/test-server/database-migration entry points retained their result
contracts. No runtime DI container was introduced.

For a public, self-contained before/after rather than application-specific
source, see [the short migration example](wire-migration-example.md).

## What the exercise uncovered

**Toolchain compatibility matters.** A generator binary compiled with an older
Go version could not analyze the newer application's packages. Rebuilding with
an appropriate Go version fixed that tooling error. This is why the installation
documentation makes the toolchain requirement explicit.

**Compare tests before and after migration.** Two HTTP scenarios initially
failed both with the old Wire output and the whyor output. The failure came
from spaces in validation tags, not dependency resolution. Fixing the malformed
tags and updating a success fixture to satisfy the declared minimum length made
all 11 scenarios pass; no assertion was removed to claim a successful migration.

**Private-module access is separate from DI.** Later updates of the shared
library and OpenTelemetry required their own module-access configuration,
compatibility changes and regression fixes. A generated file does not make
private dependency downloads or vulnerability checks disappear.

## Validation performed

- Application build and Go vet checks.
- Generated injector freshness with `whyor check`.
- Unit tests and 11 Godog HTTP scenarios.
- A baseline comparison against the previous Wire-generated wiring while
  diagnosing the initially failing HTTP scenarios.

The migration and its initial validation were recorded in
[the application's migration PR](https://github.com/medzoner-org/medzoner-go/pull/226).
Repository access may be required. The runnable public examples and tests live
in this repository; readers do not need access to the application to try whyor.

## Limits of the evidence

- The author owns both the generator and this application: independent adoption
  is still something to establish, not something claimed by this case study.
- Passing builds and tests is not evidence of long-term production reliability.
- These notes do not assert a verified deployment, load test or performance gain.
- Wire compatibility is partial; applications using other declaration forms or
  platform-specific wiring need their own migration tests.

## Try it on your project

Start with one composition package on a branch, capture the baseline test
results, regenerate and inspect the output, then run the same tests again.
If it fails, include the declaration, provider signatures, Go version and a
small reproduction in an issue—without credentials or private application data.
