package server

import (
	"context"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/lits-06/vcs-sms/pkg/grpc"
	"github.com/lits-06/vcs-sms/pkg/logger"
	"github.com/lits-06/vcs-sms/pkg/middleware"
	"github.com/lits-06/vcs-sms/pkg/postgres"
	"github.com/lits-06/vcs-sms/pkg/tracing"
	userpb "github.com/lits-06/vcs-sms/proto"
	"github.com/lits-06/vcs-sms/user_service/config"
	grpcService "github.com/lits-06/vcs-sms/user_service/internal/delivery/grpc"
	"github.com/lits-06/vcs-sms/user_service/internal/delivery/http"
	"github.com/lits-06/vcs-sms/user_service/internal/domain"
	"github.com/lits-06/vcs-sms/user_service/internal/repository"
	"github.com/lits-06/vcs-sms/user_service/internal/usecase"
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
		s.log.Error("Failed to create Jaeger tracer", "error", err)
		return err
	}
	defer closer.Close()
	opentracing.SetGlobalTracer(tracer)

	pgDB, err := postgres.NewPostgresDB(s.cfg.Postgres)
	if err != nil {
		s.log.Error("Failed to connect to Postgres", "error", err)
		return err
	}
	
	err = pgDB.AutoMigrate(&domain.User{}, &domain.Scope{})
	if err != nil {
		s.log.Error("Failed to auto migrate Postgres", "error", err)
		return err
	}

	err = domain.InitScopes(pgDB)
	if err != nil {
		s.log.Error("Failed to init scopes", "error", err)
		return err
	}

	userRepo := repository.NewUserRepository(pgDB)
	userUsecase := usecase.NewUserUsecase(userRepo)

	middleware := middleware.NewAuthMiddleware(s.cfg.JWT.SecretKey)
	userHandler := http.NewUserHandler(s.log, userUsecase, middleware)

	router := gin.Default()
	userHandler.RegisterRoutes(router)

	l, err := net.Listen("tcp", s.cfg.GRPC.Port)
	if err != nil {
		s.log.Error("Failed to listen tcp", "error", err)
		return err
	}
	defer l.Close()

	grpcServer := grpc.NewGRPCServer()
	userService := grpcService.NewUserService(s.log, userUsecase)
	userpb.RegisterUserServiceServer(grpcServer, userService)

	go func() {
		s.log.Info("Starting gRPC server", "port", s.cfg.GRPC.Port)
		s.log.Fatal(grpcServer.Serve(l))
	}()

	go func() {
		if err := router.Run(s.cfg.Port); err != nil {
			s.log.Error("Failed to run HTTP server", "error", err)
			cancel()
		}
	}()
	s.log.Info("User service is running", "http_port", s.cfg.Port, "grpc_port", s.cfg.GRPC.Port)

	<-ctx.Done()

	grpcServer.GracefulStop()
	s.log.Info("Shutting down user service")

	return nil
}