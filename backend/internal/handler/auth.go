package handler

import (
	"context"
	"strings"

	"github.com/gin-gonic/gin"

	"iclassroom/backend/internal/apperr"
	"iclassroom/backend/internal/domain"
	"iclassroom/backend/internal/middleware"
	"iclassroom/backend/internal/response"
	"iclassroom/backend/internal/service"
)

type AuthHandler struct {
	auth *service.AuthService
}

func NewAuthHandler(auth *service.AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

func (h *AuthHandler) Register(rg *gin.RouterGroup) {
	rg.POST("/auth/admin/login", h.AdminLogin)
	rg.POST("/auth/teacher/login", h.TeacherLogin)
	rg.POST("/auth/logout", h.Logout)
	rg.GET("/auth/me", h.Me)
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (h *AuthHandler) AdminLogin(c *gin.Context) {
	h.login(c, h.auth.LoginAdmin)
}

func (h *AuthHandler) TeacherLogin(c *gin.Context) {
	h.login(c, h.auth.LoginTeacher)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	token := bearerToken(c.GetHeader("Authorization"))
	if err := h.auth.Logout(c.Request.Context(), token); err != nil {
		respondError(c, err)
		return
	}
	response.Success(c, gin.H{})
}

func (h *AuthHandler) Me(c *gin.Context) {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		respondError(c, apperr.Unauthorized())
		return
	}
	response.Success(c, userJSON(*user))
}

func (h *AuthHandler) login(c *gin.Context, fn func(context.Context, string, string) (*service.LoginResult, error)) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, apperr.InvalidRequest("malformed request body"))
		return
	}
	res, err := fn(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		respondError(c, err)
		return
	}
	response.Success(c, gin.H{
		"token": res.Token,
		"user":  userJSON(res.User),
	})
}

func bearerToken(header string) string {
	parts := strings.Fields(strings.TrimSpace(header))
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return parts[1]
}

func userJSON(user domain.AuthUser) gin.H {
	return gin.H{
		"userId":      user.UserID,
		"role":        user.Role,
		"username":    user.Username,
		"displayName": user.DisplayName,
	}
}
