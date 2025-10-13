package server

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/lits-06/vcs-sms/pkg/logger"
	"github.com/lits-06/vcs-sms/pkg/middleware"
	"github.com/lits-06/vcs-sms/pkg/postgres"
	"github.com/lits-06/vcs-sms/pkg/tracing"
	"github.com/lits-06/vcs-sms/server_service/config"
	"github.com/lits-06/vcs-sms/server_service/internal/delivery/http"
	"github.com/lits-06/vcs-sms/server_service/internal/delivery/kafka"
	"github.com/lits-06/vcs-sms/server_service/internal/domain"
	"github.com/lits-06/vcs-sms/server_service/internal/repository"
	"github.com/lits-06/vcs-sms/server_service/internal/usecase"
	"github.com/opentracing/opentracing-go"
)

type server struct {
	log logger.Logger
	cfg *config.Config
}

func NewServer(log logger.Logger, cfg *config.Config) *server {
	return &server{
		log: log,
		cfg: cfg,
	}
}

func (s *server) Run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	tracer, closer, err := tracing.NewJaegerTracer(s.cfg.Jaeger)
	if err != nil {
		s.log.Error("Failed to create Jaeger tracer ", "error: ", err)
		return err
	}
	defer closer.Close()
	opentracing.SetGlobalTracer(tracer)

	pgDB, err := postgres.NewPostgresDB(s.cfg.Postgres)
	if err != nil {
		s.log.Error("Failed to connect to Postgres ", "error: ", err)
		return err
	}

	err = pgDB.AutoMigrate(&domain.Server{})
	if err != nil {
		s.log.Error("Failed to auto migrate Postgres ", "error: ", err)
		return err
	}

	serverProducer := kafka.NewProducer(s.log, s.cfg)
	serverProducer.Run()
	defer serverProducer.Close()

	serverRepo := repository.NewServerRepository(pgDB)
	serverUsecase := usecase.NewServerUsecase(serverRepo, serverProducer, s.log)

	middleware := middleware.NewAuthMiddleware(s.cfg.JWT.SecretKey)
	serverHandler := http.NewServerHandler(s.log, serverUsecase, middleware)

	router := gin.Default()
	serverHandler.RegisterRoutes(router)

	serverCG := kafka.NewConsumerGroup(s.cfg.Kafka.Brokers, s.log, serverUsecase)
	serverCG.RunConsumers(ctx, cancel)

	go func() {
		if err := router.Run(s.cfg.Port); err != nil {
			s.log.Error("Failed to run server service on HTTP server ", "error: ", err)
			cancel()
		}
	}()
	s.log.Info("Server service is running ", "http_port: ", s.cfg.Port)

	<-ctx.Done()
	s.log.Info("Shutting down server service...")

	return nil
}
