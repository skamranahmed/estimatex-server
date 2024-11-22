package entity

import "sync"

type Room struct {
	ID          string
	MaxCapacity int

	// Key: MemberID, Value: *Member
	Members sync.Map
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
