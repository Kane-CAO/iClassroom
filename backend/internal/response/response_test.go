package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() { gin.SetMode(gin.TestMode) }

func TestSuccess(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	Success(c, gin.H{"status": "ok"})

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}

	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if body["success"] != true {
		t.Errorf("success = %v, want true", body["success"])
	}
	if body["message"] != "success" {
		t.Errorf("message = %v, want success", body["message"])
	}
	if _, ok := body["data"]; !ok {
		t.Error("data field missing on success response")
	}
	if _, ok := body["errorCode"]; ok {
		t.Error("errorCode must be omitted on success response")
	}
}

func TestError(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	Error(c, http.StatusConflict, "NICKNAME_DUPLICATED", "nickname already exists")

	if w.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409", w.Code)
	}

	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if body["success"] != false {
		t.Errorf("success = %v, want false", body["success"])
	}
	if body["errorCode"] != "NICKNAME_DUPLICATED" {
		t.Errorf("errorCode = %v, want NICKNAME_DUPLICATED", body["errorCode"])
	}
	if body["message"] != "nickname already exists" {
		t.Errorf("message = %v, want nickname already exists", body["message"])
	}
	if _, ok := body["data"]; ok {
		t.Error("data must be omitted on error response")
	}
}
