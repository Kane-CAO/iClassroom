package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"iclassroom/backend/internal/config"
)

func init() { gin.SetMode(gin.TestMode) }

func TestHealthEndpoint(t *testing.T) {
	cfg := &config.Config{
		AppEnv:             "development",
		ServerPort:         "8080",
		CORSAllowedOrigins: []string{"http://localhost:5173"},
		BackendBaseURL:     "http://localhost:8080",
		UploadDir:          "./uploads",
	}
	// db is nil here: the health check must still succeed and report db "down".
	router := newRouter(cfg, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}

	var body struct {
		Success bool `json:"success"`
		Data    struct {
			Status string `json:"status"`
			Env    string `json:"env"`
			DB     string `json:"db"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if !body.Success {
		t.Error("success = false, want true")
	}
	if body.Data.Status != "ok" {
		t.Errorf("data.status = %q, want ok", body.Data.Status)
	}
	if body.Data.DB != "down" {
		t.Errorf("data.db = %q, want down (nil db)", body.Data.DB)
	}
}
