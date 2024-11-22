package session

import (
	"sync"
	"time"

	"github.com/skamranahmed/estimatex-server/internal/entity"
	"github.com/skamranahmed/estimatex-server/internal/event"
	"golang.org/x/exp/rand"
)

func init() {
	rand.Seed(uint64(time.Now().UnixNano()))
}

type Action string

const (
	ActionCreateRoom Action = "CREATE_ROOM"
	ActionJoinRoom   Action = "JOIN_ROOM"
)

func IsActionValid(input string) bool {
	switch Action(input) {
	case ActionCreateRoom, ActionJoinRoom:
		return true
	default:
		return false
	}
}

const (
	roomIDLength = 6
	letters      = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

var DefaultManager = NewManager()

// SessionManager: manages active rooms within a session, allowing thread-safe access and operations
type SessionManager struct {
	// rooms stores all active rooms in a concurrent-safe map, accessible by roomID
	rooms sync.Map
}

func NewManager() *SessionManager {
	sessionManager := &SessionManager{}
	return sessionManager
}

func (s *SessionManager) CreateRoom(maxCapacity int) *entity.Room {
	room := &entity.Room{
		ID:            s.generateRoomID(),
		MaxCapacity:   maxCapacity,
		EventHandlers: make(map[event.EventType]entity.EventHanlder),
	}
	room.SetupEventHandlers()
	s.rooms.Store(room.ID, room)
	return room
}

func (s *SessionManager) FindRoom(roomID string) *entity.Room {
	room, ok := s.rooms.Load(roomID)
	if !ok {
		return nil
	}
	return room.(*entity.Room)
}

func (s *SessionManager) generateRoomID() string {
	for {
		roomID := s.randomString(roomIDLength)
		if !s.doesRoomAlreadyExist(string(roomID)) {
			return roomID
		}
	}
}

func (m *SessionManager) randomString(length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func (m *SessionManager) doesRoomAlreadyExist(roomID string) bool {
	_, ok := m.rooms.Load(roomID)
	return ok
}
