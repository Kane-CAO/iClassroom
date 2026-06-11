package domain

import "time"

// TaskStatus is the lifecycle state of a task.
type TaskStatus string

const (
	TaskStatusPublished TaskStatus = "published"
	TaskStatusPaused    TaskStatus = "paused"
	TaskStatusClosed    TaskStatus = "closed"
)

// Valid reports whether s is a known task status.
func (s TaskStatus) Valid() bool {
	switch s {
	case TaskStatusPublished, TaskStatusPaused, TaskStatusClosed:
		return true
	default:
		return false
	}
}

// TargetType controls which students a task is published to.
type TargetType string

const (
	// TargetAll publishes to every student in the room.
	TargetAll TargetType = "all"
	// TargetGroups publishes only to the groups listed in TaskTargetGroup.
	TargetGroups TargetType = "groups"
)

// Valid reports whether t is a known target type.
func (t TargetType) Valid() bool {
	switch t {
	case TargetAll, TargetGroups:
		return true
	default:
		return false
	}
}

// Task is an assignment published by a teacher within a room.
type Task struct {
	ID            int64      `json:"taskId"`
	RoomID        int64      `json:"-"`
	Title         string     `json:"title"`
	Description   string     `json:"description"`
	AttachmentURL string     `json:"attachmentUrl"`
	DeadlineAt    time.Time  `json:"deadlineAt"`
	TargetType    TargetType `json:"targetType"`
	Status        TaskStatus `json:"status"`
	CreatedAt     time.Time  `json:"createdAt"`
	UpdatedAt     time.Time  `json:"updatedAt"`
}

// TaskTargetGroup links a TargetGroups task to one of its target groups.
// A task has one row per targeted group; (task_id, group_id) is unique.
type TaskTargetGroup struct {
	ID      int64 `json:"-"`
	TaskID  int64 `json:"taskId"`
	GroupID int64 `json:"groupId"`
}
