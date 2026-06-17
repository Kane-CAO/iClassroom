package service

import (
	"context"
	"errors"
	"time"

	"iclassroom/backend/internal/apperr"
	"iclassroom/backend/internal/domain"
	"iclassroom/backend/internal/repository"
)

type DisplayService struct {
	rooms       RoomRepository
	groups      GroupRepository
	submissions SubmissionRepository
	display     DisplayRepository
}

func NewDisplayService(rooms RoomRepository, groups GroupRepository, submissions SubmissionRepository, display DisplayRepository) *DisplayService {
	return &DisplayService{rooms: rooms, groups: groups, submissions: submissions, display: display}
}

type DisplayTaskView struct {
	TaskID             int64
	Title              string
	DeadlineAt         time.Time
	SubmittedCount     int
	TargetStudentCount int
	CompletionRate     float64
}

type DisplayFeaturedAnswerView struct {
	FeaturedID   int64
	SubmissionID int64
	TaskID       int64
	DisplayMode  domain.DisplayMode
	ContentText  string
	Score        *int
	GroupID      int64
	GroupName    string
	SubmittedAt  time.Time
}

type DisplayView struct {
	Room            *domain.Room
	Groups          []repository.GroupWithCount
	Ranking         []LeaderboardEntry
	CurrentTask     *DisplayTaskView
	FeaturedAnswers []DisplayFeaturedAnswerView
}

func (s *DisplayService) Get(ctx context.Context, roomCode string) (*DisplayView, error) {
	room, err := s.rooms.GetByRoomCode(ctx, roomCode)
	switch {
	case errors.Is(err, repository.ErrNotFound):
		return nil, apperr.RoomNotFound()
	case err != nil:
		return nil, err
	}

	groups, err := s.groups.ListByRoomID(ctx, room.ID)
	if err != nil {
		return nil, err
	}

	leaderboardItems, err := s.submissions.ListLeaderboardByRoomID(ctx, room.ID)
	if err != nil {
		return nil, err
	}

	var currentTask *DisplayTaskView
	task, err := s.display.GetCurrentTask(ctx, room.ID)
	if errors.Is(err, repository.ErrNotFound) {
		currentTask = nil
	} else if err != nil {
		return nil, err
	} else {
		currentTask = &DisplayTaskView{
			TaskID:             task.TaskID,
			Title:              task.Title,
			DeadlineAt:         task.DeadlineAt,
			SubmittedCount:     task.SubmittedCount,
			TargetStudentCount: task.TargetStudentCount,
			CompletionRate:     rate(task.SubmittedCount, task.TargetStudentCount),
		}
	}

	featuredItems, err := s.display.ListFeaturedAnswers(ctx, room.ID)
	if err != nil {
		return nil, err
	}

	featured := make([]DisplayFeaturedAnswerView, 0, len(featuredItems))
	for _, item := range featuredItems {
		view := DisplayFeaturedAnswerView{
			FeaturedID:   item.FeaturedAnswer.ID,
			SubmissionID: item.FeaturedAnswer.SubmissionID,
			TaskID:       item.TaskID,
			DisplayMode:  item.FeaturedAnswer.DisplayMode,
			ContentText:  item.ContentText,
			Score:        item.Score,
			SubmittedAt:  item.SubmittedAt,
		}
		if item.FeaturedAnswer.DisplayMode == domain.DisplayShowGroup {
			view.GroupID = item.GroupID
			view.GroupName = item.GroupName
		}
		featured = append(featured, view)
	}

	return &DisplayView{
		Room:            room,
		Groups:          groups,
		Ranking:         buildLeaderboard(room.RoomCode, leaderboardItems, 0).Entries,
		CurrentTask:     currentTask,
		FeaturedAnswers: featured,
	}, nil
}

func rate(numerator, denominator int) float64 {
	if denominator <= 0 {
		return 0
	}
	return float64(numerator) / float64(denominator)
}
