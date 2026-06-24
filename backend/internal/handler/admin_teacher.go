package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"iclassroom/backend/internal/apperr"
	"iclassroom/backend/internal/domain"
	"iclassroom/backend/internal/middleware"
	"iclassroom/backend/internal/response"
	"iclassroom/backend/internal/service"
)

type AdminTeacherHandler struct {
	auth *service.AuthService
}

func NewAdminTeacherHandler(auth *service.AuthService) *AdminTeacherHandler {
	return &AdminTeacherHandler{auth: auth}
}

func (h *AdminTeacherHandler) Register(rg *gin.RouterGroup) {
	admin := rg.Group("/admin", middleware.RequireRole(domain.RoleAdmin))
	admin.POST("/teachers", h.Create)
	admin.GET("/teachers", h.List)
	admin.PATCH("/teachers/:teacherId/status", h.UpdateStatus)
	admin.POST("/teachers/:teacherId/reset-password", h.ResetPassword)
	admin.DELETE("/teachers/:teacherId", h.Delete)
}

type createTeacherRequest struct {
	Username        string `json:"username"`
	DisplayName     string `json:"displayName"`
	InitialPassword string `json:"initialPassword"`
}

type updateTeacherStatusRequest struct {
	Status string `json:"status"`
}

type resetTeacherPasswordRequest struct {
	NewPassword string `json:"newPassword"`
}

func (h *AdminTeacherHandler) Create(c *gin.Context) {
	admin, ok := middleware.CurrentUser(c)
	if !ok || admin.Role != domain.RoleAdmin {
		respondError(c, apperr.Forbidden())
		return
	}

	var req createTeacherRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, apperr.InvalidRequest("malformed request body"))
		return
	}

	teacher, err := h.auth.CreateTeacher(c.Request.Context(), service.CreateTeacherInput{
		Username:         req.Username,
		DisplayName:      req.DisplayName,
		InitialPassword:  req.InitialPassword,
		CreatedByAdminID: admin.UserID,
	})
	if err != nil {
		respondError(c, err)
		return
	}

	response.Success(c, teacherJSON(*teacher))
}

func (h *AdminTeacherHandler) List(c *gin.Context) {
	teachers, err := h.auth.ListTeachers(c.Request.Context())
	if err != nil {
		respondError(c, err)
		return
	}

	out := make([]gin.H, 0, len(teachers))
	for _, teacher := range teachers {
		out = append(out, teacherJSON(teacher))
	}
	response.Success(c, out)
}

func (h *AdminTeacherHandler) UpdateStatus(c *gin.Context) {
	teacherID, ok := teacherIDParam(c)
	if !ok {
		return
	}

	var req updateTeacherStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, apperr.InvalidRequest("malformed request body"))
		return
	}

	teacher, err := h.auth.UpdateTeacherStatus(c.Request.Context(), teacherID, domain.AccountStatus(req.Status))
	if err != nil {
		respondError(c, err)
		return
	}
	response.Success(c, teacherJSON(*teacher))
}

func (h *AdminTeacherHandler) ResetPassword(c *gin.Context) {
	teacherID, ok := teacherIDParam(c)
	if !ok {
		return
	}

	var req resetTeacherPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, apperr.InvalidRequest("malformed request body"))
		return
	}

	teacher, password, err := h.auth.ResetTeacherPassword(c.Request.Context(), teacherID, req.NewPassword)
	if err != nil {
		respondError(c, err)
		return
	}
	res := teacherJSON(*teacher)
	res["temporaryPassword"] = password
	response.Success(c, res)
}

func (h *AdminTeacherHandler) Delete(c *gin.Context) {
	teacherID, ok := teacherIDParam(c)
	if !ok {
		return
	}
	if err := h.auth.DeleteTeacher(c.Request.Context(), teacherID); err != nil {
		respondError(c, err)
		return
	}
	response.Success(c, gin.H{"teacherId": teacherID})
}

func teacherIDParam(c *gin.Context) (int64, bool) {
	teacherID, err := strconv.ParseInt(c.Param("teacherId"), 10, 64)
	if err != nil || teacherID <= 0 {
		respondError(c, apperr.TeacherNotFound())
		return 0, false
	}
	return teacherID, true
}

func teacherJSON(teacher domain.TeacherAccount) gin.H {
	out := gin.H{
		"teacherId":   teacher.ID,
		"username":    teacher.Username,
		"displayName": teacher.DisplayName,
		"status":      teacher.Status,
		"createdAt":   teacher.CreatedAt,
		"updatedAt":   teacher.UpdatedAt,
	}
	if teacher.LastLoginAt != nil {
		out["lastLoginAt"] = teacher.LastLoginAt
	}
	return out
}
