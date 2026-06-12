// Package apperr defines the typed business error used across the service and
// handler layers. A *Error carries the HTTP status, the machine-readable
// errorCode (see docs/api.md) and a human-readable message, so the service
// layer can express *what went wrong* without importing HTTP/Gin, and the
// handler layer can translate it into the unified response envelope without
// re-deriving status codes.
//
// Anything that is NOT an *apperr.Error is treated by handlers as an
// unexpected internal failure (HTTP 500) and logged.
package apperr

import (
	"fmt"
	"net/http"
)

// Error is a business error with a stable errorCode and an HTTP status.
type Error struct {
	Status  int    // HTTP status code, e.g. 404
	Code    string // machine-readable errorCode, e.g. "ROOM_NOT_FOUND"
	Message string // human-readable message
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// New builds a business error. Prefer the named constructors below; use New
// only for one-off cases.
func New(status int, code, message string) *Error {
	return &Error{Status: status, Code: code, Message: message}
}

// --- Constructors for the codes used by the room / student entry flow. ---
// Each returns a fresh *Error so callers may further customize the message.

func RoomNotFound() *Error {
	return New(http.StatusNotFound, "ROOM_NOT_FOUND", "room not found")
}

func RoomEnded() *Error {
	return New(http.StatusConflict, "ROOM_ENDED", "classroom has ended")
}

func RoomAccessDenied() *Error {
	return New(http.StatusForbidden, "ROOM_ACCESS_DENIED", "token is not authorized for this room")
}

func RoomCreateFailed() *Error {
	return New(http.StatusInternalServerError, "ROOM_CREATE_FAILED", "failed to create room")
}

func InvalidGroupCount() *Error {
	return New(http.StatusBadRequest, "INVALID_GROUP_COUNT", "groupCount is out of range")
}

func InvalidGroupCapacity() *Error {
	return New(http.StatusBadRequest, "INVALID_GROUP_CAPACITY", "groupCapacity is out of range")
}

func InvalidTeacherToken() *Error {
	return New(http.StatusUnauthorized, "INVALID_TEACHER_TOKEN", "missing or invalid teacher token")
}

func InvalidStudentToken() *Error {
	return New(http.StatusUnauthorized, "INVALID_STUDENT_TOKEN", "missing or invalid student token")
}

func InvalidNickname() *Error {
	return New(http.StatusBadRequest, "INVALID_NICKNAME", "nickname is empty or too long")
}

func NicknameDuplicated() *Error {
	return New(http.StatusConflict, "NICKNAME_DUPLICATED", "nickname already exists in this room")
}

func GroupNotFound() *Error {
	return New(http.StatusNotFound, "GROUP_NOT_FOUND", "group not found in this room")
}

func GroupFull() *Error {
	return New(http.StatusConflict, "GROUP_FULL", "group is full")
}

func TaskNotFound() *Error {
	return New(http.StatusNotFound, "TASK_NOT_FOUND", "task not found")
}

func InvalidTaskTitle() *Error {
	return New(http.StatusBadRequest, "INVALID_TASK_TITLE", "task title is required and must be at most 255 characters")
}

func InvalidDeadline() *Error {
	return New(http.StatusBadRequest, "INVALID_DEADLINE", "deadlineAt must be later than current time")
}

func InvalidTargetType() *Error {
	return New(http.StatusBadRequest, "INVALID_TARGET_TYPE", "targetType must be all or groups")
}

func InvalidTargetGroup() *Error {
	return New(http.StatusBadRequest, "INVALID_TARGET_GROUP", "targetGroupIds must belong to this room")
}

// InvalidRequest is used for malformed request bodies / unbindable input.
func InvalidRequest(message string) *Error {
	if message == "" {
		message = "invalid request"
	}
	return New(http.StatusBadRequest, "INVALID_REQUEST", message)
}
