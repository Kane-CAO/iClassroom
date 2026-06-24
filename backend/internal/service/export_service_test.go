package service

import (
	"archive/zip"
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"iclassroom/backend/internal/domain"
	"iclassroom/backend/internal/repository"
)

type exportStore struct {
	room        *domain.Room
	tasks       []repository.TaskWithStats
	submissions map[int64][]repository.SubmissionWithStudent
}

func newExportStore() *exportStore {
	room := &domain.Room{
		ID:           1,
		RoomCode:     "ABC123",
		Title:        "Demo Class",
		Status:       domain.RoomStatusActive,
		TeacherToken: "teacher_1",
	}
	task := domain.Task{
		ID:         1,
		RoomID:     1,
		Title:      "Task 1",
		DeadlineAt: time.Now().Add(time.Hour),
		TargetType: domain.TargetAll,
		Status:     domain.TaskStatusPublished,
	}
	return &exportStore{
		room:  room,
		tasks: []repository.TaskWithStats{{Task: task}},
		submissions: map[int64][]repository.SubmissionWithStudent{
			1: {
				{
					Submission: domain.Submission{
						ID:          1,
						TaskID:      1,
						StudentID:   1,
						RoomID:      1,
						GroupID:     1,
						ContentText: "answer",
						Status:      domain.SubmissionStatusGraded,
						Comment:     "good",
						SubmittedAt: time.Date(2026, 6, 10, 10, 0, 0, 0, time.UTC),
						GradedAt:    ptrTime(time.Date(2026, 6, 10, 11, 0, 0, 0, time.UTC)),
						Images: []domain.SubmissionImage{
							{
								ID:        1,
								FileURL:   "http://localhost:8080/uploads/rooms/ABC123/tasks/1/students/1/image1.jpg",
								FilePath:  "",
								FileName:  "image1.jpg",
								FileSize:  8,
								MimeType:  "image/jpeg",
								CreatedAt: time.Now().UTC(),
							},
						},
					},
					Student: domain.Student{ID: 1, RoomID: 1, GroupID: 1, Nickname: "Tom"},
					Group:   domain.Group{ID: 1, RoomID: 1, GroupName: "第1组"},
				},
			},
		},
	}
}

func ptrTime(t time.Time) *time.Time { return &t }

func (s *exportStore) CreateRoomWithGroups(context.Context, *domain.Room) ([]domain.Group, error) {
	return nil, nil
}

func (s *exportStore) GetByRoomCode(_ context.Context, code string) (*domain.Room, error) {
	if s.room.RoomCode == code {
		return s.room, nil
	}
	return nil, repository.ErrNotFound
}

func (s *exportStore) GetByTeacherToken(_ context.Context, token string) (*domain.Room, error) {
	if s.room.TeacherToken == token {
		return s.room, nil
	}
	return nil, repository.ErrNotFound
}

func (s *exportStore) EndRoom(context.Context, int64, time.Time) error {
	return nil
}

func (s *exportStore) ListByRoomID(_ context.Context, roomID int64) ([]repository.TaskWithStats, error) {
	return s.tasks, nil
}

func (s *exportStore) GetByID(context.Context, int64) (*domain.Group, error) {
	return nil, repository.ErrNotFound
}

func (s *exportStore) Join(context.Context, int64, int64, string, string) (*domain.Student, error) {
	return nil, repository.ErrNotFound
}

func (s *exportStore) GetByClientToken(context.Context, string) (*domain.Student, error) {
	return nil, repository.ErrNotFound
}

func (s *exportStore) Create(context.Context, *domain.Task, []int64) error {
	return nil
}

func (s *exportStore) ListTargetGroupIDs(context.Context, int64) ([]int64, error) {
	return nil, nil
}

func (s *exportStore) GetRoomByTaskID(_ context.Context, taskID int64) (*domain.Room, error) {
	if taskID == 1 {
		return s.room, nil
	}
	return nil, repository.ErrNotFound
}

func (s *exportStore) UpdateStatus(context.Context, int64, domain.TaskStatus) error {
	return nil
}

func (s *exportStore) ListTasksForStudent(context.Context, int64, int64, int64) ([]repository.StudentTaskWithSubmission, error) {
	return nil, nil
}

func (s *exportStore) GetTargetedTaskForStudent(context.Context, int64, int64, int64) (*domain.Task, error) {
	return nil, repository.ErrNotFound
}

func (s *exportStore) CreateText(context.Context, int64, *domain.Student, string) (*domain.Submission, error) {
	return nil, nil
}

func (s *exportStore) CreateImages(context.Context, int64, []domain.SubmissionImage) ([]domain.SubmissionImage, error) {
	return nil, nil
}

func (s *exportStore) CreateFiles(context.Context, int64, []domain.SubmissionFile) ([]domain.SubmissionFile, error) {
	return nil, nil
}

func (s *exportStore) DeleteByID(context.Context, int64) error {
	return nil
}

func (s *exportStore) ListByTaskID(_ context.Context, taskID int64) ([]repository.SubmissionWithStudent, error) {
	return s.submissions[taskID], nil
}

func (s *exportStore) GetRoomBySubmissionID(context.Context, int64) (*domain.Room, error) {
	return s.room, nil
}

func (s *exportStore) GradeSubmission(context.Context, int64, int, string) (*domain.Submission, error) {
	return nil, nil
}

func (s *exportStore) ListLeaderboardByRoomID(context.Context, int64) ([]repository.LeaderboardItem, error) {
	return nil, nil
}

func (s *exportStore) GetCurrentTask(context.Context, int64) (*repository.DisplayTask, error) {
	return nil, repository.ErrNotFound
}

func (s *exportStore) ListFeaturedAnswers(context.Context, int64) ([]repository.FeaturedAnswerView, error) {
	return nil, nil
}

func (s *exportStore) CountStudents(context.Context, int64) (int, error) {
	return 0, nil
}

func (s *exportStore) ListGroupScores(context.Context, int64) ([]repository.GroupScore, error) {
	return nil, nil
}

func (s *exportStore) ListTaskCompletion(context.Context, int64) ([]repository.TaskCompletion, error) {
	return nil, nil
}

func (s *exportStore) ListSubmissionTimeline(context.Context, int64) ([]repository.SubmissionTimelinePoint, error) {
	return nil, nil
}

func TestExportService_SuccessWithImages(t *testing.T) {
	store := newExportStore()
	tmp := t.TempDir()
	imgPath := filepath.Join(tmp, "image1.jpg")
	if err := os.WriteFile(imgPath, []byte("jpegdata"), 0o644); err != nil {
		t.Fatalf("write image: %v", err)
	}
	store.submissions[1][0].Submission.Images[0].FilePath = imgPath

	res, err := NewExportService(store, store, store).Export(context.Background(), "ABC123", "teacher_1")
	if err != nil {
		t.Fatalf("Export error: %v", err)
	}
	if res.FileName != "export_room_ABC123.zip" {
		t.Fatalf("filename = %q", res.FileName)
	}

	zr, err := zip.NewReader(bytes.NewReader(res.Data), int64(len(res.Data)))
	if err != nil {
		t.Fatalf("open zip: %v", err)
	}
	if len(zr.File) != 2 {
		t.Fatalf("zip entries = %d, want 2", len(zr.File))
	}

	var xlsxData []byte
	foundImage := false
	for _, f := range zr.File {
		if f.Name == "submissions.xlsx" {
			rc, err := f.Open()
			if err != nil {
				t.Fatalf("open xlsx: %v", err)
			}
			xlsxData, err = io.ReadAll(rc)
			_ = rc.Close()
			if err != nil {
				t.Fatalf("read xlsx: %v", err)
			}
			continue
		}
		if f.Name == "images/task_1/group_1/student_Tom_image1.jpg" {
			foundImage = true
		}
	}
	if !foundImage {
		t.Fatalf("expected exported image entry")
	}

	xlsxReader, err := zip.NewReader(bytes.NewReader(xlsxData), int64(len(xlsxData)))
	if err != nil {
		t.Fatalf("open xlsx archive: %v", err)
	}
	var sheet []byte
	for _, f := range xlsxReader.File {
		if f.Name == "xl/worksheets/sheet1.xml" {
			rc, err := f.Open()
			if err != nil {
				t.Fatalf("open sheet: %v", err)
			}
			sheet, err = io.ReadAll(rc)
			_ = rc.Close()
			if err != nil {
				t.Fatalf("read sheet: %v", err)
			}
			break
		}
	}
	sheetText := string(sheet)
	for _, want := range []string{"ABC123", "Demo Class", "Tom", "Task 1", "answer", "image1.jpg", "good"} {
		if !strings.Contains(sheetText, want) {
			t.Fatalf("sheet xml missing %q", want)
		}
	}
}

func TestExportService_NoImages(t *testing.T) {
	store := newExportStore()
	store.submissions[1][0].Submission.Images = nil

	res, err := NewExportService(store, store, store).Export(context.Background(), "ABC123", "teacher_1")
	if err != nil {
		t.Fatalf("Export error: %v", err)
	}

	zr, err := zip.NewReader(bytes.NewReader(res.Data), int64(len(res.Data)))
	if err != nil {
		t.Fatalf("open zip: %v", err)
	}
	if len(zr.File) != 1 || zr.File[0].Name != "submissions.xlsx" {
		t.Fatalf("unexpected zip entries: %+v", zr.File)
	}
}

func TestExportService_MissingImageFile(t *testing.T) {
	store := newExportStore()
	tmp := t.TempDir()
	imgPath := filepath.Join(tmp, "image1.jpg")
	if err := os.WriteFile(imgPath, []byte("jpegdata"), 0o644); err != nil {
		t.Fatalf("write image: %v", err)
	}
	if err := os.Remove(imgPath); err != nil {
		t.Fatalf("remove image: %v", err)
	}
	store.submissions[1][0].Submission.Images[0].FilePath = imgPath

	_, err := NewExportService(store, store, store).Export(context.Background(), "ABC123", "teacher_1")
	wantCode(t, err, "IMAGE_FILE_MISSING")
}
