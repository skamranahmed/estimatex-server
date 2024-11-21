package session

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
