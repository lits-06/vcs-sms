package main

import (
	"flag"
	"log"

	"github.com/lits-06/vcs-sms/pkg/logger"
	"github.com/lits-06/vcs-sms/server_service/config"
	"github.com/lits-06/vcs-sms/server_service/internal/server"

	_ "github.com/lits-06/vcs-sms/server_service/docs"
)

func main() {
	flag.Parse()

	cfg, err := config.InitConfig()
	if err != nil {
		log.Fatal(err)
	}

	appLogger := logger.NewAppLogger(cfg.Logger)
	appLogger.InitLogger()
	appLogger.WithName("Server Service")

	s := server.NewServer(appLogger, cfg)
	appLogger.Fatal(s.Run())
}
