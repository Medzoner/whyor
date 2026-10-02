// Command whyor generates dependency injection code.
package main

import (
	"fmt"
	"os"

	"github.com/Medzoner/whyor/internal/gen"
)

func main() {
	args := os.Args[1:]
	if len(args) == 0 || args[0] != "gen" && args[0] != "check" {
		fmt.Fprintln(os.Stderr, "usage: whyor gen|check [packages]")
		os.Exit(2)
	}
	write := args[0] == "gen"
	pats := args[1:]
	if len(pats) == 0 {
		pats = []string{"./..."}
	}
	files, err := gen.Run(".", pats, write)
	if err != nil {
		fmt.Fprintln(os.Stderr, "whyor:", err)
		os.Exit(1)
	}
	for _, f := range files {
		fmt.Println(f)
	}
	if !write && len(files) > 0 {
		os.Exit(1) // generated code is stale
	}
}
