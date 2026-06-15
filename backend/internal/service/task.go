package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"iclassroom/backend/internal/apperr"
	"iclassroom/backend/internal/domain"
	"iclassroom/backend/internal/repository"
)

const maxTaskTitleLen = 255
const maxSubmissionTextLen = 5000

type TaskService struct {
	rooms       RoomRepository
	groups      GroupRepository
	students    StudentRepository
	tasks       TaskRepository
	submissions SubmissionRepository
}

func NewTaskService(
	rooms RoomRepository,
	groups GroupRepository,
	students StudentRepository,
	tasks TaskRepository,
	submissions SubmissionRepository,
) *TaskService {
	return &TaskService{
		rooms:       rooms,
		groups:      groups,
		students:    students,
		tasks:       tasks,
		submissions: submissions,
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

	return submission, nil
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

	return &TaskView{
		Task: domain.Task{
			ID:     taskID,
			RoomID: room.ID,
			Status: status,
		},
		RoomCode: room.RoomCode,
	}, nil
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
