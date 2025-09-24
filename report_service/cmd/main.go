package main

// @title Report Service API
// @version 1.0
// @description API for generating and managing uptime reports

// @host report.localhost

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

import (
	"flag"
	"log"

	"github.com/lits-06/vcs-sms/pkg/logger"
	"github.com/lits-06/vcs-sms/report_service/config"
	_ "github.com/lits-06/vcs-sms/report_service/docs"
	"github.com/lits-06/vcs-sms/report_service/internal/server"
)

func main() {
	flag.Parse()

	cfg, err := config.InitConfig()
	if err != nil {
		log.Fatal(err)
	}

	appLogger := logger.NewAppLogger(cfg.Logger)
	appLogger.InitLogger()
	appLogger.WithName("Report Service")

	s := server.NewServer(appLogger, cfg)
	appLogger.Fatal(s.Run())
}
