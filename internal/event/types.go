package event

import (
	"encoding/json"
	"errors"
)

var (
	EventHandlerNotSetError = errors.New("handler for this event type is not present")
	EventNotSupportedError  = errors.New("event is not supported")
)

type Event struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type EventType string

const (
	EventCreateRoom EventType = "CREATE_ROOM"
)

func IsIncomingEventTypeValid(input string) bool {
	switch EventType(input) {
	case EventCreateRoom:
		return true
	default:
		return false
	}
}

// CreateRoomEventData represents data specific to the "CREATE_ROOM" event
type CreateRoomEventData struct {
	RoomID string `json:"room_id"`
}
