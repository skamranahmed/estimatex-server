package cmd

import (
	"log"
	"net/http"
)

func Run() error {
	log.Println("Server is running on port 8080")
	return http.ListenAndServe("localhost:8080", nil)
}
