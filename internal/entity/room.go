package entity

import (
	"log"
	"sync"

	"github.com/skamranahmed/estimatex-server/internal/event"
)

type EventHanlder func(member *Member, event event.Event) error

type Room struct {
	ID            string
	MaxCapacity   int
	EventHandlers map[event.EventType]EventHanlder

	// Key: MemberID, Value: *Member
	Members sync.Map
}

func (r *Room) SetupEventHandlers() {
	// TODO: setup event handler functions
}

func (r *Room) HandleEvent(member *Member, receivedEvent event.Event) error {
	if event.IsIncomingEventTypeValid(receivedEvent.Type) {
		eventHandler, ok := r.EventHandlers[event.EventType(receivedEvent.Type)]
		if ok {
			err := eventHandler(member, receivedEvent)
			if err != nil {
				return err
			}
			return nil
		}

		log.Printf("the handler for the %+v event is not set", receivedEvent.Type)
		return event.EventHandlerNotSetError
	}

	log.Printf("the %+v event is not supported", receivedEvent.Type)
	return event.EventNotSupportedError
}

func (r *Room) AddMember(member *Member) {
	r.Members.Store(member.ID, member)
}

func (r *Room) GetRoomMembersCount() int {
	count := 0

	r.Members.Range(func(key interface{}, value interface{}) bool {
		count++
		return true
	})

	return count
}

func (r *Room) RemoveMember(memberID string) {
	r.Members.Delete(memberID)
}

func (r *Room) GetMembers() []*Member {
	var members []*Member

	r.Members.Range(func(key interface{}, value interface{}) bool {
		member, ok := value.(*Member)
		if ok {
			members = append(members, member)
		}
		return true
	})

	return members
}
