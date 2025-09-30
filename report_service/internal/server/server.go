package server

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/elastic/go-elasticsearch/v9"
	"github.com/gin-gonic/gin"
	"github.com/lits-06/vcs-sms/pkg/logger"
	"github.com/lits-06/vcs-sms/pkg/middleware"
	"github.com/lits-06/vcs-sms/pkg/tracing"
	"github.com/lits-06/vcs-sms/report_service/config"
	"github.com/lits-06/vcs-sms/report_service/internal/delivery/http"
	"github.com/lits-06/vcs-sms/report_service/internal/repository"
	"github.com/lits-06/vcs-sms/report_service/internal/usecase"
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

	esClient, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{s.cfg.Elasticsearch.Address},
	})
	if err != nil {
		s.log.Error("Failed to create Elasticsearch client ", "error: ", err)
		return err
	}

	esInfoRes, err := esClient.Info(esClient.Info.WithContext(ctx))
	if err != nil {
		s.log.Error("Failed to get Elasticsearch info ", "error: ", err)
		return err
	}
	if esInfoRes.IsError() {
		s.log.Error("Elasticsearch info response is error ", "error: ", esInfoRes.String())
		return err
	}

	recordRepo := repository.NewRecordRepository(esClient, s.cfg)
	reportUsecase := usecase.NewReportUseCase(recordRepo, s.cfg)

	middleware := middleware.NewAuthMiddleware(s.cfg.JWT.SecretKey)
	reportHandler := http.NewReportHandler(s.log, reportUsecase, middleware)

	router := gin.Default()
	reportHandler.RegisterRoutes(router)

	go func() {
		ticker := time.NewTicker(s.cfg.Duration)
		for {
			select {
			case <-ticker.C:
				err := reportUsecase.ReportStats(ctx, "", time.Now().Add(-s.cfg.Duration), time.Now())
				if err != nil {
					s.log.Error("Failed to generate and send reports ", "error: ", err)
				} else {
					s.log.Info("Successfully generated and sent reports")
				}
			case <-ctx.Done():
				ticker.Stop()
				return
			}
		}
	}()

	go func() {
		if err := router.Run(s.cfg.Port); err != nil {
			s.log.Error("Failed to run report service on HTTP server ", "error: ", err)
			cancel()
		}
	}()
	s.log.Info("Report service is running ", "http_port: ", s.cfg.Port)

	<-ctx.Done()
	s.log.Info("Shutting down report service...")

	return nil
}
