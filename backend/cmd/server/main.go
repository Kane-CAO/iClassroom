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
	"iclassroom/backend/internal/storage"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

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

func newRouter(cfg *config.Config, db *sql.DB) *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())
	router.Use(middleware.CORS(cfg.CORSAllowedOrigins))
	router.Static("/uploads", cfg.UploadDir)

	router.GET("/health", healthHandler(cfg, db))

	if db != nil {
		registerAPIRoutes(router, cfg, db)
	}

	return router
}

func registerAPIRoutes(router *gin.Engine, cfg *config.Config, db *sql.DB) {
	roomRepo := repository.NewRoomRepository(db)
	groupRepo := repository.NewGroupRepository(db)
	studentRepo := repository.NewStudentRepository(db)
	taskRepo := repository.NewTaskRepository(db)
	submissionRepo := repository.NewSubmissionRepository(db)
	featuredRepo := repository.NewFeaturedAnswerRepository(db)
	displayRepo := repository.NewDisplayRepository(db)
	analyticsRepo := repository.NewAnalyticsRepository(db)
	uploadStore := storage.NewLocalStorage(cfg.UploadDir, cfg.BackendBaseURL)
	uploadSvc := service.NewLocalUploadService(uploadStore)

	roomSvc := service.NewRoomService(roomRepo, groupRepo, cfg.FrontendBaseURL)
	studentSvc := service.NewStudentService(roomRepo, groupRepo, studentRepo)
	taskSvc := service.NewTaskService(roomRepo, groupRepo, studentRepo, taskRepo, submissionRepo, uploadSvc)
	featuredSvc := service.NewFeaturedAnswerService(roomRepo, featuredRepo)
	displaySvc := service.NewDisplayService(roomRepo, submissionRepo, displayRepo)
	analyticsSvc := service.NewAnalyticsService(roomRepo, analyticsRepo)
	exportSvc := service.NewExportService(roomRepo, taskRepo, submissionRepo)

	api := router.Group("/api")
	handler.NewRoomHandler(roomSvc).Register(api)
	handler.NewStudentHandler(studentSvc).Register(api)
	handler.NewTaskHandler(taskSvc).Register(api)
	handler.NewFeaturedAnswerHandler(featuredSvc).Register(api)
	handler.NewDisplayHandler(displaySvc).Register(api)
	handler.NewAnalyticsHandler(analyticsSvc).Register(api)
	handler.NewExportHandler(exportSvc).Register(api)
}

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
