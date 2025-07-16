package main

import (
	"log"

	"github.com/lits-06/vcs-sms/api/handler"
	"github.com/lits-06/vcs-sms/api/router"
	"github.com/lits-06/vcs-sms/infrastructure/database"
	"github.com/lits-06/vcs-sms/pkg/logger"
	"github.com/lits-06/vcs-sms/usecases/server"
)

func main() {
	appLogger, err := logger.NewZapLogger()
	if err != nil {
		log.Fatal("Failed to initialize logger:", err)
	}

	appLogger.Info("Starting Server Management System...")

	db, err := database.NewPostgresConnection()
	if err != nil {
		appLogger.Fatal("Failed to connect to database", "error", err)
	}
	defer db.Close()
	appLogger.Info("Database connected successfully")

	serverRepo := database.NewServerRepository(db)
	serverUsecase := server.NewServerUsecase(serverRepo)
	serverHandler := handler.NewServerHandler(serverUsecase, appLogger)

	routes := router.NewRoute(serverHandler)
	r := routes.SetupRoutes()

	appLogger.Info("Server starting on port 8080")
	if err := r.Run(":8080"); err != nil {
		appLogger.Fatal("Failed to start server", "error", err)
	}
}
