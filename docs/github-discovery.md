# GitHub discovery setup

## Repository metadata

These are proposed values for the repository's **About** panel. Committing this
document does not apply repository metadata.

**Description**

```text
Compile-time dependency injection for Go. Generates plain Go with generics, graph exports and explicit cleanup. Inspired by Wire.
```

Avoid "official successor" or "production-ready": this is an independent,
pre-1.0 project.

**Website**

```text
https://pkg.go.dev/github.com/Medzoner/whyor
```

Use the working documentation URL. Switch to the launch article only after
`https://www.medzoner.com/articles/after-wire-whyor` is actually deployed and
returns the expected article, not a 404.

**Topics**

```text
go, golang, dependency-injection, code-generation, wire, generics, developer-tools, compile-time
```

Apply through the gear beside **About** on the repository home page. If you
prefer the CLI, authenticate with `gh auth login`, then:

```sh
gh repo edit Medzoner/whyor \
  --description 'Compile-time dependency injection for Go. Generates plain Go with generics, graph exports and explicit cleanup. Inspired by Wire.' \
  --homepage 'https://pkg.go.dev/github.com/Medzoner/whyor' \
  --add-topic 'go,golang,dependency-injection,code-generation,wire,generics,developer-tools,compile-time'
```

## Social preview

Upload [social-preview.png](assets/social-preview.png) in:

**Repository Settings → General → Social preview → Edit → Upload an image**.

The image is 1280×640 (2:1), designed for GitHub link previews. Its editable
source is [social-preview.svg](assets/social-preview.svg). It contains no
unverified performance or adoption claims.

Regenerate with ffmpeg built with SVG decoding support:

```sh
ffmpeg -hide_banner -loglevel error -i docs/assets/social-preview.svg \
  -frames:v 1 docs/assets/social-preview.png
```

Uploading this asset to GitHub settings is a separate administrative action.
The README does not automatically configure the repository's social preview.

## Documentation and contribution entry points

- Verify the current release on pkg.go.dev, not only the module's cached proxy entry.
- Keep the README's installation, demo, migration and contribution links visible.
- Use the bug and feature forms to collect minimal reproductions and concrete use cases.
- Keep CHANGELOG.md as the source of release notes; do not tag each cosmetic edit.
- Separate deployed article claims from unpublished source changes.

## Access boundaries

`gh` needs authentication and permission to edit this repository. Metadata
and social-image settings cannot be applied by simply pushing these files.
Never paste tokens into issues, configuration files or documentation.
