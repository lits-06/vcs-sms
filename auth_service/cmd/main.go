package main

import (
	"flag"
	"log"

	"github.com/lits-06/vcs-sms/auth_service/config"
	_ "github.com/lits-06/vcs-sms/auth_service/docs"
	"github.com/lits-06/vcs-sms/auth_service/internal/server"
	"github.com/lits-06/vcs-sms/pkg/logger"
)

func main() {
	flag.Parse()

	cfg, err := config.InitConfig()
	if err != nil {
		log.Fatal(err)
	}

	appLogger := logger.NewAppLogger(cfg.Logger)
	appLogger.InitLogger()
	appLogger.WithName("Auth Service")

	s := server.NewServer(appLogger, cfg)
	appLogger.Fatal(s.Run())
}

