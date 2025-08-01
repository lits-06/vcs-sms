package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/lits-06/vcs-sms/api/handler"
	"github.com/lits-06/vcs-sms/api/router"
	"github.com/lits-06/vcs-sms/config"
	"github.com/lits-06/vcs-sms/infrastructure/database"
	"github.com/lits-06/vcs-sms/infrastructure/elasticsearch"
	"github.com/lits-06/vcs-sms/infrastructure/redis"
	infraServer "github.com/lits-06/vcs-sms/infrastructure/server"
	"github.com/lits-06/vcs-sms/pkg/logger"
	"github.com/lits-06/vcs-sms/services/server"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	appLogger, err := logger.NewZapLogger(&cfg.Logging)
	if err != nil {
		log.Fatal("Failed to initialize logger:", err)
	}

	db, err := database.NewGormConnection(&cfg.Database)
	if err != nil {
		appLogger.Fatal("Failed to connect to database", "error", err)
	}

	redisClient, err := redis.NewRedisClient(&cfg.Redis)
	if err != nil {
		appLogger.Fatal("Failed to connect to Redis", "error", err)
	}

	esClient, err := elasticsearch.NewElasticsearchClient(&cfg.Elasticsearch)
	if err != nil {
		appLogger.Fatal("Failed to connect to Elasticsearch", "error", err)
	}

	serverRepo := database.NewServerRepository(db)
	cacheRepo := redis.NewCacheRepository(redisClient)
	recordRepo := elasticsearch.NewRecordRepository(esClient)

	serverProvider := infraServer.NewPortServerProvider(redisClient)

	serverUsecase := server.NewServerUsecase(serverRepo, serverProvider)
	emailService := server.NewMailService(&cfg.SMTP, appLogger)
	reportService := server.NewUptimeService(recordRepo, emailService, appLogger)
	monitorService := server.NewMonitorService(
		serverRepo,
		cacheRepo,
		recordRepo,
		serverProvider,
		appLogger,
		cfg.Monitoring.Interval,
	)

	serverHandler := handler.NewServerHandler(serverUsecase, appLogger)
	reportHandler := handler.NewReportHandler(reportService, appLogger)

	routes := router.NewRoute(serverHandler, reportHandler)
	r := routes.SetupRoutes()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	monitorService.Start(ctx)

	/// Start server
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	appLogger.Info("Server starting", "address", addr)

	go func() {
		if err := r.Run(addr); err != nil {
			appLogger.Error("Failed to start server", "error", err)
			cancel()
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	appLogger.Info("Shutting down server...")

	monitorService.Stop()

	cacheRepo.DeleteServerList(ctx)

	appLogger.Info("Server exited")
}
