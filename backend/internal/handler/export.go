package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"iclassroom/backend/internal/service"
)

// ExportHandler exposes the teacher room export endpoint.
type ExportHandler struct {
	exports *service.ExportService
}

func NewExportHandler(exports *service.ExportService) *ExportHandler {
	return &ExportHandler{exports: exports}
}

func (h *ExportHandler) Register(rg *gin.RouterGroup) {
	rg.GET("/teacher/rooms/:roomCode/export", h.Get)
}

// Get handles GET /api/teacher/rooms/:roomCode/export and streams a zip file.
func (h *ExportHandler) Get(c *gin.Context) {
	roomCode := c.Param("roomCode")
	token := legacyTeacherToken(c)

	res, err := h.exports.Export(c.Request.Context(), roomCode, token, currentTeacherID(c))
	if err != nil {
		respondError(c, err)
		return
	}

	c.Header("Content-Type", "application/zip")
	c.Header("Content-Disposition", `attachment; filename="`+res.FileName+`"`)
	c.Status(http.StatusOK)
	_, _ = c.Writer.Write(res.Data)
}
