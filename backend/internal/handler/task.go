package handler

import (
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
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

	rg.POST("/teacher/submissions/:submissionId/grade", h.GradeSubmission)
	rg.GET("/student/me/results", h.StudentResults)
	rg.GET("/student/rooms/:roomCode/ranking", h.StudentRanking)

	// Compatibility routes kept temporarily for existing callers.
	// They can be removed after frontend and teammate branches migrate.
	rg.PATCH("/teacher/submissions/:submissionId/grade", h.GradeSubmission)
	rg.GET("/teacher/rooms/:roomCode/leaderboard", h.TeacherLeaderboard)
	rg.GET("/student/me/leaderboard", h.StudentLeaderboard)
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

type gradeSubmissionRequest struct {
	Score   int    `json:"score"`
	Comment string `json:"comment"`
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

	contentType := c.GetHeader("Content-Type")
	token := c.GetHeader(headerStudentToken)

	switch {
	case strings.HasPrefix(contentType, "multipart/form-data"),
		strings.HasPrefix(contentType, "application/x-www-form-urlencoded"):
		contentText := c.PostForm("contentText")
		if strings.HasPrefix(contentType, "multipart/form-data") {
			form, err := c.MultipartForm()
			if err != nil {
				respondError(c, apperr.InvalidRequest("malformed multipart form"))
				return
			}

			uploads, err := readUploadedImages(form)
			if err != nil {
				respondError(c, err)
				return
			}

			submission, err := h.tasks.SubmitWithImages(c.Request.Context(), taskID, token, contentText, uploads)
			if err != nil {
				respondError(c, err)
				return
			}

			response.Success(c, submissionJSON(submission))
			return
		}

		submission, err := h.tasks.SubmitText(c.Request.Context(), taskID, token, contentText)
		if err != nil {
			respondError(c, err)
			return
		}

		response.Success(c, submissionJSON(submission))
		return
	default:
		var req submitTaskRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			respondError(c, apperr.InvalidRequest("malformed request body"))
			return
		}

		submission, err := h.tasks.SubmitText(c.Request.Context(), taskID, token, req.ContentText)
		if err != nil {
			respondError(c, err)
			return
		}

		response.Success(c, submissionJSON(submission))
		return
	}
}

func (h *TaskHandler) GradeSubmission(c *gin.Context) {
	submissionID, err := strconv.ParseInt(c.Param("submissionId"), 10, 64)
	if err != nil || submissionID <= 0 {
		respondError(c, apperr.SubmissionNotFound())
		return
	}

	var req gradeSubmissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, apperr.InvalidRequest("malformed request body"))
		return
	}

	submission, err := h.tasks.GradeSubmission(
		c.Request.Context(),
		submissionID,
		c.GetHeader(headerTeacherToken),
		req.Score,
		req.Comment,
	)
	if err != nil {
		respondError(c, err)
		return
	}

	response.Success(c, gradeSubmissionJSON(submission))
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

	response.Success(c, out)
}

func (h *TaskHandler) TeacherLeaderboard(c *gin.Context) {
	roomCode := c.Param("roomCode")
	token := c.GetHeader(headerTeacherToken)

	view, err := h.tasks.GetLeaderboardForTeacher(c.Request.Context(), roomCode, token)
	if err != nil {
		respondError(c, err)
		return
	}

	response.Success(c, leaderboardJSON(view))
}

func (h *TaskHandler) StudentLeaderboard(c *gin.Context) {
	token := c.GetHeader(headerStudentToken)

	view, err := h.tasks.GetLeaderboardForStudent(c.Request.Context(), token)
	if err != nil {
		respondError(c, err)
		return
	}

	response.Success(c, leaderboardJSON(view))
}

func (h *TaskHandler) StudentRanking(c *gin.Context) {
	roomCode := c.Param("roomCode")
	token := c.GetHeader(headerStudentToken)

	view, err := h.tasks.GetRankingForStudent(c.Request.Context(), roomCode, token)
	if err != nil {
		respondError(c, err)
		return
	}

	response.Success(c, rankingJSON(view))
}

func (h *TaskHandler) StudentResults(c *gin.Context) {
	token := c.GetHeader(headerStudentToken)

	view, err := h.tasks.GetStudentResults(c.Request.Context(), token)
	if err != nil {
		respondError(c, err)
		return
	}

	response.Success(c, studentResultsJSON(view))
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
	mySubmissionStatus := "notSubmitted"
	var myScore any

	if task.Submission != nil {
		mySubmissionStatus = string(task.Submission.Status)
		if task.Submission.Score != nil {
			myScore = *task.Submission.Score
		}
	}

	return gin.H{
		"taskId":             task.Task.ID,
		"title":              task.Task.Title,
		"description":        task.Task.Description,
		"deadlineAt":         task.Task.DeadlineAt,
		"status":             task.Task.Status,
		"mySubmissionStatus": mySubmissionStatus,
		"myScore":            myScore,
	}
}

func submissionJSON(submission *domain.Submission) gin.H {
	return gin.H{
		"submissionId": submission.ID,
		"taskId":       submission.TaskID,
		"studentId":    submission.StudentID,
		"groupId":      submission.GroupID,
		"contentText":  submission.ContentText,
		"images":       submissionImagesJSON(submission.Images),
		"status":       submission.Status,
		"score":        submission.Score,
		"comment":      submission.Comment,
		"submittedAt":  submission.SubmittedAt,
		"gradedAt":     submission.GradedAt,
	}
}

func gradeSubmissionJSON(result *service.GradeSubmissionResult) gin.H {
	return gin.H{
		"submissionId":    result.Submission.ID,
		"score":           result.Submission.Score,
		"comment":         result.Submission.Comment,
		"status":          result.Submission.Status,
		"gradedAt":        result.Submission.GradedAt,
		"groupScoreTotal": result.GroupScoreTotal,
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
		"images":       submissionImagesJSON(item.Submission.Images),
		"status":       item.Submission.Status,
		"score":        item.Submission.Score,
		"comment":      item.Submission.Comment,
		"submittedAt":  item.Submission.SubmittedAt,
		"gradedAt":     item.Submission.GradedAt,
	}
}

func submissionImagesJSON(images []domain.SubmissionImage) []gin.H {
	out := make([]gin.H, 0, len(images))
	for _, image := range images {
		out = append(out, gin.H{
			"imageId":  image.ID,
			"fileUrl":  image.FileURL,
			"fileName": image.FileName,
			"fileSize": image.FileSize,
			"mimeType": image.MimeType,
		})
	}
	return out
}

func studentResultsJSON(view *service.StudentResultsView) gin.H {
	results := make([]gin.H, 0, len(view.Results))
	for _, result := range view.Results {
		results = append(results, gin.H{
			"taskId":           result.Task.ID,
			"taskTitle":        result.Task.Title,
			"submissionStatus": result.SubmissionStatus,
			"score":            result.Score,
			"comment":          result.Comment,
			"submittedAt":      result.SubmittedAt,
			"gradedAt":         result.GradedAt,
		})
	}

	return gin.H{
		"studentId": view.Student.ID,
		"nickname":  view.Student.Nickname,
		"groupId":   view.Group.ID,
		"groupName": view.Group.GroupName,
		"results":   results,
	}
}

func rankingJSON(view *service.LeaderboardView) []gin.H {
	entries := make([]gin.H, 0, len(view.Entries))
	for _, entry := range view.Entries {
		entries = append(entries, gin.H{
			"rank":         entry.Rank,
			"groupId":      entry.Group.ID,
			"groupName":    entry.Group.GroupName,
			"scoreTotal":   entry.Group.ScoreTotal,
			"studentCount": entry.CurrentCount,
		})
	}
	return entries
}

func readUploadedImages(form *multipart.Form) ([]service.UploadedFile, error) {
	files := form.File["images[]"]
	if len(files) == 0 {
		return nil, nil
	}
	if len(files) > 3 {
		return nil, apperr.TooManyImages()
	}

	uploads := make([]service.UploadedFile, 0, len(files))
	for _, header := range files {
		f, err := header.Open()
		if err != nil {
			return nil, apperr.UploadFailed()
		}
		data, err := io.ReadAll(f)
		_ = f.Close()
		if err != nil {
			return nil, apperr.UploadFailed()
		}

		uploads = append(uploads, service.UploadedFile{
			FileName: header.Filename,
			MimeType: http.DetectContentType(data),
			Data:     data,
		})
	}

	return uploads, nil
}

func leaderboardJSON(view *service.LeaderboardView) gin.H {
	entries := make([]gin.H, 0, len(view.Entries))

	for _, entry := range view.Entries {
		entries = append(entries, gin.H{
			"rank":         entry.Rank,
			"groupId":      entry.Group.ID,
			"groupName":    entry.Group.GroupName,
			"scoreTotal":   entry.Group.ScoreTotal,
			"currentCount": entry.CurrentCount,
			"isMyGroup":    entry.IsMyGroup,
		})
	}

	return gin.H{
		"roomCode":    view.RoomCode,
		"leaderboard": entries,
	}
}
