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

const maxTaskTitleLen = 255
const maxSubmissionTextLen = 5000

type TaskService struct {
	rooms       RoomRepository
	groups      GroupRepository
	students    StudentRepository
	tasks       TaskRepository
	submissions SubmissionRepository
	uploads     UploadService
	broadcaster EventBroadcaster
}

func NewTaskService(
	rooms RoomRepository,
	groups GroupRepository,
	students StudentRepository,
	tasks TaskRepository,
	submissions SubmissionRepository,
	broadcaster EventBroadcaster,
	uploads ...UploadService,
) *TaskService {
	var uploadSvc UploadService
	if len(uploads) > 0 {
		uploadSvc = uploads[0]
	}
	return &TaskService{
		rooms:       rooms,
		groups:      groups,
		students:    students,
		tasks:       tasks,
		submissions: submissions,
		uploads:     uploadSvc,
		broadcaster: resolveBroadcaster(broadcaster),
	}
}

type CreateTaskInput struct {
	RoomCode       string
	TeacherToken   string
	Title          string
	Description    string
	AttachmentURL  string
	DeadlineAt     time.Time
	TargetType     domain.TargetType
	TargetGroupIDs []int64
}

type TaskView struct {
	Task               domain.Task
	RoomCode           string
	TargetGroupIDs     []int64
	SubmittedCount     int
	TargetStudentCount int
}

type StudentTaskView struct {
	Task           domain.Task
	TargetGroupIDs []int64
	Submission     *domain.Submission
}

type SubmissionView struct {
	Submission domain.Submission
	Student    domain.Student
	Group      domain.Group
}

type GradeSubmissionResult struct {
	Submission      domain.Submission
	GroupScoreTotal int
}

type StudentResultItem struct {
	Task             domain.Task
	SubmissionStatus string
	Score            *int
	Comment          string
	SubmittedAt      *time.Time
	GradedAt         *time.Time
}

type StudentResultsView struct {
	Student domain.Student
	Group   domain.Group
	Results []StudentResultItem
}

func (s *TaskService) Create(ctx context.Context, in CreateTaskInput) (*TaskView, error) {
	room, err := s.verifyTeacherByRoomCode(ctx, in.RoomCode, in.TeacherToken)
	if err != nil {
		return nil, err
	}
	if room.Status == domain.RoomStatusEnded {
		return nil, apperr.RoomEnded()
	}

	title := strings.TrimSpace(in.Title)
	if title == "" || len([]rune(title)) > maxTaskTitleLen {
		return nil, apperr.InvalidTaskTitle()
	}

	if in.DeadlineAt.IsZero() || !in.DeadlineAt.After(time.Now().UTC()) {
		return nil, apperr.InvalidDeadline()
	}

	if !in.TargetType.Valid() {
		return nil, apperr.InvalidTargetType()
	}

	targetGroupIDs := make([]int64, 0)
	if in.TargetType == domain.TargetGroups {
		targetGroupIDs, err = s.validateTargetGroups(ctx, room.ID, in.TargetGroupIDs)
		if err != nil {
			return nil, err
		}
	}

	task := &domain.Task{
		RoomID:        room.ID,
		Title:         title,
		Description:   strings.TrimSpace(in.Description),
		AttachmentURL: strings.TrimSpace(in.AttachmentURL),
		DeadlineAt:    in.DeadlineAt.UTC(),
		TargetType:    in.TargetType,
		Status:        domain.TaskStatusPublished,
	}

	if err := s.tasks.Create(ctx, task, targetGroupIDs); err != nil {
		return nil, err
	}

	emit(s.broadcaster, room.RoomCode, websocket.EventTaskPublished, map[string]any{
		"taskId":     task.ID,
		"status":     task.Status,
		"targetType": task.TargetType,
	})

	return &TaskView{
		Task:           *task,
		RoomCode:       room.RoomCode,
		TargetGroupIDs: targetGroupIDs,
	}, nil
}

func (s *TaskService) ListForTeacher(ctx context.Context, roomCode, teacherToken string) ([]TaskView, error) {
	room, err := s.verifyTeacherByRoomCode(ctx, roomCode, teacherToken)
	if err != nil {
		return nil, err
	}

	items, err := s.tasks.ListByRoomID(ctx, room.ID)
	if err != nil {
		return nil, err
	}

	out := make([]TaskView, 0, len(items))
	for _, item := range items {
		out = append(out, TaskView{
			Task:               item.Task,
			RoomCode:           room.RoomCode,
			TargetGroupIDs:     item.TargetGroupIDs,
			SubmittedCount:     item.SubmittedCount,
			TargetStudentCount: item.TargetStudentCount,
		})
	}

	return out, nil
}

func (s *TaskService) ListForStudent(ctx context.Context, studentToken string) ([]StudentTaskView, error) {
	student, err := s.verifyStudent(ctx, studentToken)
	if err != nil {
		return nil, err
	}

	items, err := s.submissions.ListTasksForStudent(ctx, student.ID, student.RoomID, student.GroupID)
	if err != nil {
		return nil, err
	}

	out := make([]StudentTaskView, 0, len(items))
	for _, item := range items {
		out = append(out, StudentTaskView{
			Task:           item.Task,
			TargetGroupIDs: item.TargetGroupIDs,
			Submission:     item.Submission,
		})
	}

	return out, nil
}

func (s *TaskService) SubmitText(ctx context.Context, taskID int64, studentToken, contentText string) (*domain.Submission, error) {
	return s.SubmitWithImages(ctx, taskID, studentToken, contentText, nil)
}

func (s *TaskService) SubmitWithImages(ctx context.Context, taskID int64, studentToken, contentText string, images []UploadedFile) (*domain.Submission, error) {
	if taskID <= 0 {
		return nil, apperr.TaskNotFound()
	}

	student, err := s.verifyStudent(ctx, studentToken)
	if err != nil {
		return nil, err
	}

	content := strings.TrimSpace(contentText)
	if content == "" || len([]rune(content)) > maxSubmissionTextLen {
		return nil, apperr.InvalidSubmissionContent()
	}

	task, err := s.submissions.GetTargetedTaskForStudent(ctx, taskID, student.RoomID, student.GroupID)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, apperr.TaskNotFound()
	}
	if err != nil {
		return nil, err
	}

	room, err := s.tasks.GetRoomByTaskID(ctx, taskID)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, apperr.TaskNotFound()
	}
	if err != nil {
		return nil, err
	}
	if room.Status == domain.RoomStatusEnded {
		return nil, apperr.RoomEnded()
	}

	if task.Status != domain.TaskStatusPublished {
		return nil, apperr.TaskNotAcceptingSubmissions()
	}
	if !time.Now().UTC().Before(task.DeadlineAt.UTC()) {
		return nil, apperr.TaskNotAcceptingSubmissions()
	}

	submission, err := s.submissions.CreateText(ctx, taskID, student, content)
	if errors.Is(err, repository.ErrDuplicate) {
		return nil, apperr.SubmissionDuplicated()
	}
	if err != nil {
		return nil, err
	}

	if len(images) == 0 {
		s.emitSubmissionCreated(room.RoomCode, submission)
		return submission, nil
	}

	if s.uploads == nil {
		_ = s.submissions.DeleteByID(ctx, submission.ID)
		return nil, apperr.UploadFailed()
	}

	savedImages, err := s.uploads.SaveSubmissionImages(ctx, room.RoomCode, taskID, student.ID, images)
	if err != nil {
		_ = s.submissions.DeleteByID(ctx, submission.ID)
		return nil, err
	}

	storedImages, err := s.submissions.CreateImages(ctx, submission.ID, savedImages)
	if err != nil {
		_ = s.uploads.DeleteSubmissionImages(ctx, savedImages)
		_ = s.submissions.DeleteByID(ctx, submission.ID)
		return nil, err
	}

	submission.Images = storedImages
	s.emitSubmissionCreated(room.RoomCode, submission)
	return submission, nil
}

// emitSubmissionCreated broadcasts a submission_created event with the minimal
// identifiers clients need to refetch the affected task/submission view.
func (s *TaskService) emitSubmissionCreated(roomCode string, submission *domain.Submission) {
	emit(s.broadcaster, roomCode, websocket.EventSubmissionCreated, map[string]any{
		"submissionId": submission.ID,
		"taskId":       submission.TaskID,
		"studentId":    submission.StudentID,
		"groupId":      submission.GroupID,
	})
}

func (s *TaskService) ListSubmissionsForTeacher(ctx context.Context, taskID int64, teacherToken string) ([]SubmissionView, error) {
	if taskID <= 0 {
		return nil, apperr.TaskNotFound()
	}

	room, err := s.tasks.GetRoomByTaskID(ctx, taskID)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, apperr.TaskNotFound()
	}
	if err != nil {
		return nil, err
	}

	if err := s.verifyTeacherAgainstRoom(ctx, room, teacherToken); err != nil {
		return nil, err
	}

	items, err := s.submissions.ListByTaskID(ctx, taskID)
	if err != nil {
		return nil, err
	}

	out := make([]SubmissionView, 0, len(items))
	for _, item := range items {
		out = append(out, SubmissionView{
			Submission: item.Submission,
			Student:    item.Student,
			Group:      item.Group,
		})
	}

	return out, nil
}

func (s *TaskService) Pause(ctx context.Context, taskID int64, teacherToken string) (*TaskView, error) {
	return s.updateStatus(ctx, taskID, teacherToken, domain.TaskStatusPaused)
}

func (s *TaskService) Close(ctx context.Context, taskID int64, teacherToken string) (*TaskView, error) {
	return s.updateStatus(ctx, taskID, teacherToken, domain.TaskStatusClosed)
}

func (s *TaskService) updateStatus(ctx context.Context, taskID int64, teacherToken string, status domain.TaskStatus) (*TaskView, error) {
	if taskID <= 0 {
		return nil, apperr.TaskNotFound()
	}

	room, err := s.tasks.GetRoomByTaskID(ctx, taskID)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, apperr.TaskNotFound()
	}
	if err != nil {
		return nil, err
	}

	if err := s.verifyTeacherAgainstRoom(ctx, room, teacherToken); err != nil {
		return nil, err
	}

	if err := s.tasks.UpdateStatus(ctx, taskID, status); errors.Is(err, repository.ErrNotFound) {
		return nil, apperr.TaskNotFound()
	} else if err != nil {
		return nil, err
	}

	if evtType, ok := taskStatusEvent(status); ok {
		emit(s.broadcaster, room.RoomCode, evtType, map[string]any{
			"taskId": taskID,
			"status": status,
		})
	}

	return &TaskView{
		Task: domain.Task{
			ID:     taskID,
			RoomID: room.ID,
			Status: status,
		},
		RoomCode: room.RoomCode,
	}, nil
}

// taskStatusEvent maps a task status transition to its broadcast event type.
// The boolean is false for statuses that have no dedicated event.
func taskStatusEvent(status domain.TaskStatus) (websocket.EventType, bool) {
	switch status {
	case domain.TaskStatusPaused:
		return websocket.EventTaskPaused, true
	case domain.TaskStatusClosed:
		return websocket.EventTaskClosed, true
	default:
		return "", false
	}
}

func (s *TaskService) validateTargetGroups(ctx context.Context, roomID int64, groupIDs []int64) ([]int64, error) {
	if len(groupIDs) == 0 {
		return nil, apperr.InvalidTargetGroup()
	}

	seen := make(map[int64]struct{}, len(groupIDs))
	out := make([]int64, 0, len(groupIDs))

	for _, groupID := range groupIDs {
		if groupID <= 0 {
			return nil, apperr.InvalidTargetGroup()
		}
		if _, ok := seen[groupID]; ok {
			continue
		}

		group, err := s.groups.GetByID(ctx, groupID)
		if errors.Is(err, repository.ErrNotFound) {
			return nil, apperr.InvalidTargetGroup()
		}
		if err != nil {
			return nil, err
		}
		if group.RoomID != roomID {
			return nil, apperr.InvalidTargetGroup()
		}

		seen[groupID] = struct{}{}
		out = append(out, groupID)
	}

	if len(out) == 0 {
		return nil, apperr.InvalidTargetGroup()
	}

	return out, nil
}

func (s *TaskService) verifyStudent(ctx context.Context, studentToken string) (*domain.Student, error) {
	if studentToken == "" {
		return nil, apperr.InvalidStudentToken()
	}

	student, err := s.students.GetByClientToken(ctx, studentToken)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, apperr.InvalidStudentToken()
	}
	if err != nil {
		return nil, err
	}

	return student, nil
}

func (s *TaskService) verifyTeacherByRoomCode(ctx context.Context, roomCode, teacherToken string) (*domain.Room, error) {
	room, err := s.rooms.GetByRoomCode(ctx, roomCode)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, apperr.RoomNotFound()
	}
	if err != nil {
		return nil, err
	}

	if err := s.verifyTeacherAgainstRoom(ctx, room, teacherToken); err != nil {
		return nil, err
	}

	return room, nil
}

func (s *TaskService) verifyTeacherAgainstRoom(ctx context.Context, room *domain.Room, teacherToken string) error {
	if teacherToken == "" {
		return apperr.InvalidTeacherToken()
	}
	if teacherToken == room.TeacherToken {
		return nil
	}

	if _, err := s.rooms.GetByTeacherToken(ctx, teacherToken); err == nil {
		return apperr.RoomAccessDenied()
	} else if !errors.Is(err, repository.ErrNotFound) {
		return err
	}

	return apperr.InvalidTeacherToken()
}

type LeaderboardEntry struct {
	Rank         int
	Group        domain.Group
	CurrentCount int
	IsMyGroup    bool
}

type LeaderboardView struct {
	RoomCode string
	Entries  []LeaderboardEntry
}

func (s *TaskService) GradeSubmission(ctx context.Context, submissionID int64, teacherToken string, score int, comment string) (*GradeSubmissionResult, error) {
	if submissionID <= 0 {
		return nil, apperr.SubmissionNotFound()
	}

	if !domain.IsValidScore(score) {
		return nil, apperr.InvalidScore()
	}

	room, err := s.submissions.GetRoomBySubmissionID(ctx, submissionID)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, apperr.SubmissionNotFound()
	}
	if err != nil {
		return nil, err
	}

	if err := s.verifyTeacherAgainstRoom(ctx, room, teacherToken); err != nil {
		return nil, err
	}

	submission, err := s.submissions.GradeSubmission(ctx, submissionID, score, strings.TrimSpace(comment))
	if errors.Is(err, repository.ErrNotFound) {
		return nil, apperr.SubmissionNotFound()
	}
	if err != nil {
		return nil, err
	}

	group, err := s.groups.GetByID(ctx, submission.GroupID)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, apperr.SubmissionNotFound()
	}
	if err != nil {
		return nil, err
	}

	// Grade committed: the submission's score and the group's running total both
	// changed, so notify score listeners and refresh the leaderboard.
	emit(s.broadcaster, room.RoomCode, websocket.EventScoreUpdated, map[string]any{
		"submissionId": submission.ID,
		"score":        submission.Score,
		"groupId":      submission.GroupID,
	})
	emit(s.broadcaster, room.RoomCode, websocket.EventRankingUpdated, map[string]any{
		"groupId":         group.ID,
		"groupScoreTotal": group.ScoreTotal,
	})

	return &GradeSubmissionResult{
		Submission:      *submission,
		GroupScoreTotal: group.ScoreTotal,
	}, nil
}

func (s *TaskService) GetStudentResults(ctx context.Context, studentToken string) (*StudentResultsView, error) {
	student, err := s.verifyStudent(ctx, studentToken)
	if err != nil {
		return nil, err
	}

	group, err := s.groups.GetByID(ctx, student.GroupID)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, apperr.InvalidStudentToken()
	}
	if err != nil {
		return nil, err
	}

	items, err := s.submissions.ListTasksForStudent(ctx, student.ID, student.RoomID, student.GroupID)
	if err != nil {
		return nil, err
	}

	results := make([]StudentResultItem, 0, len(items))
	for _, item := range items {
		result := StudentResultItem{
			Task:             item.Task,
			SubmissionStatus: "notSubmitted",
		}

		if item.Submission != nil {
			result.SubmissionStatus = string(item.Submission.Status)
			result.Score = item.Submission.Score
			result.Comment = item.Submission.Comment
			submittedAt := item.Submission.SubmittedAt
			result.SubmittedAt = &submittedAt
			result.GradedAt = item.Submission.GradedAt
		}

		results = append(results, result)
	}

	return &StudentResultsView{
		Student: *student,
		Group:   *group,
		Results: results,
	}, nil
}

func (s *TaskService) GetLeaderboardForTeacher(ctx context.Context, roomCode, teacherToken string) (*LeaderboardView, error) {
	room, err := s.verifyTeacherByRoomCode(ctx, roomCode, teacherToken)
	if err != nil {
		return nil, err
	}

	items, err := s.submissions.ListLeaderboardByRoomID(ctx, room.ID)
	if err != nil {
		return nil, err
	}

	return buildLeaderboard(room.RoomCode, items, 0), nil
}

func (s *TaskService) GetLeaderboardForStudent(ctx context.Context, studentToken string) (*LeaderboardView, error) {
	student, err := s.verifyStudent(ctx, studentToken)
	if err != nil {
		return nil, err
	}

	items, err := s.submissions.ListLeaderboardByRoomID(ctx, student.RoomID)
	if err != nil {
		return nil, err
	}

	return buildLeaderboard("", items, student.GroupID), nil
}

func (s *TaskService) GetRankingForStudent(ctx context.Context, roomCode, studentToken string) (*LeaderboardView, error) {
	student, err := s.verifyStudent(ctx, studentToken)
	if err != nil {
		return nil, err
	}

	room, err := s.rooms.GetByRoomCode(ctx, roomCode)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, apperr.RoomNotFound()
	}
	if err != nil {
		return nil, err
	}

	if room.ID != student.RoomID {
		return nil, apperr.RoomAccessDenied()
	}

	items, err := s.submissions.ListLeaderboardByRoomID(ctx, room.ID)
	if err != nil {
		return nil, err
	}

	return buildLeaderboard(room.RoomCode, items, student.GroupID), nil
}

func buildLeaderboard(roomCode string, items []repository.LeaderboardItem, myGroupID int64) *LeaderboardView {
	entries := make([]LeaderboardEntry, 0, len(items))

	for i, item := range items {
		entries = append(entries, LeaderboardEntry{
			Rank:         i + 1,
			Group:        item.Group,
			CurrentCount: item.CurrentCount,
			IsMyGroup:    myGroupID > 0 && item.Group.ID == myGroupID,
		})
	}

	return &LeaderboardView{
		RoomCode: roomCode,
		Entries:  entries,
	}
}
