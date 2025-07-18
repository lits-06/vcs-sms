package main

import (
	"fmt"
	"log"

	"github.com/lits-06/vcs-sms/api/handler"
	"github.com/lits-06/vcs-sms/api/router"
	"github.com/lits-06/vcs-sms/config"
	"github.com/lits-06/vcs-sms/infrastructure/database"
	infraServer "github.com/lits-06/vcs-sms/infrastructure/server"
	"github.com/lits-06/vcs-sms/pkg/logger"
	"github.com/lits-06/vcs-sms/usecases/server"
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

	appLogger.Info("Starting Server Management System...")

	db, err := database.NewGormConnection(&cfg.Database)
	if err != nil {
		appLogger.Fatal("Failed to connect to database", "error", err)
	}

	if err := database.AutoMigrate(db); err != nil {
		appLogger.Fatal("Failed to run migrations", "error", err)
	}
	appLogger.Info("Database connected successfully")

	serverRepo := database.NewServerRepository(db)
	serverProvider := infraServer.NewPortServerProvider()
	serverUsecase := server.NewServerUsecase(serverRepo, serverProvider)
	serverHandler := handler.NewServerHandler(serverUsecase, appLogger)

	routes := router.NewRoute(serverHandler)
	r := routes.SetupRoutes()

	/// Start server
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	appLogger.Info("Server starting", "address", addr)

	if err := r.Run(addr); err != nil {
		appLogger.Fatal("Failed to start server", "error", err)
	}
}
