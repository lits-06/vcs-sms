package server

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/elastic/go-elasticsearch/v9"
	"github.com/lits-06/vcs-sms/healthcheck_service/config"
	"github.com/lits-06/vcs-sms/healthcheck_service/internal/delivery/kafka"
	"github.com/lits-06/vcs-sms/healthcheck_service/internal/repository"
	"github.com/lits-06/vcs-sms/healthcheck_service/internal/usecase"
	"github.com/lits-06/vcs-sms/pkg/logger"
	redispkg "github.com/lits-06/vcs-sms/pkg/redis"
	"github.com/lits-06/vcs-sms/pkg/tracing"
	"github.com/opentracing/opentracing-go"
	"github.com/redis/go-redis/v9"
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
		s.log.Error("Failed to create Jaeger tracer", "error", err)
		return err
	}
	defer closer.Close()
	opentracing.SetGlobalTracer(tracer)

	esClient, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{s.cfg.Elasticsearch.Address},
	})
	if err != nil {
		s.log.Error("Failed to create Elasticsearch client", "error", err)
		return err
	}

	esInfoRes, err := esClient.Info(esClient.Info.WithContext(ctx))
	if err != nil {
		s.log.Error("Failed to get Elasticsearch info", "error", err)
		return err
	}
	if esInfoRes.IsError() {
		s.log.Error("Elasticsearch info response is error", "error", esInfoRes.String())
		return err
	}

	redisUniversalClient := redispkg.NewUniversalRedisClient(s.cfg.Redis.Config)
	_, err = redisUniversalClient.Ping(ctx).Result()
	if err != nil {
		s.log.Error("Failed to connect to Redis", "error", err)
		return err
	}
	redisClient, ok := redisUniversalClient.(*redis.Client)
	if !ok {
		s.log.Error("Failed to assert Redis client type")
		return err
	}

	healthCheckProducer := kafka.NewProducer(s.log, s.cfg)
	healthCheckProducer.Run()
	defer healthCheckProducer.Close()

	esRepo := repository.NewESRepository(esClient, s.cfg)
	cacheRepo := repository.NewCacheRepository(redisClient, s.cfg)
	healthCheckUC := usecase.NewHealthCheckUseCase(s.cfg, s.log, esRepo, cacheRepo, healthCheckProducer)

	healthCheckCG := kafka.NewConsumerGroup(s.cfg.Kafka.Brokers, s.log, s.cfg, healthCheckUC)
	healthCheckCG.RunConsumers(ctx, cancel)

	go func() {
		healthCheckUC.StartHealthCheckScheduler(ctx)
		defer healthCheckUC.StopHealthCheckScheduler(ctx)
	}()

	<-ctx.Done()
	s.log.Info("Shutting down healthcheck service...")

	return nil
}
