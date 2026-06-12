package handler

import (
	"context"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"iclassroom/backend/internal/apperr"
	"iclassroom/backend/internal/domain"
	"iclassroom/backend/internal/response"
	"iclassroom/backend/internal/service"
)

type TaskHandler struct {
	tasks *service.TaskService
}

func NewTaskHandler(tasks *service.TaskService) *TaskHandler {
	return &TaskHandler{tasks: tasks}
}

func (h *TaskHandler) Register(rg *gin.RouterGroup) {
	rg.POST("/teacher/rooms/:roomCode/tasks", h.Create)
	rg.GET("/teacher/rooms/:roomCode/tasks", h.List)
	rg.PATCH("/teacher/tasks/:taskId/pause", h.Pause)
	rg.PATCH("/teacher/tasks/:taskId/close", h.Close)
}

type createTaskRequest struct {
	Title          string    `json:"title"`
	Description    string    `json:"description"`
	AttachmentURL  string    `json:"attachmentUrl"`
	DeadlineAt     time.Time `json:"deadlineAt"`
	TargetType     string    `json:"targetType"`
	TargetGroupIDs []int64   `json:"targetGroupIds"`
}

func (h *TaskHandler) Create(c *gin.Context) {
	roomCode := c.Param("roomCode")
	token := c.GetHeader(headerTeacherToken)

	var req createTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, apperr.InvalidRequest("malformed request body"))
		return
	}

	res, err := h.tasks.Create(c.Request.Context(), service.CreateTaskInput{
		RoomCode:       roomCode,
		TeacherToken:   token,
		Title:          req.Title,
		Description:    req.Description,
		AttachmentURL:  req.AttachmentURL,
		DeadlineAt:     req.DeadlineAt,
		TargetType:     domain.TargetType(req.TargetType),
		TargetGroupIDs: req.TargetGroupIDs,
	})
	if err != nil {
		respondError(c, err)
		return
	}

	response.Success(c, taskViewJSON(res))
}

func (h *TaskHandler) List(c *gin.Context) {
	roomCode := c.Param("roomCode")
	token := c.GetHeader(headerTeacherToken)

	tasks, err := h.tasks.ListForTeacher(c.Request.Context(), roomCode, token)
	if err != nil {
		respondError(c, err)
		return
	}

	out := make([]gin.H, 0, len(tasks))
	for _, task := range tasks {
		out = append(out, taskViewJSON(&task))
	}

	response.Success(c, out)
}

func (h *TaskHandler) Pause(c *gin.Context) {
	h.updateStatus(c, h.tasks.Pause)
}

func (h *TaskHandler) Close(c *gin.Context) {
	h.updateStatus(c, h.tasks.Close)
}

func (h *TaskHandler) updateStatus(c *gin.Context, update func(context.Context, int64, string) (*service.TaskView, error)) {
	taskID, err := strconv.ParseInt(c.Param("taskId"), 10, 64)
	if err != nil || taskID <= 0 {
		respondError(c, apperr.TaskNotFound())
		return
	}

	res, err := update(c.Request.Context(), taskID, c.GetHeader(headerTeacherToken))
	if err != nil {
		respondError(c, err)
		return
	}

	response.Success(c, gin.H{
		"taskId": res.Task.ID,
		"status": res.Task.Status,
	})
}

func taskViewJSON(task *service.TaskView) gin.H {
	return gin.H{
		"taskId":             task.Task.ID,
		"roomCode":           task.RoomCode,
		"title":              task.Task.Title,
		"description":        task.Task.Description,
		"attachmentUrl":      task.Task.AttachmentURL,
		"deadlineAt":         task.Task.DeadlineAt,
		"targetType":         task.Task.TargetType,
		"targetGroupIds":     task.TargetGroupIDs,
		"status":             task.Task.Status,
		"submittedCount":     task.SubmittedCount,
		"targetStudentCount": task.TargetStudentCount,
		"createdAt":          task.Task.CreatedAt,
	}
}
