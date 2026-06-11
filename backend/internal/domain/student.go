package domain

import "time"

// Student is a participant who joined a room with a nickname. Students do not
// register accounts; identity is carried by ClientToken.
//
// Nickname is unique within a room (enforced by a DB unique constraint on
// room_id + nickname).
type Student struct {
	ID       int64  `json:"studentId"`
	RoomID   int64  `json:"-"`
	GroupID  int64  `json:"groupId"`
	Nickname string `json:"nickname"`
	// ClientToken is the student's session credential. Secret — never serialize
	// it except on the join response that mints it.
	ClientToken string    `json:"-"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"-"`
}
