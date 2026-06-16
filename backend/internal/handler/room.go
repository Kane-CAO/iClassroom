package handler

import (
	"github.com/gin-gonic/gin"

	"iclassroom/backend/internal/apperr"
	"iclassroom/backend/internal/response"
	"iclassroom/backend/internal/service"
)

// headerTeacherToken carries the room-level management credential.
const headerTeacherToken = "X-Teacher-Token"

// RoomHandler exposes the teacher-facing room endpoints.
type RoomHandler struct {
	rooms *service.RoomService
}

// NewRoomHandler wires the handler to its service.
func NewRoomHandler(rooms *service.RoomService) *RoomHandler {
	return &RoomHandler{rooms: rooms}
}

// Register mounts the teacher room routes on the given group.
func (h *RoomHandler) Register(rg *gin.RouterGroup) {
	rg.POST("/teacher/rooms", h.Create)
	rg.GET("/teacher/rooms/:roomCode", h.Get)
	rg.GET("/teacher/rooms/:roomCode/overview", h.Overview)
	rg.POST("/teacher/rooms/:roomCode/end", h.End)
}

type createRoomRequest struct {
	Title            string `json:"title"`
	GroupCount       int    `json:"groupCount"`
	GroupCapacity    int    `json:"groupCapacity"`
	AllowChooseGroup bool   `json:"allowChooseGroup"`
}

type groupDTO struct {
	GroupID      int64  `json:"groupId"`
	GroupName    string `json:"groupName"`
	Capacity     int    `json:"capacity"`
	CurrentCount int    `json:"currentCount"`
}

// Create handles POST /api/teacher/rooms.
func (h *RoomHandler) Create(c *gin.Context) {
	var req createRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, apperr.InvalidRequest("malformed request body"))
		return
	}

	res, err := h.rooms.CreateRoom(c.Request.Context(), service.CreateRoomInput{
		Title:            req.Title,
		GroupCount:       req.GroupCount,
		GroupCapacity:    req.GroupCapacity,
		AllowChooseGroup: req.AllowChooseGroup,
	})
	if err != nil {
		respondError(c, err)
		return
	}

	groups := make([]groupDTO, 0, len(res.Groups))
	for _, g := range res.Groups {
		groups = append(groups, groupDTO{
			GroupID:      g.ID,
			GroupName:    g.GroupName,
			Capacity:     g.Capacity,
			CurrentCount: 0, // a brand-new room has no members yet
		})
	}

	response.Success(c, gin.H{
		"roomCode":            res.Room.RoomCode,
		"teacherToken":        res.Room.TeacherToken,
		"joinUrl":             res.JoinURL,
		"teacherDashboardUrl": res.TeacherDashboardURL,
		"groups":              groups,
	})
}

// Get handles GET /api/teacher/rooms/:roomCode.
func (h *RoomHandler) Get(c *gin.Context) {
	roomCode := c.Param("roomCode")
	token := c.GetHeader(headerTeacherToken)

	room, err := h.rooms.GetRoom(c.Request.Context(), roomCode, token)
	if err != nil {
		respondError(c, err)
		return
	}

	response.Success(c, gin.H{
		"roomCode":         room.RoomCode,
		"title":            room.Title,
		"status":           room.Status,
		"groupCount":       room.GroupCount,
		"groupCapacity":    room.GroupCapacity,
		"allowChooseGroup": room.AllowChooseGroup,
		"createdAt":        room.CreatedAt,
	})
}

// Overview handles GET /api/teacher/rooms/:roomCode/overview.
func (h *RoomHandler) Overview(c *gin.Context) {
	roomCode := c.Param("roomCode")
	token := c.GetHeader(headerTeacherToken)

	ov, err := h.rooms.GetOverview(c.Request.Context(), roomCode, token)
	if err != nil {
		respondError(c, err)
		return
	}

	groups := make([]gin.H, 0, len(ov.Groups))
	for _, g := range ov.Groups {
		groups = append(groups, gin.H{
			"groupId":      g.ID,
			"groupName":    g.GroupName,
			"capacity":     g.Capacity,
			"currentCount": g.CurrentCount,
			"scoreTotal":   g.ScoreTotal,
		})
	}

	response.Success(c, gin.H{
		"roomCode":     ov.Room.RoomCode,
		"title":        ov.Room.Title,
		"status":       ov.Room.Status,
		"joinUrl":      ov.JoinURL,
		"studentCount": ov.StudentCount,
		"groups":       groups,
		"tasks":        []any{}, // populated in a later step
	})
}

// End handles POST /api/teacher/rooms/:roomCode/end.
func (h *RoomHandler) End(c *gin.Context) {
	roomCode := c.Param("roomCode")
	token := c.GetHeader(headerTeacherToken)

	room, err := h.rooms.EndRoom(c.Request.Context(), roomCode, token)
	if err != nil {
		respondError(c, err)
		return
	}

	response.Success(c, gin.H{
		"roomCode": room.RoomCode,
		"status":   room.Status,
		"endedAt":  room.EndedAt,
	})
}
