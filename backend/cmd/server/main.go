// Command server is the HTTP entrypoint for the iClassroom backend.
//
// Backend Step 0 scope: boot the service, load config, connect to MySQL, and
// expose a health check. No business endpoints are implemented here.
package main

import (
	"database/sql"
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"iclassroom/backend/internal/config"
	"iclassroom/backend/internal/database"
	"iclassroom/backend/internal/handler"
	"iclassroom/backend/internal/middleware"
	"iclassroom/backend/internal/repository"
	"iclassroom/backend/internal/response"
	"iclassroom/backend/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	// Connect to MySQL. In development a missing DB should not stop the server
	// from booting, so /health stays reachable and reports DB status truthfully.
	// In production we fail fast, because the service is useless without a DB.
	db, err := database.New(cfg)
	if err != nil {
		if cfg.IsProduction() {
			log.Fatalf("failed to connect to database: %v", err)
		}
		log.Printf("WARNING: database not connected (dev mode, continuing): %v", err)
	} else {
		log.Printf("database connected: %s:%s/%s", cfg.DBHost, cfg.DBPort, cfg.DBName)
		defer func() { _ = db.Close() }()
	}

	router := newRouter(cfg, db)

	addr := ":" + cfg.ServerPort
	log.Printf("server listening on %s (env=%s)", addr, cfg.AppEnv)
	if err := router.Run(addr); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("server error: %v", err)
	}
}

// newRouter builds the Gin engine with shared middleware and the health check.
// It is separated from main so it can be exercised by tests. db may be nil when
// the database is unavailable in development.
func newRouter(cfg *config.Config, db *sql.DB) *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())
	router.Use(middleware.CORS(cfg.CORSAllowedOrigins))

	router.GET("/health", healthHandler(cfg, db))

	// Business endpoints require a database. In development the server still
	// boots without one (so /health stays reachable); the /api routes are
	// simply not mounted until a DB is available.
	if db != nil {
		registerAPIRoutes(router, cfg, db)
	}

	return router
}

// registerAPIRoutes wires the repository → service → handler layers and mounts
// the /api routes.
func registerAPIRoutes(router *gin.Engine, cfg *config.Config, db *sql.DB) {
	roomRepo := repository.NewRoomRepository(db)
	groupRepo := repository.NewGroupRepository(db)
	studentRepo := repository.NewStudentRepository(db)

	roomSvc := service.NewRoomService(roomRepo, groupRepo, cfg.FrontendBaseURL)
	studentSvc := service.NewStudentService(roomRepo, groupRepo, studentRepo)

	api := router.Group("/api")
	handler.NewRoomHandler(roomSvc).Register(api)
	handler.NewStudentHandler(studentSvc).Register(api)
}

// healthHandler reports service liveness and live database connectivity.
func healthHandler(cfg *config.Config, db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		dbStatus := "down"
		if db != nil {
			if err := db.Ping(); err == nil {
				dbStatus = "up"
			}
		}
		response.Success(c, gin.H{
			"status": "ok",
			"env":    cfg.AppEnv,
			"db":     dbStatus,
		})
	}
}
