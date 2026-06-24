package middleware

import (
	"context"
	"strings"

	"github.com/gin-gonic/gin"

	"iclassroom/backend/internal/apperr"
	"iclassroom/backend/internal/domain"
	"iclassroom/backend/internal/response"
)

const AuthUserKey = "authUser"

type Authenticator interface {
	Authenticate(ctx context.Context, token string) (*domain.AuthUser, error)
}

func OptionalAuth(auth Authenticator) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := bearerToken(c.GetHeader("Authorization"))
		if token == "" {
			c.Next()
			return
		}
		user, err := auth.Authenticate(c.Request.Context(), token)
		if err != nil {
			writeAuthError(c, err)
			c.Abort()
			return
		}
		c.Set(AuthUserKey, user)
		c.Next()
	}
}

func RequireRole(role domain.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := CurrentUser(c)
		if !ok {
			writeAuthError(c, apperr.Unauthorized())
			c.Abort()
			return
		}
		if user.Role != role {
			writeAuthError(c, apperr.Forbidden())
			c.Abort()
			return
		}
		c.Next()
	}
}

func CurrentUser(c *gin.Context) (*domain.AuthUser, bool) {
	value, ok := c.Get(AuthUserKey)
	if !ok {
		return nil, false
	}
	user, ok := value.(*domain.AuthUser)
	return user, ok && user != nil
}

func bearerToken(header string) string {
	header = strings.TrimSpace(header)
	if header == "" {
		return ""
	}
	parts := strings.Fields(header)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return parts[1]
}

func writeAuthError(c *gin.Context, err error) {
	if appErr, ok := err.(*apperr.Error); ok {
		response.Error(c, appErr.Status, appErr.Code, appErr.Message)
		return
	}
	response.Error(c, 401, "UNAUTHORIZED", "missing or invalid session")
}
