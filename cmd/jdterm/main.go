package main

import (
	"log"

	"jdterm/internal/app"
)

func main() {

	jdterm := app.New()

	if err := jdterm.Run(); err != nil {
		log.Fatal(err)
	}
}
