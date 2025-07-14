package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"

	"VCS-Checkpoint1/internal/config"
	"VCS-Checkpoint1/internal/delivery/http/handlers"
	"VCS-Checkpoint1/internal/delivery/http/middleware"
	"VCS-Checkpoint1/internal/delivery/http/router"
	"VCS-Checkpoint1/internal/infrastructure/database"
	"VCS-Checkpoint1/internal/infrastructure/elasticsearch"
	"VCS-Checkpoint1/internal/infrastructure/logger"
	"VCS-Checkpoint1/internal/infrastructure/redis"
	"VCS-Checkpoint1/internal/repository/cache"
	"VCS-Checkpoint1/internal/repository/postgres"
	"VCS-Checkpoint1/internal/services"
	"VCS-Checkpoint1/internal/usecase/server"
	"VCS-Checkpoint1/pkg/validator"
)

// @title Server Management API
// @version 1.0
// @description API for managing server status monitoring
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Initialize logger
	zapLogger, err := logger.NewLogger(cfg.LogLevel, cfg.LogFile)
	if err != nil {
		log.Fatal("Failed to initialize logger:", err)
	}
	defer zapLogger.Sync()

	// Initialize database
	db, err := database.NewPostgresConnection(cfg.Database)
	if err != nil {
		zapLogger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	// Initialize Redis
	redisClient, err := redis.NewRedisClient(cfg.Redis)
	if err != nil {
		zapLogger.Fatal("Failed to connect to Redis", zap.Error(err))
	}
	defer redisClient.Close()

	// Initialize Elasticsearch
	esClient, err := elasticsearch.NewClient(cfg.Elasticsearch)
	if err != nil {
		zapLogger.Fatal("Failed to connect to Elasticsearch", zap.Error(err))
	}

	// Initialize repositories
	serverRepo := postgres.NewServerRepository(db, zapLogger)
	cacheRepo := cache.NewCacheRepository(redisClient, zapLogger)

	// Initialize services
	uptimeService := services.NewUptimeService(esClient, zapLogger)
	monitoringService := services.NewMonitoringService(serverRepo, uptimeService, zapLogger)
	excelService := services.NewExcelService(serverRepo, zapLogger)
	reportService := services.NewReportService(serverRepo, uptimeService, zapLogger, cfg.SMTP)

	// Initialize validator
	v := validator.NewValidator()

	// Initialize usecase
	serverUsecase := server.NewServerUsecase(serverRepo, cacheRepo, excelService, zapLogger)

	// Initialize handlers
	serverHandler := handlers.NewServerHandler(serverUsecase, v, zapLogger)

	// Initialize middleware
	middlewareInstance := middleware.NewMiddleware(zapLogger, cfg.JWT.Secret)

	// Initialize router
	r := router.NewRouter(serverHandler, middlewareInstance)

	// Add Swagger
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Create HTTP server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Server.Port),
		Handler: r,
	}

	// Start monitoring service in background
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go monitoringService.Start(ctx, time.Duration(cfg.MonitoringInterval)*time.Second)
	go reportService.StartDailyReports(ctx)

	// Start server in a goroutine
	go func() {
		zapLogger.Info("Starting HTTP server", zap.Int("port", cfg.Server.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zapLogger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	zapLogger.Info("Shutting down server...")

	// Cancel context to stop background services
	cancel()
	monitoringService.Stop()

	// Graceful shutdown with timeout
	ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		zapLogger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	zapLogger.Info("Server exited")
}
