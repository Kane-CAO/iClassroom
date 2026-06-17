package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	gws "github.com/gorilla/websocket"

	"iclassroom/backend/internal/domain"
	"iclassroom/backend/internal/service"
	"iclassroom/backend/internal/websocket"
)

// wsTestServer spins up an httptest server exposing GET /ws backed by an
// in-memory store, and returns the dial URL plus the hub manager so tests can
// assert pool state.
func wsTestServer(t *testing.T, m *memStore) (string, *websocket.HubManager, func()) {
	t.Helper()
	manager := websocket.NewHubManager()
	r := gin.New()
	NewWSHandler(service.NewWSAuthService(m, m), manager).Register(r)

	srv := httptest.NewServer(r)
	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http")
	return wsURL, manager, srv.Close
}

// seedRoomAndStudent populates the store with one room and one student in it.
func seedRoomAndStudent(m *memStore) {
	room := &domain.Room{ID: 1, RoomCode: "ABC123", Title: "Demo", Status: domain.RoomStatusActive}
	m.rooms[room.RoomCode] = room
	m.students["stok"] = &domain.Student{ID: 10, RoomID: room.ID, GroupID: 1, Nickname: "Tom", ClientToken: "stok"}

	// A second room with its own student, to test cross-room rejection.
	other := &domain.Room{ID: 2, RoomCode: "XYZ789", Title: "Other", Status: domain.RoomStatusActive}
	m.rooms[other.RoomCode] = other
	m.students["otherTok"] = &domain.Student{ID: 20, RoomID: other.ID, GroupID: 2, Nickname: "Sam", ClientToken: "otherTok"}
}

// dial opens a WebSocket against the given query string. On a non-101 handshake
// gorilla returns ErrBadHandshake plus the *http.Response, whose body carries
// the unified error envelope.
func dial(t *testing.T, wsURL, query string) (*gws.Conn, *http.Response, error) {
	t.Helper()
	conn, resp, err := gws.DefaultDialer.Dial(wsURL+"/ws?"+query, nil)
	return conn, resp, err
}

func decodeEnvelope(t *testing.T, resp *http.Response) (int, envelope) {
	t.Helper()
	defer func() { _ = resp.Body.Close() }()
	body, _ := io.ReadAll(resp.Body)
	var env envelope
	if err := json.Unmarshal(body, &env); err != nil {
		t.Fatalf("decode envelope (status %d): %v\nbody=%s", resp.StatusCode, err, body)
	}
	return resp.StatusCode, env
}

func TestWS_TeacherAndDisplayConnect(t *testing.T) {
	m := newMemStore()
	seedRoomAndStudent(m)
	wsURL, _, closeSrv := wsTestServer(t, m)
	defer closeSrv()

	for _, role := range []string{"teacher", "display"} {
		conn, resp, err := dial(t, wsURL, "room=ABC123&role="+role)
		if err != nil {
			t.Fatalf("%s dial failed: %v (resp=%v)", role, err, resp)
		}
		if resp.StatusCode != http.StatusSwitchingProtocols {
			t.Fatalf("%s handshake status = %d, want 101", role, resp.StatusCode)
		}
		_ = conn.Close()
	}
}

func TestWS_StudentValidToken(t *testing.T) {
	m := newMemStore()
	seedRoomAndStudent(m)
	wsURL, _, closeSrv := wsTestServer(t, m)
	defer closeSrv()

	conn, resp, err := dial(t, wsURL, "room=ABC123&role=student&token=stok")
	if err != nil {
		t.Fatalf("student dial failed: %v (resp=%v)", err, resp)
	}
	if resp.StatusCode != http.StatusSwitchingProtocols {
		t.Fatalf("handshake status = %d, want 101", resp.StatusCode)
	}
	_ = conn.Close()
}

func TestWS_AuthFailures(t *testing.T) {
	m := newMemStore()
	seedRoomAndStudent(m)
	wsURL, _, closeSrv := wsTestServer(t, m)
	defer closeSrv()

	cases := []struct {
		name      string
		query     string
		wantCode  int
		wantError string
	}{
		{"invalid role", "room=ABC123&role=admin", http.StatusBadRequest, "INVALID_ROLE"},
		{"missing role", "room=ABC123", http.StatusBadRequest, "INVALID_ROLE"},
		{"room not found", "room=NOPE&role=teacher", http.StatusNotFound, "ROOM_NOT_FOUND"},
		{"student no token", "room=ABC123&role=student", http.StatusUnauthorized, "INVALID_STUDENT_TOKEN"},
		{"student unknown token", "room=ABC123&role=student&token=ghost", http.StatusUnauthorized, "INVALID_STUDENT_TOKEN"},
		{"student cross-room token", "room=ABC123&role=student&token=otherTok", http.StatusForbidden, "ROOM_ACCESS_DENIED"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			conn, resp, err := dial(t, wsURL, tc.query)
			if err == nil {
				_ = conn.Close()
				t.Fatalf("expected handshake to fail, but it succeeded")
			}
			code, env := decodeEnvelope(t, resp)
			if code != tc.wantCode {
				t.Errorf("status = %d, want %d", code, tc.wantCode)
			}
			if env.Success {
				t.Errorf("success = true, want false")
			}
			if env.ErrorCode != tc.wantError {
				t.Errorf("errorCode = %q, want %q", env.ErrorCode, tc.wantError)
			}
		})
	}
}

func TestWS_DisconnectCleansUpPool(t *testing.T) {
	m := newMemStore()
	seedRoomAndStudent(m)
	wsURL, manager, closeSrv := wsTestServer(t, m)
	defer closeSrv()

	conn, resp, err := dial(t, wsURL, "room=ABC123&role=teacher")
	if err != nil {
		t.Fatalf("dial failed: %v (resp=%v)", err, resp)
	}

	// Connection should be registered in the room's pool.
	waitFor(t, func() bool { return manager.ClientCount("ABC123") == 1 }, "client to register")

	// Client disconnects -> read pump errors -> client removed from the pool.
	_ = conn.Close()
	waitFor(t, func() bool { return manager.RoomCount() == 0 }, "pool to be cleaned up after disconnect")
}

func waitFor(t *testing.T, cond func() bool, what string) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %s", what)
}
