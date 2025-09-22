package config

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/lits-06/vcs-sms/pkg/constants"
	"github.com/lits-06/vcs-sms/pkg/kafka"
	"github.com/lits-06/vcs-sms/pkg/logger"
	"github.com/lits-06/vcs-sms/pkg/redis"
	"github.com/lits-06/vcs-sms/pkg/tracing"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

var configPath string

func init() {
	flag.StringVar(&configPath, "config", "", "Healthcheck microservice config path")
}

type Config struct {
	ServiceName string         `mapstructure:"serviceName"`
	Duration      time.Duration `mapstructure:"duration"`
	Logger      *logger.Config `mapstructure:"logger"`

	Kafka  *kafka.Config   `mapstructure:"kafka"`
	Jaeger *tracing.Config `mapstructure:"jaeger"`

	Elasticsearch Elasticsearch `mapstructure:"elasticsearch"`
	Redis         Redis         `mapstructure:"redis"`
}

type Elasticsearch struct {
	Address       string `mapstructure:"address"`
	SnapshotIndex string `mapstructure:"snapshotIndex"`
	RecordIndex   string `mapstructure:"recordIndex"`
}

type Redis struct {
	*redis.Config
	SnapshotKey string `mapstructure:"snapshotKey"`
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
			configPath = fmt.Sprintf("%s/healthcheck_service/config/config.yaml", getwd)
		}
	}

	cfg := &Config{}

	viper.SetConfigType(constants.Yaml)
	viper.SetConfigFile(configPath)

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("viper.ReadInConfig: %w", err)
	}

	if err := viper.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("viper.Unmarshal: %w", err)
	}

	return cfg, nil
}
