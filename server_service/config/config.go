package config

import (
	"flag"
	"fmt"
	"os"

	"github.com/lits-06/vcs-sms/pkg/constants"
	"github.com/lits-06/vcs-sms/pkg/kafka"
	"github.com/lits-06/vcs-sms/pkg/logger"
	"github.com/lits-06/vcs-sms/pkg/postgres"
	"github.com/lits-06/vcs-sms/pkg/probes"
	"github.com/lits-06/vcs-sms/pkg/redis"
	"github.com/lits-06/vcs-sms/pkg/tracing"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

var configPath string

func init() {
	flag.StringVar(&configPath, "config", "", "Server microservice config path")
}

type Config struct {
	ServiceName     string           `mapstructure:"serviceName"`
	Logger          *logger.Config   `mapstructure:"logger"`
	GRPC            GRPC             `mapstructure:"grpc"`
	Postgresql      *postgres.Config `mapstructure:"postgres"`
	Kafka           *kafka.Config    `mapstructure:"kafka"`
	Redis           *redis.Config    `mapstructure:"redis"`
	Probes          probes.Config    `mapstructure:"probes"`
	ServiceSettings ServiceSettings  `mapstructure:"serviceSettings"`
	Jaeger          *tracing.Config  `mapstructure:"jaeger"`
}

type GRPC struct {
	Port        string `mapstructure:"port"`
	Development bool   `mapstructure:"development"`
}

type ServiceSettings struct {
	RedisProductPrefixKey string `mapstructure:"redisProductPrefixKey"`
}

func InitConfig() (*Config, error) {
	if configPath == "" {
		configPathFromEnv := os.Getenv(constants.ConfigPath)
		if configPathFromEnv != "" {
			configPath = configPathFromEnv
		} else {
			getwd, err := os.Getwd()
			if err != nil {
				return nil, errors.Wrap(err, "os.Getwd")
			}
			configPath = fmt.Sprintf("%s/server_service/config/config.yaml", getwd)
		}
	}

	cfg := &Config{}

	viper.SetConfigType(constants.Yaml)
	viper.SetConfigFile(configPath)

	if err := viper.ReadInConfig(); err != nil {
		return nil, errors.Wrap(err, "viper.ReadInConfig")
	}

	if err := viper.Unmarshal(cfg); err != nil {
		return nil, errors.Wrap(err, "viper.Unmarshal")
	}

	grpcPort := os.Getenv(constants.GrpcPort)
	if grpcPort != "" {
		cfg.GRPC.Port = grpcPort
	}
	postgresHost := os.Getenv(constants.PostgresqlHost)
	if postgresHost != "" {
		cfg.Postgresql.Host = postgresHost
	}
	postgresPort := os.Getenv(constants.PostgresqlPort)
	if postgresPort != "" {
		cfg.Postgresql.Port = postgresPort
	}
	redisAddr := os.Getenv(constants.RedisAddr)
	if redisAddr != "" {
		cfg.Redis.Addr = redisAddr
	}
	//jaegerAddr := os.Getenv("JAEGER_HOST")
	//if jaegerAddr != "" {
	//	cfg.Jaeger.HostPort = jaegerAddr
	//}
	//kafkaBrokers := os.Getenv("KAFKA_BROKERS")
	//if kafkaBrokers != "" {
	//	cfg.Kafka.Brokers = []string{"host.docker.internal:9092"}
	//}
	kafkaBrokers := os.Getenv(constants.KafkaBrokers)
	if kafkaBrokers != "" {
		cfg.Kafka.Brokers = []string{kafkaBrokers}
	}
	jaegerAddr := os.Getenv(constants.JaegerHostPort)
	if jaegerAddr != "" {
		cfg.Jaeger.HostPort = jaegerAddr
	}

	return cfg, nil
}
