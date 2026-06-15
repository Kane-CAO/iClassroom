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

	rg.GET("/student/me/tasks", h.ListForStudent)
	rg.POST("/student/tasks/:taskId/submit", h.Submit)
	rg.GET("/teacher/tasks/:taskId/submissions", h.ListSubmissions)
}

type createTaskRequest struct {
	Title          string    `json:"title"`
	Description    string    `json:"description"`
	AttachmentURL  string    `json:"attachmentUrl"`
	DeadlineAt     time.Time `json:"deadlineAt"`
	TargetType     string    `json:"targetType"`
	TargetGroupIDs []int64   `json:"targetGroupIds"`
}

type submitTaskRequest struct {
	ContentText string `json:"contentText"`
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

func (h *TaskHandler) ListForStudent(c *gin.Context) {
	token := c.GetHeader(headerStudentToken)

	tasks, err := h.tasks.ListForStudent(c.Request.Context(), token)
	if err != nil {
		respondError(c, err)
		return
	}

	out := make([]gin.H, 0, len(tasks))
	for _, task := range tasks {
		out = append(out, studentTaskViewJSON(&task))
	}

	response.Success(c, out)
}

func (h *TaskHandler) Submit(c *gin.Context) {
	taskID, err := strconv.ParseInt(c.Param("taskId"), 10, 64)
	if err != nil || taskID <= 0 {
		respondError(c, apperr.TaskNotFound())
		return
	}

	var req submitTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, apperr.InvalidRequest("malformed request body"))
		return
	}

	submission, err := h.tasks.SubmitText(c.Request.Context(), taskID, c.GetHeader(headerStudentToken), req.ContentText)
	if err != nil {
		respondError(c, err)
		return
	}

	response.Success(c, submissionJSON(submission))
}

func (h *TaskHandler) ListSubmissions(c *gin.Context) {
	taskID, err := strconv.ParseInt(c.Param("taskId"), 10, 64)
	if err != nil || taskID <= 0 {
		respondError(c, apperr.TaskNotFound())
		return
	}

	items, err := h.tasks.ListSubmissionsForTeacher(c.Request.Context(), taskID, c.GetHeader(headerTeacherToken))
	if err != nil {
		respondError(c, err)
		return
	}

	out := make([]gin.H, 0, len(items))
	for _, item := range items {
		out = append(out, submissionWithStudentJSON(&item))
	}

	response.Success(c, gin.H{
		"taskId":      taskID,
		"submissions": out,
	})
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

func studentTaskViewJSON(task *service.StudentTaskView) gin.H {
	var submission any
	if task.Submission != nil {
		submission = submissionJSON(task.Submission)
	}

	return gin.H{
		"taskId":         task.Task.ID,
		"title":          task.Task.Title,
		"description":    task.Task.Description,
		"attachmentUrl":  task.Task.AttachmentURL,
		"deadlineAt":     task.Task.DeadlineAt,
		"targetType":     task.Task.TargetType,
		"targetGroupIds": task.TargetGroupIDs,
		"status":         task.Task.Status,
		"createdAt":      task.Task.CreatedAt,
		"submission":     submission,
	}
}

func submissionJSON(submission *domain.Submission) gin.H {
	return gin.H{
		"submissionId": submission.ID,
		"taskId":       submission.TaskID,
		"studentId":    submission.StudentID,
		"groupId":      submission.GroupID,
		"contentText":  submission.ContentText,
		"status":       submission.Status,
		"score":        submission.Score,
		"comment":      submission.Comment,
		"submittedAt":  submission.SubmittedAt,
		"gradedAt":     submission.GradedAt,
	}
}

func submissionWithStudentJSON(item *service.SubmissionView) gin.H {
	return gin.H{
		"submissionId": item.Submission.ID,
		"taskId":       item.Submission.TaskID,
		"studentId":    item.Student.ID,
		"nickname":     item.Student.Nickname,
		"groupId":      item.Group.ID,
		"groupName":    item.Group.GroupName,
		"contentText":  item.Submission.ContentText,
		"status":       item.Submission.Status,
		"score":        item.Submission.Score,
		"comment":      item.Submission.Comment,
		"submittedAt":  item.Submission.SubmittedAt,
		"gradedAt":     item.Submission.GradedAt,
	}
}
