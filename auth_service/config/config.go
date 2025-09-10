package config

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/lits-06/vcs-sms/pkg/constants"
	"github.com/lits-06/vcs-sms/pkg/kafka"
	"github.com/lits-06/vcs-sms/pkg/logger"
	"github.com/lits-06/vcs-sms/pkg/probes"
	"github.com/lits-06/vcs-sms/pkg/tracing"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

var configPath string

func init() {
	flag.StringVar(&configPath, "config", "", "Auth microservice config path")
}

type Config struct {
	ServiceName string          `mapstructure:"serviceName"`
	Logger      *logger.Config  `mapstructure:"logger"`
	KafkaTopics KafkaTopics     `mapstructure:"kafkaTopics"`
	Http        Http            `mapstructure:"http"`
	Grpc        Grpc            `mapstructure:"grpc"`
	Kafka       *kafka.Config   `mapstructure:"kafka"`
	Probes      probes.Config   `mapstructure:"probes"`
	Jaeger      *tracing.Config `mapstructure:"jaeger"`
	JWT         JWT             `mapstructure:"jwt"`
}

type Grpc struct {
	UserServicePort string `mapstructure:"userServicePort"`
}

type JWT struct {
	AccessSecretKey  string        `mapstructure:"accessSecretKey"`
	RefreshSecretKey string        `mapstructure:"refreshSecretKey"`
	AccessTTL        time.Duration `mapstructure:"accessTTL"`
	RefreshTTL       time.Duration `mapstructure:"refreshTTL"`
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
			configPath = fmt.Sprintf("%s/api_gateway_service/config/config.yaml", getwd)
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

	httpPort := os.Getenv(constants.HttpPort)
	if httpPort != "" {
		cfg.Http.Port = httpPort
	}
	kafkaBrokers := os.Getenv(constants.KafkaBrokers)
	if kafkaBrokers != "" {
		cfg.Kafka.Brokers = []string{kafkaBrokers}
	}
	jaegerAddr := os.Getenv(constants.JaegerHostPort)
	if jaegerAddr != "" {
		cfg.Jaeger.HostPort = jaegerAddr
	}
	readerServicePort := os.Getenv(constants.ReaderServicePort)
	if readerServicePort != "" {
		cfg.Grpc.ReaderServicePort = readerServicePort
	}

	return cfg, nil
}
