package service

import (
	"log"

	"iclassroom/backend/internal/websocket"
)

// EventBroadcaster is the narrow surface the services use to publish real-time
// events after a successful database operation. It is satisfied by
// *websocket.HubManager. Depending on this interface (rather than the concrete
// hub) keeps the services decoupled from the transport: production injects the
// hub manager, tests inject a fake. There is no import cycle — the websocket
// package does not import service.
type EventBroadcaster interface {
	Broadcast(roomCode string, event websocket.Event) error
}

// noopBroadcaster is the default when no broadcaster is injected: it drops every
// event. This lets services run (and most unit tests pass) without a live hub.
type noopBroadcaster struct{}

func (noopBroadcaster) Broadcast(string, websocket.Event) error { return nil }

// resolveBroadcaster returns b, or a noop if b is nil, so a service always has a
// non-nil broadcaster to call.
func resolveBroadcaster(b EventBroadcaster) EventBroadcaster {
	if b == nil {
		return noopBroadcaster{}
	}
	return b
}

// emit builds the standard event envelope and broadcasts it, logging and then
// swallowing any error. Callers MUST invoke this only after their database work
// has committed; a broadcast failure must never change the business result, so
// the error is intentionally not propagated.
func emit(b EventBroadcaster, roomCode string, typ websocket.EventType, data any) {
	if b == nil {
		return
	}
	if err := b.Broadcast(roomCode, websocket.NewEvent(typ, roomCode, data)); err != nil {
		log.Printf("service: broadcast %q for room %s failed: %v", typ, roomCode, err)
	}
}
