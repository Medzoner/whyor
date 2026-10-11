# Source code sequence diagrams

These diagrams describe the v0.12.1 implementation. Participant labels refer
to functions or internal objects, not network services. Solid arrows are calls;
dashed arrows are returns. Error branches terminate the relevant operation even
when the diagram omits the remaining return arrows for readability.

**The key distinction:** the generator inspects constructors and writes Go code;
it does not execute those constructors. Your compiled application executes the
generated injector later, without a runtime DI container.

## 1. From the CLI to generated files

Sources: [CLI dispatch](../cmd/whyor/main.go),
[package loading and generation](../internal/gen/gen.go).

```mermaid
sequenceDiagram
    actor User
    participant CLI as runWithWriters / runGen
    participant Run as gen.Run
    participant Load as load
    participant Go as go/packages
    participant Gen as generate
    participant Output as outputFile / readOutput
    User->>CLI: whyor gen ./...
    CLI->>Run: Run(".", patterns, true)
    Run->>Load: load(dir, patterns)
    Load->>Go: packages.Load with -tags=whyor
    Go-->>Load: typed packages and ASTs
    Load->>Load: Visit packages, check errors, indexVars
    Load-->>Run: packages + declaration index
    loop Each matching package
        Run->>Gen: generate(package, declarations, options)
        Gen-->>Run: formatted source, nil, or error
        opt Source was generated
            Run->>Output: readOutput(path)
            Output-->>Run: previous source or error
            Run->>Run: Compare bytes, queue changed outputs
        end
    end
    Note over Run,Output: Only after all packages pass generation and output validation
    loop Each queued output
        Run->>Output: output.write()
        Output-->>Run: success or error
    end
    Run-->>CLI: sorted changed paths
    CLI->>User: writeLines(stdout, paths)
```

- `check` follows the same path with `write=false`. It does not write files and
  fails if any generated file is stale.
- A generation/validation error prevents the write phase. An I/O error during
  that phase can leave earlier files updated: **this is not a multi-file transaction**.
- CLI output errors are returned too; successful generation does not guarantee
  that printing its result succeeds.

## 2. Turning an injector declaration into code

Sources: [generateFile and buildCall](../internal/gen/gen.go),
[inject and its helpers](../internal/gen/emit.go),
[collector](../internal/gen/collect.go).

```mermaid
sequenceDiagram
    participant Gen as generate / generateFile
    participant AST as Go AST and type information
    participant File as file.injector / inject
    participant Collector as collector
    participant Resolver as resolver
    Gen->>AST: Find function declarations and buildCall
    Note over Gen,AST: Recognize the sole panic(whyor.Build(...)) statement
    Gen->>Gen: checkInjectorBuildTag
    Gen->>File: injector(function, Build call)
    File->>File: Validate signature, parseResults
    loop Build arguments
        File->>Collector: collect(expression, package)
        alt API call such as Set, Many, Bind, Value
            Collector->>Collector: collectCall + dedicated handler
        else Function or named set reference
            Collector->>Collector: collectRef, addFunc or expand declaration
        end
        Collector->>Collector: Register bindings, AutoBind types and Closers
    end
    File->>File: Create resolver, injectParams
    Note over File,Resolver: Injector parameters become already-resolved values
    File->>Resolver: resolve(result type)
    Resolver-->>File: output local name + accumulated statements
    File->>File: writeInjector(signature, statements, return)
    File-->>Gen: success
    Gen->>Gen: checkStrayBuildCalls
    Gen->>Gen: format.Source(g.bytes())
```

The collector describes **what is available**. The resolver decides **what is
needed and in which order**. `checkStrayBuildCalls` rejects `Build` calls outside
recognized injector declarations. No constructors execute in this sequence.

## 3. Recursive resolution, reuse and errors

Sources: [resolve, lookup and emission helpers](../internal/gen/resolve.go),
[type identity](../internal/gen/types.go).

```mermaid
sequenceDiagram
    participant Caller as inject or dependent binding
    participant R as resolver.resolve(T)
    participant State as done / stack / bindings / fns
    participant E as emit / emitCall
    Caller->>R: resolve(T)
    R->>State: Look up done[type identity]
    alt Value already available
        State-->>R: local name or injector parameter
        R-->>Caller: reuse name
    else Not resolved yet
        R->>State: Check current dependency stack
        alt T is already on stack
            R-->>Caller: dependency cycle with path
        else No cycle
            R->>State: lookup(T): explicit binding, then AutoBind
            alt No unique binding
                R-->>Caller: missing provider or ambiguous AutoBind diagnostic
            else Binding found
                R->>State: Push T onto stack
                R->>E: emit(binding)
                opt Function provider not already in fns
                    E->>R: resolveArgs: recursively resolve each input
                    R-->>E: input variable names
                    E->>E: checkCallSignature, writeCall, register cleanup
                    E->>State: Cache provider instantiation in fns
                end
                Note over E: Other forms emit a value, field, struct or slice,<br/>or resolve the target of a Bind
                E-->>R: produced name or error
                R->>State: Cache name in done on success, pop stack on return
                R-->>Caller: name or contextualized error
            end
        end
    end
```

- `done` caches by Go type identity, not printed type name.
- `fns` reuses calls by provider identity **including generic type arguments**.
- `stack` detects cycles along the current resolution path, not across all
  previously visited nodes.
- Dependency resolution happens before the dependent call is emitted; this
  produces constructor calls in dependency order.

## 4. Validating and replacing an output file

Source: [output.go](../internal/gen/output.go).

```mermaid
sequenceDiagram
    participant Run as gen.Run
    participant O as outputFile.write
    participant Check as readOutput
    participant FS as filesystem
    Run->>O: write queued source
    O->>Check: Revalidate destination
    Check->>FS: Lstat(path)
    alt Destination absent
        Check-->>O: allowed
    else Destination exists
        Check->>Check: Require regular file, reject symlink
        Check->>FS: ReadFile(path)
        Check->>Check: Require whyor generated header
        Check-->>O: old source or error
    end
    O->>FS: CreateTemp in destination directory
    O->>FS: Write(source)
    O->>FS: Chmod(0644)
    O->>FS: Close()
    O->>FS: Rename(temp, destination)
    Note over O,FS: Deferred cleanup runs on success and on error
    O->>FS: Close if needed, remove temporary path
    Note over O: Missing temporary path after rename is expected.<br/>Other cleanup errors are joined with the primary error
    O-->>Run: success or contextualized error
```

This avoids truncating an existing output before replacement. It is not a
promise of multi-file atomicity, crash durability, or protection against every
concurrent filesystem race.

## 5. Watch mode and error recovery

Sources: [runWatch](../cmd/whyor/main.go),
[watch and snapshot](../cmd/whyor/watch.go).

```mermaid
sequenceDiagram
    actor User
    participant CLI as runWatch
    participant W as watch
    participant S as snapshot
    participant Callback as generation callback
    participant Gen as gen.Run
    User->>CLI: whyor gen -w ./...
    CLI->>CLI: Install interrupt context, print startup diagnostic
    CLI->>W: watch(ctx, root, 500ms, callback)
    W->>S: Initial snapshot
    S-->>W: paths mapped to modification time and size
    W->>Callback: Initial generation
    Callback->>Gen: Run(write=true)
    Gen-->>Callback: paths or generation error
    Callback-->>W: Write result or diagnostic, return output error if any
    loop Each 500ms tick until interrupt or terminal error
        W->>S: snapshot(root)
        S-->>W: current snapshot or filesystem error
        opt Snapshot differs from previous one
            W->>W: Store current snapshot
            W->>Callback: Regenerate
            Callback->>Gen: Run(write=true)
            alt Generation succeeds
                Callback->>User: Print changed files
            else Generation fails
                Callback->>User: Print diagnostic, keep watching if delivery succeeds
            end
            Callback-->>W: nil or output-delivery error
        end
    end
    W-->>CLI: nil on cancellation, error on filesystem/output failure
```

`snapshot` ignores `*_gen.go`, hidden directories, `vendor` and `testdata`.
It detects changes using paths, modification times and sizes, not content hashes.
A file disappearing during polling is a normal deletion; a missing root is an
error. Failed generation is retried **on a subsequent detected change**, not
automatically on every tick.

## 6. Runtime constructors and reverse-order cleanup

This sequence is **application runtime**, not generator execution. It follows
the actual [generated lifecycle example](../examples/lifecycle/whyor_gen.go),
specifically `InitApp`. The caller invokes the returned cleanup; it is not
scheduled automatically by whyor.

```mermaid
sequenceDiagram
    actor Caller as Application
    participant Init as generated InitApp
    participant Providers as NewLegacy / NewCache / NewSocket / NewApp
    participant Resources as acquired resources
    Caller->>Init: InitApp(ctx, events)
    Init->>Providers: NewLegacy(events)
    Providers-->>Init: legacy + cleanup1
    Init->>Providers: NewCache(ctx, legacy, events)
    Providers-->>Init: cache + cleanup2
    Init->>Providers: NewSocket(cache, events)
    Providers-->>Init: socket with Close
    Init->>Providers: NewApp(legacy, cache, socket, events)
    alt NewApp fails
        Providers-->>Init: error
        Init->>Resources: socket.Close()
        Init->>Resources: cleanup2(context.WithoutCancel(ctx))
        Init->>Resources: cleanup1()
        Init-->>Caller: nil, nil, errors.Join(initial error, cleanup errors)
    else NewApp succeeds
        Providers-->>Init: app
        Init-->>Caller: app, cleanup, nil
        Note over Caller: Application runs, then chooses a shutdown context
        Caller->>Resources: returned cleanup(shutdownCtx)
        Resources->>Resources: socket.Close()
        Resources->>Resources: cleanup2(shutdownCtx)
        Resources->>Resources: cleanup1()
        Resources-->>Caller: joined cleanup errors
    end
```

This is the context-aware cleanup form. Legacy `func()` cleanup has no error
return. Cleanup order reverses successful acquisition order; it is sequential,
not concurrent. Without an injector context, generated rollback uses
`context.Background()`. The caller must provide shutdown deadlines as needed.

## 7. Graph display and unused-provider analysis

Source: [Show and Unused](../internal/gen/gen.go),
[exportGraph and record](../internal/gen/emit.go).

```mermaid
sequenceDiagram
    actor User
    participant CLI
    participant API as gen.Show or gen.Unused
    participant Gen as generate / inject
    participant Analysis as resolver + collector
    User->>CLI: whyor show -f json OR whyor unused
    CLI->>API: Show(...) or Unused(...)
    API->>API: load packages with whyor build tag
    API->>Gen: generate with analysis options
    Gen->>Analysis: Collect declarations and resolve injector result
    alt show
        Gen->>Gen: exportGraph instead of writeInjector
        Gen-->>API: tree text or graph document entries
        API->>API: Encode JSON if requested (schema version 1)
    else unused
        Gen->>Gen: record declared and called providers
        Gen-->>API: generated source (not written)
        API->>API: Compare declared vs used, sort unused diagnostics
    end
    API-->>CLI: graph or unused-provider list
    CLI-->>User: stdout, unused returns an error if list is nonempty
```

These commands share the resolver with generation; they do not execute
constructors or write generated output files. `show -f mermaid` exports a
**dependency graph**, not the source-code sequence diagrams in this document.

## Suggested source-reading order

1. [cmd/whyor/main.go](../cmd/whyor/main.go): dispatch and CLI error handling.
2. [internal/gen/gen.go](../internal/gen/gen.go): package loading and orchestration.
3. [internal/gen/collect.go](../internal/gen/collect.go): declarations into bindings.
4. [internal/gen/emit.go](../internal/gen/emit.go): injector orchestration and rendering.
5. [internal/gen/resolve.go](../internal/gen/resolve.go): recursive resolution and cleanup emission.
6. [internal/gen/output.go](../internal/gen/output.go): filesystem operations.
7. [examples/lifecycle/whyor_gen.go](../examples/lifecycle/whyor_gen.go): the result that actually runs.
