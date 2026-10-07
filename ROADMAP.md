# Roadmap

whyor is a pre-1.0 compile-time DI generator. The priority is reliable,
inspectable Go output, not a runtime application framework. This roadmap is
a direction, not a delivery-date commitment.

## Available and tested

- Provider sets, explicit interface bindings, values and struct/field providers.
- Slice providers, automatic interface bindings and Close-based cleanup.
- Go type identity, including nested aliases.
- Explicitly instantiated generic function providers.
- Context-aware cleanup and joined errors, with legacy cleanup compatibility.
- Dependency diagnostics, tree/Mermaid/DOT/JSON graphs, watch mode and stale checks.
- Generated-name collision tests and output-file ownership checks.

See [generation contracts](docs/guarantees.md) for the precise boundaries and
the test evidence behind these statements.

## Next: independent migration feedback

- Help external users try one composition package without committing to a
  whole-application migration.
- Gather small reproductions for unsupported declarations and output errors.
- Turn compatibility reports into regression tests before broadening the API.

## Before a v1.0

- Test supported operating systems, not just cross-compile binaries.
- Define supported custom build tags and platform-specific injector behavior.
- Decide how obsolete generated outputs are detected and removed safely.
- Complete validation of imported sets and helpers declared only in tagged files.
- Tighten CLI argument validation and document deterministic output boundaries.
- Specify the compatibility policy for public declarations and graph JSON.

Local variable names and exact human-readable error wording are not intended
to become compatibility commitments.

## Small contribution candidates

These are proposed tasks, not claims that GitHub issues or assignments already exist.
Open a discussion/issue before starting so the scope can be agreed.

| Candidate | Scope | Acceptance criteria |
|---|---|---|
| Reject extra `init` arguments | CLI only, no generator rewrite | `whyor init a b` returns a usage error and creates neither directory; table-driven tests cover it. |
| Validate incomplete `show -f` | CLI flag parsing | Missing/unknown formats fail clearly before package loading; valid existing commands remain compatible. |
| Check doc links | Repository tooling | Check relative Markdown paths while skipping fenced code and external URLs; no network credential required. |
| Improve reproduction instructions | Issue template/documentation | Ask for Go version, provider/injector code and error output; warn users not to include secrets. |

Larger changes—especially scopes, inference and discovery—need an API proposal
and evidence of real use before implementation.

## Not planned for now

- A reflection-based runtime container or service locator.
- Implicit request/singleton scopes.
- Automatic discovery of constructors throughout the application.
- Taking over main, serving or shutdown through a framework lifecycle.

The generated injector should remain code you could reasonably write by hand.
