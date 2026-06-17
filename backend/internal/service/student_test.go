package service

import (
	"context"
	"testing"

	"iclassroom/backend/internal/domain"
)

func newStudentSvc(f *fakeStore) *StudentService {
	return NewStudentService(f, f, f, nil)
}

func TestJoin_Success(t *testing.T) {
	f := newFakeStore()
	created := seedRoom(t, f) // 3 groups, capacity 2
	gid := created.Groups[0].ID

	res, err := newStudentSvc(f).Join(context.Background(), created.Room.RoomCode, "Tom", gid)
	if err != nil {
		t.Fatalf("Join error: %v", err)
	}
	if res.Student.ID == 0 || res.Student.ClientToken == "" {
		t.Error("studentId / clientToken must be populated")
	}
	if res.GroupName == "" {
		t.Error("groupName must be populated")
	}
	if res.RoomCode != created.Room.RoomCode {
		t.Errorf("roomCode = %q, want %q", res.RoomCode, created.Room.RoomCode)
	}
}

func TestJoin_DuplicateNickname(t *testing.T) {
	f := newFakeStore()
	created := seedRoom(t, f)
	gid := created.Groups[0].ID
	ctx := context.Background()

	if _, err := newStudentSvc(f).Join(ctx, created.Room.RoomCode, "Tom", gid); err != nil {
		t.Fatalf("first join: %v", err)
	}
	_, err := newStudentSvc(f).Join(ctx, created.Room.RoomCode, "Tom", gid)
	wantCode(t, err, "NICKNAME_DUPLICATED")
}

func TestJoin_GroupFull(t *testing.T) {
	f := newFakeStore()
	created := seedRoom(t, f) // capacity 2
	gid := created.Groups[0].ID
	ctx := context.Background()

	for _, name := range []string{"A", "B"} {
		if _, err := newStudentSvc(f).Join(ctx, created.Room.RoomCode, name, gid); err != nil {
			t.Fatalf("join %s: %v", name, err)
		}
	}
	_, err := newStudentSvc(f).Join(ctx, created.Room.RoomCode, "C", gid)
	wantCode(t, err, "GROUP_FULL")
}

func TestJoin_GroupNotInRoom(t *testing.T) {
	f := newFakeStore()
	created := seedRoom(t, f)
	other := seedRoom(t, f)
	ctx := context.Background()

	// Group belonging to a different room must be rejected.
	_, err := newStudentSvc(f).Join(ctx, created.Room.RoomCode, "Tom", other.Groups[0].ID)
	wantCode(t, err, "GROUP_NOT_FOUND")
}

func TestJoin_InvalidNickname(t *testing.T) {
	f := newFakeStore()
	created := seedRoom(t, f)
	gid := created.Groups[0].ID
	_, err := newStudentSvc(f).Join(context.Background(), created.Room.RoomCode, "   ", gid)
	wantCode(t, err, "INVALID_NICKNAME")
}

func TestJoin_RoomNotFound(t *testing.T) {
	f := newFakeStore()
	_, err := newStudentSvc(f).Join(context.Background(), "NOPE12", "Tom", 1)
	wantCode(t, err, "ROOM_NOT_FOUND")
}

func TestJoin_RoomEnded(t *testing.T) {
	f := newFakeStore()
	created := seedRoom(t, f)
	f.rooms[created.Room.RoomCode].Status = domain.RoomStatusEnded
	_, err := newStudentSvc(f).Join(context.Background(), created.Room.RoomCode, "Tom", created.Groups[0].ID)
	wantCode(t, err, "ROOM_ENDED")
}

func TestResume_Success(t *testing.T) {
	f := newFakeStore()
	created := seedRoom(t, f)
	gid := created.Groups[0].ID
	ctx := context.Background()

	joined, err := newStudentSvc(f).Join(ctx, created.Room.RoomCode, "Tom", gid)
	if err != nil {
		t.Fatalf("join: %v", err)
	}

	res, err := newStudentSvc(f).Resume(ctx, created.Room.RoomCode, joined.Student.ClientToken)
	if err != nil {
		t.Fatalf("Resume error: %v", err)
	}
	if res.Student.Nickname != "Tom" || res.RoomStatus != domain.RoomStatusActive {
		t.Errorf("unexpected resume result: %+v", res)
	}
}

func TestResume_Errors(t *testing.T) {
	f := newFakeStore()
	created := seedRoom(t, f)
	gid := created.Groups[0].ID
	ctx := context.Background()
	joined, _ := newStudentSvc(f).Join(ctx, created.Room.RoomCode, "Tom", gid)

	t.Run("missing token", func(t *testing.T) {
		_, err := newStudentSvc(f).Resume(ctx, created.Room.RoomCode, "")
		wantCode(t, err, "INVALID_STUDENT_TOKEN")
	})

	t.Run("unknown token", func(t *testing.T) {
		_, err := newStudentSvc(f).Resume(ctx, created.Room.RoomCode, "student_bogus")
		wantCode(t, err, "INVALID_STUDENT_TOKEN")
	})

	t.Run("room not found", func(t *testing.T) {
		_, err := newStudentSvc(f).Resume(ctx, "NOPE12", joined.Student.ClientToken)
		wantCode(t, err, "ROOM_NOT_FOUND")
	})

	t.Run("token from another room", func(t *testing.T) {
		other := seedRoom(t, f)
		_, err := newStudentSvc(f).Resume(ctx, other.Room.RoomCode, joined.Student.ClientToken)
		wantCode(t, err, "INVALID_STUDENT_TOKEN")
	})

	t.Run("room ended", func(t *testing.T) {
		f.rooms[created.Room.RoomCode].Status = domain.RoomStatusEnded
		_, err := newStudentSvc(f).Resume(ctx, created.Room.RoomCode, joined.Student.ClientToken)
		wantCode(t, err, "ROOM_ENDED")
		f.rooms[created.Room.RoomCode].Status = domain.RoomStatusActive // restore
	})
}
