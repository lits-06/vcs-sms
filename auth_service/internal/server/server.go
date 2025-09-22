package server

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/lits-06/vcs-sms/auth_service/config"
	"github.com/lits-06/vcs-sms/auth_service/internal/client"
	"github.com/lits-06/vcs-sms/auth_service/internal/delivery/http"
	"github.com/lits-06/vcs-sms/auth_service/internal/repository"
	"github.com/lits-06/vcs-sms/auth_service/internal/usecase"
	"github.com/lits-06/vcs-sms/pkg/logger"
	"github.com/lits-06/vcs-sms/pkg/middleware"
	redispkg "github.com/lits-06/vcs-sms/pkg/redis"
	"github.com/lits-06/vcs-sms/pkg/tracing"
	userpb "github.com/lits-06/vcs-sms/proto"
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

	grpcConn, err := client.NewUserServiceConn(ctx, s.cfg)
	if err != nil {
		s.log.Error("Failed to connect to gRPC user service", "error", err)
		return err
	}
	defer grpcConn.Close()

	grpcClient := userpb.NewUserServiceClient(grpcConn)
	cacheRepo := repository.NewCacheRepository(redisClient, s.cfg)
	authUsecase := usecase.NewAuthUsecase(s.cfg, cacheRepo, grpcClient)
	middleware := middleware.NewAuthMiddleware(s.cfg.JWT.AccessSecretKey)

	authHandler := http.NewAuthHandler(s.log, authUsecase, middleware)
	router := gin.Default()
	authHandler.RegisterRoutes(router)

	go func() {
		if err := router.Run(s.cfg.Port); err != nil {
			s.log.Error("Failed to run HTTP server", "error", err)
			cancel()
		}
	}()
	s.log.Info("Auth service is running", "port", s.cfg.Port)

	<-ctx.Done()
	s.log.Info("Shutting down auth service...")

	return nil

}
