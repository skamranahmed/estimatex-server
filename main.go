package main

import (
	"log"
	"os"

	"github.com/skamranahmed/estimatex-server/cmd"
)

func main() {
	err := cmd.Run()
	if err != nil {
		log.Printf("Error during server startup: %+v", err)
		os.Exit(1)
	}
}
