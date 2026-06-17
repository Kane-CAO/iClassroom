package service

import (
	"context"
	"errors"

	"iclassroom/backend/internal/apperr"
	"iclassroom/backend/internal/domain"
	"iclassroom/backend/internal/repository"
	"iclassroom/backend/internal/websocket"
)

// FeaturedAnswerService implements teacher-facing featured-answer rules.
type FeaturedAnswerService struct {
	rooms       RoomRepository
	featured    FeaturedAnswerRepository
	broadcaster EventBroadcaster
}

func NewFeaturedAnswerService(rooms RoomRepository, featured FeaturedAnswerRepository, broadcaster EventBroadcaster) *FeaturedAnswerService {
	return &FeaturedAnswerService{
		rooms:       rooms,
		featured:    featured,
		broadcaster: resolveBroadcaster(broadcaster),
	}
}

func (s *FeaturedAnswerService) Feature(ctx context.Context, submissionID int64, teacherToken string, mode domain.DisplayMode) (*domain.FeaturedAnswer, error) {
	if submissionID <= 0 {
		return nil, apperr.SubmissionNotFound()
	}
	if !mode.Valid() {
		return nil, apperr.InvalidDisplayMode()
	}

	room, err := s.featured.GetRoomBySubmissionID(ctx, submissionID)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, apperr.SubmissionNotFound()
	}
	if err != nil {
		return nil, err
	}

	if err := verifyTeacherForRoom(ctx, s.rooms, room, teacherToken); err != nil {
		return nil, err
	}

	featured, err := s.featured.Upsert(ctx, room.ID, submissionID, mode)
	if err != nil {
		return nil, err
	}

	emit(s.broadcaster, room.RoomCode, websocket.EventFeaturedAnswerUpdate, map[string]any{
		"featuredId":   featured.ID,
		"submissionId": featured.SubmissionID,
		"displayMode":  featured.DisplayMode,
	})

	return featured, nil
}
