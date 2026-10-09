// Command whyor generates dependency injection code.
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"time"

	"github.com/Medzoner/whyor/internal/gen"
)

const usage = `usage:
  whyor gen   [-w] [packages]   write whyor_gen.go files (-w: regenerate on change)
  whyor check [packages]   exit 1 if generated files are stale
  whyor show  [-f tree|mermaid|dot|json] [packages]   print each injector's dependencies
  whyor unused [packages]  list providers no injector calls (exit 1 if any)
  whyor init  [dir]        create a wire.go skeleton`

func main() {
	if err := run(os.Args[1:]); err != nil {
		if reportErr := reportError(os.Stderr, err); reportErr != nil {
			// There is no reliable diagnostic destination left. Signal that the
			// command failed and its diagnostic could not be written.
			os.Exit(2)
		}
		os.Exit(1)
	}
}

func run(args []string) error {
	return runWithWriters(args, os.Stdout, os.Stderr)
}

func runWithWriters(args []string, out, diagnostic io.Writer) error {
	if len(args) == 0 {
		return fmt.Errorf("%s", usage)
	}
	cmd, rest := args[0], args[1:]
	switch cmd {
	case "init":
		dir := "."
		if len(rest) > 0 {
			dir = rest[0]
		}
		path, err := initFile(dir)
		if err != nil {
			return err
		}
		return writeLines(out, []string{path})
	case "show":
		if len(rest) == 0 {
			rest = []string{"./..."}
		}
		format := "tree"
		if len(rest) >= 2 && rest[0] == "-f" {
			format, rest = rest[1], rest[2:]
		}
		if len(rest) == 0 {
			rest = []string{"./..."}
		}
		tree, err := gen.Show(".", rest, format)
		if err != nil {
			return fmt.Errorf("show dependencies: %w", err)
		}
		if _, err := io.WriteString(out, tree); err != nil {
			return fmt.Errorf("write dependency graph: %w", err)
		}
		return nil
	case "unused":
		if len(rest) == 0 {
			rest = []string{"./..."}
		}
		list, err := gen.Unused(".", rest)
		if err != nil {
			return fmt.Errorf("check unused providers: %w", err)
		}
		if err := writeLines(out, list); err != nil {
			return err
		}
		if len(list) > 0 {
			return fmt.Errorf("%d unused provider(s)", len(list))
		}
		return nil
	case "gen", "check":
		if cmd == "gen" && len(rest) > 0 && rest[0] == "-w" {
			return runWatch(rest[1:], out, diagnostic)
		}
		if len(rest) == 0 {
			rest = []string{"./..."}
		}
		files, err := gen.Run(".", rest, cmd == "gen")
		if err != nil {
			return fmt.Errorf("%s packages: %w", cmd, err)
		}
		if err := writeLines(out, files); err != nil {
			return err
		}
		if cmd == "check" && len(files) > 0 {
			return fmt.Errorf("%d generated file(s) are stale; run whyor gen", len(files))
		}
		return nil
	}
	return fmt.Errorf("unknown command %q\n%s", cmd, usage)
}

func runWatch(pkgs []string, out, diagnostic io.Writer) error {
	if len(pkgs) == 0 {
		pkgs = []string{"./..."}
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	if _, err := fmt.Fprintln(diagnostic, "whyor: watching for changes, Ctrl+C to stop"); err != nil {
		return fmt.Errorf("write watch startup diagnostic: %w", err)
	}
	err := watch(ctx, ".", 500*time.Millisecond, func() error {
		files, err := gen.Run(".", pkgs, true)
		if err != nil {
			// Generation errors remain recoverable, unless their diagnostic cannot
			// be delivered. Broken output is terminal, not silently retried.
			return reportError(diagnostic, err)
		}
		return writeLines(out, files)
	})
	if err != nil {
		return fmt.Errorf("watch packages %q: %w", pkgs, err)
	}
	return nil
}

func writeLines(w io.Writer, lines []string) error {
	for _, line := range lines {
		if _, err := fmt.Fprintln(w, line); err != nil {
			return fmt.Errorf("write command output: %w", err)
		}
	}
	return nil
}

func reportError(w io.Writer, cause error) error {
	if _, err := fmt.Fprintln(w, "whyor:", cause); err != nil {
		return fmt.Errorf("write command diagnostic: %w", errors.Join(cause, err))
	}
	return nil
}
