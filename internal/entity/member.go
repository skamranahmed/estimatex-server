package entity

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/skamranahmed/estimatex-server/internal/event"
)

type Member struct {
	ID          string
	Name        string
	Connection  *websocket.Conn
	RoomID      string
	IsRoomAdmin bool
}

// NewMember: creates a new member with a unique ID
func NewMember(memberName string, memberWebSocketConnection *websocket.Conn, roomID string, isRoomAdmin bool) *Member {
	return &Member{
		ID:          uuid.New().String(),
		Name:        memberName,
		Connection:  memberWebSocketConnection,
		RoomID:      roomID,
		IsRoomAdmin: isRoomAdmin,
	}
}

// ReadMessages: continuously reads messages from the WebSocket connection.
// It is a blocking operation, hence it must be run as a go routine.
func (m *Member) ReadMessages(room *Room, doneChannel chan bool) {
	log.Println("Starting a go-routine to read messages from the client: ", m.Name)

	defer func() {
		// remove the member from the room
		room.RemoveMember(m.ID)

		// close the member's websocket connection
		m.Connection.Close()

		log.Printf("Read go-routine for client %s shutting down", m.Name)

		select {
		case <-doneChannel:
			// doneChannel already closed, do nothing
			return
		default:
			// the doneChannel is open, close it so that it can notify the writing go routine
			close(doneChannel)
		}
	}()

	// continuously read messages from the member's websocket connection
	for {
		select {
		case <-doneChannel:
			/*
				exit the loop if the done channel is closed (indicating that the websocket connection is closed)
			*/
			return

		default:
			// read the message from the member's websocket connection
			_, payload, err := m.Connection.ReadMessage()
			if err != nil {
				if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
					// respond to the client's close message
					log.Printf("%+v initiated close for the room id: %+v\n", m.Name, m.RoomID)

					// if the connection is closed by the room admin, then
					// all other members also need to be removed from the room
					// also, their connection has to be closed
					if m.IsRoomAdmin {
						log.Printf("Admin of room id: %+v, closed the connection. Disconnecting all the other members.", m.RoomID)
						connectedMembers := room.GetMembers()
						for _, connectedMember := range connectedMembers {
							if connectedMember.ID != m.ID {
								fmt.Printf("Closing connection for client: %+v\n", connectedMember.Name)
								connectedMember.Connection.Close()
							}
						}
					}

					m.Connection.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "Server closing connection"))
					return
				}

				// handle the case where the client's connection is abruptly closed (close code 1006)
				if strings.Contains(err.Error(), "close 1006 (abnormal closure)") {
					log.Printf("Client's websocket connection was abruptly closed, error: %+v\n", err)
					return
				}

				log.Printf("Client closed the websocket connection, error: %+v\n", err)
				return
			}

			var receivedEvent event.Event
			err = json.Unmarshal(payload, &receivedEvent)
			if err != nil {
				log.Printf("Error unmarshalling the received event message from the client: %v", err)
				// TODO: Think: do I need to return here or continue here?
				return
			}

			// TODO: setup the logic to handle different types of WebSocket messsages as events
		}
	}
}
