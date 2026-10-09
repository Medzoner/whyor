package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
)

func main() {
	if err := runApplication(os.Stdout); err != nil {
		log.Fatal(err)
	}
}

func runApplication(out io.Writer) (err error) {
	app, cleanup, err := InitApp("Go")
	if err != nil {
		return fmt.Errorf("initialize example app: %w", err)
	}
	defer func() {
		cleanup()
		if _, printErr := fmt.Fprintln(out, "cleanup: memory store"); printErr != nil {
			err = errors.Join(err, fmt.Errorf("print example cleanup: %w", printErr))
		}
	}()
	if _, err := fmt.Fprintln(out, app.Store.Message()); err != nil {
		return fmt.Errorf("print example greeting: %w", err)
	}
	return nil
}
