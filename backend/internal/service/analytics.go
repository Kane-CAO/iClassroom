package service

import (
	"context"
	"time"

	"iclassroom/backend/internal/domain"
)

type AnalyticsService struct {
	rooms     RoomRepository
	analytics AnalyticsRepository
}

func NewAnalyticsService(rooms RoomRepository, analytics AnalyticsRepository) *AnalyticsService {
	return &AnalyticsService{rooms: rooms, analytics: analytics}
}

type AnalyticsGroupScore struct {
	GroupID    int64
	GroupName  string
	ScoreTotal int
}

type AnalyticsTaskCompletion struct {
	TaskID             int64
	TaskTitle          string
	SubmittedCount     int
	TargetStudentCount int
	CompletionRate     float64
}

type AnalyticsSubmissionTimelinePoint struct {
	Time  time.Time
	Count int
}

type AnalyticsView struct {
	Room               *domain.Room
	StudentCount       int
	OnlineCount        int
	SubmissionRate     float64
	GroupScores        []AnalyticsGroupScore
	TaskCompletion     []AnalyticsTaskCompletion
	SubmissionTimeline []AnalyticsSubmissionTimelinePoint
}

func (s *AnalyticsService) Get(ctx context.Context, roomCode, teacherToken string) (*AnalyticsView, error) {
	room, err := verifyTeacherByRoomCode(ctx, s.rooms, roomCode, teacherToken)
	if err != nil {
		return nil, err
	}

	studentCount, err := s.analytics.CountStudents(ctx, room.ID)
	if err != nil {
		return nil, err
	}

	groupRows, err := s.analytics.ListGroupScores(ctx, room.ID)
	if err != nil {
		return nil, err
	}
	groupScores := make([]AnalyticsGroupScore, 0, len(groupRows))
	for _, row := range groupRows {
		groupScores = append(groupScores, AnalyticsGroupScore{
			GroupID:    row.GroupID,
			GroupName:  row.GroupName,
			ScoreTotal: row.ScoreTotal,
		})
	}

	taskRows, err := s.analytics.ListTaskCompletion(ctx, room.ID)
	if err != nil {
		return nil, err
	}
	taskCompletion := make([]AnalyticsTaskCompletion, 0, len(taskRows))
	totalSubmitted := 0
	totalTarget := 0
	for _, row := range taskRows {
		totalSubmitted += row.SubmittedCount
		totalTarget += row.TargetStudentCount
		taskCompletion = append(taskCompletion, AnalyticsTaskCompletion{
			TaskID:             row.TaskID,
			TaskTitle:          row.TaskTitle,
			SubmittedCount:     row.SubmittedCount,
			TargetStudentCount: row.TargetStudentCount,
			CompletionRate:     rate(row.SubmittedCount, row.TargetStudentCount),
		})
	}

	timelineRows, err := s.analytics.ListSubmissionTimeline(ctx, room.ID)
	if err != nil {
		return nil, err
	}
	timeline := make([]AnalyticsSubmissionTimelinePoint, 0, len(timelineRows))
	for _, row := range timelineRows {
		timeline = append(timeline, AnalyticsSubmissionTimelinePoint{
			Time:  row.Time,
			Count: row.Count,
		})
	}

	return &AnalyticsView{
		Room:               room,
		StudentCount:       studentCount,
		OnlineCount:        0,
		SubmissionRate:     rate(totalSubmitted, totalTarget),
		GroupScores:        groupScores,
		TaskCompletion:     taskCompletion,
		SubmissionTimeline: timeline,
	}, nil
}
