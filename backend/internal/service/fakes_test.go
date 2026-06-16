package service

import (
	"context"
	"time"

	"iclassroom/backend/internal/domain"
	"iclassroom/backend/internal/repository"
)

// fakeStore is an in-memory stand-in for the room/group/student repositories.
// It mirrors the real constraints (unique room_code/teacher_token,
// unique room_id+nickname, group capacity) so the service rules can be tested
// without a database. It is not concurrency-safe; tests are single-goroutine.
type fakeStore struct {
	rooms    map[string]*domain.Room    // keyed by roomCode
	groups   map[int64]*domain.Group    // keyed by groupId
	students map[string]*domain.Student // keyed by clientToken

	nextID int64

	// failCreateTimes forces CreateRoomWithGroups to return ErrDuplicate this
	// many times before succeeding, to exercise the roomCode retry path.
	failCreateTimes int
}

func newFakeStore() *fakeStore {
	return &fakeStore{
		rooms:    map[string]*domain.Room{},
		groups:   map[int64]*domain.Group{},
		students: map[string]*domain.Student{},
		nextID:   1,
	}
}

func (f *fakeStore) id() int64 {
	v := f.nextID
	f.nextID++
	return v
}

// --- RoomRepository ---

func (f *fakeStore) CreateRoomWithGroups(_ context.Context, room *domain.Room) ([]domain.Group, error) {
	if f.failCreateTimes > 0 {
		f.failCreateTimes--
		return nil, repository.ErrDuplicate
	}
	if _, exists := f.rooms[room.RoomCode]; exists {
		return nil, repository.ErrDuplicate
	}
	for _, r := range f.rooms {
		if r.TeacherToken == room.TeacherToken {
			return nil, repository.ErrDuplicate
		}
	}
	room.ID = f.id()
	f.rooms[room.RoomCode] = room

	groups := make([]domain.Group, 0, room.GroupCount)
	for i := 1; i <= room.GroupCount; i++ {
		g := domain.Group{ID: f.id(), RoomID: room.ID, GroupName: groupName(i), Capacity: room.GroupCapacity}
		f.groups[g.ID] = &g
		groups = append(groups, g)
	}
	return groups, nil
}

func (f *fakeStore) GetByRoomCode(_ context.Context, code string) (*domain.Room, error) {
	if r, ok := f.rooms[code]; ok {
		return r, nil
	}
	return nil, repository.ErrNotFound
}

func (f *fakeStore) GetByTeacherToken(_ context.Context, token string) (*domain.Room, error) {
	for _, r := range f.rooms {
		if r.TeacherToken == token {
			return r, nil
		}
	}
	return nil, repository.ErrNotFound
}

func (f *fakeStore) EndRoom(_ context.Context, roomID int64, endedAt time.Time) error {
	for _, r := range f.rooms {
		if r.ID == roomID {
			r.Status = domain.RoomStatusEnded
			t := endedAt.UTC()
			r.EndedAt = &t
			return nil
		}
	}
	return repository.ErrNotFound
}

// --- GroupRepository ---

func (f *fakeStore) ListByRoomID(_ context.Context, roomID int64) ([]repository.GroupWithCount, error) {
	var out []repository.GroupWithCount
	for _, g := range f.groups {
		if g.RoomID != roomID {
			continue
		}
		out = append(out, repository.GroupWithCount{Group: *g, CurrentCount: f.countMembers(g.ID)})
	}
	return out, nil
}

func (f *fakeStore) GetByID(_ context.Context, groupID int64) (*domain.Group, error) {
	if g, ok := f.groups[groupID]; ok {
		return g, nil
	}
	return nil, repository.ErrNotFound
}

// --- StudentRepository ---

func (f *fakeStore) Join(_ context.Context, roomID, groupID int64, nickname, clientToken string) (*domain.Student, error) {
	g, ok := f.groups[groupID]
	if !ok || g.RoomID != roomID {
		return nil, repository.ErrNotFound
	}
	if f.countMembers(groupID) >= g.Capacity {
		return nil, repository.ErrGroupFull
	}
	for _, s := range f.students {
		if s.RoomID == roomID && s.Nickname == nickname {
			return nil, repository.ErrDuplicate
		}
	}
	s := &domain.Student{ID: f.id(), RoomID: roomID, GroupID: groupID, Nickname: nickname, ClientToken: clientToken}
	f.students[clientToken] = s
	return s, nil
}

func (f *fakeStore) GetByClientToken(_ context.Context, token string) (*domain.Student, error) {
	if s, ok := f.students[token]; ok {
		return s, nil
	}
	return nil, repository.ErrNotFound
}

func (f *fakeStore) countMembers(groupID int64) int {
	n := 0
	for _, s := range f.students {
		if s.GroupID == groupID {
			n++
		}
	}
	return n
}

func groupName(i int) string {
	// Mirror the production naming ("第N组") without importing fmt at every call
	// site; only used for assertions.
	return "第" + itoa(i) + "组"
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	var b []byte
	for i > 0 {
		b = append([]byte{byte('0' + i%10)}, b...)
		i /= 10
	}
	return string(b)
}
