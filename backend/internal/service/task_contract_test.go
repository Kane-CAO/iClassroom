package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"iclassroom/backend/internal/domain"
	"iclassroom/backend/internal/repository"
)

type contractRoomRepo struct {
	byCode  map[string]*domain.Room
	byToken map[string]*domain.Room
}

func (r *contractRoomRepo) CreateRoomWithGroups(ctx context.Context, room *domain.Room) ([]domain.Group, error) {
	return nil, nil
}

func (r *contractRoomRepo) GetByRoomCode(ctx context.Context, code string) (*domain.Room, error) {
	room, ok := r.byCode[code]
	if !ok {
		return nil, repository.ErrNotFound
	}
	return room, nil
}

func (r *contractRoomRepo) GetByTeacherToken(ctx context.Context, token string) (*domain.Room, error) {
	room, ok := r.byToken[token]
	if !ok {
		return nil, repository.ErrNotFound
	}
	return room, nil
}

type contractGroupRepo struct {
	byID map[int64]*domain.Group
}

func (r *contractGroupRepo) ListByRoomID(ctx context.Context, roomID int64) ([]repository.GroupWithCount, error) {
	return nil, nil
}

func (r *contractGroupRepo) GetByID(ctx context.Context, groupID int64) (*domain.Group, error) {
	group, ok := r.byID[groupID]
	if !ok {
		return nil, repository.ErrNotFound
	}
	return group, nil
}

type contractStudentRepo struct {
	byToken map[string]*domain.Student
}

func (r *contractStudentRepo) Join(ctx context.Context, roomID, groupID int64, nickname, clientToken string) (*domain.Student, error) {
	return nil, nil
}

func (r *contractStudentRepo) GetByClientToken(ctx context.Context, token string) (*domain.Student, error) {
	student, ok := r.byToken[token]
	if !ok {
		return nil, repository.ErrNotFound
	}
	return student, nil
}

type contractTaskRepo struct {
	roomsByTask map[int64]*domain.Room
	tasks       map[int64]*domain.Task
	created     []*domain.Task
}

func (r *contractTaskRepo) Create(ctx context.Context, task *domain.Task, targetGroupIDs []int64) error {
	task.ID = int64(len(r.created) + 1)
	r.created = append(r.created, task)
	r.tasks[task.ID] = task
	return nil
}

func (r *contractTaskRepo) ListByRoomID(ctx context.Context, roomID int64) ([]repository.TaskWithStats, error) {
	return nil, nil
}

func (r *contractTaskRepo) ListTargetGroupIDs(ctx context.Context, taskID int64) ([]int64, error) {
	return nil, nil
}

func (r *contractTaskRepo) GetRoomByTaskID(ctx context.Context, taskID int64) (*domain.Room, error) {
	room, ok := r.roomsByTask[taskID]
	if !ok {
		return nil, repository.ErrNotFound
	}
	return room, nil
}

func (r *contractTaskRepo) UpdateStatus(ctx context.Context, taskID int64, status domain.TaskStatus) error {
	task, ok := r.tasks[taskID]
	if !ok {
		return repository.ErrNotFound
	}
	task.Status = status
	return nil
}

type contractSubmissionRepo struct {
	studentTasks     []repository.StudentTaskWithSubmission
	targetedTasks    map[int64]*domain.Task
	created          map[int64]*domain.Submission
	duplicateCreate  bool
	roomBySubmission map[int64]*domain.Room
	graded           map[int64]*domain.Submission
	groups           map[int64]*domain.Group
	leaderboard      []repository.LeaderboardItem
}

func (r *contractSubmissionRepo) ListTasksForStudent(ctx context.Context, studentID, roomID, groupID int64) ([]repository.StudentTaskWithSubmission, error) {
	return r.studentTasks, nil
}

func (r *contractSubmissionRepo) GetTargetedTaskForStudent(ctx context.Context, taskID, roomID, groupID int64) (*domain.Task, error) {
	task, ok := r.targetedTasks[taskID]
	if !ok {
		return nil, repository.ErrNotFound
	}
	return task, nil
}

func (r *contractSubmissionRepo) CreateText(ctx context.Context, taskID int64, student *domain.Student, contentText string) (*domain.Submission, error) {
	if r.duplicateCreate {
		return nil, repository.ErrDuplicate
	}
	sub := &domain.Submission{
		ID:          int64(len(r.created) + 1),
		TaskID:      taskID,
		StudentID:   student.ID,
		RoomID:      student.RoomID,
		GroupID:     student.GroupID,
		ContentText: contentText,
		Status:      domain.SubmissionStatusSubmitted,
		SubmittedAt: time.Now().UTC(),
	}
	r.created[sub.ID] = sub
	return sub, nil
}

func (r *contractSubmissionRepo) ListByTaskID(ctx context.Context, taskID int64) ([]repository.SubmissionWithStudent, error) {
	return nil, nil
}

func (r *contractSubmissionRepo) GetRoomBySubmissionID(ctx context.Context, submissionID int64) (*domain.Room, error) {
	room, ok := r.roomBySubmission[submissionID]
	if !ok {
		return nil, repository.ErrNotFound
	}
	return room, nil
}

func (r *contractSubmissionRepo) GradeSubmission(ctx context.Context, submissionID int64, score int, comment string) (*domain.Submission, error) {
	sub, ok := r.graded[submissionID]
	if !ok {
		return nil, repository.ErrNotFound
	}

	previous := 0
	if sub.Score != nil {
		previous = *sub.Score
	}
	if group, ok := r.groups[sub.GroupID]; ok {
		group.ScoreTotal += score - previous
	}

	scoreCopy := score
	now := time.Now().UTC()
	sub.Score = &scoreCopy
	sub.Comment = comment
	sub.Status = domain.SubmissionStatusGraded
	sub.GradedAt = &now

	return sub, nil
}

func (r *contractSubmissionRepo) ListLeaderboardByRoomID(ctx context.Context, roomID int64) ([]repository.LeaderboardItem, error) {
	return r.leaderboard, nil
}

func newContractTaskService() (*TaskService, *contractRoomRepo, *contractGroupRepo, *contractStudentRepo, *contractTaskRepo, *contractSubmissionRepo) {
	room := &domain.Room{
		ID:           1,
		RoomCode:     "ABC123",
		Title:        "Demo",
		Status:       domain.RoomStatusActive,
		TeacherToken: "teacher_1",
	}
	group1 := &domain.Group{ID: 1, RoomID: 1, GroupName: "Group 1", Capacity: 10, ScoreTotal: 0}
	group2 := &domain.Group{ID: 2, RoomID: 1, GroupName: "Group 2", Capacity: 10, ScoreTotal: 0}
	student := &domain.Student{ID: 1, RoomID: 1, GroupID: 1, Nickname: "Tom", ClientToken: "student_1"}

	rooms := &contractRoomRepo{
		byCode:  map[string]*domain.Room{"ABC123": room},
		byToken: map[string]*domain.Room{"teacher_1": room},
	}
	groups := &contractGroupRepo{byID: map[int64]*domain.Group{1: group1, 2: group2}}
	students := &contractStudentRepo{byToken: map[string]*domain.Student{"student_1": student}}
	tasks := &contractTaskRepo{
		roomsByTask: map[int64]*domain.Room{1: room},
		tasks: map[int64]*domain.Task{
			1: {
				ID:         1,
				RoomID:     1,
				Title:      "Task 1",
				DeadlineAt: time.Now().Add(time.Hour),
				TargetType: domain.TargetAll,
				Status:     domain.TaskStatusPublished,
			},
		},
	}
	submissions := &contractSubmissionRepo{
		targetedTasks: map[int64]*domain.Task{1: tasks.tasks[1]},
		created:       map[int64]*domain.Submission{},
		roomBySubmission: map[int64]*domain.Room{
			1: room,
		},
		graded: map[int64]*domain.Submission{
			1: {
				ID:          1,
				TaskID:      1,
				StudentID:   1,
				RoomID:      1,
				GroupID:     1,
				ContentText: "answer",
				Status:      domain.SubmissionStatusSubmitted,
				SubmittedAt: time.Now().Add(-time.Minute),
			},
		},
		groups: map[int64]*domain.Group{1: group1, 2: group2},
		leaderboard: []repository.LeaderboardItem{
			{Group: *group1, CurrentCount: 1},
			{Group: *group2, CurrentCount: 0},
		},
	}

	return NewTaskService(rooms, groups, students, tasks, submissions), rooms, groups, students, tasks, submissions
}

func TestTaskContract_CreateTaskValidation(t *testing.T) {
	svc, _, _, _, tasks, _ := newContractTaskService()

	_, err := svc.Create(context.Background(), CreateTaskInput{
		RoomCode:     "ABC123",
		TeacherToken: "teacher_1",
		Title:        "New task",
		DeadlineAt:   time.Now().Add(time.Hour),
		TargetType:   domain.TargetAll,
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if len(tasks.created) != 1 {
		t.Fatalf("expected one created task, got %d", len(tasks.created))
	}

	_, err = svc.Create(context.Background(), CreateTaskInput{
		RoomCode:       "ABC123",
		TeacherToken:   "teacher_1",
		Title:          "Bad groups",
		DeadlineAt:     time.Now().Add(time.Hour),
		TargetType:     domain.TargetGroups,
		TargetGroupIDs: []int64{999},
	})
	if err == nil {
		t.Fatal("expected invalid target group error")
	}
}

func TestTaskContract_TextSubmissionDuplicatePauseAndClose(t *testing.T) {
	svc, _, _, _, tasks, submissions := newContractTaskService()

	if _, err := svc.SubmitText(context.Background(), 1, "student_1", "my answer"); err != nil {
		t.Fatalf("SubmitText returned error: %v", err)
	}
	if len(submissions.created) != 1 {
		t.Fatalf("expected one submission, got %d", len(submissions.created))
	}

	submissions.duplicateCreate = true
	if _, err := svc.SubmitText(context.Background(), 1, "student_1", "again"); err == nil {
		t.Fatal("expected duplicate submission error")
	}
	submissions.duplicateCreate = false

	tasks.tasks[1].Status = domain.TaskStatusPaused
	if _, err := svc.SubmitText(context.Background(), 1, "student_1", "paused"); err == nil {
		t.Fatal("expected paused task submission error")
	}

	tasks.tasks[1].Status = domain.TaskStatusClosed
	if _, err := svc.SubmitText(context.Background(), 1, "student_1", "closed"); err == nil {
		t.Fatal("expected closed task submission error")
	}
}

func TestTaskContract_GradeRangeAndRegradeDelta(t *testing.T) {
	svc, _, _, _, _, _ := newContractTaskService()

	if _, err := svc.GradeSubmission(context.Background(), 1, "teacher_1", 0, "bad"); err == nil {
		t.Fatal("expected invalid score error for zero")
	}
	if _, err := svc.GradeSubmission(context.Background(), 1, "teacher_1", 11, "bad"); err == nil {
		t.Fatal("expected invalid score error for score greater than 10")
	}

	first, err := svc.GradeSubmission(context.Background(), 1, "teacher_1", 8, "good")
	if err != nil {
		t.Fatalf("first GradeSubmission returned error: %v", err)
	}
	if first.GroupScoreTotal != 8 {
		t.Fatalf("expected groupScoreTotal 8, got %d", first.GroupScoreTotal)
	}

	second, err := svc.GradeSubmission(context.Background(), 1, "teacher_1", 10, "better")
	if err != nil {
		t.Fatalf("second GradeSubmission returned error: %v", err)
	}
	if second.GroupScoreTotal != 10 {
		t.Fatalf("expected regrade to use delta and total 10, got %d", second.GroupScoreTotal)
	}
}

func TestTaskContract_ResultsAndRanking(t *testing.T) {
	svc, rooms, _, _, _, submissions := newContractTaskService()

	score := 9
	gradedAt := time.Now().UTC()
	submittedAt := gradedAt.Add(-time.Minute)
	submissions.studentTasks = []repository.StudentTaskWithSubmission{
		{
			Task: domain.Task{
				ID:         1,
				Title:      "Ungraded",
				RoomID:     1,
				DeadlineAt: time.Now().Add(time.Hour),
				TargetType: domain.TargetAll,
				Status:     domain.TaskStatusPublished,
			},
			Submission: &domain.Submission{
				ID:          1,
				TaskID:      1,
				StudentID:   1,
				RoomID:      1,
				GroupID:     1,
				Status:      domain.SubmissionStatusSubmitted,
				SubmittedAt: submittedAt,
			},
		},
		{
			Task: domain.Task{
				ID:         2,
				Title:      "Graded",
				RoomID:     1,
				DeadlineAt: time.Now().Add(time.Hour),
				TargetType: domain.TargetAll,
				Status:     domain.TaskStatusPublished,
			},
			Submission: &domain.Submission{
				ID:          2,
				TaskID:      2,
				StudentID:   1,
				RoomID:      1,
				GroupID:     1,
				Status:      domain.SubmissionStatusGraded,
				Score:       &score,
				Comment:     "nice",
				SubmittedAt: submittedAt,
				GradedAt:    &gradedAt,
			},
		},
		{
			Task: domain.Task{
				ID:         3,
				Title:      "Not submitted",
				RoomID:     1,
				DeadlineAt: time.Now().Add(time.Hour),
				TargetType: domain.TargetAll,
				Status:     domain.TaskStatusPublished,
			},
		},
	}

	results, err := svc.GetStudentResults(context.Background(), "student_1")
	if err != nil {
		t.Fatalf("GetStudentResults returned error: %v", err)
	}
	if len(results.Results) != 3 {
		t.Fatalf("expected 3 result rows, got %d", len(results.Results))
	}
	if results.Results[0].SubmissionStatus != string(domain.SubmissionStatusSubmitted) {
		t.Fatalf("expected submitted status, got %s", results.Results[0].SubmissionStatus)
	}
	if results.Results[1].Score == nil || *results.Results[1].Score != 9 {
		t.Fatalf("expected graded score 9")
	}
	if results.Results[2].SubmissionStatus != "notSubmitted" {
		t.Fatalf("expected notSubmitted, got %s", results.Results[2].SubmissionStatus)
	}

	ranking, err := svc.GetRankingForStudent(context.Background(), "ABC123", "student_1")
	if err != nil {
		t.Fatalf("GetRankingForStudent returned error: %v", err)
	}
	if len(ranking.Entries) != 2 {
		t.Fatalf("expected 2 ranking entries, got %d", len(ranking.Entries))
	}
	if !ranking.Entries[0].IsMyGroup {
		t.Fatalf("expected first group to be marked as my group")
	}

	rooms.byCode["OTHER"] = &domain.Room{ID: 2, RoomCode: "OTHER", Status: domain.RoomStatusActive, TeacherToken: "teacher_2"}
	if _, err := svc.GetRankingForStudent(context.Background(), "OTHER", "student_1"); err == nil {
		t.Fatal("expected room access denied for another room")
	} else if errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("expected app access error, got repository not found")
	}
}
