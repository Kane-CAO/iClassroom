package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"iclassroom/backend/internal/apperr"
	"iclassroom/backend/internal/domain"
	"iclassroom/backend/internal/repository"
	"iclassroom/backend/internal/websocket"
)

// Room creation bounds. groupCount/groupCapacity default when the request
// omits them (0); values outside [min,max] are rejected.
const (
	defaultGroupCount    = 6
	defaultGroupCapacity = 10
	minGroupCount        = 1
	maxGroupCount        = 50
	minGroupCapacity     = 1
	maxGroupCapacity     = 200

	maxTitleLen = 255

	// roomCodeMaxRetries bounds collision retries when minting a unique
	// roomCode / teacherToken.
	roomCodeMaxRetries = 5
)

// RoomService implements the teacher-facing room rules.
type RoomService struct {
	rooms       RoomRepository
	groups      GroupRepository
	frontendURL string // base URL for building joinUrl / dashboard URL
	broadcaster EventBroadcaster
}

// NewRoomService constructs a RoomService. frontendURL has no trailing slash. A
// nil broadcaster falls back to a noop.
func NewRoomService(rooms RoomRepository, groups GroupRepository, frontendURL string, broadcaster EventBroadcaster) *RoomService {
	return &RoomService{
		rooms:       rooms,
		groups:      groups,
		frontendURL: frontendURL,
		broadcaster: resolveBroadcaster(broadcaster),
	}
}

// CreateRoomInput is the validated input for creating a room.
type CreateRoomInput struct {
	TeacherID        int64
	Title            string
	GroupCount       int
	GroupCapacity    int
	AllowChooseGroup bool
}

// CreateRoomResult bundles the created room with everything the API response
// needs. Room.TeacherToken is the freshly minted management credential and is
// only ever exposed here.
type CreateRoomResult struct {
	Room                *domain.Room
	Groups              []domain.Group
	JoinURL             string
	TeacherDashboardURL string
}

// CreateRoom validates input, mints a unique roomCode + teacherToken, and
// creates the room together with its groups in one transaction. On a roomCode/
// teacherToken collision it regenerates and retries.
func (s *RoomService) CreateRoom(ctx context.Context, in CreateRoomInput) (*CreateRoomResult, error) {
	title := strings.TrimSpace(in.Title)
	if title == "" || len([]rune(title)) > maxTitleLen {
		return nil, apperr.InvalidRequest("title is required and must be at most 255 characters")
	}

	groupCount := in.GroupCount
	if groupCount == 0 {
		groupCount = defaultGroupCount
	}
	if groupCount < minGroupCount || groupCount > maxGroupCount {
		return nil, apperr.InvalidGroupCount()
	}

	groupCapacity := in.GroupCapacity
	if groupCapacity == 0 {
		groupCapacity = defaultGroupCapacity
	}
	if groupCapacity < minGroupCapacity || groupCapacity > maxGroupCapacity {
		return nil, apperr.InvalidGroupCapacity()
	}

	for attempt := 0; attempt < roomCodeMaxRetries; attempt++ {
		code, err := newRoomCode()
		if err != nil {
			return nil, err
		}
		teacherToken, err := newToken("teacher")
		if err != nil {
			return nil, err
		}

		room := &domain.Room{
			RoomCode: code,
			Title:    title,
			// MVP treats a freshly created room as immediately active.
			Status:           domain.RoomStatusActive,
			GroupCount:       groupCount,
			GroupCapacity:    groupCapacity,
			AllowChooseGroup: in.AllowChooseGroup,
			TeacherToken:     teacherToken,
		}
		if in.TeacherID > 0 {
			room.TeacherID = &in.TeacherID
		}

		groups, err := s.rooms.CreateRoomWithGroups(ctx, room)
		if errors.Is(err, repository.ErrDuplicate) {
			continue // regenerate code/token and retry
		}
		if err != nil {
			return nil, apperr.RoomCreateFailed()
		}

		return &CreateRoomResult{
			Room:                room,
			Groups:              groups,
			JoinURL:             s.joinURL(code),
			TeacherDashboardURL: s.frontendURL + "/teacher/rooms/" + code + "/dashboard",
		}, nil
	}
	return nil, apperr.RoomCreateFailed()
}

// GetRoom returns the room after verifying the teacher token belongs to it.
func (s *RoomService) GetRoom(ctx context.Context, roomCode, teacherToken string) (*domain.Room, error) {
	return s.verifyTeacher(ctx, roomCode, teacherToken)
}

type roomListRepository interface {
	ListByTeacherID(ctx context.Context, teacherID int64) ([]domain.Room, error)
}

func (s *RoomService) ListRoomsForTeacher(ctx context.Context, teacherID int64) ([]domain.Room, error) {
	if teacherID <= 0 {
		return nil, apperr.Unauthorized()
	}
	lister, ok := s.rooms.(roomListRepository)
	if !ok {
		return []domain.Room{}, nil
	}
	return lister.ListByTeacherID(ctx, teacherID)
}

func (s *RoomService) GetRoomForTeacher(ctx context.Context, roomCode string, teacherID int64, legacyToken string) (*domain.Room, error) {
	return s.verifyTeacher(ctx, roomCode, legacyToken, teacherID)
}

// RoomOverview is the teacher dashboard read model.
type RoomOverview struct {
	Room         *domain.Room
	Groups       []repository.GroupWithCount
	StudentCount int
	JoinURL      string
}

// GetOverview returns room totals (student count, per-group counts and scores)
// after verifying the teacher token. Tasks are added in a later step.
func (s *RoomService) GetOverview(ctx context.Context, roomCode, teacherToken string) (*RoomOverview, error) {
	room, err := s.verifyTeacher(ctx, roomCode, teacherToken)
	if err != nil {
		return nil, err
	}
	groups, err := s.groups.ListByRoomID(ctx, room.ID)
	if err != nil {
		return nil, err
	}
	studentCount := 0
	for _, g := range groups {
		studentCount += g.CurrentCount
	}
	return &RoomOverview{
		Room:         room,
		Groups:       groups,
		StudentCount: studentCount,
		JoinURL:      s.joinURL(room.RoomCode),
	}, nil
}

func (s *RoomService) GetOverviewForTeacher(ctx context.Context, roomCode string, teacherID int64, legacyToken string) (*RoomOverview, error) {
	room, err := s.verifyTeacher(ctx, roomCode, legacyToken, teacherID)
	if err != nil {
		return nil, err
	}
	groups, err := s.groups.ListByRoomID(ctx, room.ID)
	if err != nil {
		return nil, err
	}
	studentCount := 0
	for _, g := range groups {
		studentCount += g.CurrentCount
	}
	return &RoomOverview{
		Room:         room,
		Groups:       groups,
		StudentCount: studentCount,
		JoinURL:      s.joinURL(room.RoomCode),
	}, nil
}

// EndRoom transitions a room to ended and stamps endedAt.
func (s *RoomService) EndRoom(ctx context.Context, roomCode, teacherToken string) (*domain.Room, error) {
	room, err := s.verifyTeacher(ctx, roomCode, teacherToken)
	if err != nil {
		return nil, err
	}
	if room.Status == domain.RoomStatusEnded {
		return nil, apperr.RoomAlreadyEnded()
	}

	endedAt := time.Now().UTC()
	if err := s.rooms.EndRoom(ctx, room.ID, endedAt); errors.Is(err, repository.ErrNotFound) {
		return nil, apperr.RoomNotFound()
	} else if err != nil {
		return nil, err
	}

	room.Status = domain.RoomStatusEnded
	room.EndedAt = &endedAt

	// Room is now ended in the DB: notify everyone so students are locked out.
	emit(s.broadcaster, room.RoomCode, websocket.EventRoomEnded, map[string]any{
		"status": room.Status,
	})

	return room, nil
}

func (s *RoomService) EndRoomForTeacher(ctx context.Context, roomCode string, teacherID int64, legacyToken string) (*domain.Room, error) {
	room, err := s.verifyTeacher(ctx, roomCode, legacyToken, teacherID)
	if err != nil {
		return nil, err
	}
	if room.Status == domain.RoomStatusEnded {
		return nil, apperr.RoomAlreadyEnded()
	}

	endedAt := time.Now().UTC()
	if err := s.rooms.EndRoom(ctx, room.ID, endedAt); errors.Is(err, repository.ErrNotFound) {
		return nil, apperr.RoomNotFound()
	} else if err != nil {
		return nil, err
	}

	room.Status = domain.RoomStatusEnded
	room.EndedAt = &endedAt

	emit(s.broadcaster, room.RoomCode, websocket.EventRoomEnded, map[string]any{
		"status": room.Status,
	})

	return room, nil
}

func (s *RoomService) joinURL(code string) string {
	return s.frontendURL + "/student?room=" + code
}

// verifyTeacher loads the room and authorizes the teacher token against it.
//   - room missing            -> ROOM_NOT_FOUND
//   - token empty             -> INVALID_TEACHER_TOKEN
//   - token matches this room -> ok
//   - token belongs to another room -> ROOM_ACCESS_DENIED
//   - token unknown           -> INVALID_TEACHER_TOKEN
func (s *RoomService) verifyTeacher(ctx context.Context, roomCode, teacherToken string, teacherIDs ...int64) (*domain.Room, error) {
	room, err := s.rooms.GetByRoomCode(ctx, roomCode)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, apperr.RoomNotFound()
	}
	if err != nil {
		return nil, err
	}
	if len(teacherIDs) > 0 && teacherIDs[0] > 0 {
		if room.TeacherID != nil && *room.TeacherID == teacherIDs[0] {
			return room, nil
		}
		if teacherToken == "" {
			return nil, apperr.RoomAccessDenied()
		}
	}
	if teacherToken == "" {
		return nil, apperr.InvalidTeacherToken()
	}
	if teacherToken == room.TeacherToken {
		return room, nil
	}
	// Token does not match this room. Distinguish "valid token, wrong room"
	// from "unknown token" for a precise error code.
	if _, err := s.rooms.GetByTeacherToken(ctx, teacherToken); err == nil {
		return nil, apperr.RoomAccessDenied()
	} else if !errors.Is(err, repository.ErrNotFound) {
		return nil, err
	}
	return nil, apperr.InvalidTeacherToken()
}
