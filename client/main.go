package main

import (
	"log"

	"github.com/gen2brain/beeep"
)

func main() {
	err := beeep.Notify(
		"Build Complete",
		"Your app finished successfully",
		"",
	)
	if err != nil {
		log.Fatalf("notification failed: %v", err)
	}
}
