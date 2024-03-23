package main

import (
	"log"

	"github.com/rahul0tripathi/framecoiner/app"
)

func main() {
	err := app.Run()
	if err != nil {
		log.Fatal(err)
	}
}
