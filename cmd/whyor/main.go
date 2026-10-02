// Command whyor generates dependency injection code.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/Medzoner/whyor/internal/gen"
)

const usage = `usage:
  whyor gen   [-w] [packages]   write whyor_gen.go files (-w: regenerate on change)
  whyor check [packages]   exit 1 if generated files are stale
  whyor show  [-f tree|mermaid|dot] [packages]   print each injector's dependencies
  whyor unused [packages]  list providers no injector calls (exit 1 if any)
  whyor init  [dir]        create a wire.go skeleton`

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "whyor:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
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
		if err == nil {
			fmt.Println(path)
		}
		return err
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
		fmt.Print(tree)
		return err
	case "unused":
		if len(rest) == 0 {
			rest = []string{"./..."}
		}
		list, err := gen.Unused(".", rest)
		for _, l := range list {
			fmt.Println(l)
		}
		if err == nil && len(list) > 0 {
			err = fmt.Errorf("%d unused provider(s)", len(list))
		}
		return err
	case "gen", "check":
		if cmd == "gen" && len(rest) > 0 && rest[0] == "-w" {
			return runWatch(rest[1:])
		}
		if len(rest) == 0 {
			rest = []string{"./..."}
		}
		files, err := gen.Run(".", rest, cmd == "gen")
		for _, f := range files {
			fmt.Println(f)
		}
		if err == nil && cmd == "check" && len(files) > 0 {
			err = fmt.Errorf("%d generated file(s) are stale; run whyor gen", len(files))
		}
		return err
	}
	return fmt.Errorf("unknown command %q\n%s", cmd, usage)
}

func runWatch(pkgs []string) error {
	if len(pkgs) == 0 {
		pkgs = []string{"./..."}
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	fmt.Fprintln(os.Stderr, "whyor: watching for changes, Ctrl+C to stop")
	watch(ctx, ".", 500*time.Millisecond, func() {
		files, err := gen.Run(".", pkgs, true)
		for _, f := range files {
			fmt.Println(f)
		}
		if err != nil {
			fmt.Fprintln(os.Stderr, "whyor:", err)
		}
	})
	return nil
}
