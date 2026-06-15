package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"iclassroom/backend/internal/domain"
)

// StudentTaskWithSubmission is the student-facing task read model.
// Submission is nil when the student has not submitted the task yet.
type StudentTaskWithSubmission struct {
	Task           domain.Task
	TargetGroupIDs []int64
	Submission     *domain.Submission
}

// SubmissionWithStudent is a teacher-facing submission read model.
type SubmissionWithStudent struct {
	Submission domain.Submission
	Student    domain.Student
	Group      domain.Group
}

// SubmissionRepository persists and reads student submissions.
type SubmissionRepository struct {
	db *sql.DB
}

// NewSubmissionRepository wires the repository to a database handle.
func NewSubmissionRepository(db *sql.DB) *SubmissionRepository {
	return &SubmissionRepository{db: db}
}

func scanSubmissionFields(
	id *int64,
	taskID *int64,
	studentID *int64,
	roomID *int64,
	groupID *int64,
	contentText *string,
	status *domain.SubmissionStatus,
	score **int,
	comment *string,
	submittedAt any,
	gradedAt **sql.NullTime,
	createdAt any,
	updatedAt any,
	values ...any,
) error {
	return nil
}

func scanSubmissionRow(
	id int64,
	taskID int64,
	studentID int64,
	roomID int64,
	groupID int64,
	contentText sql.NullString,
	status domain.SubmissionStatus,
	score sql.NullInt64,
	comment sql.NullString,
	submittedAt any,
	gradedAt sql.NullTime,
	createdAt any,
	updatedAt any,
) domain.Submission {
	sub := domain.Submission{
		ID:        id,
		TaskID:    taskID,
		StudentID: studentID,
		RoomID:    roomID,
		GroupID:   groupID,
		Status:    status,
		Comment:   "",
	}
	if contentText.Valid {
		sub.ContentText = contentText.String
	}
	if score.Valid {
		v := int(score.Int64)
		sub.Score = &v
	}
	if comment.Valid {
		sub.Comment = comment.String
	}
	return sub
}

// ListTasksForStudent returns every task assigned to the student, together with
// the student's existing submission when present.
func (r *SubmissionRepository) ListTasksForStudent(ctx context.Context, studentID, roomID, groupID int64) ([]StudentTaskWithSubmission, error) {
	const q = `SELECT
t.id, t.room_id, t.title, t.description, t.attachment_url, t.deadline_at,
t.target_type, t.status, t.created_at, t.updated_at,
s.id, s.task_id, s.student_id, s.room_id, s.group_id, s.content_text,
s.status, s.score, s.comment, s.submitted_at, s.graded_at, s.created_at, s.updated_at
FROM tasks t
LEFT JOIN submissions s ON s.task_id = t.id AND s.student_id = ?
WHERE t.room_id = ?
AND (
t.target_type = 'all'
OR EXISTS (
SELECT 1 FROM task_target_groups ttg
WHERE ttg.task_id = t.id AND ttg.group_id = ?
)
)
ORDER BY t.created_at DESC, t.id DESC`

	rows, err := r.db.QueryContext(ctx, q, studentID, roomID, groupID)
	if err != nil {
		return nil, fmt.Errorf("repository: list student tasks: %w", err)
	}
	defer func() { _ = rows.Close() }()

	out := make([]StudentTaskWithSubmission, 0)
	for rows.Next() {
		var (
			task          domain.Task
			description   sql.NullString
			attachmentURL sql.NullString

			submissionID        sql.NullInt64
			submissionTaskID    sql.NullInt64
			submissionStudentID sql.NullInt64
			submissionRoomID    sql.NullInt64
			submissionGroupID   sql.NullInt64
			contentText         sql.NullString
			submissionStatus    sql.NullString
			score               sql.NullInt64
			comment             sql.NullString
			submittedAt         sql.NullTime
			gradedAt            sql.NullTime
			submissionCreatedAt sql.NullTime
			submissionUpdatedAt sql.NullTime
		)

		if err := rows.Scan(
			&task.ID,
			&task.RoomID,
			&task.Title,
			&description,
			&attachmentURL,
			&task.DeadlineAt,
			&task.TargetType,
			&task.Status,
			&task.CreatedAt,
			&task.UpdatedAt,
			&submissionID,
			&submissionTaskID,
			&submissionStudentID,
			&submissionRoomID,
			&submissionGroupID,
			&contentText,
			&submissionStatus,
			&score,
			&comment,
			&submittedAt,
			&gradedAt,
			&submissionCreatedAt,
			&submissionUpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("repository: scan student task: %w", err)
		}

		if description.Valid {
			task.Description = description.String
		}
		if attachmentURL.Valid {
			task.AttachmentURL = attachmentURL.String
		}

		item := StudentTaskWithSubmission{Task: task}

		targetGroupIDs, err := r.listTargetGroupIDs(ctx, task.ID)
		if err != nil {
			return nil, err
		}
		item.TargetGroupIDs = targetGroupIDs

		if submissionID.Valid {
			sub := domain.Submission{
				ID:          submissionID.Int64,
				TaskID:      submissionTaskID.Int64,
				StudentID:   submissionStudentID.Int64,
				RoomID:      submissionRoomID.Int64,
				GroupID:     submissionGroupID.Int64,
				Status:      domain.SubmissionStatus(submissionStatus.String),
				ContentText: "",
				Comment:     "",
			}
			if contentText.Valid {
				sub.ContentText = contentText.String
			}
			if score.Valid {
				scoreValue := int(score.Int64)
				sub.Score = &scoreValue
			}
			if comment.Valid {
				sub.Comment = comment.String
			}
			if submittedAt.Valid {
				sub.SubmittedAt = submittedAt.Time
			}
			if gradedAt.Valid {
				t := gradedAt.Time
				sub.GradedAt = &t
			}
			if submissionCreatedAt.Valid {
				sub.CreatedAt = submissionCreatedAt.Time
			}
			if submissionUpdatedAt.Valid {
				sub.UpdatedAt = submissionUpdatedAt.Time
			}
			item.Submission = &sub
		}

		out = append(out, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository: iterate student tasks: %w", err)
	}

	return out, nil
}

// GetTargetedTaskForStudent returns a task only when it belongs to the student's
// room and targets the student's group.
func (r *SubmissionRepository) GetTargetedTaskForStudent(ctx context.Context, taskID, roomID, groupID int64) (*domain.Task, error) {
	const q = `SELECT id, room_id, title, description, attachment_url, deadline_at, target_type, status, created_at, updated_at
FROM tasks t
WHERE t.id = ?
AND t.room_id = ?
AND (
t.target_type = 'all'
OR EXISTS (
SELECT 1 FROM task_target_groups ttg
WHERE ttg.task_id = t.id AND ttg.group_id = ?
)
)`

	task, err := scanTask(r.db.QueryRowContext(ctx, q, taskID, roomID, groupID))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("repository: get targeted task for student: %w", err)
	}

	return task, nil
}

// CreateText inserts a text-only submission. The database unique constraint
// guarantees that one student can submit a task only once.
func (r *SubmissionRepository) CreateText(ctx context.Context, taskID int64, student *domain.Student, contentText string) (*domain.Submission, error) {
	const q = `INSERT INTO submissions (task_id, student_id, room_id, group_id, content_text, status)
VALUES (?, ?, ?, ?, ?, ?)`

	res, err := r.db.ExecContext(
		ctx,
		q,
		taskID,
		student.ID,
		student.RoomID,
		student.GroupID,
		contentText,
		domain.SubmissionStatusSubmitted,
	)
	if err != nil {
		if isDuplicateKey(err) {
			return nil, ErrDuplicate
		}
		return nil, fmt.Errorf("repository: insert text submission: %w", err)
	}

	submissionID, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("repository: submission last insert id: %w", err)
	}

	const getQ = `SELECT id, task_id, student_id, room_id, group_id, content_text, status,
score, comment, submitted_at, graded_at, created_at, updated_at
FROM submissions
WHERE id = ?`

	var (
		sub           domain.Submission
		contentTextDB sql.NullString
		score         sql.NullInt64
		comment       sql.NullString
		gradedAt      sql.NullTime
	)
	err = r.db.QueryRowContext(ctx, getQ, submissionID).Scan(
		&sub.ID,
		&sub.TaskID,
		&sub.StudentID,
		&sub.RoomID,
		&sub.GroupID,
		&contentTextDB,
		&sub.Status,
		&score,
		&comment,
		&sub.SubmittedAt,
		&gradedAt,
		&sub.CreatedAt,
		&sub.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("repository: get inserted submission: %w", err)
	}

	if contentTextDB.Valid {
		sub.ContentText = contentTextDB.String
	}
	if score.Valid {
		scoreValue := int(score.Int64)
		sub.Score = &scoreValue
	}
	if comment.Valid {
		sub.Comment = comment.String
	}
	if gradedAt.Valid {
		t := gradedAt.Time
		sub.GradedAt = &t
	}

	return &sub, nil
}

// ListByTaskID returns all submissions for one task, including student and group
// display data for the teacher view.
func (r *SubmissionRepository) ListByTaskID(ctx context.Context, taskID int64) ([]SubmissionWithStudent, error) {
	const q = `SELECT
s.id, s.task_id, s.student_id, s.room_id, s.group_id, s.content_text,
s.status, s.score, s.comment, s.submitted_at, s.graded_at, s.created_at, s.updated_at,
st.id, st.room_id, st.group_id, st.nickname, st.client_token, st.created_at, st.updated_at,
g.id, g.room_id, g.group_name, g.capacity, g.score_total, g.created_at, g.updated_at
FROM submissions s
INNER JOIN students st ON st.id = s.student_id
INNER JOIN ` + "`groups`" + ` g ON g.id = s.group_id
WHERE s.task_id = ?
ORDER BY s.submitted_at DESC, s.id DESC`

	rows, err := r.db.QueryContext(ctx, q, taskID)
	if err != nil {
		return nil, fmt.Errorf("repository: list submissions by task: %w", err)
	}
	defer func() { _ = rows.Close() }()

	out := make([]SubmissionWithStudent, 0)
	for rows.Next() {
		var (
			item        SubmissionWithStudent
			contentText sql.NullString
			score       sql.NullInt64
			comment     sql.NullString
			gradedAt    sql.NullTime
		)

		if err := rows.Scan(
			&item.Submission.ID,
			&item.Submission.TaskID,
			&item.Submission.StudentID,
			&item.Submission.RoomID,
			&item.Submission.GroupID,
			&contentText,
			&item.Submission.Status,
			&score,
			&comment,
			&item.Submission.SubmittedAt,
			&gradedAt,
			&item.Submission.CreatedAt,
			&item.Submission.UpdatedAt,
			&item.Student.ID,
			&item.Student.RoomID,
			&item.Student.GroupID,
			&item.Student.Nickname,
			&item.Student.ClientToken,
			&item.Student.CreatedAt,
			&item.Student.UpdatedAt,
			&item.Group.ID,
			&item.Group.RoomID,
			&item.Group.GroupName,
			&item.Group.Capacity,
			&item.Group.ScoreTotal,
			&item.Group.CreatedAt,
			&item.Group.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("repository: scan submission with student: %w", err)
		}

		if contentText.Valid {
			item.Submission.ContentText = contentText.String
		}
		if score.Valid {
			scoreValue := int(score.Int64)
			item.Submission.Score = &scoreValue
		}
		if comment.Valid {
			item.Submission.Comment = comment.String
		}
		if gradedAt.Valid {
			t := gradedAt.Time
			item.Submission.GradedAt = &t
		}

		out = append(out, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository: iterate submissions: %w", err)
	}

	return out, nil
}

func (r *SubmissionRepository) listTargetGroupIDs(ctx context.Context, taskID int64) ([]int64, error) {
	const q = `SELECT group_id FROM task_target_groups WHERE task_id = ? ORDER BY group_id`

	rows, err := r.db.QueryContext(ctx, q, taskID)
	if err != nil {
		return nil, fmt.Errorf("repository: list task target groups for student tasks: %w", err)
	}
	defer func() { _ = rows.Close() }()

	groupIDs := make([]int64, 0)
	for rows.Next() {
		var groupID int64
		if err := rows.Scan(&groupID); err != nil {
			return nil, fmt.Errorf("repository: scan task target group for student tasks: %w", err)
		}
		groupIDs = append(groupIDs, groupID)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository: iterate task target groups for student tasks: %w", err)
	}

	return groupIDs, nil
}
