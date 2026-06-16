package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"

	"iclassroom/backend/internal/apperr"
	"iclassroom/backend/internal/domain"
	exportutil "iclassroom/backend/internal/export"
	"iclassroom/backend/internal/repository"
)

// ExportResult carries the finished zip archive and its suggested file name.
type ExportResult struct {
	FileName string
	Data     []byte
}

// ExportService builds the teacher export zip for one room.
type ExportService struct {
	rooms       RoomRepository
	tasks       TaskRepository
	submissions SubmissionRepository
}

func NewExportService(rooms RoomRepository, tasks TaskRepository, submissions SubmissionRepository) *ExportService {
	return &ExportService{rooms: rooms, tasks: tasks, submissions: submissions}
}

func (s *ExportService) Export(ctx context.Context, roomCode, teacherToken string) (*ExportResult, error) {
	room, err := verifyTeacherByRoomCode(ctx, s.rooms, roomCode, teacherToken)
	if err != nil {
		return nil, err
	}

	taskItems, err := s.tasks.ListByRoomID(ctx, room.ID)
	if err != nil {
		return nil, apperr.ExportFailed()
	}

	rows := make([]exportutil.SubmissionRow, 0)
	images := make([]exportutil.ImageFile, 0)

	for _, taskItem := range taskItems {
		submissions, err := s.submissions.ListByTaskID(ctx, taskItem.Task.ID)
		if err != nil {
			return nil, apperr.ExportFailed()
		}
		for _, item := range submissions {
			row, rowImages := exportSubmissionRow(room, taskItem.Task.Title, item)
			rows = append(rows, row)
			images = append(images, rowImages...)
		}
	}

	data, err := exportutil.BuildSubmissionsArchive(rows, images)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, apperr.ImageFileMissing()
		}
		if strings.Contains(err.Error(), "invalid archive path") {
			return nil, apperr.ExportFailed()
		}
		if errors.Is(err, repository.ErrNotFound) {
			return nil, apperr.ExportFailed()
		}
		if isMissingImageFile(err) {
			return nil, apperr.ImageFileMissing()
		}
		return nil, apperr.ExportFailed()
	}

	return &ExportResult{
		FileName: fmt.Sprintf("export_room_%s.zip", room.RoomCode),
		Data:     data,
	}, nil
}

func exportSubmissionRow(room *domain.Room, taskTitle string, item repository.SubmissionWithStudent) (exportutil.SubmissionRow, []exportutil.ImageFile) {
	imageFileNames := make([]string, 0, len(item.Submission.Images))
	images := make([]exportutil.ImageFile, 0, len(item.Submission.Images))
	for _, image := range item.Submission.Images {
		imageFileNames = append(imageFileNames, image.FileName)
		images = append(images, exportutil.ImageFile{
			ArchivePath: path.Join(
				"task_"+strconv.FormatInt(item.Submission.TaskID, 10),
				"group_"+strconv.FormatInt(item.Submission.GroupID, 10),
				"student_"+sanitizeSegment(item.Student.Nickname)+"_"+image.FileName,
			),
			FilePath: image.FilePath,
		})
	}

	row := exportutil.SubmissionRow{
		RoomCode:        room.RoomCode,
		RoomTitle:       room.Title,
		GroupName:       item.Group.GroupName,
		StudentNickname: item.Student.Nickname,
		TaskTitle:       taskTitle,
		ContentText:     item.Submission.ContentText,
		ImageFileNames:  strings.Join(imageFileNames, ", "),
		Comment:         item.Submission.Comment,
		SubmittedAt:     item.Submission.SubmittedAt.UTC().Format("2006-01-02T15:04:05Z"),
	}
	if item.Submission.Score != nil {
		row.Score = strconv.Itoa(*item.Submission.Score)
	}
	if item.Submission.GradedAt != nil {
		row.GradedAt = item.Submission.GradedAt.UTC().Format("2006-01-02T15:04:05Z")
	}

	return row, images
}

func sanitizeSegment(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return "unknown"
	}
	s = strings.ReplaceAll(s, "/", "_")
	s = strings.ReplaceAll(s, "\\", "_")
	s = strings.ReplaceAll(s, ":", "_")
	return s
}

func isMissingImageFile(err error) bool {
	return errors.Is(err, os.ErrNotExist)
}
