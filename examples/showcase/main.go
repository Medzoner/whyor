package main

import (
	"fmt"
	"log"
)

func main() {
	app, cleanup, err := InitApp("Go")
	if err != nil {
		log.Fatal(err)
	}
	defer cleanup()
	fmt.Println(app.Store.Message())
}
