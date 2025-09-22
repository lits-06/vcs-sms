package config

import (
	"flag"
	"fmt"
	"os"

	"github.com/lits-06/vcs-sms/pkg/constants"
	"github.com/lits-06/vcs-sms/pkg/jwt"
	"github.com/lits-06/vcs-sms/pkg/kafka"
	"github.com/lits-06/vcs-sms/pkg/logger"
	"github.com/lits-06/vcs-sms/pkg/postgres"
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
	Port            string           `mapstructure:"port"`
	Logger          *logger.Config   `mapstructure:"logger"`
	Postgres        *postgres.Config `mapstructure:"postgres"`
	Kafka           *kafka.Config    `mapstructure:"kafka"`
	Jaeger          *tracing.Config  `mapstructure:"jaeger"`
	JWT             *jwt.Config      `mapstructure:"jwt"`
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

	return cfg, nil
}
