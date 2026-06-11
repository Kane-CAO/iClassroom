// Package response defines the unified JSON envelope used by every HTTP
// endpoint (except binary export endpoints), as specified in docs/api.md.
//
// Success: {"success": true,  "message": "success", "data": {...}}
// Error:   {"success": false, "message": "...",     "errorCode": "..."}
package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Body is the unified response envelope. Field names are camelCase to match
// the API contract. Data is omitted on errors; errorCode is omitted on success.
type Body struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	ErrorCode string      `json:"errorCode,omitempty"`
}

// Success writes a 200 OK success envelope with the given payload.
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Body{
		Success: true,
		Message: "success",
		Data:    data,
	})
}

// Error writes an error envelope with the given HTTP status, machine-readable
// errorCode (see docs/api.md), and human-readable message.
func Error(c *gin.Context, status int, errorCode, message string) {
	c.JSON(status, Body{
		Success:   false,
		Message:   message,
		ErrorCode: errorCode,
	})
}
