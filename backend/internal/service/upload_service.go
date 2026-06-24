package service

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"iclassroom/backend/internal/apperr"
	"iclassroom/backend/internal/domain"
	"iclassroom/backend/internal/storage"
)

const (
	maxSubmissionImages = 3
	maxSubmissionImageMB = 5
	maxSubmissionImageSize = maxSubmissionImageMB * 1024 * 1024
	maxSubmissionFiles = 5
	maxSubmissionFileMB = 20
	maxSubmissionFileSize = maxSubmissionFileMB * 1024 * 1024
)

// UploadedFile is one multipart image payload already read into memory by the
// HTTP handler.
type UploadedFile struct {
	FileName string
	MimeType string
	Data     []byte
}

// UploadService persists uploaded submission images and cleans them up on
// failure.
type UploadService interface {
	SaveSubmissionImages(ctx context.Context, roomCode string, taskID, studentID int64, files []UploadedFile) ([]domain.SubmissionImage, error)
	SaveSubmissionFiles(ctx context.Context, roomCode string, taskID, studentID int64, files []UploadedFile) ([]domain.SubmissionFile, error)
	DeleteSubmissionImages(ctx context.Context, images []domain.SubmissionImage) error
	DeleteSubmissionFiles(ctx context.Context, files []domain.SubmissionFile) error
}

// LocalUploadService stores files on the local filesystem.
type LocalUploadService struct {
	store *storage.LocalStorage
}

// NewLocalUploadService wires the uploader to local storage.
func NewLocalUploadService(store *storage.LocalStorage) *LocalUploadService {
	return &LocalUploadService{store: store}
}

func (s *LocalUploadService) SaveSubmissionImages(ctx context.Context, roomCode string, taskID, studentID int64, files []UploadedFile) ([]domain.SubmissionImage, error) {
	_ = ctx

	if len(files) == 0 {
		return nil, nil
	}
	if len(files) > maxSubmissionImages {
		return nil, apperr.TooManyImages()
	}

	saved := make([]storage.SavedFile, 0, len(files))
	out := make([]domain.SubmissionImage, 0, len(files))
	for idx, file := range files {
		mimeType := normalizeImageMimeType(file.MimeType)
		if mimeType == "" {
			return nil, apperr.InvalidImageType()
		}
		if len(file.Data) > maxSubmissionImageSize {
			s.deleteSavedFiles(saved)
			return nil, apperr.ImageTooLarge()
		}

		persisted, err := s.store.SaveSubmissionImage(roomCode, taskID, studentID, idx+1, mimeType, file.Data)
		if err != nil {
			s.deleteSavedFiles(saved)
			return nil, wrapUploadErr(err)
		}
		saved = append(saved, persisted)
		out = append(out, domain.SubmissionImage{
			FileURL:  persisted.FileURL,
			FilePath: persisted.FilePath,
			FileName: persisted.FileName,
			FileSize: persisted.FileSize,
			MimeType: persisted.MimeType,
		})
	}

	return out, nil
}

func (s *LocalUploadService) SaveSubmissionFiles(ctx context.Context, roomCode string, taskID, studentID int64, files []UploadedFile) ([]domain.SubmissionFile, error) {
	_ = ctx

	if len(files) == 0 {
		return nil, nil
	}
	if len(files) > maxSubmissionFiles {
		return nil, apperr.InvalidRequest("at most 5 files can be uploaded")
	}

	saved := make([]storage.SavedFile, 0, len(files))
	out := make([]domain.SubmissionFile, 0, len(files))
	for idx, file := range files {
		mimeType := normalizeFileMimeType(file.FileName, file.MimeType)
		if mimeType == "" {
			s.deleteSavedFiles(saved)
			return nil, apperr.InvalidFileType()
		}
		if len(file.Data) > maxSubmissionFileSize {
			s.deleteSavedFiles(saved)
			return nil, apperr.FileTooLarge()
		}

		persisted, err := s.store.SaveSubmissionFile(roomCode, taskID, studentID, idx+1, file.FileName, mimeType, file.Data)
		if err != nil {
			s.deleteSavedFiles(saved)
			return nil, wrapUploadErr(err)
		}
		saved = append(saved, persisted)
		out = append(out, domain.SubmissionFile{
			Kind:             "file",
			FileURL:          persisted.FileURL,
			FilePath:         persisted.FilePath,
			OriginalFileName: file.FileName,
			StoredFileName:   persisted.FileName,
			FileSize:         persisted.FileSize,
			MimeType:         persisted.MimeType,
		})
	}

	return out, nil
}

func (s *LocalUploadService) DeleteSubmissionImages(ctx context.Context, images []domain.SubmissionImage) error {
	_ = ctx
	if len(images) == 0 {
		return nil
	}

	files := make([]storage.SavedFile, 0, len(images))
	for _, image := range images {
		files = append(files, storage.SavedFile{FilePath: image.FilePath})
	}
	s.store.DeleteFiles(files)
	return nil
}

func (s *LocalUploadService) DeleteSubmissionFiles(ctx context.Context, attachments []domain.SubmissionFile) error {
	_ = ctx
	if len(attachments) == 0 {
		return nil
	}

	files := make([]storage.SavedFile, 0, len(attachments))
	for _, attachment := range attachments {
		files = append(files, storage.SavedFile{FilePath: attachment.FilePath})
	}
	s.store.DeleteFiles(files)
	return nil
}

func (s *LocalUploadService) deleteSavedFiles(files []storage.SavedFile) {
	s.store.DeleteFiles(files)
}

func normalizeImageMimeType(mimeType string) string {
	switch strings.ToLower(strings.TrimSpace(mimeType)) {
	case "image/jpeg", "image/jpg":
		return "image/jpeg"
	case "image/png":
		return "image/png"
	case "image/webp":
		return "image/webp"
	default:
		return ""
	}
}

func normalizeFileMimeType(fileName, mimeType string) string {
	mimeType = strings.ToLower(strings.TrimSpace(mimeType))
	if strings.HasPrefix(mimeType, "video/") {
		return ""
	}
	if normalizeImageMimeType(mimeType) != "" {
		return normalizeImageMimeType(mimeType)
	}

	ext := strings.ToLower(filepath.Ext(fileName))
	switch ext {
	case ".pdf":
		return "application/pdf"
	case ".doc":
		return "application/msword"
	case ".docx":
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case ".xls":
		return "application/vnd.ms-excel"
	case ".xlsx":
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	case ".ppt":
		return "application/vnd.ms-powerpoint"
	case ".pptx":
		return "application/vnd.openxmlformats-officedocument.presentationml.presentation"
	case ".zip":
		return "application/zip"
	case ".rar":
		return "application/vnd.rar"
	case ".7z":
		return "application/x-7z-compressed"
	case ".txt", ".md", ".csv", ".tsv":
		if mimeType != "" && !strings.HasPrefix(mimeType, "video/") {
			return mimeType
		}
		return "text/plain"
	default:
		return ""
	}
}

func wrapUploadErr(err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%w: %v", apperr.UploadFailed(), err)
}
