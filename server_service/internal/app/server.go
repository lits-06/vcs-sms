package app

import (
	"github.com/go-playground/validator"
	"github.com/lits-06/vcs-sms/pkg/logger"
	"github.com/lits-06/vcs-sms/server_service/config"
	"github.com/lits-06/vcs-sms/server_service/internal/metrics"
	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
)

type server struct {
	log         logger.Logger
	cfg         *config.Config
	v           *validator.Validate
	kafkaConn   *kafka.Conn
	im          interceptors.InterceptorManager
	redisClient redis.UniversalClient
	ps          *service.ProductService
	metrics     *metrics.ReaderServiceMetrics
}