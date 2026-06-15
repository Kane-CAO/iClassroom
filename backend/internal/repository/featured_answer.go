package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"iclassroom/backend/internal/domain"
)

// FeaturedAnswerRepository persists featured-answer markers.
type FeaturedAnswerRepository struct {
	db *sql.DB
}

// NewFeaturedAnswerRepository wires the repository to a database handle.
func NewFeaturedAnswerRepository(db *sql.DB) *FeaturedAnswerRepository {
	return &FeaturedAnswerRepository{db: db}
}

// GetRoomBySubmissionID loads the room that owns a submission.
func (r *FeaturedAnswerRepository) GetRoomBySubmissionID(ctx context.Context, submissionID int64) (*domain.Room, error) {
	const q = `SELECT rm.id, rm.room_code, rm.title, rm.status, rm.group_count, rm.group_capacity,
rm.allow_choose_group, rm.teacher_token, rm.created_at, rm.updated_at, rm.ended_at
FROM rooms rm
INNER JOIN submissions s ON s.room_id = rm.id
WHERE s.id = ?`

	room, err := scanRoom(r.db.QueryRowContext(ctx, q, submissionID))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("repository: get room by featured submission id: %w", err)
	}
	return room, nil
}

// Upsert marks a submission as featured and returns the persisted marker.
func (r *FeaturedAnswerRepository) Upsert(ctx context.Context, roomID, submissionID int64, mode domain.DisplayMode) (*domain.FeaturedAnswer, error) {
	const q = `INSERT INTO featured_answers (room_id, submission_id, display_mode)
VALUES (?, ?, ?)
ON DUPLICATE KEY UPDATE display_mode = VALUES(display_mode)`

	if _, err := r.db.ExecContext(ctx, q, roomID, submissionID, mode); err != nil {
		return nil, fmt.Errorf("repository: upsert featured answer: %w", err)
	}

	return r.GetBySubmissionID(ctx, submissionID)
}

// GetBySubmissionID loads a featured marker by submission id.
func (r *FeaturedAnswerRepository) GetBySubmissionID(ctx context.Context, submissionID int64) (*domain.FeaturedAnswer, error) {
	const q = `SELECT id, room_id, submission_id, display_mode, created_at
FROM featured_answers
WHERE submission_id = ?`

	var fa domain.FeaturedAnswer
	err := r.db.QueryRowContext(ctx, q, submissionID).Scan(
		&fa.ID,
		&fa.RoomID,
		&fa.SubmissionID,
		&fa.DisplayMode,
		&fa.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("repository: get featured answer by submission id: %w", err)
	}
	return &fa, nil
}
