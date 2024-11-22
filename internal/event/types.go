package event

import "encoding/json"

type Event struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type EventType string

const (
	EventCreateRoom EventType = "CREATE_ROOM"
)

// CreateRoomEventData represents data specific to the "CREATE_ROOM" event
type CreateRoomEventData struct {
	RoomID string `json:"room_id"`
}
