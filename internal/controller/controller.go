package controller

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/skamranahmed/estimatex-server/internal/api"
	"github.com/skamranahmed/estimatex-server/internal/entity"
	"github.com/skamranahmed/estimatex-server/internal/session"
)

var (
	websocketUpgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	sessionManager = session.DefaultManager
)

func ServeWS(w http.ResponseWriter, r *http.Request) {
	// upgrading the HTTP request to a websocket request
	wsConnection, err := websocketUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	// connection established

	actionValue, clientName, err := validateRequest(r)
	if err != nil {
		api.SendErrorResponse(wsConnection, err.Error())
		return
	}

	isRoomAdmin := false

	if actionValue == string(session.ActionCreateRoom) {
		maxRoomCapacityString := strings.TrimSpace(r.URL.Query().Get("max_room_capacity"))
		maxRoomCapacityInteger, err := strconv.Atoi(maxRoomCapacityString)
		if err != nil {
			log.Printf("[BAD_REQUEST_ERROR]: Got invalid value for max_room_capacity, error: %+v\n", err)
			api.SendErrorResponse(wsConnection, "invalid max_room_capacity value provided")
			return
		}

		// the client who creates the room is the room admin
		isRoomAdmin = true

		// create a new room
		room := sessionManager.CreateRoom(maxRoomCapacityInteger)

		// create a new client (i.e member)
		member := entity.NewMember(clientName, wsConnection, room.ID, isRoomAdmin)

		// add member to the room
		room.AddMember(member)
	}

	if actionValue == string(session.ActionJoinRoom) {
		roomID := strings.TrimSpace(r.URL.Query().Get("room_id"))

		fmt.Println(roomID)

		// TODO:
		/*
			1. Check if the room with the provided roomID exists or not
			2. If the room exists, add the member (client) to the room
			   but if the room's max capacity has already been reached,
			   then we must NOT add the member to the room, rather throw an error
			3. Create a new client (i.e member)
			4. Add the member to the room
		*/
	}

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
