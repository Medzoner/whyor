#!/usr/bin/env bash
set -euo pipefail

# Execute with bash docs/demo/run.sh from any working directory.
# All generated demo files live in an owned temporary directory, never in the checkout.
root=$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd -P)
pause=${DEMO_PAUSE:-12}
case "$pause" in
  ''|*[!0-9]*) printf 'DEMO_PAUSE must be a non-negative integer\n' >&2; exit 2 ;;
esac
tmp=$(mktemp -d "${TMPDIR:-/tmp}/whyor-demo.XXXXXX")
tmp=$(cd "$tmp" && pwd -P)
trap 'rm -rf "$tmp"' EXIT
mkdir "$tmp/app"

printf 'Preparing the demo from this checkout (Go 1.27.1+ required)...\n'
go -C "$root" build -o "$tmp/whyor" ./cmd/whyor
cp "$root/examples/showcase/"{app.go,wire.go,main.go,app_test.go} "$tmp/app/"
go -C "$tmp/app" mod init example.com/whyor-demo
go -C "$tmp/app" mod edit -go=1.27.1 \
  -require=github.com/Medzoner/whyor@v0.11.0 \
  -replace="github.com/Medzoner/whyor=$root"

cd "$tmp/app"
screen() {
  if [ -t 1 ]; then printf '\033[2J\033[H'; fi
  printf '\n── %s ──\n\n' "$1"
}
wait_for_viewer() { sleep "$pause"; }

screen '1 / Ordinary constructors. No container API.'
cat app.go
wait_for_viewer

screen '2 / Declare the wiring with a build tag.'
cat wire.go
wait_for_viewer

screen '3 / Generate Go you can inspect and commit.'
printf '$ whyor gen .\n'
"$tmp/whyor" gen .
cat whyor_gen.go
wait_for_viewer

screen '4 / See the dependency graph.'
printf '$ whyor show .\n'
"$tmp/whyor" show .
printf '\n$ whyor show -f json . > dependencies.json\n'
"$tmp/whyor" show -f json . > dependencies.json
wait_for_viewer

screen '5 / Run ordinary Go, including cleanup.'
printf '$ go run .\n'
go run .
printf '\n$ go test .\n'
go test .
printf '\n$ whyor check .\n'
"$tmp/whyor" check .
wait_for_viewer

printf '\nDemo complete. Temporary workspace removed on exit.\n'
