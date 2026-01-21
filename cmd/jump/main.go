package main

import (
	"log"

	"github.com/MrSnakeDoc/jump/internal/app"
)

func main() {
	if err := app.New().Run(); err != nil {
		log.Fatalf("❌ jump failed to start: %v", err)
	}
}
