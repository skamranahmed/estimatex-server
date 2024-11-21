package controller

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var (
	websocketUpgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
)

func ServeWS(w http.ResponseWriter, r *http.Request) {
	// upgrading the HTTP request to a websocket request
	wsConnection, err := websocketUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	// connection established

	fmt.Println(wsConnection.LocalAddr())
}
