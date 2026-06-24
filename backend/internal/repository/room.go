package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"iclassroom/backend/internal/domain"
)

// RoomRepository persists rooms and their groups.
type RoomRepository struct {
	db *sql.DB
}

// NewRoomRepository wires the repository to a database handle.
func NewRoomRepository(db *sql.DB) *RoomRepository {
	return &RoomRepository{db: db}
}

// rowScanner is satisfied by both *sql.Row and *sql.Rows.
type rowScanner interface {
	Scan(dest ...any) error
}

const roomColumns = `id, teacher_id, room_code, title, status, group_count, group_capacity,
	allow_choose_group, teacher_token, created_at, updated_at, ended_at`

// scanRoom reads one room row. endedAt is nullable.
func scanRoom(s rowScanner) (*domain.Room, error) {
	var (
		r       domain.Room
		teacherID sql.NullInt64
		endedAt sql.NullTime
	)
	if err := s.Scan(
		&r.ID, &teacherID, &r.RoomCode, &r.Title, &r.Status, &r.GroupCount, &r.GroupCapacity,
		&r.AllowChooseGroup, &r.TeacherToken, &r.CreatedAt, &r.UpdatedAt, &endedAt,
	); err != nil {
		return nil, err
	}
	if teacherID.Valid {
		v := teacherID.Int64
		r.TeacherID = &v
	}
	if endedAt.Valid {
		t := endedAt.Time
		r.EndedAt = &t
	}
	return &r, nil
}

func nullableInt64(v *int64) any {
	if v == nil {
		return nil
	}
	return *v
}

// CreateRoomWithGroups inserts the room and its groups in a single transaction.
// room.GroupCount groups are named "第1组"… and given room.GroupCapacity. On
// success it sets room.ID and returns the created groups (currentCount = 0).
// A UNIQUE violation (room_code / teacher_token) returns ErrDuplicate so the
// service can retry with a freshly generated code/token.
func (r *RoomRepository) CreateRoomWithGroups(ctx context.Context, room *domain.Room) ([]domain.Group, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("repository: begin tx: %w", err)
	}
	// Roll back unless we reach an explicit Commit. Rollback after Commit is a
	// no-op, so this is safe.
	defer func() { _ = tx.Rollback() }()

	const insertRoom = `INSERT INTO rooms
		(teacher_id, room_code, title, status, group_count, group_capacity, allow_choose_group, teacher_token)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	res, err := tx.ExecContext(ctx, insertRoom,
		nullableInt64(room.TeacherID), room.RoomCode, room.Title, room.Status, room.GroupCount, room.GroupCapacity,
		room.AllowChooseGroup, room.TeacherToken,
	)
	if err != nil {
		if isDuplicateKey(err) {
			return nil, ErrDuplicate
		}
		return nil, fmt.Errorf("repository: insert room: %w", err)
	}
	roomID, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("repository: room last insert id: %w", err)
	}
	room.ID = roomID

	const insertGroup = `INSERT INTO ` + "`groups`" + ` (room_id, group_name, capacity) VALUES (?, ?, ?)`
	groups := make([]domain.Group, 0, room.GroupCount)
	for i := 1; i <= room.GroupCount; i++ {
		name := fmt.Sprintf("第%d组", i)
		gres, err := tx.ExecContext(ctx, insertGroup, roomID, name, room.GroupCapacity)
		if err != nil {
			return nil, fmt.Errorf("repository: insert group %d: %w", i, err)
		}
		gid, err := gres.LastInsertId()
		if err != nil {
			return nil, fmt.Errorf("repository: group last insert id: %w", err)
		}
		groups = append(groups, domain.Group{
			ID:        gid,
			RoomID:    roomID,
			GroupName: name,
			Capacity:  room.GroupCapacity,
		})
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("repository: commit: %w", err)
	}
	return groups, nil
}

// GetByRoomCode loads a room by its code. Returns ErrNotFound if absent.
func (r *RoomRepository) GetByRoomCode(ctx context.Context, code string) (*domain.Room, error) {
	const q = `SELECT ` + roomColumns + ` FROM rooms WHERE room_code = ?`
	room, err := scanRoom(r.db.QueryRowContext(ctx, q, code))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("repository: get room by code: %w", err)
	}
	return room, nil
}

// GetByTeacherToken loads a room by its teacher token. Returns ErrNotFound if
// no room owns that token. Used to distinguish an unknown token from a token
// that belongs to a *different* room (ROOM_ACCESS_DENIED).
func (r *RoomRepository) GetByTeacherToken(ctx context.Context, token string) (*domain.Room, error) {
	const q = `SELECT ` + roomColumns + ` FROM rooms WHERE teacher_token = ?`
	room, err := scanRoom(r.db.QueryRowContext(ctx, q, token))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("repository: get room by teacher token: %w", err)
	}
	return room, nil
}

func (r *RoomRepository) ListByTeacherID(ctx context.Context, teacherID int64) ([]domain.Room, error) {
	const q = `SELECT ` + roomColumns + ` FROM rooms WHERE teacher_id = ? ORDER BY created_at DESC, id DESC`
	rows, err := r.db.QueryContext(ctx, q, teacherID)
	if err != nil {
		return nil, fmt.Errorf("repository: list rooms by teacher: %w", err)
	}
	defer func() { _ = rows.Close() }()

	out := make([]domain.Room, 0)
	for rows.Next() {
		room, err := scanRoom(rows)
		if err != nil {
			return nil, fmt.Errorf("repository: scan teacher room: %w", err)
		}
		out = append(out, *room)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository: iterate teacher rooms: %w", err)
	}
	return out, nil
}

// EndRoom marks a room as ended and records the end timestamp.
func (r *RoomRepository) EndRoom(ctx context.Context, roomID int64, endedAt time.Time) error {
	const q = `UPDATE rooms SET status = ?, ended_at = ?, updated_at = ? WHERE id = ?`
	res, err := r.db.ExecContext(ctx, q, domain.RoomStatusEnded, endedAt.UTC(), endedAt.UTC(), roomID)
	if err != nil {
		return fmt.Errorf("repository: end room: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("repository: end room rows affected: %w", err)
	}
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}
