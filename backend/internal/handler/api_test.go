package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"iclassroom/backend/internal/domain"
	"iclassroom/backend/internal/repository"
	"iclassroom/backend/internal/service"
)

func init() { gin.SetMode(gin.TestMode) }

// memStore is an in-memory implementation of the service repository ports,
// good enough to drive the handlers over real HTTP.
type memStore struct {
	rooms    map[string]*domain.Room
	groups   map[int64]*domain.Group
	students map[string]*domain.Student
	next     int64
}

func newMemStore() *memStore {
	return &memStore{
		rooms:    map[string]*domain.Room{},
		groups:   map[int64]*domain.Group{},
		students: map[string]*domain.Student{},
		next:     1,
	}
}

func (m *memStore) id() int64 { v := m.next; m.next++; return v }

func (m *memStore) CreateRoomWithGroups(_ context.Context, room *domain.Room) ([]domain.Group, error) {
	if _, ok := m.rooms[room.RoomCode]; ok {
		return nil, repository.ErrDuplicate
	}
	room.ID = m.id()
	m.rooms[room.RoomCode] = room
	groups := make([]domain.Group, 0, room.GroupCount)
	for i := 1; i <= room.GroupCount; i++ {
		g := domain.Group{ID: m.id(), RoomID: room.ID, GroupName: fmt.Sprintf("第%d组", i), Capacity: room.GroupCapacity}
		m.groups[g.ID] = &g
		groups = append(groups, g)
	}
	return groups, nil
}

func (m *memStore) GetByRoomCode(_ context.Context, code string) (*domain.Room, error) {
	if r, ok := m.rooms[code]; ok {
		return r, nil
	}
	return nil, repository.ErrNotFound
}

func (m *memStore) GetByTeacherToken(_ context.Context, token string) (*domain.Room, error) {
	for _, r := range m.rooms {
		if r.TeacherToken == token {
			return r, nil
		}
	}
	return nil, repository.ErrNotFound
}

func (m *memStore) ListByRoomID(_ context.Context, roomID int64) ([]repository.GroupWithCount, error) {
	var out []repository.GroupWithCount
	for _, g := range m.groups {
		if g.RoomID == roomID {
			out = append(out, repository.GroupWithCount{Group: *g, CurrentCount: m.count(g.ID)})
		}
	}
	return out, nil
}

func (m *memStore) GetByID(_ context.Context, id int64) (*domain.Group, error) {
	if g, ok := m.groups[id]; ok {
		return g, nil
	}
	return nil, repository.ErrNotFound
}

func (m *memStore) Join(_ context.Context, roomID, groupID int64, nickname, token string) (*domain.Student, error) {
	g, ok := m.groups[groupID]
	if !ok || g.RoomID != roomID {
		return nil, repository.ErrNotFound
	}
	if m.count(groupID) >= g.Capacity {
		return nil, repository.ErrGroupFull
	}
	for _, s := range m.students {
		if s.RoomID == roomID && s.Nickname == nickname {
			return nil, repository.ErrDuplicate
		}
	}
	s := &domain.Student{ID: m.id(), RoomID: roomID, GroupID: groupID, Nickname: nickname, ClientToken: token}
	m.students[token] = s
	return s, nil
}

func (m *memStore) GetByClientToken(_ context.Context, token string) (*domain.Student, error) {
	if s, ok := m.students[token]; ok {
		return s, nil
	}
	return nil, repository.ErrNotFound
}

func (m *memStore) count(groupID int64) int {
	n := 0
	for _, s := range m.students {
		if s.GroupID == groupID {
			n++
		}
	}
	return n
}

// testRouter wires the handlers against an in-memory store.
func testRouter(m *memStore) *gin.Engine {
	r := gin.New()
	api := r.Group("/api")
	NewRoomHandler(service.NewRoomService(m, m, "http://localhost:5173")).Register(api)
	NewStudentHandler(service.NewStudentService(m, m, m)).Register(api)
	return r
}

// envelope is the decoded unified response.
type envelope struct {
	Success   bool            `json:"success"`
	Message   string          `json:"message"`
	ErrorCode string          `json:"errorCode"`
	Data      json.RawMessage `json:"data"`
}

func do(t *testing.T, r *gin.Engine, method, path string, body any, headers map[string]string) (int, envelope) {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatalf("encode body: %v", err)
		}
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	var env envelope
	if err := json.Unmarshal(w.Body.Bytes(), &env); err != nil {
		t.Fatalf("decode response (%d): %v\nbody=%s", w.Code, err, w.Body.String())
	}
	return w.Code, env
}

// TestFullEntryFlow walks the first main link end to end over HTTP:
// create room -> student views room -> join -> resume.
func TestFullEntryFlow(t *testing.T) {
	r := testRouter(newMemStore())

	// 1. Teacher creates a room.
	code, env := do(t, r, http.MethodPost, "/api/teacher/rooms", gin.H{
		"title": "Demo Class", "groupCount": 3, "groupCapacity": 2, "allowChooseGroup": true,
	}, nil)
	if code != http.StatusOK || !env.Success {
		t.Fatalf("create room: code=%d env=%+v", code, env)
	}
	var created struct {
		RoomCode     string `json:"roomCode"`
		TeacherToken string `json:"teacherToken"`
		JoinURL      string `json:"joinUrl"`
		Groups       []struct {
			GroupID   int64  `json:"groupId"`
			GroupName string `json:"groupName"`
		} `json:"groups"`
	}
	mustData(t, env, &created)
	if created.RoomCode == "" || created.TeacherToken == "" || len(created.Groups) != 3 {
		t.Fatalf("unexpected create payload: %+v", created)
	}

	// 2. Student fetches the room (public, no token).
	code, env = do(t, r, http.MethodGet, "/api/student/rooms/"+created.RoomCode, nil, nil)
	if code != http.StatusOK || !env.Success {
		t.Fatalf("student get room: code=%d env=%+v", code, env)
	}

	// 3. Student joins.
	gid := created.Groups[0].GroupID
	code, env = do(t, r, http.MethodPost, "/api/student/rooms/"+created.RoomCode+"/join",
		gin.H{"nickname": "Tom", "groupId": gid}, nil)
	if code != http.StatusOK || !env.Success {
		t.Fatalf("join: code=%d env=%+v", code, env)
	}
	var joined struct {
		StudentID   int64  `json:"studentId"`
		ClientToken string `json:"clientToken"`
	}
	mustData(t, env, &joined)
	if joined.ClientToken == "" {
		t.Fatal("join must return clientToken")
	}

	// 4. Resume with the client token restores identity.
	code, env = do(t, r, http.MethodPost, "/api/student/rooms/"+created.RoomCode+"/resume",
		nil, map[string]string{headerStudentToken: joined.ClientToken})
	if code != http.StatusOK || !env.Success {
		t.Fatalf("resume: code=%d env=%+v", code, env)
	}

	// 5. Teacher overview reflects the new student.
	code, env = do(t, r, http.MethodGet, "/api/teacher/rooms/"+created.RoomCode+"/overview",
		nil, map[string]string{headerTeacherToken: created.TeacherToken})
	if code != http.StatusOK || !env.Success {
		t.Fatalf("overview: code=%d env=%+v", code, env)
	}
	var ov struct {
		StudentCount int `json:"studentCount"`
	}
	mustData(t, env, &ov)
	if ov.StudentCount != 1 {
		t.Errorf("overview studentCount = %d, want 1", ov.StudentCount)
	}
}

func TestErrorCases(t *testing.T) {
	r := testRouter(newMemStore())
	// Seed a room.
	_, env := do(t, r, http.MethodPost, "/api/teacher/rooms", gin.H{"title": "X", "groupCount": 1, "groupCapacity": 1}, nil)
	var created struct {
		RoomCode     string `json:"roomCode"`
		TeacherToken string `json:"teacherToken"`
		Groups       []struct {
			GroupID int64 `json:"groupId"`
		} `json:"groups"`
	}
	mustData(t, env, &created)
	gid := created.Groups[0].GroupID

	cases := []struct {
		name       string
		method     string
		path       string
		body       any
		headers    map[string]string
		wantStatus int
		wantCode   string
	}{
		{"room not found", http.MethodGet, "/api/student/rooms/NOPE12", nil, nil, http.StatusNotFound, "ROOM_NOT_FOUND"},
		{"overview missing token", http.MethodGet, "/api/teacher/rooms/" + created.RoomCode + "/overview", nil, nil, http.StatusUnauthorized, "INVALID_TEACHER_TOKEN"},
		{"join bad group", http.MethodPost, "/api/student/rooms/" + created.RoomCode + "/join", gin.H{"nickname": "Z", "groupId": 99999}, nil, http.StatusNotFound, "GROUP_NOT_FOUND"},
		{"resume missing token", http.MethodPost, "/api/student/rooms/" + created.RoomCode + "/resume", nil, nil, http.StatusUnauthorized, "INVALID_STUDENT_TOKEN"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			status, env := do(t, r, c.method, c.path, c.body, c.headers)
			if status != c.wantStatus {
				t.Errorf("status = %d, want %d", status, c.wantStatus)
			}
			if env.Success {
				t.Error("success = true, want false")
			}
			if env.ErrorCode != c.wantCode {
				t.Errorf("errorCode = %q, want %q", env.ErrorCode, c.wantCode)
			}
		})
	}

	// Group full: capacity 1, second join must fail with GROUP_FULL.
	do(t, r, http.MethodPost, "/api/student/rooms/"+created.RoomCode+"/join", gin.H{"nickname": "A", "groupId": gid}, nil)
	status, env2 := do(t, r, http.MethodPost, "/api/student/rooms/"+created.RoomCode+"/join", gin.H{"nickname": "B", "groupId": gid}, nil)
	if status != http.StatusConflict || env2.ErrorCode != "GROUP_FULL" {
		t.Errorf("group full: status=%d code=%q", status, env2.ErrorCode)
	}

	// Duplicate nickname: use a room with spare capacity so the duplicate
	// (not a full group) is what trips the rule.
	_, env4 := do(t, r, http.MethodPost, "/api/teacher/rooms", gin.H{"title": "Y", "groupCount": 1, "groupCapacity": 5}, nil)
	var room2 struct {
		RoomCode string `json:"roomCode"`
		Groups   []struct {
			GroupID int64 `json:"groupId"`
		} `json:"groups"`
	}
	mustData(t, env4, &room2)
	g2 := room2.Groups[0].GroupID
	do(t, r, http.MethodPost, "/api/student/rooms/"+room2.RoomCode+"/join", gin.H{"nickname": "Tom", "groupId": g2}, nil)
	status, env5 := do(t, r, http.MethodPost, "/api/student/rooms/"+room2.RoomCode+"/join", gin.H{"nickname": "Tom", "groupId": g2}, nil)
	if status != http.StatusConflict || env5.ErrorCode != "NICKNAME_DUPLICATED" {
		t.Errorf("dup nickname: status=%d code=%q", status, env5.ErrorCode)
	}
}

func mustData(t *testing.T, env envelope, v any) {
	t.Helper()
	if err := json.Unmarshal(env.Data, v); err != nil {
		t.Fatalf("decode data: %v (data=%s)", err, env.Data)
	}
}
