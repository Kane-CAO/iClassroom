package service

import (
	"context"
	"errors"
	"fmt"

	"iclassroom/backend/internal/apperr"
	"iclassroom/backend/internal/repository"
)

// WebSocket connection roles. Kept as plain strings here so the service layer
// stays independent of the websocket transport package (per CLAUDE.md, the
// WebSocket connection logic must not live in services).
const (
	wsRoleTeacher = "teacher"
	wsRoleStudent = "student"
	wsRoleDisplay = "display"
)

// WSAuthService validates an incoming WebSocket connection request before the
// HTTP handler upgrades it. It enforces the business rules (room must exist,
// students must present a valid token bound to that room) and leaves the
// transport concerns to the handler and the websocket package.
type WSAuthService struct {
	rooms    RoomRepository
	students StudentRepository
}

// NewWSAuthService wires the auth service to its repositories.
func NewWSAuthService(rooms RoomRepository, students StudentRepository) *WSAuthService {
	return &WSAuthService{rooms: rooms, students: students}
}

// WSConnInfo is the validated identity of an authorised connection. ClientID is
// a stable, human-readable label used for logging; StudentID is 0 for
// teacher/display connections.
type WSConnInfo struct {
	Role      string
	RoomCode  string
	ClientID  string
	StudentID int64
}

// Authorize validates role, room, and (for students) the session token.
//
//   - role must be teacher, student, or display          -> INVALID_ROLE
//   - room must exist                                     -> ROOM_NOT_FOUND
//   - student: token required                             -> INVALID_STUDENT_TOKEN
//   - student: token must resolve to a student            -> INVALID_STUDENT_TOKEN
//   - student: that student must belong to this room      -> ROOM_ACCESS_DENIED
//
// teacher and display connections only require the room to exist.
func (s *WSAuthService) Authorize(ctx context.Context, roomCode, role, token string) (*WSConnInfo, error) {
	switch role {
	case wsRoleTeacher, wsRoleStudent, wsRoleDisplay:
	default:
		return nil, apperr.InvalidRole()
	}

	room, err := s.rooms.GetByRoomCode(ctx, roomCode)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, apperr.RoomNotFound()
	}
	if err != nil {
		return nil, err
	}

	info := &WSConnInfo{Role: role, RoomCode: room.RoomCode}

	if role != wsRoleStudent {
		info.ClientID = fmt.Sprintf("%s:%s", role, room.RoomCode)
		return info, nil
	}

	if token == "" {
		return nil, apperr.InvalidStudentToken()
	}
	student, err := s.students.GetByClientToken(ctx, token)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, apperr.InvalidStudentToken()
	}
	if err != nil {
		return nil, err
	}
	if student.RoomID != room.ID {
		return nil, apperr.RoomAccessDenied()
	}

	info.StudentID = student.ID
	info.ClientID = fmt.Sprintf("student:%d", student.ID)
	return info, nil
}
