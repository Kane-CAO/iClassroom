package service

import (
	"context"
	"errors"
	"testing"

	"iclassroom/backend/internal/domain"
	"iclassroom/backend/internal/websocket"
)

// recordingBroadcaster captures every broadcast so tests can assert what was
// (and was not) published. When err is non-nil, Broadcast reports it, letting
// us verify a broadcast failure never affects the business result.
type recordingBroadcaster struct {
	events []websocket.Event
	err    error
}

func (r *recordingBroadcaster) Broadcast(_ string, evt websocket.Event) error {
	r.events = append(r.events, evt)
	return r.err
}

func (r *recordingBroadcaster) only(t *testing.T) websocket.Event {
	t.Helper()
	if len(r.events) != 1 {
		t.Fatalf("expected exactly 1 broadcast, got %d: %+v", len(r.events), r.events)
	}
	return r.events[0]
}

// Operation failure must not broadcast: a rejected join (unknown group) leaves
// the pool untouched, so no student_joined event may be emitted.
func TestBroadcast_NotSentWhenOperationFails(t *testing.T) {
	f := newFakeStore()
	created := seedRoom(t, f)
	rec := &recordingBroadcaster{}
	svc := NewStudentService(f, f, f, rec)

	_, err := svc.Join(context.Background(), created.Room.RoomCode, "Tom", 999999)
	if err == nil {
		t.Fatal("expected join to fail for an unknown group")
	}
	if len(rec.events) != 0 {
		t.Fatalf("no event should be broadcast on failure, got %+v", rec.events)
	}
}

// Operation success must broadcast the right event with the minimal payload.
func TestBroadcast_SentOnOperationSuccess(t *testing.T) {
	f := newFakeStore()
	created := seedRoom(t, f)
	rec := &recordingBroadcaster{}
	svc := NewStudentService(f, f, f, rec)
	gid := created.Groups[0].ID

	res, err := svc.Join(context.Background(), created.Room.RoomCode, "Tom", gid)
	if err != nil {
		t.Fatalf("join: %v", err)
	}

	evt := rec.only(t)
	if evt.Type != websocket.EventStudentJoined {
		t.Errorf("event type = %q, want %q", evt.Type, websocket.EventStudentJoined)
	}
	if evt.RoomCode != created.Room.RoomCode {
		t.Errorf("roomCode = %q, want %q", evt.RoomCode, created.Room.RoomCode)
	}
	data, ok := evt.Data.(map[string]any)
	if !ok {
		t.Fatalf("data is %T, want map[string]any", evt.Data)
	}
	if data["studentId"] != res.Student.ID {
		t.Errorf("data.studentId = %v, want %v", data["studentId"], res.Student.ID)
	}
	if data["nickname"] != "Tom" {
		t.Errorf("data.nickname = %v, want Tom", data["nickname"])
	}
	if data["groupId"] != gid {
		t.Errorf("data.groupId = %v, want %v", data["groupId"], gid)
	}
}

// A broadcast that returns an error must be swallowed: the business operation
// still succeeds and returns its normal result.
func TestBroadcast_ErrorDoesNotFailBusiness(t *testing.T) {
	f := newFakeStore()
	created := seedRoom(t, f)
	rec := &recordingBroadcaster{err: errors.New("hub unavailable")}
	svc := NewStudentService(f, f, f, rec)
	gid := created.Groups[0].ID

	res, err := svc.Join(context.Background(), created.Room.RoomCode, "Tom", gid)
	if err != nil {
		t.Fatalf("business must succeed despite broadcast error, got: %v", err)
	}
	if res == nil || res.Student == nil {
		t.Fatal("expected a join result")
	}
	if len(rec.events) != 1 {
		t.Fatalf("broadcast should still have been attempted once, got %d", len(rec.events))
	}
}

// EndRoom success emits room_ended carrying the new status.
func TestBroadcast_EndRoomEmitsRoomEnded(t *testing.T) {
	f := newFakeStore()
	created := seedRoom(t, f)
	rec := &recordingBroadcaster{}
	svc := NewRoomService(f, f, testFrontend, rec)

	_, err := svc.EndRoom(context.Background(), created.Room.RoomCode, created.Room.TeacherToken)
	if err != nil {
		t.Fatalf("end room: %v", err)
	}

	evt := rec.only(t)
	if evt.Type != websocket.EventRoomEnded {
		t.Errorf("event type = %q, want %q", evt.Type, websocket.EventRoomEnded)
	}
	if evt.RoomCode != created.Room.RoomCode {
		t.Errorf("roomCode = %q, want %q", evt.RoomCode, created.Room.RoomCode)
	}
	data, ok := evt.Data.(map[string]any)
	if !ok {
		t.Fatalf("data is %T, want map[string]any", evt.Data)
	}
	if data["status"] != domain.RoomStatusEnded {
		t.Errorf("data.status = %v, want %v", data["status"], domain.RoomStatusEnded)
	}
}

// A nil broadcaster must default to noop and never panic.
func TestBroadcast_NilBroadcasterDefaultsToNoop(t *testing.T) {
	f := newFakeStore()
	created := seedRoom(t, f)
	svc := NewStudentService(f, f, f, nil)

	if _, err := svc.Join(context.Background(), created.Room.RoomCode, "Tom", created.Groups[0].ID); err != nil {
		t.Fatalf("join with noop broadcaster: %v", err)
	}
}
