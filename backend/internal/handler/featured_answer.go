package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"iclassroom/backend/internal/apperr"
	"iclassroom/backend/internal/domain"
	"iclassroom/backend/internal/response"
	"iclassroom/backend/internal/service"
)

type FeaturedAnswerHandler struct {
	featured *service.FeaturedAnswerService
}

func NewFeaturedAnswerHandler(featured *service.FeaturedAnswerService) *FeaturedAnswerHandler {
	return &FeaturedAnswerHandler{featured: featured}
}

func (h *FeaturedAnswerHandler) Register(rg *gin.RouterGroup) {
	rg.POST("/teacher/submissions/:submissionId/feature", h.Feature)
}

type featureSubmissionRequest struct {
	DisplayMode string `json:"displayMode"`
}

func (h *FeaturedAnswerHandler) Feature(c *gin.Context) {
	submissionID, err := strconv.ParseInt(c.Param("submissionId"), 10, 64)
	if err != nil || submissionID <= 0 {
		respondError(c, apperr.SubmissionNotFound())
		return
	}

	var req featureSubmissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, apperr.InvalidRequest("malformed request body"))
		return
	}

	fa, err := h.featured.Feature(
		c.Request.Context(),
		submissionID,
		legacyTeacherToken(c),
		domain.DisplayMode(req.DisplayMode),
		currentTeacherID(c),
	)
	if err != nil {
		respondError(c, err)
		return
	}

	response.Success(c, gin.H{
		"featuredId":   fa.ID,
		"submissionId": fa.SubmissionID,
		"displayMode":  fa.DisplayMode,
	})
}
