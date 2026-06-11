package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"iclassroom/backend/internal/domain"
)

// GroupRepository reads groups. Group creation lives in RoomRepository because
// groups are only ever created together with their room (same transaction).
type GroupRepository struct {
	db *sql.DB
}

// NewGroupRepository wires the repository to a database handle.
func NewGroupRepository(db *sql.DB) *GroupRepository {
	return &GroupRepository{db: db}
}

// GroupWithCount is a group plus its live member count (derived via COUNT,
// never stored — see domain.Group). It is the read model for the student room
// view and the teacher overview.
type GroupWithCount struct {
	domain.Group
	CurrentCount int `json:"currentCount"`
}

// ListByRoomID returns the room's groups ordered by id, each with its current
// member count. A room with no students still returns every group (count 0).
func (r *GroupRepository) ListByRoomID(ctx context.Context, roomID int64) ([]GroupWithCount, error) {
	const q = `SELECT g.id, g.room_id, g.group_name, g.capacity, g.score_total, COUNT(s.id)
		FROM ` + "`groups`" + ` g
		LEFT JOIN students s ON s.group_id = g.id
		WHERE g.room_id = ?
		GROUP BY g.id, g.room_id, g.group_name, g.capacity, g.score_total
		ORDER BY g.id`
	rows, err := r.db.QueryContext(ctx, q, roomID)
	if err != nil {
		return nil, fmt.Errorf("repository: list groups: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []GroupWithCount
	for rows.Next() {
		var g GroupWithCount
		if err := rows.Scan(&g.ID, &g.RoomID, &g.GroupName, &g.Capacity, &g.ScoreTotal, &g.CurrentCount); err != nil {
			return nil, fmt.Errorf("repository: scan group: %w", err)
		}
		out = append(out, g)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository: iterate groups: %w", err)
	}
	return out, nil
}

// GetByID loads a single group. Returns ErrNotFound if absent. Used to resolve
// groupName for join / resume responses.
func (r *GroupRepository) GetByID(ctx context.Context, groupID int64) (*domain.Group, error) {
	const q = `SELECT id, room_id, group_name, capacity, score_total, created_at, updated_at
		FROM ` + "`groups`" + ` WHERE id = ?`
	var g domain.Group
	err := r.db.QueryRowContext(ctx, q, groupID).Scan(
		&g.ID, &g.RoomID, &g.GroupName, &g.Capacity, &g.ScoreTotal, &g.CreatedAt, &g.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("repository: get group: %w", err)
	}
	return &g, nil
}
