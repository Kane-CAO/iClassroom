package service

import (
	"context"
	"errors"
	"testing"

	"iclassroom/backend/internal/apperr"
	"iclassroom/backend/internal/domain"
)

const testFrontend = "http://localhost:5173"

func newRoomSvc(f *fakeStore) *RoomService {
	return NewRoomService(f, f, testFrontend)
}

// wantCode asserts err is an *apperr.Error with the given errorCode.
func wantCode(t *testing.T, err error, code string) {
	t.Helper()
	var ae *apperr.Error
	if !errors.As(err, &ae) {
		t.Fatalf("error = %v, want *apperr.Error with code %s", err, code)
	}
	if ae.Code != code {
		t.Fatalf("errorCode = %s, want %s", ae.Code, code)
	}
}

func TestCreateRoom_Defaults(t *testing.T) {
	f := newFakeStore()
	res, err := newRoomSvc(f).CreateRoom(context.Background(), CreateRoomInput{Title: "Demo Class"})
	if err != nil {
		t.Fatalf("CreateRoom error: %v", err)
	}
	if res.Room.GroupCount != defaultGroupCount || res.Room.GroupCapacity != defaultGroupCapacity {
		t.Errorf("defaults not applied: count=%d capacity=%d", res.Room.GroupCount, res.Room.GroupCapacity)
	}
	if len(res.Groups) != defaultGroupCount {
		t.Errorf("created %d groups, want %d", len(res.Groups), defaultGroupCount)
	}
	if res.Room.Status != domain.RoomStatusActive {
		t.Errorf("status = %s, want active", res.Room.Status)
	}
	if res.Room.TeacherToken == "" || res.Room.RoomCode == "" {
		t.Error("roomCode/teacherToken must be populated")
	}
	wantJoin := testFrontend + "/student?room=" + res.Room.RoomCode
	if res.JoinURL != wantJoin {
		t.Errorf("joinURL = %q, want %q", res.JoinURL, wantJoin)
	}
}

func TestCreateRoom_Validation(t *testing.T) {
	cases := []struct {
		name string
		in   CreateRoomInput
		code string
	}{
		{"empty title", CreateRoomInput{Title: "   "}, "INVALID_REQUEST"},
		{"negative groupCount", CreateRoomInput{Title: "x", GroupCount: -1}, "INVALID_GROUP_COUNT"},
		{"huge groupCount", CreateRoomInput{Title: "x", GroupCount: 999}, "INVALID_GROUP_COUNT"},
		{"negative capacity", CreateRoomInput{Title: "x", GroupCapacity: -5}, "INVALID_GROUP_CAPACITY"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := newRoomSvc(newFakeStore()).CreateRoom(context.Background(), c.in)
			wantCode(t, err, c.code)
		})
	}
}

func TestCreateRoom_RetriesOnDuplicateCode(t *testing.T) {
	f := newFakeStore()
	f.failCreateTimes = 2 // first two attempts collide, third succeeds
	res, err := newRoomSvc(f).CreateRoom(context.Background(), CreateRoomInput{Title: "Demo"})
	if err != nil {
		t.Fatalf("CreateRoom should recover from collisions: %v", err)
	}
	if res.Room.RoomCode == "" {
		t.Error("expected a room after retry")
	}
}

func TestCreateRoom_GivesUpAfterMaxRetries(t *testing.T) {
	f := newFakeStore()
	f.failCreateTimes = roomCodeMaxRetries + 1
	_, err := newRoomSvc(f).CreateRoom(context.Background(), CreateRoomInput{Title: "Demo"})
	wantCode(t, err, "ROOM_CREATE_FAILED")
}

func seedRoom(t *testing.T, f *fakeStore) *CreateRoomResult {
	t.Helper()
	res, err := newRoomSvc(f).CreateRoom(context.Background(), CreateRoomInput{Title: "Demo", GroupCount: 3, GroupCapacity: 2})
	if err != nil {
		t.Fatalf("seed CreateRoom: %v", err)
	}
	return res
}

func TestGetRoom_TeacherAuth(t *testing.T) {
	f := newFakeStore()
	created := seedRoom(t, f)
	svc := newRoomSvc(f)
	ctx := context.Background()

	t.Run("valid token", func(t *testing.T) {
		room, err := svc.GetRoom(ctx, created.Room.RoomCode, created.Room.TeacherToken)
		if err != nil {
			t.Fatalf("GetRoom error: %v", err)
		}
		if room.RoomCode != created.Room.RoomCode {
			t.Error("wrong room returned")
		}
	})

	t.Run("missing token", func(t *testing.T) {
		_, err := svc.GetRoom(ctx, created.Room.RoomCode, "")
		wantCode(t, err, "INVALID_TEACHER_TOKEN")
	})

	t.Run("unknown token", func(t *testing.T) {
		_, err := svc.GetRoom(ctx, created.Room.RoomCode, "teacher_bogus")
		wantCode(t, err, "INVALID_TEACHER_TOKEN")
	})

	t.Run("room not found", func(t *testing.T) {
		_, err := svc.GetRoom(ctx, "ZZZZZZ", created.Room.TeacherToken)
		wantCode(t, err, "ROOM_NOT_FOUND")
	})

	t.Run("token of another room", func(t *testing.T) {
		other := seedRoom(t, f)
		_, err := svc.GetRoom(ctx, created.Room.RoomCode, other.Room.TeacherToken)
		wantCode(t, err, "ROOM_ACCESS_DENIED")
	})
}

func TestGetOverview_Counts(t *testing.T) {
	f := newFakeStore()
	created := seedRoom(t, f)
	// Add two students to the first group.
	studentSvc := NewStudentService(f, f, f)
	gid := created.Groups[0].ID
	for _, name := range []string{"Tom", "Jerry"} {
		if _, err := studentSvc.Join(context.Background(), created.Room.RoomCode, name, gid); err != nil {
			t.Fatalf("seed join %s: %v", name, err)
		}
	}

	ov, err := newRoomSvc(f).GetOverview(context.Background(), created.Room.RoomCode, created.Room.TeacherToken)
	if err != nil {
		t.Fatalf("GetOverview error: %v", err)
	}
	if ov.StudentCount != 2 {
		t.Errorf("studentCount = %d, want 2", ov.StudentCount)
	}
	if len(ov.Groups) != 3 {
		t.Errorf("groups = %d, want 3", len(ov.Groups))
	}
}

func TestEndRoom_Success(t *testing.T) {
	f := newFakeStore()
	created := seedRoom(t, f)
	res, err := newRoomSvc(f).EndRoom(context.Background(), created.Room.RoomCode, created.Room.TeacherToken)
	if err != nil {
		t.Fatalf("EndRoom error: %v", err)
	}
	if res.Status != domain.RoomStatusEnded || res.EndedAt == nil {
		t.Fatalf("ended room = %+v", res)
	}
	if got := f.rooms[created.Room.RoomCode].Status; got != domain.RoomStatusEnded {
		t.Fatalf("store status = %s, want ended", got)
	}
}

func TestEndRoom_Errors(t *testing.T) {
	f := newFakeStore()
	created := seedRoom(t, f)
	svc := newRoomSvc(f)

	t.Run("missing token", func(t *testing.T) {
		_, err := svc.EndRoom(context.Background(), created.Room.RoomCode, "")
		wantCode(t, err, "INVALID_TEACHER_TOKEN")
	})

	t.Run("wrong token", func(t *testing.T) {
		other := seedRoom(t, f)
		_, err := svc.EndRoom(context.Background(), created.Room.RoomCode, other.Room.TeacherToken)
		wantCode(t, err, "ROOM_ACCESS_DENIED")
	})

	t.Run("room not found", func(t *testing.T) {
		_, err := svc.EndRoom(context.Background(), "NOPE12", created.Room.TeacherToken)
		wantCode(t, err, "ROOM_NOT_FOUND")
	})

	t.Run("already ended", func(t *testing.T) {
		_, err := svc.EndRoom(context.Background(), created.Room.RoomCode, created.Room.TeacherToken)
		if err != nil {
			t.Fatalf("first end: %v", err)
		}
		_, err = svc.EndRoom(context.Background(), created.Room.RoomCode, created.Room.TeacherToken)
		wantCode(t, err, "ROOM_ALREADY_ENDED")
	})
}
