package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"iclassroom/backend/internal/domain"
)

// TaskWithStats is a teacher-facing task read model with submission statistics.
type TaskWithStats struct {
	Task               domain.Task
	TargetGroupIDs     []int64
	SubmittedCount     int
	TargetStudentCount int
}

// TaskRepository persists tasks and their target groups.
type TaskRepository struct {
	db *sql.DB
}

// NewTaskRepository wires the repository to a database handle.
func NewTaskRepository(db *sql.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

const taskColumns = `id, room_id, title, description, attachment_url, deadline_at, target_type, status, created_at, updated_at`

func scanTask(s rowScanner) (*domain.Task, error) {
	var (
		t             domain.Task
		description   sql.NullString
		attachmentURL sql.NullString
	)

	if err := s.Scan(
		&t.ID,
		&t.RoomID,
		&t.Title,
		&description,
		&attachmentURL,
		&t.DeadlineAt,
		&t.TargetType,
		&t.Status,
		&t.CreatedAt,
		&t.UpdatedAt,
	); err != nil {
		return nil, err
	}

	if description.Valid {
		t.Description = description.String
	}
	if attachmentURL.Valid {
		t.AttachmentURL = attachmentURL.String
	}

	return &t, nil
}

func nullableString(v string) any {
	if v == "" {
		return nil
	}
	return v
}

// Create inserts a task and, when targetType is groups, its target group rows in
// one transaction. On success it sets task.ID.
func (r *TaskRepository) Create(ctx context.Context, task *domain.Task, targetGroupIDs []int64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("repository: begin task tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	const insertTask = `INSERT INTO tasks
(room_id, title, description, attachment_url, deadline_at, target_type, status)
VALUES (?, ?, ?, ?, ?, ?, ?)`

	res, err := tx.ExecContext(
		ctx,
		insertTask,
		task.RoomID,
		task.Title,
		nullableString(task.Description),
		nullableString(task.AttachmentURL),
		task.DeadlineAt,
		task.TargetType,
		task.Status,
	)
	if err != nil {
		return fmt.Errorf("repository: insert task: %w", err)
	}

	taskID, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("repository: task last insert id: %w", err)
	}
	task.ID = taskID

	if len(targetGroupIDs) > 0 {
		const insertTarget = `INSERT INTO task_target_groups (task_id, group_id) VALUES (?, ?)`
		for _, groupID := range targetGroupIDs {
			if _, err := tx.ExecContext(ctx, insertTarget, taskID, groupID); err != nil {
				return fmt.Errorf("repository: insert task target group: %w", err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("repository: commit task tx: %w", err)
	}

	return nil
}

// ListByRoomID returns all tasks in a room with submittedCount and
// targetStudentCount. It also includes targetGroupIds for group-targeted tasks.
func (r *TaskRepository) ListByRoomID(ctx context.Context, roomID int64) ([]TaskWithStats, error) {
	const q = `SELECT
t.id, t.room_id, t.title, t.description, t.attachment_url, t.deadline_at,
t.target_type, t.status, t.created_at, t.updated_at,
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
GROUP BY
t.id, t.room_id, t.title, t.description, t.attachment_url, t.deadline_at,
t.target_type, t.status, t.created_at, t.updated_at
ORDER BY t.created_at DESC, t.id DESC`

	rows, err := r.db.QueryContext(ctx, q, roomID)
	if err != nil {
		return nil, fmt.Errorf("repository: list tasks: %w", err)
	}
	defer func() { _ = rows.Close() }()

	out := make([]TaskWithStats, 0)
	for rows.Next() {
		var (
			t             domain.Task
			description   sql.NullString
			attachmentURL sql.NullString
			item          TaskWithStats
		)

		if err := rows.Scan(
			&t.ID,
			&t.RoomID,
			&t.Title,
			&description,
			&attachmentURL,
			&t.DeadlineAt,
			&t.TargetType,
			&t.Status,
			&t.CreatedAt,
			&t.UpdatedAt,
			&item.SubmittedCount,
			&item.TargetStudentCount,
		); err != nil {
			return nil, fmt.Errorf("repository: scan task list item: %w", err)
		}

		if description.Valid {
			t.Description = description.String
		}
		if attachmentURL.Valid {
			t.AttachmentURL = attachmentURL.String
		}

		item.Task = t
		targetGroupIDs, err := r.ListTargetGroupIDs(ctx, t.ID)
		if err != nil {
			return nil, err
		}
		item.TargetGroupIDs = targetGroupIDs

		out = append(out, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository: iterate task rows: %w", err)
	}

	return out, nil
}

// ListTargetGroupIDs returns the group ids targeted by a group-targeted task.
func (r *TaskRepository) ListTargetGroupIDs(ctx context.Context, taskID int64) ([]int64, error) {
	const q = `SELECT group_id FROM task_target_groups WHERE task_id = ? ORDER BY group_id`

	rows, err := r.db.QueryContext(ctx, q, taskID)
	if err != nil {
		return nil, fmt.Errorf("repository: list task target groups: %w", err)
	}
	defer func() { _ = rows.Close() }()

	groupIDs := make([]int64, 0)
	for rows.Next() {
		var groupID int64
		if err := rows.Scan(&groupID); err != nil {
			return nil, fmt.Errorf("repository: scan task target group: %w", err)
		}
		groupIDs = append(groupIDs, groupID)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository: iterate task target groups: %w", err)
	}

	return groupIDs, nil
}

// GetRoomByTaskID loads the room that owns the given task. It is used for
// teacher-token authorization on task-id based routes.
func (r *TaskRepository) GetRoomByTaskID(ctx context.Context, taskID int64) (*domain.Room, error) {
	const q = `SELECT r.id, r.room_code, r.title, r.status, r.group_count, r.group_capacity,
r.allow_choose_group, r.teacher_token, r.created_at, r.updated_at, r.ended_at
FROM rooms r
INNER JOIN tasks t ON t.room_id = r.id
WHERE t.id = ?`

	room, err := scanRoom(r.db.QueryRowContext(ctx, q, taskID))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("repository: get room by task id: %w", err)
	}
	return room, nil
}

// UpdateStatus updates a task lifecycle status.
func (r *TaskRepository) UpdateStatus(ctx context.Context, taskID int64, status domain.TaskStatus) error {
	const q = `UPDATE tasks SET status = ? WHERE id = ?`

	res, err := r.db.ExecContext(ctx, q, status, taskID)
	if err != nil {
		return fmt.Errorf("repository: update task status: %w", err)
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("repository: task rows affected: %w", err)
	}
	if affected == 0 {
		return ErrNotFound
	}

	return nil
}
