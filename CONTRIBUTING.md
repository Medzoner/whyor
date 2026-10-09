# Contributing to whyor

Thanks for trying the project. Compatibility reports from real Wire migrations
are particularly useful while the API is pre-1.0.

## Before a change

- Read the [README](README.md), [roadmap](ROADMAP.md) and
  [generation contracts](docs/guarantees.md).
- For a new API or structural change, open an issue first with a concrete use case.
- Keep bug fixes focused. Do not introduce a runtime container to solve a
  generation problem.

## Reproducing a bug

Include the Go version, whyor version/tag, command, provider signatures,
injector declaration and error output. A minimal module is better than a large
private application dump. Remove secrets, credentials and personal data.

## Development

Go 1.27.1+ is required for the current release.

`make check` also requires golangci-lint v2.13.2 (or a compatible version built
with Go 1.27+). The configuration checks returned errors, including blank assignments,
wrapping and error comparisons. It does not cap repeated issues or hide stdout/stderr
failures through the default exclusion presets. In-memory writers with documented
infallible writes are the only targeted errcheck exemptions:

```sh
go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.13.2
```

```sh
go run ./cmd/whyor gen ./examples/...
make check
go test -race ./...
go build ./...
```

For a generation change, add a regression test. Compile generated Go when a
format-only assertion would miss a type or scope error. Keep example outputs
versioned and fresh. For a CLI change, include invalid argument cases.

## Pull requests

Describe the original problem, the behavior change and the checks actually
run. Note any compatibility impact. Do not claim a check passed if it was not
executed, and keep local IDE files or generated build artifacts out of the PR.

## Documentation and demos

Examples should be runnable or clearly marked as illustrative. Do not turn a
successful author-maintained migration into a claim of independent adoption,
production validation or benchmark superiority.

The [showcase demo](docs/demo/README.md) runs in a temporary module using this
checkout and can be verified without recording software:

```sh
DEMO_PAUSE=0 bash docs/demo/run.sh
```
