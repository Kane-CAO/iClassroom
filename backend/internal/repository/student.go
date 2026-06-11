package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"iclassroom/backend/internal/domain"
)

// StudentRepository persists students and resolves them by client token.
type StudentRepository struct {
	db *sql.DB
}

// NewStudentRepository wires the repository to a database handle.
func NewStudentRepository(db *sql.DB) *StudentRepository {
	return &StudentRepository{db: db}
}

// Join atomically admits a student to a group within a room. It runs in a
// transaction and locks the target group row (SELECT … FOR UPDATE) so that
// concurrent joins to the same group cannot exceed capacity.
//
// Error mapping:
//   - group missing / not in this room        -> ErrNotFound
//   - group already at capacity               -> ErrGroupFull
//   - nickname taken (room_id+nickname unique) -> ErrDuplicate
//
// On success it returns the created student (ID populated).
func (r *StudentRepository) Join(ctx context.Context, roomID, groupID int64, nickname, clientToken string) (*domain.Student, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("repository: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// Lock the group row and read its capacity. Scoping by room_id enforces
	// "group must belong to this room" in the same query.
	var capacity int
	const lockGroup = "SELECT capacity FROM `groups` WHERE id = ? AND room_id = ? FOR UPDATE"
	err = tx.QueryRowContext(ctx, lockGroup, groupID, roomID).Scan(&capacity)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("repository: lock group: %w", err)
	}

	// Count current members under the lock; reject if full.
	var count int
	const countMembers = `SELECT COUNT(*) FROM students WHERE group_id = ?`
	if err := tx.QueryRowContext(ctx, countMembers, groupID).Scan(&count); err != nil {
		return nil, fmt.Errorf("repository: count group members: %w", err)
	}
	if count >= capacity {
		return nil, ErrGroupFull
	}

	const insertStudent = `INSERT INTO students (room_id, group_id, nickname, client_token)
		VALUES (?, ?, ?, ?)`
	res, err := tx.ExecContext(ctx, insertStudent, roomID, groupID, nickname, clientToken)
	if err != nil {
		if isDuplicateKey(err) {
			// room_id+nickname (or, vanishingly unlikely, client_token) collided.
			return nil, ErrDuplicate
		}
		return nil, fmt.Errorf("repository: insert student: %w", err)
	}
	studentID, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("repository: student last insert id: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("repository: commit: %w", err)
	}

	return &domain.Student{
		ID:          studentID,
		RoomID:      roomID,
		GroupID:     groupID,
		Nickname:    nickname,
		ClientToken: clientToken,
	}, nil
}

// GetByClientToken resolves a student by their session token. Returns
// ErrNotFound if the token is unknown.
func (r *StudentRepository) GetByClientToken(ctx context.Context, token string) (*domain.Student, error) {
	const q = `SELECT id, room_id, group_id, nickname, client_token, created_at, updated_at
		FROM students WHERE client_token = ?`
	var s domain.Student
	err := r.db.QueryRowContext(ctx, q, token).Scan(
		&s.ID, &s.RoomID, &s.GroupID, &s.Nickname, &s.ClientToken, &s.CreatedAt, &s.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("repository: get student by token: %w", err)
	}
	return &s, nil
}
