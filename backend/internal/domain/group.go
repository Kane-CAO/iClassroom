package domain

import "time"

// Group is a team within a room. Students join exactly one group.
//
// ScoreTotal is a redundant running total kept in sync transactionally when a
// submission is graded (see grading rules in CLAUDE.md / docs/api.md). The live
// member count (currentCount in API responses) is derived via COUNT(students)
// rather than stored, to avoid drift.
type Group struct {
	ID         int64     `json:"groupId"`
	RoomID     int64     `json:"-"`
	GroupName  string    `json:"groupName"`
	Capacity   int       `json:"capacity"`
	ScoreTotal int       `json:"scoreTotal"`
	CreatedAt  time.Time `json:"-"`
	UpdatedAt  time.Time `json:"-"`
}
