package service

import (
	"context"
	"errors"
	"strings"

	"iclassroom/backend/internal/apperr"
	"iclassroom/backend/internal/domain"
	"iclassroom/backend/internal/repository"
)

const maxNicknameLen = 64

// StudentService implements the student entry rules: viewing a room, joining,
// and resuming a session by token.
type StudentService struct {
	rooms    RoomRepository
	groups   GroupRepository
	students StudentRepository
}

// NewStudentService constructs a StudentService.
func NewStudentService(rooms RoomRepository, groups GroupRepository, students StudentRepository) *StudentService {
	return &StudentService{rooms: rooms, groups: groups, students: students}
}

// StudentRoomView is the public room view shown before a student joins.
type StudentRoomView struct {
	Room   *domain.Room
	Groups []repository.GroupWithCount
}

// GetRoomForStudent returns the room and its joinable groups. The room must
// exist and must not be ended.
func (s *StudentService) GetRoomForStudent(ctx context.Context, roomCode string) (*StudentRoomView, error) {
	room, err := s.loadActiveRoom(ctx, roomCode)
	if err != nil {
		return nil, err
	}
	groups, err := s.groups.ListByRoomID(ctx, room.ID)
	if err != nil {
		return nil, err
	}
	return &StudentRoomView{Room: room, Groups: groups}, nil
}

// JoinResult is the data returned to a student who just joined.
type JoinResult struct {
	Student   *domain.Student
	RoomCode  string
	GroupName string
}

// Join admits a student to a group. It enforces: room exists and is active,
// nickname is valid, group belongs to the room, group is not full, and the
// nickname is unique within the room (the latter three are enforced
// transactionally and race-safe in the repository). A fresh, unpredictable
// clientToken is minted on success.
func (s *StudentService) Join(ctx context.Context, roomCode, nickname string, groupID int64) (*JoinResult, error) {
	room, err := s.loadActiveRoom(ctx, roomCode)
	if err != nil {
		return nil, err
	}

	nickname = strings.TrimSpace(nickname)
	if nickname == "" || len([]rune(nickname)) > maxNicknameLen {
		return nil, apperr.InvalidNickname()
	}
	if groupID <= 0 {
		return nil, apperr.GroupNotFound()
	}

	clientToken, err := newToken("student")
	if err != nil {
		return nil, err
	}

	student, err := s.students.Join(ctx, room.ID, groupID, nickname, clientToken)
	switch {
	case errors.Is(err, repository.ErrNotFound):
		return nil, apperr.GroupNotFound()
	case errors.Is(err, repository.ErrGroupFull):
		return nil, apperr.GroupFull()
	case errors.Is(err, repository.ErrDuplicate):
		return nil, apperr.NicknameDuplicated()
	case err != nil:
		return nil, err
	}

	group, err := s.groups.GetByID(ctx, groupID)
	if err != nil {
		return nil, err
	}
	return &JoinResult{Student: student, RoomCode: room.RoomCode, GroupName: group.GroupName}, nil
}

// ResumeResult is the data returned when a student session is restored.
type ResumeResult struct {
	Student    *domain.Student
	RoomCode   string
	GroupName  string
	RoomStatus domain.RoomStatus
}

// Resume restores a student identity from roomCode + clientToken. The token
// must be valid and belong to the addressed room; the room must not be ended.
func (s *StudentService) Resume(ctx context.Context, roomCode, clientToken string) (*ResumeResult, error) {
	room, err := s.rooms.GetByRoomCode(ctx, roomCode)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, apperr.RoomNotFound()
	}
	if err != nil {
		return nil, err
	}

	if clientToken == "" {
		return nil, apperr.InvalidStudentToken()
	}
	student, err := s.students.GetByClientToken(ctx, clientToken)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, apperr.InvalidStudentToken()
	}
	if err != nil {
		return nil, err
	}
	// A valid token that belongs to a different room must not unlock this one.
	if student.RoomID != room.ID {
		return nil, apperr.InvalidStudentToken()
	}
	if room.Status == domain.RoomStatusEnded {
		return nil, apperr.RoomEnded()
	}

	group, err := s.groups.GetByID(ctx, student.GroupID)
	if err != nil {
		return nil, err
	}
	return &ResumeResult{
		Student:    student,
		RoomCode:   room.RoomCode,
		GroupName:  group.GroupName,
		RoomStatus: room.Status,
	}, nil
}

// loadActiveRoom fetches a room and rejects it if ended.
func (s *StudentService) loadActiveRoom(ctx context.Context, roomCode string) (*domain.Room, error) {
	room, err := s.rooms.GetByRoomCode(ctx, roomCode)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, apperr.RoomNotFound()
	}
	if err != nil {
		return nil, err
	}
	if room.Status == domain.RoomStatusEnded {
		return nil, apperr.RoomEnded()
	}
	return room, nil
}
