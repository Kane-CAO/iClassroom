package handler

import (
	"log"

	"github.com/gin-gonic/gin"

	"iclassroom/backend/internal/service"
	"iclassroom/backend/internal/websocket"
)

// WSHandler upgrades authorised HTTP requests to WebSocket connections and hands
// them to the hub manager. It performs no business logic itself: authorisation
// lives in WSAuthService, and connection pooling/broadcasting lives in the
// websocket package.
type WSHandler struct {
	auth    *service.WSAuthService
	manager *websocket.HubManager
}

// NewWSHandler wires the handler to the auth service and the shared hub manager.
func NewWSHandler(auth *service.WSAuthService, manager *websocket.HubManager) *WSHandler {
	return &WSHandler{auth: auth, manager: manager}
}

// Register mounts the WebSocket endpoint at the application root (NOT under
// /api, as the real-time channel is separate from the REST surface).
func (h *WSHandler) Register(r gin.IRouter) {
	r.GET("/ws", h.Connect)
}

// Connect handles GET /ws?room=ABC123&role=teacher|student|display[&token=xxx].
//
// Authorisation runs first and, on failure, returns the unified JSON error
// envelope with an appropriate status (400/401/403/404). Only after the request
// is authorised do we attempt the WebSocket upgrade; an upgrade failure is
// handled by gorilla (which writes its own HTTP error) and merely logged here.
func (h *WSHandler) Connect(c *gin.Context) {
	roomCode := c.Query("room")
	role := c.Query("role")
	token := c.Query("token")

	info, err := h.auth.Authorize(c.Request.Context(), roomCode, role, token)
	if err != nil {
		respondError(c, err)
		return
	}

	conn, err := websocket.DefaultUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		// Upgrade already wrote the HTTP error response (e.g. 400 for a
		// non-WebSocket request); nothing more to send to the client.
		log.Printf("websocket: upgrade failed for room %s role %s: %v", info.RoomCode, info.Role, err)
		return
	}

	// Serve registers the connection in the room's pool and starts its pumps;
	// the pumps clean the client out of the pool on disconnect.
	h.manager.Serve(conn, websocket.Role(info.Role), info.RoomCode, info.ClientID)
}
