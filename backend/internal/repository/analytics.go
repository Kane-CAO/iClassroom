package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// GroupScore is a group score row for analytics.
type GroupScore struct {
	GroupID    int64
	GroupName  string
	ScoreTotal int
}

// TaskCompletion is a task completion row for analytics.
type TaskCompletion struct {
	TaskID             int64
	TaskTitle          string
	SubmittedCount     int
	TargetStudentCount int
}

// SubmissionTimelinePoint is an aggregated submission bucket.
type SubmissionTimelinePoint struct {
	Time  time.Time
	Count int
}

// AnalyticsRepository reads teacher analytics data.
type AnalyticsRepository struct {
	db *sql.DB
}

// NewAnalyticsRepository wires the repository to a database handle.
func NewAnalyticsRepository(db *sql.DB) *AnalyticsRepository {
	return &AnalyticsRepository{db: db}
}

func (r *AnalyticsRepository) CountStudents(ctx context.Context, roomID int64) (int, error) {
	const q = `SELECT COUNT(*) FROM students WHERE room_id = ?`
	var count int
	if err := r.db.QueryRowContext(ctx, q, roomID).Scan(&count); err != nil {
		return 0, fmt.Errorf("repository: count analytics students: %w", err)
	}
	return count, nil
}

func (r *AnalyticsRepository) ListGroupScores(ctx context.Context, roomID int64) ([]GroupScore, error) {
	const q = `SELECT id, group_name, score_total
FROM ` + "`groups`" + `
WHERE room_id = ?
ORDER BY score_total DESC, id ASC`

	rows, err := r.db.QueryContext(ctx, q, roomID)
	if err != nil {
		return nil, fmt.Errorf("repository: list analytics group scores: %w", err)
	}
	defer func() { _ = rows.Close() }()

	out := make([]GroupScore, 0)
	for rows.Next() {
		var item GroupScore
		if err := rows.Scan(&item.GroupID, &item.GroupName, &item.ScoreTotal); err != nil {
			return nil, fmt.Errorf("repository: scan analytics group score: %w", err)
		}
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository: iterate analytics group scores: %w", err)
	}
	return out, nil
}

func (r *AnalyticsRepository) ListTaskCompletion(ctx context.Context, roomID int64) ([]TaskCompletion, error) {
	const q = `SELECT
t.id,
t.title,
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
WHERE t.room_id = ?
GROUP BY t.id, t.title, t.target_type, t.room_id, t.created_at
ORDER BY t.created_at DESC, t.id DESC`

	rows, err := r.db.QueryContext(ctx, q, roomID)
	if err != nil {
		return nil, fmt.Errorf("repository: list analytics task completion: %w", err)
	}
	defer func() { _ = rows.Close() }()

	out := make([]TaskCompletion, 0)
	for rows.Next() {
		var item TaskCompletion
		if err := rows.Scan(&item.TaskID, &item.TaskTitle, &item.SubmittedCount, &item.TargetStudentCount); err != nil {
			return nil, fmt.Errorf("repository: scan analytics task completion: %w", err)
		}
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository: iterate analytics task completion: %w", err)
	}
	return out, nil
}

func (r *AnalyticsRepository) ListSubmissionTimeline(ctx context.Context, roomID int64) ([]SubmissionTimelinePoint, error) {
	const q = `SELECT
STR_TO_DATE(DATE_FORMAT(submitted_at, '%Y-%m-%d %H:%i:00'), '%Y-%m-%d %H:%i:%s') AS bucket_time,
COUNT(*) AS count
FROM submissions
WHERE room_id = ?
GROUP BY bucket_time
ORDER BY bucket_time ASC`

	rows, err := r.db.QueryContext(ctx, q, roomID)
	if err != nil {
		return nil, fmt.Errorf("repository: list analytics submission timeline: %w", err)
	}
	defer func() { _ = rows.Close() }()

	out := make([]SubmissionTimelinePoint, 0)
	for rows.Next() {
		var item SubmissionTimelinePoint
		if err := rows.Scan(&item.Time, &item.Count); err != nil {
			return nil, fmt.Errorf("repository: scan analytics submission timeline: %w", err)
		}
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository: iterate analytics submission timeline: %w", err)
	}
	return out, nil
}
