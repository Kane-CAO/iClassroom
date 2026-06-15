package service

import (
	"context"
	"errors"

	"iclassroom/backend/internal/apperr"
	"iclassroom/backend/internal/domain"
	"iclassroom/backend/internal/repository"
)

// FeaturedAnswerService implements teacher-facing featured-answer rules.
type FeaturedAnswerService struct {
	rooms    RoomRepository
	featured FeaturedAnswerRepository
}

func NewFeaturedAnswerService(rooms RoomRepository, featured FeaturedAnswerRepository) *FeaturedAnswerService {
	return &FeaturedAnswerService{rooms: rooms, featured: featured}
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

	return s.featured.Upsert(ctx, room.ID, submissionID, mode)
}

