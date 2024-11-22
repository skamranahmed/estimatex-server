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

	// Key: TicketID, Value: slice of Vote
	TicketVotesMap      map[string][]*Vote
	TicketVotesMapMutex sync.Mutex

	// Key: MemberID, Value: Vote
	MemberVoteMap map[string]*Vote
}

func (r *Room) SetupEventHandlers() {
	r.EventHandlers[event.EventJoinRoom] = r.JoinRoomEventHandler
	r.EventHandlers[event.EventBeginVoting] = r.BeginVotingEventHandler
	r.EventHandlers[event.EventMemberVoted] = r.MemberVotedEventHandler
	r.EventHandlers[event.EventRevealVotes] = r.RevealVotesEventHandler
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

func (r *Room) MemberVotedEventHandler(member *Member, receivedEvent event.Event) error {
	var memberVotedEventData event.MemberVotedEventData
	err := json.Unmarshal(receivedEvent.Data, &memberVotedEventData)
	if err != nil {
		log.Println("unable to handle MEMBER_VOTED event")
		return err
	}

	r.SaveTicketVote(member, memberVotedEventData.TicketID, memberVotedEventData.Vote)

	membersInRoom := r.GetMembers()

	for _, memberInRoom := range membersInRoom {
		memberInRoom.MessageChannel <- fmt.Sprintf("%v voted for the ticket id %v", member.Name, memberVotedEventData.TicketID)
	}

	if len(r.TicketVotesMap[memberVotedEventData.TicketID]) == r.GetRoomMembersCount() {
		r.SaveAllMemberVotes(memberVotedEventData.TicketID)

		for _, memberInRoom := range membersInRoom {
			if memberInRoom.IsRoomAdmin {
				messageToBeSentToAdminMember := fmt.Sprintf("✅ Voting has completed for the ticket id: %s\n> 👉 You will now be prompted for confirmation to reveal the votes.", memberVotedEventData.TicketID)
				memberInRoom.SendVotingCompletedEvent(messageToBeSentToAdminMember)
				memberInRoom.SendRevealVotesPromptEvent("", memberVotedEventData.TicketID)
				continue
			}
			messageToBeSentToNonAdminMember := fmt.Sprintf("✅ Voting has completed for the ticket id: %s\n> ⏳ Waiting for the admin to reveal the votes.", memberVotedEventData.TicketID)
			memberInRoom.SendVotingCompletedEvent(messageToBeSentToNonAdminMember)
		}
	}

	return nil
}

func (r *Room) RevealVotesEventHandler(member *Member, receivedEvent event.Event) error {
	var revealVotesEventData event.RevealVotesEventData
	err := json.Unmarshal(receivedEvent.Data, &revealVotesEventData)
	if err != nil {
		log.Println("unable to handle REVEAL_VOTES event")
		return err
	}

	// event received to reveal votesm broadcast a message to all participants, including the admin,
	// and reveal the votes for the given ticket ID
	memberVotesMapInterface := make(map[string]interface{}, len(r.MemberVoteMap))
	for memberID, vote := range r.MemberVoteMap {
		memberVotesMapInterface[memberID] = vote
	}

	for _, memberInRoom := range r.GetMembers() {
		if memberInRoom.IsRoomAdmin {
			memberInRoom.SendVotesRevealedEvent(revealVotesEventData.TicketID, memberVotesMapInterface)

			// send another prompt to the admin to enter the ticket id for the next vote
			memberInRoom.SendBeginVotingPromptEvent("📝 Enter the ticket id for which you want to start voting next:")
			continue
		}

		memberInRoom.SendVotesRevealedEvent(revealVotesEventData.TicketID, memberVotesMapInterface)

		// also send message to the member that they need to wait for the admin to begin voting for the next ticket
		memberInRoom.SendAwaitingAdminVoteStartEvent("⏳ Waiting for the admin to begin voting for next ticket")
	}

	// delete the TicketID entry from the TicketVotesMap
	delete(r.TicketVotesMap, revealVotesEventData.TicketID)

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

func (r *Room) SaveTicketVote(member *Member, ticketID string, voteValue string) {
	r.TicketVotesMapMutex.Lock()
	defer r.TicketVotesMapMutex.Unlock()

	r.TicketVotesMap[ticketID] = append(r.TicketVotesMap[ticketID], &Vote{
		Value:      voteValue,
		MemberID:   member.ID,
		MemberName: member.Name,
	})
}

func (r *Room) SaveAllMemberVotes(ticketID string) {
	for _, vote := range r.TicketVotesMap[ticketID] {
		r.MemberVoteMap[vote.MemberID] = &Vote{
			Value:      vote.Value,
			MemberID:   vote.MemberID,
			MemberName: vote.MemberName,
		}
	}
}
