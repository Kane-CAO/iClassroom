package websocket

import "sync"

// HubManager owns one RoomHub per roomCode. The hubs map is guarded by a mutex.
// All mutations that decide whether a hub exists (register, unregister and the
// empty-hub cleanup) happen atomically under that lock so a client can never be
// registered onto a hub that is concurrently being removed from the map.
//
// Broadcast looks the hub up under the lock but releases it before fanning out,
// keeping the manager lock off the per-client I/O path.
type HubManager struct {
	mu   sync.Mutex
	hubs map[string]*RoomHub
}

// NewHubManager creates an empty manager with no rooms.
func NewHubManager() *HubManager {
	return &HubManager{
		hubs: make(map[string]*RoomHub),
	}
}

// Register adds a client to its room's hub, creating the hub on first use.
func (m *HubManager) Register(c *Client) {
	m.mu.Lock()
	defer m.mu.Unlock()

	h, ok := m.hubs[c.RoomCode]
	if !ok {
		h = newRoomHub(c.RoomCode)
		m.hubs[c.RoomCode] = h
	}
	h.add(c)
}

// unregister removes a client from its hub and drops the hub once it is empty,
// freeing memory for ended or idle rooms. Idempotent: removing a client that is
// already gone leaves the hub untouched.
func (m *HubManager) unregister(c *Client) {
	m.mu.Lock()
	defer m.mu.Unlock()

	h, ok := m.hubs[c.RoomCode]
	if !ok {
		return
	}
	h.remove(c)
	if h.isEmpty() {
		delete(m.hubs, c.RoomCode)
	}
}

// Broadcast delivers an event to every client in the named room. It is the only
// method business services call. Unknown or empty rooms are silently ignored
// (returns nil) so services never need to know whether anyone is connected. The
// only non-nil error is a marshal failure of the event payload.
func (m *HubManager) Broadcast(roomCode string, evt Event) error {
	m.mu.Lock()
	h, ok := m.hubs[roomCode]
	m.mu.Unlock()
	if !ok {
		return nil
	}
	return h.Broadcast(evt)
}

// RoomCount returns the number of rooms with at least one connected client.
func (m *HubManager) RoomCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.hubs)
}

// ClientCount returns the number of clients connected to the given room, or 0
// if the room has no hub.
func (m *HubManager) ClientCount(roomCode string) int {
	m.mu.Lock()
	h, ok := m.hubs[roomCode]
	m.mu.Unlock()
	if !ok {
		return 0
	}
	return h.Count()
}
