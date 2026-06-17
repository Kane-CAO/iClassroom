package handler

import (
	"github.com/gin-gonic/gin"

	"iclassroom/backend/internal/domain"
	"iclassroom/backend/internal/response"
	"iclassroom/backend/internal/service"
)

type DisplayHandler struct {
	display *service.DisplayService
}

func NewDisplayHandler(display *service.DisplayService) *DisplayHandler {
	return &DisplayHandler{display: display}
}

func (h *DisplayHandler) Register(rg *gin.RouterGroup) {
	rg.GET("/teacher/rooms/:roomCode/display", h.Get)
}

func (h *DisplayHandler) Get(c *gin.Context) {
	view, err := h.display.Get(c.Request.Context(), c.Param("roomCode"))
	if err != nil {
		respondError(c, err)
		return
	}

	groups := make([]gin.H, 0, len(view.Groups))
	for _, group := range view.Groups {
		groups = append(groups, gin.H{
			"groupId":      group.ID,
			"groupName":    group.GroupName,
			"capacity":     group.Capacity,
			"currentCount": group.CurrentCount,
			"scoreTotal":   group.ScoreTotal,
		})
	}

	ranking := make([]gin.H, 0, len(view.Ranking))
	for _, entry := range view.Ranking {
		ranking = append(ranking, gin.H{
			"rank":       entry.Rank,
			"groupId":    entry.Group.ID,
			"groupName":  entry.Group.GroupName,
			"scoreTotal": entry.Group.ScoreTotal,
		})
	}

	featured := make([]gin.H, 0, len(view.FeaturedAnswers))
	for _, item := range view.FeaturedAnswers {
		row := gin.H{
			"featuredId":   item.FeaturedID,
			"submissionId": item.SubmissionID,
			"taskId":       item.TaskID,
			"displayMode":  item.DisplayMode,
			"contentText":  item.ContentText,
			"score":        item.Score,
			"submittedAt":  item.SubmittedAt,
		}
		if item.DisplayMode == domain.DisplayShowGroup {
			row["groupId"] = item.GroupID
			row["groupName"] = item.GroupName
		}
		featured = append(featured, row)
	}

	response.Success(c, gin.H{
		"roomCode":        view.Room.RoomCode,
		"title":           view.Room.Title,
		"status":          view.Room.Status,
		"groups":          groups,
		"ranking":         ranking,
		"currentTask":     displayTaskJSON(view.CurrentTask),
		"featuredAnswers": featured,
	})
}

func displayTaskJSON(task *service.DisplayTaskView) any {
	if task == nil {
		return nil
	}
	return gin.H{
		"taskId":             task.TaskID,
		"title":              task.Title,
		"deadlineAt":         task.DeadlineAt,
		"submittedCount":     task.SubmittedCount,
		"targetStudentCount": task.TargetStudentCount,
		"completionRate":     task.CompletionRate,
	}
}
