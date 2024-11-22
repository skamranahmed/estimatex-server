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

	// the `done` channel is used for the communication between the websocket reading and websocket writing goroutine
	// and to coordinate the termination of each other
	done := make(chan bool)

	var member *entity.Member
	var room *entity.Room

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
		room = sessionManager.CreateRoom(maxRoomCapacityInteger)

		// create a new client (i.e member)
		member = entity.NewMember(clientName, wsConnection, room.ID, isRoomAdmin)

		// add member to the room
		room.AddMember(member)
	}

	if actionValue == string(session.ActionJoinRoom) {
		roomID := strings.TrimSpace(r.URL.Query().Get("room_id"))

		// check if the room with the provided roomID exists or not
		room = sessionManager.FindRoom(roomID)
		if room == nil {
			log.Printf("[BAD_REQUEST_ERROR]: Trying to join a room that doesn't exist")
			errMessage := fmt.Sprintf("⚠️ Room id: %s does not exist. Please check the room id and try again.", roomID)
			api.SendErrorResponse(wsConnection, errMessage)
			return
		}

		/*
			if the room exists, add the member (client) to the room
			but if the room's max capacity has already been reached,
			then we must NOT add the member to the room, rather throw an error
		*/
		if room.GetRoomMembersCount() >= room.MaxCapacity {
			log.Printf("[BAD_REQUEST_ERROR]: Trying to join a room that is already at maximum capacity")
			errMessage := fmt.Sprintf("😢 Room %s is full. You cannot join. Please try again later or choose a different room.", roomID)
			api.SendErrorResponse(wsConnection, errMessage)
			return
		}

		// create a new client (i.e member)
		member = entity.NewMember(clientName, wsConnection, roomID, isRoomAdmin)

		// add member to the room
		room.AddMember(member)
	}

	// start a go routine which would continuously read messages from the client (member)
	go member.ReadMessages(room, done)

	// start a go routine which would write messages to the client (member)
	go member.WriteMessages(done)

	if actionValue == string(session.ActionCreateRoom) {
		// inform the client (member) that the room has been created
		member.SendCreateRoomEvent(room.ID)
	}

	return
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
