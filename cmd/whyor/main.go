// Command whyor generates dependency injection code.
package main

import (
	"fmt"
	"os"

	"github.com/Medzoner/whyor/internal/gen"
)

const usage = `usage:
  whyor gen   [packages]   write whyor_gen.go files
  whyor check [packages]   exit 1 if generated files are stale
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
	case "gen", "check":
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
