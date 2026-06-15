package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"iclassroom/backend/internal/domain"
)

// DisplayTask is the current task read model for the big-screen display.
type DisplayTask struct {
	TaskID             int64
	Title              string
	DeadlineAt         time.Time
	SubmittedCount     int
	TargetStudentCount int
}

// FeaturedAnswerView is a featured answer with display data.
type FeaturedAnswerView struct {
	FeaturedAnswer domain.FeaturedAnswer
	TaskID         int64
	GroupID        int64
	GroupName      string
	ContentText    string
	Score          *int
	SubmittedAt    time.Time
}

// DisplayRepository reads big-screen display data.
type DisplayRepository struct {
	db *sql.DB
}

// NewDisplayRepository wires the repository to a database handle.
func NewDisplayRepository(db *sql.DB) *DisplayRepository {
	return &DisplayRepository{db: db}
}

// GetCurrentTask returns the newest non-closed task in a room with completion stats.
func (r *DisplayRepository) GetCurrentTask(ctx context.Context, roomID int64) (*DisplayTask, error) {
	const q = `SELECT
t.id,
t.title,
t.deadline_at,
COUNT(DISTINCT s.id) AS submitted_count,
CASE
WHEN t.target_type = 'all' THEN (
SELECT COUNT(*) FROM students st WHERE st.room_id = t.room_id
)
ELSE (
SELECT COUNT(*)
FROM students st
INNER JOIN task_target_groups ttg ON ttg.group_id = st.group_id
WHERE ttg.task_id = t.id
)
END AS target_student_count
FROM tasks t
LEFT JOIN submissions s ON s.task_id = t.id
WHERE t.room_id = ? AND t.status <> 'closed'
GROUP BY t.id, t.title, t.deadline_at, t.target_type, t.room_id, t.created_at
ORDER BY t.created_at DESC, t.id DESC
LIMIT 1`

	var task DisplayTask
	err := r.db.QueryRowContext(ctx, q, roomID).Scan(
		&task.TaskID,
		&task.Title,
		&task.DeadlineAt,
		&task.SubmittedCount,
		&task.TargetStudentCount,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("repository: get display current task: %w", err)
	}
	return &task, nil
}

// ListFeaturedAnswers returns featured answers for a room, newest first.
func (r *DisplayRepository) ListFeaturedAnswers(ctx context.Context, roomID int64) ([]FeaturedAnswerView, error) {
	const q = `SELECT
fa.id, fa.room_id, fa.submission_id, fa.display_mode, fa.created_at,
s.task_id, s.group_id, s.content_text, s.score, s.submitted_at,
g.group_name
FROM featured_answers fa
INNER JOIN submissions s ON s.id = fa.submission_id
INNER JOIN ` + "`groups`" + ` g ON g.id = s.group_id
WHERE fa.room_id = ?
ORDER BY fa.created_at DESC, fa.id DESC`

	rows, err := r.db.QueryContext(ctx, q, roomID)
	if err != nil {
		return nil, fmt.Errorf("repository: list display featured answers: %w", err)
	}
	defer func() { _ = rows.Close() }()

	out := make([]FeaturedAnswerView, 0)
	for rows.Next() {
		var (
			item        FeaturedAnswerView
			contentText sql.NullString
			score       sql.NullInt64
		)
		if err := rows.Scan(
			&item.FeaturedAnswer.ID,
			&item.FeaturedAnswer.RoomID,
			&item.FeaturedAnswer.SubmissionID,
			&item.FeaturedAnswer.DisplayMode,
			&item.FeaturedAnswer.CreatedAt,
			&item.TaskID,
			&item.GroupID,
			&contentText,
			&score,
			&item.SubmittedAt,
			&item.GroupName,
		); err != nil {
			return nil, fmt.Errorf("repository: scan display featured answer: %w", err)
		}
		if contentText.Valid {
			item.ContentText = contentText.String
		}
		if score.Valid {
			scoreValue := int(score.Int64)
			item.Score = &scoreValue
		}
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository: iterate display featured answers: %w", err)
	}
	return out, nil
}
