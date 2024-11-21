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
