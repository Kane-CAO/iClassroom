// Package service holds the business rules of the iClassroom MVP: validation,
// state transitions, transactions and token minting. It depends on the
// repository layer through the narrow interfaces declared here (the consumer
// owns the interface), which keeps the services unit-testable with fakes and
// free of any HTTP concern.
package service

import (
	"context"

	"iclassroom/backend/internal/domain"
	"iclassroom/backend/internal/repository"
)

// RoomRepository is the persistence surface the services need for rooms.
type RoomRepository interface {
	CreateRoomWithGroups(ctx context.Context, room *domain.Room) ([]domain.Group, error)
	GetByRoomCode(ctx context.Context, code string) (*domain.Room, error)
	GetByTeacherToken(ctx context.Context, token string) (*domain.Room, error)
}

// GroupRepository is the persistence surface the services need for groups.
type GroupRepository interface {
	ListByRoomID(ctx context.Context, roomID int64) ([]repository.GroupWithCount, error)
	GetByID(ctx context.Context, groupID int64) (*domain.Group, error)
}

// StudentRepository is the persistence surface the services need for students.
type StudentRepository interface {
	Join(ctx context.Context, roomID, groupID int64, nickname, clientToken string) (*domain.Student, error)
	GetByClientToken(ctx context.Context, token string) (*domain.Student, error)
}

// TaskRepository is the persistence surface the services need for tasks.
type TaskRepository interface {
	Create(ctx context.Context, task *domain.Task, targetGroupIDs []int64) error
	ListByRoomID(ctx context.Context, roomID int64) ([]repository.TaskWithStats, error)
	ListTargetGroupIDs(ctx context.Context, taskID int64) ([]int64, error)
	GetRoomByTaskID(ctx context.Context, taskID int64) (*domain.Room, error)
	UpdateStatus(ctx context.Context, taskID int64, status domain.TaskStatus) error
}
