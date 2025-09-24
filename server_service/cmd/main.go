package main

// @title		Server Service API
// @version		1.0
// @description	This is the Server Service API for VCS-SMS system
// @termsOfService	http://swagger.io/terms/
//
// @host server.localhost
//
// @securityDefinitions.apikey	BearerAuth
// @in							header
// @name						Authorization
// @description				Type "Bearer" followed by a space and JWT token.

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
