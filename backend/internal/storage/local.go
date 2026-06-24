package storage

import (
	"fmt"
	"mime"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// SavedFile describes one uploaded file persisted on disk.
type SavedFile struct {
	FileURL  string
	FilePath string
	FileName string
	FileSize int64
	MimeType string
}

// LocalStorage stores uploaded files in a local filesystem tree and exposes
// them under a public base URL.
type LocalStorage struct {
	rootDir       string
	publicBaseURL string
}

// NewLocalStorage constructs a local file storage rooted at rootDir.
func NewLocalStorage(rootDir, publicBaseURL string) *LocalStorage {
	return &LocalStorage{
		rootDir:       filepath.Clean(rootDir),
		publicBaseURL: strings.TrimRight(publicBaseURL, "/"),
	}
}

// SaveSubmissionImage writes one image to the local uploads tree and returns
// the persisted metadata.
func (s *LocalStorage) SaveSubmissionImage(roomCode string, taskID, studentID int64, index int, mimeType string, data []byte) (SavedFile, error) {
	ext, err := extensionForMimeType(mimeType)
	if err != nil {
		return SavedFile{}, err
	}

	fileName := fmt.Sprintf("image%d%s", index, ext)
	relDir := filepath.Join("rooms", roomCode, "tasks", fmt.Sprintf("%d", taskID), "students", fmt.Sprintf("%d", studentID))
	absDir := filepath.Join(s.rootDir, relDir)
	if err := os.MkdirAll(absDir, 0o755); err != nil {
		return SavedFile{}, fmt.Errorf("storage: create upload dir: %w", err)
	}

	absPath := filepath.Join(absDir, fileName)
	if err := os.WriteFile(absPath, data, 0o644); err != nil {
		return SavedFile{}, fmt.Errorf("storage: write upload file: %w", err)
	}

	relPath := filepath.ToSlash(filepath.Join(relDir, fileName))
	fileURL := s.publicBaseURL + "/uploads/" + relPath

	return SavedFile{
		FileURL:  fileURL,
		FilePath: absPath,
		FileName: fileName,
		FileSize: int64(len(data)),
		MimeType: mimeType,
	}, nil
}

// SaveSubmissionFile writes one non-image attachment to the local uploads tree.
func (s *LocalStorage) SaveSubmissionFile(roomCode string, taskID, studentID int64, index int, originalName, mimeType string, data []byte) (SavedFile, error) {
	ext := strings.ToLower(filepath.Ext(originalName))
	if ext == "" {
		if exts, _ := mime.ExtensionsByType(mimeType); len(exts) > 0 {
			ext = strings.ToLower(exts[0])
		}
	}
	if ext == "" {
		ext = ".bin"
	}

	base := sanitizeUploadName(strings.TrimSuffix(filepath.Base(originalName), filepath.Ext(originalName)))
	if base == "" {
		base = "file"
	}
	fileName := fmt.Sprintf("file%d_%s%s", index, base, ext)
	relDir := filepath.Join("rooms", roomCode, "tasks", fmt.Sprintf("%d", taskID), "students", fmt.Sprintf("%d", studentID), "files")
	absDir := filepath.Join(s.rootDir, relDir)
	if err := os.MkdirAll(absDir, 0o755); err != nil {
		return SavedFile{}, fmt.Errorf("storage: create upload dir: %w", err)
	}

	absPath := filepath.Join(absDir, fileName)
	if err := os.WriteFile(absPath, data, 0o644); err != nil {
		return SavedFile{}, fmt.Errorf("storage: write upload file: %w", err)
	}

	relPath := filepath.ToSlash(filepath.Join(relDir, fileName))
	fileURL := s.publicBaseURL + "/uploads/" + relPath

	return SavedFile{
		FileURL:  fileURL,
		FilePath: absPath,
		FileName: fileName,
		FileSize: int64(len(data)),
		MimeType: mimeType,
	}, nil
}

// DeleteFiles removes uploaded files best-effort. Missing files are ignored.
func (s *LocalStorage) DeleteFiles(files []SavedFile) {
	for _, file := range files {
		_ = os.Remove(file.FilePath)
	}
}

func extensionForMimeType(mimeType string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(mimeType)) {
	case "image/jpeg":
		return ".jpg", nil
	case "image/png":
		return ".png", nil
	case "image/webp":
		return ".webp", nil
	}

	if exts, _ := mime.ExtensionsByType(mimeType); len(exts) > 0 {
		for _, ext := range exts {
			switch strings.ToLower(ext) {
			case ".jpg", ".jpeg", ".png", ".webp":
				if ext == ".jpeg" {
					return ".jpg", nil
				}
				return ext, nil
			}
		}
	}

	return "", fmt.Errorf("storage: unsupported mime type %q", mimeType)
}

var unsafeUploadName = regexp.MustCompile(`[^A-Za-z0-9._-]+`)

func sanitizeUploadName(name string) string {
	name = strings.TrimSpace(name)
	name = unsafeUploadName.ReplaceAllString(name, "_")
	name = strings.Trim(name, "._-")
	if len(name) > 80 {
		name = name[:80]
	}
	return name
}
