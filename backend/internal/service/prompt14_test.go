package service

import (
	"context"
	"testing"
	"time"

	"iclassroom/backend/internal/domain"
	"iclassroom/backend/internal/repository"
)

type prompt14Store struct {
	roomsByCode  map[string]*domain.Room
	roomsByToken map[string]*domain.Room

	submissionRooms map[int64]*domain.Room
	featured        map[int64]*domain.FeaturedAnswer

	leaderboard []repository.LeaderboardItem
	currentTask *repository.DisplayTask
	featuredOut []repository.FeaturedAnswerView
	groups      []repository.GroupWithCount

	studentCount int
	groupScores  []repository.GroupScore
	tasks        []repository.TaskCompletion
	timeline     []repository.SubmissionTimelinePoint
}

func newPrompt14Store() *prompt14Store {
	room := &domain.Room{ID: 1, RoomCode: "ABC123", Title: "Demo Class", TeacherToken: "teacher_ok"}
	other := &domain.Room{ID: 2, RoomCode: "XYZ999", Title: "Other Class", TeacherToken: "teacher_other"}
	return &prompt14Store{
		roomsByCode: map[string]*domain.Room{
			room.RoomCode:  room,
			other.RoomCode: other,
		},
		roomsByToken: map[string]*domain.Room{
			room.TeacherToken:  room,
			other.TeacherToken: other,
		},
		submissionRooms: map[int64]*domain.Room{10: room},
		featured:        map[int64]*domain.FeaturedAnswer{},
	}
}

func (s *prompt14Store) CreateRoomWithGroups(context.Context, *domain.Room) ([]domain.Group, error) {
	return nil, nil
}

func (s *prompt14Store) GetByRoomCode(_ context.Context, code string) (*domain.Room, error) {
	if room, ok := s.roomsByCode[code]; ok {
		return room, nil
	}
	return nil, repository.ErrNotFound
}

func (s *prompt14Store) GetByTeacherToken(_ context.Context, token string) (*domain.Room, error) {
	if room, ok := s.roomsByToken[token]; ok {
		return room, nil
	}
	return nil, repository.ErrNotFound
}

func (s *prompt14Store) EndRoom(_ context.Context, roomID int64, endedAt time.Time) error {
	for _, room := range s.roomsByCode {
		if room.ID == roomID {
			room.Status = domain.RoomStatusEnded
			t := endedAt.UTC()
			room.EndedAt = &t
			return nil
		}
	}
	return repository.ErrNotFound
}

func (s *prompt14Store) GetRoomBySubmissionID(_ context.Context, submissionID int64) (*domain.Room, error) {
	if room, ok := s.submissionRooms[submissionID]; ok {
		return room, nil
	}
	return nil, repository.ErrNotFound
}

func (s *prompt14Store) ListByRoomID(context.Context, int64) ([]repository.GroupWithCount, error) {
	return s.groups, nil
}

func (s *prompt14Store) GetByID(_ context.Context, groupID int64) (*domain.Group, error) {
	for _, group := range s.groups {
		if group.ID == groupID {
			g := group.Group
			return &g, nil
		}
	}
	return nil, repository.ErrNotFound
}

func (s *prompt14Store) Upsert(_ context.Context, roomID, submissionID int64, mode domain.DisplayMode) (*domain.FeaturedAnswer, error) {
	fa := &domain.FeaturedAnswer{ID: 99, RoomID: roomID, SubmissionID: submissionID, DisplayMode: mode}
	s.featured[submissionID] = fa
	return fa, nil
}

func (s *prompt14Store) ListTasksForStudent(context.Context, int64, int64, int64) ([]repository.StudentTaskWithSubmission, error) {
	return nil, nil
}

func (s *prompt14Store) GetTargetedTaskForStudent(context.Context, int64, int64, int64) (*domain.Task, error) {
	return nil, nil
}

func (s *prompt14Store) CreateText(context.Context, int64, *domain.Student, string) (*domain.Submission, error) {
	return nil, nil
}

func (s *prompt14Store) CreateImages(context.Context, int64, []domain.SubmissionImage) ([]domain.SubmissionImage, error) {
	return nil, nil
}

func (s *prompt14Store) CreateFiles(context.Context, int64, []domain.SubmissionFile) ([]domain.SubmissionFile, error) {
	return nil, nil
}

func (s *prompt14Store) DeleteByID(context.Context, int64) error {
	return nil
}

func (s *prompt14Store) ListByTaskID(context.Context, int64) ([]repository.SubmissionWithStudent, error) {
	return nil, nil
}

func (s *prompt14Store) GradeSubmission(context.Context, int64, int, string) (*domain.Submission, error) {
	return nil, nil
}

func (s *prompt14Store) ListLeaderboardByRoomID(context.Context, int64) ([]repository.LeaderboardItem, error) {
	return s.leaderboard, nil
}

func (s *prompt14Store) GetCurrentTask(context.Context, int64) (*repository.DisplayTask, error) {
	if s.currentTask == nil {
		return nil, repository.ErrNotFound
	}
	return s.currentTask, nil
}

func (s *prompt14Store) ListFeaturedAnswers(context.Context, int64) ([]repository.FeaturedAnswerView, error) {
	return s.featuredOut, nil
}

func (s *prompt14Store) CountStudents(context.Context, int64) (int, error) {
	return s.studentCount, nil
}

func (s *prompt14Store) ListGroupScores(context.Context, int64) ([]repository.GroupScore, error) {
	return s.groupScores, nil
}

func (s *prompt14Store) ListTaskCompletion(context.Context, int64) ([]repository.TaskCompletion, error) {
	return s.tasks, nil
}

func (s *prompt14Store) ListSubmissionTimeline(context.Context, int64) ([]repository.SubmissionTimelinePoint, error) {
	return s.timeline, nil
}

func TestFeaturedAnswerService(t *testing.T) {
	ctx := context.Background()

	t.Run("invalid display mode", func(t *testing.T) {
		store := newPrompt14Store()
		_, err := NewFeaturedAnswerService(store, store, nil).Feature(ctx, 10, "teacher_ok", domain.DisplayMode("bad"))
		wantCode(t, err, "INVALID_DISPLAY_MODE")
	})

	t.Run("missing token", func(t *testing.T) {
		store := newPrompt14Store()
		_, err := NewFeaturedAnswerService(store, store, nil).Feature(ctx, 10, "", domain.DisplayAnonymous)
		wantCode(t, err, "INVALID_TEACHER_TOKEN")
	})

	t.Run("wrong room token", func(t *testing.T) {
		store := newPrompt14Store()
		_, err := NewFeaturedAnswerService(store, store, nil).Feature(ctx, 10, "teacher_other", domain.DisplayAnonymous)
		wantCode(t, err, "ROOM_ACCESS_DENIED")
	})

	t.Run("upsert featured answer", func(t *testing.T) {
		store := newPrompt14Store()
		fa, err := NewFeaturedAnswerService(store, store, nil).Feature(ctx, 10, "teacher_ok", domain.DisplayShowGroup)
		if err != nil {
			t.Fatalf("Feature error: %v", err)
		}
		if fa.SubmissionID != 10 || fa.DisplayMode != domain.DisplayShowGroup {
			t.Fatalf("unexpected featured answer: %+v", fa)
		}
	})
}

func TestDisplayServiceEmptyData(t *testing.T) {
	store := newPrompt14Store()
	store.groups = []repository.GroupWithCount{
		{Group: domain.Group{ID: 1, GroupName: "Group A", Capacity: 4}, CurrentCount: 0},
	}
	view, err := NewDisplayService(store, store, store, store).Get(context.Background(), "ABC123")
	if err != nil {
		t.Fatalf("Display Get error: %v", err)
	}
	if view.CurrentTask != nil {
		t.Fatalf("CurrentTask = %+v, want nil", view.CurrentTask)
	}
	if len(view.Ranking) != 0 || len(view.FeaturedAnswers) != 0 {
		t.Fatalf("expected empty ranking and featured answers: %+v", view)
	}
	if len(view.Groups) != 1 {
		t.Fatalf("expected display groups: %+v", view.Groups)
	}
}

func TestDisplayServiceRoomAuth(t *testing.T) {
	store := newPrompt14Store()
	_, err := NewDisplayService(store, store, store, store).Get(context.Background(), "NOPE12")
	wantCode(t, err, "ROOM_NOT_FOUND")
}

func TestDisplayServiceDoesNotRequireTeacherToken(t *testing.T) {
	store := newPrompt14Store()
	_, err := NewDisplayService(store, store, store, store).Get(context.Background(), "ABC123")
	if err != nil {
		t.Fatalf("Display Get without teacher token error: %v", err)
	}
}

func TestAnalyticsServiceRates(t *testing.T) {
	ctx := context.Background()

	t.Run("empty data returns zero rate", func(t *testing.T) {
		store := newPrompt14Store()
		view, err := NewAnalyticsService(store, store).Get(ctx, "ABC123", "teacher_ok")
		if err != nil {
			t.Fatalf("Analytics Get error: %v", err)
		}
		if view.SubmissionRate != 0 || view.OnlineCount != 0 {
			t.Fatalf("rates = submission %v online %d, want 0/0", view.SubmissionRate, view.OnlineCount)
		}
		if len(view.GroupScores) != 0 || len(view.TaskCompletion) != 0 || len(view.SubmissionTimeline) != 0 {
			t.Fatalf("expected empty arrays: %+v", view)
		}
	})

	t.Run("task completion rate uses zero-safe division", func(t *testing.T) {
		store := newPrompt14Store()
		store.tasks = []repository.TaskCompletion{
			{TaskID: 1, TaskTitle: "A", SubmittedCount: 2, TargetStudentCount: 4},
			{TaskID: 2, TaskTitle: "B", SubmittedCount: 0, TargetStudentCount: 0},
		}
		store.timeline = []repository.SubmissionTimelinePoint{{Time: time.Now().UTC(), Count: 1}}

		view, err := NewAnalyticsService(store, store).Get(ctx, "ABC123", "teacher_ok")
		if err != nil {
			t.Fatalf("Analytics Get error: %v", err)
		}
		if view.SubmissionRate != 0.5 {
			t.Fatalf("SubmissionRate = %v, want 0.5", view.SubmissionRate)
		}
		if view.TaskCompletion[1].CompletionRate != 0 {
			t.Fatalf("zero denominator completion = %v, want 0", view.TaskCompletion[1].CompletionRate)
		}
	})
}
