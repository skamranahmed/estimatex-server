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
	// Incoming Events
	EventJoinRoom    EventType = "JOIN_ROOM"
	EventBeginVoting EventType = "BEGIN_VOTING"
	EventMemberVoted EventType = "MEMBER_VOTED"

	// Outgoing Events
	EventRoomJoinUpdates     EventType = "ROOM_JOIN_UPDATES"
	EventRoomCapacityReached EventType = "ROOM_CAPACITY_REACHED"
	EventBeginVotingPrompt   EventType = "BEGIN_VOTING_PROMPT"
	EventAskForVote          EventType = "ASK_FOR_VOTE"
	EventVotingCompleted     EventType = "VOTING_COMPLETED"
	EventRevealVotesPrompt   EventType = "REVEAL_VOTES_PROMPT"

	// Incoming + Outgoing Events
	EventCreateRoom EventType = "CREATE_ROOM"
)

func IsIncomingEventTypeValid(input string) bool {
	switch EventType(input) {
	case EventCreateRoom, EventJoinRoom, EventBeginVoting, EventMemberVoted:
		return true
	default:
		return false
	}
}

// CreateRoomEventData represents data specific to the "CREATE_ROOM" event
type CreateRoomEventData struct {
	RoomID string `json:"room_id"`
}

// RoomJoinUpdatesEventData represents data specific to the "ROOM_JOIN_UPDATES" event
type RoomJoinUpdatesEventData struct {
	Message string `json:"message"`
}

// RoomCapacityReachedEventData represents data specific to the "ROOM_CAPACITY_REACHED" event
type RoomCapacityReachedEventData struct {
	Message string `json:"message"`
}

// BeginVotingPromptEventData represents data specific to the "BEGIN_VOTING_PROMPT" event
type BeginVotingPromptEventData struct {
	Message string `json:"message"`
}

// BeginVotingEventData represents data specific to the "BEGIN_VOTING" event
type BeginVotingEventData struct {
	TicketID string `json:"ticket_id"`
}

// AskForVoteEventData represents data specific to the "ASK_FOR_VOTE" event
type AskForVoteEventData struct {
	TicketID string `json:"ticket_id"`
}

// MemberVotedEventData represents data specific to the "MEMBER_VOTED" event
type MemberVotedEventData struct {
	TicketID string `json:"ticket_id"`
	Vote     string `json:"vote"`
}

// VotingCompletedEventData represents data specific to the "VOTING_COMPLETED" event
type VotingCompletedEventData struct {
	TicketID string `json:"ticket_id"`
	Message  string `json:"message"`
}

// RevealVotesPromptEventData represents data specific to the "REVEAL_VOTES_PROMPT" event
type RevealVotesPromptEventData struct {
	TicketID string `json:"ticket_id"`
	Message  string `json:"message"`
}
