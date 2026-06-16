package websocket

import (
	"encoding/json"
	"log"
	"sync"
)

// RoomHub is the connection pool for a single roomCode. Teacher, student and
// display clients all live in the same hub. The clients map is guarded by a
// RWMutex; broadcasts take only a read lock and never perform I/O while holding
// it, so a slow or dead connection cannot stall registration or other senders.
type RoomHub struct {
	roomCode string

	mu      sync.RWMutex
	clients map[*Client]struct{}
}

// newRoomHub creates an empty hub for the given room.
func newRoomHub(roomCode string) *RoomHub {
	return &RoomHub{
		roomCode: roomCode,
		clients:  make(map[*Client]struct{}),
	}
}

// add inserts a client into the pool.
func (h *RoomHub) add(c *Client) {
	h.mu.Lock()
	h.clients[c] = struct{}{}
	h.mu.Unlock()
}

// remove deletes a client from the pool. It is a no-op if the client was
// already removed, so it is safe to call from both pumps.
func (h *RoomHub) remove(c *Client) {
	h.mu.Lock()
	delete(h.clients, c)
	h.mu.Unlock()
}

// isEmpty reports whether the pool currently holds no clients.
func (h *RoomHub) isEmpty() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients) == 0
}

// Count returns the number of connected clients.
func (h *RoomHub) Count() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// Broadcast marshals the event once and enqueues it for every client. A failure
// to marshal is logged and swallowed; per-client delivery is non-blocking, so
// one stuck client never blocks the room and nothing here can panic.
func (h *RoomHub) Broadcast(evt Event) {
	payload, err := json.Marshal(evt)
	if err != nil {
		log.Printf("websocket: marshal event %q for room %s failed: %v", evt.Type, h.roomCode, err)
		return
	}

	// Snapshot the client set under a read lock, then release it before
	// enqueuing so delivery happens without holding the pool lock.
	h.mu.RLock()
	targets := make([]*Client, 0, len(h.clients))
	for c := range h.clients {
		targets = append(targets, c)
	}
	h.mu.RUnlock()

	for _, c := range targets {
		c.enqueue(payload)
	}
}
