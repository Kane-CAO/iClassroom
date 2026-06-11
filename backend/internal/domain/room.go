// Package domain defines the core entities of the iClassroom MVP. These are
// plain data structures plus small value-validation helpers; they contain no
// persistence or HTTP logic (that belongs to repository / handler layers).
//
// JSON tags use camelCase to match the API contract (docs/api.md). Secret or
// server-internal fields are tagged `json:"-"` so they are never serialized to
// clients by accident.
package domain

import "time"

// RoomStatus is the lifecycle state of a room.
type RoomStatus string

const (
	RoomStatusCreated RoomStatus = "created"
	RoomStatusActive  RoomStatus = "active"
	RoomStatusEnded   RoomStatus = "ended"
)

// Valid reports whether s is a known room status.
func (s RoomStatus) Valid() bool {
	switch s {
	case RoomStatusCreated, RoomStatusActive, RoomStatusEnded:
		return true
	default:
		return false
	}
}

// Room is a classroom created by a teacher. It owns groups, students and tasks.
type Room struct {
	ID               int64      `json:"-"`
	RoomCode         string     `json:"roomCode"`
	Title            string     `json:"title"`
	Status           RoomStatus `json:"status"`
	GroupCount       int        `json:"groupCount"`
	GroupCapacity    int        `json:"groupCapacity"`
	AllowChooseGroup bool       `json:"allowChooseGroup"`
	// TeacherToken is the room-level management credential. Secret — never
	// serialize it except on the room-creation response that mints it.
	TeacherToken string     `json:"-"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
	EndedAt      *time.Time `json:"endedAt,omitempty"`
}
