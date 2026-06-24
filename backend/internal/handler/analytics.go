package handler

import (
	"github.com/gin-gonic/gin"

	"iclassroom/backend/internal/response"
	"iclassroom/backend/internal/service"
)

type AnalyticsHandler struct {
	analytics *service.AnalyticsService
}

func NewAnalyticsHandler(analytics *service.AnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{analytics: analytics}
}

func (h *AnalyticsHandler) Register(rg *gin.RouterGroup) {
	rg.GET("/teacher/rooms/:roomCode/analytics", h.Get)
}

func (h *AnalyticsHandler) Get(c *gin.Context) {
	view, err := h.analytics.Get(c.Request.Context(), c.Param("roomCode"), legacyTeacherToken(c), currentTeacherID(c))
	if err != nil {
		respondError(c, err)
		return
	}

	groupScores := make([]gin.H, 0, len(view.GroupScores))
	for _, item := range view.GroupScores {
		groupScores = append(groupScores, gin.H{
			"groupId":    item.GroupID,
			"groupName":  item.GroupName,
			"scoreTotal": item.ScoreTotal,
		})
	}

	taskCompletion := make([]gin.H, 0, len(view.TaskCompletion))
	for _, item := range view.TaskCompletion {
		taskCompletion = append(taskCompletion, gin.H{
			"taskId":             item.TaskID,
			"taskTitle":          item.TaskTitle,
			"submittedCount":     item.SubmittedCount,
			"targetStudentCount": item.TargetStudentCount,
			"completionRate":     item.CompletionRate,
		})
	}

	timeline := make([]gin.H, 0, len(view.SubmissionTimeline))
	for _, item := range view.SubmissionTimeline {
		timeline = append(timeline, gin.H{
			"time":  item.Time,
			"count": item.Count,
		})
	}

	response.Success(c, gin.H{
		"studentCount":       view.StudentCount,
		"onlineCount":        view.OnlineCount,
		"submissionRate":     view.SubmissionRate,
		"groupScores":        groupScores,
		"taskCompletion":     taskCompletion,
		"submissionTimeline": timeline,
	})
}
