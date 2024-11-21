package entity

import (
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
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
