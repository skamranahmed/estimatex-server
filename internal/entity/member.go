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
	ID             string
	Name           string
	Connection     *websocket.Conn
	RoomID         string
	IsRoomAdmin    bool
	MessageChannel chan string
}

// NewMember: creates a new member with a unique ID
func NewMember(memberName string, memberWebSocketConnection *websocket.Conn, roomID string, isRoomAdmin bool) *Member {
	return &Member{
		ID:             uuid.New().String(),
		Name:           memberName,
		Connection:     memberWebSocketConnection,
		RoomID:         roomID,
		IsRoomAdmin:    isRoomAdmin,
		MessageChannel: make(chan string),
	}
}

// ReadMessages: continuously reads messages from the WebSocket connection.
// It is a blocking operation, hence it must be run as a go routine.
func (m *Member) ReadMessages(room *Room, doneChannel chan bool) {
	log.Println("Starting a go-routine to read messages from the client: ", m.Name)

	defer func() {
		log.Println("Shutting down the read go-routine for the client: ", m.Name)

		// remove the member from the room
		room.RemoveMember(m.ID)

		// close the member's websocket connection
		m.Connection.Close()

		select {
		case <-doneChannel:
			// doneChannel already closed, do nothing
			return
		default:
			// the doneChannel is open, close it so that it can notify the `WriteMessages` method that is running as a go-routine
			close(doneChannel)
		}
	}()

	// continuously read messages from the member's websocket connection
	for {
		select {
		case <-doneChannel:
			/*
				Exit the loop if the done channel is closed (indicating that the websocket connection is closed).
				This signal is being read because the `WriteMessages` method which is running in another go-routine
				will close the `doneChannel` channel once the connection is broken, and once that happens, we also
				need to stop the `ReadMessages` method which is also running as a go-routine. If the `ReadMessages`
				go-routine is not exited, then it will lead to a go-routine leak.
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
				// TODO: Think: if the error has happened with the admin, do I need to terminate the connection for other members too?
				return
			}

			// logic to handle different types of WebSocket messsages as events
			err = room.HandleEvent(m, receivedEvent)
			if err != nil {
				log.Printf("Error while handling the received event %s", receivedEvent.Type)
				// TODO:
				// Think: do I need to return here or continue here?
				// Do I need to inform the client that something has gone wrong on the server?
				// Or should I simply close the connection?
				return
			}
		}
	}
}

// WriteMessages: sends messages to the WebSocket connection.
// It is a blocking operation, hence it must be run as a go routine.
func (m *Member) WriteMessages(doneChannel chan bool) {
	log.Println("Starting a go-routine to write messages to the client: ", m.Name)

	defer func() {
		log.Println("Shutting down the write go-routine for the client: ", m.Name)

		select {
		case <-doneChannel:
			// doneChannel already closed, do nothing
			return
		default:
			// the doneChannel is open, close it so that it can notify the `ReadMessages` method which is running as a go-routine
			close(doneChannel)
		}
	}()

	for {
		select {
		case messageToBeSentToMember, _ := <-m.MessageChannel:
			err := m.Connection.WriteMessage(websocket.TextMessage, []byte(messageToBeSentToMember))
			if err != nil {
				log.Printf("Error while sending message to the client, error: %+v\n", err)
				// in case of an error, make an early return and close the connection
				return
			}

		case <-doneChannel:
			/*
				Exit the loop if the done channel is closed (indicating that the websocket connection is closed).
				This signal is being read because the `ReadMessages` method which is running in another go-routine
				will close the `doneChannel` channel once the connection is broken, and once that happens, we also
				need to stop the `WriteMessages` method which is also running as a go-routine. If the `WriteMessages`
				go-routine is not exited, then it will lead to a go-routine leak.
			*/
			return
		}
	}
}

func (m *Member) SendCreateRoomEvent(roomID string) {
	createRoomEvent := event.CreateRoomEventData{
		RoomID: roomID,
	}
	createRoomEvenJsonData, _ := json.Marshal(createRoomEvent)
	eventToBeSent := event.Event{
		Type: string(event.EventCreateRoom),
		Data: json.RawMessage(createRoomEvenJsonData),
	}
	m.sendEvent(eventToBeSent)
}

func (m *Member) SendRoomJoinUpdatesEvent(message string) {
	roomJoinUpdatesEvent := event.RoomJoinUpdatesEventData{
		Message: message,
	}
	roomJoinUpdatesEventJsonData, _ := json.Marshal(roomJoinUpdatesEvent)
	eventToBeSent := event.Event{
		Type: string(event.EventRoomJoinUpdates),
		Data: json.RawMessage(roomJoinUpdatesEventJsonData),
	}
	m.sendEvent(eventToBeSent)
}

func (m *Member) SendRoomCapacityReachedEvent(message string) {
	roomCapacityReachedEvent := event.RoomCapacityReachedEventData{
		Message: message,
	}
	roomCapacityReachedEventJsonData, _ := json.Marshal(roomCapacityReachedEvent)
	eventToBeSent := event.Event{
		Type: string(event.EventRoomCapacityReached),
		Data: json.RawMessage(roomCapacityReachedEventJsonData),
	}
	m.sendEvent(eventToBeSent)
}

func (m *Member) SendBeginVotingPromptEvent(message string) {
	beginVotingPromptEvent := event.BeginVotingPromptEventData{
		Message: message,
	}
	beginVotingPromptEventJsonData, _ := json.Marshal(beginVotingPromptEvent)
	eventToBeSent := event.Event{
		Type: string(event.EventBeginVotingPrompt),
		Data: json.RawMessage(beginVotingPromptEventJsonData),
	}
	m.sendEvent(eventToBeSent)
}

func (m *Member) SendAskForVoteEvent(ticketId string) {
	askForVoteEvent := event.AskForVoteEventData{
		TicketID: ticketId,
	}
	askForVoteEventJsonData, _ := json.Marshal(askForVoteEvent)
	eventToBeSent := event.Event{
		Type: string(event.EventAskForVote),
		Data: json.RawMessage(askForVoteEventJsonData),
	}
	m.sendEvent(eventToBeSent)
}

func (m *Member) SendVotingCompletedEvent(message string) {
	votingCompletedEvent := event.VotingCompletedEventData{
		Message: message,
	}
	votingCompletedEventJsonData, _ := json.Marshal(votingCompletedEvent)
	eventToBeSent := event.Event{
		Type: string(event.EventVotingCompleted),
		Data: json.RawMessage(votingCompletedEventJsonData),
	}
	m.sendEvent(eventToBeSent)
}

func (m *Member) SendRevealVotesPromptEvent(message string, ticketId string) {
	revealVotesPromptEvent := event.RevealVotesPromptEventData{
		Message:  message,
		TicketID: ticketId,
	}
	revealVotesPromptEventJsonData, _ := json.Marshal(revealVotesPromptEvent)
	eventToBeSent := event.Event{
		Type: string(event.EventRevealVotesPrompt),
		Data: json.RawMessage(revealVotesPromptEventJsonData),
	}
	m.sendEvent(eventToBeSent)
}

func (m *Member) sendEvent(eventToBeSent event.Event) {
	jsonMessage, err := json.Marshal(eventToBeSent)
	if err != nil {
		log.Printf("unable to marshal message: %+v, error: %+v", eventToBeSent, err)
	}

	m.MessageChannel <- string(jsonMessage)
}
