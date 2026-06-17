package websocket

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"testing"
	"time"
)

// fakeConn is an in-memory wsConn used to drive the pumps without a real
// network handshake. It records written frames and unblocks ReadMessage when
// closed, simulating a peer disconnect.
type fakeConn struct {
	mu      sync.Mutex
	written [][]byte
	closed  bool
	closeCh chan struct{}
}

func newFakeConn() *fakeConn {
	return &fakeConn{closeCh: make(chan struct{})}
}

var errConnClosed = errors.New("fake conn closed")

func (f *fakeConn) WriteMessage(_ int, data []byte) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.closed {
		return errConnClosed
	}
	cp := make([]byte, len(data))
	copy(cp, data)
	f.written = append(f.written, cp)
	return nil
}

func (f *fakeConn) ReadMessage() (int, []byte, error) {
	<-f.closeCh
	return 0, nil, errConnClosed
}

func (f *fakeConn) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.closed {
		return nil
	}
	f.closed = true
	close(f.closeCh)
	return nil
}

func (f *fakeConn) SetWriteDeadline(time.Time) error      { return nil }
func (f *fakeConn) SetReadDeadline(time.Time) error       { return nil }
func (f *fakeConn) SetReadLimit(int64)                    {}
func (f *fakeConn) SetPongHandler(func(string) error)     {}
func (f *fakeConn) WriteControl(int, []byte, time.Time) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.closed {
		return errConnClosed
	}
	return nil
}

func (f *fakeConn) frames() [][]byte {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([][]byte, len(f.written))
	copy(out, f.written)
	return out
}

// register attaches a client (with started pumps) to the manager and returns
// both so the test can inspect delivery and trigger disconnects.
func register(m *HubManager, role Role, roomCode, id string) (*Client, *fakeConn) {
	conn := newFakeConn()
	c := newClient(m, conn, role, roomCode, id)
	m.Register(c)
	c.start()
	return c, conn
}

func TestNewHubManagerEmpty(t *testing.T) {
	m := NewHubManager()
	if got := m.RoomCount(); got != 0 {
		t.Fatalf("RoomCount() = %d, want 0", got)
	}
	if got := m.ClientCount("ABC123"); got != 0 {
		t.Fatalf("ClientCount of unknown room = %d, want 0", got)
	}
}

func TestPoolCreatedOnRegister(t *testing.T) {
	m := NewHubManager()
	c := newClient(m, newFakeConn(), RoleTeacher, "ABC123", "t1")
	m.Register(c)

	if got := m.RoomCount(); got != 1 {
		t.Fatalf("RoomCount() = %d, want 1", got)
	}
	if got := m.ClientCount("ABC123"); got != 1 {
		t.Fatalf("ClientCount() = %d, want 1", got)
	}
}

func TestAllRolesShareRoom(t *testing.T) {
	m := NewHubManager()
	for _, role := range []Role{RoleTeacher, RoleStudent, RoleDisplay} {
		c := newClient(m, newFakeConn(), role, "ABC123", string(role))
		m.Register(c)
	}

	if got := m.RoomCount(); got != 1 {
		t.Fatalf("RoomCount() = %d, want 1 (single shared room)", got)
	}
	if got := m.ClientCount("ABC123"); got != 3 {
		t.Fatalf("ClientCount() = %d, want 3", got)
	}
}

func TestUnregisterRemovesEmptyHub(t *testing.T) {
	m := NewHubManager()
	a := newClient(m, newFakeConn(), RoleStudent, "ABC123", "a")
	b := newClient(m, newFakeConn(), RoleStudent, "ABC123", "b")
	m.Register(a)
	m.Register(b)

	m.unregister(a)
	if got := m.ClientCount("ABC123"); got != 1 {
		t.Fatalf("after one unregister ClientCount() = %d, want 1", got)
	}
	if got := m.RoomCount(); got != 1 {
		t.Fatalf("hub should still exist with one client, RoomCount() = %d, want 1", got)
	}

	m.unregister(b)
	if got := m.RoomCount(); got != 0 {
		t.Fatalf("empty hub should be removed, RoomCount() = %d, want 0", got)
	}

	// Idempotent: unregistering again must not panic or go negative.
	m.unregister(b)
	if got := m.RoomCount(); got != 0 {
		t.Fatalf("RoomCount() after double unregister = %d, want 0", got)
	}
}

func TestBroadcastUnknownRoomIsNoop(t *testing.T) {
	m := NewHubManager()
	// Must not panic when no hub exists.
	m.Broadcast("NOPE", NewEvent(EventTaskPublished, "NOPE", nil))
}

func TestBroadcastDeliversEventJSON(t *testing.T) {
	m := NewHubManager()
	c, conn := register(m, RoleStudent, "ABC123", "s1")
	defer c.close()

	evt := NewEvent(EventTaskPublished, "ABC123", map[string]any{"taskId": 7})
	m.Broadcast("ABC123", evt)

	frame := waitForFrame(t, conn)
	var got Event
	if err := json.Unmarshal(frame, &got); err != nil {
		t.Fatalf("unmarshal broadcast frame: %v", err)
	}
	if got.Type != EventTaskPublished {
		t.Errorf("event type = %q, want %q", got.Type, EventTaskPublished)
	}
	if got.RoomCode != "ABC123" {
		t.Errorf("roomCode = %q, want ABC123", got.RoomCode)
	}
	if got.OccurredAt.IsZero() {
		t.Error("occurredAt should be set")
	}
}

func TestClientCleanedUpOnDisconnect(t *testing.T) {
	m := NewHubManager()
	_, conn := register(m, RoleStudent, "ABC123", "s1")

	if got := m.ClientCount("ABC123"); got != 1 {
		t.Fatalf("ClientCount() = %d, want 1", got)
	}

	// Simulate the peer dropping the connection: ReadMessage unblocks with an
	// error, the read pump exits and cleans the client out of the pool.
	_ = conn.Close()

	waitFor(t, func() bool { return m.RoomCount() == 0 }, "client to be removed after disconnect")
}

// TestConcurrentBroadcastSafe exercises the mutexes: many clients register and
// unregister while broadcasts fan out concurrently. Run with -race to catch any
// unsynchronised map access; the assertion is simply that nothing panics and
// the manager converges to empty.
func TestConcurrentBroadcastSafe(t *testing.T) {
	m := NewHubManager()
	const rooms = 8
	const clientsPerRoom = 16

	var wg sync.WaitGroup

	clients := make([][]*Client, rooms)
	for r := 0; r < rooms; r++ {
		room := "ROOM" + strconv.Itoa(r)
		clients[r] = make([]*Client, clientsPerRoom)
		for i := 0; i < clientsPerRoom; i++ {
			c := newClient(m, newFakeConn(), RoleStudent, room, fmt.Sprintf("%s-%d", room, i))
			m.Register(c)
			clients[r][i] = c
		}
	}

	// Concurrent broadcasters.
	for r := 0; r < rooms; r++ {
		room := "ROOM" + strconv.Itoa(r)
		wg.Add(1)
		go func() {
			defer wg.Done()
			for n := 0; n < 200; n++ {
				m.Broadcast(room, NewEvent(EventScoreUpdated, room, n))
			}
		}()
	}

	// Concurrent churn: unregister and re-register clients while broadcasting.
	for r := 0; r < rooms; r++ {
		r := r
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < clientsPerRoom; i++ {
				m.unregister(clients[r][i])
				m.Register(clients[r][i])
			}
		}()
	}

	wg.Wait()

	// Tear everything down and confirm the manager drains to zero rooms.
	for r := 0; r < rooms; r++ {
		for i := 0; i < clientsPerRoom; i++ {
			m.unregister(clients[r][i])
		}
	}
	if got := m.RoomCount(); got != 0 {
		t.Fatalf("after draining all clients RoomCount() = %d, want 0", got)
	}
}

// TestSlowConsumerDoesNotBlockBroadcast fills a client's send buffer past
// capacity without a write pump draining it. enqueue must drop the slow client
// instead of blocking, so the broadcast loop always returns promptly.
func TestSlowConsumerDoesNotBlockBroadcast(t *testing.T) {
	m := NewHubManager()
	// Client without started pumps: nothing drains its send channel.
	c := newClient(m, newFakeConn(), RoleDisplay, "ABC123", "slow")
	m.Register(c)

	done := make(chan struct{})
	go func() {
		for n := 0; n < sendBuffer*4; n++ {
			m.Broadcast("ABC123", NewEvent(EventRankingUpdated, "ABC123", n))
		}
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("broadcast blocked on a slow consumer")
	}

	// The slow consumer should have been closed once its buffer overflowed.
	select {
	case <-c.done:
	default:
		t.Error("slow consumer was not closed after buffer overflow")
	}
}

func TestRoleValid(t *testing.T) {
	for _, r := range []Role{RoleTeacher, RoleStudent, RoleDisplay} {
		if !r.Valid() {
			t.Errorf("%q should be valid", r)
		}
	}
	if Role("admin").Valid() {
		t.Error("unknown role should be invalid")
	}
}

// --- helpers ---

func waitForFrame(t *testing.T, conn *fakeConn) []byte {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if frames := conn.frames(); len(frames) > 0 {
			return frames[0]
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal("timed out waiting for a broadcast frame")
	return nil
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
