// Package websocket implements the real-time fan-out layer for iClassroom.
//
// It is intentionally decoupled from the business services: services never
// import this package's connection types, they only call HubManager.Broadcast
// with a plain Event after their database work has committed. This keeps the
// transport concern (connection pools, pumps, locking) out of the domain layer
// as required by CLAUDE.md.
//
// Wire format of every broadcast event:
//
//	{
//	  "type": "task_published",
//	  "roomCode": "ABC123",
//	  "data": { ... },
//	  "occurredAt": "2026-06-16T10:00:00Z"
//	}
package websocket

import "time"

// EventType enumerates the real-time events broadcast to a room. The string
// values are part of the client contract and must not change casually.
type EventType string

const (
	EventStudentJoined        EventType = "student_joined"
	EventTaskPublished        EventType = "task_published"
	EventTaskPaused           EventType = "task_paused"
	EventTaskClosed           EventType = "task_closed"
	EventSubmissionCreated    EventType = "submission_created"
	EventScoreUpdated         EventType = "score_updated"
	EventRankingUpdated       EventType = "ranking_updated"
	EventFeaturedAnswerUpdate EventType = "featured_answer_updated"
	EventRoomEnded            EventType = "room_ended"
)

// Event is the unified envelope broadcast to every client in a room. Data is
// any JSON-serialisable payload (typically a struct or map) and may be nil.
type Event struct {
	Type       EventType   `json:"type"`
	RoomCode   string      `json:"roomCode"`
	Data       interface{} `json:"data"`
	OccurredAt time.Time   `json:"occurredAt"`
}

// NewEvent builds an Event stamped with the current UTC time. Services should
// use this so the occurredAt field is consistent across all events.
func NewEvent(typ EventType, roomCode string, data interface{}) Event {
	return Event{
		Type:       typ,
		RoomCode:   roomCode,
		Data:       data,
		OccurredAt: time.Now().UTC(),
	}
}
