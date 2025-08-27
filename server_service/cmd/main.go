package main

import (
	"flag"

	"github.com/lits-06/vcs-sms/pkg/logger"
	"github.com/lits-06/vcs-sms/server_service/config"
)

func main() {
	flag.Parse()

	cfg, err := config.InitConfig()
	if err != nil {
		log.Fatal(err)
	}

	appLogger := logger.NewAppLogger(cfg.Logger)
	appLogger.InitLogger()
	appLogger.WithName("ServerService")

	s := server.NewServer(appLogger, cfg)
	appLogger.Fatal(s.Run())
}
