package cmd

import (
	"log"
	"net/http"

	"github.com/skamranahmed/estimatex-server/internal/controller"
)

func Run() error {
	http.HandleFunc("/ws", controller.ServeWS)

	log.Println("Server is running on port 8080")
	return http.ListenAndServe(":8080", nil)
}
