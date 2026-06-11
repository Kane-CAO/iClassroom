package handler

import (
	"github.com/gin-gonic/gin"

	"iclassroom/backend/internal/apperr"
	"iclassroom/backend/internal/response"
	"iclassroom/backend/internal/service"
)

// headerStudentToken carries the student session credential.
const headerStudentToken = "X-Student-Token"

// StudentHandler exposes the student-facing room entry endpoints.
type StudentHandler struct {
	students *service.StudentService
}

// NewStudentHandler wires the handler to its service.
func NewStudentHandler(students *service.StudentService) *StudentHandler {
	return &StudentHandler{students: students}
}

// Register mounts the student room routes on the given group.
func (h *StudentHandler) Register(rg *gin.RouterGroup) {
	rg.GET("/student/rooms/:roomCode", h.GetRoom)
	rg.POST("/student/rooms/:roomCode/join", h.Join)
	rg.POST("/student/rooms/:roomCode/resume", h.Resume)
}

// GetRoom handles GET /api/student/rooms/:roomCode.
func (h *StudentHandler) GetRoom(c *gin.Context) {
	roomCode := c.Param("roomCode")

	view, err := h.students.GetRoomForStudent(c.Request.Context(), roomCode)
	if err != nil {
		respondError(c, err)
		return
	}

	groups := make([]gin.H, 0, len(view.Groups))
	for _, g := range view.Groups {
		groups = append(groups, gin.H{
			"groupId":      g.ID,
			"groupName":    g.GroupName,
			"capacity":     g.Capacity,
			"currentCount": g.CurrentCount,
			"available":    g.CurrentCount < g.Capacity,
		})
	}

	response.Success(c, gin.H{
		"roomCode": view.Room.RoomCode,
		"title":    view.Room.Title,
		"status":   view.Room.Status,
		"groups":   groups,
	})
}

type joinRequest struct {
	Nickname string `json:"nickname"`
	GroupID  int64  `json:"groupId"`
}

// Join handles POST /api/student/rooms/:roomCode/join.
func (h *StudentHandler) Join(c *gin.Context) {
	roomCode := c.Param("roomCode")

	var req joinRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, apperr.InvalidRequest("malformed request body"))
		return
	}

	res, err := h.students.Join(c.Request.Context(), roomCode, req.Nickname, req.GroupID)
	if err != nil {
		respondError(c, err)
		return
	}

	response.Success(c, gin.H{
		"studentId":   res.Student.ID,
		"clientToken": res.Student.ClientToken,
		"roomCode":    res.RoomCode,
		"nickname":    res.Student.Nickname,
		"groupId":     res.Student.GroupID,
		"groupName":   res.GroupName,
	})
}

// Resume handles POST /api/student/rooms/:roomCode/resume.
func (h *StudentHandler) Resume(c *gin.Context) {
	roomCode := c.Param("roomCode")
	token := c.GetHeader(headerStudentToken)

	res, err := h.students.Resume(c.Request.Context(), roomCode, token)
	if err != nil {
		respondError(c, err)
		return
	}

	response.Success(c, gin.H{
		"studentId":  res.Student.ID,
		"roomCode":   res.RoomCode,
		"nickname":   res.Student.Nickname,
		"groupId":    res.Student.GroupID,
		"groupName":  res.GroupName,
		"roomStatus": res.RoomStatus,
	})
}
