package entity

import (
	"encoding/json"
	"fmt"
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
	r.EventHandlers[event.EventJoinRoom] = r.JoinRoomEventHandler
	r.EventHandlers[event.EventBeginVoting] = r.BeginVotingEventHandler
}

func (r *Room) JoinRoomEventHandler(member *Member, receivedEvent event.Event) error {
	/*
		When a member joins the room, the following things need to be done:

		1. The member needs to be informed that they are now present inside the room which they intended to join

		2. The member needs to be informed about the other members who are already present in the room

		3. The member's name needs to be logged which would indicate that they have joined the room

		4. The existing members need to be informed that a new member has joined the room
	*/

	// satisfies requirement 1
	messageToBeSentToMember := fmt.Sprintf("🧠 You are now present in the room: %+v", r.ID)
	member.SendRoomJoinUpdatesEvent(messageToBeSentToMember)

	// satisfies requirement 2
	alreadyPresentMembers := r.GetMembers()
	for _, alreadyPresentMember := range alreadyPresentMembers {
		if alreadyPresentMember.ID != member.ID {
			messageToBeSentToMember := fmt.Sprintf("👤 %s joined", alreadyPresentMember.Name)

			if alreadyPresentMember.IsRoomAdmin {
				messageToBeSentToMember = fmt.Sprintf("👑👤 %s (ADMIN) joined", alreadyPresentMember.Name)
			}

			member.SendRoomJoinUpdatesEvent(messageToBeSentToMember)
		}
	}

	// satisfies requirement 3
	messageToBeSentToMember = fmt.Sprintf("👤 %s joined", member.Name)
	if member.IsRoomAdmin {
		messageToBeSentToMember = fmt.Sprintf("👑👤 %s (ADMIN) joined", member.Name)
	}
	member.SendRoomJoinUpdatesEvent(messageToBeSentToMember)

	// satisfies requirement 4
	for _, alreadyPresentMember := range alreadyPresentMembers {
		if alreadyPresentMember.ID != member.ID {
			// a member who joins a room later cannot be an admin, hence only a single message type is needed here
			messageToBeSentToAlreadyPresentMember := fmt.Sprintf("👤 %s joined", member.Name)
			alreadyPresentMember.SendRoomJoinUpdatesEvent(messageToBeSentToAlreadyPresentMember)
		}
	}

	// when a room's capacity is reached, the voting for the ticket needs to begin
	if r.GetRoomMembersCount() == r.MaxCapacity {
		for _, member := range alreadyPresentMembers {
			if member.IsRoomAdmin {
				member.SendRoomCapacityReachedEvent("🟢 Room capacity reached. You will now be prompted to begin voting.")
				member.SendBeginVotingPromptEvent("📝 Enter the ticket id for which you want to start voting:")
				continue
			}
			member.SendRoomCapacityReachedEvent("🟢 Room capacity reached. Waiting for the admin to begin voting.")
		}
	}

	return nil
}

func (r *Room) BeginVotingEventHandler(member *Member, receivedEvent event.Event) error {
	var beginVotingEventData event.BeginVotingEventData
	err := json.Unmarshal(receivedEvent.Data, &beginVotingEventData)
	if err != nil {
		log.Printf("unable to handle %+v event\n", receivedEvent.Type)
		return err
	}

	// we got the ticket id for which the admin wants to begin voting
	// now, we need to send a broadcast message to everyone in the room to ask for their vote
	for _, member := range r.GetMembers() {
		member.SendAskForVoteEvent(beginVotingEventData.TicketID)
	}

	return nil
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
