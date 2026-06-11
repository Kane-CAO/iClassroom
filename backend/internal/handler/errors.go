// Package handler holds the HTTP layer: request binding, calling the service,
// and writing the unified response envelope. Handlers contain no business
// rules — those live in the service layer.
package handler

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"iclassroom/backend/internal/apperr"
	"iclassroom/backend/internal/response"
)

// respondError translates an error into the unified error response. Known
// business errors (*apperr.Error) carry their own status + errorCode;
// everything else is an unexpected internal failure (500) and is logged.
func respondError(c *gin.Context, err error) {
	var ae *apperr.Error
	if errors.As(err, &ae) {
		response.Error(c, ae.Status, ae.Code, ae.Message)
		return
	}
	log.Printf("handler: unexpected error on %s %s: %v", c.Request.Method, c.Request.URL.Path, err)
	response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
}
