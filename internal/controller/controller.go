package controller

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/skamranahmed/estimatex-server/internal/api"
	"github.com/skamranahmed/estimatex-server/internal/session"
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

	actionValue, clientName, err := validateRequest(r)
	if err != nil {
		api.SendErrorResponse(wsConnection, err.Error())
		return
	}

	fmt.Println(actionValue)
	fmt.Println(clientName)

}

func validateRequest(r *http.Request) (actionValue string, clientName string, err error) {
	actionValue = strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("action")))

	if !session.IsActionValid(actionValue) {
		log.Printf("[BAD_REQUEST_ERROR]: Got invalid action value: %+v\n", actionValue)
		return "", "", fmt.Errorf("invalid action value: %s", actionValue)
	}

	clientName = strings.TrimSpace(r.URL.Query().Get("name"))
	if clientName == "" {
		log.Printf("[BAD_REQUEST_ERROR]: Got an empty client name\n")
		return "", "", fmt.Errorf("name cannot be empty")
	}

	return actionValue, clientName, nil
}
