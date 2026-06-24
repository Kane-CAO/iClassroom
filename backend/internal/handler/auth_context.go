package handler

import (
	"github.com/gin-gonic/gin"

	"iclassroom/backend/internal/domain"
	"iclassroom/backend/internal/middleware"
)

func currentTeacherID(c *gin.Context) int64 {
	user, ok := middleware.CurrentUser(c)
	if !ok || user.Role != domain.RoleTeacher {
		return 0
	}
	return user.UserID
}

func legacyTeacherToken(c *gin.Context) string {
	return c.GetHeader(headerTeacherToken)
}
