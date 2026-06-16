package websocket

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Connection timing/sizing constants for the read and write pumps.
const (
	// writeWait is the time allowed to write a single message to the peer.
	writeWait = 10 * time.Second
	// pongWait is how long we wait for a pong before considering the peer dead.
	pongWait = 60 * time.Second
	// pingPeriod must be smaller than pongWait so a ping is always in flight.
	pingPeriod = (pongWait * 9) / 10
	// maxMessageSize caps inbound messages; clients are receivers, not senders.
	maxMessageSize = 4 * 1024
	// sendBuffer is the per-client outbound queue depth. When it overflows the
	// client is treated as a slow consumer and dropped so one stuck connection
	// can never block a broadcast to the whole room.
	sendBuffer = 64
)

// Role identifies the kind of participant on a connection. All three roles may
// subscribe to the same room.
type Role string

const (
	RoleTeacher Role = "teacher"
	RoleStudent Role = "student"
	RoleDisplay Role = "display"
)

// Valid reports whether the role is one of the recognised participant kinds.
func (r Role) Valid() bool {
	switch r {
	case RoleTeacher, RoleStudent, RoleDisplay:
		return true
	default:
		return false
	}
}

// wsConn is the minimal subset of *websocket.Conn the client needs. Abstracting
// it keeps the pumps unit-testable without a real network handshake.
type wsConn interface {
	WriteMessage(messageType int, data []byte) error
	ReadMessage() (messageType int, p []byte, err error)
	Close() error
	SetWriteDeadline(t time.Time) error
	SetReadDeadline(t time.Time) error
	SetReadLimit(limit int64)
	SetPongHandler(h func(appData string) error)
	WriteControl(messageType int, data []byte, deadline time.Time) error
}

// DefaultUpgrader upgrades HTTP requests to WebSocket connections. CheckOrigin
// is permissive here for local development; production wiring should replace it
// with an origin allow-list. This package only provides the upgrader — the HTTP
// handler that uses it lives outside the core package.
var DefaultUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

// Client represents one WebSocket connection bound to a single room. It is safe
// for concurrent use: outbound delivery goes through the buffered send channel,
// and shutdown is guarded by a sync.Once.
type Client struct {
	ID       string
	Role     Role
	RoomCode string

	manager *HubManager
	conn    wsConn
	send    chan []byte
	done    chan struct{}
	once    sync.Once
}

// newClient constructs a client without starting its pumps. Exposed within the
// package (and to tests) so a fake connection can be injected.
func newClient(manager *HubManager, conn wsConn, role Role, roomCode, id string) *Client {
	return &Client{
		ID:       id,
		Role:     role,
		RoomCode: roomCode,
		manager:  manager,
		conn:     conn,
		send:     make(chan []byte, sendBuffer),
		done:     make(chan struct{}),
	}
}

// Serve registers a freshly upgraded connection with its room and starts the
// read/write pumps. It is the single entry point an HTTP handler needs.
func (m *HubManager) Serve(conn *websocket.Conn, role Role, roomCode, id string) *Client {
	c := newClient(m, conn, role, roomCode, id)
	m.Register(c)
	c.start()
	return c
}

// start launches the read and write pumps. Each pump runs in its own goroutine
// and tears the client down on exit.
func (c *Client) start() {
	go c.writePump()
	go c.readPump()
}

// enqueue queues a pre-marshalled message for delivery. It never blocks: if the
// client is already shutting down the message is dropped, and if the outbound
// buffer is full the client is treated as a slow consumer and closed. This is
// what guarantees a broadcast can never stall on one bad connection.
func (c *Client) enqueue(msg []byte) {
	select {
	case <-c.done:
		return
	default:
	}

	select {
	case c.send <- msg:
	case <-c.done:
	default:
		// Buffer full: drop the slow consumer. close() is idempotent and the
		// send channel is never closed, so this can never panic a broadcaster.
		c.close()
	}
}

// close signals shutdown exactly once. The send channel is deliberately not
// closed so concurrent enqueue calls can never send on a closed channel.
func (c *Client) close() {
	c.once.Do(func() {
		close(c.done)
		_ = c.conn.Close()
	})
}

// cleanup detaches the client from its room and closes the connection. Safe to
// call from either pump; HubManager.unregister is idempotent.
func (c *Client) cleanup() {
	c.manager.unregister(c)
	c.close()
}

// readPump drains inbound frames. Clients are receivers, so messages are
// discarded; the pump exists to detect disconnects and service pong keepalives.
func (c *Client) readPump() {
	defer c.cleanup()

	c.conn.SetReadLimit(maxMessageSize)
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		return c.conn.SetReadDeadline(time.Now().Add(pongWait))
	})

	for {
		if _, _, err := c.conn.ReadMessage(); err != nil {
			return
		}
	}
}

// writePump delivers queued messages and sends periodic pings. Any write error
// terminates the pump and cleans up the client.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.cleanup()
	}()

	for {
		select {
		case msg := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				log.Printf("websocket: write to client %s in room %s failed: %v", c.ID, c.RoomCode, err)
				return
			}
		case <-ticker.C:
			if err := c.conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(writeWait)); err != nil {
				return
			}
		case <-c.done:
			return
		}
	}
}
