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
