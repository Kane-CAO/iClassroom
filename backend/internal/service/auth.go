package service

import (
	"context"
	"errors"

	"iclassroom/backend/internal/apperr"
	"iclassroom/backend/internal/domain"
	"iclassroom/backend/internal/repository"
)

func verifyTeacherForRoom(ctx context.Context, rooms RoomRepository, room *domain.Room, teacherToken string, teacherIDs ...int64) error {
	if len(teacherIDs) > 0 && teacherIDs[0] > 0 {
		if room.TeacherID != nil && *room.TeacherID == teacherIDs[0] {
			return nil
		}
		if teacherToken == "" {
			return apperr.RoomAccessDenied()
		}
	}
	if teacherToken == "" {
		return apperr.InvalidTeacherToken()
	}
	if teacherToken == room.TeacherToken {
		return nil
	}

	if _, err := rooms.GetByTeacherToken(ctx, teacherToken); err == nil {
		return apperr.RoomAccessDenied()
	} else if !errors.Is(err, repository.ErrNotFound) {
		return err
	}

	return apperr.InvalidTeacherToken()
}

func verifyTeacherByRoomCode(ctx context.Context, rooms RoomRepository, roomCode, teacherToken string, teacherIDs ...int64) (*domain.Room, error) {
	room, err := rooms.GetByRoomCode(ctx, roomCode)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, apperr.RoomNotFound()
	}
	if err != nil {
		return nil, err
	}
	if err := verifyTeacherForRoom(ctx, rooms, room, teacherToken, teacherIDs...); err != nil {
		return nil, err
	}
	return room, nil
}
