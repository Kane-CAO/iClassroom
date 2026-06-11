package domain

import "time"

// SubmissionStatus is the lifecycle state of a submission.
type SubmissionStatus string

const (
	SubmissionStatusSubmitted SubmissionStatus = "submitted"
	SubmissionStatusGraded    SubmissionStatus = "graded"
)

// Valid reports whether s is a known submission status.
func (s SubmissionStatus) Valid() bool {
	switch s {
	case SubmissionStatusSubmitted, SubmissionStatusGraded:
		return true
	default:
		return false
	}
}

// Score bounds. 0 is explicitly disallowed; scores are integers in [1, 10].
const (
	ScoreMin = 1
	ScoreMax = 10
)

// IsValidScore reports whether score is a permitted grade (integer 1–10).
func IsValidScore(score int) bool {
	return score >= ScoreMin && score <= ScoreMax
}

// Submission is a student's answer to a task. A student may submit a given task
// at most once (enforced by a DB unique constraint on task_id + student_id).
//
// RoomID and GroupID are denormalized snapshots that simplify teacher queries,
// grouping and export. Score is nil until graded; when set it is in [1, 10]
// (enforced by both IsValidScore and a DB CHECK constraint).
type Submission struct {
	ID          int64            `json:"submissionId"`
	TaskID      int64            `json:"taskId"`
	StudentID   int64            `json:"studentId"`
	RoomID      int64            `json:"-"`
	GroupID     int64            `json:"groupId"`
	ContentText string           `json:"contentText"`
	Status      SubmissionStatus `json:"status"`
	Score       *int             `json:"score"`
	Comment     string           `json:"comment"`
	SubmittedAt time.Time        `json:"submittedAt"`
	GradedAt    *time.Time       `json:"gradedAt"`
	CreatedAt   time.Time        `json:"-"`
	UpdatedAt   time.Time        `json:"-"`

	// Images is populated by the repository/service when loading a submission
	// with its attachments; it is not a database column.
	Images []SubmissionImage `json:"images,omitempty"`
}

// SubmissionImage is a single uploaded image belonging to a submission.
// FilePath is the server-side storage location and is never exposed to clients;
// FileURL is the publicly accessible URL.
type SubmissionImage struct {
	ID           int64     `json:"imageId"`
	SubmissionID int64     `json:"-"`
	FileURL      string    `json:"fileUrl"`
	FilePath     string    `json:"-"`
	FileName     string    `json:"fileName"`
	FileSize     int64     `json:"fileSize"`
	MimeType     string    `json:"mimeType"`
	CreatedAt    time.Time `json:"-"`
}
